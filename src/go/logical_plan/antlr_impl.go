// antlerのListenr用メソッドの実装および補助関数の定義
//
// パーサーの定義ファイル変更時に編集する
package logical_plan

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
	if ctx.COALESCE() != nil {
		var props []string
		var alias string

		// 全ての引数からプロパティ名を抽出
		for _, exprCtx := range ctx.AllExpression() {
			if exprCtx != nil {
				txt := exprCtx.GetText()
				parts := strings.Split(txt, ".")

				if len(parts) == 2 {
					if alias == "" {
						alias = parts[0]
					}
					props = append(props, parts[1])
				}
			}
		}
		l.returnItems = append(l.returnItems, plan.ReturnItem{Name: ctx.GetText(), Alias: alias, Props: props, IsCoalesce: true})

	} else {
		// propertyは一つだけ
		exprCtx := ctx.Expression(0)
		if exprCtx != nil {
			txt := exprCtx.GetText()
			parts := strings.Split(txt, ".")
			if len(parts) == 2 {
				l.returnItems = append(l.returnItems, plan.ReturnItem{Name: txt, Alias: parts[0], Props: []string{parts[1]}, IsCoalesce: false})
			} else {
				// 今回はない
				l.returnItems = append(l.returnItems, plan.ReturnItem{Name: txt, Alias: parts[0], Props: []string{""}, IsCoalesce: false})
			}
		}
	}

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

	if exprCtx != nil {
		txt := exprCtx.GetText()
		parts := strings.Split(txt, ".")
		if len(parts) == 2 {
			l.orderItems = append(l.orderItems, plan.OrderItem{Alias: parts[0], Prop: parts[1], Direction: dir})
		} else {
			fmt.Println("parser error: cannot sort complex data structure")
		}
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
