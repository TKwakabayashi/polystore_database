package volcano_async_executor

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"

	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/semaphore"
)

// Mode は処理粒度。AsyncMode（並行化方式）とは直交する軸。
type Mode int

const (
	// ModeVolcano は tuple-at-a-time（vectorWidth = 1）。
	ModeVolcano Mode = iota
	// ModeVectorized は batch-at-a-time（vectorWidth = VectorSize）。
	ModeVectorized
)

func (m Mode) String() string {
	if m == ModeVectorized {
		return "Vectorized"
	}
	return "Volcano"
}

// Processor は非同期 pull 実行系の実行コンテキスト。DB 接続・モード・計測を保持する。
//
// 同期版との差分:
//   - asyncMode / policy / sem: 並行化の方式と並行度。
//   - mu: metrics を複数ワーカーから守る（同期版は単一 goroutine 前提で不要だった）。
//   - roundTrips: atomic で加算（metrics.go）。
type Processor struct {
	rg  *storage.Registry
	ctx context.Context

	mode        Mode
	vectorWidth int // Volcano: 1 / Vectorized: VectorSize

	asyncMode AsyncMode
	policy    ExecPolicy
	sem       *semaphore.Weighted // ExecDynamic 用：全演算子で共有する全体上限

	neoDriver neo4j.DriverWithContext
	mDb       *mongo.Database
	ldb       *leveldb.DB
	sqlDb     *sql.DB
	cqlSes    *gocql.Session

	mu         sync.Mutex       // metrics 用
	roundTrips int64            // DB 往復回数（atomic）
	metrics    map[int]*Metrics // step -> 計測
	nextStep   int              // build 時に演算子へ割り当てる連番（build は単一 goroutine）

	results []map[string]interface{}
}

// NewProcessor は cfg で 5 ストアへ接続し、指定モードの Processor を返す。
// vectorSize は ModeVectorized のときのベクトル長（ModeVolcano では無視され 1 になる）。
// policy は並行度。ゼロ値を渡した場合は DefaultExecPolicy() を使う。
func NewProcessor(ctx context.Context, cfg storage.Config, mode Mode, vectorSize int, async AsyncMode, policy ExecPolicy) (*Processor, error) {
	rg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("registry 初期化に失敗: %w", err)
	}

	width := 1
	if mode == ModeVectorized {
		if vectorSize < 1 {
			vectorSize = 1
		}
		width = vectorSize
	}

	if policy.GlobalMaxConcurrency < 1 && len(policy.PerOp) == 0 && policy.Default.Workers < 1 {
		policy = DefaultExecPolicy()
	}

	p := &Processor{
		rg:          rg,
		ctx:         ctx,
		mode:        mode,
		vectorWidth: width,
		asyncMode:   async,
		policy:      policy,
		metrics:     make(map[int]*Metrics),
	}
	if policy.Mode == ExecDynamic {
		p.sem = semaphore.NewWeighted(int64(policy.globalMax()))
	}

	if d, ok := rg.Neo4j(); ok {
		p.neoDriver = d
	}
	if d, ok := rg.Mongo(); ok {
		p.mDb = d
	}
	if d, ok := rg.LevelDB(); ok {
		p.ldb = d
	}
	if d, ok := rg.MySQL(); ok {
		p.sqlDb = d
	}
	if s, ok := rg.Cassandra(); ok {
		p.cqlSes = s
	}
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
	p.mu.Unlock()
	p.roundTrips = 0
	p.nextStep = 0
	p.results = nil
	if p.policy.Mode == ExecDynamic {
		p.sem = semaphore.NewWeighted(int64(p.policy.globalMax()))
	}
}

// Mode は現在の処理粒度を返す。
func (p *Processor) Mode() Mode { return p.mode }

// VectorWidth は現在のベクトル長を返す。
func (p *Processor) VectorWidth() int { return p.vectorWidth }

// AsyncMode は現在の並行化方式を返す。
func (p *Processor) AsyncMode() AsyncMode { return p.asyncMode }

// Policy は現在の並行度ポリシーを返す。
func (p *Processor) Policy() ExecPolicy { return p.policy }

// Run は plan ツリーを pull 実行し、最終結果行を返す。
// tail（Projection/Aggregate/Sort/Limit/Return）は row（map）で、record パイプライン
// （EntityScan/Filter/Expand/VarLengthExpand）は非同期 pull イテレータで実行する。
// Projection がその橋渡し点。
func (p *Processor) Run(op plan.PlanNode) ([]map[string]interface{}, error) {
	if op == nil {
		return nil, fmt.Errorf("nil plan node")
	}
	rows, err := p.execRowLevel(op)
	if err != nil {
		return nil, err
	}
	p.results = rows
	return rows, nil
}

