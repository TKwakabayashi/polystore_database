package volcano_executor

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"

	"github.com/syndtr/goleveldb/leveldb/util"
	"go.mongodb.org/mongo-driver/bson"
)

// scanIterator は EntityScan の pull 実装。Open で対象ストアから uuid を全件取得し、
// Next で vectorWidth 件ずつバッチとして払い出す（スキャン自体のクエリは 1 往復/ラベル）。
type scanIterator struct {
	p    *Processor
	o    *plan.EntityScan
	step int

	slotCount int
	aliasIdx  int
	ids       []string
	pos       int
}

func (s *scanIterator) Open(ctx context.Context) error {
	start := time.Now()
	s.slotCount = len(s.o.OutputSlot.VarToSlot)
	s.aliasIdx = s.o.OutputSlot.VarToSlot[s.o.Alias]

	ids, err := s.p.scanIDs(s.o)
	if err != nil {
		return err
	}
	s.ids = ids
	s.p.recordOp(s.step, "EntityScan", time.Since(start), 0)
	return nil
}

func (s *scanIterator) Next(ctx context.Context) (*Batch, error) {
	if s.pos >= len(s.ids) {
		return nil, nil
	}
	start := time.Now()
	end := s.pos + s.p.vectorWidth
	if end > len(s.ids) {
		end = len(s.ids)
	}
	b := newBatch(s.slotCount, end-s.pos)
	for ; s.pos < end; s.pos++ {
		row := make([]string, s.slotCount)
		row[s.aliasIdx] = s.ids[s.pos]
		b.appendRow(row)
	}
	s.p.recordOp(s.step, "EntityScan", time.Since(start), b.n)
	return b, nil
}

func (s *scanIterator) Close(ctx context.Context) error { return nil }

// scanIDs はストア種別に応じて uuid 一覧を取得する。
func (p *Processor) scanIDs(o *plan.EntityScan) ([]string, error) {
	switch o.DataStore {
	case "graph", "", "unknown":
		return p.scanGraphIDs(o)
	case "document":
		return p.scanDocIDs(o)
	case "kvs":
		return p.scanKvsIDs(o)
	case "relational":
		return p.scanRdbIDs(o)
	case "columnar":
		return p.scanColIDs(o)
	default:
		return nil, fmt.Errorf("未知のスキャン対象ストア: %s", o.DataStore)
	}
}

// ---------- graph (Neo4j) ----------

func (p *Processor) scanGraphIDs(o *plan.EntityScan) ([]string, error) {
	var whereSections []string
	params := make(map[string]interface{})

	if len(o.Labels) > 0 {
		var lc []string
		for _, l := range o.Labels {
			lc = append(lc, fmt.Sprintf("n:%s", l))
		}
		whereSections = append(whereSections, "("+strings.Join(lc, " OR ")+")")
	}
	for i, cond := range o.Filter {
		op, err := cypherOp(cond.Type)
		if err != nil {
			return nil, err
		}
		pn := fmt.Sprintf("val%d", i)
		whereSections = append(whereSections, fmt.Sprintf("n.%s %s $%s", cond.Property, op, pn))
		params[pn], _ = codec.ConvertToNativeType(cond.Value, cond.DataType)
	}

	query := "MATCH (n)"
	if len(whereSections) > 0 {
		query += "\nWHERE " + strings.Join(whereSections, " AND ")
	}
	query += "\nRETURN n.uuid AS id"

	sess := p.newReadSession()
	defer sess.Close(p.ctx)

	p.countRoundTrip()
	res, err := sess.Run(p.ctx, query, params)
	if err != nil {
		return nil, err
	}
	var ids []string
	for res.Next(p.ctx) {
		if v, ok := res.Record().Get("id"); ok && v != nil {
			if s, ok := v.(string); ok {
				ids = append(ids, s)
			}
		}
	}
	return ids, res.Err()
}

// ---------- document (MongoDB) ----------

func (p *Processor) scanDocIDs(o *plan.EntityScan) ([]string, error) {
	if p.mDb == nil {
		return nil, nil
	}
	query := bson.D{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		if cond.DataType == "int" || cond.DataType == "integer" {
			if v32, ok := val.(int32); ok {
				val = int64(v32)
			}
		}
		query = append(query, bson.E{Key: cond.Property, Value: bson.M{mongoOp(cond.Type): val}})
	}

	var ids []string
	for _, label := range o.Labels {
		p.countRoundTrip()
		cur, err := p.mDb.Collection(label).Find(p.ctx, query)
		if err != nil {
			return nil, err
		}
		for cur.Next(p.ctx) {
			var rec struct {
				UUID string `bson:"uuid"`
			}
			if err := cur.Decode(&rec); err != nil || rec.UUID == "" {
				continue
			}
			ids = append(ids, rec.UUID)
		}
		cur.Close(p.ctx)
	}
	return ids, nil
}

