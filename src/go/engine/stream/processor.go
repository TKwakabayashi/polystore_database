package stream

import (
	"context"
	"database/sql"
	"fmt"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
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
	metrics   map[int]Metrics
	metricsMu sync.Mutex
	counts    map[string]int

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

type Metrics struct {
	StepNum  int           // 実行順序
	OpType   string        // オペレーター種別
	Duration time.Duration // 実行時間
	RowCount int           // そのステップでの結果数
}

func NewProcessor(ctx context.Context) (*Processor, error) {
	st := &plan.SlotTable{
		VarToSlot: make(map[string]int),
		SlotToVar: []string{},
	}

	qp := &Processor{
		records:   []Record{},
		slotTable: st,
		results:   []map[string]interface{}{},
		metrics:   make(map[int]Metrics),
		counts:    make(map[string]int),
		ctx:       ctx,
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
	}

	qp := &Processor{
		records:   []Record{},
		slotTable: st,
		results:   []map[string]interface{}{},
		metrics:   make(map[int]Metrics),
		counts:    make(map[string]int),
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

	// メトリクス（実行時間など）のクリア
	qp.metrics = make(map[int]Metrics)

	// カウント情報のクリア
	qp.counts = make(map[string]int)
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

		var opType string
		var rowCount int
		var err error

		switch o := op.(type) {
		case *plan.EntityScan:
			opType = "EntityScan"
			rowCount, err = scanByStore(qp, o, outputStream)

		case *plan.Expand:
			opType = "Expand"
			rowCount, err = ExpandGraphStream(qp, o, inputStream, outputStream)

		case *plan.VarLengthExpand:
			opType = "VarLengthExpand"
			rowCount, err = streamVarLengthExpand(qp, o, inputStream, outputStream)

		case *plan.Filter:
			opType = "Filter"
			rowCount, err = filterByStore(qp, o, inputStream, outputStream)

		default:
			fmt.Printf("Unknown operator: %T\n", o)
		}

		if err != nil {
			fmt.Printf("Error in step %d (%s): %v\n", currentStep, opType, err)
		}

		qp.metricsMu.Lock()
		qp.metrics[currentStep] = Metrics{
			StepNum:  currentStep,
			OpType:   opType,
			RowCount: rowCount,
		}
		qp.metricsMu.Unlock()
	}()

	return outputStream, nil
}

func scanByStore(qp *Processor, o *plan.EntityScan, out chan<- []Record) (int, error) {
	switch o.DataStore {
	case store.Graph:
		return ScanGraphStream(qp, o, out)
	case store.Document:
		return ScanDocStream(qp, o, out)
	case store.Kvs:
		return ScanKvsStream(qp, o, out)
	case store.Relational:
		return ScanRdbStream(qp, o, out)
	case store.Columnar:
		return ScanColStream(qp, o, out)
	default:
		return 0, fmt.Errorf("unknown datastore for scan: %s", o.DataStore)
	}
}

func filterByStore(qp *Processor, o *plan.Filter, in <-chan []Record, out chan<- []Record) (int, error) {
	switch o.DataStore {
	case store.Graph:
		return streamFilterGraph(qp, o, in, out)
	case store.Document:
		return FilterDocStream(qp, o, in, out)
	case store.Kvs:
		return FilterKvsStream(qp, o, in, out)
	case store.Relational:
		return FilterRdbStream(qp, o, in, out)
	case store.Columnar:
		return FilterColStream(qp, o, in, out)
	default:
		return 0, fmt.Errorf("unknown datastore for filter: %s", o.DataStore)
	}
}
