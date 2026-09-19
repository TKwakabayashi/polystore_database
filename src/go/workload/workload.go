package workload

import (
	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	"sort"
	"strings"
)

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
	"TP1":  DefineWorkloadTP1,
	// クエリ表現拡充（feat/query_expansipn）の検証用ワークロード。
	"INLIST":   DefineWorkloadINLIST,
	"ISNULL":   DefineWorkloadISNULL,
	"VARCMP":   DefineWorkloadVARCMP,
	"MULTIPAT": DefineWorkloadMULTIPAT,
	"PATHFILT": DefineWorkloadPATHFILT,
	// 公式 LDBC BI（基本文法で表現可能な read）。
	"BI3": DefineWorkloadBI3,
	"BI9": DefineWorkloadBI9,
	// Phase A: graph verbatim 委譲で走る公式 read（WITH/OPTIONAL/CASE/collect/UNWIND/否定/循環/shortestPath）。
	"IS7":  DefineWorkloadIS7,
	"IC1":  DefineWorkloadIC1,
	"IC3":  DefineWorkloadIC3,
	"IC4":  DefineWorkloadIC4,
	"IC6":  DefineWorkloadIC6,
	"IC12": DefineWorkloadIC12,
	"BI5":  DefineWorkloadBI5,
	"BI6":  DefineWorkloadBI6,
	"BI7":  DefineWorkloadBI7,
	"BI8":  DefineWorkloadBI8,
	"BI11": DefineWorkloadBI11,
	"BI12": DefineWorkloadBI12,
	"BI14": DefineWorkloadBI14,
	"BI18": DefineWorkloadBI18,
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

// TP1: tail pushdown 検証用。all-graph traversal で集めた comment 作者ごとの返信数を、
// 集約キー（firstName/lastName）で GROUP BY する。seed フィルタは Person.id（graph に残す）を使い、
// tail が参照するのは Person.{firstName,lastName} のみ。これらを単一ストアへ移すと traversal は
// all-graph のまま tail だけが単一ストアに解決でき、tail pushdown（中間 UUID→一時テーブル→
// ネイティブ集約）の A/B 計測が可能になる。AGG1 は author.id を GROUP BY するため id を移すと
// seed フィルタも非 graph 化してしまい、この検証には使えない。
func DefineWorkloadTP1(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)\n" +
		"      <-[:REPLY_OF]-(comment:Comment)-[:HAS_CREATOR]->(author:Person)\n" +
		"RETURN author.firstName, author.lastName, count(comment) AS replyCount\n" +
		"ORDER BY replyCount DESC, author.lastName ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"firstName", "lastName"}, Mode: mode},
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

// =====================================================================
// クエリ表現拡充（feat/query_expansipn）の検証用ワークロード。
// それぞれ新機能を最小構成で突く。実値は既存 SNB dump の分布に合わせ要検証（0件回避）。
// =====================================================================

