// Code generated from Cypher.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Cypher
import "github.com/antlr4-go/antlr/v4"

// BaseCypherListener is a complete listener for a parse tree produced by CypherParser.
type BaseCypherListener struct{}

var _ CypherListener = &BaseCypherListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseCypherListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseCypherListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseCypherListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseCypherListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterCypher is called when production cypher is entered.
func (s *BaseCypherListener) EnterCypher(ctx *CypherContext) {}

// ExitCypher is called when production cypher is exited.
func (s *BaseCypherListener) ExitCypher(ctx *CypherContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseCypherListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseCypherListener) ExitStatement(ctx *StatementContext) {}

// EnterReadingClause is called when production readingClause is entered.
func (s *BaseCypherListener) EnterReadingClause(ctx *ReadingClauseContext) {}

// ExitReadingClause is called when production readingClause is exited.
func (s *BaseCypherListener) ExitReadingClause(ctx *ReadingClauseContext) {}

// EnterWithClause is called when production withClause is entered.
func (s *BaseCypherListener) EnterWithClause(ctx *WithClauseContext) {}

// ExitWithClause is called when production withClause is exited.
func (s *BaseCypherListener) ExitWithClause(ctx *WithClauseContext) {}

// EnterMatchClause is called when production matchClause is entered.
func (s *BaseCypherListener) EnterMatchClause(ctx *MatchClauseContext) {}

// ExitMatchClause is called when production matchClause is exited.
func (s *BaseCypherListener) ExitMatchClause(ctx *MatchClauseContext) {}

// EnterReturnClause is called when production returnClause is entered.
func (s *BaseCypherListener) EnterReturnClause(ctx *ReturnClauseContext) {}

// ExitReturnClause is called when production returnClause is exited.
func (s *BaseCypherListener) ExitReturnClause(ctx *ReturnClauseContext) {}

// EnterWhereClause is called when production whereClause is entered.
func (s *BaseCypherListener) EnterWhereClause(ctx *WhereClauseContext) {}

// ExitWhereClause is called when production whereClause is exited.
func (s *BaseCypherListener) ExitWhereClause(ctx *WhereClauseContext) {}

// EnterCallClause is called when production callClause is entered.
func (s *BaseCypherListener) EnterCallClause(ctx *CallClauseContext) {}

// ExitCallClause is called when production callClause is exited.
func (s *BaseCypherListener) ExitCallClause(ctx *CallClauseContext) {}

// EnterAsClause is called when production asClause is entered.
func (s *BaseCypherListener) EnterAsClause(ctx *AsClauseContext) {}

// ExitAsClause is called when production asClause is exited.
func (s *BaseCypherListener) ExitAsClause(ctx *AsClauseContext) {}

// EnterPattern is called when production pattern is entered.
func (s *BaseCypherListener) EnterPattern(ctx *PatternContext) {}

// ExitPattern is called when production pattern is exited.
func (s *BaseCypherListener) ExitPattern(ctx *PatternContext) {}

// EnterNode is called when production node is entered.
func (s *BaseCypherListener) EnterNode(ctx *NodeContext) {}

// ExitNode is called when production node is exited.
func (s *BaseCypherListener) ExitNode(ctx *NodeContext) {}

// EnterRelationship is called when production relationship is entered.
func (s *BaseCypherListener) EnterRelationship(ctx *RelationshipContext) {}

// ExitRelationship is called when production relationship is exited.
func (s *BaseCypherListener) ExitRelationship(ctx *RelationshipContext) {}

// EnterReturnItems is called when production returnItems is entered.
func (s *BaseCypherListener) EnterReturnItems(ctx *ReturnItemsContext) {}

// ExitReturnItems is called when production returnItems is exited.
func (s *BaseCypherListener) ExitReturnItems(ctx *ReturnItemsContext) {}

// EnterReturnItem is called when production returnItem is entered.
func (s *BaseCypherListener) EnterReturnItem(ctx *ReturnItemContext) {}

