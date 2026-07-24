package stream

import (
	"fmt"
	"polystore_database/src/go/codec"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (qp *Processor) newReadSession() neo4j.SessionWithContext {
	return qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, FetchSize: neo4j.FetchAll})
}

func (qp *Processor) closeSession(s neo4j.SessionWithContext) { _ = s.Close(qp.ctx) }

func ScanGraphStream(qp *Processor,
	o *plan.EntityScan, output chan<- []Record) (int, error) {
	var whereSections []string
	params := make(map[string]interface{})

	// 1. ラベル条件の構築 (n:L1 OR n:L2 ...)
	if len(o.Labels) > 0 {
		var labelConditions []string
		for _, l := range o.Labels {
			labelConditions = append(labelConditions, fmt.Sprintf("n:%s", l))
		}
		whereSections = append(whereSections, "("+strings.Join(labelConditions, " OR ")+")")
	}

	// 2. プロパティフィルタ条件の構築 (n.prop = $val ...)
	for i, cond := range o.Filter {
		var operator string
		switch cond.Type {
		case plan.CondEq:
			operator = "="
		case plan.CondNeq:
			operator = "<>"
		case plan.CondGreater:
			operator = ">"
		case plan.CondLess:
			operator = "<"
		case plan.CondGreaterEq:
			operator = ">="
		case plan.CondLessEq:
			operator = "<="
		default:
			return 0, fmt.Errorf("unknown operator")
		}

		paramName := fmt.Sprintf("val%d", i)
		whereSections = append(whereSections, fmt.Sprintf("n.%s %s $%s", cond.Property, operator, paramName))
		params[paramName], _ = codec.ConvertToNativeType(cond.Value, cond.DataType)
	}

	// 3. クエリの組み立て
	query := "MATCH (n)"
	if len(whereSections) > 0 {
		query += "\nWHERE " + strings.Join(whereSections, " AND ")
	}
	query += "\nRETURN n.uuid AS id"

	sess := qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, FetchSize: neo4j.FetchAll})
	defer sess.Close(qp.ctx)

	// 4. 実行
	res, err := sess.Run(qp.ctx, query, params)
	if err != nil {
		return 0, err
	}

	// 5. ストリーミング処理
	const outputBatchSize = 2000
	rowCount := 0

	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]

	currentBatch := make([]Record, 0, outputBatchSize)
	for res.Next(qp.ctx) {
		if idVal, ok := res.Record().Get("id"); ok && idVal != nil {
			newSlots := make([]uid.UUID, newSlotCount)
			newSlots[aliasIdx] = uid.FromAny(idVal)

			currentBatch = append(currentBatch, Record{Slots: newSlots})
			rowCount++

			if len(currentBatch) >= outputBatchSize {
				output <- currentBatch
				currentBatch = make([]Record, 0, outputBatchSize)
			}
		}
	}
	if len(currentBatch) > 0 {
		output <- currentBatch
	}

	return rowCount, res.Err()

}

