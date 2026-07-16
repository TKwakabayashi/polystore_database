package volcano_async_executor

import (
	"fmt"
	"strings"

	"polystore_database/src/go/plan"
)

// expandOp は 1 段固定長 Expand の 1 バッチ処理。バッチ内 src.uuid をまとめて
// 1 回の Cypher (WHERE src.uuid IN $ids) で展開する。
//
//	Volcano(width=1)    : 1 行 = 1 往復（tuple-at-a-time の往復増幅）
//	Vectorized(width=N) : 1 バッチ = 1 往復（往復 = ⌈rows/N⌉）
//
// 同期版と違い、これらの往復は asyncDriver の W ワーカーによって並行に発行される。
// クエリ本体・remap は同期版 (volcano_exec/op_expand.go) と同一。
//
// query は prepare() で 1 度だけ組み立て、以降は読み取り専用なので
// 複数ワーカーから同時に process を呼んで安全。
type expandOp struct {
	p *Processor
	o *plan.Expand

	query     string
	srcIdx    int
	relIdxOut int
	tgtIdxOut int
	hasRel    bool
	hasTarget bool
	slotCount int
}

// prepare は build 時（単一 goroutine）に Cypher とスロット位置を確定させる。
func (e *expandOp) prepare() {
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
}

func (e *expandOp) process(in *Batch) (*Batch, error) {
	// src.uuid -> その uuid を持つ入力行（複数あり得る）
	srcIds := make([]string, 0, in.n)
	recordMap := make(map[string][][]string)
	for i := 0; i < in.n; i++ {
		id := in.get(i, e.srcIdx)
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id)
		}
		recordMap[id] = append(recordMap[id], in.row(i))
	}

	// セッションはワーカーごとに都度作る（neo4j の Session は goroutine 安全でないため、
	// 共有せず process のスコープに閉じ込める）。
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
		sid, ok := dbRec.Values[0].(string)
		if !ok {
			continue
		}
		var rid, tid string
		if e.hasRel {
			if v, ok := dbRec.Get("rid"); ok && v != nil {
				rid, _ = v.(string)
			}
		}
		if e.hasTarget {
			if v, ok := dbRec.Get("tid"); ok && v != nil {
				tid, _ = v.(string)
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

// varExpandOp は可変長 Expand の 1 バッチ処理。
type varExpandOp struct {
	p *Processor
	o *plan.VarLengthExpand

	query     string
	srcIdx    int
	tgtIdxOut int
	hasTarget bool
	slotCount int
}

func (e *varExpandOp) prepare() {
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
	e.query = fmt.Sprintf(`
		MATCH (src:Entity)%s(tgt%s)
		WHERE src.uuid IN $ids
		RETURN DISTINCT src.uuid AS sid, tgt.uuid AS tid`,
		relPattern, tgtConstraint,
	)
}

func (e *varExpandOp) process(in *Batch) (*Batch, error) {
	srcIds := make([]string, 0, in.n)
	recordMap := make(map[string][][]string)
	for i := 0; i < in.n; i++ {
		id := in.get(i, e.srcIdx)
		if _, exists := recordMap[id]; !exists {
			srcIds = append(srcIds, id)
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
	carry := func(origin []string, targetID string) {
		newSlots := remap(origin, e.o.InputSlot, e.o.OutputSlot)
		if e.hasTarget {
			newSlots[e.tgtIdxOut] = targetID
		}
		out.appendRow(newSlots)
	}

	reached := make(map[string]struct{})
	for res.Next(e.p.ctx) {
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
	if e.o.MinHops == 0 {
		for _, sid := range srcIds {
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