// ---------- kvs (LevelDB) ----------

func (p *Processor) scanKvsIDs(o *plan.EntityScan) ([]string, error) {
	if p.ldb == nil {
		return nil, nil
	}
	var ids []string
	seen := make(map[string]struct{})
	primaryEq := findPrimaryEqCondition(o.Filter)

	for _, label := range o.Labels {
		var prefix []byte
		if primaryEq != nil {
			var valBytes []byte
			switch primaryEq.DataType {
			case "int", "integer", "long":
				v, _ := strconv.ParseInt(primaryEq.Value, 10, 64)
				valBytes = codec.EncodeInt(v)
			default:
				valBytes = []byte(primaryEq.Value)
			}
			prefix = []byte("index" + codec.Sep + label + codec.Sep + primaryEq.Property + codec.Sep)
			prefix = append(prefix, valBytes...)
			prefix = append(prefix, []byte(codec.Sep)...)
		} else {
			prefix = []byte(label + codec.Sep)
		}

		p.countRoundTrip()
		iter := p.ldb.NewIterator(util.BytesPrefix(prefix), nil)
		for iter.Next() {
			key := iter.Key()
			var uuid string
			if primaryEq != nil {
				if idx := bytes.LastIndex(key, []byte(codec.Sep)); idx != -1 {
					uuid = string(key[idx+1:])
				}
			} else {
				parts := bytes.Split(key, []byte(codec.Sep))
				if len(parts) >= 2 {
					uuid = string(parts[1])
				}
			}
			if uuid == "" {
				continue
			}
			if _, done := seen[uuid]; done {
				continue
			}
			if p.matchConditionsKVS(label, uuid, o.Filter) {
				seen[uuid] = struct{}{}
				ids = append(ids, uuid)
			}
		}
		iter.Release()
	}
	return ids, nil
}

// ---------- relational (MySQL) ----------

func (p *Processor) scanRdbIDs(o *plan.EntityScan) ([]string, error) {
	if p.sqlDb == nil {
		return nil, nil
	}
	var clauses []string
	var args []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		clauses = append(clauses, fmt.Sprintf("%s %s ?", cond.Property, sqlOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		args = append(args, val)
	}
	whereStr := ""
	if len(clauses) > 0 {
		whereStr = "WHERE " + strings.Join(clauses, " AND ")
	}

	var ids []string
	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM %s %s", label, whereStr)
		p.countRoundTrip()
		rows, err := p.sqlDb.QueryContext(p.ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil || id == "" {
				continue
			}
			ids = append(ids, id)
		}
		rows.Close()
	}
	return ids, nil
}

// ---------- columnar (Cassandra) ----------

func (p *Processor) scanColIDs(o *plan.EntityScan) ([]string, error) {
	if p.cqlSes == nil {
		return nil, nil
	}
	var whereClauses []string
	var args []interface{}
	for _, cond := range o.Filter {
		if cond == nil {
			continue
		}
		whereClauses = append(whereClauses, fmt.Sprintf("\"%s\" %s ?", cond.Property, cqlOp(cond.Type)))
		val, _ := codec.ConvertToNativeType(cond.Value, cond.DataType)
		args = append(args, val)
	}
	whereStr := ""
	if len(whereClauses) > 0 {
		whereStr = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	var ids []string
	for _, label := range o.Labels {
		query := fmt.Sprintf("SELECT uuid FROM \"%s\" %s ALLOW FILTERING", label, whereStr)
		p.countRoundTrip()
		iter := p.cqlSes.Query(query, args...).Iter()
		var id string
		for iter.Scan(&id) {
			if id != "" {
				ids = append(ids, id)
			}
		}
		if err := iter.Close(); err != nil {
			return ids, err
		}
	}
	return ids, nil
}

// cypherOp は比較演算子を Cypher 表記へ。
func cypherOp(t plan.ConditionType) (string, error) {
	switch t {
	case plan.CondEq:
		return "=", nil
	case plan.CondNeq:
		return "<>", nil
	case plan.CondGreater:
		return ">", nil
	case plan.CondLess:
		return "<", nil
	default:
		return "", fmt.Errorf("unknown operator")
	}
}
