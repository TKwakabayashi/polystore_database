package test

import (
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"sort"
	"strings"
	"time"
)

type TrialResult struct {
	WorkloadName string
	Mode         string
	TotalTime    time.Duration
	Steps        []StepMetric
}

// workloadDef はワークロード定義関数のシグネチャ。
type workloadDef func(migrator.MigrationMode, bool) (string, map[string]string, []migrator.MigrationConfig)

// Registry は名前引きできるワークロード一覧（データセット依存）。
// 別データセットのワークロードを新ファイルに足したら、ここに登録する。
var Registry = map[string]workloadDef{
	"Q2":  DefineWorkloadQ2,
	"Q8":  DefineWorkloadQ8,
	"Q9":  DefineWorkloadQ9,
	"Q11": DefineWorkloadQ11,
	"IS1": DefineWorkloadIS1,
	"IS2": DefineWorkloadIS2,
	"IS3": DefineWorkloadIS3,
	"IS4": DefineWorkloadIS4,
	"IS5": DefineWorkloadIS5,
	"IS6": DefineWorkloadIS6,
}

// AvailableWorkloads は登録済みワークロード名をソートして返す。
func AvailableWorkloads() string {
	names := make([]string, 0, len(Registry))
	for n := range Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
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
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id", "content", "imageFile", "creationDate"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Comment", Properties: []string{"creationDate", "id", "content"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"creationDate", "id", "content", "imageFile"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName", "browserUsed", "creationDate", "email", "gender", "locationIP", "speaks"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Organisation", Properties: []string{"type", "name"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Place", Properties: []string{"type", "name"}, Mode: mode},
			{ObjType: plan.Relationship, Entity: "WORK_AT", Properties: []string{"workFrom"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName", "birthday", "gender", "locationIP", "browserUsed", "creationDate"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Place", Properties: []string{"type", "id"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id", "content", "imageFile", "creationDate"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: plan.Relationship, Entity: "KNOWS", Properties: []string{"creationDate"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"creationDate", "content", "imageFile"}},
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
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
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
			{ObjType: plan.Entity, Entity: "Forum", Properties: []string{"id", "title"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}
