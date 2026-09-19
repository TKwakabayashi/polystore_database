// antlerのListenr用メソッドの実装および補助関数の定義
//
// パーサーの定義ファイル変更時に編集する
package planner

import (
	"fmt"
	parser "polystore_database/src/go/parser/output"
	plan "polystore_database/src/go/plan"
	"strconv"
	"strings"
)

// EnterWithClause は WITH 節（多段クエリ）の開始。以降 returnItems/orderItems/limit/whereClause は
// 現 WITH ステージのバッファへ振り分ける（inWith=true）。
func (l *QueryPlannerListener) EnterWithClause(ctx *parser.WithClauseContext) {
	l.hasWith = true
	l.inWith = true
	if ctx.DISTINCT() != nil {
		l.withDistinct = true
	}
}

// ExitWithClause は WITH 節の終端。現ステージ（record パイプライン状態＋WITH 項目＋post-WHERE）を
// スナップショットして stages へ積み、instance の record 状態をリセットして次ステージを
// 「WITH の束縛項目を RowSource から供給される carried 束縛」で種付けする。
func (l *QueryPlannerListener) ExitWithClause(ctx *parser.WithClauseContext) {
	l.snapshotStage()
	l.inWith = false
}

// EnterPattern はパターン先頭の p= 束縛（パス変数）を捕捉する。
// pattern: (variable EQ)? node (relationship node)* の variable が path var。
func (l *QueryPlannerListener) EnterPattern(ctx *parser.PatternContext) {
	l.pendingPathVar = ""
	if ctx.EQ() != nil && ctx.Variable() != nil {
		l.pendingPathVar = ctx.Variable().GetText()
	}
}

// EnterNode is called when production node is entered.
func (l *QueryPlannerListener) EnterNode(ctx *parser.NodeContext) {
	variable := ""
	if ctx.Variable() != nil {
		variable = ctx.Variable().GetText()
	}

	labels := []string{}
	if ctx.Labels() != nil {
		for _, label := range ctx.Labels().AllLabel() {
			labels = append(labels, label.IDENTIFIER().GetText())
		}
	}

	// 1. エイリアスがない場合は自動生成
	if variable == "" {
		variable = fmt.Sprintf("Nalias_%d", l.varEntCounter)
		l.varEntCounter++
	}

	// 2. EntityInfoの登録/更新
	if _, exists := l.symbolEntTable[variable]; !exists {
		info := &EntityInfo{alias: variable, labels: labels}
		l.entityInfo = append(l.entityInfo, info)
		l.symbolEntTable[variable] = len(l.entityInfo) - 1
	} else {
		idx := l.symbolEntTable[variable]
		if len(l.entityInfo[idx].labels) == 0 && len(labels) > 0 {
			l.entityInfo[idx].labels = labels
		}
	}

	// 3. インラインプロパティの処理
	if ctx.Properties() != nil {
		// properties: LCURLY property (COMMA property)* RCURLY のケース
		for _, prop := range ctx.Properties().AllProperty() {
			valStr := l.parseValue(prop.Value())
			l.symbolCondMapping[variable] = append(l.symbolCondMapping[variable], &plan.ConditionNode{
				Type:     plan.CondEq,
				Labels:   labels,
				Alias:    variable,
				Property: prop.IDENTIFIER().GetText(),
				Value:    valStr,
			})
		}
		// properties: whereClause のケースがある場合は ExitWhereClause で処理されるためここでは不要
	}

	// 4. 重要：未完了のリレーションシップのターゲットとしてこのノードを確定させる
	if len(l.relInfo) > 0 {
		lastRel := l.relInfo[len(l.relInfo)-1]
		if lastRel.target == "" {
			lastRel.target = variable
			// sourceEntMapping は EnterRelationship で設定済み。ここでは target 側を登録。
			l.targetEntMapping[variable] = append(l.targetEntMapping[variable], len(l.relInfo)-1)
		}
	}

	l.lastNodeAlias = variable
}

