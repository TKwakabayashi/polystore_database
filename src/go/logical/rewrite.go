package logical

// rewrite.go は段階 C: 論理書き換え（docs/operator_plan.md §3 / §5.4）。ストアは見ない。
//
//   - DeferOutputFetch: 出力にしか使わない GetProperty を Sort / Limit の後ろへ移す（遅延実体化）。
//     `RETURN f.firstName ORDER BY f.id LIMIT 10` の firstName は Limit 後の 10 行分だけ取ればよい。
//   - Prune: 上で使われない GetProperty / GetListProperty / ProjectPath を除く。
//
// リテラルの型は AST の段階で int64 / float64 / string に決まっているので、パラメータの型付けは不要
// （文法にパラメータ記号が無く、パース前に値へ置換されるため）。ストアの型への変換は R4 の lowering が行う。

// Rewrite は論理書き換えをすべて適用する。
func Rewrite(n Node) Node {
	n = DeferOutputFetch(n)
	return Prune(n)
}

// ================================
// 遅延実体化
// ================================

// DeferOutputFetch は「GetProperty* → Projection → [Sort] → [Limit] → [Projection]」の形で、
// Projection がそのまま出力するだけの列の GetProperty を Limit の後ろへ移す。
// Distinct が挟まる形は対象外（重複除去の対象列が変わるため）。
func DeferOutputFetch(n Node) Node {
	if n == nil {
		return nil
	}
	// 上から順に見る（Sort/Limit の連なりの最上位で 1 回だけ当てるため）。
	if r, below, ok := deferAt(n); ok {
		if below != nil {
			setChild(below, DeferOutputFetch(child(below)))
		}
		return r
	}
	switch x := n.(type) {
	case *NodeHashJoin:
		x.Left = DeferOutputFetch(x.Left)
		x.Right = DeferOutputFetch(x.Right)
		return x
	}
	if c := child(n); c != nil {
		setChild(n, DeferOutputFetch(c))
	}
	return n
}

// deferAt は n を最上位とする形に遅延実体化を当てる。当てたら新しい根と、再帰を続ける
// 下側の Projection（その入力から先を書き換える）を返す。
func deferAt(n Node) (Node, Node, bool) {
	// n を「最上位の Sort/Limit/trim」として、その下の Projection を探す。
	var trim *Projection
	top := n
	if p, ok := top.(*Projection); ok && isPassThrough(p) {
		trim = p
		top = p.Input
	}
	var sortLimit []Node // 上から順
	for {
		switch x := top.(type) {
		case *Limit:
			sortLimit = append(sortLimit, x)
			top = x.Input
			continue
		case *Sort:
			sortLimit = append(sortLimit, x)
			top = x.Input
			continue
		}
		break
	}
	proj, ok := top.(*Projection)
	if !ok || len(sortLimit) == 0 {
		return n, nil, false
	}

	// Projection 直下で連続する GetProperty のうち、素通しの出力にしか使われないもの。
	// Sort が参照する項目（並べ替えキー）は素通しとみなさない。
	sortRefs := map[string]bool{}
	for _, sl := range sortLimit {
		if s, ok := sl.(*Sort); ok {
			for _, k := range s.Keys {
				for _, c := range Columns(k.Expr) {
					sortRefs[c] = true
				}
			}
		}
	}
	used := map[string]int{} // Projection 内での列の参照回数（素通し以外）
	passItem := map[string]int{}
	for i, it := range proj.Items {
		if c, ok := it.Expr.(*Column); ok && !sortRefs[it.Name] {
			if _, dup := passItem[c.Key]; !dup {
				passItem[c.Key] = i
				continue
			}
		}
		for _, k := range Columns(it.Expr) {
			used[k]++
		}
	}

	var deferred []*GetProperty
	below := proj.Input
	var keep []*GetProperty
	for {
		g, ok := below.(*GetProperty)
		if !ok {
			break
		}
		_, isPass := passItem[g.Key()]
		if isPass && used[g.Key()] == 0 {
			deferred = append(deferred, g)
		} else {
			keep = append(keep, g)
		}
		below = g.Input
	}
	if len(deferred) == 0 {
		return n, nil, false
	}

	// 下側を組み直す（残す GetProperty は元の順で）。
	rebuilt := below
	for i := len(keep) - 1; i >= 0; i-- {
		g := *keep[i]
		g.Input = rebuilt
		rebuilt = &g
	}
	proj.Input = rebuilt

	// Projection から移す列の項目を外し、代わりに識別列を運ぶ。
	outName := map[string]string{} // 列キー → 元の項目名
	order := make([]string, 0, len(proj.Items))
	for _, it := range proj.Items {
		order = append(order, it.Name)
	}
	isDeferred := map[string]bool{}
	for _, g := range deferred {
		isDeferred[g.Key()] = true
	}
	var items []ProjItem
	names := map[string]bool{}
	for _, it := range proj.Items {
		if c, ok := it.Expr.(*Column); ok && isDeferred[c.Key] && passItem[c.Key] == indexOf(proj.Items, it.Name) {
			outName[c.Key] = it.Name
			continue
		}
		items = append(items, it)
		names[it.Name] = true
	}
	for _, g := range deferred {
		if !names[g.Alias] {
			items = append(items, ProjItem{Name: g.Alias, Expr: &Column{Key: g.Alias}})
			names[g.Alias] = true
		}
	}
	proj.Items = items

	// Sort/Limit の上に GetProperty を積み直し、最後の Projection で元の項目順へ戻す。
	upper := Node(sortLimit[0])
	for i := len(deferred) - 1; i >= 0; i-- {
		g := *deferred[i]
		g.Input = upper
		upper = &g
	}
	final := make([]ProjItem, 0, len(order))
	for _, name := range order {
		key := name
		for col, on := range outName {
			if on == name {
				key = col
			}
		}
		final = append(final, ProjItem{Name: name, Expr: &Column{Key: key}})
	}
	if trim != nil {
		// 既存の trim の項目名（＝元の出力名）を保ち、移した列だけ参照先を差し替える。
		for i, it := range trim.Items {
			for col, on := range outName {
				if c, ok := it.Expr.(*Column); ok && c.Key == on {
					trim.Items[i].Expr = &Column{Key: col}
				}
			}
		}
		trim.Input = upper
		return trim, proj, true
	}
	return &Projection{Input: upper, Items: final}, proj, true
}

