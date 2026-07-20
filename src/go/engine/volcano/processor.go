package volcano

import (
	"context"
	"database/sql"
	"fmt"

	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"

	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
)

// Mode は実行モデル。
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

// Processor は pull 型実行系の実行コンテキスト。DB 接続・モード・計測を保持する。
type Processor struct {
	rg  *storage.Registry
	ctx context.Context

	mode        Mode
	vectorWidth int // Volcano: 1 / Vectorized: VectorSize

	neoDriver neo4j.DriverWithContext
	mDb       *mongo.Database
	ldb       *leveldb.DB
	sqlDb     *sql.DB
	cqlSes    *gocql.Session

	roundTrips int64            // DB 往復回数
	metrics    map[int]*Metrics // step -> 計測
	nextStep   int              // build 時に演算子へ割り当てる連番

	results []map[string]interface{}
}

// NewProcessor は cfg で 5 ストアへ接続し、指定モードの Processor を返す。
// vectorSize は ModeVectorized のときのベクトル長（ModeVolcano では無視され 1 になる）。
func NewProcessor(ctx context.Context, cfg storage.Config, mode Mode, vectorSize int) (*Processor, error) {
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

	p := &Processor{
		rg:          rg,
		ctx:         ctx,
		mode:        mode,
		vectorWidth: width,
		metrics:     make(map[int]*Metrics),
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
	p.roundTrips = 0
	p.metrics = make(map[int]*Metrics)
	p.nextStep = 0
	p.results = nil
}

// Mode は現在のモードを返す。
func (p *Processor) Mode() Mode { return p.mode }

// VectorWidth は現在のベクトル長を返す。
func (p *Processor) VectorWidth() int { return p.vectorWidth }

// newStep は演算子へ step 番号を採番する（葉→根で加算）。
func (p *Processor) newStep() int {
	p.nextStep++
	return p.nextStep
}

// Run は plan ツリーを pull 実行し、最終結果行を返す。
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
		// tail の無いプラン（補助経路）。
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
		return &filterIterator{p: p, o: o, child: child, step: p.nextStep}, nil
	case *plan.Expand:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		p.nextStep++
		return &expandIterator{p: p, o: o, child: child, step: p.nextStep}, nil
	case *plan.VarLengthExpand:
		child, err := p.build(o.Input)
		if err != nil {
			return nil, err
		}
		p.nextStep++
		return &varExpandIterator{p: p, o: o, child: child, step: p.nextStep}, nil
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
