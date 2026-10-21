package logical

import (
	"math"

	"polystore_database/src/go/ast"
	"polystore_database/src/go/stats"
)

// cost.go は統計（stats.Stats: ラベル件数とタイプ別の関係数）による行数の推定。翻訳はこれで
//   - 成分の起点（推定コストが最小の頂点）
//   - 起点が複数ある成分を分割するか（分割した方が推定コストが小さいときだけ）
//   - 辺をたどる順序（各段で行数の増え方が最小の辺から）
//
// を決める。コストは各段でストアが作る行数（述語で絞る前）の和。値の分布は持たないので、述語の選択率は既定値
// （一意キーの等値は 1/件数、それ以外の等値 1/10、範囲 1/3。System R の既定値に倣う。type のような値の種類が
// 少ない列を小さく見積もり過ぎないよう、等値を強く効かせない）。統計が無ければ（b.st == nil）規則に戻る。

const (
	eqSel    = 0.1
	rangeSel = 1.0 / 3
	unkSel   = 0.5
	// 上限の無い可変長は、下限から maxExtraHops ホップ先までで見積もる。
	maxExtraHops = 2
	// UNWIND 1 行あたりの要素数の見積もり。
	unwindFanout = 10
)

func (b *builder) labelsOf(g *growCtx, v string) []string {
	if vx, ok := g.vertices[v]; ok {
		return vx.labels
	}
	return b.scope[v].labels
}

// sel は述語 1 項の選択率。
func (b *builder) sel(g *growCtx, e ast.Expr) float64 {
	switch x := e.(type) {
	case *ast.Bin:
		switch x.Op {
		case "=":
			if pr, ok := x.L.(*ast.PropRef); ok && isUniqueKey(pr.Prop) {
				if _, lit := x.R.(*ast.Lit); lit {
					return 1 / b.st.Count(b.labelsOf(g, pr.Var))
				}
			}
			return eqSel
		case "<>":
			return 1 - eqSel
		case "<", ">", "<=", ">=":
			return rangeSel
		case "IN":
			if l, ok := x.R.(*ast.List); ok {
				return math.Min(1, float64(len(l.Elems))*eqSel)
			}
			return 10 * eqSel
		case "AND":
			return b.sel(g, x.L) * b.sel(g, x.R)
		case "OR":
			l, r := b.sel(g, x.L), b.sel(g, x.R)
			return l + r - l*r
		}
	case *ast.Un:
		switch x.Op {
		case "NOT":
			return 1 - b.sel(g, x.X)
		case "IS NULL":
			return eqSel
		case "IS NOT NULL":
			return 1 - eqSel
		}
	}
	return unkSel
}

// predFactor は、bound に added を足したとき新しく評価できるようになる述語の選択率の積。
func (b *builder) predFactor(g *growCtx, pending []ast.Expr, bound map[string]bool, added ...string) float64 {
	f := 1.0
	for _, e := range pending {
		ready, touches := true, false
		for _, d := range b.deps(e) {
			isNew := false
			for _, a := range added {
				if d == a {
					isNew = true
				}
			}
			if isNew {
				touches = true
			} else if !bound[d] {
				ready = false
			}
		}
		if ready && touches {
			f *= b.sel(g, e)
		}
	}
	return f
}

func statsDir(d Direction) stats.Dir {
	switch d {
	case Out:
		return stats.Out
	case In:
		return stats.In
	}
	return stats.Both
}

// edgeFactor は辺 e をたどったときの行数の倍率: 述語で絞る前（work。展開が作る行）と、その時点で評価できる
// 述語で絞った後（out）。
func (b *builder) edgeFactor(g *growCtx, e *patEdge, bound map[string]bool, pending []ast.Expr) (work, out float64) {
	from, to, dir := orient(e, bound)
	fl, tl := b.labelsOf(g, from), b.labelsOf(g, to)
	sd := statsDir(dir)
	added := []string{e.rel.Var}
	if !bound[to] {
		added = append(added, to)
	}
	var f float64
	if e.rel.Range == nil {
		f = b.st.Degree(fl, e.rel.Types, sd, tl)
	} else {
		// 1 ホップの次数 step（端のラベルの割合 frac は最後のホップだけに掛ける）。
		step := b.st.Degree(fl, e.rel.Types, sd, nil)
		frac := 1.0
		if step > 0 {
			frac = b.st.Degree(fl, e.rel.Types, sd, tl) / step
		}
		r := g.segRange[e.rel.Var]
		hi := r.Max
		if hi < 0 {
			hi = r.Min + maxExtraHops
		}
		for k := r.Min; k <= hi; k++ {
			if k == 0 {
				f++
			} else {
				f += math.Pow(step, float64(k)) * frac
			}
		}
	}
	work = f
	if bound[to] { // 両端が束縛済み: 着いた先がその頂点である確率
		f /= b.st.Count(tl)
	}
	return work, f * b.predFactor(g, pending, bound, added...)
}

