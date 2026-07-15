// Generated from /Users/wakabayashitaku/polystore_database/src/go/parser/Cypher.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class CypherParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		T__0=1, MATCH=2, RETURN=3, WHERE=4, DISTINCT=5, AS=6, WITH=7, NEQ=8, DOT=9, 
		LARROW=10, RARROW=11, LANGLE=12, RANGLE=13, COLON=14, COMMA=15, LPAREN=16, 
		RPAREN=17, LSQUARE=18, RSQUARE=19, LCURLY=20, RCURLY=21, MINUS=22, SQUOTE=23, 
		STAR=24, DOUBLE_DOT=25, CREATE=26, DELETE=27, ORDER_BY=28, ASC=29, DESC=30, 
		LIMIT=31, OPTIONAL=32, UNWIND=33, FINISH=34, SET=35, EQ=36, AND=37, OR=38, 
		NOT=39, XOR=40, COUNT=41, REDUCE=42, SUM=43, AVG=44, MIN=45, MAX=46, COALESCE=47, 
		IN=48, ALL=49, ANY=50, NONE=51, SINGLE=52, CALL=53, STRING=54, NUMBER=55, 
		IDENTIFIER=56, WS=57;
	public static final int
		RULE_cypher = 0, RULE_statement = 1, RULE_matchClause = 2, RULE_returnClause = 3, 
		RULE_whereClause = 4, RULE_callClause = 5, RULE_asClause = 6, RULE_pattern = 7, 
		RULE_node = 8, RULE_relationship = 9, RULE_returnItems = 10, RULE_returnItem = 11, 
		RULE_aggregateFunc = 12, RULE_aggArg = 13, RULE_orderItems = 14, RULE_orderItem = 15, 
		RULE_limitNum = 16, RULE_labels = 17, RULE_label = 18, RULE_properties = 19, 
		RULE_property = 20, RULE_condition = 21, RULE_variable = 22, RULE_types = 23, 
		RULE_expression = 24, RULE_value = 25, RULE_range = 26, RULE_rangeLiteral = 27;
	private static String[] makeRuleNames() {
		return new String[] {
			"cypher", "statement", "matchClause", "returnClause", "whereClause", 
			"callClause", "asClause", "pattern", "node", "relationship", "returnItems", 
			"returnItem", "aggregateFunc", "aggArg", "orderItems", "orderItem", "limitNum", 
			"labels", "label", "properties", "property", "condition", "variable", 
			"types", "expression", "value", "range", "rangeLiteral"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "'|'", null, null, null, null, null, "'WITH'", "'<>'", "'.'", "'<-'", 
			"'->'", "'<'", "'>'", "':'", "','", "'('", "')'", "'['", "']'", "'{'", 
			"'}'", "'-'", "'''", "'*'", "'..'", "'CREATE'", "'DELETE'", "'ORDER BY'", 
			"'ASC'", "'DESC'", "'LIMIT'", "'OPTIONAL'", "'UNWIND'", "'FINISH'", "'SET'", 
			"'='", "'AND'", "'OR'", "'NOT'", "'XOR'"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, null, "MATCH", "RETURN", "WHERE", "DISTINCT", "AS", "WITH", "NEQ", 
			"DOT", "LARROW", "RARROW", "LANGLE", "RANGLE", "COLON", "COMMA", "LPAREN", 
			"RPAREN", "LSQUARE", "RSQUARE", "LCURLY", "RCURLY", "MINUS", "SQUOTE", 
			"STAR", "DOUBLE_DOT", "CREATE", "DELETE", "ORDER_BY", "ASC", "DESC", 
			"LIMIT", "OPTIONAL", "UNWIND", "FINISH", "SET", "EQ", "AND", "OR", "NOT", 
			"XOR", "COUNT", "REDUCE", "SUM", "AVG", "MIN", "MAX", "COALESCE", "IN", 
			"ALL", "ANY", "NONE", "SINGLE", "CALL", "STRING", "NUMBER", "IDENTIFIER", 
			"WS"
		};
	}
	private static final String[] _SYMBOLIC_NAMES = makeSymbolicNames();
	public static final Vocabulary VOCABULARY = new VocabularyImpl(_LITERAL_NAMES, _SYMBOLIC_NAMES);

	/**
	 * @deprecated Use {@link #VOCABULARY} instead.
	 */
	@Deprecated
	public static final String[] tokenNames;
	static {
		tokenNames = new String[_SYMBOLIC_NAMES.length];
		for (int i = 0; i < tokenNames.length; i++) {
			tokenNames[i] = VOCABULARY.getLiteralName(i);
			if (tokenNames[i] == null) {
				tokenNames[i] = VOCABULARY.getSymbolicName(i);
			}

			if (tokenNames[i] == null) {
				tokenNames[i] = "<INVALID>";
			}
		}
	}

	@Override
	@Deprecated
	public String[] getTokenNames() {
		return tokenNames;
	}

	@Override

	public Vocabulary getVocabulary() {
		return VOCABULARY;
	}

	@Override
	public String getGrammarFileName() { return "Cypher.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public CypherParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CypherContext extends ParserRuleContext {
		public TerminalNode EOF() { return getToken(CypherParser.EOF, 0); }
		public List<StatementContext> statement() {
			return getRuleContexts(StatementContext.class);
		}
		public StatementContext statement(int i) {
			return getRuleContext(StatementContext.class,i);
		}
		public CypherContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_cypher; }
	}

	public final CypherContext cypher() throws RecognitionException {
		CypherContext _localctx = new CypherContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_cypher);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(57); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(56);
				statement();
				}
				}
				setState(59); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==MATCH );
			setState(61);
			match(EOF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class StatementContext extends ParserRuleContext {
		public MatchClauseContext matchClause() {
			return getRuleContext(MatchClauseContext.class,0);
		}
		public ReturnClauseContext returnClause() {
			return getRuleContext(ReturnClauseContext.class,0);
		}
		public WhereClauseContext whereClause() {
			return getRuleContext(WhereClauseContext.class,0);
		}
		public StatementContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_statement; }
	}

	public final StatementContext statement() throws RecognitionException {
		StatementContext _localctx = new StatementContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_statement);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(63);
			matchClause();
			setState(65);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WHERE) {
				{
				setState(64);
				whereClause();
				}
			}

			setState(67);
			returnClause();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class MatchClauseContext extends ParserRuleContext {
		public TerminalNode MATCH() { return getToken(CypherParser.MATCH, 0); }
		public PatternContext pattern() {
			return getRuleContext(PatternContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(CypherParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(CypherParser.WS, i);
		}
		public MatchClauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_matchClause; }
	}

	public final MatchClauseContext matchClause() throws RecognitionException {
		MatchClauseContext _localctx = new MatchClauseContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_matchClause);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(69);
			match(MATCH);
			setState(73);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(70);
				match(WS);
				}
				}
				setState(75);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(76);
			pattern();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ReturnClauseContext extends ParserRuleContext {
		public TerminalNode RETURN() { return getToken(CypherParser.RETURN, 0); }
		public ReturnItemsContext returnItems() {
			return getRuleContext(ReturnItemsContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(CypherParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(CypherParser.WS, i);
		}
		public OrderItemsContext orderItems() {
			return getRuleContext(OrderItemsContext.class,0);
		}
		public LimitNumContext limitNum() {
			return getRuleContext(LimitNumContext.class,0);
		}
		public ReturnClauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_returnClause; }
	}

	public final ReturnClauseContext returnClause() throws RecognitionException {
		ReturnClauseContext _localctx = new ReturnClauseContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_returnClause);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(78);
			match(RETURN);
			setState(82);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(79);
				match(WS);
				}
				}
				setState(84);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(85);
			returnItems();
			setState(87);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==ORDER_BY) {
				{
				setState(86);
				orderItems();
				}
			}

			setState(90);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LIMIT) {
				{
				setState(89);
				limitNum();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class WhereClauseContext extends ParserRuleContext {
		public TerminalNode WHERE() { return getToken(CypherParser.WHERE, 0); }
		public ConditionContext condition() {
			return getRuleContext(ConditionContext.class,0);
		}
		public List<TerminalNode> WS() { return getTokens(CypherParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(CypherParser.WS, i);
		}
		public WhereClauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_whereClause; }
	}

	public final WhereClauseContext whereClause() throws RecognitionException {
		WhereClauseContext _localctx = new WhereClauseContext(_ctx, getState());
		enterRule(_localctx, 8, RULE_whereClause);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(92);
			match(WHERE);
			setState(96);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(93);
				match(WS);
				}
				}
				setState(98);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(99);
			condition(0);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class CallClauseContext extends ParserRuleContext {
		public TerminalNode CALL() { return getToken(CypherParser.CALL, 0); }
		public TerminalNode STRING() { return getToken(CypherParser.STRING, 0); }
		public List<TerminalNode> WS() { return getTokens(CypherParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(CypherParser.WS, i);
		}
		public CallClauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_callClause; }
	}

	public final CallClauseContext callClause() throws RecognitionException {
		CallClauseContext _localctx = new CallClauseContext(_ctx, getState());
		enterRule(_localctx, 10, RULE_callClause);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(101);
			match(CALL);
			setState(105);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(102);
				match(WS);
				}
				}
				setState(107);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(108);
			match(STRING);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AsClauseContext extends ParserRuleContext {
		public TerminalNode AS() { return getToken(CypherParser.AS, 0); }
		public TerminalNode IDENTIFIER() { return getToken(CypherParser.IDENTIFIER, 0); }
		public AsClauseContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_asClause; }
	}

	public final AsClauseContext asClause() throws RecognitionException {
		AsClauseContext _localctx = new AsClauseContext(_ctx, getState());
		enterRule(_localctx, 12, RULE_asClause);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(110);
			match(AS);
			setState(111);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PatternContext extends ParserRuleContext {
		public List<NodeContext> node() {
			return getRuleContexts(NodeContext.class);
		}
		public NodeContext node(int i) {
			return getRuleContext(NodeContext.class,i);
		}
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public TerminalNode EQ() { return getToken(CypherParser.EQ, 0); }
		public List<RelationshipContext> relationship() {
			return getRuleContexts(RelationshipContext.class);
		}
		public RelationshipContext relationship(int i) {
			return getRuleContext(RelationshipContext.class,i);
		}
		public PatternContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_pattern; }
	}

	public final PatternContext pattern() throws RecognitionException {
		PatternContext _localctx = new PatternContext(_ctx, getState());
		enterRule(_localctx, 14, RULE_pattern);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(116);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==IDENTIFIER) {
				{
				setState(113);
				variable();
				setState(114);
				match(EQ);
				}
			}

			setState(118);
			node();
			setState(124);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LARROW || _la==MINUS) {
				{
				{
				setState(119);
				relationship();
				setState(120);
				node();
				}
				}
				setState(126);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class NodeContext extends ParserRuleContext {
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public LabelsContext labels() {
			return getRuleContext(LabelsContext.class,0);
		}
		public PropertiesContext properties() {
			return getRuleContext(PropertiesContext.class,0);
		}
		public NodeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_node; }
	}

	public final NodeContext node() throws RecognitionException {
		NodeContext _localctx = new NodeContext(_ctx, getState());
		enterRule(_localctx, 16, RULE_node);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(127);
			match(LPAREN);
			setState(129);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==IDENTIFIER) {
				{
				setState(128);
				variable();
				}
			}

			setState(132);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==COLON) {
				{
				setState(131);
				labels();
				}
			}

			setState(135);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LCURLY) {
				{
				setState(134);
				properties();
				}
			}

			setState(137);
			match(RPAREN);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RelationshipContext extends ParserRuleContext {
		public List<TerminalNode> MINUS() { return getTokens(CypherParser.MINUS); }
		public TerminalNode MINUS(int i) {
			return getToken(CypherParser.MINUS, i);
		}
		public TerminalNode LSQUARE() { return getToken(CypherParser.LSQUARE, 0); }
		public TerminalNode RSQUARE() { return getToken(CypherParser.RSQUARE, 0); }
		public TerminalNode RARROW() { return getToken(CypherParser.RARROW, 0); }
		public List<TerminalNode> WS() { return getTokens(CypherParser.WS); }
		public TerminalNode WS(int i) {
			return getToken(CypherParser.WS, i);
		}
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public TypesContext types() {
			return getRuleContext(TypesContext.class,0);
		}
		public RangeContext range() {
			return getRuleContext(RangeContext.class,0);
		}
		public PropertiesContext properties() {
			return getRuleContext(PropertiesContext.class,0);
		}
		public TerminalNode LARROW() { return getToken(CypherParser.LARROW, 0); }
		public RelationshipContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_relationship; }
	}

	public final RelationshipContext relationship() throws RecognitionException {
		RelationshipContext _localctx = new RelationshipContext(_ctx, getState());
		enterRule(_localctx, 18, RULE_relationship);
		int _la;
		try {
			setState(223);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,31,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(139);
				match(MINUS);
				setState(143);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(140);
					match(WS);
					}
					}
					setState(145);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(146);
				match(LSQUARE);
				setState(148);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==IDENTIFIER) {
					{
					setState(147);
					variable();
					}
				}

				setState(151);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==COLON) {
					{
					setState(150);
					types();
					}
				}

				setState(154);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==STAR) {
					{
					setState(153);
					range();
					}
				}

				setState(157);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==LCURLY) {
					{
					setState(156);
					properties();
					}
				}

				setState(159);
				match(RSQUARE);
				setState(163);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(160);
					match(WS);
					}
					}
					setState(165);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(166);
				match(RARROW);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(167);
				match(LARROW);
				setState(171);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(168);
					match(WS);
					}
					}
					setState(173);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(174);
				match(LSQUARE);
				setState(176);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==IDENTIFIER) {
					{
					setState(175);
					variable();
					}
				}

				setState(179);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==COLON) {
					{
					setState(178);
					types();
					}
				}

				setState(182);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==STAR) {
					{
					setState(181);
					range();
					}
				}

				setState(185);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==LCURLY) {
					{
					setState(184);
					properties();
					}
				}

				setState(187);
				match(RSQUARE);
				setState(191);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(188);
					match(WS);
					}
					}
					setState(193);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(194);
				match(MINUS);
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(195);
				match(MINUS);
				setState(199);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(196);
					match(WS);
					}
					}
					setState(201);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(202);
				match(LSQUARE);
				setState(204);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==IDENTIFIER) {
					{
					setState(203);
					variable();
					}
				}

				setState(207);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==COLON) {
					{
					setState(206);
					types();
					}
				}

				setState(210);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==STAR) {
					{
					setState(209);
					range();
					}
				}

				setState(213);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==LCURLY) {
					{
					setState(212);
					properties();
					}
				}

				setState(215);
				match(RSQUARE);
				setState(219);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(216);
					match(WS);
					}
					}
					setState(221);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(222);
				match(MINUS);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ReturnItemsContext extends ParserRuleContext {
		public List<ReturnItemContext> returnItem() {
			return getRuleContexts(ReturnItemContext.class);
		}
		public ReturnItemContext returnItem(int i) {
			return getRuleContext(ReturnItemContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(CypherParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(CypherParser.COMMA, i);
		}
		public ReturnItemsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_returnItems; }
	}

	public final ReturnItemsContext returnItems() throws RecognitionException {
		ReturnItemsContext _localctx = new ReturnItemsContext(_ctx, getState());
		enterRule(_localctx, 20, RULE_returnItems);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(225);
			returnItem();
			setState(230);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(226);
				match(COMMA);
				setState(227);
				returnItem();
				}
				}
				setState(232);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ReturnItemContext extends ParserRuleContext {
		public AggregateFuncContext aggregateFunc() {
			return getRuleContext(AggregateFuncContext.class,0);
		}
		public TerminalNode AS() { return getToken(CypherParser.AS, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode COALESCE() { return getToken(CypherParser.COALESCE, 0); }
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public List<TerminalNode> COMMA() { return getTokens(CypherParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(CypherParser.COMMA, i);
		}
		public ReturnItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_returnItem; }
	}

	public final ReturnItemContext returnItem() throws RecognitionException {
		ReturnItemContext _localctx = new ReturnItemContext(_ctx, getState());
		enterRule(_localctx, 22, RULE_returnItem);
		int _la;
		try {
			setState(258);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case COUNT:
			case SUM:
			case AVG:
			case MIN:
			case MAX:
				enterOuterAlt(_localctx, 1);
				{
				setState(233);
				aggregateFunc();
				setState(236);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS) {
					{
					setState(234);
					match(AS);
					setState(235);
					variable();
					}
				}

				}
				break;
			case IDENTIFIER:
				enterOuterAlt(_localctx, 2);
				{
				setState(238);
				expression();
				setState(241);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS) {
					{
					setState(239);
					match(AS);
					setState(240);
					variable();
					}
				}

				}
				break;
			case COALESCE:
				enterOuterAlt(_localctx, 3);
				{
				setState(243);
				match(COALESCE);
				setState(244);
				match(LPAREN);
				setState(245);
				expression();
				setState(250);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(246);
					match(COMMA);
					setState(247);
					expression();
					}
					}
					setState(252);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(253);
				match(RPAREN);
				setState(256);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS) {
					{
					setState(254);
					match(AS);
					setState(255);
					variable();
					}
				}

				}
				break;
			default:
				throw new NoViableAltException(this);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AggregateFuncContext extends ParserRuleContext {
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public TerminalNode COUNT() { return getToken(CypherParser.COUNT, 0); }
		public TerminalNode SUM() { return getToken(CypherParser.SUM, 0); }
		public TerminalNode AVG() { return getToken(CypherParser.AVG, 0); }
		public TerminalNode MIN() { return getToken(CypherParser.MIN, 0); }
		public TerminalNode MAX() { return getToken(CypherParser.MAX, 0); }
		public TerminalNode STAR() { return getToken(CypherParser.STAR, 0); }
		public AggArgContext aggArg() {
			return getRuleContext(AggArgContext.class,0);
		}
		public TerminalNode DISTINCT() { return getToken(CypherParser.DISTINCT, 0); }
		public AggregateFuncContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_aggregateFunc; }
	}

	public final AggregateFuncContext aggregateFunc() throws RecognitionException {
		AggregateFuncContext _localctx = new AggregateFuncContext(_ctx, getState());
		enterRule(_localctx, 24, RULE_aggregateFunc);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(260);
			_la = _input.LA(1);
			if ( !((((_la) & ~0x3f) == 0 && ((1L << _la) & 134140418588672L) != 0)) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			setState(261);
			match(LPAREN);
			setState(267);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case STAR:
				{
				setState(262);
				match(STAR);
				}
				break;
			case DISTINCT:
			case IDENTIFIER:
				{
				setState(264);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==DISTINCT) {
					{
					setState(263);
					match(DISTINCT);
					}
				}

				setState(266);
				aggArg();
				}
				break;
			default:
				throw new NoViableAltException(this);
			}
			setState(269);
			match(RPAREN);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class AggArgContext extends ParserRuleContext {
		public List<TerminalNode> IDENTIFIER() { return getTokens(CypherParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(CypherParser.IDENTIFIER, i);
		}
		public TerminalNode DOT() { return getToken(CypherParser.DOT, 0); }
		public AggArgContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_aggArg; }
	}

	public final AggArgContext aggArg() throws RecognitionException {
		AggArgContext _localctx = new AggArgContext(_ctx, getState());
		enterRule(_localctx, 26, RULE_aggArg);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(271);
			match(IDENTIFIER);
			setState(274);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==DOT) {
				{
				setState(272);
				match(DOT);
				setState(273);
				match(IDENTIFIER);
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OrderItemsContext extends ParserRuleContext {
		public TerminalNode ORDER_BY() { return getToken(CypherParser.ORDER_BY, 0); }
		public List<OrderItemContext> orderItem() {
			return getRuleContexts(OrderItemContext.class);
		}
		public OrderItemContext orderItem(int i) {
			return getRuleContext(OrderItemContext.class,i);
		}
		public List<TerminalNode> COMMA() { return getTokens(CypherParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(CypherParser.COMMA, i);
		}
		public OrderItemsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_orderItems; }
	}

	public final OrderItemsContext orderItems() throws RecognitionException {
		OrderItemsContext _localctx = new OrderItemsContext(_ctx, getState());
		enterRule(_localctx, 28, RULE_orderItems);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(276);
			match(ORDER_BY);
			setState(277);
			orderItem();
			setState(282);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(278);
				match(COMMA);
				setState(279);
				orderItem();
				}
				}
				setState(284);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class OrderItemContext extends ParserRuleContext {
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode ASC() { return getToken(CypherParser.ASC, 0); }
		public TerminalNode DESC() { return getToken(CypherParser.DESC, 0); }
		public OrderItemContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_orderItem; }
	}

	public final OrderItemContext orderItem() throws RecognitionException {
		OrderItemContext _localctx = new OrderItemContext(_ctx, getState());
		enterRule(_localctx, 30, RULE_orderItem);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(285);
			expression();
			setState(286);
			_la = _input.LA(1);
			if ( !(_la==ASC || _la==DESC) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class LimitNumContext extends ParserRuleContext {
		public TerminalNode LIMIT() { return getToken(CypherParser.LIMIT, 0); }
		public TerminalNode NUMBER() { return getToken(CypherParser.NUMBER, 0); }
		public LimitNumContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_limitNum; }
	}

	public final LimitNumContext limitNum() throws RecognitionException {
		LimitNumContext _localctx = new LimitNumContext(_ctx, getState());
		enterRule(_localctx, 32, RULE_limitNum);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(288);
			match(LIMIT);
			setState(289);
			match(NUMBER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class LabelsContext extends ParserRuleContext {
		public TerminalNode COLON() { return getToken(CypherParser.COLON, 0); }
		public List<LabelContext> label() {
			return getRuleContexts(LabelContext.class);
		}
		public LabelContext label(int i) {
			return getRuleContext(LabelContext.class,i);
		}
		public LabelsContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_labels; }
	}

	public final LabelsContext labels() throws RecognitionException {
		LabelsContext _localctx = new LabelsContext(_ctx, getState());
		enterRule(_localctx, 34, RULE_labels);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(291);
			match(COLON);
			setState(293); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(292);
				label();
				}
				}
				setState(295); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==IDENTIFIER );
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class LabelContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(CypherParser.IDENTIFIER, 0); }
		public LabelContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_label; }
	}

	public final LabelContext label() throws RecognitionException {
		LabelContext _localctx = new LabelContext(_ctx, getState());
		enterRule(_localctx, 36, RULE_label);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(297);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PropertiesContext extends ParserRuleContext {
		public TerminalNode LCURLY() { return getToken(CypherParser.LCURLY, 0); }
		public List<PropertyContext> property() {
			return getRuleContexts(PropertyContext.class);
		}
		public PropertyContext property(int i) {
			return getRuleContext(PropertyContext.class,i);
		}
		public TerminalNode RCURLY() { return getToken(CypherParser.RCURLY, 0); }
		public List<TerminalNode> COMMA() { return getTokens(CypherParser.COMMA); }
		public TerminalNode COMMA(int i) {
			return getToken(CypherParser.COMMA, i);
		}
		public PropertiesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_properties; }
	}

	public final PropertiesContext properties() throws RecognitionException {
		PropertiesContext _localctx = new PropertiesContext(_ctx, getState());
		enterRule(_localctx, 38, RULE_properties);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(299);
			match(LCURLY);
			setState(300);
			property();
			setState(305);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(301);
				match(COMMA);
				setState(302);
				property();
				}
				}
				setState(307);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(308);
			match(RCURLY);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class PropertyContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(CypherParser.IDENTIFIER, 0); }
		public TerminalNode COLON() { return getToken(CypherParser.COLON, 0); }
		public ValueContext value() {
			return getRuleContext(ValueContext.class,0);
		}
		public PropertyContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_property; }
	}

	public final PropertyContext property() throws RecognitionException {
		PropertyContext _localctx = new PropertyContext(_ctx, getState());
		enterRule(_localctx, 40, RULE_property);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(310);
			match(IDENTIFIER);
			setState(311);
			match(COLON);
			setState(312);
			value();
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ConditionContext extends ParserRuleContext {
		public ConditionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_condition; }
	 
		public ConditionContext() { }
		public void copyFrom(ConditionContext ctx) {
			super.copyFrom(ctx);
		}
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionAndContext extends ConditionContext {
		public List<ConditionContext> condition() {
			return getRuleContexts(ConditionContext.class);
		}
		public ConditionContext condition(int i) {
			return getRuleContext(ConditionContext.class,i);
		}
		public TerminalNode AND() { return getToken(CypherParser.AND, 0); }
		public ConditionAndContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionOrContext extends ConditionContext {
		public List<ConditionContext> condition() {
			return getRuleContexts(ConditionContext.class);
		}
		public ConditionContext condition(int i) {
			return getRuleContext(ConditionContext.class,i);
		}
		public TerminalNode OR() { return getToken(CypherParser.OR, 0); }
		public ConditionOrContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionNotContext extends ConditionContext {
		public TerminalNode NOT() { return getToken(CypherParser.NOT, 0); }
		public ConditionContext condition() {
			return getRuleContext(ConditionContext.class,0);
		}
		public ConditionNotContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionParenContext extends ConditionContext {
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public ConditionContext condition() {
			return getRuleContext(ConditionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public ConditionParenContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionNoneContext extends ConditionContext {
		public TerminalNode NONE() { return getToken(CypherParser.NONE, 0); }
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public TerminalNode IN() { return getToken(CypherParser.IN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode WHERE() { return getToken(CypherParser.WHERE, 0); }
		public ConditionContext condition() {
			return getRuleContext(ConditionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public ConditionNoneContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionAllContext extends ConditionContext {
		public TerminalNode ALL() { return getToken(CypherParser.ALL, 0); }
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public TerminalNode IN() { return getToken(CypherParser.IN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode WHERE() { return getToken(CypherParser.WHERE, 0); }
		public ConditionContext condition() {
			return getRuleContext(ConditionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public ConditionAllContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionGreaterContext extends ConditionContext {
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode RANGLE() { return getToken(CypherParser.RANGLE, 0); }
		public ValueContext value() {
			return getRuleContext(ValueContext.class,0);
		}
		public ConditionGreaterContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionAnyContext extends ConditionContext {
		public TerminalNode ANY() { return getToken(CypherParser.ANY, 0); }
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public TerminalNode IN() { return getToken(CypherParser.IN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode WHERE() { return getToken(CypherParser.WHERE, 0); }
		public ConditionContext condition() {
			return getRuleContext(ConditionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public ConditionAnyContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionNotEqualityContext extends ConditionContext {
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode NEQ() { return getToken(CypherParser.NEQ, 0); }
		public ValueContext value() {
			return getRuleContext(ValueContext.class,0);
		}
		public ConditionNotEqualityContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionLessContext extends ConditionContext {
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode LANGLE() { return getToken(CypherParser.LANGLE, 0); }
		public ValueContext value() {
			return getRuleContext(ValueContext.class,0);
		}
		public ConditionLessContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionSingleContext extends ConditionContext {
		public TerminalNode SINGLE() { return getToken(CypherParser.SINGLE, 0); }
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public TerminalNode IN() { return getToken(CypherParser.IN, 0); }
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode WHERE() { return getToken(CypherParser.WHERE, 0); }
		public ConditionContext condition() {
			return getRuleContext(ConditionContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public ConditionSingleContext(ConditionContext ctx) { copyFrom(ctx); }
	}
	@SuppressWarnings("CheckReturnValue")
	public static class ConditionEqualityContext extends ConditionContext {
		public ExpressionContext expression() {
			return getRuleContext(ExpressionContext.class,0);
		}
		public TerminalNode EQ() { return getToken(CypherParser.EQ, 0); }
		public ValueContext value() {
			return getRuleContext(ValueContext.class,0);
		}
		public ConditionEqualityContext(ConditionContext ctx) { copyFrom(ctx); }
	}

	public final ConditionContext condition() throws RecognitionException {
		return condition(0);
	}

	private ConditionContext condition(int _p) throws RecognitionException {
		ParserRuleContext _parentctx = _ctx;
		int _parentState = getState();
		ConditionContext _localctx = new ConditionContext(_ctx, _parentState);
		ConditionContext _prevctx = _localctx;
		int _startState = 42;
		enterRecursionRule(_localctx, 42, RULE_condition, _p);
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(373);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,44,_ctx) ) {
			case 1:
				{
				_localctx = new ConditionParenContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;

				setState(315);
				match(LPAREN);
				setState(316);
				condition(0);
				setState(317);
				match(RPAREN);
				}
				break;
			case 2:
				{
				_localctx = new ConditionNotContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(319);
				match(NOT);
				setState(320);
				condition(11);
				}
				break;
			case 3:
				{
				_localctx = new ConditionAllContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(321);
				match(ALL);
				setState(322);
				match(LPAREN);
				setState(323);
				variable();
				setState(324);
				match(IN);
				setState(325);
				expression();
				setState(326);
				match(WHERE);
				setState(327);
				condition(0);
				setState(328);
				match(RPAREN);
				}
				break;
			case 4:
				{
				_localctx = new ConditionAnyContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(330);
				match(ANY);
				setState(331);
				match(LPAREN);
				setState(332);
				variable();
				setState(333);
				match(IN);
				setState(334);
				expression();
				setState(335);
				match(WHERE);
				setState(336);
				condition(0);
				setState(337);
				match(RPAREN);
				}
				break;
			case 5:
				{
				_localctx = new ConditionNoneContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(339);
				match(NONE);
				setState(340);
				match(LPAREN);
				setState(341);
				variable();
				setState(342);
				match(IN);
				setState(343);
				expression();
				setState(344);
				match(WHERE);
				setState(345);
				condition(0);
				setState(346);
				match(RPAREN);
				}
				break;
			case 6:
				{
				_localctx = new ConditionSingleContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(348);
				match(SINGLE);
				setState(349);
				match(LPAREN);
				setState(350);
				variable();
				setState(351);
				match(IN);
				setState(352);
				expression();
				setState(353);
				match(WHERE);
				setState(354);
				condition(0);
				setState(355);
				match(RPAREN);
				}
				break;
			case 7:
				{
				_localctx = new ConditionEqualityContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(357);
				expression();
				setState(358);
				match(EQ);
				setState(359);
				value();
				}
				break;
			case 8:
				{
				_localctx = new ConditionNotEqualityContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(361);
				expression();
				setState(362);
				match(NEQ);
				setState(363);
				value();
				}
				break;
			case 9:
				{
				_localctx = new ConditionGreaterContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(365);
				expression();
				setState(366);
				match(RANGLE);
				setState(367);
				value();
				}
				break;
			case 10:
				{
				_localctx = new ConditionLessContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(369);
				expression();
				setState(370);
				match(LANGLE);
				setState(371);
				value();
				}
				break;
			}
			_ctx.stop = _input.LT(-1);
			setState(383);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,46,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					setState(381);
					_errHandler.sync(this);
					switch ( getInterpreter().adaptivePredict(_input,45,_ctx) ) {
					case 1:
						{
						_localctx = new ConditionAndContext(new ConditionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_condition);
						setState(375);
						if (!(precpred(_ctx, 10))) throw new FailedPredicateException(this, "precpred(_ctx, 10)");
						setState(376);
						match(AND);
						setState(377);
						condition(11);
						}
						break;
					case 2:
						{
						_localctx = new ConditionOrContext(new ConditionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_condition);
						setState(378);
						if (!(precpred(_ctx, 9))) throw new FailedPredicateException(this, "precpred(_ctx, 9)");
						setState(379);
						match(OR);
						setState(380);
						condition(10);
						}
						break;
					}
					} 
				}
				setState(385);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,46,_ctx);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			unrollRecursionContexts(_parentctx);
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class VariableContext extends ParserRuleContext {
		public TerminalNode IDENTIFIER() { return getToken(CypherParser.IDENTIFIER, 0); }
		public VariableContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_variable; }
	}

	public final VariableContext variable() throws RecognitionException {
		VariableContext _localctx = new VariableContext(_ctx, getState());
		enterRule(_localctx, 44, RULE_variable);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(386);
			match(IDENTIFIER);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class TypesContext extends ParserRuleContext {
		public TerminalNode COLON() { return getToken(CypherParser.COLON, 0); }
		public List<TerminalNode> IDENTIFIER() { return getTokens(CypherParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(CypherParser.IDENTIFIER, i);
		}
		public TypesContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_types; }
	}

	public final TypesContext types() throws RecognitionException {
		TypesContext _localctx = new TypesContext(_ctx, getState());
		enterRule(_localctx, 46, RULE_types);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(388);
			match(COLON);
			setState(389);
			match(IDENTIFIER);
			setState(394);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==T__0) {
				{
				{
				setState(390);
				match(T__0);
				setState(391);
				match(IDENTIFIER);
				}
				}
				setState(396);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ExpressionContext extends ParserRuleContext {
		public List<TerminalNode> IDENTIFIER() { return getTokens(CypherParser.IDENTIFIER); }
		public TerminalNode IDENTIFIER(int i) {
			return getToken(CypherParser.IDENTIFIER, i);
		}
		public List<TerminalNode> DOT() { return getTokens(CypherParser.DOT); }
		public TerminalNode DOT(int i) {
			return getToken(CypherParser.DOT, i);
		}
		public TerminalNode LPAREN() { return getToken(CypherParser.LPAREN, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
		}
		public TerminalNode RPAREN() { return getToken(CypherParser.RPAREN, 0); }
		public ExpressionContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_expression; }
	}

	public final ExpressionContext expression() throws RecognitionException {
		ExpressionContext _localctx = new ExpressionContext(_ctx, getState());
		enterRule(_localctx, 48, RULE_expression);
		int _la;
		try {
			setState(410);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,49,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(397);
				match(IDENTIFIER);
				setState(402);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==DOT) {
					{
					{
					setState(398);
					match(DOT);
					setState(399);
					match(IDENTIFIER);
					}
					}
					setState(404);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(405);
				match(IDENTIFIER);
				setState(406);
				match(LPAREN);
				setState(407);
				variable();
				setState(408);
				match(RPAREN);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ValueContext extends ParserRuleContext {
		public TerminalNode STRING() { return getToken(CypherParser.STRING, 0); }
		public TerminalNode NUMBER() { return getToken(CypherParser.NUMBER, 0); }
		public ValueContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_value; }
	}

	public final ValueContext value() throws RecognitionException {
		ValueContext _localctx = new ValueContext(_ctx, getState());
		enterRule(_localctx, 50, RULE_value);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(412);
			_la = _input.LA(1);
			if ( !(_la==STRING || _la==NUMBER) ) {
			_errHandler.recoverInline(this);
			}
			else {
				if ( _input.LA(1)==Token.EOF ) matchedEOF = true;
				_errHandler.reportMatch(this);
				consume();
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RangeContext extends ParserRuleContext {
		public TerminalNode STAR() { return getToken(CypherParser.STAR, 0); }
		public RangeLiteralContext rangeLiteral() {
			return getRuleContext(RangeLiteralContext.class,0);
		}
		public RangeContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_range; }
	}

	public final RangeContext range() throws RecognitionException {
		RangeContext _localctx = new RangeContext(_ctx, getState());
		enterRule(_localctx, 52, RULE_range);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(414);
			match(STAR);
			setState(416);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==DOUBLE_DOT || _la==NUMBER) {
				{
				setState(415);
				rangeLiteral();
				}
			}

			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class RangeLiteralContext extends ParserRuleContext {
		public TerminalNode DOUBLE_DOT() { return getToken(CypherParser.DOUBLE_DOT, 0); }
		public List<TerminalNode> NUMBER() { return getTokens(CypherParser.NUMBER); }
		public TerminalNode NUMBER(int i) {
			return getToken(CypherParser.NUMBER, i);
		}
		public RangeLiteralContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_rangeLiteral; }
	}

	public final RangeLiteralContext rangeLiteral() throws RecognitionException {
		RangeLiteralContext _localctx = new RangeLiteralContext(_ctx, getState());
		enterRule(_localctx, 54, RULE_rangeLiteral);
		int _la;
		try {
			setState(426);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,53,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(419);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==NUMBER) {
					{
					setState(418);
					match(NUMBER);
					}
				}

				setState(421);
				match(DOUBLE_DOT);
				setState(423);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==NUMBER) {
					{
					setState(422);
					match(NUMBER);
					}
				}

				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(425);
				match(NUMBER);
				}
				break;
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public boolean sempred(RuleContext _localctx, int ruleIndex, int predIndex) {
		switch (ruleIndex) {
		case 21:
			return condition_sempred((ConditionContext)_localctx, predIndex);
		}
		return true;
	}
	private boolean condition_sempred(ConditionContext _localctx, int predIndex) {
		switch (predIndex) {
		case 0:
			return precpred(_ctx, 10);
		case 1:
			return precpred(_ctx, 9);
		}
		return true;
	}

	public static final String _serializedATN =
		"\u0004\u00019\u01ad\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004\u0002"+
		"\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007\u0002"+
		"\b\u0007\b\u0002\t\u0007\t\u0002\n\u0007\n\u0002\u000b\u0007\u000b\u0002"+
		"\f\u0007\f\u0002\r\u0007\r\u0002\u000e\u0007\u000e\u0002\u000f\u0007\u000f"+
		"\u0002\u0010\u0007\u0010\u0002\u0011\u0007\u0011\u0002\u0012\u0007\u0012"+
		"\u0002\u0013\u0007\u0013\u0002\u0014\u0007\u0014\u0002\u0015\u0007\u0015"+
		"\u0002\u0016\u0007\u0016\u0002\u0017\u0007\u0017\u0002\u0018\u0007\u0018"+
		"\u0002\u0019\u0007\u0019\u0002\u001a\u0007\u001a\u0002\u001b\u0007\u001b"+
		"\u0001\u0000\u0004\u0000:\b\u0000\u000b\u0000\f\u0000;\u0001\u0000\u0001"+
		"\u0000\u0001\u0001\u0001\u0001\u0003\u0001B\b\u0001\u0001\u0001\u0001"+
		"\u0001\u0001\u0002\u0001\u0002\u0005\u0002H\b\u0002\n\u0002\f\u0002K\t"+
		"\u0002\u0001\u0002\u0001\u0002\u0001\u0003\u0001\u0003\u0005\u0003Q\b"+
		"\u0003\n\u0003\f\u0003T\t\u0003\u0001\u0003\u0001\u0003\u0003\u0003X\b"+
		"\u0003\u0001\u0003\u0003\u0003[\b\u0003\u0001\u0004\u0001\u0004\u0005"+
		"\u0004_\b\u0004\n\u0004\f\u0004b\t\u0004\u0001\u0004\u0001\u0004\u0001"+
		"\u0005\u0001\u0005\u0005\u0005h\b\u0005\n\u0005\f\u0005k\t\u0005\u0001"+
		"\u0005\u0001\u0005\u0001\u0006\u0001\u0006\u0001\u0006\u0001\u0007\u0001"+
		"\u0007\u0001\u0007\u0003\u0007u\b\u0007\u0001\u0007\u0001\u0007\u0001"+
		"\u0007\u0001\u0007\u0005\u0007{\b\u0007\n\u0007\f\u0007~\t\u0007\u0001"+
		"\b\u0001\b\u0003\b\u0082\b\b\u0001\b\u0003\b\u0085\b\b\u0001\b\u0003\b"+
		"\u0088\b\b\u0001\b\u0001\b\u0001\t\u0001\t\u0005\t\u008e\b\t\n\t\f\t\u0091"+
		"\t\t\u0001\t\u0001\t\u0003\t\u0095\b\t\u0001\t\u0003\t\u0098\b\t\u0001"+
		"\t\u0003\t\u009b\b\t\u0001\t\u0003\t\u009e\b\t\u0001\t\u0001\t\u0005\t"+
		"\u00a2\b\t\n\t\f\t\u00a5\t\t\u0001\t\u0001\t\u0001\t\u0005\t\u00aa\b\t"+
		"\n\t\f\t\u00ad\t\t\u0001\t\u0001\t\u0003\t\u00b1\b\t\u0001\t\u0003\t\u00b4"+
		"\b\t\u0001\t\u0003\t\u00b7\b\t\u0001\t\u0003\t\u00ba\b\t\u0001\t\u0001"+
		"\t\u0005\t\u00be\b\t\n\t\f\t\u00c1\t\t\u0001\t\u0001\t\u0001\t\u0005\t"+
		"\u00c6\b\t\n\t\f\t\u00c9\t\t\u0001\t\u0001\t\u0003\t\u00cd\b\t\u0001\t"+
		"\u0003\t\u00d0\b\t\u0001\t\u0003\t\u00d3\b\t\u0001\t\u0003\t\u00d6\b\t"+
		"\u0001\t\u0001\t\u0005\t\u00da\b\t\n\t\f\t\u00dd\t\t\u0001\t\u0003\t\u00e0"+
		"\b\t\u0001\n\u0001\n\u0001\n\u0005\n\u00e5\b\n\n\n\f\n\u00e8\t\n\u0001"+
		"\u000b\u0001\u000b\u0001\u000b\u0003\u000b\u00ed\b\u000b\u0001\u000b\u0001"+
		"\u000b\u0001\u000b\u0003\u000b\u00f2\b\u000b\u0001\u000b\u0001\u000b\u0001"+
		"\u000b\u0001\u000b\u0001\u000b\u0005\u000b\u00f9\b\u000b\n\u000b\f\u000b"+
		"\u00fc\t\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0003\u000b\u0101\b"+
		"\u000b\u0003\u000b\u0103\b\u000b\u0001\f\u0001\f\u0001\f\u0001\f\u0003"+
		"\f\u0109\b\f\u0001\f\u0003\f\u010c\b\f\u0001\f\u0001\f\u0001\r\u0001\r"+
		"\u0001\r\u0003\r\u0113\b\r\u0001\u000e\u0001\u000e\u0001\u000e\u0001\u000e"+
		"\u0005\u000e\u0119\b\u000e\n\u000e\f\u000e\u011c\t\u000e\u0001\u000f\u0001"+
		"\u000f\u0001\u000f\u0001\u0010\u0001\u0010\u0001\u0010\u0001\u0011\u0001"+
		"\u0011\u0004\u0011\u0126\b\u0011\u000b\u0011\f\u0011\u0127\u0001\u0012"+
		"\u0001\u0012\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0005\u0013"+
		"\u0130\b\u0013\n\u0013\f\u0013\u0133\t\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0014\u0001\u0014\u0001\u0014\u0001\u0014\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0003\u0015\u0176\b\u0015\u0001\u0015\u0001"+
		"\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0005\u0015\u017e"+
		"\b\u0015\n\u0015\f\u0015\u0181\t\u0015\u0001\u0016\u0001\u0016\u0001\u0017"+
		"\u0001\u0017\u0001\u0017\u0001\u0017\u0005\u0017\u0189\b\u0017\n\u0017"+
		"\f\u0017\u018c\t\u0017\u0001\u0018\u0001\u0018\u0001\u0018\u0005\u0018"+
		"\u0191\b\u0018\n\u0018\f\u0018\u0194\t\u0018\u0001\u0018\u0001\u0018\u0001"+
		"\u0018\u0001\u0018\u0001\u0018\u0003\u0018\u019b\b\u0018\u0001\u0019\u0001"+
		"\u0019\u0001\u001a\u0001\u001a\u0003\u001a\u01a1\b\u001a\u0001\u001b\u0003"+
		"\u001b\u01a4\b\u001b\u0001\u001b\u0001\u001b\u0003\u001b\u01a8\b\u001b"+
		"\u0001\u001b\u0003\u001b\u01ab\b\u001b\u0001\u001b\u0000\u0001*\u001c"+
		"\u0000\u0002\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016\u0018\u001a"+
		"\u001c\u001e \"$&(*,.0246\u0000\u0003\u0002\u0000))+.\u0001\u0000\u001d"+
		"\u001e\u0001\u000067\u01d0\u00009\u0001\u0000\u0000\u0000\u0002?\u0001"+
		"\u0000\u0000\u0000\u0004E\u0001\u0000\u0000\u0000\u0006N\u0001\u0000\u0000"+
		"\u0000\b\\\u0001\u0000\u0000\u0000\ne\u0001\u0000\u0000\u0000\fn\u0001"+
		"\u0000\u0000\u0000\u000et\u0001\u0000\u0000\u0000\u0010\u007f\u0001\u0000"+
		"\u0000\u0000\u0012\u00df\u0001\u0000\u0000\u0000\u0014\u00e1\u0001\u0000"+
		"\u0000\u0000\u0016\u0102\u0001\u0000\u0000\u0000\u0018\u0104\u0001\u0000"+
		"\u0000\u0000\u001a\u010f\u0001\u0000\u0000\u0000\u001c\u0114\u0001\u0000"+
		"\u0000\u0000\u001e\u011d\u0001\u0000\u0000\u0000 \u0120\u0001\u0000\u0000"+
		"\u0000\"\u0123\u0001\u0000\u0000\u0000$\u0129\u0001\u0000\u0000\u0000"+
		"&\u012b\u0001\u0000\u0000\u0000(\u0136\u0001\u0000\u0000\u0000*\u0175"+
		"\u0001\u0000\u0000\u0000,\u0182\u0001\u0000\u0000\u0000.\u0184\u0001\u0000"+
		"\u0000\u00000\u019a\u0001\u0000\u0000\u00002\u019c\u0001\u0000\u0000\u0000"+
		"4\u019e\u0001\u0000\u0000\u00006\u01aa\u0001\u0000\u0000\u00008:\u0003"+
		"\u0002\u0001\u000098\u0001\u0000\u0000\u0000:;\u0001\u0000\u0000\u0000"+
		";9\u0001\u0000\u0000\u0000;<\u0001\u0000\u0000\u0000<=\u0001\u0000\u0000"+
		"\u0000=>\u0005\u0000\u0000\u0001>\u0001\u0001\u0000\u0000\u0000?A\u0003"+
		"\u0004\u0002\u0000@B\u0003\b\u0004\u0000A@\u0001\u0000\u0000\u0000AB\u0001"+
		"\u0000\u0000\u0000BC\u0001\u0000\u0000\u0000CD\u0003\u0006\u0003\u0000"+
		"D\u0003\u0001\u0000\u0000\u0000EI\u0005\u0002\u0000\u0000FH\u00059\u0000"+
		"\u0000GF\u0001\u0000\u0000\u0000HK\u0001\u0000\u0000\u0000IG\u0001\u0000"+
		"\u0000\u0000IJ\u0001\u0000\u0000\u0000JL\u0001\u0000\u0000\u0000KI\u0001"+
		"\u0000\u0000\u0000LM\u0003\u000e\u0007\u0000M\u0005\u0001\u0000\u0000"+
		"\u0000NR\u0005\u0003\u0000\u0000OQ\u00059\u0000\u0000PO\u0001\u0000\u0000"+
		"\u0000QT\u0001\u0000\u0000\u0000RP\u0001\u0000\u0000\u0000RS\u0001\u0000"+
		"\u0000\u0000SU\u0001\u0000\u0000\u0000TR\u0001\u0000\u0000\u0000UW\u0003"+
		"\u0014\n\u0000VX\u0003\u001c\u000e\u0000WV\u0001\u0000\u0000\u0000WX\u0001"+
		"\u0000\u0000\u0000XZ\u0001\u0000\u0000\u0000Y[\u0003 \u0010\u0000ZY\u0001"+
		"\u0000\u0000\u0000Z[\u0001\u0000\u0000\u0000[\u0007\u0001\u0000\u0000"+
		"\u0000\\`\u0005\u0004\u0000\u0000]_\u00059\u0000\u0000^]\u0001\u0000\u0000"+
		"\u0000_b\u0001\u0000\u0000\u0000`^\u0001\u0000\u0000\u0000`a\u0001\u0000"+
		"\u0000\u0000ac\u0001\u0000\u0000\u0000b`\u0001\u0000\u0000\u0000cd\u0003"+
		"*\u0015\u0000d\t\u0001\u0000\u0000\u0000ei\u00055\u0000\u0000fh\u0005"+
		"9\u0000\u0000gf\u0001\u0000\u0000\u0000hk\u0001\u0000\u0000\u0000ig\u0001"+
		"\u0000\u0000\u0000ij\u0001\u0000\u0000\u0000jl\u0001\u0000\u0000\u0000"+
		"ki\u0001\u0000\u0000\u0000lm\u00056\u0000\u0000m\u000b\u0001\u0000\u0000"+
		"\u0000no\u0005\u0006\u0000\u0000op\u00058\u0000\u0000p\r\u0001\u0000\u0000"+
		"\u0000qr\u0003,\u0016\u0000rs\u0005$\u0000\u0000su\u0001\u0000\u0000\u0000"+
		"tq\u0001\u0000\u0000\u0000tu\u0001\u0000\u0000\u0000uv\u0001\u0000\u0000"+
		"\u0000v|\u0003\u0010\b\u0000wx\u0003\u0012\t\u0000xy\u0003\u0010\b\u0000"+
		"y{\u0001\u0000\u0000\u0000zw\u0001\u0000\u0000\u0000{~\u0001\u0000\u0000"+
		"\u0000|z\u0001\u0000\u0000\u0000|}\u0001\u0000\u0000\u0000}\u000f\u0001"+
		"\u0000\u0000\u0000~|\u0001\u0000\u0000\u0000\u007f\u0081\u0005\u0010\u0000"+
		"\u0000\u0080\u0082\u0003,\u0016\u0000\u0081\u0080\u0001\u0000\u0000\u0000"+
		"\u0081\u0082\u0001\u0000\u0000\u0000\u0082\u0084\u0001\u0000\u0000\u0000"+
		"\u0083\u0085\u0003\"\u0011\u0000\u0084\u0083\u0001\u0000\u0000\u0000\u0084"+
		"\u0085\u0001\u0000\u0000\u0000\u0085\u0087\u0001\u0000\u0000\u0000\u0086"+
		"\u0088\u0003&\u0013\u0000\u0087\u0086\u0001\u0000\u0000\u0000\u0087\u0088"+
		"\u0001\u0000\u0000\u0000\u0088\u0089\u0001\u0000\u0000\u0000\u0089\u008a"+
		"\u0005\u0011\u0000\u0000\u008a\u0011\u0001\u0000\u0000\u0000\u008b\u008f"+
		"\u0005\u0016\u0000\u0000\u008c\u008e\u00059\u0000\u0000\u008d\u008c\u0001"+
		"\u0000\u0000\u0000\u008e\u0091\u0001\u0000\u0000\u0000\u008f\u008d\u0001"+
		"\u0000\u0000\u0000\u008f\u0090\u0001\u0000\u0000\u0000\u0090\u0092\u0001"+
		"\u0000\u0000\u0000\u0091\u008f\u0001\u0000\u0000\u0000\u0092\u0094\u0005"+
		"\u0012\u0000\u0000\u0093\u0095\u0003,\u0016\u0000\u0094\u0093\u0001\u0000"+
		"\u0000\u0000\u0094\u0095\u0001\u0000\u0000\u0000\u0095\u0097\u0001\u0000"+
		"\u0000\u0000\u0096\u0098\u0003.\u0017\u0000\u0097\u0096\u0001\u0000\u0000"+
		"\u0000\u0097\u0098\u0001\u0000\u0000\u0000\u0098\u009a\u0001\u0000\u0000"+
		"\u0000\u0099\u009b\u00034\u001a\u0000\u009a\u0099\u0001\u0000\u0000\u0000"+
		"\u009a\u009b\u0001\u0000\u0000\u0000\u009b\u009d\u0001\u0000\u0000\u0000"+
		"\u009c\u009e\u0003&\u0013\u0000\u009d\u009c\u0001\u0000\u0000\u0000\u009d"+
		"\u009e\u0001\u0000\u0000\u0000\u009e\u009f\u0001\u0000\u0000\u0000\u009f"+
		"\u00a3\u0005\u0013\u0000\u0000\u00a0\u00a2\u00059\u0000\u0000\u00a1\u00a0"+
		"\u0001\u0000\u0000\u0000\u00a2\u00a5\u0001\u0000\u0000\u0000\u00a3\u00a1"+
		"\u0001\u0000\u0000\u0000\u00a3\u00a4\u0001\u0000\u0000\u0000\u00a4\u00a6"+
		"\u0001\u0000\u0000\u0000\u00a5\u00a3\u0001\u0000\u0000\u0000\u00a6\u00e0"+
		"\u0005\u000b\u0000\u0000\u00a7\u00ab\u0005\n\u0000\u0000\u00a8\u00aa\u0005"+
		"9\u0000\u0000\u00a9\u00a8\u0001\u0000\u0000\u0000\u00aa\u00ad\u0001\u0000"+
		"\u0000\u0000\u00ab\u00a9\u0001\u0000\u0000\u0000\u00ab\u00ac\u0001\u0000"+
		"\u0000\u0000\u00ac\u00ae\u0001\u0000\u0000\u0000\u00ad\u00ab\u0001\u0000"+
		"\u0000\u0000\u00ae\u00b0\u0005\u0012\u0000\u0000\u00af\u00b1\u0003,\u0016"+
		"\u0000\u00b0\u00af\u0001\u0000\u0000\u0000\u00b0\u00b1\u0001\u0000\u0000"+
		"\u0000\u00b1\u00b3\u0001\u0000\u0000\u0000\u00b2\u00b4\u0003.\u0017\u0000"+
		"\u00b3\u00b2\u0001\u0000\u0000\u0000\u00b3\u00b4\u0001\u0000\u0000\u0000"+
		"\u00b4\u00b6\u0001\u0000\u0000\u0000\u00b5\u00b7\u00034\u001a\u0000\u00b6"+
		"\u00b5\u0001\u0000\u0000\u0000\u00b6\u00b7\u0001\u0000\u0000\u0000\u00b7"+
		"\u00b9\u0001\u0000\u0000\u0000\u00b8\u00ba\u0003&\u0013\u0000\u00b9\u00b8"+
		"\u0001\u0000\u0000\u0000\u00b9\u00ba\u0001\u0000\u0000\u0000\u00ba\u00bb"+
		"\u0001\u0000\u0000\u0000\u00bb\u00bf\u0005\u0013\u0000\u0000\u00bc\u00be"+
		"\u00059\u0000\u0000\u00bd\u00bc\u0001\u0000\u0000\u0000\u00be\u00c1\u0001"+
		"\u0000\u0000\u0000\u00bf\u00bd\u0001\u0000\u0000\u0000\u00bf\u00c0\u0001"+
		"\u0000\u0000\u0000\u00c0\u00c2\u0001\u0000\u0000\u0000\u00c1\u00bf\u0001"+
		"\u0000\u0000\u0000\u00c2\u00e0\u0005\u0016\u0000\u0000\u00c3\u00c7\u0005"+
		"\u0016\u0000\u0000\u00c4\u00c6\u00059\u0000\u0000\u00c5\u00c4\u0001\u0000"+
		"\u0000\u0000\u00c6\u00c9\u0001\u0000\u0000\u0000\u00c7\u00c5\u0001\u0000"+
		"\u0000\u0000\u00c7\u00c8\u0001\u0000\u0000\u0000\u00c8\u00ca\u0001\u0000"+
		"\u0000\u0000\u00c9\u00c7\u0001\u0000\u0000\u0000\u00ca\u00cc\u0005\u0012"+
		"\u0000\u0000\u00cb\u00cd\u0003,\u0016\u0000\u00cc\u00cb\u0001\u0000\u0000"+
		"\u0000\u00cc\u00cd\u0001\u0000\u0000\u0000\u00cd\u00cf\u0001\u0000\u0000"+
		"\u0000\u00ce\u00d0\u0003.\u0017\u0000\u00cf\u00ce\u0001\u0000\u0000\u0000"+
		"\u00cf\u00d0\u0001\u0000\u0000\u0000\u00d0\u00d2\u0001\u0000\u0000\u0000"+
		"\u00d1\u00d3\u00034\u001a\u0000\u00d2\u00d1\u0001\u0000\u0000\u0000\u00d2"+
		"\u00d3\u0001\u0000\u0000\u0000\u00d3\u00d5\u0001\u0000\u0000\u0000\u00d4"+
		"\u00d6\u0003&\u0013\u0000\u00d5\u00d4\u0001\u0000\u0000\u0000\u00d5\u00d6"+
		"\u0001\u0000\u0000\u0000\u00d6\u00d7\u0001\u0000\u0000\u0000\u00d7\u00db"+
		"\u0005\u0013\u0000\u0000\u00d8\u00da\u00059\u0000\u0000\u00d9\u00d8\u0001"+
		"\u0000\u0000\u0000\u00da\u00dd\u0001\u0000\u0000\u0000\u00db\u00d9\u0001"+
		"\u0000\u0000\u0000\u00db\u00dc\u0001\u0000\u0000\u0000\u00dc\u00de\u0001"+
		"\u0000\u0000\u0000\u00dd\u00db\u0001\u0000\u0000\u0000\u00de\u00e0\u0005"+
		"\u0016\u0000\u0000\u00df\u008b\u0001\u0000\u0000\u0000\u00df\u00a7\u0001"+
		"\u0000\u0000\u0000\u00df\u00c3\u0001\u0000\u0000\u0000\u00e0\u0013\u0001"+
		"\u0000\u0000\u0000\u00e1\u00e6\u0003\u0016\u000b\u0000\u00e2\u00e3\u0005"+
		"\u000f\u0000\u0000\u00e3\u00e5\u0003\u0016\u000b\u0000\u00e4\u00e2\u0001"+
		"\u0000\u0000\u0000\u00e5\u00e8\u0001\u0000\u0000\u0000\u00e6\u00e4\u0001"+
		"\u0000\u0000\u0000\u00e6\u00e7\u0001\u0000\u0000\u0000\u00e7\u0015\u0001"+
		"\u0000\u0000\u0000\u00e8\u00e6\u0001\u0000\u0000\u0000\u00e9\u00ec\u0003"+
		"\u0018\f\u0000\u00ea\u00eb\u0005\u0006\u0000\u0000\u00eb\u00ed\u0003,"+
		"\u0016\u0000\u00ec\u00ea\u0001\u0000\u0000\u0000\u00ec\u00ed\u0001\u0000"+
		"\u0000\u0000\u00ed\u0103\u0001\u0000\u0000\u0000\u00ee\u00f1\u00030\u0018"+
		"\u0000\u00ef\u00f0\u0005\u0006\u0000\u0000\u00f0\u00f2\u0003,\u0016\u0000"+
		"\u00f1\u00ef\u0001\u0000\u0000\u0000\u00f1\u00f2\u0001\u0000\u0000\u0000"+
		"\u00f2\u0103\u0001\u0000\u0000\u0000\u00f3\u00f4\u0005/\u0000\u0000\u00f4"+
		"\u00f5\u0005\u0010\u0000\u0000\u00f5\u00fa\u00030\u0018\u0000\u00f6\u00f7"+
		"\u0005\u000f\u0000\u0000\u00f7\u00f9\u00030\u0018\u0000\u00f8\u00f6\u0001"+
		"\u0000\u0000\u0000\u00f9\u00fc\u0001\u0000\u0000\u0000\u00fa\u00f8\u0001"+
		"\u0000\u0000\u0000\u00fa\u00fb\u0001\u0000\u0000\u0000\u00fb\u00fd\u0001"+
		"\u0000\u0000\u0000\u00fc\u00fa\u0001\u0000\u0000\u0000\u00fd\u0100\u0005"+
		"\u0011\u0000\u0000\u00fe\u00ff\u0005\u0006\u0000\u0000\u00ff\u0101\u0003"+
		",\u0016\u0000\u0100\u00fe\u0001\u0000\u0000\u0000\u0100\u0101\u0001\u0000"+
		"\u0000\u0000\u0101\u0103\u0001\u0000\u0000\u0000\u0102\u00e9\u0001\u0000"+
		"\u0000\u0000\u0102\u00ee\u0001\u0000\u0000\u0000\u0102\u00f3\u0001\u0000"+
		"\u0000\u0000\u0103\u0017\u0001\u0000\u0000\u0000\u0104\u0105\u0007\u0000"+
		"\u0000\u0000\u0105\u010b\u0005\u0010\u0000\u0000\u0106\u010c\u0005\u0018"+
		"\u0000\u0000\u0107\u0109\u0005\u0005\u0000\u0000\u0108\u0107\u0001\u0000"+
		"\u0000\u0000\u0108\u0109\u0001\u0000\u0000\u0000\u0109\u010a\u0001\u0000"+
		"\u0000\u0000\u010a\u010c\u0003\u001a\r\u0000\u010b\u0106\u0001\u0000\u0000"+
		"\u0000\u010b\u0108\u0001\u0000\u0000\u0000\u010c\u010d\u0001\u0000\u0000"+
		"\u0000\u010d\u010e\u0005\u0011\u0000\u0000\u010e\u0019\u0001\u0000\u0000"+
		"\u0000\u010f\u0112\u00058\u0000\u0000\u0110\u0111\u0005\t\u0000\u0000"+
		"\u0111\u0113\u00058\u0000\u0000\u0112\u0110\u0001\u0000\u0000\u0000\u0112"+
		"\u0113\u0001\u0000\u0000\u0000\u0113\u001b\u0001\u0000\u0000\u0000\u0114"+
		"\u0115\u0005\u001c\u0000\u0000\u0115\u011a\u0003\u001e\u000f\u0000\u0116"+
		"\u0117\u0005\u000f\u0000\u0000\u0117\u0119\u0003\u001e\u000f\u0000\u0118"+
		"\u0116\u0001\u0000\u0000\u0000\u0119\u011c\u0001\u0000\u0000\u0000\u011a"+
		"\u0118\u0001\u0000\u0000\u0000\u011a\u011b\u0001\u0000\u0000\u0000\u011b"+
		"\u001d\u0001\u0000\u0000\u0000\u011c\u011a\u0001\u0000\u0000\u0000\u011d"+
		"\u011e\u00030\u0018\u0000\u011e\u011f\u0007\u0001\u0000\u0000\u011f\u001f"+
		"\u0001\u0000\u0000\u0000\u0120\u0121\u0005\u001f\u0000\u0000\u0121\u0122"+
		"\u00057\u0000\u0000\u0122!\u0001\u0000\u0000\u0000\u0123\u0125\u0005\u000e"+
		"\u0000\u0000\u0124\u0126\u0003$\u0012\u0000\u0125\u0124\u0001\u0000\u0000"+
		"\u0000\u0126\u0127\u0001\u0000\u0000\u0000\u0127\u0125\u0001\u0000\u0000"+
		"\u0000\u0127\u0128\u0001\u0000\u0000\u0000\u0128#\u0001\u0000\u0000\u0000"+
		"\u0129\u012a\u00058\u0000\u0000\u012a%\u0001\u0000\u0000\u0000\u012b\u012c"+
		"\u0005\u0014\u0000\u0000\u012c\u0131\u0003(\u0014\u0000\u012d\u012e\u0005"+
		"\u000f\u0000\u0000\u012e\u0130\u0003(\u0014\u0000\u012f\u012d\u0001\u0000"+
		"\u0000\u0000\u0130\u0133\u0001\u0000\u0000\u0000\u0131\u012f\u0001\u0000"+
		"\u0000\u0000\u0131\u0132\u0001\u0000\u0000\u0000\u0132\u0134\u0001\u0000"+
		"\u0000\u0000\u0133\u0131\u0001\u0000\u0000\u0000\u0134\u0135\u0005\u0015"+
		"\u0000\u0000\u0135\'\u0001\u0000\u0000\u0000\u0136\u0137\u00058\u0000"+
		"\u0000\u0137\u0138\u0005\u000e\u0000\u0000\u0138\u0139\u00032\u0019\u0000"+
		"\u0139)\u0001\u0000\u0000\u0000\u013a\u013b\u0006\u0015\uffff\uffff\u0000"+
		"\u013b\u013c\u0005\u0010\u0000\u0000\u013c\u013d\u0003*\u0015\u0000\u013d"+
		"\u013e\u0005\u0011\u0000\u0000\u013e\u0176\u0001\u0000\u0000\u0000\u013f"+
		"\u0140\u0005\'\u0000\u0000\u0140\u0176\u0003*\u0015\u000b\u0141\u0142"+
		"\u00051\u0000\u0000\u0142\u0143\u0005\u0010\u0000\u0000\u0143\u0144\u0003"+
		",\u0016\u0000\u0144\u0145\u00050\u0000\u0000\u0145\u0146\u00030\u0018"+
		"\u0000\u0146\u0147\u0005\u0004\u0000\u0000\u0147\u0148\u0003*\u0015\u0000"+
		"\u0148\u0149\u0005\u0011\u0000\u0000\u0149\u0176\u0001\u0000\u0000\u0000"+
		"\u014a\u014b\u00052\u0000\u0000\u014b\u014c\u0005\u0010\u0000\u0000\u014c"+
		"\u014d\u0003,\u0016\u0000\u014d\u014e\u00050\u0000\u0000\u014e\u014f\u0003"+
		"0\u0018\u0000\u014f\u0150\u0005\u0004\u0000\u0000\u0150\u0151\u0003*\u0015"+
		"\u0000\u0151\u0152\u0005\u0011\u0000\u0000\u0152\u0176\u0001\u0000\u0000"+
		"\u0000\u0153\u0154\u00053\u0000\u0000\u0154\u0155\u0005\u0010\u0000\u0000"+
		"\u0155\u0156\u0003,\u0016\u0000\u0156\u0157\u00050\u0000\u0000\u0157\u0158"+
		"\u00030\u0018\u0000\u0158\u0159\u0005\u0004\u0000\u0000\u0159\u015a\u0003"+
		"*\u0015\u0000\u015a\u015b\u0005\u0011\u0000\u0000\u015b\u0176\u0001\u0000"+
		"\u0000\u0000\u015c\u015d\u00054\u0000\u0000\u015d\u015e\u0005\u0010\u0000"+
		"\u0000\u015e\u015f\u0003,\u0016\u0000\u015f\u0160\u00050\u0000\u0000\u0160"+
		"\u0161\u00030\u0018\u0000\u0161\u0162\u0005\u0004\u0000\u0000\u0162\u0163"+
		"\u0003*\u0015\u0000\u0163\u0164\u0005\u0011\u0000\u0000\u0164\u0176\u0001"+
		"\u0000\u0000\u0000\u0165\u0166\u00030\u0018\u0000\u0166\u0167\u0005$\u0000"+
		"\u0000\u0167\u0168\u00032\u0019\u0000\u0168\u0176\u0001\u0000\u0000\u0000"+
		"\u0169\u016a\u00030\u0018\u0000\u016a\u016b\u0005\b\u0000\u0000\u016b"+
		"\u016c\u00032\u0019\u0000\u016c\u0176\u0001\u0000\u0000\u0000\u016d\u016e"+
		"\u00030\u0018\u0000\u016e\u016f\u0005\r\u0000\u0000\u016f\u0170\u0003"+
		"2\u0019\u0000\u0170\u0176\u0001\u0000\u0000\u0000\u0171\u0172\u00030\u0018"+
		"\u0000\u0172\u0173\u0005\f\u0000\u0000\u0173\u0174\u00032\u0019\u0000"+
		"\u0174\u0176\u0001\u0000\u0000\u0000\u0175\u013a\u0001\u0000\u0000\u0000"+
		"\u0175\u013f\u0001\u0000\u0000\u0000\u0175\u0141\u0001\u0000\u0000\u0000"+
		"\u0175\u014a\u0001\u0000\u0000\u0000\u0175\u0153\u0001\u0000\u0000\u0000"+
		"\u0175\u015c\u0001\u0000\u0000\u0000\u0175\u0165\u0001\u0000\u0000\u0000"+
		"\u0175\u0169\u0001\u0000\u0000\u0000\u0175\u016d\u0001\u0000\u0000\u0000"+
		"\u0175\u0171\u0001\u0000\u0000\u0000\u0176\u017f\u0001\u0000\u0000\u0000"+
		"\u0177\u0178\n\n\u0000\u0000\u0178\u0179\u0005%\u0000\u0000\u0179\u017e"+
		"\u0003*\u0015\u000b\u017a\u017b\n\t\u0000\u0000\u017b\u017c\u0005&\u0000"+
		"\u0000\u017c\u017e\u0003*\u0015\n\u017d\u0177\u0001\u0000\u0000\u0000"+
		"\u017d\u017a\u0001\u0000\u0000\u0000\u017e\u0181\u0001\u0000\u0000\u0000"+
		"\u017f\u017d\u0001\u0000\u0000\u0000\u017f\u0180\u0001\u0000\u0000\u0000"+
		"\u0180+\u0001\u0000\u0000\u0000\u0181\u017f\u0001\u0000\u0000\u0000\u0182"+
		"\u0183\u00058\u0000\u0000\u0183-\u0001\u0000\u0000\u0000\u0184\u0185\u0005"+
		"\u000e\u0000\u0000\u0185\u018a\u00058\u0000\u0000\u0186\u0187\u0005\u0001"+
		"\u0000\u0000\u0187\u0189\u00058\u0000\u0000\u0188\u0186\u0001\u0000\u0000"+
		"\u0000\u0189\u018c\u0001\u0000\u0000\u0000\u018a\u0188\u0001\u0000\u0000"+
		"\u0000\u018a\u018b\u0001\u0000\u0000\u0000\u018b/\u0001\u0000\u0000\u0000"+
		"\u018c\u018a\u0001\u0000\u0000\u0000\u018d\u0192\u00058\u0000\u0000\u018e"+
		"\u018f\u0005\t\u0000\u0000\u018f\u0191\u00058\u0000\u0000\u0190\u018e"+
		"\u0001\u0000\u0000\u0000\u0191\u0194\u0001\u0000\u0000\u0000\u0192\u0190"+
		"\u0001\u0000\u0000\u0000\u0192\u0193\u0001\u0000\u0000\u0000\u0193\u019b"+
		"\u0001\u0000\u0000\u0000\u0194\u0192\u0001\u0000\u0000\u0000\u0195\u0196"+
		"\u00058\u0000\u0000\u0196\u0197\u0005\u0010\u0000\u0000\u0197\u0198\u0003"+
		",\u0016\u0000\u0198\u0199\u0005\u0011\u0000\u0000\u0199\u019b\u0001\u0000"+
		"\u0000\u0000\u019a\u018d\u0001\u0000\u0000\u0000\u019a\u0195\u0001\u0000"+
		"\u0000\u0000\u019b1\u0001\u0000\u0000\u0000\u019c\u019d\u0007\u0002\u0000"+
		"\u0000\u019d3\u0001\u0000\u0000\u0000\u019e\u01a0\u0005\u0018\u0000\u0000"+
		"\u019f\u01a1\u00036\u001b\u0000\u01a0\u019f\u0001\u0000\u0000\u0000\u01a0"+
		"\u01a1\u0001\u0000\u0000\u0000\u01a15\u0001\u0000\u0000\u0000\u01a2\u01a4"+
		"\u00057\u0000\u0000\u01a3\u01a2\u0001\u0000\u0000\u0000\u01a3\u01a4\u0001"+
		"\u0000\u0000\u0000\u01a4\u01a5\u0001\u0000\u0000\u0000\u01a5\u01a7\u0005"+
		"\u0019\u0000\u0000\u01a6\u01a8\u00057\u0000\u0000\u01a7\u01a6\u0001\u0000"+
		"\u0000\u0000\u01a7\u01a8\u0001\u0000\u0000\u0000\u01a8\u01ab\u0001\u0000"+
		"\u0000\u0000\u01a9\u01ab\u00057\u0000\u0000\u01aa\u01a3\u0001\u0000\u0000"+
		"\u0000\u01aa\u01a9\u0001\u0000\u0000\u0000\u01ab7\u0001\u0000\u0000\u0000"+
		"6;AIRWZ`it|\u0081\u0084\u0087\u008f\u0094\u0097\u009a\u009d\u00a3\u00ab"+
		"\u00b0\u00b3\u00b6\u00b9\u00bf\u00c7\u00cc\u00cf\u00d2\u00d5\u00db\u00df"+
		"\u00e6\u00ec\u00f1\u00fa\u0100\u0102\u0108\u010b\u0112\u011a\u0127\u0131"+
		"\u0175\u017d\u017f\u018a\u0192\u019a\u01a0\u01a3\u01a7\u01aa";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}