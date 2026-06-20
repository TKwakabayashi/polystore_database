package test

import (
	"context"
	"fmt"
	"log"
	planner "polystore_database/src/go/logical_plan"
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/storage"
	executor "polystore_database/src/go/stream_exec"
	"strconv"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const expMappingPath = "./schema/mapping.json"
const expMongoDbName = "polystore_doc"

// --- 実験結果を管理する構造体 ---
type StepMetric struct {
	Name     string
	Duration time.Duration
	Rows     int
}

type TrialResult struct {
	WorkloadName string
	Mode         string
	TotalTime    time.Duration
	Steps        []StepMetric
}

// --- メイン処理 ---

func main() {
	/*
		f, err := os.Create("trace.out")
		if err != nil {
			log.Fatalf("failed to create trace file: %v", err)
		}
		defer f.Close()

		// 2. トレースの開始
		if err := trace.Start(f); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
		defer trace.Stop() // プログラム終了時に必ず止める
	*/
	mappingPath := "./schema/mapping.json"

	fmt.Printf("=== 実験開始 ===\n\n")

	cypher, params, migs := DefineWorkloadQ9(migrator.ModeKvsToGraph, false)
	RunWorkflow(context.Background(), "Q9", mappingPath, cypher, params, migs)
}

// --- 実行用ヘルパー関数 ---

type ExperimentSummary struct {
	TotalLatency time.Duration
	SumOpTime    time.Duration
	RowCount     int
	Details      []string
}

// runWorkflow は 1 ワークロードを実行する。
// migs が空でなければ先に migration を行い、その後ベースラインと自作システムを比較する。
func RunWorkflow(ctx context.Context, name, mappingPath, cypher string, params map[string]string, migs []migrator.MigrationConfig) {
	fmt.Printf("===== Workload %s =====\n", name)

	scfg := storage.DefaultConfig()

	// 1. migration（指定があるときだけ）
	if len(migs) > 0 {
		for _, mc := range migs {
			t := time.Now()
			if err := migrator.MigrateData(mc, scfg); err != nil {
				fmt.Printf("Migration Error (%s): %v\n", mc.Entity, err)
			} else {
				fmt.Printf("Migrated %s in %v\n", mc.Entity, time.Since(t))
			}
		}
	}

	// 2. Neo4j ベースライン（接続情報は storage.Config から）
	if scfg.Neo4j != nil {
		fmt.Println(">> Neo4j 実行中...")
		if neoResult, err := RunNeo4j(ctx, *scfg.Neo4j, cypher, toValuedParams(params)); err != nil {
			log.Printf("Neo4j 実行エラー: %v", err)
		} else {
			PrintResult("Neo4j (Baseline)", neoResult)
		}
	}

	// 3. 自作システム
	fmt.Println(">> 自作システム 実行中...")
	if customResult, err := RunCustom(ctx, cypher, mappingPath, params); err != nil {
		log.Printf("自作システム 実行エラー: %v", err)
	} else {
		PrintResult("Custom System", customResult)
	}
}

// toValuedParams は string の params を Neo4j 用の typed params に変換する。
// 整数→int、RFC3339→time.Time、それ以外は文字列のまま。
func toValuedParams(params map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(params))
	for k, v := range params {
		if n, err := strconv.Atoi(v); err == nil {
			out[k] = n
		} else if t, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = t
		} else {
			out[k] = v
		}
	}
	return out
}

// --- Neo4j 実行用関数 ---
func RunNeo4j(ctx context.Context, cfg storage.Neo4jConfig, cypher string, params map[string]interface{}) (ExperimentSummary, error) {

	const n = 1

	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.User, cfg.Password, ""),
	)
	if err != nil {
		return ExperimentSummary{}, err
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var totalLatency time.Duration
	var totalRows int

	for i := 0; i < n; i++ {

		start := time.Now()
		count := 0

		_, err = session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {

			res, err := tx.Run(ctx, cypher, params)
			if err != nil {
				return nil, err
			}

			results := make(map[int]map[string]interface{})
			idx := 0

			for res.Next(ctx) {

				record := res.Record()

				row := make(map[string]interface{})

				for i, key := range record.Keys {
					row[key] = record.Values[i]
				}

				results[idx] = row
				idx++
			}

			count = len(results)

			return results, res.Err()
		})

		if err != nil {
			return ExperimentSummary{}, err
		}

		latency := time.Since(start)

		totalLatency += latency
		totalRows += count
	}

	avgLatency := totalLatency / time.Duration(n)
	avgRows := totalRows / n

	return ExperimentSummary{
		TotalLatency: avgLatency,
		SumOpTime:    avgLatency,
		RowCount:     avgRows,
	}, nil
}

