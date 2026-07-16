// Generated from /Users/wakabayashitaku/polystore_database/src/go/parser/Cypher.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.tree.ParseTreeListener;

/**
 * This interface defines a complete listener for a parse tree produced by
 * {@link CypherParser}.
 */
public interface CypherListener extends ParseTreeListener {
	/**
	 * Enter a parse tree produced by {@link CypherParser#cypher}.
	 * @param ctx the parse tree
	 */
	void enterCypher(CypherParser.CypherContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#cypher}.
	 * @param ctx the parse tree
	 */
	void exitCypher(CypherParser.CypherContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#statement}.
	 * @param ctx the parse tree
	 */
	void enterStatement(CypherParser.StatementContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#statement}.
	 * @param ctx the parse tree
	 */
	void exitStatement(CypherParser.StatementContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#matchClause}.
	 * @param ctx the parse tree
	 */
	void enterMatchClause(CypherParser.MatchClauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#matchClause}.
	 * @param ctx the parse tree
	 */
	void exitMatchClause(CypherParser.MatchClauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#returnClause}.
	 * @param ctx the parse tree
	 */
	void enterReturnClause(CypherParser.ReturnClauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#returnClause}.
	 * @param ctx the parse tree
	 */
	void exitReturnClause(CypherParser.ReturnClauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#whereClause}.
	 * @param ctx the parse tree
	 */
	void enterWhereClause(CypherParser.WhereClauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#whereClause}.
	 * @param ctx the parse tree
	 */
	void exitWhereClause(CypherParser.WhereClauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#callClause}.
	 * @param ctx the parse tree
	 */
	void enterCallClause(CypherParser.CallClauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#callClause}.
	 * @param ctx the parse tree
	 */
	void exitCallClause(CypherParser.CallClauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#asClause}.
	 * @param ctx the parse tree
	 */
	void enterAsClause(CypherParser.AsClauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#asClause}.
	 * @param ctx the parse tree
	 */
	void exitAsClause(CypherParser.AsClauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#pattern}.
	 * @param ctx the parse tree
	 */
	void enterPattern(CypherParser.PatternContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#pattern}.
	 * @param ctx the parse tree
	 */
	void exitPattern(CypherParser.PatternContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#node}.
	 * @param ctx the parse tree
	 */
	void enterNode(CypherParser.NodeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#node}.
	 * @param ctx the parse tree
	 */
	void exitNode(CypherParser.NodeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#relationship}.
	 * @param ctx the parse tree
	 */
	void enterRelationship(CypherParser.RelationshipContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#relationship}.
	 * @param ctx the parse tree
	 */
	void exitRelationship(CypherParser.RelationshipContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#returnItems}.
	 * @param ctx the parse tree
	 */
	void enterReturnItems(CypherParser.ReturnItemsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#returnItems}.
	 * @param ctx the parse tree
	 */
	void exitReturnItems(CypherParser.ReturnItemsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#returnItem}.
	 * @param ctx the parse tree
	 */
	void enterReturnItem(CypherParser.ReturnItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#returnItem}.
	 * @param ctx the parse tree
	 */
	void exitReturnItem(CypherParser.ReturnItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#aggregateFunc}.
	 * @param ctx the parse tree
	 */
	void enterAggregateFunc(CypherParser.AggregateFuncContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#aggregateFunc}.
	 * @param ctx the parse tree
	 */
	void exitAggregateFunc(CypherParser.AggregateFuncContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#aggArg}.
	 * @param ctx the parse tree
	 */
	void enterAggArg(CypherParser.AggArgContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#aggArg}.
	 * @param ctx the parse tree
	 */
	void exitAggArg(CypherParser.AggArgContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#orderItems}.
	 * @param ctx the parse tree
	 */
	void enterOrderItems(CypherParser.OrderItemsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#orderItems}.
	 * @param ctx the parse tree
	 */
	void exitOrderItems(CypherParser.OrderItemsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#orderItem}.
	 * @param ctx the parse tree
	 */
	void enterOrderItem(CypherParser.OrderItemContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#orderItem}.
	 * @param ctx the parse tree
	 */
	void exitOrderItem(CypherParser.OrderItemContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#limitNum}.
	 * @param ctx the parse tree
	 */
	void enterLimitNum(CypherParser.LimitNumContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#limitNum}.
	 * @param ctx the parse tree
	 */
	void exitLimitNum(CypherParser.LimitNumContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#labels}.
	 * @param ctx the parse tree
	 */
	void enterLabels(CypherParser.LabelsContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#labels}.
	 * @param ctx the parse tree
	 */
	void exitLabels(CypherParser.LabelsContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#label}.
	 * @param ctx the parse tree
	 */
	void enterLabel(CypherParser.LabelContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#label}.
	 * @param ctx the parse tree
	 */
	void exitLabel(CypherParser.LabelContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#properties}.
	 * @param ctx the parse tree
	 */
	void enterProperties(CypherParser.PropertiesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#properties}.
	 * @param ctx the parse tree
	 */
	void exitProperties(CypherParser.PropertiesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#property}.
	 * @param ctx the parse tree
	 */
	void enterProperty(CypherParser.PropertyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#property}.
	 * @param ctx the parse tree
	 */
	void exitProperty(CypherParser.PropertyContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionAnd}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionAnd(CypherParser.ConditionAndContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionAnd}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionAnd(CypherParser.ConditionAndContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionNot}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionNot(CypherParser.ConditionNotContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionNot}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionNot(CypherParser.ConditionNotContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionGreater}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionGreater(CypherParser.ConditionGreaterContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionGreater}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionGreater(CypherParser.ConditionGreaterContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionAny}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionAny(CypherParser.ConditionAnyContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionAny}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionAny(CypherParser.ConditionAnyContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionOr}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionOr(CypherParser.ConditionOrContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionOr}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionOr(CypherParser.ConditionOrContext ctx);
	/**
	 * Enter a parse tree produced by the {@code conditionParen}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionParen(CypherParser.ConditionParenContext ctx);
	/**
	 * Exit a parse tree produced by the {@code conditionParen}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionParen(CypherParser.ConditionParenContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionLessEqual}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionLessEqual(CypherParser.ConditionLessEqualContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionLessEqual}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionLessEqual(CypherParser.ConditionLessEqualContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionNone}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionNone(CypherParser.ConditionNoneContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionNone}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionNone(CypherParser.ConditionNoneContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionGreaterEqual}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionGreaterEqual(CypherParser.ConditionGreaterEqualContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionGreaterEqual}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionGreaterEqual(CypherParser.ConditionGreaterEqualContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionAll}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionAll(CypherParser.ConditionAllContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionAll}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionAll(CypherParser.ConditionAllContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionNotEquality}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionNotEquality(CypherParser.ConditionNotEqualityContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionNotEquality}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionNotEquality(CypherParser.ConditionNotEqualityContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionLess}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionLess(CypherParser.ConditionLessContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionLess}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionLess(CypherParser.ConditionLessContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionSingle}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionSingle(CypherParser.ConditionSingleContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionSingle}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionSingle(CypherParser.ConditionSingleContext ctx);
	/**
	 * Enter a parse tree produced by the {@code ConditionEquality}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void enterConditionEquality(CypherParser.ConditionEqualityContext ctx);
	/**
	 * Exit a parse tree produced by the {@code ConditionEquality}
	 * labeled alternative in {@link CypherParser#condition}.
	 * @param ctx the parse tree
	 */
	void exitConditionEquality(CypherParser.ConditionEqualityContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#variable}.
	 * @param ctx the parse tree
	 */
	void enterVariable(CypherParser.VariableContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#variable}.
	 * @param ctx the parse tree
	 */
	void exitVariable(CypherParser.VariableContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#types}.
	 * @param ctx the parse tree
	 */
	void enterTypes(CypherParser.TypesContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#types}.
	 * @param ctx the parse tree
	 */
	void exitTypes(CypherParser.TypesContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#expression}.
	 * @param ctx the parse tree
	 */
	void enterExpression(CypherParser.ExpressionContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#expression}.
	 * @param ctx the parse tree
	 */
	void exitExpression(CypherParser.ExpressionContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#value}.
	 * @param ctx the parse tree
	 */
	void enterValue(CypherParser.ValueContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#value}.
	 * @param ctx the parse tree
	 */
	void exitValue(CypherParser.ValueContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#range}.
	 * @param ctx the parse tree
	 */
	void enterRange(CypherParser.RangeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#range}.
	 * @param ctx the parse tree
	 */
	void exitRange(CypherParser.RangeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CypherParser#rangeLiteral}.
	 * @param ctx the parse tree
	 */
	void enterRangeLiteral(CypherParser.RangeLiteralContext ctx);
	/**
	 * Exit a parse tree produced by {@link CypherParser#rangeLiteral}.
	 * @param ctx the parse tree
	 */
	void exitRangeLiteral(CypherParser.RangeLiteralContext ctx);
}