// EnterRelationship is called when production relationship is entered.
func (l *QueryPlannerListener) EnterRelationship(ctx *parser.RelationshipContext) {
	rel := &RelInfo{minHops: -1, maxHops: -1}

	variable := ""
	if ctx.Variable() != nil {
		variable = ctx.Variable().GetText()
	}
	label := ""
	if ctx.Types() != nil && len(ctx.Types().AllIDENTIFIER()) > 0 {
		label = ctx.Types().IDENTIFIER(0).GetText()
	}

	// assign alias if not
	if variable == "" {
		variable = fmt.Sprintf("Ealias_%d", l.varRelCounter)
		l.varRelCounter++
	}

	dir := plan.Bidirectional
	if ctx.RARROW() != nil {
		dir = plan.Outgoing
	}
	if ctx.LARROW() != nil {
		dir = plan.Incoming
	}

	rel.alias = variable
	rel.label = label
	rel.dir = dir

	rel.source = l.lastNodeAlias

	if ctx.Range_() != nil {
		rel.isVarLength = true
		rngCtx := ctx.Range_().(*parser.RangeContext)
		rel = parseRange(rngCtx, rel)

		// パス束縛 p=(a)-[:R*..]-(b) の p を、この可変長 rel へ対応付ける。
		// パス全体フィルタ（all/any/none/single・length(p)）の紐付けに使う。
		if l.pendingPathVar != "" {
			l.varLenRelByPathVar[l.pendingPathVar] = variable
		}
	}

	l.relInfo = append(l.relInfo, rel)
	l.symbolRelTable[variable] = len(l.relInfo) - 1

	l.sorceEntMapping[l.lastNodeAlias] = append(l.sorceEntMapping[l.lastNodeAlias], len(l.relInfo)-1)

}

// EnterReturnItem is called when production returnItem is entered.
func (l *QueryPlannerListener) EnterReturnItem(ctx *parser.ReturnItemContext) {
	asName := ""
	if ctx.AS() != nil && ctx.Variable() != nil {
		asName = ctx.Variable().GetText()
	}

	// CASE 式・valueExpr（計算式）は子の解析後に組む必要があるため ExitReturnItem で処理する。
	if ctx.CaseExpr() != nil || ctx.ValueExpr() != nil {
		return
	}

	// 1) 集約関数 count/sum/avg/min/max
	if agg := ctx.AggregateFunc(); agg != nil {
		item := l.parseAggregate(agg.(*parser.AggregateFuncContext))
		if asName != "" {
			item.OutName = asName
		}
		l.addProjItem(plan.ReturnItem{
			Name:        item.OutName,
			Alias:       item.Alias,
			IsAggregate: true,
			Agg:         &item,
		})
		return
	}

	// 2) coalesce(...)
	if ctx.COALESCE() != nil {
		var props []string
		var alias string
		for _, exprCtx := range ctx.AllExpression() {
			if exprCtx != nil {
				parts := strings.Split(exprCtx.GetText(), ".")
				if len(parts) == 2 {
					if alias == "" {
						alias = parts[0]
					}
					props = append(props, parts[1])
				}
			}
		}
		name := ctx.GetText()
		if asName != "" {
			name = asName
		}
		l.addProjItem(plan.ReturnItem{Name: name, Alias: alias, Props: props, IsCoalesce: true})
		return
	}

	// 3) 通常の式 alias.prop
	exprCtx := ctx.Expression(0)
	if exprCtx != nil {
		txt := exprCtx.GetText()
		parts := strings.Split(txt, ".")
		name := txt
		if asName != "" {
			name = asName
		}
		if len(parts) == 2 {
			l.addProjItem(plan.ReturnItem{Name: name, Alias: parts[0], Props: []string{parts[1]}, IsCoalesce: false})
		} else {
			l.addProjItem(plan.ReturnItem{Name: name, Alias: parts[0], Props: []string{""}, IsCoalesce: false})
		}
	}
}

// addProjItem は projection 項目を現在のクローズ先へ振り分ける。
// WITH 節の中なら現 WITH ステージの項目バッファへ、そうでなければ最終 RETURN の項目へ。
func (l *QueryPlannerListener) addProjItem(item plan.ReturnItem) {
	if l.inWith {
		l.withItemsBuf = append(l.withItemsBuf, item)
	} else {
		l.returnItems = append(l.returnItems, item)
	}
}