// --- 自作システム 実行用関数 ---
func RunCustom(ctx context.Context, cypher string, mappingPath string, params map[string]string) (ExperimentSummary, error) {
	const n = 1

	var totalLatency time.Duration
	var totalRows int

	qp, err := executor.NewQueryProcessor(ctx)
	if err != nil {
		return ExperimentSummary{}, err
	}
	defer qp.Close()

	for i := 0; i < n; i++ {
		qp.Reset()

		start := time.Now()
		op, err := planner.ParseQuery(cypher, mappingPath, params)
		if err != nil {
			return ExperimentSummary{}, err
		}

		results, err := qp.ProcessQueryStream(op)
		latency := time.Since(start)
		if err != nil {
			return ExperimentSummary{}, err
		}

		totalLatency += latency
		totalRows += len(results)
	}

	return ExperimentSummary{
		TotalLatency: totalLatency / time.Duration(n),
		RowCount:     totalRows / n,
	}, nil
}

// --- 結果表示用関数 ---
func PrintResult(title string, s ExperimentSummary) {
	fmt.Printf("[%s]\n", title)
	fmt.Printf("  - 全体実行時間 (Latency): %v\n", s.TotalLatency)
	fmt.Printf("  - オペレータ合計時間 (Sum): %v\n", s.SumOpTime)
	fmt.Printf("  - 最終結果数: %d\n", s.RowCount)
	/*
		if len(s.Details) > 0 {
			fmt.Println("  - 実行詳細:")
			for _, d := range s.Details {
				fmt.Printf("      %s\n", d)
			}
		}
	*/
	fmt.Println()
}

// =====================================================================
// ワークロード定義：cypher・params（ハードコード）・migration 設定（接続情報なし）を返す。
// migration の実行（MigrateData）は別関数が migs を受けて行う。
// =====================================================================