// pickEdge は allowed の未処理の辺のうち、片端以上が束縛済みで行数の倍率が最小のもの（同じなら出現順）。
func (b *builder) pickEdge(g *growCtx, allowed func(*patEdge) bool, bound map[string]bool, done map[int]bool, pending []ast.Expr) *patEdge {
	var best *patEdge
	bestF := 0.0
	for _, e := range g.edges {
		if done[e.order] || !allowed(e) || !(bound[e.left] || bound[e.right]) {
			continue
		}
		if _, f := b.edgeFactor(g, e, bound, pending); best == nil || f < bestF {
			best, bestF = e, f
		}
	}
	return best
}

// simulate は bound0 から allowed の辺を pickEdge の順にたどったときの推定コスト（各段の work の和）と最後の行数。
func (b *builder) simulate(g *growCtx, bound0 map[string]bool, allowed func(*patEdge) bool, pending []ast.Expr, rows float64) (cost, out float64) {
	bound := map[string]bool{}
	for k := range bound0 {
		bound[k] = true
	}
	done := map[int]bool{}
	for {
		e := b.pickEdge(g, allowed, bound, done, pending)
		if e == nil {
			return cost, rows
		}
		done[e.order] = true
		work, out := b.edgeFactor(g, e, bound, pending)
		cost += rows * work
		rows *= out
		bound[e.rel.Var], bound[e.left], bound[e.right] = true, true, true
	}
}

// startRows は頂点 v を起点にしたときの最初の行数（件数 × v だけで評価できる述語の選択率）。
func (b *builder) startRows(g *growCtx, v string, pending []ast.Expr) float64 {
	return b.st.Count(b.labelsOf(g, v)) * b.predFactor(g, pending, map[string]bool{}, v)
}

// compPlan は成分の組み方の推定結果。
type compPlan struct {
	split bool
	start string  // split でないときの起点
	rows  float64 // 成分を組んだ後の推定行数（束縛済みの成分なら前の行数を含む）
}

// planComp は統計で成分の組み方を決める: 1 つの起点から伸ばす最良の形と、（起点の組が 2 つ以上なら）分割した形の
// 推定コストを比べる。
func (b *builder) planComp(g *growCtx, comp []string, groups [][]string, inComp map[string]bool, bound map[string]bool, pending []ast.Expr, splitOK bool) compPlan {
	all := func(e *patEdge) bool { return inComp[e.left] }
	var single compPlan
	singleCost := math.Inf(1)
	for _, v := range comp {
		if bound[v] { // 束縛済みの頂点があればそこから（前の計画の上に伸ばす）
			c, r := b.simulate(g, bound, all, pending, b.est)
			single, singleCost = compPlan{start: v, rows: r}, c
			break
		}
	}
	if math.IsInf(singleCost, 1) {
		for _, v := range comp {
			r0 := b.startRows(g, v, pending)
			c, r := b.simulate(g, map[string]bool{v: true}, all, pending, r0)
			if c+r0 < singleCost {
				single, singleCost = compPlan{start: v, rows: r}, c+r0
			}
		}
	}
	if !splitOK || len(groups) < 2 {
		return single
	}
	side := b.splitSide(g, groups, inComp)
	hasBound := bound[groups[0][0]]
	var splitCost, accRows float64
	accBound := map[string]bool{}
	for gi, grp := range groups {
		sb := map[string]bool{}
		var r0 float64
		if gi == 0 && hasBound {
			for k := range bound {
				sb[k] = true
			}
			r0 = b.est
		} else {
			sb[grp[0]] = true
			r0 = b.startRows(g, grp[0], pending)
			splitCost += r0
		}
		gi := gi
		c, r := b.simulate(g, sb, func(e *patEdge) bool { return all(e) && side(e) == gi }, pending, r0)
		splitCost += c
		sideBound := map[string]bool{}
		for k := range sb {
			sideBound[k] = true
		}
		for _, e := range g.edges {
			if all(e) && side(e) == gi {
				sideBound[e.left], sideBound[e.right] = true, true
			}
		}
		if gi == 0 {
			accRows = r
		} else {
			j := accRows * r
			for _, v := range comp {
				if accBound[v] && sideBound[v] {
					j /= b.st.Count(b.labelsOf(g, v))
				}
			}
			accRows = j
			splitCost += j
		}
		for k := range sideBound {
			accBound[k] = true
		}
	}
	if splitCost < singleCost {
		return compPlan{split: true, rows: accRows}
	}
	return single
}