// ExitReturnItem は CASE 式の項目を組み立てる（WHEN 条件は stackCond に積まれている）。
// searched CASE のみ対応: CASE (WHEN condition THEN result)+ (ELSE result)? END。
func (l *QueryPlannerListener) ExitReturnItem(ctx *parser.ReturnItemContext) {
	// valueExpr 項目（算術/関数/list 計算式・単純プロパティ参照）。
	if ve := ctx.ValueExpr(); ve != nil {
		l.exitValueExprItem(ctx, ve)
		return
	}
	ce := ctx.CaseExpr()
	if ce == nil {
		return
	}
	cec := ce.(*parser.CaseExprContext)
	n := len(cec.AllWHEN())

	// WHEN 条件を stackCond から取り出す（積まれた順の逆順で pop）。
	conds := make([]*plan.ConditionNode, n)
	for i := n - 1; i >= 0; i-- {
		conds[i] = l.popCond()
	}
	results := cec.AllCaseResult()
	hasElse := cec.ELSE() != nil

	var whens []plan.CaseWhen
	for i := 0; i < n; i++ {
		w, err := condToExpr(conds[i])
		if err != nil {
			l.exprBuildFailed = true // 未対応 WHEN（変数間比較等）→ verbatim フォールバック
			return
		}
		whens = append(whens, plan.CaseWhen{When: w, Then: l.caseResultExpr(results[i])})
	}
	var elseE plan.Expr
	if hasElse {
		elseE = l.caseResultExpr(results[n])
	}

	name := ""
	if ctx.AS() != nil && ctx.Variable() != nil {
		name = ctx.Variable().GetText()
	}
	if name == "" {
		name = ctx.GetText()
	}
	l.addProjItem(plan.ReturnItem{Name: name, Expr: &plan.Case{Whens: whens, Else: elseE}})
}

// caseResultExpr は CASE の THEN/ELSE 結果（value または expression）を Expr へ変換する。
func (l *QueryPlannerListener) caseResultExpr(cr parser.ICaseResultContext) plan.Expr {
	c := cr.(*parser.CaseResultContext)
	if c.Value() != nil {
		return litOf(l.parseValue(c.Value()))
	}
	if c.Expression() != nil {
		a, p := l.parseExpression(c.Expression())
		return colExpr(a, p)
	}
	return &plan.Lit{Val: nil}
}

func (l *QueryPlannerListener) parseAggregate(ctx *parser.AggregateFuncContext) plan.AggregateItem {
	var fn plan.AggFunc
	switch {
	case ctx.COUNT() != nil:
		fn = plan.AggCount
	case ctx.SUM() != nil:
		fn = plan.AggSum
	case ctx.AVG() != nil:
		fn = plan.AggAvg
	case ctx.MIN() != nil:
		fn = plan.AggMin
	case ctx.MAX() != nil:
		fn = plan.AggMax
	}

	item := plan.AggregateItem{Func: fn, Distinct: ctx.DISTINCT() != nil, OutName: ctx.GetText()}

	if ctx.STAR() != nil {
		return item // count(*)
	}
	if arg := ctx.AggArg(); arg != nil {
		ac := arg.(*parser.AggArgContext)
		idents := ac.AllIDENTIFIER()
		if len(idents) > 0 {
			item.Alias = idents[0].GetText()
		}
		if len(idents) > 1 {
			item.Prop = idents[1].GetText()
		}
	}
	return item
}