func DefineWorkloadQ2(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (:Person {id: $personId})-[:KNOWS]-(friend:Person)<-[:HAS_CREATOR]-(m:Message)\n" +
		"WHERE m.creationDate <= $maxDate\n" +
		"RETURN friend.id, friend.firstName, friend.lastName,\n" +
		"       m.id, coalesce(m.content, m.imageFile), m.creationDate\n" +
		"ORDER BY m.creationDate DESC, m.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
		"maxDate":  "2011-06-16T00:00:00.000Z",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{ // 導出:要確認
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id", "content", "imageFile", "creationDate"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadQ8(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)\n" +
		"      <-[:REPLY_OF]-(comment:Comment)-[:HAS_CREATOR]->(author:Person)\n" +
		"RETURN author.id, author.firstName, author.lastName,\n" +
		"       comment.creationDate, comment.id, comment.content\n" +
		"ORDER BY comment.creationDate DESC, comment.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{ // 導出:要確認
			{ObjType: plan.Entity, Entity: "Comment", Properties: []string{"creationDate", "id", "content"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadQ9(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})-[:KNOWS*1..2]-(other:Person)\n" +
		"      <-[:HAS_CREATOR]-(m:Message)\n" +
		"WHERE m.creationDate < $maxDate\n" +
		"RETURN other.id, other.firstName, other.lastName,\n" +
		"       m.id, coalesce(m.content, m.imageFile), m.creationDate\n" +
		"ORDER BY m.creationDate DESC, m.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
		"maxDate":  "2011-06-16T00:00:00.000Z",
	}
	var migs []migrator.MigrationConfig
	if isMigration { // authoritative（旧 executeWorkFlowQ9 より）
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"creationDate", "id", "content", "imageFile"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadQ11(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})-[:KNOWS*1..3]-(friend:Person)\n" +
		"      -[work:WORK_AT]->(comp:Organisation {type: \"Company\"})\n" +
		"      -[:IS_LOCATED_IN]->(:Place {type: \"Country\", name: $countryName})\n" +
		"WHERE work.workFrom < $workFromYear\n" +
		"RETURN friend.id, friend.firstName, friend.lastName,\n" +
		"       comp.name, work.workFrom\n" +
		"ORDER BY work.workFrom ASC, friend.id ASC, comp.name DESC\n" +
		"LIMIT 10"
	params := map[string]string{
		"personId":     "15393162799448",
		"countryName":  "Germany",
		"workFromYear": "2008",
	}
	var migs []migrator.MigrationConfig
	if isMigration { // authoritative（旧 executeWorkFlowQ11 より）
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName", "browserUsed", "creationDate", "email", "gender", "locationIP", "speaks"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Organisation", Properties: []string{"type", "name"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Place", Properties: []string{"type", "name"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Relationship, Entity: "WORK_AT", Properties: []string{"workFrom"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS1(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})-[:IS_LOCATED_IN]->(c:City)\n" +
		"RETURN p.firstName, p.lastName, p.birthday,\n" +
		"       p.locationIP, p.browserUsed,\n" +
		"       c.id, p.gender, p.creationDate"
	params := map[string]string{
		"personId": "21990232558284",
	}
	var migs []migrator.MigrationConfig
	if isMigration { // authoritative（旧 executeWorkFlowIS1 より）
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName", "birthday", "gender", "locationIP", "browserUsed", "creationDate"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Place", Properties: []string{"type", "id"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS2(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)\n" +
		"RETURN m.id, m.content, m.imageFile, m.creationDate,\n" +
		"       p.id, p.firstName, p.lastName\n" +
		"ORDER BY m.creationDate DESC\n" +
		"LIMIT 10"
	params := map[string]string{
		"personId": "21990232558284", // TODO: 実値に調整
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{ // 導出:要確認
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id", "content", "imageFile", "creationDate"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS3(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})-[r:KNOWS]-(friend:Person)\n" +
		"RETURN friend.id, friend.firstName, friend.lastName, r.creationDate\n" +
		"ORDER BY r.creationDate DESC, friend.id ASC"
	params := map[string]string{
		"personId": "21990232558284", // TODO: 実値に調整
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{ // 導出:要確認
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Relationship, Entity: "KNOWS", Properties: []string{"creationDate"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS4(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (m:Message {id: $messageId})\n" +
		"RETURN m.creationDate, coalesce(m.content, m.imageFile)"
	params := map[string]string{
		"messageId": "1030792151051", // TODO: 実値に調整
	}
	var migs []migrator.MigrationConfig
	if isMigration { // authoritative（旧 executeWorkFlowIS4 より）
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"creationDate", "content", "imageFile"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS5(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (m:Message {id: $messageId})-[:HAS_CREATOR]->(p:Person)\n" +
		"RETURN p.id, p.firstName, p.lastName"
	params := map[string]string{
		"messageId": "1030792151051", // TODO: 実値に調整
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{ // 導出:要確認
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}

func DefineWorkloadIS6(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (m:Message {id: $messageId})-[:REPLY_OF*0..]->(:Post)\n" +
		"      <-[:CONTAINER_OF]-(f:Forum)\n" +
		"      -[:HAS_MODERATOR]->(mod:Person)\n" +
		"RETURN f.id, f.title,\n" +
		"       mod.id, mod.firstName, mod.lastName"
	params := map[string]string{
		"messageId": "1030792151051", // TODO: 実値に調整
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{ // 導出:要確認
			{ObjType: plan.Entity, Entity: "Forum", Properties: []string{"id", "title"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode, MappingPath: expMappingPath, MongoDbName: expMongoDbName},
		}
	}
	return cypher, params, migs
}