func indexOf(items []ProjItem, name string) int {
	for i, it := range items {
		if it.Name == name {
			return i
		}
	}
	return -1
}

// isPassThrough は Projection がすべて素通し（Column(name) → name）か。
func isPassThrough(p *Projection) bool {
	for _, it := range p.Items {
		c, ok := it.Expr.(*Column)
		if !ok || c.Key != it.Name {
			return false
		}
	}
	return true
}

// ================================
// 刈り込み
// ================================

// Prune は上で使われない列を作る GetProperty / GetListProperty / ProjectPath を除く。
func Prune(n Node) Node {
	return prune(n, nil, true)
}

// prune は need（nil かつ all=true なら全列が必要）を受け取り、不要な演算子を除いた木を返す。
func prune(n Node, need map[string]bool, all bool) Node {
	if n == nil {
		return nil
	}
	needs := func(k string) bool { return all || need[k] }
	switch x := n.(type) {
	case *GetProperty:
		if !needs(x.Key()) {
			return prune(x.Input, need, all)
		}
	case *GetListProperty:
		if !needs(x.Key()) {
			return prune(x.Input, need, all)
		}
	case *ProjectPath:
		if !needs(x.Path) {
			return prune(x.Input, need, all)
		}
	}

	// 子へ伝える要求列。
	below := map[string]bool{}
	belowAll := false
	switch x := n.(type) {
	case *Projection:
		for _, it := range x.Items {
			for _, k := range Columns(it.Expr) {
				below[k] = true
			}
		}
	case *Aggregation:
		for _, it := range x.Keys {
			for _, k := range Columns(it.Expr) {
				below[k] = true
			}
		}
		for _, a := range x.Aggs {
			if a.Arg != nil {
				for _, k := range Columns(a.Arg) {
					below[k] = true
				}
			}
		}
	case *Distinct:
		belowAll = all
		for k := range need {
			below[k] = true
		}
	default:
		belowAll = all
		for k := range need {
			below[k] = true
		}
		for _, k := range Refs(n) {
			below[k] = true
		}
	}

	switch x := n.(type) {
	case *NodeHashJoin:
		x.Left = prune(x.Left, below, belowAll)
		x.Right = prune(x.Right, below, belowAll)
	default:
		if c := child(n); c != nil {
			setChild(n, prune(c, below, belowAll))
		}
	}
	return n
}

