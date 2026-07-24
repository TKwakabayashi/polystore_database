package volcano

import (
	"context"
	"fmt"
	"strings"
	"time"

	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
)

// expandIterator は 1 段固定長 Expand の pull 実装。子から 1 バッチ pull し、
// バッチ内 src.uuid をまとめて 1 回の Cypher (WHERE src.uuid IN $ids) で展開する。
//   - Volcano(width=1): 1 行ごとに 1 往復（tuple-at-a-time の往復増幅）。
//   - Vectorized(width=N): 1 バッチごとに 1 往復。
type expandIterator struct {
	p     *Processor
	o     *plan.Expand
	child Iterator
	step  int

	query     string
	srcIdx    int
	relIdxOut int
	tgtIdxOut int
	hasRel    bool
	hasTarget bool
	slotCount int
}

func (e *expandIterator) Open(ctx context.Context) error {
	if err := e.child.Open(ctx); err != nil {
		return err
	}
	o := e.o
	e.srcIdx = o.InputSlot.VarToSlot[o.SourceEntity]
	e.relIdxOut, e.hasRel = o.OutputSlot.VarToSlot[o.Alias]
	e.tgtIdxOut, e.hasTarget = o.OutputSlot.VarToSlot[o.TargetEntity]
	e.slotCount = len(o.OutputSlot.VarToSlot)

	returns := "src.uuid AS sid"
	if e.hasRel {
		returns += fmt.Sprintf(", %s.uuid AS rid", o.Alias)
	}
	if e.hasTarget {
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
	e.query = fmt.Sprintf("MATCH %s WHERE src.uuid IN $ids RETURN %s", pattern, returns)
	return nil
}

func (e *expandIterator) Close(ctx context.Context) error { return e.child.Close(ctx) }

func (e *expandIterator) Next(ctx context.Context) (*Batch, error) {
	for {
		in, err := e.child.Next(ctx)
		if err != nil {
			return nil, err
		}
		if in == nil {
			return nil, nil
		}
		start := time.Now()
		out, err := e.process(in)
		if err != nil {
			return nil, err
		}
		e.p.recordOp(e.step, "Expand", time.Since(start), out.n)
		if out.n > 0 {
			return out, nil
		}
	}
}

func (e *expandIterator) process(in *Batch) (*Batch, error) {
	// src.uuid -> その uuid を持つ入力行（複数あり得る）
	srcIds := make([]string, 0, in.n)
	recordMap := make(map[uid.UUID][][]uid.UUID)
	for i := 0; i < in.n; i++ {
		id := in.get(i, e.srcIdx)
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id.String())
		}
		recordMap[id] = append(recordMap[id], in.row(i))
	}

	sess := e.p.newReadSession()
	defer sess.Close(e.p.ctx)

	e.p.countRoundTrip()
	res, err := sess.Run(e.p.ctx, e.query, map[string]interface{}{"ids": srcIds})
	if err != nil {
		return nil, err
	}

	out := newBatch(e.slotCount, in.n)
	for res.Next(e.p.ctx) {
		dbRec := res.Record()
		sid := uid.FromAny(dbRec.Values[0])
		var rid, tid uid.UUID
		if e.hasRel {
			if v, ok := dbRec.Get("rid"); ok && v != nil {
				rid = uid.FromAny(v)
			}
		}
		if e.hasTarget {
			if v, ok := dbRec.Get("tid"); ok && v != nil {
				tid = uid.FromAny(v)
			}
		}
		for _, origin := range recordMap[sid] {
			newSlots := remap(origin, e.o.InputSlot, e.o.OutputSlot)
			if e.hasRel {
				newSlots[e.relIdxOut] = rid
			}
			if e.hasTarget {
				newSlots[e.tgtIdxOut] = tid
			}
			out.appendRow(newSlots)
		}
	}
	return out, res.Err()
}

// varExpandIterator は可変長 Expand の pull 実装。
type varExpandIterator struct {
	p     *Processor
	o     *plan.VarLengthExpand
	child Iterator
	step  int

	query     string
	srcIdx    int
	tgtIdxOut int
	hasTarget bool
	slotCount int
}

func (e *varExpandIterator) Open(ctx context.Context) error {
	if err := e.child.Open(ctx); err != nil {
		return err
	}
	o := e.o
	e.srcIdx = o.InputSlot.VarToSlot[o.SourceEntity]
	e.tgtIdxOut, e.hasTarget = o.OutputSlot.VarToSlot[o.TargetEntity]
	e.slotCount = len(o.OutputSlot.VarToSlot)

	relLabel := ""
	if o.RelLabel != "" {
		relLabel = ":" + o.RelLabel
	}
	relContent := fmt.Sprintf("[%s%s%s]", o.Alias, relLabel, core.VarLengthRange(o.MinHops, o.MaxHops))
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
	e.query = fmt.Sprintf(`
		MATCH (src:Entity)%s(tgt%s)
		WHERE src.uuid IN $ids
		RETURN DISTINCT src.uuid AS sid, tgt.uuid AS tid`,
		relPattern, tgtConstraint,
	)
	return nil
}

func (e *varExpandIterator) Close(ctx context.Context) error { return e.child.Close(ctx) }

func (e *varExpandIterator) Next(ctx context.Context) (*Batch, error) {
	for {
		in, err := e.child.Next(ctx)
		if err != nil {
			return nil, err
		}
		if in == nil {
			return nil, nil
		}
		start := time.Now()
		out, err := e.process(in)
		if err != nil {
			return nil, err
		}
		e.p.recordOp(e.step, "VarLengthExpand", time.Since(start), out.n)
		if out.n > 0 {
			return out, nil
		}
	}
}

func (e *varExpandIterator) process(in *Batch) (*Batch, error) {
	srcIds := make([]string, 0, in.n)
	recordMap := make(map[uid.UUID][][]uid.UUID)
	for i := 0; i < in.n; i++ {
		id := in.get(i, e.srcIdx)
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id.String())
		}
		recordMap[id] = append(recordMap[id], in.row(i))
	}

	sess := e.p.newReadSession()
	defer sess.Close(e.p.ctx)

	e.p.countRoundTrip()
	res, err := sess.Run(e.p.ctx, e.query, map[string]interface{}{"ids": srcIds})
	if err != nil {
		return nil, err
	}

	out := newBatch(e.slotCount, in.n)
	carry := func(origin []uid.UUID, targetID uid.UUID) {
		newSlots := remap(origin, e.o.InputSlot, e.o.OutputSlot)
		if e.hasTarget {
			newSlots[e.tgtIdxOut] = targetID
		}
		out.appendRow(newSlots)
	}

	reached := make(map[uid.UUID]struct{})
	for res.Next(e.p.ctx) {
		rec := res.Record()
		sid := uid.FromAny(rec.Values[0])
		tid := uid.FromAny(rec.Values[1])
		reached[sid] = struct{}{}
		for _, origin := range recordMap[sid] {
			carry(origin, tid)
		}
	}
	if err := res.Err(); err != nil {
		return nil, err
	}

	// 0 ホップ（自分自身）
	if e.o.MinHops == 0 {
		for _, sidStr := range srcIds {
			sid := uid.UUID(sidStr)
			if _, ok := reached[sid]; ok {
				continue
			}
			for _, origin := range recordMap[sid] {
				carry(origin, sid)
			}
		}
	}
	return out, nil
}