// execRowLevel は row レベルの tail を再帰実行する。Projection に達したらそこで
// record パイプラインを pull 実行してプロパティを実体化し、以降の Return/Sort/Limit/
// Aggregate を上位で適用する。step 番号は子（葉側）→親（根側）の後順で採番する。
//
// Aggregate/Sort/Limit/Return はいずれもメモリ内処理（DB 往復なし）なので、
// 同期版と同一のまま直列で実行する。並行化の対象は Projection より下。
func (p *Processor) execRowLevel(op plan.PlanNode) ([]map[string]interface{}, error) {
	switch o := op.(type) {
	case *plan.Projection:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		if err := child.Open(p.ctx); err != nil {
			return nil, err
		}
		defer child.Close(p.ctx)
		return p.runProjection(o, child)

	case *plan.Aggregate:
		in, err := p.execRowLevel(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applyAggregate(o, in)
		p.nextStep++
		p.recordOp(p.nextStep, "Aggregate", time.Since(start), len(rows))
		return rows, nil

	case *plan.Sort:
		in, err := p.execRowLevel(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applySort(o, in)
		p.nextStep++
		p.recordOp(p.nextStep, "Sort", time.Since(start), len(rows))
		return rows, nil

	case *plan.Limit:
		in, err := p.execRowLevel(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applyLimit(o, in)
		p.nextStep++
		p.recordOp(p.nextStep, "Limit", time.Since(start), len(rows))
		return rows, nil

	case *plan.Return:
		in, err := p.execRowLevel(o.Input)
		if err != nil {
			return nil, err
		}
		start := time.Now()
		rows := applyReturn(o, in)
		p.nextStep++
		p.recordOp(p.nextStep, "Return", time.Since(start), len(rows))
		return rows, nil

	case *plan.StorePushdown:
		return nil, fmt.Errorf("volcano_async_exec: StorePushdown は未対応です（集約 pushdown は stream/bulk を使用してください）")

	default:
		// Projection の無いプラン（RETURN で終わらない補助経路）。
		it, err := p.build(op)
		if err != nil {
			return nil, err
		}
		if err := it.Open(p.ctx); err != nil {
			return nil, err
		}
		defer it.Close(p.ctx)
		return p.drainRaw(it)
	}
}

// build は plan.PlanNode を pull iterator へ変換する（葉→根で step 番号を付与）。
// DB へ往復する Filter/Expand/VarLengthExpand は asyncDriver（exchange）で包み、
// その 1 バッチ処理だけをワーカーで並行実行させる。
// EntityScan は Open で 1 回だけ全件取得し Next は純メモリ処理なので、包まず素通しする。
func (p *Processor) build(op plan.PlanNode) (Iterator, error) {
	switch o := op.(type) {
	case *plan.EntityScan:
		p.nextStep++
		return &scanIterator{p: p, o: o, step: p.nextStep}, nil

	case *plan.Filter:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		p.nextStep++
		fo := &filterOp{p: p, o: o}
		return newAsyncDriver(p, child, OpFilter, p.nextStep, fo.process), nil

	case *plan.Expand:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		p.nextStep++
		eo := &expandOp{p: p, o: o}
		eo.prepare()
		return newAsyncDriver(p, child, OpExpand, p.nextStep, eo.process), nil

	case *plan.VarLengthExpand:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		p.nextStep++
		vo := &varExpandOp{p: p, o: o}
		vo.prepare()
		return newAsyncDriver(p, child, OpVarLengthExpand, p.nextStep, vo.process), nil

	case *plan.Projection:
		return nil, fmt.Errorf("projection は root 以外に置けません")

	default:
		return nil, fmt.Errorf("未知の演算子: %T", op)
	}
}

// drainRaw は Projection の無いプランで、上流の全バッチを引き出し行を map 化する。
// この系のクエリは通常 RETURN(=Projection) で終わるため補助的な経路。
func (p *Processor) drainRaw(it Iterator) ([]map[string]interface{}, error) {
	var out []map[string]interface{}
	for {
		b, err := it.Next(p.ctx)
		if err != nil {
			return nil, err
		}
		if b == nil {
			break
		}
		for i := 0; i < b.n; i++ {
			row := make(map[string]interface{}, b.slotCount())
			for s := 0; s < b.slotCount(); s++ {
				row[fmt.Sprintf("slot%d", s)] = b.get(i, s)
			}
			out = append(out, row)
		}
	}
	return out, nil
}
