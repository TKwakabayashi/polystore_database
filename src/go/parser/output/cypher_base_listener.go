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

// EnterConditionOr is called when production ConditionOr is entered.
func (s *BaseCypherListener) EnterConditionOr(ctx *ConditionOrContext) {}

// ExitConditionOr is called when production ConditionOr is exited.
func (s *BaseCypherListener) ExitConditionOr(ctx *ConditionOrContext) {}

// EnterConditionNot is called when production ConditionNot is entered.
func (s *BaseCypherListener) EnterConditionNot(ctx *ConditionNotContext) {}

// ExitConditionNot is called when production ConditionNot is exited.
func (s *BaseCypherListener) ExitConditionNot(ctx *ConditionNotContext) {}

// EnterConditionParen is called when production conditionParen is entered.
func (s *BaseCypherListener) EnterConditionParen(ctx *ConditionParenContext) {}

// ExitConditionParen is called when production conditionParen is exited.
func (s *BaseCypherListener) ExitConditionParen(ctx *ConditionParenContext) {}

// EnterConditionNone is called when production ConditionNone is entered.
func (s *BaseCypherListener) EnterConditionNone(ctx *ConditionNoneContext) {}

// ExitConditionNone is called when production ConditionNone is exited.
func (s *BaseCypherListener) ExitConditionNone(ctx *ConditionNoneContext) {}

// EnterConditionAll is called when production ConditionAll is entered.
func (s *BaseCypherListener) EnterConditionAll(ctx *ConditionAllContext) {}

// ExitConditionAll is called when production ConditionAll is exited.
func (s *BaseCypherListener) ExitConditionAll(ctx *ConditionAllContext) {}

// EnterConditionGreater is called when production ConditionGreater is entered.
func (s *BaseCypherListener) EnterConditionGreater(ctx *ConditionGreaterContext) {}

// ExitConditionGreater is called when production ConditionGreater is exited.
func (s *BaseCypherListener) ExitConditionGreater(ctx *ConditionGreaterContext) {}

// EnterConditionAny is called when production ConditionAny is entered.
func (s *BaseCypherListener) EnterConditionAny(ctx *ConditionAnyContext) {}

// ExitConditionAny is called when production ConditionAny is exited.
func (s *BaseCypherListener) ExitConditionAny(ctx *ConditionAnyContext) {}

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
