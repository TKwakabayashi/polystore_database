package executor

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/id"
	"polystore_database/src/go/native"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/store"
)

// backend.go はストアへの実アクセス（storage.Registry の接続を使う）。ドライバ境界はここに閉じ、
// 識別子は uuid 文字列で受け渡す（内側では id.UUID。EXTENDING §2）。

// StoreBackend は storage.Registry 上の Backend。
type StoreBackend struct {
	reg *storage.Registry
}

// NewStoreBackend は接続済みの Registry から Backend を作る。
func NewStoreBackend(reg *storage.Registry) *StoreBackend { return &StoreBackend{reg: reg} }

func (b *StoreBackend) Run(ctx context.Context, q *native.Query, in []Key) ([][]any, error) {
	switch q.Store {
	case store.Graph:
		return b.runGraph(ctx, q, in)
	case store.Relational:
		return b.runSQL(ctx, q, in)
	case store.Columnar:
		return b.runCQL(ctx, q, in)
	case store.Document:
		return b.runMongo(ctx, q, in)
	case store.Kvs:
		return b.runKVS(ctx, q, in)
	}
	return nil, fmt.Errorf("executor: no backend for %s", q.Store)
}

func uuidStrings(in []Key) []string {
	out := make([]string, len(in))
	for i, k := range in {
		out[i] = k[0].String()
	}
	return out
}

// ---- graph ----

func (b *StoreBackend) neo4j() (neo4j.DriverWithContext, error) {
	d, ok := b.reg.Neo4j()
	if !ok {
		return nil, fmt.Errorf("executor: neo4j not connected")
	}
	return d, nil
}

func (b *StoreBackend) runGraph(ctx context.Context, q *native.Query, in []Key) ([][]any, error) {
	d, err := b.neo4j()
	if err != nil {
		return nil, err
	}
	params := map[string]any{}
	for k, v := range q.Params {
		params[k] = v
	}
	if len(q.InKeys) == 1 {
		params[native.InParam] = uuidStrings(in)
	} else if len(q.InKeys) > 1 {
		rows := make([]map[string]any, len(in))
		for i, k := range in {
			m := make(map[string]any, len(q.InKeys))
			for j, name := range q.InKeys {
				m[name] = k[j].String()
			}
			rows[i] = m
		}
		params[native.InParam] = rows
	}
	sess := d.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close(ctx)
	res, err := sess.Run(ctx, q.Text, params)
	if err != nil {
		return nil, fmt.Errorf("cypher %q: %w", q.Text, err)
	}
	var out [][]any
	for res.Next(ctx) {
		rec := res.Record()
		vals := make([]any, len(q.Columns))
		for i := range q.Columns {
			v, _ := rec.Get(fmt.Sprintf("c%d", i))
			vals[i] = v
		}
		out = append(out, vals)
	}
	return out, res.Err()
}

func (b *StoreBackend) Verbatim(ctx context.Context, cypher string, params map[string]any) ([]map[string]any, error) {
	d, err := b.neo4j()
	if err != nil {
		return nil, err
	}
	sess := d.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer sess.Close(ctx)
	res, err := sess.Run(ctx, cypher, params)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	for res.Next(ctx) {
		rec := res.Record()
		row := make(map[string]any, len(rec.Keys))
		for i, k := range rec.Keys {
			row[k] = rec.Values[i]
		}
		rows = append(rows, row)
	}
	return rows, res.Err()
}

// ---- relational ----

// expandIn は InArg 番目の `?` を入力件数分の `?, ?, …` に広げ、引数にも uuid を展開する。
func expandIn(text string, args []any, inArg int, ids []string) (string, []any) {
	if inArg < 0 {
		return text, args
	}
	var b strings.Builder
	n := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '?' {
			if n == inArg {
				b.WriteString(strings.TrimSuffix(strings.Repeat("?, ", len(ids)), ", "))
				n++
				continue
			}
			n++
		}
		b.WriteByte(text[i])
	}
	out := make([]any, 0, len(args)+len(ids))
	for i, a := range args {
		if i == inArg {
			for _, s := range ids {
				out = append(out, s)
			}
			continue
		}
		out = append(out, a)
	}
	return b.String(), out
}

