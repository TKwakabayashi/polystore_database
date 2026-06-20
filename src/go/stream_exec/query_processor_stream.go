package stream_executor

import (
	"context"
	"database/sql"
	"fmt"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
)

type Record struct {
	Slots []string
}

type QueryProcessor struct {
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

	rg *storage.Registry
}

type Metrics struct {
	StepNum  int           // 実行順序
	OpType   string        // オペレーター種別
	Duration time.Duration // 実行時間
	RowCount int           // そのステップでの結果数
}

func NewQueryProcessor(ctx context.Context) (*QueryProcessor, error) {
	st := &plan.SlotTable{
		VarToSlot: make(map[string]int),
		SlotToVar: []string{},
	}

	qp := &QueryProcessor{
		records:   []Record{},
		slotTable: st,
		results:   []map[string]interface{}{},
		metrics:   make(map[int]Metrics),
		counts:    make(map[string]int),
		ctx:       ctx,
	}
	cfg, _ := storage.LoadConfig("")
	rg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	qp.rg = rg
	// Neo4j の設定
	neoDriver, ok := rg.Neo4j()
	if ok {
		qp.neoDriver = neoDriver
	} else {

	}

	// MongoDB の設定
	mDb, ok := rg.Mongo()
	if ok {
		qp.mDb = mDb
	} else {

	}

	// LevelDB の設定
	ldb, ok := rg.LevelDB()
	if ok {
		qp.ldb = ldb
	} else {

	}

	// MySQL の設定
	sqlDb, ok := rg.MySQL()
	if ok {
		qp.sqlDb = sqlDb
	} else {

	}

	// Cassandra (gocql) の設定
	cqlSes, ok := rg.Cassandra()
	if ok {
		qp.cqlSes = cqlSes
	} else {

	}

	return qp, nil
}

func NewQueryProcessorWithConfig(ctx context.Context, cfg storage.Config) (*QueryProcessor, error) {
	st := &plan.SlotTable{
		VarToSlot: make(map[string]int),
		SlotToVar: []string{},
	}

	qp := &QueryProcessor{
		records:   []Record{},
		slotTable: st,
		results:   []map[string]interface{}{},
		metrics:   make(map[int]Metrics),
		counts:    make(map[string]int),
		ctx:       ctx,
	}

	rg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	qp.rg = rg
	// Neo4j の設定
	neoDriver, ok := rg.Neo4j()
	if ok {
		qp.neoDriver = neoDriver
	} else {

	}

	// MongoDB の設定
	mDb, ok := rg.Mongo()
	if ok {
		qp.mDb = mDb
	} else {

	}

	// LevelDB の設定
	ldb, ok := rg.LevelDB()
	if ok {
		qp.ldb = ldb
	} else {

	}

	// MySQL の設定
	sqlDb, ok := rg.MySQL()
	if ok {
		qp.sqlDb = sqlDb
	} else {

	}

	// Cassandra (gocql) の設定
	cqlSes, ok := rg.Cassandra()
	if ok {
		qp.cqlSes = cqlSes
	} else {

	}

	return qp, nil
}

func (qp *QueryProcessor) Close() error {
	if qp.rg == nil {
		return nil
	}
	return qp.rg.Close(qp.ctx)
}

func (qp *QueryProcessor) Reset() {
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

func (qp *QueryProcessor) ProcessQueryStream(op plan.PlanNode) ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	// rootquery := ParseQuery(cypher)
	counter := 0
	_, err := ExecuteOperatorStream(qp, op, &counter, &wg)

	wg.Wait()

	return qp.results, err
}

func ExecuteOperatorStream(qp *QueryProcessor, op plan.PlanNode, counter *int, wg *sync.WaitGroup) (chan []Record, error) {
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
			rowCount, err = ScanGraphStream(qp, o, outputStream)

		case *plan.Expand:
			opType = "Expand"
			rowCount, err = ExpandGraphStream(qp, o, inputStream, outputStream)

		case *plan.VarLengthExpand:
			opType = "VarLengthExpand"
			rowCount, err = streamVarLengthExpand(qp, o, inputStream, outputStream)

		case *plan.Filter:
			opType = "Filter"
			rowCount, err = streamFilterGraph(qp, o, inputStream, outputStream)

		case *plan.Projection:
			opType = "Projection"
			err = streamProjection(qp, o, inputStream)

			rowCount = len(qp.results)

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
