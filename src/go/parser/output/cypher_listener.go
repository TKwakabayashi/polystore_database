// Code generated from Cypher.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Cypher
import "github.com/antlr4-go/antlr/v4"

// CypherListener is a complete listener for a parse tree produced by CypherParser.
type CypherListener interface {
	antlr.ParseTreeListener

	// EnterCypher is called when entering the cypher production.
	EnterCypher(c *CypherContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// EnterReadingClause is called when entering the readingClause production.
	EnterReadingClause(c *ReadingClauseContext)

	// EnterWithClause is called when entering the withClause production.
	EnterWithClause(c *WithClauseContext)

	// EnterMatchClause is called when entering the matchClause production.
	EnterMatchClause(c *MatchClauseContext)

	// EnterReturnClause is called when entering the returnClause production.
	EnterReturnClause(c *ReturnClauseContext)

	// EnterWhereClause is called when entering the whereClause production.
	EnterWhereClause(c *WhereClauseContext)

	// EnterCallClause is called when entering the callClause production.
	EnterCallClause(c *CallClauseContext)

	// EnterAsClause is called when entering the asClause production.
	EnterAsClause(c *AsClauseContext)

	// EnterPattern is called when entering the pattern production.
	EnterPattern(c *PatternContext)

	// EnterNode is called when entering the node production.
	EnterNode(c *NodeContext)

	// EnterRelationship is called when entering the relationship production.
	EnterRelationship(c *RelationshipContext)

	// EnterReturnItems is called when entering the returnItems production.
	EnterReturnItems(c *ReturnItemsContext)

	// EnterReturnItem is called when entering the returnItem production.
	EnterReturnItem(c *ReturnItemContext)

	// EnterCaseSearched is called when entering the caseSearched production.
	EnterCaseSearched(c *CaseSearchedContext)

	// EnterCaseSimple is called when entering the caseSimple production.
	EnterCaseSimple(c *CaseSimpleContext)

	// EnterCaseKey is called when entering the caseKey production.
	EnterCaseKey(c *CaseKeyContext)

	// EnterCaseResult is called when entering the caseResult production.
	EnterCaseResult(c *CaseResultContext)

	// EnterVeList is called when entering the veList production.
	EnterVeList(c *VeListContext)

	// EnterVeMulDiv is called when entering the veMulDiv production.
	EnterVeMulDiv(c *VeMulDivContext)

	// EnterVeAddSub is called when entering the veAddSub production.
	EnterVeAddSub(c *VeAddSubContext)

	// EnterVeLit is called when entering the veLit production.
	EnterVeLit(c *VeLitContext)

	// EnterVeCall is called when entering the veCall production.
	EnterVeCall(c *VeCallContext)

	// EnterVeParen is called when entering the veParen production.
	EnterVeParen(c *VeParenContext)

	// EnterVeNeg is called when entering the veNeg production.
	EnterVeNeg(c *VeNegContext)

	// EnterVePropRef is called when entering the vePropRef production.
	EnterVePropRef(c *VePropRefContext)

	// EnterAggregateFunc is called when entering the aggregateFunc production.
	EnterAggregateFunc(c *AggregateFuncContext)

	// EnterAggArg is called when entering the aggArg production.
	EnterAggArg(c *AggArgContext)

	// EnterOrderItems is called when entering the orderItems production.
	EnterOrderItems(c *OrderItemsContext)

	// EnterOrderItem is called when entering the orderItem production.
	EnterOrderItem(c *OrderItemContext)

	// EnterLimitNum is called when entering the limitNum production.
	EnterLimitNum(c *LimitNumContext)

	// EnterLabels is called when entering the labels production.
	EnterLabels(c *LabelsContext)

	// EnterLabel is called when entering the label production.
	EnterLabel(c *LabelContext)

	// EnterProperties is called when entering the properties production.
	EnterProperties(c *PropertiesContext)

	// EnterProperty is called when entering the property production.
	EnterProperty(c *PropertyContext)

	// EnterConditionAnd is called when entering the ConditionAnd production.
	EnterConditionAnd(c *ConditionAndContext)

	// EnterConditionNot is called when entering the ConditionNot production.
	EnterConditionNot(c *ConditionNotContext)

	// EnterConditionVarNeq is called when entering the ConditionVarNeq production.
	EnterConditionVarNeq(c *ConditionVarNeqContext)

	// EnterConditionVarGreaterEqual is called when entering the ConditionVarGreaterEqual production.
	EnterConditionVarGreaterEqual(c *ConditionVarGreaterEqualContext)

	// EnterConditionIsNull is called when entering the ConditionIsNull production.
	EnterConditionIsNull(c *ConditionIsNullContext)

	// EnterConditionGreater is called when entering the ConditionGreater production.
	EnterConditionGreater(c *ConditionGreaterContext)

	// EnterConditionIn is called when entering the ConditionIn production.
	EnterConditionIn(c *ConditionInContext)

	// EnterConditionAny is called when entering the ConditionAny production.
	EnterConditionAny(c *ConditionAnyContext)

	// EnterConditionVarEq is called when entering the ConditionVarEq production.
	EnterConditionVarEq(c *ConditionVarEqContext)

	// EnterConditionVarLessEqual is called when entering the ConditionVarLessEqual production.
	EnterConditionVarLessEqual(c *ConditionVarLessEqualContext)

	// EnterConditionOr is called when entering the ConditionOr production.
	EnterConditionOr(c *ConditionOrContext)

	// EnterConditionIsNotNull is called when entering the ConditionIsNotNull production.
	EnterConditionIsNotNull(c *ConditionIsNotNullContext)

	// EnterConditionParen is called when entering the conditionParen production.
	EnterConditionParen(c *ConditionParenContext)

	// EnterConditionLessEqual is called when entering the ConditionLessEqual production.
	EnterConditionLessEqual(c *ConditionLessEqualContext)

	// EnterConditionNone is called when entering the ConditionNone production.
	EnterConditionNone(c *ConditionNoneContext)

	// EnterConditionGreaterEqual is called when entering the ConditionGreaterEqual production.
	EnterConditionGreaterEqual(c *ConditionGreaterEqualContext)

	// EnterConditionVarGreater is called when entering the ConditionVarGreater production.
	EnterConditionVarGreater(c *ConditionVarGreaterContext)

	// EnterConditionAll is called when entering the ConditionAll production.
	EnterConditionAll(c *ConditionAllContext)

	// EnterConditionVarLess is called when entering the ConditionVarLess production.
	EnterConditionVarLess(c *ConditionVarLessContext)

	// EnterConditionNotEquality is called when entering the ConditionNotEquality production.
	EnterConditionNotEquality(c *ConditionNotEqualityContext)

	// EnterConditionLess is called when entering the ConditionLess production.
	EnterConditionLess(c *ConditionLessContext)

	// EnterConditionSingle is called when entering the ConditionSingle production.
	EnterConditionSingle(c *ConditionSingleContext)

	// EnterConditionEquality is called when entering the ConditionEquality production.
	EnterConditionEquality(c *ConditionEqualityContext)

	// EnterList is called when entering the list production.
	EnterList(c *ListContext)

	// EnterVariable is called when entering the variable production.
	EnterVariable(c *VariableContext)

	// EnterTypes is called when entering the types production.
	EnterTypes(c *TypesContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterValue is called when entering the value production.
	EnterValue(c *ValueContext)

	// EnterRange is called when entering the range production.
	EnterRange(c *RangeContext)

	// EnterRangeLiteral is called when entering the rangeLiteral production.
	EnterRangeLiteral(c *RangeLiteralContext)

	// ExitCypher is called when exiting the cypher production.
	ExitCypher(c *CypherContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)

	// ExitReadingClause is called when exiting the readingClause production.
	ExitReadingClause(c *ReadingClauseContext)

	// ExitWithClause is called when exiting the withClause production.
	ExitWithClause(c *WithClauseContext)

	// ExitMatchClause is called when exiting the matchClause production.
	ExitMatchClause(c *MatchClauseContext)

	// ExitReturnClause is called when exiting the returnClause production.
	ExitReturnClause(c *ReturnClauseContext)

	// ExitWhereClause is called when exiting the whereClause production.
	ExitWhereClause(c *WhereClauseContext)

	// ExitCallClause is called when exiting the callClause production.
	ExitCallClause(c *CallClauseContext)

	// ExitAsClause is called when exiting the asClause production.
	ExitAsClause(c *AsClauseContext)

	// ExitPattern is called when exiting the pattern production.
	ExitPattern(c *PatternContext)

	// ExitNode is called when exiting the node production.
	ExitNode(c *NodeContext)

	// ExitRelationship is called when exiting the relationship production.
	ExitRelationship(c *RelationshipContext)

	// ExitReturnItems is called when exiting the returnItems production.
	ExitReturnItems(c *ReturnItemsContext)

	// ExitReturnItem is called when exiting the returnItem production.
	ExitReturnItem(c *ReturnItemContext)

	// ExitCaseSearched is called when exiting the caseSearched production.
	ExitCaseSearched(c *CaseSearchedContext)

	// ExitCaseSimple is called when exiting the caseSimple production.
	ExitCaseSimple(c *CaseSimpleContext)

	// ExitCaseKey is called when exiting the caseKey production.
	ExitCaseKey(c *CaseKeyContext)

	// ExitCaseResult is called when exiting the caseResult production.
	ExitCaseResult(c *CaseResultContext)

	// ExitVeList is called when exiting the veList production.
	ExitVeList(c *VeListContext)

	// ExitVeMulDiv is called when exiting the veMulDiv production.
	ExitVeMulDiv(c *VeMulDivContext)

	// ExitVeAddSub is called when exiting the veAddSub production.
	ExitVeAddSub(c *VeAddSubContext)

	// ExitVeLit is called when exiting the veLit production.
	ExitVeLit(c *VeLitContext)

	// ExitVeCall is called when exiting the veCall production.
	ExitVeCall(c *VeCallContext)

	// ExitVeParen is called when exiting the veParen production.
	ExitVeParen(c *VeParenContext)

	// ExitVeNeg is called when exiting the veNeg production.
	ExitVeNeg(c *VeNegContext)

	// ExitVePropRef is called when exiting the vePropRef production.
	ExitVePropRef(c *VePropRefContext)

	// ExitAggregateFunc is called when exiting the aggregateFunc production.
	ExitAggregateFunc(c *AggregateFuncContext)

	// ExitAggArg is called when exiting the aggArg production.
	ExitAggArg(c *AggArgContext)

	// ExitOrderItems is called when exiting the orderItems production.
	ExitOrderItems(c *OrderItemsContext)

	// ExitOrderItem is called when exiting the orderItem production.
	ExitOrderItem(c *OrderItemContext)

	// ExitLimitNum is called when exiting the limitNum production.
	ExitLimitNum(c *LimitNumContext)

	// ExitLabels is called when exiting the labels production.
	ExitLabels(c *LabelsContext)

	// ExitLabel is called when exiting the label production.
	ExitLabel(c *LabelContext)

	// ExitProperties is called when exiting the properties production.
	ExitProperties(c *PropertiesContext)

	// ExitProperty is called when exiting the property production.
	ExitProperty(c *PropertyContext)

	// ExitConditionAnd is called when exiting the ConditionAnd production.
	ExitConditionAnd(c *ConditionAndContext)

	// ExitConditionNot is called when exiting the ConditionNot production.
	ExitConditionNot(c *ConditionNotContext)

	// ExitConditionVarNeq is called when exiting the ConditionVarNeq production.
	ExitConditionVarNeq(c *ConditionVarNeqContext)

	// ExitConditionVarGreaterEqual is called when exiting the ConditionVarGreaterEqual production.
	ExitConditionVarGreaterEqual(c *ConditionVarGreaterEqualContext)

	// ExitConditionIsNull is called when exiting the ConditionIsNull production.
	ExitConditionIsNull(c *ConditionIsNullContext)

	// ExitConditionGreater is called when exiting the ConditionGreater production.
	ExitConditionGreater(c *ConditionGreaterContext)

	// ExitConditionIn is called when exiting the ConditionIn production.
	ExitConditionIn(c *ConditionInContext)

	// ExitConditionAny is called when exiting the ConditionAny production.
	ExitConditionAny(c *ConditionAnyContext)

	// ExitConditionVarEq is called when exiting the ConditionVarEq production.
	ExitConditionVarEq(c *ConditionVarEqContext)

	// ExitConditionVarLessEqual is called when exiting the ConditionVarLessEqual production.
	ExitConditionVarLessEqual(c *ConditionVarLessEqualContext)

	// ExitConditionOr is called when exiting the ConditionOr production.
	ExitConditionOr(c *ConditionOrContext)

	// ExitConditionIsNotNull is called when exiting the ConditionIsNotNull production.
	ExitConditionIsNotNull(c *ConditionIsNotNullContext)

	// ExitConditionParen is called when exiting the conditionParen production.
	ExitConditionParen(c *ConditionParenContext)

	// ExitConditionLessEqual is called when exiting the ConditionLessEqual production.
	ExitConditionLessEqual(c *ConditionLessEqualContext)

	// ExitConditionNone is called when exiting the ConditionNone production.
	ExitConditionNone(c *ConditionNoneContext)

	// ExitConditionGreaterEqual is called when exiting the ConditionGreaterEqual production.
	ExitConditionGreaterEqual(c *ConditionGreaterEqualContext)

	// ExitConditionVarGreater is called when exiting the ConditionVarGreater production.
	ExitConditionVarGreater(c *ConditionVarGreaterContext)

	// ExitConditionAll is called when exiting the ConditionAll production.
	ExitConditionAll(c *ConditionAllContext)

	// ExitConditionVarLess is called when exiting the ConditionVarLess production.
	ExitConditionVarLess(c *ConditionVarLessContext)

	// ExitConditionNotEquality is called when exiting the ConditionNotEquality production.
	ExitConditionNotEquality(c *ConditionNotEqualityContext)

	// ExitConditionLess is called when exiting the ConditionLess production.
	ExitConditionLess(c *ConditionLessContext)

	// ExitConditionSingle is called when exiting the ConditionSingle production.
	ExitConditionSingle(c *ConditionSingleContext)

	// ExitConditionEquality is called when exiting the ConditionEquality production.
	ExitConditionEquality(c *ConditionEqualityContext)

	// ExitList is called when exiting the list production.
	ExitList(c *ListContext)

	// ExitVariable is called when exiting the variable production.
	ExitVariable(c *VariableContext)

	// ExitTypes is called when exiting the types production.
	ExitTypes(c *TypesContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitValue is called when exiting the value production.
	ExitValue(c *ValueContext)

	// ExitRange is called when exiting the range production.
	ExitRange(c *RangeContext)

	// ExitRangeLiteral is called when exiting the rangeLiteral production.
	ExitRangeLiteral(c *RangeLiteralContext)
}
