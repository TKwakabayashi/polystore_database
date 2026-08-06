package volcano

import (
	"context"
	"fmt"
	"time"

	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

// scanIterator は EntityScan の pull 実装。Open で対象ストアから uuid を全件取得し、
// Next で vectorWidth 件ずつバッチとして払い出す（スキャン自体のクエリは 1 往復/ラベル）。
type scanIterator struct {
	p    *Processor
	o    *plan.EntityScan
	step int

	slotCount int
	aliasIdx  int
	ids       []string
	pos       int
}

func (s *scanIterator) Open(ctx context.Context) error {
	start := time.Now()
	s.slotCount = len(s.o.OutputSlot.VarToSlot)
	s.aliasIdx = s.o.OutputSlot.VarToSlot[s.o.Alias]

	ids, err := s.p.scanIDs(s.o)
	if err != nil {
		return err
	}
	s.ids = ids
	s.p.recordOp(s.step, "EntityScan", time.Since(start), 0)
	return nil
}

func (s *scanIterator) Next(ctx context.Context) (*Batch, error) {
	if s.pos >= len(s.ids) {
		return nil, nil
	}
	start := time.Now()
	end := s.pos + s.p.vectorWidth
	if end > len(s.ids) {
		end = len(s.ids)
	}
	b := newBatch(s.slotCount, end-s.pos)
	for ; s.pos < end; s.pos++ {
		row := make([]uid.UUID, s.slotCount)
		row[s.aliasIdx] = uid.UUID(s.ids[s.pos])
		b.appendRow(row)
	}
	s.p.recordOp(s.step, "EntityScan", time.Since(start), b.n)
	return b, nil
}

func (s *scanIterator) Close(ctx context.Context) error { return nil }

// fragmentIterator は record-mode StoreFragment（融合 graph traversal）の pull 実装。
// Open で融合 Cypher を 1 往復実行して束縛 UUID 行をマテリアライズし、Next で vectorWidth 件ずつ払い出す。
type fragmentIterator struct {
	p       *Processor
	cypher  string
	params  map[string]interface{}
	outSlot plan.SlotTable
	step    int

	slotCount int
	rows      [][]string
	pos       int
}

func (it *fragmentIterator) Open(ctx context.Context) error {
	start := time.Now()
	it.slotCount = len(it.outSlot.VarToSlot)
	rows, err := it.p.runGraphRecordFragment(it.cypher, it.params, it.outSlot)
	if err != nil {
		return err
	}
	it.rows = rows
	it.p.recordOp(it.step, "Fragment", time.Since(start), 0)
	return nil
}

func (it *fragmentIterator) Next(ctx context.Context) (*Batch, error) {
	if it.pos >= len(it.rows) {
		return nil, nil
	}
	start := time.Now()
	end := it.pos + it.p.vectorWidth
	if end > len(it.rows) {
		end = len(it.rows)
	}
	b := newBatch(it.slotCount, end-it.pos)
	for ; it.pos < end; it.pos++ {
		r := it.rows[it.pos]
		row := make([]uid.UUID, it.slotCount)
		for s := range row {
			row[s] = uid.UUID(r[s])
		}
		b.appendRow(row)
	}
	it.p.recordOp(it.step, "Fragment", time.Since(start), b.n)
	return b, nil
}

func (it *fragmentIterator) Close(ctx context.Context) error { return nil }

// scanIDs はストア種別に応じて uuid 一覧を取得する（実装は access_<store>.go）。
func (p *Processor) scanIDs(o *plan.EntityScan) ([]string, error) {
	switch o.DataStore {
	case store.Graph:
		return p.scanGraphIDs(o)
	case store.Document:
		return p.scanDocIDs(o)
	case store.Kvs:
		return p.scanKvsIDs(o)
	case store.Relational:
		return p.scanRdbIDs(o)
	case store.Columnar:
		return p.scanColIDs(o)
	default:
		return nil, fmt.Errorf("未知のスキャン対象ストア: %s", o.DataStore)
	}
}
