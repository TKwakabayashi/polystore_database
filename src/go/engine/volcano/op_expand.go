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
	recordMap := make(map[uid.UUID][]int) // src uuid -> 入力行インデックス（carried 値の引継ぎに使う）
	for i := 0; i < in.n; i++ {
		id := in.get(i, e.srcIdx)
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id.String())
		}
		recordMap[id] = append(recordMap[id], i)
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
		for _, idx := range recordMap[sid] {
			newSlots := remap(in.row(idx), e.o.InputSlot, e.o.OutputSlot)
			if e.hasRel {
				newSlots[e.relIdxOut] = rid
			}
			if e.hasTarget {
				newSlots[e.tgtIdxOut] = tid
			}
			out.appendRowVals(newSlots, in.valsAt(idx))
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

	query      string
	pathParams map[string]interface{} // パス全体フィルタ用の静的パラメータ
	srcIdx     int
	tgtIdxOut  int
	hasTarget  bool
	slotCount  int
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

	// 可変長パス全体への述語（all/any/none/single・length(p)）を WHERE に付加する。
	matchHead := "(src:Entity)"
	whereExtra := ""
	e.pathParams = map[string]interface{}{}
	if len(o.Filters) > 0 {
		const pathVar = "vp"
		matchHead = pathVar + " = (src:Entity)"
		seq := 0
		clauses, err := core.BuildPathFilterClauses(pathVar, o.Filters, e.pathParams, &seq)
		if err != nil {
			return err
		}
		if len(clauses) > 0 {
			whereExtra = " AND " + strings.Join(clauses, " AND ")
		}
	}

	e.query = fmt.Sprintf(`
		MATCH %s%s(tgt%s)
		WHERE src.uuid IN $ids%s
		RETURN DISTINCT src.uuid AS sid, tgt.uuid AS tid`,
		matchHead, relPattern, tgtConstraint, whereExtra,
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
	recordMap := make(map[uid.UUID][]int) // src uuid -> 入力行インデックス（carried 値の引継ぎに使う）
	for i := 0; i < in.n; i++ {
		id := in.get(i, e.srcIdx)
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id.String())
		}
		recordMap[id] = append(recordMap[id], i)
	}

	sess := e.p.newReadSession()
	defer sess.Close(e.p.ctx)

	runParams := make(map[string]interface{}, len(e.pathParams)+1)
	runParams["ids"] = srcIds
	for k, v := range e.pathParams {
		runParams[k] = v
	}

	e.p.countRoundTrip()
	res, err := sess.Run(e.p.ctx, e.query, runParams)
	if err != nil {
		return nil, err
	}

	out := newBatch(e.slotCount, in.n)
	carry := func(idx int, targetID uid.UUID) {
		newSlots := remap(in.row(idx), e.o.InputSlot, e.o.OutputSlot)
		if e.hasTarget {
			newSlots[e.tgtIdxOut] = targetID
		}
		out.appendRowVals(newSlots, in.valsAt(idx))
	}

	reached := make(map[uid.UUID]struct{})
	for res.Next(e.p.ctx) {
		rec := res.Record()
		sid := uid.FromAny(rec.Values[0])
		tid := uid.FromAny(rec.Values[1])
		reached[sid] = struct{}{}
		for _, idx := range recordMap[sid] {
			carry(idx, tid)
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
			for _, idx := range recordMap[sid] {
				carry(idx, sid)
			}
		}
	}
	return out, nil
}
