package stream

import (
	"fmt"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"strings"
	"time"
)

func ScanColStream(qp *Processor,
	o *plan.EntityScan, step int, output chan<- []Record) (int, error) {
	if qp.cqlSes == nil {
		return 0, nil
	}
	t0 := time.Now()

	// --- DB特有: WHERE 構築 ---
	var whereClauses []string
	var args []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		whereClauses = append(whereClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, core.CQLOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		args = append(args, val)
	}
	whereStr := ""
	if len(whereClauses) > 0 {
		whereStr = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// --- graph と同じストリーミング骨格（batch 幅は VectorWidth に統一）---
	outputBatchSize := qp.exec.vectorWidth()
	rowCount := 0
	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]
	currentBatch := make([]Record, 0, outputBatchSize)

	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM \"%s\" %s ALLOW FILTERING", label, whereStr)
		qp.countRoundTrip()
		iter := qp.cqlSes.Query(query, args...).Iter()
		var id string
		for iter.Scan(&id) {
			if id == "" {
				continue
			}
			newSlots := make([]uid.UUID, newSlotCount)
			newSlots[aliasIdx] = uid.UUID(id)
			currentBatch = append(currentBatch, Record{Slots: newSlots})
			rowCount++
			if len(currentBatch) >= outputBatchSize {
				output <- currentBatch
				currentBatch = make([]Record, 0, outputBatchSize)
			}
		}
		if err := iter.Close(); err != nil {
			return rowCount, err
		}
	}
	if len(currentBatch) > 0 {
		output <- currentBatch
	}
	qp.recordScan(step, rowCount, t0)
	return rowCount, nil
}

func FilterColStream(qp *Processor,
	o *plan.Filter, step int, inputStream <-chan []Record, outputStream chan<- []Record) (int, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	// --- DB特有: 共通条件 ---
	var commonClauses []string
	var commonArgs []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		commonClauses = append(commonClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, core.CQLOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		commonArgs = append(commonArgs, val)
	}
	whereClause := strings.Join(append([]string{"uuid IN ?"}, commonClauses...), " AND ")

	return runBatches(
		qp.ctx, qp.exec, qp.sem, OpFilter, qp, step, inputStream, outputStream,
		noResource, closeNoResource,
		func(_ struct{}, batch []Record) ([]Record, error) {
			idMap := make(map[uid.UUID]struct{})
			for _, r := range batch {
				idMap[r.Slots[filterIdxIn]] = struct{}{}
			}
			uniqueIDs := make([]string, 0, len(idMap))
			for id := range idMap {
				uniqueIDs = append(uniqueIDs, id.String())
			}

			// --- DB特有: valid 抽出 ---
			queryArgs := append([]interface{}{uniqueIDs}, commonArgs...)
			validMap := make(map[uid.UUID]struct{})
			for _, label := range o.Labels {
				query := fmt.Sprintf("SELECT uuid FROM \"%s\" WHERE %s ALLOW FILTERING", label, whereClause)
				qp.countRoundTrip()
				iter := qp.cqlSes.Query(query, queryArgs...).Iter()
				var id string
				for iter.Scan(&id) {
					validMap[uid.UUID(id)] = struct{}{}
				}
				if err := iter.Close(); err != nil {
					return nil, err
				}
			}

			// --- graph と同じ列引き継ぎ ---
			out := make([]Record, 0, len(batch))
			for _, r := range batch {
				if _, ok := validMap[r.Slots[filterIdxIn]]; ok {
					newRec := Record{Slots: make([]uid.UUID, newSlotCount)}
					for alias, outIdx := range o.OutputSlot.VarToSlot {
						if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
							newRec.Slots[outIdx] = r.Slots[inIdx]
						}
					}
					out = append(out, newRec)
				}
			}
			return out, nil
		},
	)
}

func fetchColPropsStream(qp *Processor,
	ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if qp.cqlSes == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}

	const batchSize = 500
	quoted := make([]string, len(fetch.Props))
	for i, p := range fetch.Props {
		quoted[i] = fmt.Sprintf("\"%s\"", p)
	}
	propList := strings.Join(quoted, ", ")

	for _, table := range unit.Labels {
		for i := 0; i < len(ids); i += batchSize {
			end := i + batchSize
			if end > len(ids) {
				end = len(ids)
			}
			batch := ids[i:end]

			query := fmt.Sprintf("SELECT uuid, %s FROM \"%s\" WHERE uuid IN ?", propList, table)
			qp.countRoundTrip()
			iter := qp.cqlSes.Query(query, batch).Iter()
			for {
				row := make(map[string]interface{})
				if !iter.MapScan(row) {
					break
				}
				idRaw, ok := row["uuid"]
				if !ok || idRaw == nil {
					continue
				}
				id := fmt.Sprintf("%v", idRaw)
				if _, exists := result[id]; !exists {
					result[id] = make(map[string]interface{})
				}
				for _, p := range fetch.Props {
					if val, ok := row[p]; ok && val != nil {
						result[id][p], _ = codec.ConvertToNativeType(val, fetch.TypeMap[p])
					}
				}
			}
			if err := iter.Close(); err != nil {
				continue
			}
		}
	}
	return result
}
