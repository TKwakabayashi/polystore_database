package vecstream

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"

	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/semaphore"
)

// Processor は pull × 列 Batch × parallel(exchange) 実行系の実行コンテキスト。
// DB 接続・並行度・ベクトル幅・計測を保持する。access_*.go / op_*.go は engine/volcano から
// 移植したため、フィールド名（neoDriver 等）とメソッド名（newReadSession/countRoundTrip）を
// volcano に揃えてある。
type Processor struct {
	rg  *storage.Registry
	ctx context.Context

	exec ExecPolicy
	sem  *semaphore.Weighted // システム全体の同時 DB アクセス上限（全 exchange ワーカーで共有）

	neoDriver neo4j.DriverWithContext
	mDb       *mongo.Database
	ldb       *leveldb.DB
	sqlDb     *sql.DB
	cqlSes    *gocql.Session

	mu         sync.Mutex       // metrics / nextStep 保護
	roundTrips int64            // DB 往復回数（atomic）
	metrics    map[int]*Metrics // step -> 計測
	nextStep   int              // build 時に演算子へ割り当てる連番

	results []map[string]interface{}
}

// NewProcessorWithConfig は cfg で 5 ストアへ接続し Processor を返す。
// 並行度の既定は engine/stream と同一（globalMax=8 / Expand4・Filter4・VarLen2・Projection4）。
// VectorWidth は settings.VectorSize（vectorized と同じ幅ノブ）を反映する。
func NewProcessorWithConfig(ctx context.Context, cfg storage.Config) (*Processor, error) {
	deps, err := core.Open(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("registry 初期化に失敗: %w", err)
	}

	exec := ExecPolicy{
		Default:              OpConcurrency{Workers: 2},
		GlobalMaxConcurrency: 8,
		PerOp: map[OpKind]OpConcurrency{
			OpExpand:          {Workers: 4},
			OpVarLengthExpand: {Workers: 2},
			OpFilter:          {Workers: 4},
			OpProjection:      {Workers: 4},
		},
		VectorWidth: settings.VectorSize,
	}

	p := &Processor{
		rg:      deps.Registry,
		ctx:     ctx,
		exec:    exec,
		sem:     semaphore.NewWeighted(int64(exec.globalMax())),
		metrics: make(map[int]*Metrics),
	}
	p.neoDriver = deps.Neo
	p.mDb = deps.Mongo
	p.ldb = deps.LevelDB
	p.sqlDb = deps.MySQL
	p.cqlSes = deps.Cassandra
	return p, nil
}

// Close は接続を閉じる。
func (p *Processor) Close() error {
	if p.rg == nil {
		return nil
	}
	return p.rg.Close(p.ctx)
}

// Reset は試行間で計測・結果をクリアする（接続は維持）。
func (p *Processor) Reset() {
	p.mu.Lock()
	p.metrics = make(map[int]*Metrics)
	p.nextStep = 0
	p.mu.Unlock()
	p.roundTrips = 0
	p.results = nil
	p.sem = semaphore.NewWeighted(int64(p.exec.globalMax()))
}

// VectorWidth は現在のベクトル幅を返す（core.Result 用）。
func (p *Processor) VectorWidth() int { return p.exec.vectorWidth() }

// newStep は演算子へ step 番号を採番する（build/runRow は単一 goroutine 内で呼ぶ）。
func (p *Processor) newStep() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nextStep++
	return p.nextStep
}

// Run は plan ツリーを実行し最終結果行を返す。
// 現行 IR の root は tail 演算子（Return/Limit/Sort/Aggregate/Projection/StorePushdown）。
func (p *Processor) Run(op plan.PlanNode) ([]map[string]interface{}, error) {
	if op == nil {
		return nil, fmt.Errorf("nil plan node")
	}
	switch op.(type) {
	case *plan.StorePushdown, *plan.Projection, *plan.Aggregate, *plan.Sort, *plan.Limit, *plan.Return:
		rows, err := p.runRow(op)
		if err != nil {
			return nil, err
		}
		p.results = rows
		return rows, nil
	default:
		return nil, fmt.Errorf("vecstream: root は tail 演算子のみ対応（got %T）", op)
	}
}

// build は plan.PlanNode を pull Iterator ツリーへ変換する（葉→根で step 採番）。
// DB へ往復する Filter/Expand/VarLengthExpand は exchangeIterator で包み、W ワーカーが
// Neo4j セッションを 1 本ずつ使い回して process を並列実行する。EntityScan は Open で
// 1 回だけ全件取得し Next は純メモリなので包まず素通しする。
func (p *Processor) build(op plan.PlanNode) (Iterator, error) {
	switch o := op.(type) {
	case *plan.EntityScan:
		step := p.newStep()
		return &scanIterator{p: p, o: o, step: step}, nil

	case *plan.Filter:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		step := p.newStep()
		fo := &filterOp{p: p, o: o}
		return newExchange(p, child, OpFilter, "Filter", step,
			p.newReadSession, p.closeSession, fo.process), nil

	case *plan.Expand:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		step := p.newStep()
		eo := &expandOp{p: p, o: o}
		eo.prepare()
		return newExchange(p, child, OpExpand, "Expand", step,
			p.newReadSession, p.closeSession, eo.process), nil

	case *plan.VarLengthExpand:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		step := p.newStep()
		vo := &varExpandOp{p: p, o: o}
		vo.prepare()
		return newExchange(p, child, OpVarLengthExpand, "VarLengthExpand", step,
			p.newReadSession, p.closeSession, vo.process), nil

	case *plan.Projection:
		return nil, fmt.Errorf("projection は record パイプラインに置けません")

	default:
		return nil, fmt.Errorf("未知の演算子: %T", op)
	}
}
