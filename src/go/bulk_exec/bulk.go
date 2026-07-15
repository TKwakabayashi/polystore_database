// Package bulk_executor は、演算子間で中間結果を全件マテリアライズする
// 最も単純な逐次実行モデルを提供する。各演算子は全行に対し 1 回だけ実行され、
// 「何件に対してどれだけ時間がかかったか」を演算子単位でクリーンに計測する。
//
// 設計:
//   - 中間結果は行指向の Record ([]string、スロット index → 値)。plan.SlotTable で解釈。
//   - execute(op) が再帰で子を先に全件実行し、自演算子を全件処理して []Record を返す。
//   - Expand は全入力行の src.uuid をまとめて 1 本の Cypher で展開する（1 演算子 = 1 往復）。
//   - Projection は sink で、全行分のプロパティを一括取得して射影・ソート・リミットする。
//
// 既存 stream_exec / volcano_exec / plan / logical_plan / storage / codec には手を加えず、
// plan.PlanNode を消費して実行する。ストアアクセス層は volcano_exec と同一ロジック。
package bulk_executor

import (
	"context"
	"database/sql"
	"fmt"

	"polystore_database/src/go/storage"

	"github.com/gocql/gocql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"go.mongodb.org/mongo-driver/mongo"
)

// Record は 1 行（スロット index → 値）。SlotTable.VarToSlot で解釈する。
type Record []string

// Processor は全件マテリアライズ実行の実行コンテキスト。DB 接続と計測を保持する。
type Processor struct {
	rg  *storage.Registry
	ctx context.Context

	neoDriver neo4j.DriverWithContext
	mDb       *mongo.Database
	ldb       *leveldb.DB
	sqlDb     *sql.DB
	cqlSes    *gocql.Session

	roundTrips int64            // DB 往復回数
	metrics    map[int]*Metrics // step -> 計測
	nextStep   int              // 実行時に演算子へ割り当てる連番（葉→根）
	results    []map[string]interface{}
}

// NewProcessor は cfg で 5 ストアへ接続した Processor を返す。
func NewProcessor(ctx context.Context, cfg storage.Config) (*Processor, error) {
	rg, err := storage.NewRegistry(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("registry 初期化に失敗: %w", err)
	}

	p := &Processor{
		rg:      rg,
		ctx:     ctx,
		metrics: make(map[int]*Metrics),
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

// newReadSession は Neo4j の読み取りセッションを生成する。
func (p *Processor) newReadSession() neo4j.SessionWithContext {
	return p.neoDriver.NewSession(p.ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
		FetchSize:  neo4j.FetchAll,
	})
}
