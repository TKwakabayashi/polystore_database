package executor

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
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

// Executor は物理計画を評価する。反復モデル（bulk / volcano / vectorized / stream / vecstream）は
// ExecuteModel で選ぶ。ストアへの発行・入力行との結合・値の取得・コーディネータの演算子の意味は
// すべてのモデルで共通（storeQuery / listFetch / runCoord）で、違うのは反復の仕方だけ。
type Executor struct {
	ctx        context.Context
	be         Backend
	types      native.TypeFn
	roundTrips atomic.Int64

	mu      sync.Mutex
	acc     map[physical.Node]*stepAcc
	order   []physical.Node // 計測の並び（葉 → 根）
	lowered map[*physical.StoreQuery]*native.Query
	lists   map[*physical.ListFetch][]*native.Query
}

type stepAcc struct {
	op      string
	dur     time.Duration
	in, out int
}

// New は Executor を作る。types は native.Lower に渡すプロパティ型の引き当て。
func New(ctx context.Context, be Backend, types native.TypeFn) *Executor {
	return &Executor{ctx: ctx, be: be, types: types}
}

// Steps は直前の実行の演算子ごとの計測（葉 → 根。並列モデルではバッチごとの時間の合計）。
func (e *Executor) Steps() []core.StepMetric {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []core.StepMetric
	for _, n := range e.order {
		if a, ok := e.acc[n]; ok {
			out = append(out, core.StepMetric{Step: len(out) + 1, Op: a.op, Duration: a.dur, InRows: a.in, OutRows: a.out})
		}
	}
	return out
}

// RoundTrips は直前の実行でストアへ投げたクエリの数。
func (e *Executor) RoundTrips() int64 { return e.roundTrips.Load() }

// reset は実行前に計測と変換の控えを初期化する。
func (e *Executor) reset(root physical.Node) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.roundTrips.Store(0)
	e.acc = map[physical.Node]*stepAcc{}
	e.lowered = map[*physical.StoreQuery]*native.Query{}
	e.lists = map[*physical.ListFetch][]*native.Query{}
	e.order = nil
	var walk func(physical.Node)
	walk = func(n physical.Node) {
		for _, c := range n.Children() {
			walk(c)
		}
		e.order = append(e.order, n)
	}
	walk(root)
}

// record は演算子 n の 1 回分（バッチ 1 つ、bulk なら全件）の計測を足し込む。
func (e *Executor) record(n physical.Node, dur time.Duration, in, out int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	a, ok := e.acc[n]
	if !ok {
		a = &stepAcc{op: opName(n)}
		e.acc[n] = a
	}
	a.dur += dur
	a.in += in
	a.out += out
}

func opName(n physical.Node) string {
	switch x := n.(type) {
	case *physical.StoreQuery:
		if x.IsFetch() {
			return "Fetch[" + x.Store.String() + "]"
		}
		return "StoreQuery[" + x.Store.String() + "]"
	case *physical.ListFetch:
		return "ListFetch[" + x.Store.String() + "]"
	case *physical.Coord:
		return fmt.Sprintf("%T", x.Op)[len("*logical."):]
	}
	return fmt.Sprintf("%T", n)
}

// lower は StoreQuery のネイティブクエリ（実行ごとに 1 回だけ変換する）。
func (e *Executor) lower(q *physical.StoreQuery) (*native.Query, error) {
	e.mu.Lock()
	nq, ok := e.lowered[q]
	e.mu.Unlock()
	if ok {
		return nq, nil
	}
	nq, err := native.Lower(q, e.types)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.lowered[q] = nq
	e.mu.Unlock()
	return nq, nil
}

func (e *Executor) lowerList(f *physical.ListFetch) ([]*native.Query, error) {
	e.mu.Lock()
	qs, ok := e.lists[f]
	e.mu.Unlock()
	if ok {
		return qs, nil
	}
	qs, err := native.LowerList(f, e.types)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	e.lists[f] = qs
	e.mu.Unlock()
	return qs, nil
}

