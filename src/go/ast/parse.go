package ast

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	parser "polystore_database/src/go/parser/output"
)

// ErrUnsupported は現行文法・AST で表せない構文。呼び出し側は graph への原文委譲（verbatim）に落とす。
var ErrUnsupported = errors.New("ast: unsupported syntax")

// Parse は Cypher を AST にする。$param は従来どおり値へ置換してからパースする
// （文法にパラメータ記号が無いため。数値はそのまま、それ以外は単一引用符で囲む）。
func Parse(query string, params map[string]string) (*Query, error) {
	text, err := substituteParams(query, params)
	if err != nil {
		return nil, err
	}

	errs := &errorCounter{DefaultErrorListener: antlr.NewDefaultErrorListener()}
	lexer := parser.NewCypherLexer(antlr.NewInputStream(text))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errs)
	p := parser.NewCypherParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))
	p.RemoveErrorListeners()
	p.AddErrorListener(errs)

	tree := p.Cypher()
	if errs.n > 0 {
		return nil, fmt.Errorf("%w: %d syntax error(s)", ErrUnsupported, errs.n)
	}

	b := &builder{src: text}
	q, err := b.cypher(tree)
	if err != nil {
		return nil, err
	}
	return q, nil
}

type errorCounter struct {
	*antlr.DefaultErrorListener
	n int
}

func (e *errorCounter) SyntaxError(antlr.Recognizer, interface{}, int, int, string, antlr.RecognitionException) {
	e.n++
}

var paramRe = regexp.MustCompile(`\$(\w+)`)

func substituteParams(query string, params map[string]string) (string, error) {
	var missing []string
	out := paramRe.ReplaceAllStringFunc(query, func(m string) string {
		v, ok := params[m[1:]]
		if !ok {
			missing = append(missing, m[1:])
			return m
		}
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			return v
		}
		return "'" + strings.ReplaceAll(v, "'", "\\'") + "'"
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("ast: missing query parameters: %s", strings.Join(missing, ", "))
	}
	return out, nil
}

// builder はパース木を歩いて AST を組む。名前の無いノード・関係に連番を振る。
type builder struct {
	src  string
	nAno int
	rAno int
}

func (b *builder) unsupported(what string) error {
	return fmt.Errorf("%w: %s", ErrUnsupported, what)
}

// text は ctx が覆う元のクエリ文字列（項目名に使う。空白を保つ）。
func (b *builder) text(ctx antlr.ParserRuleContext) string {
	start, stop := ctx.GetStart(), ctx.GetStop()
	if start == nil || stop == nil {
		return ctx.GetText()
	}
	s, e := start.GetStart(), stop.GetStop()
	if s < 0 || e < s || e >= len(b.src) {
		return ctx.GetText()
	}
	return b.src[s : e+1]
}

func (b *builder) cypher(ctx parser.ICypherContext) (*Query, error) {
	stmts := ctx.AllStatement()
	if len(stmts) != 1 {
		return nil, b.unsupported("multiple statements")
	}
	st := stmts[0]
	q := &Query{}
	for _, rc := range st.AllReadingClause() {
		switch {
		case rc.MatchClause() != nil:
			m, err := b.match(rc.MatchClause(), rc.WhereClause())
			if err != nil {
				return nil, err
			}
			m.Optional = rc.OPTIONAL() != nil
			q.Clauses = append(q.Clauses, m)
		case rc.WithClause() != nil:
			w, err := b.with(rc.WithClause())
			if err != nil {
				return nil, err
			}
			q.Clauses = append(q.Clauses, w)
		}
	}
	r, err := b.ret(st.ReturnClause())
	if err != nil {
		return nil, err
	}
	q.Clauses = append(q.Clauses, r)
	return q, nil
}

// ---- MATCH ----

func (b *builder) match(mc parser.IMatchClauseContext, wc parser.IWhereClauseContext) (*MatchClause, error) {
	m := &MatchClause{}
	for _, pc := range mc.AllPattern() {
		p, err := b.pattern(pc)
		if err != nil {
			return nil, err
		}
		m.Patterns = append(m.Patterns, p)
	}
	if wc != nil {
		w, err := b.cond(wc.Condition())
		if err != nil {
			return nil, err
		}
		m.Where = w
	}
	return m, nil
}