func (b *StoreBackend) runSQL(ctx context.Context, q *native.Query, in []Key) ([][]any, error) {
	db, ok := b.reg.MySQL()
	if !ok {
		return nil, fmt.Errorf("executor: mysql not connected")
	}
	text, args := expandIn(q.Text, q.Args, q.InArg, uuidStrings(in))
	rows, err := db.QueryContext(ctx, text, args...)
	if err != nil {
		return nil, fmt.Errorf("sql %q: %w", text, err)
	}
	defer rows.Close()
	return scanSQL(rows, len(q.Columns))
}

func scanSQL(rows *sql.Rows, n int) ([][]any, error) {
	var out [][]any
	for rows.Next() {
		vals := make([]any, n)
		ptrs := make([]any, n)
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		for i, v := range vals {
			if bs, ok := v.([]byte); ok {
				vals[i] = string(bs)
			}
		}
		out = append(out, vals)
	}
	return out, rows.Err()
}

// ---- columnar ----

func (b *StoreBackend) runCQL(ctx context.Context, q *native.Query, in []Key) ([][]any, error) {
	sess, ok := b.reg.Cassandra()
	if !ok {
		return nil, fmt.Errorf("executor: cassandra not connected")
	}
	args := append([]any{}, q.Args...)
	if q.InArg >= 0 {
		args[q.InArg] = uuidStrings(in)
	}
	iter := sess.Query(q.Text, args...).WithContext(ctx).Iter()
	// MapScan は null をゼロ値で返すので使わない。列ごとに **T へ読み、null は nil のまま残す。
	cols := iter.Columns()
	pos := make([]int, len(q.Columns))
	for i := range q.Columns {
		pos[i] = -1
		for j, c := range cols {
			if c.Name == fmt.Sprintf("c%d", i) {
				pos[i] = j
			}
		}
	}
	var out [][]any
	for {
		dest := make([]any, len(cols))
		for j, c := range cols {
			dest[j] = reflect.New(reflect.TypeOf(c.TypeInfo.New())).Interface()
		}
		if !iter.Scan(dest...) {
			break
		}
		vals := make([]any, len(q.Columns))
		for i, j := range pos {
			if j < 0 {
				continue
			}
			pp := reflect.ValueOf(dest[j]).Elem() // *T
			if !pp.IsNil() {
				vals[i] = pp.Elem().Interface()
			}
		}
		out = append(out, vals)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("cql %q: %w", q.Text, err)
	}
	return out, nil
}

// ---- document ----

// toBSON は native.Doc を bson.D に、入力 uuid の目印を ids に置き換えて変換する。
func toBSON(v any, ids []string) any {
	switch x := v.(type) {
	case native.Doc:
		d := bson.D{}
		for _, f := range x {
			d = append(d, bson.E{Key: f.Key, Value: toBSON(f.Val, ids)})
		}
		return d
	case []any:
		out := make(bson.A, len(x))
		for i, e := range x {
			out[i] = toBSON(e, ids)
		}
		return out
	}
	if v == native.InMarker {
		return ids
	}
	return v
}

func (b *StoreBackend) runMongo(ctx context.Context, q *native.Query, in []Key) ([][]any, error) {
	db, ok := b.reg.Mongo()
	if !ok {
		return nil, fmt.Errorf("executor: mongo not connected")
	}
	ids := uuidStrings(in)
	pipeline := mongo.Pipeline{}
	for _, s := range q.Pipeline {
		body := toBSON(s.Body, ids)
		if s.Op == "$limit" || s.Op == "$skip" {
			body = s.Body[0].Val // {$limit: n}
		}
		pipeline = append(pipeline, bson.D{{Key: s.Op, Value: body}})
	}
	cur, err := db.Collection(q.Collection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("mongo %s: %w", q.Collection, err)
	}
	defer cur.Close(ctx)
	var out [][]any
	for cur.Next(ctx) {
		var doc bson.M
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		vals := make([]any, len(q.Columns))
		for i := range q.Columns {
			vals[i] = fromBSON(doc[fmt.Sprintf("c%d", i)])
		}
		out = append(out, vals)
	}
	return out, cur.Err()
}