// EnterOrderItems is called when production orderItems is entered.
func (l *QueryPlannerListener) EnterOrderItem(ctx *parser.OrderItemContext) {
	exprCtx := ctx.Expression()
	var dir plan.OrderDir
	if ctx.ASC() != nil {
		dir = plan.OrderAsc
	} else if ctx.DESC() != nil {
		dir = plan.OrderDesc
	} else {
		fmt.Println("invalid order direction")
	}
	if exprCtx == nil {
		return
	}
	txt := exprCtx.GetText()
	parts := strings.Split(txt, ".")
	var oi plan.OrderItem
	if len(parts) == 2 {
		oi = plan.OrderItem{Alias: parts[0], Prop: parts[1], Direction: dir, Key: parts[0] + "." + parts[1]}
	} else {
		oi = plan.OrderItem{Key: txt, Direction: dir} // 集約別名など単一トークン
	}
	if l.inWith {
		l.withOrderBuf = append(l.withOrderBuf, oi)
	} else {
		l.orderItems = append(l.orderItems, oi)
	}
}

// EnterLimitNum is called when production limitNum is entered.
func (l *QueryPlannerListener) EnterLimitNum(ctx *parser.LimitNumContext) {
	n, _ := strconv.Atoi(ctx.NUMBER().GetText())
	if l.inWith {
		l.withLimitBuf = n
	} else {
		l.limitNum = n
	}
}

func (l *QueryPlannerListener) ExitWhereClause(ctx *parser.WhereClauseContext) {
	if len(l.stackCond) == 0 {
		return
	}

	rootCond := l.popCond()

	// WITH 節の末尾 WHERE（post-WITH フィルタ）は Row 世界の RowFilter として扱う。
	// record 世界の symbolCondMapping には入れず、現 WITH ステージ確定時に取り出す（curPostWhere）。
	if l.inWith {
		l.curPostWhere = rootCond
		return
	}

	condList := plan.DecomposeOuterAndOp(rootCond)

	for _, cond := range condList {
		l.routeCondition(cond)
	}
}

// routeCondition は分解済みの葉条件を種別に応じて振り分ける:
//   - CondVarCmp（変数間比較）→ crossFilters（PostFilter として積む）
//   - 量化子（all/any/none/single）→ 可変長パスの要素述語
//   - PathFunc 付き（length(p) …）→ 可変長パスの大域述語
//   - それ以外（単一エイリアス述語）→ symbolCondMapping
func (l *QueryPlannerListener) routeCondition(cond *plan.ConditionNode) {
	switch {
	case cond.Type == plan.CondVarCmp:
		l.crossFilters = append(l.crossFilters, cond)
	case cond.Type.IsQuantifier():
		l.attachPathElementFilter(cond)
	case cond.PathFunc != "":
		l.attachGlobalPathFilter(cond)
	default:
		if cond.Alias != "" {
			l.symbolCondMapping[cond.Alias] = append(l.symbolCondMapping[cond.Alias], cond)
		}
	}
}

// attachPathElementFilter は量化子 ConditionNode（Alias=path var, Property=nodes/relationships,
// Value=反復変数, Child=内側述語）を VarLengthFilter へ変換し、対応する可変長 rel へ紐付ける。
func (l *QueryPlannerListener) attachPathElementFilter(cond *plan.ConditionNode) {
	relAlias, ok := l.varLenRelByPathVar[cond.Alias]
	if !ok {
		return // path var が可変長 rel に束縛されていない → 無視（安全側）
	}
	target := plan.TargetNode
	if cond.Property == "relationships" {
		target = plan.TargetRelationship
	}
	vlf := plan.VarLengthFilter{
		Phase:   plan.PhaseStep,
		Target:  target,
		Quant:   cond.Type,
		IterVar: cond.Value,
		Cond:    cond.Child,
	}
	l.pathFiltersByRelAlias[relAlias] = append(l.pathFiltersByRelAlias[relAlias], vlf)
}

// attachGlobalPathFilter は length(p) OP val（PathFunc 付き ConditionNode）を
// VarLengthFilter(TargetGlobalPath) へ変換して紐付ける。
func (l *QueryPlannerListener) attachGlobalPathFilter(cond *plan.ConditionNode) {
	relAlias, ok := l.varLenRelByPathVar[cond.Alias]
	if !ok {
		return
	}
	vlf := plan.VarLengthFilter{
		Phase:  plan.PhasePost,
		Target: plan.TargetGlobalPath,
		Func:   cond.PathFunc,
		CmpOp:  cond.Type,
		Value:  cond.Value,
	}
	l.pathFiltersByRelAlias[relAlias] = append(l.pathFiltersByRelAlias[relAlias], vlf)
}

