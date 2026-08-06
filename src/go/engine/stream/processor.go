package stream

import (
	"context"
	"database/sql"
	"fmt"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/store"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/semaphore"
)

func noResource() struct{}       { return struct{}{} }
func closeNoResource(_ struct{}) {}

// Record は core.Record への型エイリアス（Slots []id.UUID を bulk と共有）。
type Record = core.Record

type Row = map[string]interface{}

type Processor struct {
	records   []Record
	slotTable *plan.SlotTable
	results   []map[string]interface{}
	instr     *core.Instr // 往復・演算子時間・フロー計測（vecstream と共有する core 実装）

	neoDriver neo4j.DriverWithContext
	// neoSes    neo4j.SessionWithContext
	mDb    *mongo.Database
	ldb    *leveldb.DB
	sqlDb  *sql.DB
	cqlSes *gocql.Session
	ctx    context.Context

	rg   *storage.Registry
	exec ExecPolicy
	sem  *semaphore.Weighted // ExecDynamic 用：全演算子で共有する全体上限
}

// 計測は core.Instr へ委譲する薄いラッパ（vecstream と同一意味論）。
func (qp *Processor) recordOp(step int, op string, dur time.Duration, rows int) {
	qp.instr.RecordOp(step, op, dur, rows)
}
func (qp *Processor) recordFlow(step int, op string, batIn, batOut, rowIn, rowOut, queries int64, t0, t1 time.Time) {
	qp.instr.RecordFlow(step, op, batIn, batOut, rowIn, rowOut, queries, t0, t1)
}
func (qp *Processor) countRoundTrip() { qp.instr.CountRoundTrip() }

// recordScan は 1 scan の計測を記録する（クエリ1往復・emit batches・rowOut・wall）。
// scan は VectorWidth 幅で払い出すため batOut = ⌈rows/VectorWidth⌉。
func (qp *Processor) recordScan(step, rows int, t0 time.Time) {
	t1 := time.Now()
	qp.recordOp(step, "EntityScan", t1.Sub(t0), rows)
	vw := qp.exec.vectorWidth()
	batches := 0
	if rows > 0 {
		batches = (rows + vw - 1) / vw
	}
	qp.recordFlow(step, "EntityScan", 0, int64(batches), 0, int64(rows), 1, t0, t1)
}
func (qp *Processor) RoundTrips() int64              { return qp.instr.RoundTrips() }
func (qp *Processor) StepMetrics() []core.StepMetric { return qp.instr.StepMetrics() }
func (qp *Processor) FlowMetrics() []core.FlowMetric { return qp.instr.FlowMetrics() }
func (qp *Processor) VectorWidth() int               { return qp.exec.vectorWidth() }

