package bulk

import (
	"fmt"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
)

// bulkStorePushdown は集約を単一ストアへ委譲して最終行を得る。
// 出力行のキーは RETURN 別名（＝ ReturnItem.Name）に一致させ、コーディネータ経路と揃える。
func bulkStorePushdown(qp *Processor, o *plan.StorePushdown) ([]Row, error) {
	switch o.Store {
	case store.Graph:
		return runGraphPushdown(qp, o.Query, o.Params)
	case store.Relational:
		return runRelationalPushdown(qp, o)
	case store.Document:
		return runDocumentPushdown(qp, o)
	case store.Columnar:
		return runColumnarPushdown(qp, o)
	default:
		return nil, fmt.Errorf("StorePushdown: unsupported store %q", o.Store)
	}
}

// ===== graph: パラメータ埋め込み済み Cypher をそのまま実行 =====

func runGraphPushdown(qp *Processor, query string, params map[string]string) ([]Row, error) {
	if qp.neoDriver == nil {
		return nil, fmt.Errorf("StorePushdown[graph]: neo4j driver is nil")
	}
	session := qp.neoDriver.NewSession(qp.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(qp.ctx)

	res, err := session.Run(qp.ctx, query, core.TypeParams(params))
	if err != nil {
		return nil, fmt.Errorf("StorePushdown[graph] run error: %w", err)
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
		return nil, fmt.Errorf("StorePushdown[graph] iterate error: %w", err)
	}
	return rows, nil
}

// ===== relational (MySQL): SELECT ... aggexprs ... GROUP BY ... ORDER BY ... LIMIT =====

func runRelationalPushdown(qp *Processor, o *plan.StorePushdown) ([]Row, error) {
	if qp.sqlDb == nil {
		return nil, nil
	}
	sql, args := core.BuildRelationalSQL(o)
	rows, err := qp.sqlDb.QueryContext(qp.ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("StorePushdown[relational] error: %w\n  SQL: %s", err, sql)
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
	return result, nil
}

// ===== document (Mongo): $match → $group → ($addFields for DISTINCT) → $sort → $limit =====

func runDocumentPushdown(qp *Processor, o *plan.StorePushdown) ([]Row, error) {
	if qp.mDb == nil {
		return nil, nil
	}
	pipeline := core.BuildMongoPipeline(o)
	cur, err := qp.mDb.Collection(o.Table).Aggregate(qp.ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("StorePushdown[document] error: %w", err)
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
	return result, nil
}

// ===== columnar (Cassandra): 全体集約のみ（GROUP BY / ORDER BY / DISTINCT はプランナで除外済み） =====

func runColumnarPushdown(qp *Processor, o *plan.StorePushdown) ([]Row, error) {
	if qp.cqlSes == nil {
		return nil, nil
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
		return nil, fmt.Errorf("StorePushdown[columnar] error: %w\n  CQL: %s", err, cql)
	}
	if len(row) > 0 {
		return []Row{row}, nil
	}
	return nil, nil
}

// ===== 共通ヘルパ =====

// typeParams は string パラメータを Neo4j 用の typed 値へ変換する
// （int → time.Time(RFC3339) → string の順。RunNeo4j の toValuedParams と同一規則）。

// coerceScalar は SQL ドライバ等が返す []byte を数値/文字列へ寄せる。