// ExitReturnItem is called when production returnItem is exited.
func (s *BaseCypherListener) ExitReturnItem(ctx *ReturnItemContext) {}

// EnterCaseSearched is called when production caseSearched is entered.
func (s *BaseCypherListener) EnterCaseSearched(ctx *CaseSearchedContext) {}

// ExitCaseSearched is called when production caseSearched is exited.
func (s *BaseCypherListener) ExitCaseSearched(ctx *CaseSearchedContext) {}

// EnterCaseSimple is called when production caseSimple is entered.
func (s *BaseCypherListener) EnterCaseSimple(ctx *CaseSimpleContext) {}

// ExitCaseSimple is called when production caseSimple is exited.
func (s *BaseCypherListener) ExitCaseSimple(ctx *CaseSimpleContext) {}

// EnterCaseKey is called when production caseKey is entered.
func (s *BaseCypherListener) EnterCaseKey(ctx *CaseKeyContext) {}

// ExitCaseKey is called when production caseKey is exited.
func (s *BaseCypherListener) ExitCaseKey(ctx *CaseKeyContext) {}

// EnterCaseResult is called when production caseResult is entered.
func (s *BaseCypherListener) EnterCaseResult(ctx *CaseResultContext) {}

// ExitCaseResult is called when production caseResult is exited.
func (s *BaseCypherListener) ExitCaseResult(ctx *CaseResultContext) {}

// EnterVeList is called when production veList is entered.
func (s *BaseCypherListener) EnterVeList(ctx *VeListContext) {}

// ExitVeList is called when production veList is exited.
func (s *BaseCypherListener) ExitVeList(ctx *VeListContext) {}

// EnterVeMulDiv is called when production veMulDiv is entered.
func (s *BaseCypherListener) EnterVeMulDiv(ctx *VeMulDivContext) {}

// ExitVeMulDiv is called when production veMulDiv is exited.
func (s *BaseCypherListener) ExitVeMulDiv(ctx *VeMulDivContext) {}

// EnterVeAddSub is called when production veAddSub is entered.
func (s *BaseCypherListener) EnterVeAddSub(ctx *VeAddSubContext) {}

// ExitVeAddSub is called when production veAddSub is exited.
func (s *BaseCypherListener) ExitVeAddSub(ctx *VeAddSubContext) {}

// EnterVeLit is called when production veLit is entered.
func (s *BaseCypherListener) EnterVeLit(ctx *VeLitContext) {}

// ExitVeLit is called when production veLit is exited.
func (s *BaseCypherListener) ExitVeLit(ctx *VeLitContext) {}

// EnterVeCall is called when production veCall is entered.
func (s *BaseCypherListener) EnterVeCall(ctx *VeCallContext) {}

// ExitVeCall is called when production veCall is exited.
func (s *BaseCypherListener) ExitVeCall(ctx *VeCallContext) {}

// EnterVeParen is called when production veParen is entered.
func (s *BaseCypherListener) EnterVeParen(ctx *VeParenContext) {}

// ExitVeParen is called when production veParen is exited.
func (s *BaseCypherListener) ExitVeParen(ctx *VeParenContext) {}

// EnterVeNeg is called when production veNeg is entered.
func (s *BaseCypherListener) EnterVeNeg(ctx *VeNegContext) {}

// ExitVeNeg is called when production veNeg is exited.
func (s *BaseCypherListener) ExitVeNeg(ctx *VeNegContext) {}

// EnterVePropRef is called when production vePropRef is entered.
func (s *BaseCypherListener) EnterVePropRef(ctx *VePropRefContext) {}

// ExitVePropRef is called when production vePropRef is exited.
func (s *BaseCypherListener) ExitVePropRef(ctx *VePropRefContext) {}

// EnterAggregateFunc is called when production aggregateFunc is entered.
func (s *BaseCypherListener) EnterAggregateFunc(ctx *AggregateFuncContext) {}

// ExitAggregateFunc is called when production aggregateFunc is exited.
func (s *BaseCypherListener) ExitAggregateFunc(ctx *AggregateFuncContext) {}