// pushValueComparison は「expression OP value」をプッシュする共通処理。
// length(p) などのパス関数式は大域パス述語（PathFunc 付き）として扱う。
func (l *QueryPlannerListener) pushValueComparison(t plan.ConditionType, expr parser.IExpressionContext, val parser.IValueContext) {
	if fn, pathVar, ok := parseExprFunc(expr); ok {
		// length(p) OP val → 大域パス述語。他のパス関数は現状 length のみ対応。
		l.pushCond(&plan.ConditionNode{Type: t, PathFunc: fn, Alias: pathVar, Value: l.parseValue(val)})
		return
	}
	alias, prop := l.parseExpression(expr)
	c := &plan.ConditionNode{Type: t, Alias: alias, Property: prop, Value: l.parseValue(val)}
	if idx, ok := l.symbolEntTable[alias]; ok {
		c.Labels = l.entityInfo[idx].labels
	}
	l.pushCond(c)
}

// pushVarComparison は「expression OP expression」（クロスエイリアス比較）をプッシュする。
func (l *QueryPlannerListener) pushVarComparison(op plan.ConditionType, lExpr, rExpr parser.IExpressionContext) {
	la, lp := l.parseExpression(lExpr)
	ra, rp := l.parseExpression(rExpr)
	l.pushCond(&plan.ConditionNode{
		Type: plan.CondVarCmp, Alias: la, Property: lp, RAlias: ra, RProperty: rp, CmpOp: op,
	})
}

func (l *QueryPlannerListener) ExitConditionEquality(ctx *parser.ConditionEqualityContext) {
	l.pushValueComparison(plan.CondEq, ctx.Expression(), ctx.Value())
}

func (l *QueryPlannerListener) ExitConditionGreater(ctx *parser.ConditionGreaterContext) {
	l.pushValueComparison(plan.CondGreater, ctx.Expression(), ctx.Value())
}

func (l *QueryPlannerListener) ExitConditionLess(ctx *parser.ConditionLessContext) {
	l.pushValueComparison(plan.CondLess, ctx.Expression(), ctx.Value())
}

func (l *QueryPlannerListener) ExitConditionGreaterEqual(ctx *parser.ConditionGreaterEqualContext) {
	l.pushValueComparison(plan.CondGreaterEq, ctx.Expression(), ctx.Value())
}

func (l *QueryPlannerListener) ExitConditionLessEqual(ctx *parser.ConditionLessEqualContext) {
	l.pushValueComparison(plan.CondLessEq, ctx.Expression(), ctx.Value())
}

func (l *QueryPlannerListener) ExitConditionNotEquality(ctx *parser.ConditionNotEqualityContext) {
	l.pushValueComparison(plan.CondNeq, ctx.Expression(), ctx.Value())
}

// ---- 新規: IN リスト / IS NULL / IS NOT NULL ----

func (l *QueryPlannerListener) ExitConditionIn(ctx *parser.ConditionInContext) {
	alias, prop := l.parseExpression(ctx.Expression())
	var vals []string
	if lst := ctx.List(); lst != nil {
		for _, v := range lst.(*parser.ListContext).AllValue() {
			vals = append(vals, l.parseValue(v))
		}
	}
	l.pushCond(&plan.ConditionNode{Type: plan.CondIn, Alias: alias, Property: prop, Values: vals})
}

func (l *QueryPlannerListener) ExitConditionIsNull(ctx *parser.ConditionIsNullContext) {
	alias, prop := l.parseExpression(ctx.Expression())
	l.pushCond(&plan.ConditionNode{Type: plan.CondIsNull, Alias: alias, Property: prop})
}

func (l *QueryPlannerListener) ExitConditionIsNotNull(ctx *parser.ConditionIsNotNullContext) {
	alias, prop := l.parseExpression(ctx.Expression())
	l.pushCond(&plan.ConditionNode{Type: plan.CondIsNotNull, Alias: alias, Property: prop})
}

