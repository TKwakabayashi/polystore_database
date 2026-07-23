// Package bulk_executor は、演算子間で中間結果を全件マテリアライズする最も単純な
// 逐次実行モデルを提供する。各演算子は全行に対し 1 回だけ実行され、「何件に対して」
// 「どれだけの時間」かかったかを演算子単位でクリーンに計測する。
//
// ストリーミング（channel / goroutine / ExecPolicy）を排し、中間結果 []Record を
// そのまま return で受け渡す逐次版。既存 stream_exec / plan / logical_plan / storage /
package bulk

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
)

type Record struct {
	Slots []string
}

type Row = map[string]interface{}

type Processor struct {
	results []map[string]interface{}
	metrics map[int]Metrics

	neoDriver neo4j.DriverWithContext
	mDb       *mongo.Database
	ldb       *leveldb.DB
	sqlDb     *sql.DB
	cqlSes    *gocql.Session
	ctx       context.Context

	rg *storage.Registry
}

type Metrics struct {
	StepNum  int           // 実行順序（葉→根）
	OpType   string        // オペレーター種別
	Duration time.Duration // その演算子が自身の処理に費やした時間（子の実行は除外）
	InRows   int           // その演算子への入力行数（EntityScan は 0）
	RowCount int           // その演算子が出力した行数
}

func NewProcessor(ctx context.Context) (*Processor, error) {
	cfg, _ := storage.LoadConfig("")
	return NewProcessorWithConfig(ctx, cfg)
}

func NewProcessorWithConfig(ctx context.Context, cfg storage.Config) (*Processor, error) {
	qp := &Processor{
		results: []map[string]interface{}{},
		metrics: make(map[int]Metrics),
		ctx:     ctx,
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
	qp.results = []map[string]interface{}{}
	qp.metrics = make(map[int]Metrics)
}

// StepMetrics は StepNum 昇順の演算子計測一覧を返す。
func (qp *Processor) StepMetrics() []Metrics {
	out := make([]Metrics, 0, len(qp.metrics))
	for step := 1; step <= len(qp.metrics); step++ {
		if m, ok := qp.metrics[step]; ok {
			out = append(out, m)
		}
	}
	return out
}

// ProcessQueryBulk は plan ツリーを全件マテリアライズ実行し、最終結果行を返す。
// tail（Projection/Aggregate/Sort/Limit/Return/StorePushdown）は row ストリームで、
// record パイプライン（EntityScan/Filter/Expand）は []Record で実行する。
func (qp *Processor) ProcessQueryBulk(op plan.PlanNode) ([]map[string]interface{}, error) {
	counter := 0
	rows, err := executeRowBulk(qp, op, &counter)
	if err != nil {
		return nil, err
	}
	qp.results = rows
	return qp.results, nil
}

// ExecuteOperatorBulk は record パイプライン（EntityScan/Filter/Expand/VarLengthExpand）を
// 全件マテリアライズ実行する。上流（子）を先に全件実行してから自演算子を処理し、出力 []Record を
// 返す。演算子ごとに step 番号（葉→根）・入出力件数・時間を記録する。
func ExecuteOperatorBulk(qp *Processor, op plan.PlanNode, counter *int) ([]Record, error) {
	if op == nil {
		return nil, fmt.Errorf("Empty Operator Passed")
	}

	// 1. 再帰的に上流（Child）を全件実行（この時間は計測に含めない）
	var input []Record
	if len(op.Children()) > 0 {
		var err error
		input, err = ExecuteOperatorBulk(qp, op.Children()[0], counter)
		if err != nil {
			return nil, err
		}
	}

	*counter++
	currentStep := *counter

	var (
		opType string
		output []Record
		err    error
	)
	inRows := len(input)

	// 2. 自演算子を全件処理（この区間だけ計測）
	start := time.Now()
	switch o := op.(type) {
	case *plan.EntityScan:
		opType = "EntityScan"
		output, err = scanByStore(qp, o)
	case *plan.Expand:
		opType = "Expand"
		output, err = ExpandGraphBulk(qp, o, input)
	case *plan.VarLengthExpand:
		opType = "VarLengthExpand"
		output, err = bulkVarLengthExpand(qp, o, input)
	case *plan.Filter:
		opType = "Filter"
		output, err = filterByStore(qp, o, input)
	default:
		return nil, fmt.Errorf("unexpected record operator: %T", op)
	}
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}

	rowCount := len(output)
	qp.metrics[currentStep] = Metrics{
		StepNum:  currentStep,
		OpType:   opType,
		Duration: duration,
		InRows:   inRows,
		RowCount: rowCount,
	}
	return output, nil
}

func scanByStore(qp *Processor, o *plan.EntityScan) ([]Record, error) {
	switch o.DataStore {
	case "graph", "", "unknown":
		return ScanGraphBulk(qp, o)
	case "document":
		return ScanDocBulk(qp, o)
	case "kvs":
		return ScanKvsBulk(qp, o)
	case "relational":
		return ScanRdbBulk(qp, o)
	case "columnar":
		return ScanColBulk(qp, o)
	default:
		return nil, fmt.Errorf("unknown datastore for scan: %s", o.DataStore)
	}
}

func filterByStore(qp *Processor, o *plan.Filter, in []Record) ([]Record, error) {
	switch o.DataStore {
	case "graph", "", "unknown":
		return bulkFilterGraph(qp, o, in)
	case "document":
		return FilterDocBulk(qp, o, in)
	case "kvs":
		return FilterKvsBulk(qp, o, in)
	case "relational":
		return FilterRdbBulk(qp, o, in)
	case "columnar":
		return FilterColBulk(qp, o, in)
	default:
		return nil, fmt.Errorf("unknown datastore for filter: %s", o.DataStore)
	}
}
