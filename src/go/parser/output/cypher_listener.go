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

	// EnterConditionOr is called when entering the ConditionOr production.
	EnterConditionOr(c *ConditionOrContext)

	// EnterConditionNot is called when entering the ConditionNot production.
	EnterConditionNot(c *ConditionNotContext)

	// EnterConditionParen is called when entering the conditionParen production.
	EnterConditionParen(c *ConditionParenContext)

	// EnterConditionNone is called when entering the ConditionNone production.
	EnterConditionNone(c *ConditionNoneContext)

	// EnterConditionAll is called when entering the ConditionAll production.
	EnterConditionAll(c *ConditionAllContext)

	// EnterConditionGreater is called when entering the ConditionGreater production.
	EnterConditionGreater(c *ConditionGreaterContext)

	// EnterConditionAny is called when entering the ConditionAny production.
	EnterConditionAny(c *ConditionAnyContext)

	// EnterConditionNotEquality is called when entering the ConditionNotEquality production.
	EnterConditionNotEquality(c *ConditionNotEqualityContext)

	// EnterConditionLess is called when entering the ConditionLess production.
	EnterConditionLess(c *ConditionLessContext)

	// EnterConditionSingle is called when entering the ConditionSingle production.
	EnterConditionSingle(c *ConditionSingleContext)

	// EnterConditionEquality is called when entering the ConditionEquality production.
	EnterConditionEquality(c *ConditionEqualityContext)

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

	// ExitConditionOr is called when exiting the ConditionOr production.
	ExitConditionOr(c *ConditionOrContext)

	// ExitConditionNot is called when exiting the ConditionNot production.
	ExitConditionNot(c *ConditionNotContext)

	// ExitConditionParen is called when exiting the conditionParen production.
	ExitConditionParen(c *ConditionParenContext)

	// ExitConditionNone is called when exiting the ConditionNone production.
	ExitConditionNone(c *ConditionNoneContext)

	// ExitConditionAll is called when exiting the ConditionAll production.
	ExitConditionAll(c *ConditionAllContext)

	// ExitConditionGreater is called when exiting the ConditionGreater production.
	ExitConditionGreater(c *ConditionGreaterContext)

	// ExitConditionAny is called when exiting the ConditionAny production.
	ExitConditionAny(c *ConditionAnyContext)

	// ExitConditionNotEquality is called when exiting the ConditionNotEquality production.
	ExitConditionNotEquality(c *ConditionNotEqualityContext)

	// ExitConditionLess is called when exiting the ConditionLess production.
	ExitConditionLess(c *ConditionLessContext)

	// ExitConditionSingle is called when exiting the ConditionSingle production.
	ExitConditionSingle(c *ConditionSingleContext)

	// ExitConditionEquality is called when exiting the ConditionEquality production.
	ExitConditionEquality(c *ConditionEqualityContext)

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