func streamFilterGraph(qp *Processor, o *plan.Filter, inputStream <-chan []Record, outputStream chan<- []Record) (int, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	var targetVar, matchPattern string
	if o.ObjType == plan.Relationship {
		targetVar = "r"
		matchPattern = "()-[r]->()"
	} else {
		targetVar = "n"
		matchPattern = "(n:Entity)"
	}

	var labelConditions []string
	for _, l := range o.Labels {
		labelConditions = append(labelConditions, fmt.Sprintf("%s:%s", targetVar, l))
	}
	labelFilter := strings.Join(labelConditions, " OR ")

	var whereClauses []string
	params := make(map[string]interface{})
	for i, cond := range o.Filter {
		operator := "="
		switch cond.Type {
		case plan.CondEq:
			operator = "="
		case plan.CondNeq:
			operator = "<>"
		case plan.CondGreater:
			operator = ">"
		case plan.CondLess:
			operator = "<"
		case plan.CondGreaterEq:
			operator = ">="
		case plan.CondLessEq:
			operator = "<="
		}
		paramName := fmt.Sprintf("val%d", i)
		whereClauses = append(whereClauses, fmt.Sprintf("%s.%s %s $%s", targetVar, cond.Property, operator, paramName))
		params[paramName], _ = codec.ConvertToNativeType(cond.Value, cond.DataType)
	}

	finalQuery := fmt.Sprintf(`
        MATCH %s
        WHERE (%s) 
          AND %s.uuid IN $ids 
          AND %s 
        RETURN %s.uuid AS id`,
		matchPattern, labelFilter, targetVar, strings.Join(whereClauses, " AND "), targetVar,
	)

	return runBatches(
		qp.ctx, qp.exec, qp.sem, OpFilter, inputStream, outputStream,
		qp.newReadSession, qp.closeSession,
		func(sess neo4j.SessionWithContext, batch []Record) ([]Record, error) {
			idMap := make(map[uid.UUID]struct{})
			for _, r := range batch {
				idMap[r.Slots[filterIdxIn]] = struct{}{}
			}
			uniqueIDs := make([]string, 0, len(idMap))
			for id := range idMap {
				uniqueIDs = append(uniqueIDs, id.String())
			}

			localParams := make(map[string]interface{}, len(params)+1)
			for k, v := range params {
				localParams[k] = v
			}
			localParams["ids"] = uniqueIDs

			res, err := sess.Run(qp.ctx, finalQuery, localParams)
			if err != nil {
				return nil, err
			}

			validMap := make(map[uid.UUID]struct{})
			for res.Next(qp.ctx) {
				if id, ok := res.Record().Get("id"); ok && id != nil {
					validMap[uid.FromAny(id)] = struct{}{}
				}
			}
			if err := res.Err(); err != nil {
				return nil, err
			}

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

func fetchGraphPropsStream(qp *Processor,
	ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if len(ids) == 0 || len(fetch.Props) == 0 {
		return result
	}
	// 1. objType に応じてターゲット変数と MATCH パターンを切り替え
	var targetVar string
	var matchPattern string
	if unit.ObjType == plan.Relationship {
		targetVar = "r"
		matchPattern = fmt.Sprintf("()-[%s]->()", targetVar)
	} else if unit.ObjType == plan.Entity {
		targetVar = "n"
		matchPattern = fmt.Sprintf("(%s:Entity)", targetVar)
	} else {
		panic("unknown ObjectType passed")
	}

	// 2. ラベル（またはリレーションタイプ）条件の構築
	var labelFilter string
	if len(unit.Labels) > 0 {
		var labelConditions []string
		for _, l := range unit.Labels {
			labelConditions = append(labelConditions, fmt.Sprintf("%s:%s", targetVar, l))
		}
		labelFilter = fmt.Sprintf("WHERE (%s) AND ", strings.Join(labelConditions, " OR "))
	} else {
		labelFilter = "WHERE "
	}

	// 3. RETURN 句のプロパティ指定を構築
	var propReturns []string
	for _, p := range fetch.Props {
		propReturns = append(propReturns, fmt.Sprintf("%s.%s AS %s", targetVar, p, p))
	}

	// 4. クエリの組み立て
	query := fmt.Sprintf(`
        MATCH %s
        %s %s.uuid IN $ids
        RETURN %s.uuid AS uuid, %s`,
		matchPattern,
		labelFilter,
		targetVar,
		targetVar,
		strings.Join(propReturns, ", "),
	)
	// 5. 実行
	sess := qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead, FetchSize: neo4j.FetchAll})
	defer sess.Close(qp.ctx)

	res, err := sess.Run(qp.ctx, query, map[string]interface{}{"ids": ids})
	if err != nil {
		// この関数は error を返さないシグネチャのため伝播できない。
		// 空ブランチのままだと直後の res.Next で nil 参照 panic するので、空結果を返す。
		return result
	}

	// 5. 結果のパースと型変換
	for res.Next(qp.ctx) {
		rec := res.Record()
		idVal, _ := rec.Get("uuid")
		id, ok := idVal.(string)
		if !ok {
			continue
		}

		propsMap := make(map[string]interface{})
		for _, p := range fetch.Props {
			if val, ok := rec.Get(p); ok && val != nil {
				// FetchPlan の typeMap に基づいて型変換を実行
				typeName := fetch.TypeMap[p]
				propsMap[p], _ = codec.ConvertToNativeType(val, typeName)
			}
		}
		result[id] = propsMap
	}

	return result
}