// ---- 新規: 変数間比較（a.x OP b.y）----

func (l *QueryPlannerListener) ExitConditionVarEq(ctx *parser.ConditionVarEqContext) {
	l.pushVarComparison(plan.CondEq, ctx.Expression(0), ctx.Expression(1))
}

func (l *QueryPlannerListener) ExitConditionVarNeq(ctx *parser.ConditionVarNeqContext) {
	l.pushVarComparison(plan.CondNeq, ctx.Expression(0), ctx.Expression(1))
}

func (l *QueryPlannerListener) ExitConditionVarGreater(ctx *parser.ConditionVarGreaterContext) {
	l.pushVarComparison(plan.CondGreater, ctx.Expression(0), ctx.Expression(1))
}

func (l *QueryPlannerListener) ExitConditionVarLess(ctx *parser.ConditionVarLessContext) {
	l.pushVarComparison(plan.CondLess, ctx.Expression(0), ctx.Expression(1))
}

func (l *QueryPlannerListener) ExitConditionVarGreaterEqual(ctx *parser.ConditionVarGreaterEqualContext) {
	l.pushVarComparison(plan.CondGreaterEq, ctx.Expression(0), ctx.Expression(1))
}

func (l *QueryPlannerListener) ExitConditionVarLessEqual(ctx *parser.ConditionVarLessEqualContext) {
	l.pushVarComparison(plan.CondLessEq, ctx.Expression(0), ctx.Expression(1))
}

