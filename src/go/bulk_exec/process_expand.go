package bulk_executor

import (
	"fmt"
	"strings"
	"time"

	"polystore_database/src/go/plan"
)

// execExpand は 1 段固定長 Expand の全件実装。子を全件実行し、全入力行の src.uuid を
// まとめて 1 回の Cypher (WHERE src.uuid IN $ids) で展開する。
func (p *Processor) execExpand(o *plan.Expand) ([]Record, error) {
	in, err := p.execute(o.Input)
	if err != nil {
		return nil, err
	}
	p.nextStep++
	step := p.nextStep
	start := time.Now()

	srcIdx := o.InputSlot.VarToSlot[o.SourceEntity]
	relIdxOut, hasRel := o.OutputSlot.VarToSlot[o.Alias]
	tgtIdxOut, hasTarget := o.OutputSlot.VarToSlot[o.TargetEntity]

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
	query := fmt.Sprintf("MATCH %s WHERE src.uuid IN $ids RETURN %s", pattern, returns)

	// src.uuid -> その uuid を持つ入力行（複数あり得る）
	srcIds := make([]string, 0, len(in))
	recordMap := make(map[string][][]string)
	for _, r := range in {
		id := r[srcIdx]
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id)
		}
		recordMap[id] = append(recordMap[id], []string(r))
	}

	sess := p.newReadSession()
	defer sess.Close(p.ctx)

	p.countRoundTrip()
	res, err := sess.Run(p.ctx, query, map[string]interface{}{"ids": srcIds})
	if err != nil {
		return nil, err
	}

	out := make([]Record, 0, len(in))
	for res.Next(p.ctx) {
		dbRec := res.Record()
		sid, ok := dbRec.Values[0].(string)
		if !ok {
			continue
		}
		var rid, tid string
		if hasRel {
			if v, ok := dbRec.Get("rid"); ok && v != nil {
				rid, _ = v.(string)
			}
		}
		if hasTarget {
			if v, ok := dbRec.Get("tid"); ok && v != nil {
				tid, _ = v.(string)
			}
		}
		for _, origin := range recordMap[sid] {
			newSlots := remap(origin, o.InputSlot, o.OutputSlot)
			if hasRel {
				newSlots[relIdxOut] = rid
			}
			if hasTarget {
				newSlots[tgtIdxOut] = tid
			}
			out = append(out, Record(newSlots))
		}
	}
	if err := res.Err(); err != nil {
		return nil, err
	}
	p.recordOp(step, "Expand", time.Since(start), len(in), len(out))
	return out, nil
}

// execVarExpand は可変長 Expand の全件実装。
func (p *Processor) execVarExpand(o *plan.VarLengthExpand) ([]Record, error) {
	in, err := p.execute(o.Input)
	if err != nil {
		return nil, err
	}
	p.nextStep++
	step := p.nextStep
	start := time.Now()

	srcIdx := o.InputSlot.VarToSlot[o.SourceEntity]
	tgtIdxOut, hasTarget := o.OutputSlot.VarToSlot[o.TargetEntity]

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
	query := fmt.Sprintf(`
		MATCH (src:Entity)%s(tgt%s)
		WHERE src.uuid IN $ids
		RETURN DISTINCT src.uuid AS sid, tgt.uuid AS tid`,
		relPattern, tgtConstraint,
	)

	srcIds := make([]string, 0, len(in))
	recordMap := make(map[string][][]string)
	for _, r := range in {
		id := r[srcIdx]
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id)
		}
		recordMap[id] = append(recordMap[id], []string(r))
	}

	sess := p.newReadSession()
	defer sess.Close(p.ctx)

	p.countRoundTrip()
	res, err := sess.Run(p.ctx, query, map[string]interface{}{"ids": srcIds})
	if err != nil {
		return nil, err
	}

	out := make([]Record, 0, len(in))
	carry := func(origin []string, targetID string) {
		newSlots := remap(origin, o.InputSlot, o.OutputSlot)
		if hasTarget {
			newSlots[tgtIdxOut] = targetID
		}
		out = append(out, Record(newSlots))
	}

	reached := make(map[string]struct{})
	for res.Next(p.ctx) {
		rec := res.Record()
		sid, _ := rec.Values[0].(string)
		tid, _ := rec.Values[1].(string)
		reached[sid] = struct{}{}
		for _, origin := range recordMap[sid] {
			carry(origin, tid)
		}
	}
	if err := res.Err(); err != nil {
		return nil, err
	}

	// 0 ホップ（自分自身）
	if o.MinHops == 0 {
		for _, sid := range srcIds {
			if _, ok := reached[sid]; ok {
				continue
			}
			for _, origin := range recordMap[sid] {
				carry(origin, sid)
			}
		}
	}
	p.recordOp(step, "VarLengthExpand", time.Since(start), len(in), len(out))
	return out, nil
}