func (b *builder) pattern(pc parser.IPatternContext) (PatternPath, error) {
	var p PatternPath
	if pc.EQ() != nil && pc.Variable() != nil {
		p.PathVar = pc.Variable().GetText()
	}
	for _, nc := range pc.AllNode() {
		n, err := b.node(nc)
		if err != nil {
			return p, err
		}
		p.Nodes = append(p.Nodes, n)
	}
	for _, rc := range pc.AllRelationship() {
		r, err := b.rel(rc)
		if err != nil {
			return p, err
		}
		p.Rels = append(p.Rels, r)
	}
	return p, nil
}

func (b *builder) node(nc parser.INodeContext) (NodePat, error) {
	var n NodePat
	if nc.Variable() != nil {
		n.Var = nc.Variable().GetText()
	} else {
		n.Var = fmt.Sprintf("_n%d", b.nAno)
		b.nAno++
	}
	if lc := nc.Labels(); lc != nil {
		for _, l := range lc.AllLabel() {
			n.Labels = append(n.Labels, l.GetText())
		}
	}
	props, err := b.props(nc.Properties())
	if err != nil {
		return n, err
	}
	n.Props = props
	return n, nil
}

func (b *builder) rel(rc parser.IRelationshipContext) (RelPat, error) {
	var r RelPat
	if rc.Variable() != nil {
		r.Var = rc.Variable().GetText()
	} else {
		r.Var = fmt.Sprintf("_r%d", b.rAno)
		b.rAno++
	}
	switch {
	case rc.RARROW() != nil:
		r.Dir = DirRight
	case rc.LARROW() != nil:
		r.Dir = DirLeft
	default:
		r.Dir = DirBoth
	}
	if tc := rc.Types(); tc != nil {
		for _, id := range tc.AllIDENTIFIER() {
			r.Types = append(r.Types, id.GetText())
		}
	}
	if rg := rc.Range_(); rg != nil {
		h := HopRange{}
		if lit := rg.RangeLiteral(); lit != nil {
			nums := lit.AllNUMBER()
			atoi := func(t antlr.TerminalNode) *int {
				v, _ := strconv.Atoi(t.GetText())
				return &v
			}
			switch {
			case lit.DOUBLE_DOT() == nil: // *n
				h.Min, h.Max = atoi(nums[0]), atoi(nums[0])
			case len(nums) == 2: // *a..b
				h.Min, h.Max = atoi(nums[0]), atoi(nums[1])
			case len(nums) == 1 && lit.GetStart().GetTokenType() == parser.CypherLexerDOUBLE_DOT: // *..b
				h.Max = atoi(nums[0])
			case len(nums) == 1: // *a..
				h.Min = atoi(nums[0])
			}
		}
		r.Range = &h
	}
	props, err := b.props(rc.Properties())
	if err != nil {
		return r, err
	}
	r.Props = props
	return r, nil
}

func (b *builder) props(pc parser.IPropertiesContext) ([]PropEq, error) {
	if pc == nil {
		return nil, nil
	}
	var out []PropEq
	for _, p := range pc.AllProperty() {
		v, err := b.value(p.Value())
		if err != nil {
			return nil, err
		}
		out = append(out, PropEq{Key: p.IDENTIFIER().GetText(), Value: v})
	}
	return out, nil
}

// ---- WITH / RETURN ----

func (b *builder) with(wc parser.IWithClauseContext) (*ProjectionClause, error) {
	pc := &ProjectionClause{Distinct: wc.DISTINCT() != nil}
	if err := b.projBody(pc, wc.ReturnItems(), wc.OrderItems(), wc.LimitNum()); err != nil {
		return nil, err
	}
	if w := wc.WhereClause(); w != nil {
		e, err := b.cond(w.Condition())
		if err != nil {
			return nil, err
		}
		pc.Where = e
	}
	return pc, nil
}

func (b *builder) ret(rc parser.IReturnClauseContext) (*ProjectionClause, error) {
	pc := &ProjectionClause{IsReturn: true}
	if err := b.projBody(pc, rc.ReturnItems(), rc.OrderItems(), rc.LimitNum()); err != nil {
		return nil, err
	}
	return pc, nil
}

func (b *builder) projBody(pc *ProjectionClause, items parser.IReturnItemsContext,
	order parser.IOrderItemsContext, limit parser.ILimitNumContext) error {
	for _, ic := range items.AllReturnItem() {
		it, err := b.projItem(ic)
		if err != nil {
			return err
		}
		pc.Items = append(pc.Items, it)
	}
	if order != nil {
		for _, oc := range order.AllOrderItem() {
			e, err := b.expression(oc.Expression())
			if err != nil {
				return err
			}
			pc.OrderBy = append(pc.OrderBy, SortItem{Expr: e, Desc: oc.DESC() != nil})
		}
	}
	if limit != nil {
		n, err := strconv.Atoi(limit.NUMBER().GetText())
		if err != nil {
			return b.unsupported("non-integer LIMIT")
		}
		pc.Limit = &n
	}
	return nil
}

