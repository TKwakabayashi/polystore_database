package vecstream

import (
	"fmt"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.mongodb.org/mongo-driver/bson"
)

// runStoreFragment は委譲フラグメントを単一ストアで実行して最終行を得る（Plan を lowering して発行）。
// 出力行のキーは RETURN 別名（＝ ReturnItem.Name）に一致させ、コーディネータ経路と揃える。
// 委譲はストア内で完結するため DB 往復は 1 回。
func (p *Processor) runStoreFragment(f *plan.StoreFragment) ([]Row, error) {
	o := core.LowerFragment(f)
	switch o.Store {
	case store.Graph:
		return p.runGraphPushdown(o.Verbatim, o.Params)
	case store.Relational:
		return p.runRelationalPushdown(o)
	case store.Document:
		return p.runDocumentPushdown(o)
	case store.Columnar:
		return p.runColumnarPushdown(o)
	default:
		return nil, fmt.Errorf("StoreFragment: unsupported store %q", o.Store)
	}
}

// ===== graph: 原クエリ＋params を session.Run で発行（baseline と同一） =====

func (p *Processor) runGraphPushdown(query string, params map[string]string) ([]Row, error) {
	if p.neoDriver == nil {
		return nil, fmt.Errorf("StoreFragment[graph]: neo4j driver is nil")
	}
	session := p.neoDriver.NewSession(p.ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(p.ctx)

	p.countRoundTrip()
	res, err := session.Run(p.ctx, query, core.TypeParams(params))
	if err != nil {
		return nil, fmt.Errorf("StoreFragment[graph] run error: %w", err)
	}
	var rows []Row
	for res.Next(p.ctx) {
		rec := res.Record()
		row := make(Row, len(rec.Keys))
		for i, k := range rec.Keys {
			row[k] = rec.Values[i]
		}
		rows = append(rows, row)
	}
	if err := res.Err(); err != nil {
		return nil, fmt.Errorf("StoreFragment[graph] iterate error: %w", err)
	}
	return rows, nil
}

// ===== relational (MySQL): SELECT ... aggexprs ... GROUP BY ... ORDER BY ... LIMIT =====

func (p *Processor) runRelationalPushdown(o core.FragmentSpec) ([]Row, error) {
	if p.sqlDb == nil {
		return nil, nil
	}
	sqlText, args := core.BuildRelationalSQL(o)
	p.countRoundTrip()
	rows, err := p.sqlDb.QueryContext(p.ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("StoreFragment[relational] error: %w\n  SQL: %s", err, sqlText)
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

func (p *Processor) runDocumentPushdown(o core.FragmentSpec) ([]Row, error) {
	if p.mDb == nil {
		return nil, nil
	}
	pipeline := core.BuildMongoPipeline(o)
	p.countRoundTrip()
	cur, err := p.mDb.Collection(o.Table).Aggregate(p.ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("StoreFragment[document] error: %w", err)
	}
	defer cur.Close(p.ctx)

	var result []Row
	for cur.Next(p.ctx) {
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

func (p *Processor) runColumnarPushdown(o core.FragmentSpec) ([]Row, error) {
	if p.cqlSes == nil {
		return nil, nil
	}
	cql, args := core.BuildColumnarCQL(o)
	p.countRoundTrip()
	iter := p.cqlSes.Query(cql, args...).Iter()

	scanned := make(map[string]interface{})
	row := make(Row, len(o.Items))
	if iter.MapScan(scanned) {
		for i, it := range o.Items {
			row[it.Name] = scanned[fmt.Sprintf("c%d", i)]
		}
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("StoreFragment[columnar] error: %w\n  CQL: %s", err, cql)
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
