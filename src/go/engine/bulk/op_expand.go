package bulk

import (
	"fmt"
	"strings"

	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

type PathResult struct {
	OriginID string   // 探索開始ノードのUUID
	TargetID string   // 到着ノードのUUID
	RelIDs   []string // パス上の関係IDリスト
	NodeIDs  []string // パス上のノードUUIDリスト（ターゲット含む）
	HopLen   int
}

type VarPathResult struct {
	TargetID string
	// RelIDs   []string // 必要に応じて保持
}

// ExpandGraphBulk は 1 段固定長 Expand。全入力行の src.uuid をまとめて
// 1 回の Cypher (WHERE src.uuid IN $ids) で展開する。
func ExpandGraphBulk(qp *Processor, o *plan.Expand, in []Record) ([]Record, error) {
	srcIdx := o.InputSlot.VarToSlot[o.SourceEntity]
	relIdxOut, hasRel := o.OutputSlot.VarToSlot[o.Alias]
	tgtIdxOut, hasTarget := o.OutputSlot.VarToSlot[o.TargetEntity]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	returns := "src.uuid AS sid"
	if hasRel {
		returns += fmt.Sprintf(", %s.uuid AS rid", o.Alias)
	}
	if hasTarget {
		returns += ", tgt.uuid AS tid"
	}

	relConstraint := ""
	if o.RelLabel != "" {
		relConstraint = ":" + o.RelLabel
	}
	tgtConstraint := ""
	if len(o.TargetLabels) > 0 {
		tgtConstraint = ":" + strings.Join(o.TargetLabels, "|:")
	}
	relDef := fmt.Sprintf("[%s%s]", o.Alias, relConstraint)
	var pattern string
	switch o.Dir {
	case plan.Outgoing:
		pattern = fmt.Sprintf("(src:Entity)-%s->(tgt%s)", relDef, tgtConstraint)
	case plan.Incoming:
		pattern = fmt.Sprintf("(src:Entity)<-%s-(tgt%s)", relDef, tgtConstraint)
	case plan.Bidirectional:
		pattern = fmt.Sprintf("(src:Entity)-%s-(tgt%s)", relDef, tgtConstraint)
	default:
		pattern = fmt.Sprintf("(src:Entity)-%s->(tgt%s)", relDef, tgtConstraint)
	}
	finalQuery := fmt.Sprintf("MATCH %s WHERE src.uuid IN $ids RETURN %s", pattern, returns)

	srcIds := make([]string, 0, len(in))
	recordMap := make(map[uid.UUID][]Record)
	for _, r := range in {
		id := r.Slots[srcIdx]
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id.String())
		}
		recordMap[id] = append(recordMap[id], r)
	}

	sess := qp.newReadSession()
	defer qp.closeSession(sess)

	res, err := sess.Run(qp.ctx, finalQuery, map[string]interface{}{"ids": srcIds})
	if err != nil {
		return nil, err
	}

	out := make([]Record, 0, len(in))
	for res.Next(qp.ctx) {
		dbRec := res.Record()
		sid := uid.FromAny(dbRec.Values[0])
		for _, originalRec := range recordMap[sid] {
			newSlots := make([]uid.UUID, newSlotCount)
			for alias, outIdx := range o.OutputSlot.VarToSlot {
				if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
					newSlots[outIdx] = originalRec.Slots[inIdx]
				}
			}
			if hasRel {
				if rid, ok := dbRec.Get("rid"); ok && rid != nil {
					newSlots[relIdxOut] = uid.FromAny(rid)
				}
			}
			if hasTarget {
				if tid, ok := dbRec.Get("tid"); ok && tid != nil {
					newSlots[tgtIdxOut] = uid.FromAny(tid)
				}
			}
			out = append(out, Record{Slots: newSlots})
		}
	}
	return out, res.Err()
}

// bulkVarLengthExpand は可変長 Expand。
func bulkVarLengthExpand(qp *Processor, o *plan.VarLengthExpand, in []Record) ([]Record, error) {
	srcIdxIn := o.InputSlot.VarToSlot[o.SourceEntity]
	tgtIdxOut, hasTarget := o.OutputSlot.VarToSlot[o.TargetEntity]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	relLabel := ""
	if o.RelLabel != "" {
		relLabel = ":" + o.RelLabel
	}
	relContent := fmt.Sprintf("[%s%s*%d..%d]", o.Alias, relLabel, o.MinHops, o.MaxHops)
	var relPattern string
	switch o.Dir {
	case plan.Incoming:
		relPattern = fmt.Sprintf("<-%s-", relContent)
	case plan.Bidirectional:
		relPattern = fmt.Sprintf("-%s-", relContent)
	default:
		relPattern = fmt.Sprintf("-%s->", relContent)
	}
	tgtConstraint := ""
	if len(o.TargetLabels) > 0 {
		tgtConstraint = ":" + strings.Join(o.TargetLabels, "|:")
	}
	finalQuery := fmt.Sprintf(`
		MATCH (src:Entity)%s(tgt%s)
		WHERE src.uuid IN $ids
		RETURN DISTINCT src.uuid AS sid, tgt.uuid AS tid`,
		relPattern, tgtConstraint,
	)

	srcIds := make([]string, 0, len(in))
	recordMap := make(map[uid.UUID][]Record)
	for _, r := range in {
		id := r.Slots[srcIdxIn]
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id.String())
		}
		recordMap[id] = append(recordMap[id], r)
	}

	sess := qp.newReadSession()
	defer qp.closeSession(sess)

	res, err := sess.Run(qp.ctx, finalQuery, map[string]interface{}{"ids": srcIds})
	if err != nil {
		return nil, err
	}

	out := make([]Record, 0, len(in))
	reachedSids := make(map[uid.UUID]struct{})
	carry := func(originalRec Record, targetID uid.UUID) {
		newSlots := make([]uid.UUID, newSlotCount)
		for alias, outIdx := range o.OutputSlot.VarToSlot {
			if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
				newSlots[outIdx] = originalRec.Slots[inIdx]
			}
		}
		if hasTarget {
			newSlots[tgtIdxOut] = targetID
		}
		out = append(out, Record{Slots: newSlots})
	}

	for res.Next(qp.ctx) {
		rec := res.Record()
		sid := uid.FromAny(rec.Values[0])
		tid := uid.FromAny(rec.Values[1])
		reachedSids[sid] = struct{}{}
		for _, originalRec := range recordMap[sid] {
			carry(originalRec, tid)
		}
	}
	if err := res.Err(); err != nil {
		return nil, err
	}

	// 0ホップ（自分自身）
	if o.MinHops == 0 {
		for _, sidStr := range srcIds {
			sid := uid.UUID(sidStr)
			if _, ok := reachedSids[sid]; ok {
				continue
			}
			for _, originalRec := range recordMap[sid] {
				carry(originalRec, sid)
			}
		}
	}
	return out, nil
}