// Execute は bulk（全件実体化）で評価する。
func (e *Executor) Execute(n physical.Node) ([]Row, error) {
	return e.ExecuteModel(n, settings.EngineBulk)
}

// step は物理演算子 1 個を入力行に適用する（全モデル共通の意味）。inputs は子ごとの入力。
func (e *Executor) step(n physical.Node, inputs [][]Row) ([]Row, error) {
	start := time.Now()
	in := 0
	for _, rs := range inputs {
		in += len(rs)
	}
	var (
		out []Row
		err error
	)
	switch x := n.(type) {
	case *physical.StoreQuery:
		var rows []Row
		if len(inputs) > 0 {
			rows = inputs[0]
		}
		out, err = e.storeQuery(x, rows)
	case *physical.ListFetch:
		out, err = e.listFetch(x, inputs[0])
	case *physical.Coord:
		out, err = runCoord(x.Op, inputs)
	default:
		err = fmt.Errorf("executor: physical node %T", n)
	}
	if err != nil {
		return nil, err
	}
	e.record(n, time.Since(start), in, len(out))
	return out, nil
}

// bulk は葉から全件を実体化して評価する。
func (e *Executor) bulk(n physical.Node) ([]Row, error) {
	var inputs [][]Row
	for _, c := range n.Children() {
		rows, err := e.bulk(c)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, rows)
	}
	return e.step(n, inputs)
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
			e.roundTrips.Add(1)
			return lb.RunLogical(q, keys)
		}
		nq, err := e.lower(q)
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
		e.roundTrips.Add(1)
		return e.be.Run(e.ctx, q, keys)
	}
	var out [][]any
	for i := 0; i < len(keys); i += size {
		j := i + size
		if j > len(keys) {
			j = len(keys)
		}
		e.roundTrips.Add(1)
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
		e.roundTrips.Add(1)
		vs, err := lb.ListValues(f, ids)
		if err != nil {
			return nil, err
		}
		values = vs
	} else if len(keys) > 0 {
		qs, err := e.lowerList(f)
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

// Run は Cypher を bulk で計画・実行して結果を返す。構造化できないクエリは graph への原文委譲（verbatim）。
func Run(ctx context.Context, be Backend, md *schema.MappingDictionary, cypher string, params map[string]string) (core.Result, error) {
	return RunModel(ctx, be, md, cypher, params, settings.EngineBulk)
}

// RunModel は反復モデルを指定して Cypher を計画・実行する。
func RunModel(ctx context.Context, be Backend, md *schema.MappingDictionary, cypher string, params map[string]string, model settings.EngineKind) (core.Result, error) {
	t0 := time.Now()
	p, cat, err := Plan(cypher, params, md)
	planT := time.Since(t0)
	if errors.Is(err, ast.ErrUnsupported) {
		start := time.Now()
		rows, verr := be.Verbatim(ctx, cypher, typedParams(params))
		if verr != nil {
			return core.Result{}, verr
		}
		return core.Result{Rows: rows, PlanTime: planT, ExecTime: time.Since(start), Engine: "redesign-" + string(model) + "(verbatim)", RoundTrips: 1}, nil
	}
	if err != nil {
		return core.Result{}, err
	}
	ex := New(ctx, be, cat.TypeOf)
	start := time.Now()
	rows, err := ex.ExecuteModel(p, model)
	if err != nil {
		return core.Result{}, err
	}
	res := core.Result{Rows: OutputRows(p, rows), PlanTime: planT, ExecTime: time.Since(start),
		Steps: ex.Steps(), RoundTrips: ex.RoundTrips(), Engine: "redesign-" + string(model)}
	if model == settings.EngineVectorized || model == settings.EngineVecStream || model == settings.EngineStream {
		res.VectorWidth = vectorWidth()
	}
	return res, nil
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
