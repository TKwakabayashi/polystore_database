package executor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/id"
	"polystore_database/src/go/logical"
	"polystore_database/src/go/native"
	"polystore_database/src/go/physical"
	"polystore_database/src/go/schema"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// Key は境界の識別列の組（native.Query.InKeys と同じ並び）。
type Key []id.UUID

// Backend はストアへのアクセス。結果は native.Query.Columns と同じ並びの生の値。
type Backend interface {
	Run(ctx context.Context, q *native.Query, in []Key) ([][]any, error)
	// Verbatim は原文の Cypher をそのまま graph へ投げる（構造化できないクエリの安全弁）。
	Verbatim(ctx context.Context, cypher string, params map[string]any) ([]map[string]any, error)
}

// logicalBackend はテスト用の口: ネイティブクエリを経ずに、StoreQuery の論理部分木を直接評価する
// バックエンド（DB なしで計画と実行器の組み合わせを検証するため）。
type logicalBackend interface {
	RunLogical(q *physical.StoreQuery, in []Key) ([]Row, error)
	ListValues(f *physical.ListFetch, ids []id.UUID) (map[id.UUID]any, error)
}

// Executor は物理計画を全件実体化で評価する（bulk）。
type Executor struct {
	ctx        context.Context
	be         Backend
	types      native.TypeFn
	steps      []core.StepMetric
	roundTrips int64
}

// New は Executor を作る。types は native.Lower に渡すプロパティ型の引き当て。
func New(ctx context.Context, be Backend, types native.TypeFn) *Executor {
	return &Executor{ctx: ctx, be: be, types: types}
}

// Steps は直前の実行の演算子ごとの計測（葉 → 根）。
func (e *Executor) Steps() []core.StepMetric { return e.steps }

// RoundTrips は直前の実行でストアへ投げたクエリの数。
func (e *Executor) RoundTrips() int64 { return e.roundTrips }

// Execute は物理計画を評価して行を返す。
func (e *Executor) Execute(n physical.Node) ([]Row, error) {
	e.steps, e.roundTrips = nil, 0
	return e.exec(n)
}

func (e *Executor) exec(n physical.Node) ([]Row, error) {
	var inputs [][]Row
	inRows := 0
	for _, c := range n.Children() {
		rows, err := e.exec(c)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, rows)
		inRows += len(rows)
	}
	start := time.Now()
	var (
		out []Row
		err error
		op  string
	)
	switch x := n.(type) {
	case *physical.StoreQuery:
		op = "StoreQuery[" + x.Store.String() + "]"
		if x.IsFetch() {
			op = "Fetch[" + x.Store.String() + "]"
		}
		var in []Row
		if len(inputs) > 0 {
			in = inputs[0]
		}
		out, err = e.storeQuery(x, in)
	case *physical.ListFetch:
		op = "ListFetch[" + x.Store.String() + "]"
		out, err = e.listFetch(x, inputs[0])
	case *physical.Coord:
		op = fmt.Sprintf("%T", x.Op)[len("*logical."):]
		out, err = runCoord(x.Op, inputs)
	default:
		err = fmt.Errorf("executor: physical node %T", n)
	}
	if err != nil {
		return nil, err
	}
	e.steps = append(e.steps, core.StepMetric{Step: len(e.steps) + 1, Op: op, Duration: time.Since(start), InRows: inRows, OutRows: len(out)})
	return out, nil
}

// chunkSize は 1 クエリに載せる識別子の数（0 = 分割しない）。min(希望値, 物理上限)。
func chunkSize(k store.Kind) int {
	lim := store.LimitsOf(k)
	want := lim.DefaultChunk
	if settings.MaterializeChunkSize > 0 {
		want = settings.MaterializeChunkSize
	}
	if lim.MaxInList > 0 && (want <= 0 || want > lim.MaxInList) {
		want = lim.MaxInList
	}
	if want < 0 {
		return 0
	}
	return want
}

