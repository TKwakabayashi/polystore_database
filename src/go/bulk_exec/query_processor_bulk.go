package bulk_executor

/*
import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/plan"
)

type Record struct {
	Slots []string
}


type QueryProcessor struct {
	records   []Record
	slotTable *SlotTable
	results   []map[string]interface{}
	metrics   map[int]Metrics
	counts    map[string]int

	neoDriver neo4j.DriverWithContext
	neoSes    neo4j.SessionWithContext
	mDb       *mongo.Database
	ldb       *leveldb.DB
	sqlDb     *sql.DB
	cqlSes    *gocql.Session
	ctx       context.Context
}

func NewSlotTable() *SlotTable {
	return &SlotTable{
		VarToSlot: make(map[string]int),
	}
}

type Metrics struct {
	StepNum  int           // 実行順序
	OpType   string        // オペレーター種別
	Duration time.Duration // 実行時間
	RowCount int           // そのステップでの結果数
}

func NewQueryProcessor(ctx context.Context) (*QueryProcessor, error) {
	st := &SlotTable{
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

	// 各種データベースの初期化

	// Neo4j の設定
	neoUri := "bolt://localhost:7690"
	neoUser := "neo4j"
	neoPass := "password"
	neoDriver, err := neo4j.NewDriverWithContext(neoUri, neo4j.BasicAuth(neoUser, neoPass, ""))
	if err != nil {
		return nil, fmt.Errorf("neo4j driver error: %w", err)
	}
	qp.neoDriver = neoDriver
	qp.neoSes = neoDriver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})

	// MongoDB の設定
	mongoUri := "mongodb://localhost:27017"
	mClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, fmt.Errorf("mongodb connect error: %w", err)
	}
	qp.mDb = mClient.Database("polystore_doc")

	// LevelDB の設定
	ldbPath := "./polystore_kvs"
	ldb, err := leveldb.OpenFile(ldbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("leveldb open error: %w", err)
	}
	qp.ldb = ldb

	// SQL (MySQL/PostgreSQL等) の設定
	sqlDb, err := sql.Open("mysql", "user:pass@tcp(127.0.0.1:3306)/polystore_relational")
	if err != nil {
		return nil, fmt.Errorf("sql open error: %w", err)
	}
	qp.sqlDb = sqlDb

	// Cassandra (gocql) の設定

	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "polystore_columnar"
	cqlSes, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("cassandra session error: %w", err)
	}
	qp.cqlSes = cqlSes

	return qp, nil
}

func (qp *QueryProcessor) Close() error {
	var errs []string

	// 1. Neo4j Session
	if qp.neoSes != nil {
		if err := qp.neoSes.Close(qp.ctx); err != nil {
			errs = append(errs, fmt.Sprintf("neo4j close error: %v", err))
		}
	}

	// 2. MongoDB (Clientの切断が必要)
	if qp.mDb != nil {
		if err := qp.mDb.Client().Disconnect(qp.ctx); err != nil {
			errs = append(errs, fmt.Sprintf("mongodb disconnect error: %v", err))
		}
	}

	// 3. LevelDB
	if qp.ldb != nil {
		if err := qp.ldb.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("leveldb close error: %v", err))
		}
	}

	// 4. SQL (MySQL/PostgreSQL)
	if qp.sqlDb != nil {
		if err := qp.sqlDb.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("sql close error: %v", err))
		}
	}

	// 5. Cassandra (gocql)
	if qp.cqlSes != nil {
		qp.cqlSes.Close()
	}

	// 複数のエラーが発生した場合はまとめて返す
	if len(errs) > 0 {
		return fmt.Errorf("errors occurred during close: %s", strings.Join(errs, "; "))
	}

	return nil
}

func (qp *QueryProcessor) Reset() {
	// 中間レコードのクリア
	qp.records = []Record{}

	// スロットテーブルの初期化（変数の割り当て情報をクリア）
	qp.slotTable = &SlotTable{
		VarToSlot: make(map[string]int),
		SlotToVar: []string{},
	}

	// 最終結果のクリア
	qp.results = []map[string]interface{}{}

	// メトリクス（実行時間など）のクリア
	qp.metrics = make(map[int]Metrics)

	// カウント情報のクリア
	qp.counts = make(map[string]int)
}

func (qp *QueryProcessor) ProcessQuery(op plan.PlanNode) ([]map[string]interface{}, error) {

	// rootquery := ParseQuery(cypher)
	counter := 0
	cpuFile, errl := os.Create("cpu.prof")
	if errl != nil {
		log.Fatal(errl)
	}
	pprof.StartCPUProfile(cpuFile)
	defer pprof.StopCPUProfile()

	err := ExecuteOperator(qp, op, &counter)
	return qp.results, err
}

func ExecuteOperator(qp *QueryProcessor, op planner.Operator, counter *int) error {

	if op == nil {
		return fmt.Errorf("Empty Operator Passed")
	}

	if len(op.Children()) > 0 {
		err := ExecuteOperator(qp, op.Children()[0], counter)
		if err != nil {
			return err
		}
	}

	*counter++
	currentStep := *counter
	var opType string
	var duration time.Duration

	switch o := op.(type) {
	case *plan.EntityScan:
		opType = "EntityScan"
		start := time.Now()
		var Ids []string
		var err error
		switch o.DataStore {
		case "graph":
			Ids, err = scanGraph(qp, o)
			if err != nil {
				fmt.Println(err)
			}
		case "document":
			Ids, _ = scanDocument(qp, o)
		case "kvs":
			Ids, _ = scanKVS(qp, o)
		case "relational":
			Ids, _ = scanRelational(qp, o)
		case "columnar":
			Ids, _ = scanColumnar(qp, o)
		}
		finalizeScan(qp, o.Alias, Ids)

		duration = time.Since(start)

	case *planner.Expand:
		opType = "Expand"
		start := time.Now()

		expandGraph(qp, o)

		duration = time.Since(start)

	case *planner.VarLengthExpand:
		opType = "VarLengthExpand"
		start := time.Now()

		ProcessVarLengthExpand(qp, o)

		duration = time.Since(start)

	case *planner.Filter:
		opType = "Filter"
		start := time.Now()

		var validIDs []string
		switch o.DataStore {
		case "graph":
			validIDs, _ = filterGraph(qp, o)
		case "document":
			validIDs, _ = filterDocument(qp, o)
		case "kvs":
			validIDs, _ = filterKVS(qp, o)
		case "relational":
			validIDs, _ = filterRelational(qp, o)
		case "columnar":
			validIDs, _ = filterColumnar(qp, o)
		}
		applyFilter(qp, o, validIDs)
		duration = time.Since(start)

	case *plan.Projection:
		opType = "Projection"
		start := time.Now()

		projectMulti(qp, o)

		duration = time.Since(start)
	}

	var rowcount int
	if opType == "Projection" {
		rowcount = len(qp.results)
	} else {
		rowcount = len(qp.records)
	}

	qp.metrics[currentStep] = Metrics{
		StepNum:  currentStep,
		OpType:   opType,
		Duration: duration,
		RowCount: rowcount,
	}
	return nil
}

// また今度
func ProcessRelationshipScan() {

}
*/