// EnterAggArg is called when production aggArg is entered.
func (s *BaseCypherListener) EnterAggArg(ctx *AggArgContext) {}

// ExitAggArg is called when production aggArg is exited.
func (s *BaseCypherListener) ExitAggArg(ctx *AggArgContext) {}

// EnterOrderItems is called when production orderItems is entered.
func (s *BaseCypherListener) EnterOrderItems(ctx *OrderItemsContext) {}

// ExitOrderItems is called when production orderItems is exited.
func (s *BaseCypherListener) ExitOrderItems(ctx *OrderItemsContext) {}

// EnterOrderItem is called when production orderItem is entered.
func (s *BaseCypherListener) EnterOrderItem(ctx *OrderItemContext) {}

// ExitOrderItem is called when production orderItem is exited.
func (s *BaseCypherListener) ExitOrderItem(ctx *OrderItemContext) {}

// EnterLimitNum is called when production limitNum is entered.
func (s *BaseCypherListener) EnterLimitNum(ctx *LimitNumContext) {}

// ExitLimitNum is called when production limitNum is exited.
func (s *BaseCypherListener) ExitLimitNum(ctx *LimitNumContext) {}

// EnterLabels is called when production labels is entered.
func (s *BaseCypherListener) EnterLabels(ctx *LabelsContext) {}

// ExitLabels is called when production labels is exited.
func (s *BaseCypherListener) ExitLabels(ctx *LabelsContext) {}

// EnterLabel is called when production label is entered.
func (s *BaseCypherListener) EnterLabel(ctx *LabelContext) {}

// ExitLabel is called when production label is exited.
func (s *BaseCypherListener) ExitLabel(ctx *LabelContext) {}

// EnterProperties is called when production properties is entered.
func (s *BaseCypherListener) EnterProperties(ctx *PropertiesContext) {}

// ExitProperties is called when production properties is exited.
func (s *BaseCypherListener) ExitProperties(ctx *PropertiesContext) {}

// EnterProperty is called when production property is entered.
func (s *BaseCypherListener) EnterProperty(ctx *PropertyContext) {}

// ExitProperty is called when production property is exited.
func (s *BaseCypherListener) ExitProperty(ctx *PropertyContext) {}

// EnterConditionAnd is called when production ConditionAnd is entered.
func (s *BaseCypherListener) EnterConditionAnd(ctx *ConditionAndContext) {}

// ExitConditionAnd is called when production ConditionAnd is exited.
func (s *BaseCypherListener) ExitConditionAnd(ctx *ConditionAndContext) {}

// EnterConditionNot is called when production ConditionNot is entered.
func (s *BaseCypherListener) EnterConditionNot(ctx *ConditionNotContext) {}

// ExitConditionNot is called when production ConditionNot is exited.
func (s *BaseCypherListener) ExitConditionNot(ctx *ConditionNotContext) {}

// EnterConditionVarNeq is called when production ConditionVarNeq is entered.
func (s *BaseCypherListener) EnterConditionVarNeq(ctx *ConditionVarNeqContext) {}

// ExitConditionVarNeq is called when production ConditionVarNeq is exited.
func (s *BaseCypherListener) ExitConditionVarNeq(ctx *ConditionVarNeqContext) {}

// EnterConditionVarGreaterEqual is called when production ConditionVarGreaterEqual is entered.
func (s *BaseCypherListener) EnterConditionVarGreaterEqual(ctx *ConditionVarGreaterEqualContext) {}

// ExitConditionVarGreaterEqual is called when production ConditionVarGreaterEqual is exited.
func (s *BaseCypherListener) ExitConditionVarGreaterEqual(ctx *ConditionVarGreaterEqualContext) {}

// EnterConditionIsNull is called when production ConditionIsNull is entered.
func (s *BaseCypherListener) EnterConditionIsNull(ctx *ConditionIsNullContext) {}