func (b *builder) projItem(ic parser.IReturnItemContext) (ProjItem, error) {
	var (
		e    Expr
		err  error
		body antlr.ParserRuleContext
	)
	switch {
	case ic.AggregateFunc() != nil:
		body = ic.AggregateFunc()
		e, err = b.aggregate(ic.AggregateFunc())
	case ic.CaseExpr() != nil:
		body = ic.CaseExpr()
		e, err = b.caseExpr(ic.CaseExpr())
	case ic.ValueExpr() != nil:
		body = ic.ValueExpr()
		e, err = b.valueExpr(ic.ValueExpr())
	case ic.COALESCE() != nil:
		body = ic
		var args []Expr
		for _, x := range ic.AllExpression() {
			a, aerr := b.expression(x)
			if aerr != nil {
				return ProjItem{}, aerr
			}
			args = append(args, a)
		}
		e = &Call{Fn: "coalesce", Args: args}
	default:
		return ProjItem{}, b.unsupported("return item")
	}
	if err != nil {
		return ProjItem{}, err
	}
	it := ProjItem{Expr: e}
	if ic.AS() != nil && ic.Variable() != nil {
		it.Name, it.Alias = ic.Variable().GetText(), true
	} else if ic.COALESCE() != nil {
		it.Name = strings.TrimSpace(b.text(ic))
	} else {
		it.Name = strings.TrimSpace(b.text(body))
	}
	return it, nil
}

func (b *builder) aggregate(ac parser.IAggregateFuncContext) (Expr, error) {
	var fn string
	switch {
	case ac.COUNT() != nil:
		fn = "count"
	case ac.SUM() != nil:
		fn = "sum"
	case ac.AVG() != nil:
		fn = "avg"
	case ac.MIN() != nil:
		fn = "min"
	case ac.MAX() != nil:
		fn = "max"
	}
	a := &Agg{Fn: fn, Distinct: ac.DISTINCT() != nil}
	if arg := ac.AggArg(); arg != nil {
		ids := arg.AllIDENTIFIER()
		if len(ids) == 1 {
			a.Arg = &VarRef{Name: ids[0].GetText()}
		} else {
			a.Arg = &PropRef{Var: ids[0].GetText(), Prop: ids[1].GetText()}
		}
	}
	return a, nil
}

func (b *builder) caseExpr(cc parser.ICaseExprContext) (Expr, error) {
	c := &Case{}
	switch x := cc.(type) {
	case *parser.CaseSearchedContext:
		conds, results := x.AllCondition(), x.AllCaseResult()
		for i, cond := range conds {
			w, err := b.cond(cond)
			if err != nil {
				return nil, err
			}
			t, err := b.caseResult(results[i])
			if err != nil {
				return nil, err
			}
			c.Whens = append(c.Whens, CaseWhen{When: w, Then: t})
		}
		if x.ELSE() != nil && len(results) > len(conds) {
			e, err := b.caseResult(results[len(conds)])
			if err != nil {
				return nil, err
			}
			c.Else = e
		}
		return c, nil
	case *parser.CaseSimpleContext:
		subj, err := b.valueExpr(x.ValueExpr())
		if err != nil {
			return nil, err
		}
		c.Subject = subj
		keys, results := x.AllCaseKey(), x.AllCaseResult()
		for i, k := range keys {
			kv, err := b.caseKey(k)
			if err != nil {
				return nil, err
			}
			t, err := b.caseResult(results[i])
			if err != nil {
				return nil, err
			}
			c.Whens = append(c.Whens, CaseWhen{When: kv, Then: t})
		}
		if x.ELSE() != nil && len(results) > len(keys) {
			e, err := b.caseResult(results[len(keys)])
			if err != nil {
				return nil, err
			}
			c.Else = e
		}
		return c, nil
	}
	return nil, b.unsupported("case expression")
}

func (b *builder) caseKey(kc parser.ICaseKeyContext) (Expr, error) {
	switch {
	case kc.Value() != nil:
		return b.value(kc.Value())
	case kc.NULL() != nil:
		return &Lit{Val: nil}, nil
	case kc.TRUE() != nil:
		return &Lit{Val: true}, nil
	case kc.FALSE() != nil:
		return &Lit{Val: false}, nil
	}
	return b.expression(kc.Expression())
}