// storeQuery はネイティブクエリを発行し、入力行があれば識別列で結合する。
func (e *Executor) storeQuery(q *physical.StoreQuery, in []Row) ([]Row, error) {
	issue := func(keys []Key) ([]Row, error) {
		if lb, ok := e.be.(logicalBackend); ok {
			e.roundTrips++
			return lb.RunLogical(q, keys)
		}
		nq, err := native.Lower(q, e.types)
		if err != nil {
			return nil, err
		}
		raw, err := e.run(nq, keys)
		if err != nil {
			return nil, err
		}
		return toRows(nq, raw)
	}
	if q.Leaf() {
		return issue(nil)
	}

	// 入力の識別列の組を重複なく集める。
	var keys []Key
	seen := map[string]bool{}
	keyOf := func(r Row) (Key, string, bool) {
		k := make(Key, len(q.Keys))
		vals := make([]any, len(q.Keys))
		for i, c := range q.Keys {
			u, ok := r[c].(id.UUID)
			if !ok || u.Empty() {
				return nil, "", false
			}
			k[i], vals[i] = u, u
		}
		return k, keyString(vals...), true
	}
	for _, r := range in {
		k, ks, ok := keyOf(r)
		if ok && !seen[ks] {
			seen[ks] = true
			keys = append(keys, k)
		}
	}
	var results []Row
	if len(keys) > 0 {
		var err error
		results, err = issue(keys)
		if err != nil {
			return nil, err
		}
	}
	byKey := map[string][]Row{}
	for _, r := range results {
		_, ks, ok := keyOf(r)
		if ok {
			byKey[ks] = append(byKey[ks], r)
		}
	}

	// 入力行ごとに結合（入力の重複は保つ）。値の取得だけのクエリは左外部結合（無ければ null）。
	fetch := q.IsFetch()
	var out []Row
	for _, r := range in {
		_, ks, ok := keyOf(r)
		matches := byKey[ks]
		if !ok || len(matches) == 0 {
			if fetch {
				nr := r.clone()
				for _, c := range q.Outputs {
					if _, has := nr[c]; !has {
						nr[c] = nil
					}
				}
				out = append(out, nr)
			}
			continue
		}
		for _, m := range matches {
			nr := r.clone()
			for k, v := range m {
				nr[k] = v
			}
			out = append(out, nr)
		}
	}
	return out, nil
}

// run は入力をチャンクに分けて発行する。
func (e *Executor) run(q *native.Query, keys []Key) ([][]any, error) {
	size := chunkSize(q.Store)
	if keys == nil || size <= 0 || len(keys) <= size {
		e.roundTrips++
		return e.be.Run(e.ctx, q, keys)
	}
	var out [][]any
	for i := 0; i < len(keys); i += size {
		j := i + size
		if j > len(keys) {
			j = len(keys)
		}
		e.roundTrips++
		part, err := e.be.Run(e.ctx, q, keys[i:j])
		if err != nil {
			return nil, err
		}
		out = append(out, part...)
	}
	return out, nil
}

// toRows は生の結果を行にする（識別列は id.UUID、区間・パス列は *ListVal、値は型名で復元）。
func toRows(q *native.Query, raw [][]any) ([]Row, error) {
	out := make([]Row, 0, len(raw))
	for _, vals := range raw {
		if len(vals) != len(q.Columns) {
			return nil, fmt.Errorf("executor: %s returned %d columns, want %d", q.Store, len(vals), len(q.Columns))
		}
		r := make(Row, len(vals))
		for i, c := range q.Columns {
			r[c] = decode(vals[i], q.ColKinds[i], q.ColTypes[i])
		}
		out = append(out, r)
	}
	return out, nil
}

func decode(v any, kind native.ColKind, typ string) any {
	switch kind {
	case native.KindIdent:
		if v == nil {
			return nil
		}
		return id.FromAny(normalize(v))
	case native.KindList:
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		lv := &ListVal{}
		for _, u := range asList(m["nodes"]) {
			lv.Nodes = append(lv.Nodes, id.FromAny(normalize(u)))
		}
		for _, u := range asList(m["rels"]) {
			lv.Rels = append(lv.Rels, id.FromAny(normalize(u)))
		}
		return lv
	}
	v = normalize(v)
	s, ok := v.(string)
	if typ == "" || !ok {
		return v
	}
	if typ == "date" || typ == "datetime" {
		// MySQL は DATETIME を "2006-01-02 15:04:05[.ffffff]" で返す（RFC3339 ではない）。
		for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999", "2006-01-02"} {
			if t, err := time.Parse(layout, s); err == nil {
				return t.UTC()
			}
		}
		return v
	}
	switch typ {
	case "double", "float":
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		return v
	case "int", "integer", "long":
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(f)
		}
		return v
	}
	if t, err := codec.ConvertToNativeType(s, typ); err == nil {
		return normalize(t)
	}
	return v
}

func asList(v any) []any {
	switch x := v.(type) {
	case []any:
		return x
	case []string:
		out := make([]any, len(x))
		for i, s := range x {
			out[i] = s
		}
		return out
	}
	return nil
}

// listFetch はリスト列の要素（全行の和集合）について値を取り、値のリスト列を作る。
func (e *Executor) listFetch(f *physical.ListFetch, in []Row) ([]Row, error) {
	elems := func(r Row) []id.UUID {
		lv, _ := r[f.Source].(*ListVal)
		if lv == nil {
			return nil
		}
		if f.Part == logical.SegRels {
			return lv.Rels
		}
		return lv.Nodes
	}
	var keys []Key
	seen := map[id.UUID]bool{}
	for _, r := range in {
		for _, u := range elems(r) {
			if !u.Empty() && !seen[u] {
				seen[u] = true
				keys = append(keys, Key{u})
			}
		}
	}
	values := map[id.UUID]any{}
	if lb, ok := e.be.(logicalBackend); ok && len(keys) > 0 {
		ids := make([]id.UUID, len(keys))
		for i, k := range keys {
			ids[i] = k[0]
		}
		e.roundTrips++
		vs, err := lb.ListValues(f, ids)
		if err != nil {
			return nil, err
		}
		values = vs
	} else if len(keys) > 0 {
		qs, err := native.LowerList(f, e.types)
		if err != nil {
			return nil, err
		}
		for _, q := range qs {
			raw, err := e.run(q, keys)
			if err != nil {
				return nil, err
			}
			rows, err := toRows(q, raw)
			if err != nil {
				return nil, err
			}
			for _, r := range rows {
				u, _ := r[q.Columns[0]].(id.UUID)
				values[u] = r[q.Columns[1]]
			}
		}
	}
	key := f.Key()
	out := make([]Row, 0, len(in))
	for _, r := range in {
		es := elems(r)
		list := make([]any, len(es))
		for i, u := range es {
			list[i] = values[u]
		}
		nr := r.clone()
		nr[key] = list
		out = append(out, nr)
	}
	return out, nil
}

