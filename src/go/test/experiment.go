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
	"Q2":   DefineWorkloadQ2,
	"Q8":   DefineWorkloadQ8,
	"Q9":   DefineWorkloadQ9,
	"Q11":  DefineWorkloadQ11,
	"IS1":  DefineWorkloadIS1,
	"IS2":  DefineWorkloadIS2,
	"IS3":  DefineWorkloadIS3,
	"IS4":  DefineWorkloadIS4,
	"IS5":  DefineWorkloadIS5,
	"IS6":  DefineWorkloadIS6,
	"AGG1": DefineWorkloadAGG1,
	"AGG2": DefineWorkloadAGG2,
	"AGG3": DefineWorkloadAGG3,
	"AGG4": DefineWorkloadAGG4,
	"AGG5": DefineWorkloadAGG5,
	"AGG6": DefineWorkloadAGG6,
}

// AvailableWorkloads は登録済みワークロード名をソートして返す。
func AvailableWorkloads() string {
	return strings.Join(AllWorkloadNames(), ", ")
}

// AllWorkloadNames は登録済みワークロード名をソート済みスライスで返す。
func AllWorkloadNames() []string {
	names := make([]string, 0, len(Registry))
	for n := range Registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
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
		// 現行ダンプの実在Person。KNOWS*1..2 の 2ホップで HAS_CREATOR メッセージ約76万件に到達する高次数ノード。
		// maxDate でフィルタすると約7.2万件（数万〜数十万イメージ）に絞り込まれる。
		"personId": "17592186049378",
		"maxDate":  "2011-06-16T00:00:00.000Z",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
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
	if isMigration {
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
	if isMigration {
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

// AGG1: 集約(count) + 暗黙GROUP BY + ORDER BY(集約別名) + LIMIT の検証用。
// あるPersonのMessageへ返信したComment作者ごとの返信数を集計する（LDBC IC風）。
func DefineWorkloadAGG1(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)\n" +
		"      <-[:REPLY_OF]-(comment:Comment)-[:HAS_CREATOR]->(author:Person)\n" +
		"RETURN author.id, author.firstName, author.lastName, count(comment) AS replyCount\n" +
		"ORDER BY replyCount DESC, author.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG2: グループ集約（count + avg + max）+ ORDER BY(集約別名) + LIMIT。
// 友人ごとのメッセージ数と本文長の平均・最大を集計する。
// Comment/Post は :Message ラベルで書けばマルチラベルで一緒にマッチし、
// プロパティも Message 定義で解決される（type=NULL のため type 条件は使わない）。
func DefineWorkloadAGG2(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})-[:KNOWS]-(friend:Person)<-[:HAS_CREATOR]-(m:Message)\n" +
		"RETURN friend.id, friend.firstName, count(m) AS msgCount, avg(m.length) AS avgLen, max(m.length) AS maxLen\n" +
		"ORDER BY msgCount DESC, friend.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"length"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG3: 全体集約（GROUP BY なし）。count(*) / count(DISTINCT ...) / sum / avg / min / max を網羅。
func DefineWorkloadAGG3(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})-[:KNOWS]-(friend:Person)<-[:HAS_CREATOR]-(m:Message)\n" +
		"RETURN count(*) AS total, count(DISTINCT friend.id) AS friends,\n" +
		"       sum(m.length) AS totLen, avg(m.length) AS avgLen, min(m.length) AS minLen, max(m.length) AS maxLen"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"length"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG4: 新規文法 >= / <= の範囲フィルタ（集約なし）+ ORDER BY + LIMIT。
// 数値プロパティ m.length を使う（creationDate は graph に文字列格納されており、
// マッピングの datetime 型と不一致で比較が 0 件になるため範囲検証には不適）。
func DefineWorkloadAGG4(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)\n" +
		"WHERE m.length >= $minLen AND m.length <= $maxLen\n" +
		"RETURN m.id, m.length\n" +
		"ORDER BY m.length DESC, m.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
		"minLen":   "50",
		"maxLen":   "200",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id", "length"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG5: traversal 無しの単一エンティティ全体集約。非graph 単一ストア pushdown の検証用。
// Organisation.id を対象ストアへ migrate すると、RDB/Document/Columnar いずれもネイティブ集約へ委譲される。
func DefineWorkloadAGG5(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (o:Organisation)\n" +
		"WHERE o.id >= $minId AND o.id <= $maxId\n" +
		"RETURN count(*) AS cnt, min(o.id) AS mn, max(o.id) AS mx, sum(o.id) AS sm, avg(o.id) AS av"
	params := map[string]string{
		"minId": "100",
		"maxId": "5000",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Organisation", Properties: []string{"id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// AGG6: traversal 無しのグループ集約（GROUP BY o.type）+ ORDER BY。
// RDB/Document は GROUP BY 対応でネイティブ委譲、Columnar(CQL) は非対応でエンジンへフォールバック。
func DefineWorkloadAGG6(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (o:Organisation)\n" +
		"RETURN o.type, count(*) AS cnt, min(o.id) AS mn, max(o.id) AS mx\n" +
		"ORDER BY cnt DESC"
	params := map[string]string{}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Organisation", Properties: []string{"type", "id"}, Mode: mode},
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
