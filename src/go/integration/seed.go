//go:build integration

// seed.go は結合テスト用の小さな合成 mini-SNB（LDBC SNB を模した数十ノード）を Neo4j へ投入する。
// 実 LDBC dump（約 300 万ノード）は CI では非現実的なため、ワークロードが参照する id/構造を
// 決定的に再現した最小グラフで「システムが機能するか」を検証する。
//
// 設計方針:
//   - 各ノードに `:Entity` 共通ラベルと `uuid` を付与する（record パイプラインの前提。
//     本番は setup で付与するが、ここでは seed が直接付ける）。
//   - creationDate は Neo4j の datetime 型、birthday は date 型で格納する（実 SF1 と同じ。
//     文字列で入れるとエンジンの datetime 比較で 0 件になる）。
//   - ワークロード（Q2/Q8/Q9/Q11・IS1〜6・AGG1〜6・BI3/BI9）が参照する実 id
//     （personId=15393162799448 等・messageId=1030792151051・Organisation id 100..5000・
//     country=Germany・tagClass=MusicalArtist）を実在させ、全て非空を返すように接続する。
package integration

import (
	"context"
	"fmt"

	"polystore_database/src/go/storage"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// seedCypher は mini-SNB を作る単一 CREATE 文（先頭で全削除）。
const seedCypher = `
CREATE
 (pA:Entity:Person {uuid:'N-pA', id:15393162799448, firstName:'Alice', lastName:'Anderson', birthday:date('1985-01-01'), gender:'female', locationIP:'10.0.0.1', browserUsed:'Firefox', creationDate:datetime('2010-01-01T00:00:00.000Z'), email:'alice@example.com', speaks:'en'}),
 (pB:Entity:Person {uuid:'N-pB', id:17592186049378, firstName:'Bob', lastName:'Brown', birthday:date('1986-02-02'), gender:'male', locationIP:'10.0.0.2', browserUsed:'Chrome', creationDate:datetime('2010-01-02T00:00:00.000Z'), email:'bob@example.com', speaks:'en'}),
 (pC:Entity:Person {uuid:'N-pC', id:21990232558284, firstName:'Carol', lastName:'Clark', birthday:date('1987-03-03'), gender:'female', locationIP:'10.0.0.3', browserUsed:'Safari', creationDate:datetime('2010-01-03T00:00:00.000Z'), email:'carol@example.com', speaks:'de'}),
 (pD:Entity:Person {uuid:'N-pD', id:100000000001, firstName:'Dave', lastName:'Davis', birthday:date('1988-04-04'), gender:'male', locationIP:'10.0.0.4', browserUsed:'Edge', creationDate:datetime('2010-01-04T00:00:00.000Z'), email:'dave@example.com', speaks:'de'}),
 (pE:Entity:Person {uuid:'N-pE', id:100000000002, firstName:'Eve', lastName:'Evans', birthday:date('1989-05-05'), gender:'female', locationIP:'10.0.0.5', browserUsed:'Firefox', creationDate:datetime('2010-01-05T00:00:00.000Z'), email:'eve@example.com', speaks:'en'}),
 (pF:Entity:Person {uuid:'N-pF', id:100000000003, firstName:'Frank', lastName:'Foster', birthday:date('1990-06-06'), gender:'male', locationIP:'10.0.0.6', browserUsed:'Chrome', creationDate:datetime('2010-01-06T00:00:00.000Z'), email:'frank@example.com', speaks:'en'}),
 (pG:Entity:Person {uuid:'N-pG', id:100000000004, firstName:'Grace', lastName:'Green', birthday:date('1991-07-07'), gender:'female', locationIP:'10.0.0.7', browserUsed:'Safari', creationDate:datetime('2010-01-07T00:00:00.000Z'), email:'grace@example.com', speaks:'en'}),
 (deCountry:Entity:Place:Country {uuid:'N-de', id:1001, type:'Country', name:'Germany', url:'http://dbpedia.org/Germany'}),
 (berlin:Entity:Place:City {uuid:'N-berlin', id:1002, type:'City', name:'Berlin', url:'http://dbpedia.org/Berlin'}),
 (acme:Entity:Organisation {uuid:'N-acme', id:200, type:'Company', name:'ACME'}),
 (uni:Entity:Organisation {uuid:'N-uni', id:300, type:'University', name:'TU Berlin'}),
 (f1:Entity:Forum {uuid:'N-f1', id:7001, title:'Music Forum', creationDate:datetime('2010-02-01T00:00:00.000Z')}),
 (tag1:Entity:Tag {uuid:'N-tag1', id:8001, name:'TheBeatles', url:'http://dbpedia.org/TheBeatles'}),
 (tc1:Entity:TagClass {uuid:'N-tc1', id:9001, name:'MusicalArtist', url:'http://dbpedia.org/MusicalArtist'}),
 (m1:Entity:Message:Post {uuid:'N-m1', id:1030792151051, content:'post one content', length:100, creationDate:datetime('2011-06-15T10:00:00.000Z'), locationIP:'10.0.0.1', browserUsed:'Firefox', language:'en'}),
 (m2:Entity:Message:Post {uuid:'N-m2', id:2001, content:'alice music post', length:120, creationDate:datetime('2011-05-10T10:00:00.000Z'), locationIP:'10.0.0.1', browserUsed:'Firefox', language:'en'}),
 (c1:Entity:Message:Comment {uuid:'N-c1', id:3001, content:'nice comment', length:60, creationDate:datetime('2011-05-11T10:00:00.000Z'), locationIP:'10.0.0.5', browserUsed:'Firefox', language:'en'}),
 (m3:Entity:Message:Post {uuid:'N-m3', id:4001, content:'eve post', length:80, creationDate:datetime('2011-06-15T09:00:00.000Z'), locationIP:'10.0.0.5', browserUsed:'Firefox', language:'en'}),
 (m4:Entity:Message:Post {uuid:'N-m4', id:5001, content:'grace post', length:90, creationDate:datetime('2011-05-01T00:00:00.000Z'), locationIP:'10.0.0.7', browserUsed:'Safari', language:'en'}),
 (m5:Entity:Message:Post {uuid:'N-m5', id:6001, content:'carol post', length:70, creationDate:datetime('2011-04-01T00:00:00.000Z'), locationIP:'10.0.0.3', browserUsed:'Safari', language:'de'}),
 (berlin)-[:IS_PART_OF {uuid:'E-1'}]->(deCountry),
 (acme)-[:IS_LOCATED_IN {uuid:'E-2'}]->(deCountry),
 (uni)-[:IS_LOCATED_IN {uuid:'E-3'}]->(deCountry),
 (pC)-[:IS_LOCATED_IN {uuid:'E-4'}]->(berlin),
 (pD)-[:IS_LOCATED_IN {uuid:'E-5'}]->(berlin),
 (pA)-[:KNOWS {uuid:'E-6', creationDate:datetime('2010-03-01T00:00:00.000Z')}]->(pD),
 (pA)-[:KNOWS {uuid:'E-7', creationDate:datetime('2010-03-02T00:00:00.000Z')}]->(pE),
 (pA)-[:KNOWS {uuid:'E-8', creationDate:datetime('2010-03-03T00:00:00.000Z')}]->(pC),
 (pB)-[:KNOWS {uuid:'E-9', creationDate:datetime('2010-03-04T00:00:00.000Z')}]->(pG),
 (pC)-[:KNOWS {uuid:'E-10', creationDate:datetime('2010-03-05T00:00:00.000Z')}]->(pF),
 (pD)-[:WORK_AT {uuid:'E-11', workFrom:2005}]->(acme),
 (m1)-[:HAS_CREATOR {uuid:'E-12'}]->(pA),
 (m2)-[:HAS_CREATOR {uuid:'E-13'}]->(pA),
 (c1)-[:HAS_CREATOR {uuid:'E-14'}]->(pE),
 (m3)-[:HAS_CREATOR {uuid:'E-15'}]->(pE),
 (m4)-[:HAS_CREATOR {uuid:'E-16'}]->(pG),
 (m5)-[:HAS_CREATOR {uuid:'E-17'}]->(pC),
 (c1)-[:REPLY_OF {uuid:'E-18'}]->(m2),
 (f1)-[:CONTAINER_OF {uuid:'E-19'}]->(m1),
 (f1)-[:CONTAINER_OF {uuid:'E-20'}]->(m2),
 (f1)-[:HAS_MODERATOR {uuid:'E-21'}]->(pC),
 (m2)-[:HAS_TAG {uuid:'E-22'}]->(tag1),
 (tag1)-[:HAS_TYPE {uuid:'E-23'}]->(tc1)
`

// SeedMiniSNB は Neo4j を全削除してから mini-SNB を投入する（べき等）。
// 隔離された結合テスト用 Neo4j（datastore/env/citest.env 等）に対して使うこと。
func SeedMiniSNB(ctx context.Context, cfg storage.Config) error {
	if cfg.Neo4j == nil {
		return fmt.Errorf("neo4j 設定がありません")
	}
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
	if err != nil {
		return fmt.Errorf("neo4j ドライバ生成: %w", err)
	}
	defer driver.Close(ctx)

	sess := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer sess.Close(ctx)

	if _, err := sess.Run(ctx, "MATCH (n) DETACH DELETE n", nil); err != nil {
		return fmt.Errorf("既存データ削除: %w", err)
	}
	if _, err := sess.Run(ctx, seedCypher, nil); err != nil {
		return fmt.Errorf("mini-SNB 投入: %w", err)
	}
	return nil
}