func (b *builder) caseResult(rc parser.ICaseResultContext) (Expr, error) {
	switch {
	case rc.Value() != nil:
		return b.value(rc.Value())
	case rc.NULL() != nil:
		return &Lit{Val: nil}, nil
	case rc.TRUE() != nil:
		return &Lit{Val: true}, nil
	case rc.FALSE() != nil:
		return &Lit{Val: false}, nil
	}
	return b.expression(rc.Expression())
}

func (b *builder) valueExpr(vc parser.IValueExprContext) (Expr, error) {
	switch c := vc.(type) {
	case *parser.VeNegContext:
		x, err := b.valueExpr(c.ValueExpr())
		if err != nil {
			return nil, err
		}
		return &Un{Op: "-", X: x}, nil
	case *parser.VeMulDivContext:
		return b.binVE(c.GetOp().GetText(), c.ValueExpr(0), c.ValueExpr(1))
	case *parser.VeAddSubContext:
		return b.binVE(c.GetOp().GetText(), c.ValueExpr(0), c.ValueExpr(1))
	case *parser.VeCallContext:
		call := &Call{Fn: strings.ToLower(c.IDENTIFIER().GetText())}
		for _, a := range c.AllValueExpr() {
			x, err := b.valueExpr(a)
			if err != nil {
				return nil, err
			}
			call.Args = append(call.Args, x)
		}
		return call, nil
	case *parser.VePropRefContext:
		return b.ref(c.AllIDENTIFIER())
	case *parser.VeListContext:
		l := &List{}
		for _, a := range c.AllValueExpr() {
			x, err := b.valueExpr(a)
			if err != nil {
				return nil, err
			}
			l.Elems = append(l.Elems, x)
		}
		return l, nil
	case *parser.VeLitContext:
		if c.STRING() != nil {
			return &Lit{Val: unquote(c.STRING().GetText())}, nil
		}
		return numberLit(c.NUMBER().GetText()), nil
	case *parser.VeParenContext:
		return b.valueExpr(c.ValueExpr())
	}
	return nil, b.unsupported("value expression")
}

func (b *builder) binVE(op string, l, r parser.IValueExprContext) (Expr, error) {
	le, err := b.valueExpr(l)
	if err != nil {
		return nil, err
	}
	re, err := b.valueExpr(r)
	if err != nil {
		return nil, err
	}
	return &Bin{Op: op, L: le, R: re}, nil
}

func (b *builder) ref(ids []antlr.TerminalNode) (Expr, error) {
	switch len(ids) {
	case 1:
		return &VarRef{Name: ids[0].GetText()}, nil
	case 2:
		return &PropRef{Var: ids[0].GetText(), Prop: ids[1].GetText()}, nil
	}
	return nil, b.unsupported("nested property access")
}

// ---- 条件 ----

