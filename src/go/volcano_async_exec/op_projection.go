package volcano_async_executor

import (
	"context"
	"sync"
	"time"

	"polystore_database/src/go/plan"
)

// runProjection は Projection の実体化。子から全バッチを pull しつつ、バッチ単位で
// プロパティをクロスストア収集し、wide row（キー: 束縛 uuid の "alias" と "alias.prop"）を
// 生成して返す。整形(Return)・ソート(Sort)・リミット(Limit) は上位の tail 演算子が行う。
//
// Projection もバッチごとに DB へ往復する（fetchPropertiesBulk）ため、ここも並行化の
// 対象になる。Volcano では射影のプロパティ取得も tuple-at-a-time となり、
// 往復数は ⌈rows/vectorWidth⌉ × Fetch 数（ストア毎）に比例する。
//
// 構造は asyncDriver と同じ（child は直列 pull / fetch だけ W ワーカーで並行）だが、
// 出力が Batch ではなく []Row のため Iterator にはせず、ここに直接書いている。
//
// 順序: ワーカーの完了順で allResults へ積むため、行順は非決定的になる。
// ORDER BY は上位の Sort が担保するので結果の正しさには影響しない
// （ORDER BY の無いクエリの行順が実行ごとに変わる点は stream_exec と同じ性質）。
func (p *Processor) runProjection(o *plan.Projection, child Iterator) ([]map[string]interface{}, error) {
	step := p.nextStep + 1
	p.nextStep = step

	workers := p.policy.workersFor(OpProjection, p.asyncMode)

	ctx, cancel := context.WithCancel(p.ctx)
	defer cancel()

	work := make(chan *Batch, workers)

	var (
		mu         sync.Mutex
		allResults []map[string]interface{}
		errMu      sync.Mutex
		firstErr   error
	)
	setErr := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for b := range work {
				if p.sem != nil {
					if err := p.sem.Acquire(ctx, 1); err != nil {
						return
					}
				}
				start := time.Now()
				rows := p.projectBatch(o, b)
				if p.sem != nil {
					p.sem.Release(1)
				}
				p.recordOp(step, "Projection", time.Since(start), b.n)

				mu.Lock()
				allResults = append(allResults, rows...)
				mu.Unlock()
			}
		}()
	}

	// driver: child から直列に pull して work へ配る。
drive:
	for {
		b, err := child.Next(ctx)
		if err != nil {
			setErr(err)
			break
		}
		if b == nil {
			break // EOF
		}
		select {
		case work <- b:
		case <-ctx.Done():
			break drive
		}
	}
	close(work)
	wg.Wait()

	errMu.Lock()
	err := firstErr
	errMu.Unlock()
	if err != nil {
		return nil, err
	}
	return allResults, nil
}

// projectBatch は 1 バッチ分のプロパティ収集と wide row 生成。
// Processor の共有状態を書き換えないため、複数ワーカーから同時に呼んで安全
// （計測は recordOp が mu で、往復数は countRoundTrip が atomic で保護する）。
func (p *Processor) projectBatch(o *plan.Projection, batch *Batch) []map[string]interface{} {
	aliasToSlot := o.InputSlot.VarToSlot

	// A. バッチ内の Unit ごとの ID 集合
	unitIDMap := make(map[string]map[string]struct{})
	for _, unit := range o.Units {
		unitIDMap[unit.Alias] = make(map[string]struct{})
	}
	for i := 0; i < batch.n; i++ {
		for _, unit := range o.Units {
			slotIdx, ok := aliasToSlot[unit.Alias]
			if !ok {
				continue
			}
			id := batch.get(i, slotIdx)
			if id != "" {
				unitIDMap[unit.Alias][id] = struct{}{}
			}
		}
	}

	// B. プロパティフェッチ cache[alias][id][prop] = value
	cache := make(map[string]map[string]map[string]interface{})
	for ui := range o.Units {
		unit := o.Units[ui]
		idSet := unitIDMap[unit.Alias]
		if len(idSet) == 0 {
			continue
		}
		ids := make([]string, 0, len(idSet))
		for id := range idSet {
			ids = append(ids, id)
		}
		cache[unit.Alias] = make(map[string]map[string]interface{})
		for i := range unit.Fetches {
			data := p.fetchPropertiesBulk(ids, &unit, &unit.Fetches[i])
			for id, propsMap := range data {
				if _, ok := cache[unit.Alias][id]; !ok {
					cache[unit.Alias][id] = make(map[string]interface{})
				}
				for pName, pVal := range propsMap {
					cache[unit.Alias][id][pName] = pVal
				}
			}
		}
	}

	// C. wide row 生成（キー: "alias"=束縛uuid, "alias.prop"=値）
	out := make([]map[string]interface{}, 0, batch.n)
	for i := 0; i < batch.n; i++ {
		row := make(map[string]interface{})
		for alias, slotIdx := range aliasToSlot {
			row[alias] = batch.get(i, slotIdx) // 束縛 uuid
		}
		for _, unit := range o.Units {
			slotIdx, ok := aliasToSlot[unit.Alias]
			if !ok {
				continue
			}
			props := cache[unit.Alias][batch.get(i, slotIdx)]
			for _, f := range unit.Fetches {
				for _, pName := range f.Props {
					row[unit.Alias+"."+pName] = props[pName]
				}
			}
		}
		out = append(out, row)
	}
	return out
}

// fetchPropertiesBulk は各種ストアからプロパティを一括取得する共通の入り口（実装は access_<store>.go）。
func (p *Processor) fetchPropertiesBulk(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	switch fetch.Store {
	case "graph":
		return p.fetchGraphProps(ids, unit, fetch)
	case "document":
		return p.fetchDocProps(ids, unit, fetch)
	case "kvs":
		return p.fetchKvsProps(ids, unit, fetch)
	case "columnar":
		return p.fetchColProps(ids, unit, fetch)
	case "relational":
		return p.fetchRdbProps(ids, unit, fetch)
	default:
		return nil
	}
}