// Refs は演算子が入力行から読む列（識別列を含む）。
func Refs(n Node) []string {
	var out []string
	add := func(ks ...string) { out = append(out, ks...) }
	switch x := n.(type) {
	case *GetProperty:
		add(x.Alias)
	case *GetListProperty:
		add(x.Source)
	case *Expand:
		add(x.From)
	case *ExpandInto:
		add(x.From, x.To)
	case *VarLengthExpand:
		add(x.From)
	case *Filter:
		add(Columns(x.Pred)...)
		addQuantLists(x.Pred, &out)
	case *Projection:
		for _, it := range x.Items {
			add(Columns(it.Expr)...)
			addQuantLists(it.Expr, &out)
		}
	case *ProjectPath:
		for _, e := range x.Elems {
			if e.Seg {
				add(SegColumn(e.Var))
			} else {
				add(e.Var)
			}
		}
	case *Aggregation:
		for _, it := range x.Keys {
			add(Columns(it.Expr)...)
		}
		for _, a := range x.Aggs {
			if a.Arg != nil {
				add(Columns(a.Arg)...)
			}
		}
	case *Sort:
		for _, k := range x.Keys {
			add(Columns(k.Expr)...)
		}
	case *Distinct:
		add(x.Cols...)
	case *NodeHashJoin:
		add(x.Keys...)
	case *RelUniqueFilter:
		add(x.Rels...)
		for _, s := range x.Segs {
			add(SegColumn(s))
		}
	}
	return out
}

// addQuantLists は量化子の要素参照に対応するリスト値列（GetListProperty の出力）を要求列に加える。
func addQuantLists(e Expr, out *[]string) {
	Walk(e, func(x Expr) {
		q, ok := x.(*Quant)
		if !ok {
			return
		}
		call, ok := q.List.(*Call)
		if !ok || len(call.Args) != 1 {
			return
		}
		src, ok := call.Args[0].(*Column)
		if !ok {
			return
		}
		part := SegNodes
		if call.Fn == "relationships" {
			part = SegRels
		}
		Walk(q.Pred, func(y Expr) {
			if ep, ok := y.(*ElemProp); ok && ep.Var == q.Var {
				*out = append(*out, ListPropKey(src.Key, part, ep.Prop))
			}
		})
	})
}

// child / setChild は単項演算子の入力を読み書きする。
func child(n Node) Node {
	switch x := n.(type) {
	case *GetProperty:
		return x.Input
	case *GetListProperty:
		return x.Input
	case *Expand:
		return x.Input
	case *ExpandInto:
		return x.Input
	case *VarLengthExpand:
		return x.Input
	case *Filter:
		return x.Input
	case *Projection:
		return x.Input
	case *ProjectPath:
		return x.Input
	case *Aggregation:
		return x.Input
	case *Sort:
		return x.Input
	case *Limit:
		return x.Input
	case *Distinct:
		return x.Input
	case *RelUniqueFilter:
		return x.Input
	}
	return nil
}

func setChild(n, c Node) {
	switch x := n.(type) {
	case *GetProperty:
		x.Input = c
	case *GetListProperty:
		x.Input = c
	case *Expand:
		x.Input = c
	case *ExpandInto:
		x.Input = c
	case *VarLengthExpand:
		x.Input = c
	case *Filter:
		x.Input = c
	case *Projection:
		x.Input = c
	case *ProjectPath:
		x.Input = c
	case *Aggregation:
		x.Input = c
	case *Sort:
		x.Input = c
	case *Limit:
		x.Input = c
	case *Distinct:
		x.Input = c
	case *RelUniqueFilter:
		x.Input = c
	}
}

// Child は単項演算子の入力（二項・葉は nil）。physical から使う。
func Child(n Node) Node { return child(n) }

// WithInput は単項演算子 n の浅い複製を作り、入力を c に差し替える。
func WithInput(n Node, c Node) Node {
	var cp Node
	switch x := n.(type) {
	case *GetProperty:
		y := *x
		cp = &y
	case *GetListProperty:
		y := *x
		cp = &y
	case *Expand:
		y := *x
		cp = &y
	case *ExpandInto:
		y := *x
		cp = &y
	case *VarLengthExpand:
		y := *x
		cp = &y
	case *Filter:
		y := *x
		cp = &y
	case *Projection:
		y := *x
		cp = &y
	case *ProjectPath:
		y := *x
		cp = &y
	case *Aggregation:
		y := *x
		cp = &y
	case *Sort:
		y := *x
		cp = &y
	case *Limit:
		y := *x
		cp = &y
	case *Distinct:
		y := *x
		cp = &y
	case *RelUniqueFilter:
		y := *x
		cp = &y
	default:
		return n
	}
	setChild(cp, c)
	return cp
}
