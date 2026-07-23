package stream

import (
	"fmt"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
)

// streamStorePushdown は集約を単一ストアへ委譲して最終行を得る。
// 出力行のキーは RETURN 別名（＝ ReturnItem.Name）に一致させ、コーディネータ経路と揃える。
func streamStorePushdown(qp *Processor, o *plan.StorePushdown, out chan<- []Row) int {
	switch o.Store {
	case "graph":
		return runGraphPushdown(qp, o.Query, o.Params, out)
	case "relational":
		return runRelationalPushdown(qp, o, out)
	case "document":
		return runDocumentPushdown(qp, o, out)
	case "columnar":
		return runColumnarPushdown(qp, o, out)
	default:
		fmt.Printf("StorePushdown: unsupported store %q\n", o.Store)
		return 0
	}
}

// ===== graph: パラメータ埋め込み済み Cypher をそのまま実行 =====

// baseline(RunNeo4j) と同じ session.Run(原クエリ, params) に揃えて公平に比較する。
func runGraphPushdown(qp *Processor, query string, params map[string]string, out chan<- []Row) int {
	if qp.neoDriver == nil {
		fmt.Println("StorePushdown[graph]: neo4j driver is nil")
		return 0
	}
	session := qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(qp.ctx)

	res, err := session.Run(qp.ctx, query, core.TypeParams(params))
	if err != nil {
		fmt.Printf("StorePushdown[graph] run error: %v\n", err)
		return 0
	}
	var rows []Row
	for res.Next(qp.ctx) {
		rec := res.Record()
		row := make(Row, len(rec.Keys))
		for i, k := range rec.Keys {
			row[k] = rec.Values[i]
		}
		rows = append(rows, row)
	}
	if err := res.Err(); err != nil {
		fmt.Printf("StorePushdown[graph] iterate error: %v\n", err)
	}
	if len(rows) > 0 {
		out <- rows
	}
	return len(rows)
}

// ===== relational (MySQL): SELECT ... aggexprs ... GROUP BY ... ORDER BY ... LIMIT =====

func runRelationalPushdown(qp *Processor, o *plan.StorePushdown, out chan<- []Row) int {
	if qp.sqlDb == nil {
		return 0
	}
	sql, args := core.BuildRelationalSQL(o)
	rows, err := qp.sqlDb.QueryContext(qp.ctx, sql, args...)
	if err != nil {
		fmt.Printf("StorePushdown[relational] error: %v\n  SQL: %s\n", err, sql)
		return 0
	}
	defer rows.Close()
	cols, _ := rows.Columns()

	var result []Row
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(Row, len(cols))
		for i, c := range cols {
			row[c] = core.CoerceScalar(vals[i])
		}
		result = append(result, row)
	}
	if len(result) > 0 {
		out <- result
	}
	return len(result)
}

// ===== document (Mongo): $match → $group → ($addFields for DISTINCT) → $sort → $limit =====

func runDocumentPushdown(qp *Processor, o *plan.StorePushdown, out chan<- []Row) int {
	if qp.mDb == nil {
		return 0
	}
	pipeline := core.BuildMongoPipeline(o)
	cur, err := qp.mDb.Collection(o.Table).Aggregate(qp.ctx, pipeline)
	if err != nil {
		fmt.Printf("StorePushdown[document] error: %v\n", err)
		return 0
	}
	defer cur.Close(qp.ctx)

	var result []Row
	for cur.Next(qp.ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			continue
		}
		var idDoc bson.M
		if raw, ok := doc["_id"].(bson.M); ok {
			idDoc = raw
		}
		row := make(Row, len(o.Items))
		gi, ai := 0, 0
		for _, it := range o.Items {
			if it.IsAggregate {
				row[it.Name] = doc[fmt.Sprintf("a%d", ai)]
				ai++
			} else {
				if idDoc != nil {
					row[it.Name] = idDoc[fmt.Sprintf("g%d", gi)]
				}
				gi++
			}
		}
		result = append(result, row)
	}
	if len(result) > 0 {
		out <- result
	}
	return len(result)
}

// ===== columnar (Cassandra): 全体集約のみ（GROUP BY / ORDER BY / DISTINCT はプランナで除外済み） =====

func runColumnarPushdown(qp *Processor, o *plan.StorePushdown, out chan<- []Row) int {
	if qp.cqlSes == nil {
		return 0
	}
	cql, args := core.BuildColumnarCQL(o)
	iter := qp.cqlSes.Query(cql, args...).Iter()

	scanned := make(map[string]interface{})
	row := make(Row, len(o.Items))
	if iter.MapScan(scanned) {
		for i, it := range o.Items {
			row[it.Name] = scanned[fmt.Sprintf("c%d", i)]
		}
	}
	if err := iter.Close(); err != nil {
		fmt.Printf("StorePushdown[columnar] error: %v\n  CQL: %s\n", err, cql)
		return 0
	}
	if len(row) > 0 {
		out <- []Row{row}
	}
	return 1
}

// ===== 共通ヘルパ =====

// typeParams は string パラメータを Neo4j 用の typed 値へ変換する
// （int → time.Time(RFC3339) → string の順。RunNeo4j の toValuedParams と同一規則）。

// coerceScalar は SQL ドライバ等が返す []byte を数値/文字列へ寄せる。