// ================================
// 入口: Cypher → 結果
// ================================

// Plan は Cypher を物理計画にする。構造化できなければ ast.ErrUnsupported を包んだエラー。
func Plan(cypher string, params map[string]string, md *schema.MappingDictionary) (physical.Node, *physical.Catalog, error) {
	a, err := ast.Parse(cypher, params)
	if err != nil {
		return nil, nil, err
	}
	l, err := logical.Translate(a)
	if err != nil {
		return nil, nil, err
	}
	cat := physical.NewCatalog(md)
	p, err := physical.Build(logical.Rewrite(l), cat)
	if err != nil {
		return nil, nil, err
	}
	// ネイティブクエリが作れない形も、実行前に verbatim へ落とせるよう確かめておく。
	if err := checkLowerable(p, cat); err != nil {
		return nil, nil, err
	}
	return p, cat, nil
}

func checkLowerable(n physical.Node, cat *physical.Catalog) error {
	switch x := n.(type) {
	case *physical.StoreQuery:
		if _, err := native.Lower(x, cat.TypeOf); err != nil {
			return err
		}
	case *physical.ListFetch:
		if _, err := native.LowerList(x, cat.TypeOf); err != nil {
			return err
		}
	}
	for _, c := range n.Children() {
		if err := checkLowerable(c, cat); err != nil {
			return err
		}
	}
	return nil
}

// Run は Cypher を計画・実行して結果を返す。構造化できないクエリは graph への原文委譲（verbatim）。
func Run(ctx context.Context, be Backend, md *schema.MappingDictionary, cypher string, params map[string]string) (core.Result, error) {
	t0 := time.Now()
	p, cat, err := Plan(cypher, params, md)
	planT := time.Since(t0)
	if errors.Is(err, ast.ErrUnsupported) {
		start := time.Now()
		rows, verr := be.Verbatim(ctx, cypher, typedParams(params))
		if verr != nil {
			return core.Result{}, verr
		}
		return core.Result{Rows: rows, PlanTime: planT, ExecTime: time.Since(start), Engine: "redesign(verbatim)", RoundTrips: 1}, nil
	}
	if err != nil {
		return core.Result{}, err
	}
	ex := New(ctx, be, cat.TypeOf)
	start := time.Now()
	rows, err := ex.Execute(p)
	if err != nil {
		return core.Result{}, err
	}
	return core.Result{Rows: OutputRows(p, rows), PlanTime: planT, ExecTime: time.Since(start),
		Steps: ex.Steps(), RoundTrips: ex.RoundTrips(), Engine: "redesign"}, nil
}

// OutputRows は最終行を出力列だけの map にする（識別子は文字列、区間・パスは uuid 文字列のリスト）。
func OutputRows(p physical.Node, rows []Row) []map[string]any {
	cols := outputColumns(p)
	out := make([]map[string]any, len(rows))
	for i, r := range rows {
		m := make(map[string]any, len(cols))
		for _, c := range cols {
			m[c] = external(r[c])
		}
		out[i] = m
	}
	return out
}

func outputColumns(p physical.Node) []string {
	switch x := p.(type) {
	case *physical.StoreQuery:
		return x.Outputs
	case *physical.Coord:
		switch op := x.Op.(type) {
		case *logical.Projection:
			names := make([]string, len(op.Items))
			for i, it := range op.Items {
				names[i] = it.Name
			}
			return names
		case *logical.Aggregation:
			var names []string
			for _, k := range op.Keys {
				names = append(names, k.Name)
			}
			for _, a := range op.Aggs {
				names = append(names, a.Name)
			}
			return names
		}
		if len(x.Inputs) > 0 {
			return outputColumns(x.Inputs[0])
		}
	}
	return nil
}

func external(v any) any {
	switch x := v.(type) {
	case id.UUID:
		return x.String()
	case *ListVal:
		nodes := make([]any, len(x.Nodes))
		for i, u := range x.Nodes {
			nodes[i] = u.String()
		}
		rels := make([]any, len(x.Rels))
		for i, u := range x.Rels {
			rels[i] = u.String()
		}
		return map[string]any{"nodes": nodes, "rels": rels}
	}
	return v
}

// typedParams は verbatim 用に文字列パラメータを型付けする（bench.RunNeo4j と同じ規則）。
func typedParams(params map[string]string) map[string]any {
	out := make(map[string]any, len(params))
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