func (l *QueryPlannerListener) ExitConditionParen(ctx *parser.ConditionParenContext) {
	child := l.popCond()
	c := &plan.ConditionNode{Type: plan.CondParen, Child: child}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionAnd(ctx *parser.ConditionAndContext) {
	right := l.popCond()
	left := l.popCond()
	c := &plan.ConditionNode{Type: plan.CondAnd, Left: left, Right: right}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionOr(ctx *parser.ConditionOrContext) {
	right := l.popCond()
	left := l.popCond()
	c := &plan.ConditionNode{Type: plan.CondOr, Left: left, Right: right}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionNot(ctx *parser.ConditionNotContext) {
	child := l.popCond()
	c := &plan.ConditionNode{Type: plan.CondNot, Child: child}
	l.pushCond(c)
}

// pushQuantifier は all/any/none/single(iterVar IN nodes(p)|relationships(p) WHERE inner)
// を ConditionNode（Alias=path var, Property=関数名, Value=反復変数, Child=内側述語）へ変換する。
func (l *QueryPlannerListener) pushQuantifier(quant plan.ConditionType, iterVar string, exprCtx parser.IExpressionContext) {
	innerCond := l.popCond()

	var listAlias, listFunc string
	if exprCtx.LPAREN() != nil { // nodes(p) / relationships(p)
		listFunc = exprCtx.IDENTIFIER(0).GetText()
		listAlias = exprCtx.Variable().GetText()
	} else {
		listAlias = exprCtx.GetText()
	}

	l.pushCond(&plan.ConditionNode{
		Type:     quant,
		Alias:    listAlias, // path var "p"
		Property: listFunc,  // "nodes" / "relationships"
		Value:    iterVar,   // 反復変数 "x"
		Child:    innerCond,
	})
}

func (l *QueryPlannerListener) ExitConditionAll(ctx *parser.ConditionAllContext) {
	l.pushQuantifier(plan.CondAll, ctx.Variable().GetText(), ctx.Expression())
}

func (l *QueryPlannerListener) ExitConditionAny(ctx *parser.ConditionAnyContext) {
	l.pushQuantifier(plan.CondAny, ctx.Variable().GetText(), ctx.Expression())
}

func (l *QueryPlannerListener) ExitConditionNone(ctx *parser.ConditionNoneContext) {
	l.pushQuantifier(plan.CondNone, ctx.Variable().GetText(), ctx.Expression())
}

func (l *QueryPlannerListener) ExitConditionSingle(ctx *parser.ConditionSingleContext) {
	l.pushQuantifier(plan.CondSingle, ctx.Variable().GetText(), ctx.Expression())
}

// parseExprFunc は expression が関数呼び出し形（IDENTIFIER LPAREN variable RPAREN、例 length(p)）
// なら (関数名, 引数変数, true) を返す。
func parseExprFunc(exprCtx parser.IExpressionContext) (fn, arg string, ok bool) {
	ec, cast := exprCtx.(*parser.ExpressionContext)
	if !cast || ec.LPAREN() == nil || ec.Variable() == nil {
		return "", "", false
	}
	return ec.IDENTIFIER(0).GetText(), ec.Variable().GetText(), true
}

// ================================
// parser用helper関数
// ================================
func parseRange(ctx *parser.RangeContext, rel *RelInfo) *RelInfo {
	if ctx.RangeLiteral() != nil {
		rLit := ctx.RangeLiteral().(*parser.RangeLiteralContext)
		rel = extractRangeLiteral(rLit, rel)
	}

	return rel
}

func extractRangeLiteral(rLit *parser.RangeLiteralContext, rng *RelInfo) *RelInfo {
	if rLit.DOUBLE_DOT() != nil {
		nums := rLit.AllNUMBER()
		switch len(nums) {
		case 2:
			minS := nums[0].GetText()
			maxS := nums[1].GetText()
			minV, _ := strconv.Atoi(minS)
			maxV, _ := strconv.Atoi(maxS)
			rng.minHops = minV
			rng.maxHops = maxV
		case 1:
			only := nums[0].GetText()
			if rLit.GetStart().GetText() != ".." {
				minV, _ := strconv.Atoi(only)
				rng.minHops = minV
			} else {
				maxV, _ := strconv.Atoi(only)
				rng.maxHops = maxV
			}
		default:
		}
	} else {
		numStr := rLit.NUMBER(0).GetText()
		val, _ := strconv.Atoi(numStr)
		rng.minHops = val
		rng.maxHops = val
	}

	return rng
}

func (l *QueryPlannerListener) parseExpression(exprCtx parser.IExpressionContext) (ent, prop string) {
	idents := exprCtx.(*parser.ExpressionContext).AllIDENTIFIER()
	if len(idents) > 0 {
		ent = idents[0].GetText()
		if len(idents) > 1 {
			prop = idents[1].GetText()
		}
	}
	return
}

func (l *QueryPlannerListener) parseValue(valCtx parser.IValueContext) string {
	if valCtx == nil {
		return ""
	}
	txt := valCtx.GetText()
	return strings.Trim(txt, "'\"")
}

// exitValueExprItem は returnItem の valueExpr を Expr（計算式）または従来の材料化列へ振り分ける。
// 単純なプロパティ参照（a / a.b）は golden 互換のため従来の ReturnItem{Alias,Props} 形で扱い、
// 複合式（算術/関数/list）は ReturnItem{Expr} として Return が core.EvalExpr で評価する。
func (l *QueryPlannerListener) exitValueExprItem(ctx *parser.ReturnItemContext, ve parser.IValueExprContext) {
	e := l.popExpr() // この項目の valueExpr サブツリーが積んだ式

	asName := ""
	if ctx.AS() != nil && ctx.Variable() != nil {
		asName = ctx.Variable().GetText()
	}

	// 単純プロパティ参照（DOT のみ・関数/演算子/list なし）は従来の材料化パスへ。
	if pr, ok := ve.(*parser.VePropRefContext); ok {
		idents := pr.AllIDENTIFIER()
		alias := ""
		if len(idents) > 0 {
			alias = idents[0].GetText()
		}
		name := ve.GetText()
		if asName != "" {
			name = asName
		}
		prop := ""
		if len(idents) >= 2 {
			prop = idents[1].GetText()
		}
		l.addProjItem(plan.ReturnItem{Name: name, Alias: alias, Props: []string{prop}, IsCoalesce: false})
		return
	}

	name := ve.GetText()
	if asName != "" {
		name = asName
	}
	l.addProjItem(plan.ReturnItem{Name: name, Expr: e})
}

// ---- valueExpr（RETURN 計算式）の bottom-up 構築 ----

func (l *QueryPlannerListener) pushExpr(e plan.Expr) { l.stackExpr = append(l.stackExpr, e) }

func (l *QueryPlannerListener) popExpr() plan.Expr {
	if len(l.stackExpr) == 0 {
		return &plan.Lit{Val: nil}
	}
	e := l.stackExpr[len(l.stackExpr)-1]
	l.stackExpr = l.stackExpr[:len(l.stackExpr)-1]
	return e
}

// supportedExprFunc は core.EvalExpr が評価できる関数の許可リスト。
// 未対応関数は exprBuildFailed を立てて verbatim へフォールバックさせる（誤った null を返さない）。
func supportedExprFunc(fn string) bool {
	switch strings.ToLower(fn) {
	case "coalesce", "size", "length", "head", "last", "tointeger", "tofloat", "abs", "floor":
		return true
	}
	return false
}

func binOpOfText(s string) plan.BinOp {
	switch s {
	case "+":
		return plan.OpAdd
	case "-":
		return plan.OpSub
	case "*":
		return plan.OpMul
	case "/":
		return plan.OpDiv
	case "%":
		return plan.OpMod
	}
	return plan.OpAdd
}

func (l *QueryPlannerListener) ExitVeNeg(ctx *parser.VeNegContext) {
	l.pushExpr(&plan.Un{Op: plan.OpNeg, X: l.popExpr()})
}

func (l *QueryPlannerListener) ExitVeMulDiv(ctx *parser.VeMulDivContext) {
	r := l.popExpr()
	le := l.popExpr()
	l.pushExpr(&plan.Bin{Op: binOpOfText(ctx.GetOp().GetText()), L: le, R: r})
}

func (l *QueryPlannerListener) ExitVeAddSub(ctx *parser.VeAddSubContext) {
	r := l.popExpr()
	le := l.popExpr()
	l.pushExpr(&plan.Bin{Op: binOpOfText(ctx.GetOp().GetText()), L: le, R: r})
}

func (l *QueryPlannerListener) ExitVeCall(ctx *parser.VeCallContext) {
	n := len(ctx.AllValueExpr())
	args := make([]plan.Expr, n)
	for i := n - 1; i >= 0; i-- {
		args[i] = l.popExpr()
	}
	fn := ctx.IDENTIFIER().GetText()
	if !supportedExprFunc(fn) {
		l.exprBuildFailed = true
	}
	l.pushExpr(&plan.Call{Fn: fn, Args: args})
}

func (l *QueryPlannerListener) ExitVePropRef(ctx *parser.VePropRefContext) {
	idents := ctx.AllIDENTIFIER()
	parts := make([]string, len(idents))
	for i, id := range idents {
		parts[i] = id.GetText()
	}
	l.pushExpr(&plan.Column{Key: strings.Join(parts, ".")})
}

func (l *QueryPlannerListener) ExitVeList(ctx *parser.VeListContext) {
	n := len(ctx.AllValueExpr())
	elems := make([]plan.Expr, n)
	for i := n - 1; i >= 0; i-- {
		elems[i] = l.popExpr()
	}
	l.pushExpr(&plan.List{Elems: elems})
}

func (l *QueryPlannerListener) ExitVeLit(ctx *parser.VeLitContext) {
	if ctx.STRING() != nil {
		l.pushExpr(&plan.Lit{Val: strings.Trim(ctx.STRING().GetText(), "'\"")})
		return
	}
	l.pushExpr(litOf(ctx.NUMBER().GetText()))
}

// ExitVeParen は括弧（識別子的ラッパ）。子の式はすでにスタック上にあるので何もしない。
func (l *QueryPlannerListener) ExitVeParen(ctx *parser.VeParenContext) {}

func (l *QueryPlannerListener) pushCond(c *plan.ConditionNode) {
	l.stackCond = append(l.stackCond, c)
}

func (l *QueryPlannerListener) popCond() *plan.ConditionNode {
	if len(l.stackCond) == 0 {
		return nil
	}
	c := l.stackCond[len(l.stackCond)-1]
	l.stackCond = l.stackCond[:len(l.stackCond)-1]
	return c
}