// ExitConditionIsNull is called when production ConditionIsNull is exited.
func (s *BaseCypherListener) ExitConditionIsNull(ctx *ConditionIsNullContext) {}

// EnterConditionGreater is called when production ConditionGreater is entered.
func (s *BaseCypherListener) EnterConditionGreater(ctx *ConditionGreaterContext) {}

// ExitConditionGreater is called when production ConditionGreater is exited.
func (s *BaseCypherListener) ExitConditionGreater(ctx *ConditionGreaterContext) {}

// EnterConditionIn is called when production ConditionIn is entered.
func (s *BaseCypherListener) EnterConditionIn(ctx *ConditionInContext) {}

// ExitConditionIn is called when production ConditionIn is exited.
func (s *BaseCypherListener) ExitConditionIn(ctx *ConditionInContext) {}

// EnterConditionAny is called when production ConditionAny is entered.
func (s *BaseCypherListener) EnterConditionAny(ctx *ConditionAnyContext) {}

// ExitConditionAny is called when production ConditionAny is exited.
func (s *BaseCypherListener) ExitConditionAny(ctx *ConditionAnyContext) {}

// EnterConditionVarEq is called when production ConditionVarEq is entered.
func (s *BaseCypherListener) EnterConditionVarEq(ctx *ConditionVarEqContext) {}

// ExitConditionVarEq is called when production ConditionVarEq is exited.
func (s *BaseCypherListener) ExitConditionVarEq(ctx *ConditionVarEqContext) {}

// EnterConditionVarLessEqual is called when production ConditionVarLessEqual is entered.
func (s *BaseCypherListener) EnterConditionVarLessEqual(ctx *ConditionVarLessEqualContext) {}

// ExitConditionVarLessEqual is called when production ConditionVarLessEqual is exited.
func (s *BaseCypherListener) ExitConditionVarLessEqual(ctx *ConditionVarLessEqualContext) {}

// EnterConditionOr is called when production ConditionOr is entered.
func (s *BaseCypherListener) EnterConditionOr(ctx *ConditionOrContext) {}

// ExitConditionOr is called when production ConditionOr is exited.
func (s *BaseCypherListener) ExitConditionOr(ctx *ConditionOrContext) {}

// EnterConditionIsNotNull is called when production ConditionIsNotNull is entered.
func (s *BaseCypherListener) EnterConditionIsNotNull(ctx *ConditionIsNotNullContext) {}

// ExitConditionIsNotNull is called when production ConditionIsNotNull is exited.
func (s *BaseCypherListener) ExitConditionIsNotNull(ctx *ConditionIsNotNullContext) {}

// EnterConditionParen is called when production conditionParen is entered.
func (s *BaseCypherListener) EnterConditionParen(ctx *ConditionParenContext) {}

// ExitConditionParen is called when production conditionParen is exited.
func (s *BaseCypherListener) ExitConditionParen(ctx *ConditionParenContext) {}

// EnterConditionLessEqual is called when production ConditionLessEqual is entered.
func (s *BaseCypherListener) EnterConditionLessEqual(ctx *ConditionLessEqualContext) {}

// ExitConditionLessEqual is called when production ConditionLessEqual is exited.
func (s *BaseCypherListener) ExitConditionLessEqual(ctx *ConditionLessEqualContext) {}

// EnterConditionNone is called when production ConditionNone is entered.
func (s *BaseCypherListener) EnterConditionNone(ctx *ConditionNoneContext) {}

// ExitConditionNone is called when production ConditionNone is exited.
func (s *BaseCypherListener) ExitConditionNone(ctx *ConditionNoneContext) {}

// EnterConditionGreaterEqual is called when production ConditionGreaterEqual is entered.
func (s *BaseCypherListener) EnterConditionGreaterEqual(ctx *ConditionGreaterEqualContext) {}

// ExitConditionGreaterEqual is called when production ConditionGreaterEqual is exited.
func (s *BaseCypherListener) ExitConditionGreaterEqual(ctx *ConditionGreaterEqualContext) {}

