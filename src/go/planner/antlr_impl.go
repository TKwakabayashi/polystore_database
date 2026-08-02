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

	// 1) 集約関数 count/sum/avg/min/max
	if agg := ctx.AggregateFunc(); agg != nil {
		item := l.parseAggregate(agg.(*parser.AggregateFuncContext))
		if asName != "" {
			item.OutName = asName
		}
		l.returnItems = append(l.returnItems, plan.ReturnItem{
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
		l.returnItems = append(l.returnItems, plan.ReturnItem{Name: name, Alias: alias, Props: props, IsCoalesce: true})
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
			l.returnItems = append(l.returnItems, plan.ReturnItem{Name: name, Alias: parts[0], Props: []string{parts[1]}, IsCoalesce: false})
		} else {
			l.returnItems = append(l.returnItems, plan.ReturnItem{Name: name, Alias: parts[0], Props: []string{""}, IsCoalesce: false})
		}
	}
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
	if len(parts) == 2 {
		l.orderItems = append(l.orderItems, plan.OrderItem{Alias: parts[0], Prop: parts[1], Direction: dir, Key: parts[0] + "." + parts[1]})
	} else {
		// 集約別名など単一トークンでの並べ替え
		l.orderItems = append(l.orderItems, plan.OrderItem{Key: txt, Direction: dir})
	}
}

// EnterLimitNum is called when production limitNum is entered.
func (l *QueryPlannerListener) EnterLimitNum(ctx *parser.LimitNumContext) {
	l.limitNum, _ = strconv.Atoi(ctx.NUMBER().GetText())
}

func (l *QueryPlannerListener) ExitWhereClause(ctx *parser.WhereClauseContext) {
	if len(l.stackCond) == 0 {
		return
	}

	rootCond := l.popCond()

	condList := plan.DecomposeOuterAndOp(rootCond)

	for _, cond := range condList {
		if cond.Alias != "" {
			l.symbolCondMapping[cond.Alias] = append(l.symbolCondMapping[cond.Alias], cond)
		}
	}
}

func (l *QueryPlannerListener) ExitConditionEquality(ctx *parser.ConditionEqualityContext) {
	expr := ctx.Expression()
	val := ctx.Value()
	alias, prop := l.parseExpression(expr)
	c := &plan.ConditionNode{
		Type:     plan.CondEq,
		Alias:    alias,
		Labels:   l.entityInfo[l.symbolEntTable[alias]].labels,
		Property: prop,
		Value:    l.parseValue(val),
	}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionGreater(ctx *parser.ConditionGreaterContext) {
	expr := ctx.Expression()
	val := ctx.Value()
	alias, prop := l.parseExpression(expr)
	c := &plan.ConditionNode{
		Type:     plan.CondGreater,
		Alias:    alias,
		Labels:   l.entityInfo[l.symbolEntTable[alias]].labels,
		Property: prop,
		Value:    l.parseValue(val),
	}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionLess(ctx *parser.ConditionLessContext) {
	expr := ctx.Expression()
	val := ctx.Value()
	alias, prop := l.parseExpression(expr)
	c := &plan.ConditionNode{
		Type:     plan.CondLess,
		Alias:    alias,
		Labels:   l.entityInfo[l.symbolEntTable[alias]].labels,
		Property: prop,
		Value:    l.parseValue(val),
	}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionGreaterEqual(ctx *parser.ConditionGreaterEqualContext) {
	expr := ctx.Expression()
	val := ctx.Value()
	alias, prop := l.parseExpression(expr)
	c := &plan.ConditionNode{
		Type:     plan.CondGreaterEq,
		Alias:    alias,
		Labels:   l.entityInfo[l.symbolEntTable[alias]].labels,
		Property: prop,
		Value:    l.parseValue(val),
	}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionLessEqual(ctx *parser.ConditionLessEqualContext) {
	expr := ctx.Expression()
	val := ctx.Value()
	alias, prop := l.parseExpression(expr)
	c := &plan.ConditionNode{
		Type:     plan.CondLessEq,
		Alias:    alias,
		Labels:   l.entityInfo[l.symbolEntTable[alias]].labels,
		Property: prop,
		Value:    l.parseValue(val),
	}
	l.pushCond(c)
}

func (l *QueryPlannerListener) ExitConditionNotEquality(ctx *parser.ConditionNotEqualityContext) {
	expr := ctx.Expression()
	val := ctx.Value()
	alias, prop := l.parseExpression(expr)
	c := &plan.ConditionNode{
		Type:     plan.CondNeq,
		Alias:    alias,
		Labels:   l.entityInfo[l.symbolEntTable[alias]].labels,
		Property: prop,
		Value:    l.parseValue(val),
	}
	l.pushCond(c)
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

func (l *QueryPlannerListener) ExitConditionAll(ctx *parser.ConditionAllContext) {
	// 1. 内部条件を回収
	innerCond := l.popCond()

	// 2. ループ変数 (n)
	iterVar := ctx.Variable().GetText()

	// 3. Expressionの解析
	exprCtx := ctx.Expression()
	var listAlias string
	var listFunc string

	if exprCtx.LPAREN() != nil {
		listFunc = exprCtx.IDENTIFIER(0).GetText()
		listAlias = exprCtx.Variable().GetText()
	} else {
		listAlias = exprCtx.GetText()
		listFunc = ""
	}

	// 4. ConditionNodeの構築
	c := &plan.ConditionNode{
		Type:     plan.CondAll,
		Alias:    listAlias, // "p" or "answers"
		Property: listFunc,  // "nodes" or ""
		Value:    iterVar,   // "n" (ループ内でのバインド名)
		Child:    innerCond,
	}

	l.pushCond(c)
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