func fromBSON(v any) any {
	switch x := v.(type) {
	case primitive.DateTime:
		return x.Time().UTC()
	case bson.A:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = fromBSON(e)
		}
		return out
	}
	return v
}

// ---- kvs ----

func (b *StoreBackend) runKVS(ctx context.Context, q *native.Query, in []Key) ([][]any, error) {
	ldb, ok := b.reg.LevelDB()
	if !ok {
		return nil, fmt.Errorf("executor: leveldb not connected")
	}
	k := q.KVS
	var ids []id.UUID
	if k.Scan {
		var err error
		if ids, err = scanKVS(ctx, ldb, k); err != nil {
			return nil, err
		}
	} else {
		for _, key := range in {
			ids = append(ids, key[0])
		}
	}
	var out [][]any
	for i, u := range ids {
		if i%1024 == 0 && ctx.Err() != nil { // 打ち切り（context）を一定件数ごとに確かめる
			return nil, ctx.Err()
		}
		r := Row{k.Alias: u}
		for _, p := range k.Props {
			raw, err := ldb.Get(codec.BuildEntityKey(k.Label, u, p), nil)
			switch {
			case err == leveldb.ErrNotFound:
				r[k.Alias+"."+p] = nil
			case err != nil:
				return nil, err
			default:
				// キーがあって値が空なのは空文字列（DecodeValue は空を nil にするので先に扱う）。
				// 値は宣言型へ戻してからフィルタを評価する（日時を文字列のまま比べないため）。
				var v any
				if len(raw) == 0 && isStringType(k.Types[p]) {
					v = ""
				} else {
					v = codec.DecodeValue(raw, k.Types[p])
				}
				r[k.Alias+"."+p] = decode(v, native.KindValue, k.Types[p])
			}
		}
		keep := true
		for _, f := range k.Filters {
			v, err := Eval(f, r)
			if err != nil {
				return nil, err
			}
			if !truthy(v) {
				keep = false
				break
			}
		}
		if !keep {
			continue
		}
		vals := make([]any, len(q.Columns))
		for i, c := range q.Columns {
			if c == k.Alias {
				vals[i] = u.String()
			} else {
				vals[i] = r[c]
			}
		}
		out = append(out, vals)
	}
	return out, nil
}

func isStringType(t string) bool {
	switch strings.ToLower(t) {
	case "", "string", "text":
		return true
	}
	return false
}

// scanKVS はラベルの uuid を集める（等値条件があれば逆引きインデックスの前方一致）。
func scanKVS(ctx context.Context, ldb *leveldb.DB, k *native.KVSQuery) ([]id.UUID, error) {
	var prefix []byte
	index := false
	if eq := k.PrimaryEq; eq != nil {
		var vb []byte
		switch v := eq.Value.(type) {
		case int64:
			vb = codec.EncodeInt(v)
		case string:
			vb = []byte(v)
		}
		if vb != nil {
			index = true
			prefix = []byte("index" + codec.Sep + k.Label + codec.Sep + eq.Prop + codec.Sep)
			prefix = append(prefix, vb...)
			prefix = append(prefix, []byte(codec.Sep)...)
		}
	}
	if !index {
		prefix = []byte(k.Label + codec.Sep)
	}
	seen := map[string]bool{}
	var out []id.UUID
	iter := ldb.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	n := 0
	for iter.Next() {
		if n++; n%4096 == 0 && ctx.Err() != nil {
			return nil, ctx.Err()
		}
		key := iter.Key()
		var s string
		if index {
			if i := bytes.LastIndex(key, []byte(codec.Sep)); i >= 0 {
				s = string(key[i+1:])
			}
		} else if parts := bytes.Split(key, []byte(codec.Sep)); len(parts) >= 2 {
			s = string(parts[1])
		}
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, id.FromAny(s))
		}
	}
	return out, iter.Error()
}
