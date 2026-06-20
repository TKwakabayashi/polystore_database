package stream_executor

import (
	"fmt"
	"polystore_database/src/go/plan"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
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

func ExpandGraphStream(qp *QueryProcessor, o *plan.Expand, inputStream <-chan []Record, outputStream chan<- []Record) (int, error) {
	const batchSize = 500
	totalRowCount := 0

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

	// 3. 【メインループ】バッチ処理
	const workerCount = 1
	var wg sync.WaitGroup
	var atomicTotalCount int64

	// バッチをWorkerに渡すための内部チャネル
	batchChan := make(chan []Record, workerCount)

	// 1. Worker ゴルーチンの起動
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				// --- A. IDの整理と逆引き用Map (Worker内で行う) ---
				srcIds := make([]string, 0, len(batch))
				recordMap := make(map[string][]Record)
				for _, r := range batch {
					id := r.Slots[srcIdx]
					if _, exists := recordMap[id]; !exists {
						srcIds = append(srcIds, id)
					}
					recordMap[id] = append(recordMap[id], r)
				}

				sess := qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})

				// --- B. 実行 (並列でNeo4jにリクエスト) ---
				res, err := sess.Run(qp.ctx, finalQuery, map[string]interface{}{"ids": srcIds})
				if err != nil {
					// エラーログ出力など
					continue
				}

				localCount := 0
				// --- C. 結果の変換と送信 ---
				resultsBuffer := make([]Record, 0, batchSize)
				for res.Next(qp.ctx) {
					dbRec := res.Record()
					sidStr := dbRec.Values[0].(string)

					for _, originalRec := range recordMap[sidStr] {
						newSlots := make([]string, newSlotCount)

						// 1. 既存列の引き継ぎ (inputSlot -> outputSlot へのマッピング)
						for alias, outIdx := range o.OutputSlot.VarToSlot {
							if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
								newSlots[outIdx] = originalRec.Slots[inIdx]
							}
						}

						// 2. 新規追加列 (Relation/Target) の書き込み
						if hasRel {
							if rid, ok := dbRec.Get("rid"); ok && rid != nil {
								newSlots[relIdxOut] = rid.(string)
							}
						}
						if hasTarget {
							if tid, ok := dbRec.Get("tid"); ok && tid != nil {
								newSlots[tgtIdxOut] = tid.(string)
							}
						}
						resultsBuffer = append(resultsBuffer, Record{Slots: newSlots})

						// バッファがいっぱいになったら「塊」として送信
						if len(resultsBuffer) >= batchSize {
							outputStream <- resultsBuffer
							resultsBuffer = make([]Record, 0, batchSize) // 新しいスライスを確保
						}

						localCount++
					}
				}

				if len(resultsBuffer) > 0 {
					outputStream <- resultsBuffer
				}
				atomic.AddInt64(&atomicTotalCount, int64(localCount))

				sess.Close(qp.ctx)
			}
		}()
	}

	// 2. メインルーチン：入力からバッチを切り出して batchChan に投げる
	for batch := range inputStream {
		batchChan <- batch
	}

	// 3. 後処理：全ての投入が終わったらチャネルを閉じ、Workerの完了を待つ
	close(batchChan)
	wg.Wait()

	totalRowCount = int(atomicTotalCount)

	return totalRowCount, nil
}

func streamVarLengthExpand(qp *QueryProcessor,
	o *plan.VarLengthExpand, inputStream <-chan []Record, outputStream chan<- []Record) (int, error) {
	const batchSize = 100
	const outputBatchSize = 500
	const workerCount = 1
	var wg sync.WaitGroup
	var atomicTotalCount int64

	srcIdxIn := o.InputSlot.VarToSlot[o.SourceEntity]

	// 出力のどこにターゲットIDを配置するか
	tgtIdxOut, hasTarget := o.OutputSlot.VarToSlot[o.TargetEntity]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	// 2. 【ループ外】クエリテンプレートの構築
	relLabel := ""
	if o.RelLabel != "" {
		relLabel = ":" + o.RelLabel
	}

	// 可変長パスの指定 (*min..max)
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

	// 3. Worker Pool の起動
	batchChan := make(chan []Record, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				srcIds := make([]string, 0, len(batch))
				recordMap := make(map[string][]Record)
				for _, r := range batch {
					id := r.Slots[srcIdxIn]
					if _, exists := recordMap[id]; !exists {
						srcIds = append(srcIds, id)
					}
					recordMap[id] = append(recordMap[id], r)
				}

				// Neo4j 実行
				sess := qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
				defer sess.Close(qp.ctx)

				res, err := sess.Run(qp.ctx, finalQuery, map[string]interface{}{"ids": srcIds})
				if err != nil {
					continue
				}

				localCount := 0
				resultsBuffer := make([]Record, 0, outputBatchSize)
				reachedSids := make(map[string]struct{})

				for res.Next(qp.ctx) {
					rec := res.Record()
					sid := rec.Values[0].(string)
					tid := rec.Values[1].(string)
					reachedSids[sid] = struct{}{}

					for _, originalRec := range recordMap[sid] {
						newSlots := make([]string, newSlotCount)
						// 1. 既存列の引き継ぎ (inputSlot -> outputSlot へのマッピング)
						for alias, outIdx := range o.OutputSlot.VarToSlot {
							if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
								newSlots[outIdx] = originalRec.Slots[inIdx]
							}
						}
						if hasTarget {
							newSlots[tgtIdxOut] = tid
						}

						resultsBuffer = append(resultsBuffer, Record{Slots: newSlots})
						localCount++

						if len(resultsBuffer) >= outputBatchSize {
							outputStream <- resultsBuffer
							resultsBuffer = make([]Record, 0, outputBatchSize)
						}
					}
				}

				// 0ホップ（自分自身）の処理
				if o.MinHops == 0 {
					for _, sid := range srcIds {
						if _, ok := reachedSids[sid]; !ok {
							for _, originalRec := range recordMap[sid] {
								newSlots := make([]string, newSlotCount)
								for alias, outIdx := range o.OutputSlot.VarToSlot {
									if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
										newSlots[outIdx] = originalRec.Slots[inIdx]
									}
								}
								if hasTarget {
									newSlots[tgtIdxOut] = sid
								}
								resultsBuffer = append(resultsBuffer, Record{Slots: newSlots})
								localCount++

								if len(resultsBuffer) >= outputBatchSize {
									outputStream <- resultsBuffer
									resultsBuffer = make([]Record, 0, outputBatchSize)
								}
							}
						}
					}
				}

				if len(resultsBuffer) > 0 {
					outputStream <- resultsBuffer
				}
				atomic.AddInt64(&atomicTotalCount, int64(localCount))

				sess.Close(qp.ctx)
			}
		}()
	}

	// 4. メインループ
	for batch := range inputStream {
		batchChan <- batch
	}

	close(batchChan)
	wg.Wait()
	return int(atomicTotalCount), nil
}