// EnterConditionVarGreater is called when production ConditionVarGreater is entered.
func (s *BaseCypherListener) EnterConditionVarGreater(ctx *ConditionVarGreaterContext) {}

// ExitConditionVarGreater is called when production ConditionVarGreater is exited.
func (s *BaseCypherListener) ExitConditionVarGreater(ctx *ConditionVarGreaterContext) {}

// EnterConditionAll is called when production ConditionAll is entered.
func (s *BaseCypherListener) EnterConditionAll(ctx *ConditionAllContext) {}

// ExitConditionAll is called when production ConditionAll is exited.
func (s *BaseCypherListener) ExitConditionAll(ctx *ConditionAllContext) {}

// EnterConditionVarLess is called when production ConditionVarLess is entered.
func (s *BaseCypherListener) EnterConditionVarLess(ctx *ConditionVarLessContext) {}

// ExitConditionVarLess is called when production ConditionVarLess is exited.
func (s *BaseCypherListener) ExitConditionVarLess(ctx *ConditionVarLessContext) {}

// EnterConditionNotEquality is called when production ConditionNotEquality is entered.
func (s *BaseCypherListener) EnterConditionNotEquality(ctx *ConditionNotEqualityContext) {}

// ExitConditionNotEquality is called when production ConditionNotEquality is exited.
func (s *BaseCypherListener) ExitConditionNotEquality(ctx *ConditionNotEqualityContext) {}

// EnterConditionLess is called when production ConditionLess is entered.
func (s *BaseCypherListener) EnterConditionLess(ctx *ConditionLessContext) {}

// ExitConditionLess is called when production ConditionLess is exited.
func (s *BaseCypherListener) ExitConditionLess(ctx *ConditionLessContext) {}

// EnterConditionSingle is called when production ConditionSingle is entered.
func (s *BaseCypherListener) EnterConditionSingle(ctx *ConditionSingleContext) {}

// ExitConditionSingle is called when production ConditionSingle is exited.
func (s *BaseCypherListener) ExitConditionSingle(ctx *ConditionSingleContext) {}

// EnterConditionEquality is called when production ConditionEquality is entered.
func (s *BaseCypherListener) EnterConditionEquality(ctx *ConditionEqualityContext) {}

// ExitConditionEquality is called when production ConditionEquality is exited.
func (s *BaseCypherListener) ExitConditionEquality(ctx *ConditionEqualityContext) {}

// EnterList is called when production list is entered.
func (s *BaseCypherListener) EnterList(ctx *ListContext) {}

// ExitList is called when production list is exited.
func (s *BaseCypherListener) ExitList(ctx *ListContext) {}

// EnterVariable is called when production variable is entered.
func (s *BaseCypherListener) EnterVariable(ctx *VariableContext) {}

// ExitVariable is called when production variable is exited.
func (s *BaseCypherListener) ExitVariable(ctx *VariableContext) {}

// EnterTypes is called when production types is entered.
func (s *BaseCypherListener) EnterTypes(ctx *TypesContext) {}

// ExitTypes is called when production types is exited.
func (s *BaseCypherListener) ExitTypes(ctx *TypesContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseCypherListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseCypherListener) ExitExpression(ctx *ExpressionContext) {}

// EnterValue is called when production value is entered.
func (s *BaseCypherListener) EnterValue(ctx *ValueContext) {}

// ExitValue is called when production value is exited.
func (s *BaseCypherListener) ExitValue(ctx *ValueContext) {}

// EnterRange is called when production range is entered.
func (s *BaseCypherListener) EnterRange(ctx *RangeContext) {}

// ExitRange is called when production range is exited.
func (s *BaseCypherListener) ExitRange(ctx *RangeContext) {}

// EnterRangeLiteral is called when production rangeLiteral is entered.
func (s *BaseCypherListener) EnterRangeLiteral(ctx *RangeLiteralContext) {}

// ExitRangeLiteral is called when production rangeLiteral is exited.
func (s *BaseCypherListener) ExitRangeLiteral(ctx *RangeLiteralContext) {}