func (b *builder) cond(cc parser.IConditionContext) (Expr, error) {
	switch c := cc.(type) {
	case *parser.ConditionParenContext:
		return b.cond(c.Condition())
	case *parser.ConditionNotContext:
		x, err := b.cond(c.Condition())
		if err != nil {
			return nil, err
		}
		return &Un{Op: "NOT", X: x}, nil
	case *parser.ConditionAndContext:
		return b.binCond("AND", c.Condition(0), c.Condition(1))
	case *parser.ConditionOrContext:
		return b.binCond("OR", c.Condition(0), c.Condition(1))
	case *parser.ConditionAllContext:
		return b.quant("all", c.Variable(), c.Expression(), c.Condition())
	case *parser.ConditionAnyContext:
		return b.quant("any", c.Variable(), c.Expression(), c.Condition())
	case *parser.ConditionNoneContext:
		return b.quant("none", c.Variable(), c.Expression(), c.Condition())
	case *parser.ConditionSingleContext:
		return b.quant("single", c.Variable(), c.Expression(), c.Condition())
	case *parser.ConditionIsNullContext:
		x, err := b.expression(c.Expression())
		if err != nil {
			return nil, err
		}
		return &Un{Op: "IS NULL", X: x}, nil
	case *parser.ConditionIsNotNullContext:
		x, err := b.expression(c.Expression())
		if err != nil {
			return nil, err
		}
		return &Un{Op: "IS NOT NULL", X: x}, nil
	case *parser.ConditionInContext:
		x, err := b.expression(c.Expression())
		if err != nil {
			return nil, err
		}
		l := &List{}
		for _, v := range c.List().AllValue() {
			lv, err := b.value(v)
			if err != nil {
				return nil, err
			}
			l.Elems = append(l.Elems, lv)
		}
		return &Bin{Op: "IN", L: x, R: l}, nil
	case *parser.ConditionEqualityContext:
		return b.cmpValue("=", c.Expression(), c.Value())
	case *parser.ConditionNotEqualityContext:
		return b.cmpValue("<>", c.Expression(), c.Value())
	case *parser.ConditionGreaterContext:
		return b.cmpValue(">", c.Expression(), c.Value())
	case *parser.ConditionLessContext:
		return b.cmpValue("<", c.Expression(), c.Value())
	case *parser.ConditionGreaterEqualContext:
		return b.cmpValue(">=", c.Expression(), c.Value())
	case *parser.ConditionLessEqualContext:
		return b.cmpValue("<=", c.Expression(), c.Value())
	case *parser.ConditionVarEqContext:
		return b.cmpExpr("=", c.Expression(0), c.Expression(1))
	case *parser.ConditionVarNeqContext:
		return b.cmpExpr("<>", c.Expression(0), c.Expression(1))
	case *parser.ConditionVarGreaterContext:
		return b.cmpExpr(">", c.Expression(0), c.Expression(1))
	case *parser.ConditionVarLessContext:
		return b.cmpExpr("<", c.Expression(0), c.Expression(1))
	case *parser.ConditionVarGreaterEqualContext:
		return b.cmpExpr(">=", c.Expression(0), c.Expression(1))
	case *parser.ConditionVarLessEqualContext:
		return b.cmpExpr("<=", c.Expression(0), c.Expression(1))
	}
	return nil, b.unsupported("condition")
}

func (b *builder) binCond(op string, l, r parser.IConditionContext) (Expr, error) {
	le, err := b.cond(l)
	if err != nil {
		return nil, err
	}
	re, err := b.cond(r)
	if err != nil {
		return nil, err
	}
	return &Bin{Op: op, L: le, R: re}, nil
}

func (b *builder) quant(kind string, v parser.IVariableContext, list parser.IExpressionContext, pred parser.IConditionContext) (Expr, error) {
	le, err := b.expression(list)
	if err != nil {
		return nil, err
	}
	pe, err := b.cond(pred)
	if err != nil {
		return nil, err
	}
	return &Quant{Kind: kind, Var: v.GetText(), List: le, Pred: pe}, nil
}

func (b *builder) cmpValue(op string, l parser.IExpressionContext, r parser.IValueContext) (Expr, error) {
	le, err := b.expression(l)
	if err != nil {
		return nil, err
	}
	re, err := b.value(r)
	if err != nil {
		return nil, err
	}
	return &Bin{Op: op, L: le, R: re}, nil
}

func (b *builder) cmpExpr(op string, l, r parser.IExpressionContext) (Expr, error) {
	le, err := b.expression(l)
	if err != nil {
		return nil, err
	}
	re, err := b.expression(r)
	if err != nil {
		return nil, err
	}
	return &Bin{Op: op, L: le, R: re}, nil
}

// expression は `a` / `a.p` / `f(v)`（length(p)・nodes(p)・relationships(p) など）。
func (b *builder) expression(ec parser.IExpressionContext) (Expr, error) {
	if ec.LPAREN() != nil {
		return &Call{
			Fn:   strings.ToLower(ec.IDENTIFIER(0).GetText()),
			Args: []Expr{&VarRef{Name: ec.Variable().GetText()}},
		}, nil
	}
	return b.ref(ec.AllIDENTIFIER())
}

func (b *builder) value(vc parser.IValueContext) (Expr, error) {
	if vc.STRING() != nil {
		return &Lit{Val: unquote(vc.STRING().GetText())}, nil
	}
	if vc.NUMBER() != nil {
		return numberLit(vc.NUMBER().GetText()), nil
	}
	return nil, b.unsupported("value")
}

func numberLit(s string) Expr {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &Lit{Val: i}
	}
	f, _ := strconv.ParseFloat(s, 64)
	return &Lit{Val: f}
}

// unquote は '…' / "…" を外し、バックスラッシュエスケープを解く。
func unquote(s string) string {
	if len(s) >= 2 {
		s = s[1 : len(s)-1]
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