func NewProcessor(ctx context.Context) (*Processor, error) {
	st := &plan.SlotTable{
		VarToSlot: make(map[string]int),
		SlotToVar: []string{},
	}

	qp := &Processor{
		records:   []Record{},
		slotTable: st,
		results:   []map[string]interface{}{},
		instr:     core.NewInstr(),
		ctx:       ctx,
		exec:      ExecPolicy{VectorWidth: settings.VectorSize},
	}
	cfg, _ := storage.LoadConfig("")
	deps, err := core.Open(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	qp.rg = deps.Registry
	qp.neoDriver = deps.Neo
	qp.mDb = deps.Mongo
	qp.ldb = deps.LevelDB
	qp.sqlDb = deps.MySQL
	qp.cqlSes = deps.Cassandra

	return qp, nil
}

func NewProcessorWithConfig(ctx context.Context, cfg storage.Config) (*Processor, error) {
	st := &plan.SlotTable{
		VarToSlot: make(map[string]int),
		SlotToVar: []string{},
	}

	exec := ExecPolicy{
		Mode:                 ExecDynamic,
		Default:              OpConcurrency{Workers: 2},
		GlobalMaxConcurrency: 8, // ExecDynamic 時のシステム全体の同時DB上限
		PerOp: map[OpKind]OpConcurrency{
			OpExpand:          {Workers: 4},
			OpVarLengthExpand: {Workers: 2},
			OpFilter:          {Workers: 4},
			OpProjection:      {Workers: 4},
		},
		VectorWidth: settings.VectorSize, // vecstream と同じ batch 幅ノブ
	}

	qp := &Processor{
		records:   []Record{},
		slotTable: st,
		results:   []map[string]interface{}{},
		instr:     core.NewInstr(),
		ctx:       ctx,
		exec:      exec,
		sem:       semaphore.NewWeighted(int64(exec.globalMax())),
	}

	deps, err := core.Open(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	qp.rg = deps.Registry
	qp.neoDriver = deps.Neo
	qp.mDb = deps.Mongo
	qp.ldb = deps.LevelDB
	qp.sqlDb = deps.MySQL
	qp.cqlSes = deps.Cassandra

	return qp, nil
}

func (qp *Processor) Close() error {
	if qp.rg == nil {
		return nil
	}
	return qp.rg.Close(qp.ctx)
}

func (qp *Processor) Reset() {
	// 中間レコードのクリア
	qp.records = []Record{}

	// スロットテーブルの初期化
	qp.slotTable = &plan.SlotTable{
		VarToSlot: make(map[string]int),
		SlotToVar: []string{},
	}

	// 最終結果をクリア
	qp.results = []map[string]interface{}{}

	// 計測のクリア（往復・演算子時間・フロー）
	qp.instr.Reset()
}

func (qp *Processor) ProcessQueryStream(op plan.PlanNode) ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	counter := 0
	rowCh, err := executeRowStream(qp, op, &counter, &wg)
	if err != nil {
		wg.Wait()
		return nil, err
	}
	var results []map[string]interface{}
	if rowCh != nil {
		for batch := range rowCh {
			results = append(results, batch...)
		}
	}
	wg.Wait()
	qp.results = results
	return qp.results, nil
}

func ExecuteOperatorStream(qp *Processor, op plan.PlanNode, counter *int, wg *sync.WaitGroup) (chan []Record, error) {
	if op == nil {
		return nil, fmt.Errorf("Empty Operator Passed")
	}

	// record-mode StoreFragment（部分融合）: graph traversal を 1 本の Cypher に融合して束縛 UUID を
	// source として流す。生成不能な構造は元の部分木を通常実行してフォールバック（結果は等価）。
	if frag, ok := op.(*plan.StoreFragment); ok {
		aliases := make([]string, 0, len(frag.OutputSlot.VarToSlot))
		for a := range frag.OutputSlot.VarToSlot {
			aliases = append(aliases, a)
		}
		cypher, params := core.BuildGraphRecordCypher(frag.Ops, aliases)
		if cypher == "" {
			return ExecuteOperatorStream(qp, frag.Ops, counter, wg)
		}
		outputStream := make(chan []Record, 500)
		*counter++
		step := *counter
		outSlot := frag.OutputSlot
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(outputStream)
			if _, err := streamGraphRecordFragment(qp, cypher, params, outSlot, step, outputStream); err != nil {
				fmt.Printf("Error in step %d (Fragment): %v\n", step, err)
			}
		}()
		return outputStream, nil
	}

	// 1. 再帰的に上流（Child）のチャネルを取得
	var inputStream chan []Record
	if len(op.Children()) > 0 {
		var err error
		// ここでは inputStream を受け取る
		inputStream, err = ExecuteOperatorStream(qp, op.Children()[0], counter, wg)
		if err != nil {
			return nil, err
		}
	}

	// 2. この演算子の出力チャネルを作成　チャネルのサイズを調整する必要あり
	outputStream := make(chan []Record, 500)

	*counter++
	currentStep := *counter
	// 3. 演算子の実行 (ゴルーチンによる非同期処理)
	wg.Add(1)
	go func() {
		// 次のステージへ終了を伝えるために必ず閉じる
		defer wg.Done()
		defer close(outputStream)

		var err error
		// 演算子計測（Duration/フロー）は各 op 内（runBatches / scan）で currentStep に記録する。
		switch o := op.(type) {
		case *plan.EntityScan:
			_, err = scanByStore(qp, o, currentStep, outputStream)
		case *plan.Expand:
			_, err = ExpandGraphStream(qp, o, currentStep, inputStream, outputStream)
		case *plan.VarLengthExpand:
			_, err = streamVarLengthExpand(qp, o, currentStep, inputStream, outputStream)
		case *plan.Filter:
			_, err = filterByStore(qp, o, currentStep, inputStream, outputStream)
		default:
			fmt.Printf("Unknown operator: %T\n", o)
		}
		if err != nil {
			fmt.Printf("Error in step %d: %v\n", currentStep, err)
		}
	}()

	return outputStream, nil
}

func scanByStore(qp *Processor, o *plan.EntityScan, step int, out chan<- []Record) (int, error) {
	switch o.DataStore {
	case store.Graph:
		return ScanGraphStream(qp, o, step, out)
	case store.Document:
		return ScanDocStream(qp, o, step, out)
	case store.Kvs:
		return ScanKvsStream(qp, o, step, out)
	case store.Relational:
		return ScanRdbStream(qp, o, step, out)
	case store.Columnar:
		return ScanColStream(qp, o, step, out)
	default:
		return 0, fmt.Errorf("unknown datastore for scan: %s", o.DataStore)
	}
}

func filterByStore(qp *Processor, o *plan.Filter, step int, in <-chan []Record, out chan<- []Record) (int, error) {
	switch o.DataStore {
	case store.Graph:
		return streamFilterGraph(qp, o, step, in, out)
	case store.Document:
		return FilterDocStream(qp, o, step, in, out)
	case store.Kvs:
		return FilterKvsStream(qp, o, step, in, out)
	case store.Relational:
		return FilterRdbStream(qp, o, step, in, out)
	case store.Columnar:
		return FilterColStream(qp, o, step, in, out)
	default:
		return 0, fmt.Errorf("unknown datastore for filter: %s", o.DataStore)
	}
}