// INLIST: WHERE ... IN [list]（単一エイリアス述語）。全ストアへ push 可能。
// Organisation.type は Company/University の 2 値なので必ず非空。
func DefineWorkloadINLIST(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (o:Organisation)\n" +
		"WHERE o.type IN [\"Company\", \"University\"]\n" +
		"RETURN o.id, o.type\n" +
		"ORDER BY o.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Organisation", Properties: []string{"id", "type"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// ISNULL: WHERE ... IS NOT NULL（単一エイリアス述語）。CQL は非対応でコーディネータ評価へ
// フォールバック（capability で弾く）。Post は content=NULL、Comment は content 有りなので
// IS NOT NULL で本文ありメッセージに絞る。
func DefineWorkloadISNULL(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})<-[:HAS_CREATOR]-(m:Message)\n" +
		"WHERE m.content IS NOT NULL\n" +
		"RETURN m.id, m.content\n" +
		"ORDER BY m.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id", "content"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// VARCMP: 変数間比較 a.id < b.id（クロスエイリアス）。PostFilter でコーディネータ評価。
// 友人ペアのうち相手 id が自分より大きいものに絞る（対称重複除去の定石）。
func DefineWorkloadVARCMP(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (a:Person {id: $personId})-[:KNOWS]-(b:Person)\n" +
		"WHERE a.id < b.id\n" +
		"RETURN a.id, b.id, b.firstName\n" +
		"ORDER BY b.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// MULTIPAT: 複数パターン MATCH（カンマ区切り・変数 friend を共有・非循環）。
// 論理的には線形連鎖だが、カンマ区切りのパース／連結を検証する。
func DefineWorkloadMULTIPAT(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId})-[:KNOWS]-(friend:Person),\n" +
		"      (friend)-[:IS_LOCATED_IN]->(city:Place)\n" +
		"RETURN friend.id, friend.firstName, city.id\n" +
		"ORDER BY friend.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "15393162799448",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Place", Properties: []string{"id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// PATHFILT: 可変長パス全体への述語。要素述語 all(r IN relationships(p) WHERE ...) ＋
// 大域述語 length(p) <= 3。パスフィルタを持つ可変長展開は融合せず per-op 実行される。
func DefineWorkloadPATHFILT(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH p=(person:Person {id: $personId})-[:KNOWS*1..3]-(other:Person)\n" +
		"WHERE all(r IN relationships(p) WHERE r.creationDate > $minDate)\n" +
		"      AND length(p) <= 3\n" +
		"RETURN other.id, other.firstName\n" +
		"ORDER BY other.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId": "17592186049378", // Q9 と同じ高次数 Person（KNOWS*1..2 で多数到達）
		"minDate":  "2010-01-01T00:00:00.000Z",
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// BI3: 公式 LDBC BI Q3「ある国のあるタグクラスで人気のトピック(フォーラム)」。
// 単一線形パス（国←市←Person←Forum→Post←*Message→Tag→TagClass）＋ count(DISTINCT) の
// 単段グループ集約。既存の基本文法で表現できる BI。City/Country は現行カタログでは
// Place（type 付き）でモデル化されるため :Place を使う（cf. Q11）。
// TODO(実値): $country/$tagClass はロード中の SNB dump に実在する値へ要検証（0件回避）。
func DefineWorkloadBI3(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (country:Place {type: \"Country\", name: $country})<-[:IS_PART_OF]-(city:Place)\n" +
		"      <-[:IS_LOCATED_IN]-(person:Person)<-[:HAS_MODERATOR]-(forum:Forum)\n" +
		"      -[:CONTAINER_OF]->(post:Post)<-[:REPLY_OF*0..]-(message:Message)\n" +
		"      -[:HAS_TAG]->(tag:Tag)-[:HAS_TYPE]->(tagclass:TagClass {name: $tagClass})\n" +
		"RETURN forum.id, forum.title, forum.creationDate, person.id, count(DISTINCT message.id) AS messageCount\n" +
		"ORDER BY messageCount DESC, forum.id ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"country":  "Germany",       // TODO: 実値に調整
		"tagClass": "MusicalArtist", // TODO: 実値に調整
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Forum", Properties: []string{"id", "title", "creationDate"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// BI9: 公式 LDBC BI Q9「ある期間に最も活発なスレッド開始者」。
// 単一線形パス（Person←Post←*Message）＋日付範囲フィルタ＋ count(DISTINCT) ×2 の単段集約。
// creationDate の範囲比較を使うため、非graph 配置では型整合（datetime↔string）に依存する
// （graph 配置は文字列比較で機能する）。Post/Comment は現行カタログでは Message にモデル化される。
// TODO(実値): $startDate/$endDate はロード中の SNB dump の分布に合わせ要検証（0件回避）。
func DefineWorkloadBI9(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (person:Person)<-[:HAS_CREATOR]-(post:Post)<-[:REPLY_OF*0..]-(reply:Message)\n" +
		"WHERE post.creationDate >= $startDate AND post.creationDate <= $endDate\n" +
		"      AND reply.creationDate >= $startDate AND reply.creationDate <= $endDate\n" +
		"RETURN person.id, person.firstName, person.lastName,\n" +
		"       count(DISTINCT post.id) AS threadCount, count(DISTINCT reply.id) AS messageCount\n" +
		"ORDER BY messageCount DESC, person.id ASC\n" +
		"LIMIT 100"
	params := map[string]string{
		"startDate": "2010-01-01T00:00:00.000Z", // TODO: 実値に調整
		"endDate":   "2013-01-01T00:00:00.000Z", // TODO: 実値に調整
	}
	var migs []migrator.MigrationConfig
	if isMigration {
		migs = []migrator.MigrationConfig{
			{ObjType: plan.Entity, Entity: "Person", Properties: []string{"id", "firstName", "lastName"}, Mode: mode},
			{ObjType: plan.Entity, Entity: "Message", Properties: []string{"id", "creationDate"}, Mode: mode},
		}
	}
	return cypher, params, migs
}

// =====================================================================
// Phase A: graph verbatim 委譲で走る公式 LDBC read（WITH/OPTIONAL/CASE/collect/否定/循環）。
// 現行文法では構造化できないため ParseQuery が「all-graph verbatim」へフォールバックし、
// 原クエリを Neo4j へ丸ごと発行する（Neo4j がネイティブ処理）。graph 配置専用（migration なし）。
// 公式クエリを本ダンプの schema へ調整: City/Country→Place{type}, Company/University→Organisation{type}。
// 実値 param はロード中の SNB dump に合わせ要検証（0件回避）。出典は docs/LDBC_WORKLOADS.md。
// =====================================================================

// IS7（公式 Interactive Short 7）: あるメッセージへの返信と、返信者が原作者と KNOWS かを CASE 判定。
// OPTIONAL MATCH ＋ CASE。schema 調整不要。
func DefineWorkloadIS7(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (m:Message {id: $messageId})<-[:REPLY_OF]-(c:Comment)-[:HAS_CREATOR]->(p:Person)\n" +
		"OPTIONAL MATCH (m)-[:HAS_CREATOR]->(a:Person)-[r:KNOWS]-(p)\n" +
		"RETURN c.id AS commentId, c.content AS commentContent, c.creationDate AS commentCreationDate,\n" +
		"       p.id AS replyAuthorId, p.firstName AS replyAuthorFirstName, p.lastName AS replyAuthorLastName,\n" +
		"       CASE r WHEN null THEN false ELSE true END AS replyAuthorKnowsOriginalMessageAuthor\n" +
		"ORDER BY commentCreationDate DESC, replyAuthorId"
	params := map[string]string{
		"messageId": "1030792151051", // IS6 と同じ実在メッセージ（TODO: 返信を持つ id へ要検証）
	}
	return cypher, params, nil // graph verbatim（非graph 配置なし）
}

// IC3（公式 Interactive Complex 3）: 2 国に投稿している友人（1..2 ホップ）を国別に集計。
// WITH 多段・collect・否定（NOT ... IN cities）・CASE・sum・連鎖比較。
// schema 調整: Country→Place{type:"Country"}, City→Place{type:"City"}。
func DefineWorkloadIC3(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (countryX:Place {type: \"Country\", name: $countryXName}),\n" +
		"      (countryY:Place {type: \"Country\", name: $countryYName}),\n" +
		"      (person:Person {id: $personId})\n" +
		"WITH person, countryX, countryY\n" +
		"LIMIT 1\n" +
		"MATCH (city:Place {type: \"City\"})-[:IS_PART_OF]->(country:Place {type: \"Country\"})\n" +
		"WHERE country IN [countryX, countryY]\n" +
		"WITH person, countryX, countryY, collect(city) AS cities\n" +
		"MATCH (person)-[:KNOWS*1..2]-(friend)-[:IS_LOCATED_IN]->(city)\n" +
		"WHERE NOT person=friend AND NOT city IN cities\n" +
		"WITH DISTINCT friend, countryX, countryY\n" +
		"MATCH (friend)<-[:HAS_CREATOR]-(message)-[:IS_LOCATED_IN]->(country)\n" +
		"WHERE $endDate > message.creationDate >= $startDate AND country IN [countryX, countryY]\n" +
		"WITH friend, CASE WHEN country=countryX THEN 1 ELSE 0 END AS messageX,\n" +
		"             CASE WHEN country=countryY THEN 1 ELSE 0 END AS messageY\n" +
		"WITH friend, sum(messageX) AS xCount, sum(messageY) AS yCount\n" +
		"WHERE xCount>0 AND yCount>0\n" +
		"RETURN friend.id AS friendId, friend.firstName AS friendFirstName, friend.lastName AS friendLastName,\n" +
		"       xCount, yCount, xCount + yCount AS xyCount\n" +
		"ORDER BY xyCount DESC, friendId ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId":     "15393162799448",
		"countryXName": "Germany", // TODO: 実値に調整
		"countryYName": "China",   // TODO: 実値に調整
		"startDate":    "2010-01-01T00:00:00.000Z",
		"endDate":      "2011-01-01T00:00:00.000Z",
	}
	return cypher, params, nil
}

// IC4（公式 Interactive Complex 4）: 友人の Post に付いたタグのうち、期間内に初出のものを集計。
// WITH 多段・CASE・sum・日付連鎖比較。schema 調整不要（Post/HAS_TAG は実在）。
func DefineWorkloadIC4(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (person:Person {id: $personId})-[:KNOWS]-(friend:Person),\n" +
		"      (friend)<-[:HAS_CREATOR]-(post:Post)-[:HAS_TAG]->(tag)\n" +
		"WITH DISTINCT tag, post\n" +
		"WITH tag,\n" +
		"     CASE WHEN $startDate <= post.creationDate < $endDate THEN 1 ELSE 0 END AS valid,\n" +
		"     CASE WHEN post.creationDate < $startDate THEN 1 ELSE 0 END AS inValid\n" +
		"WITH tag, sum(valid) AS postCount, sum(inValid) AS inValidPostCount\n" +
		"WHERE postCount>0 AND inValidPostCount=0\n" +
		"RETURN tag.name AS tagName, postCount\n" +
		"ORDER BY postCount DESC, tagName ASC\n" +
		"LIMIT 10"
	params := map[string]string{
		"personId":  "15393162799448",
		"startDate": "2010-10-01T00:00:00.000Z", // TODO: 実値に調整
		"endDate":   "2010-11-01T00:00:00.000Z", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// BI7（公式 BI 7）: あるタグの付いたメッセージへの返信コメントに付く「関連タグ」を集計。
// 否定パターン NOT (comment)-[:HAS_TAG]->(tag) ＋ 複数パターン。schema 調整不要。
func DefineWorkloadBI7(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (tag:Tag {name: $tag})<-[:HAS_TAG]-(message:Message)<-[:REPLY_OF]-(comment:Comment)\n" +
		"      -[:HAS_TAG]->(relatedTag:Tag)\n" +
		"WHERE NOT (comment)-[:HAS_TAG]->(tag)\n" +
		"RETURN relatedTag.name AS relatedTagName, count(DISTINCT comment) AS count\n" +
		"ORDER BY count DESC, relatedTagName ASC\n" +
		"LIMIT 100"
	params := map[string]string{
		"tag": "Genghis_Khan", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// BI11（公式 BI 11）: ある国の 3 人が相互に KNOWS な三角形の数（循環パターン＋変数間比較）。
// schema 調整: Country→Place{type:"Country"}, City→Place{type:"City"}。
func DefineWorkloadBI11(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (country:Place {type: \"Country\", name: $country})<-[:IS_PART_OF]-(:Place {type: \"City\"})<-[:IS_LOCATED_IN]-(a:Person),\n" +
		"      (country)<-[:IS_PART_OF]-(:Place {type: \"City\"})<-[:IS_LOCATED_IN]-(b:Person),\n" +
		"      (country)<-[:IS_PART_OF]-(:Place {type: \"City\"})<-[:IS_LOCATED_IN]-(c:Person),\n" +
		"      (a)-[k1:KNOWS]-(b), (b)-[k2:KNOWS]-(c), (c)-[k3:KNOWS]-(a)\n" +
		"WHERE a.id < b.id AND b.id < c.id\n" +
		"      AND $startDate <= k1.creationDate AND $startDate <= k2.creationDate AND $startDate <= k3.creationDate\n" +
		"RETURN count(*) AS count"
	params := map[string]string{
		"country":   "Germany",                  // TODO: 実値に調整
		"startDate": "2010-01-01T00:00:00.000Z", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// IC6（公式 Interactive Complex 6）: 友人（1..2 ホップ）の Post のうち指定タグを含むものに
// 共起する他タグを集計。WITH・collect・UNWIND・複数パターン・否定。schema 調整不要。
func DefineWorkloadIC6(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (knownTag:Tag {name: $tagName})\n" +
		"WITH knownTag.id AS knownTagId\n" +
		"MATCH (person:Person {id: $personId})-[:KNOWS*1..2]-(friend)\n" +
		"WHERE NOT person=friend\n" +
		"WITH knownTagId, collect(DISTINCT friend) AS friends\n" +
		"UNWIND friends AS f\n" +
		"    MATCH (f)<-[:HAS_CREATOR]-(post:Post),\n" +
		"          (post)-[:HAS_TAG]->(t:Tag {id: knownTagId}),\n" +
		"          (post)-[:HAS_TAG]->(tag:Tag)\n" +
		"    WHERE NOT t = tag\n" +
		"    WITH tag.name AS tagName, count(post) AS postCount\n" +
		"RETURN tagName, postCount\n" +
		"ORDER BY postCount DESC, tagName ASC\n" +
		"LIMIT 10"
	params := map[string]string{
		"personId": "15393162799448",
		"tagName":  "Genghis_Khan", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// IC12（公式 Interactive Complex 12）: 友人が、指定タグクラス配下のタグを持つ Post へ返信した
// Comment 数を集計。relationship 型 alternation の可変長 [:HAS_TYPE|IS_SUBCLASS_OF*0..]・
// collect・toInteger。schema 調整不要（TagClass 階層は BI3 で HAS_TYPE 使用実績あり）。
func DefineWorkloadIC12(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (tag:Tag)-[:HAS_TYPE|IS_SUBCLASS_OF*0..]->(baseTagClass:TagClass)\n" +
		"WHERE tag.name = $tagClassName OR baseTagClass.name = $tagClassName\n" +
		"WITH collect(tag.id) AS tags\n" +
		"MATCH (:Person {id: $personId})-[:KNOWS]-(friend:Person)<-[:HAS_CREATOR]-(comment:Comment)-[:REPLY_OF]->(:Post)-[:HAS_TAG]->(tag:Tag)\n" +
		"WHERE tag.id IN tags\n" +
		"RETURN friend.id AS personId, friend.firstName AS personFirstName, friend.lastName AS personLastName,\n" +
		"       collect(DISTINCT tag.name) AS tagNames, count(DISTINCT comment) AS replyCount\n" +
		"ORDER BY replyCount DESC, toInteger(personId) ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId":     "15393162799448",
		"tagClassName": "Monarch", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// IC1（公式 Interactive Complex 1）: 名前一致する友人（1..3 ホップ、shortestPath 距離）と
// その所属校/勤務先。shortestPath（Neo4j 組込・verbatim で動作）＋OPTIONAL＋collect＋CASE。
// schema 調整: City→Place{type:"City"}, University/Company→Organisation{type}, Country→Place{type:"Country"}。
func DefineWorkloadIC1(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (p:Person {id: $personId}), (friend:Person {firstName: $firstName})\n" +
		"WHERE NOT p=friend\n" +
		"WITH p, friend\n" +
		"MATCH path = shortestPath((p)-[:KNOWS*1..3]-(friend))\n" +
		"WITH min(length(path)) AS distance, friend\n" +
		"ORDER BY distance ASC, friend.lastName ASC, toInteger(friend.id) ASC\n" +
		"LIMIT 20\n" +
		"MATCH (friend)-[:IS_LOCATED_IN]->(friendCity:Place {type: \"City\"})\n" +
		"OPTIONAL MATCH (friend)-[studyAt:STUDY_AT]->(uni:Organisation {type: \"University\"})-[:IS_LOCATED_IN]->(uniCity:Place {type: \"City\"})\n" +
		"WITH friend, collect(CASE uni.name WHEN null THEN null ELSE [uni.name, studyAt.classYear, uniCity.name] END) AS unis, friendCity, distance\n" +
		"OPTIONAL MATCH (friend)-[workAt:WORK_AT]->(company:Organisation {type: \"Company\"})-[:IS_LOCATED_IN]->(companyCountry:Place {type: \"Country\"})\n" +
		"WITH friend, collect(CASE company.name WHEN null THEN null ELSE [company.name, workAt.workFrom, companyCountry.name] END) AS companies, unis, friendCity, distance\n" +
		"RETURN friend.id AS friendId, friend.lastName AS friendLastName, distance AS distanceFromPerson,\n" +
		"       friend.birthday AS friendBirthday, friend.creationDate AS friendCreationDate, friend.gender AS friendGender,\n" +
		"       friend.browserUsed AS friendBrowserUsed, friend.locationIP AS friendLocationIp, friend.email AS friendEmails,\n" +
		"       friend.speaks AS friendLanguages, friendCity.name AS friendCityName, unis AS friendUniversities, companies AS friendCompanies\n" +
		"ORDER BY distanceFromPerson ASC, friendLastName ASC, toInteger(friendId) ASC\n" +
		"LIMIT 20"
	params := map[string]string{
		"personId":  "15393162799448",
		"firstName": "John", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// BI5（公式 BI 5）: あるタグの最も活発な投稿者（message/reply/like の加重スコア）。
// OPTIONAL＋LIKES/REPLY_OF＋算術スコア。temporal なし。LIKES 欠なら likeCount 0（エラーにならない）。
func DefineWorkloadBI5(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (tag:Tag {name: $tag})<-[:HAS_TAG]-(message:Message)-[:HAS_CREATOR]->(person:Person)\n" +
		"OPTIONAL MATCH (message)<-[likes:LIKES]-(:Person)\n" +
		"WITH person, message, count(likes) AS likeCount\n" +
		"OPTIONAL MATCH (message)<-[:REPLY_OF]-(reply:Comment)\n" +
		"WITH person, message, likeCount, count(reply) AS replyCount\n" +
		"WITH person, count(message) AS messageCount, sum(likeCount) AS likeCount, sum(replyCount) AS replyCount\n" +
		"RETURN person.id, replyCount, likeCount, messageCount,\n" +
		"       1*messageCount + 2*replyCount + 10*likeCount AS score\n" +
		"ORDER BY score DESC, person.id ASC\n" +
		"LIMIT 100"
	params := map[string]string{"tag": "Abbas_I_of_Persia"} // TODO: 実値に調整
	return cypher, params, nil
}

// BI6（公式 BI 6）: あるタグの最も権威あるユーザ（返信に付く like の連鎖）。OPTIONAL×2＋LIKES。temporal なし。
func DefineWorkloadBI6(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (tag:Tag {name: $tag})<-[:HAS_TAG]-(message1:Message)-[:HAS_CREATOR]->(person1:Person)\n" +
		"OPTIONAL MATCH (message1)<-[:LIKES]-(person2:Person)\n" +
		"OPTIONAL MATCH (person2)<-[:HAS_CREATOR]-(message2:Message)<-[like:LIKES]-(person3:Person)\n" +
		"RETURN person1.id, count(DISTINCT like) AS authorityScore\n" +
		"ORDER BY authorityScore DESC, person1.id ASC\n" +
		"LIMIT 100"
	params := map[string]string{"tag": "Arnold_Schwarzenegger"} // TODO: 実値に調整
	return cypher, params, nil
}

// BI8（公式 BI 8）: あるタグの中心人物（関心＋期間内投稿のスコア、友人スコア込み）。
// OPTIONAL＋collect＋UNWIND＋list 内包 size＋日付文字列比較。temporal-safe（文字列比較）。HAS_INTEREST は実在。
func DefineWorkloadBI8(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (tag:Tag {name: $tag})\n" +
		"OPTIONAL MATCH (tag)<-[interest:HAS_INTEREST]-(person:Person)\n" +
		"WITH tag, collect(person) AS interestedPersons\n" +
		"OPTIONAL MATCH (tag)<-[:HAS_TAG]-(message:Message)-[:HAS_CREATOR]->(person:Person)\n" +
		"         WHERE $startDate < message.creationDate AND message.creationDate < $endDate\n" +
		"WITH tag, interestedPersons, interestedPersons + collect(person) AS persons\n" +
		"UNWIND persons AS person\n" +
		"WITH DISTINCT tag, person\n" +
		"WITH tag, person,\n" +
		"  100 * size([(tag)<-[interest:HAS_INTEREST]-(person) | interest]) + size([(tag)<-[:HAS_TAG]-(message:Message)-[:HAS_CREATOR]->(person) WHERE $startDate < message.creationDate AND message.creationDate < $endDate | message]) AS score\n" +
		"OPTIONAL MATCH (person)-[:KNOWS]-(friend)\n" +
		"WITH person, score,\n" +
		"  100 * size([(tag)<-[interest:HAS_INTEREST]-(friend) | interest]) + size([(tag)<-[:HAS_TAG]-(message:Message)-[:HAS_CREATOR]->(friend) WHERE $startDate < message.creationDate AND message.creationDate < $endDate | message]) AS friendScore\n" +
		"RETURN person.id, score, sum(friendScore) AS friendsScore\n" +
		"ORDER BY score + friendsScore DESC, person.id ASC\n" +
		"LIMIT 100"
	params := map[string]string{
		"tag":       "Che_Guevara",              // TODO: 実値に調整
		"startDate": "2011-07-20T00:00:00.000Z", // TODO: 実値に調整
		"endDate":   "2011-07-25T00:00:00.000Z", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// BI12（公式 BI 12）: メッセージ数ごとの人数分布（本文あり・長さ・期間・言語で絞る）。
// OPTIONAL＋IS NOT NULL＋日付文字列比較＋IN リスト（言語はクエリにインライン＝list param 回避）。temporal-safe。
func DefineWorkloadBI12(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (person:Person)\n" +
		"OPTIONAL MATCH (person)<-[:HAS_CREATOR]-(message:Message)-[:REPLY_OF*0..]->(post:Post)\n" +
		"WHERE message.content IS NOT NULL AND message.length < $lengthThreshold\n" +
		"  AND message.creationDate > $startDate AND post.language IN ['ar', 'hu']\n" +
		"WITH person, count(message) AS messageCount\n" +
		"RETURN messageCount, count(person) AS personCount\n" +
		"ORDER BY personCount DESC, messageCount DESC"
	params := map[string]string{
		"lengthThreshold": "20",
		"startDate":       "2010-07-22T00:00:00.000Z", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// BI14（公式 BI 14）: 2 国間で最も交流のある人物ペア（都市別トップ）。
// OPTIONAL×4＋CASE＋LIKES＋collect 索引。temporal なし。schema 調整: Country/City→Place{type}。
func DefineWorkloadBI14(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH\n" +
		"  (country1:Place {type: \"Country\", name: $country1})<-[:IS_PART_OF]-(city1:Place {type: \"City\"})<-[:IS_LOCATED_IN]-(person1:Person),\n" +
		"  (country2:Place {type: \"Country\", name: $country2})<-[:IS_PART_OF]-(city2:Place {type: \"City\"})<-[:IS_LOCATED_IN]-(person2:Person),\n" +
		"  (person1)-[:KNOWS]-(person2)\n" +
		"WITH person1, person2, city1, 0 AS score\n" +
		"OPTIONAL MATCH (person1)<-[:HAS_CREATOR]-(c:Comment)-[:REPLY_OF]->(:Message)-[:HAS_CREATOR]->(person2)\n" +
		"WITH DISTINCT person1, person2, city1, score + (CASE WHEN c IS NULL THEN 0 ELSE 4 END) AS score\n" +
		"OPTIONAL MATCH (person1)<-[:HAS_CREATOR]-(m:Message)<-[:REPLY_OF]-(:Comment)-[:HAS_CREATOR]->(person2)\n" +
		"WITH DISTINCT person1, person2, city1, score + (CASE WHEN m IS NULL THEN 0 ELSE 1 END) AS score\n" +
		"OPTIONAL MATCH (person1)-[:LIKES]->(m:Message)-[:HAS_CREATOR]->(person2)\n" +
		"WITH DISTINCT person1, person2, city1, score + (CASE WHEN m IS NULL THEN 0 ELSE 10 END) AS score\n" +
		"OPTIONAL MATCH (person1)<-[:HAS_CREATOR]-(m:Message)<-[:LIKES]-(person2)\n" +
		"WITH DISTINCT person1, person2, city1, score + (CASE WHEN m IS NULL THEN 0 ELSE 1 END) AS score\n" +
		"ORDER BY city1.name ASC, score DESC, person1.id ASC, person2.id ASC\n" +
		"WITH city1, collect({score: score, person1Id: person1.id, person2Id: person2.id})[0] AS top\n" +
		"RETURN top.person1Id, top.person2Id, city1.name, top.score\n" +
		"ORDER BY top.score DESC, top.person1Id ASC, top.person2Id ASC\n" +
		"LIMIT 100"
	params := map[string]string{
		"country1": "Chile",     // TODO: 実値に調整
		"country2": "Argentina", // TODO: 実値に調整
	}
	return cypher, params, nil
}

// BI18（公式 BI 18）: あるタグに関心を持つ未接続ペアと共通友人数（友人推薦）。
// 複数パターン＋否定パターン＋変数間不等。temporal なし。HAS_INTEREST/KNOWS は実在。
func DefineWorkloadBI18(mode migrator.MigrationMode, isMigration bool) (string, map[string]string, []migrator.MigrationConfig) {
	cypher := "MATCH (tag:Tag {name: $tag})<-[:HAS_INTEREST]-(person1:Person)-[:KNOWS]-(mutualFriend:Person)-[:KNOWS]-(person2:Person)-[:HAS_INTEREST]->(tag)\n" +
		"WHERE person1 <> person2 AND NOT (person1)-[:KNOWS]-(person2)\n" +
		"RETURN person1.id AS person1Id, person2.id AS person2Id, count(DISTINCT mutualFriend) AS mutualFriendCount\n" +
		"ORDER BY mutualFriendCount DESC, person1Id ASC, person2Id ASC\n" +
		"LIMIT 20"
	params := map[string]string{"tag": "Frank_Sinatra"} // TODO: 実値に調整
	return cypher, params, nil
}
