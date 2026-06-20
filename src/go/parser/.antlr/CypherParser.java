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
		NOT=39, XOR=40, COUNT=41, REDUCE=42, SUM=43, AVG=44, COALESCE=45, IN=46, 
		ALL=47, ANY=48, NONE=49, SINGLE=50, CALL=51, STRING=52, NUMBER=53, IDENTIFIER=54, 
		WS=55;
	public static final int
		RULE_cypher = 0, RULE_statement = 1, RULE_matchClause = 2, RULE_returnClause = 3, 
		RULE_whereClause = 4, RULE_callClause = 5, RULE_asClause = 6, RULE_pattern = 7, 
		RULE_node = 8, RULE_relationship = 9, RULE_returnItems = 10, RULE_returnItem = 11, 
		RULE_orderItems = 12, RULE_orderItem = 13, RULE_limitNum = 14, RULE_labels = 15, 
		RULE_label = 16, RULE_properties = 17, RULE_property = 18, RULE_condition = 19, 
		RULE_variable = 20, RULE_types = 21, RULE_expression = 22, RULE_value = 23, 
		RULE_range = 24, RULE_rangeLiteral = 25;
	private static String[] makeRuleNames() {
		return new String[] {
			"cypher", "statement", "matchClause", "returnClause", "whereClause", 
			"callClause", "asClause", "pattern", "node", "relationship", "returnItems", 
			"returnItem", "orderItems", "orderItem", "limitNum", "labels", "label", 
			"properties", "property", "condition", "variable", "types", "expression", 
			"value", "range", "rangeLiteral"
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
			"XOR", "COUNT", "REDUCE", "SUM", "AVG", "COALESCE", "IN", "ALL", "ANY", 
			"NONE", "SINGLE", "CALL", "STRING", "NUMBER", "IDENTIFIER", "WS"
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
			setState(53); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(52);
				statement();
				}
				}
				setState(55); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==MATCH );
			setState(57);
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
			setState(59);
			matchClause();
			setState(61);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==WHERE) {
				{
				setState(60);
				whereClause();
				}
			}

			setState(63);
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
			setState(65);
			match(MATCH);
			setState(69);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(66);
				match(WS);
				}
				}
				setState(71);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(72);
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
			setState(74);
			match(RETURN);
			setState(78);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(75);
				match(WS);
				}
				}
				setState(80);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(81);
			returnItems();
			setState(83);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==ORDER_BY) {
				{
				setState(82);
				orderItems();
				}
			}

			setState(86);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LIMIT) {
				{
				setState(85);
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
			setState(88);
			match(WHERE);
			setState(92);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(89);
				match(WS);
				}
				}
				setState(94);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(95);
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
			setState(97);
			match(CALL);
			setState(101);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==WS) {
				{
				{
				setState(98);
				match(WS);
				}
				}
				setState(103);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(104);
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
			setState(106);
			match(AS);
			setState(107);
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
			setState(112);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==IDENTIFIER) {
				{
				setState(109);
				variable();
				setState(110);
				match(EQ);
				}
			}

			setState(114);
			node();
			setState(120);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==LARROW || _la==MINUS) {
				{
				{
				setState(115);
				relationship();
				setState(116);
				node();
				}
				}
				setState(122);
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
			setState(123);
			match(LPAREN);
			setState(125);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==IDENTIFIER) {
				{
				setState(124);
				variable();
				}
			}

			setState(128);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==COLON) {
				{
				setState(127);
				labels();
				}
			}

			setState(131);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==LCURLY) {
				{
				setState(130);
				properties();
				}
			}

			setState(133);
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
			setState(219);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,31,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(135);
				match(MINUS);
				setState(139);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(136);
					match(WS);
					}
					}
					setState(141);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(142);
				match(LSQUARE);
				setState(144);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==IDENTIFIER) {
					{
					setState(143);
					variable();
					}
				}

				setState(147);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==COLON) {
					{
					setState(146);
					types();
					}
				}

				setState(150);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==STAR) {
					{
					setState(149);
					range();
					}
				}

				setState(153);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==LCURLY) {
					{
					setState(152);
					properties();
					}
				}

				setState(155);
				match(RSQUARE);
				setState(159);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(156);
					match(WS);
					}
					}
					setState(161);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(162);
				match(RARROW);
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(163);
				match(LARROW);
				setState(167);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(164);
					match(WS);
					}
					}
					setState(169);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(170);
				match(LSQUARE);
				setState(172);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==IDENTIFIER) {
					{
					setState(171);
					variable();
					}
				}

				setState(175);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==COLON) {
					{
					setState(174);
					types();
					}
				}

				setState(178);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==STAR) {
					{
					setState(177);
					range();
					}
				}

				setState(181);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==LCURLY) {
					{
					setState(180);
					properties();
					}
				}

				setState(183);
				match(RSQUARE);
				setState(187);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(184);
					match(WS);
					}
					}
					setState(189);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(190);
				match(MINUS);
				}
				break;
			case 3:
				enterOuterAlt(_localctx, 3);
				{
				setState(191);
				match(MINUS);
				setState(195);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(192);
					match(WS);
					}
					}
					setState(197);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(198);
				match(LSQUARE);
				setState(200);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==IDENTIFIER) {
					{
					setState(199);
					variable();
					}
				}

				setState(203);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==COLON) {
					{
					setState(202);
					types();
					}
				}

				setState(206);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==STAR) {
					{
					setState(205);
					range();
					}
				}

				setState(209);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==LCURLY) {
					{
					setState(208);
					properties();
					}
				}

				setState(211);
				match(RSQUARE);
				setState(215);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==WS) {
					{
					{
					setState(212);
					match(WS);
					}
					}
					setState(217);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(218);
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
			setState(221);
			returnItem();
			setState(226);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(222);
				match(COMMA);
				setState(223);
				returnItem();
				}
				}
				setState(228);
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
		public List<ExpressionContext> expression() {
			return getRuleContexts(ExpressionContext.class);
		}
		public ExpressionContext expression(int i) {
			return getRuleContext(ExpressionContext.class,i);
		}
		public TerminalNode AS() { return getToken(CypherParser.AS, 0); }
		public VariableContext variable() {
			return getRuleContext(VariableContext.class,0);
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
			setState(246);
			_errHandler.sync(this);
			switch (_input.LA(1)) {
			case IDENTIFIER:
				enterOuterAlt(_localctx, 1);
				{
				setState(229);
				expression();
				setState(232);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==AS) {
					{
					setState(230);
					match(AS);
					setState(231);
					variable();
					}
				}

				}
				break;
			case COALESCE:
				enterOuterAlt(_localctx, 2);
				{
				setState(234);
				match(COALESCE);
				setState(235);
				match(LPAREN);
				setState(236);
				expression();
				setState(241);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==COMMA) {
					{
					{
					setState(237);
					match(COMMA);
					setState(238);
					expression();
					}
					}
					setState(243);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				setState(244);
				match(RPAREN);
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
		enterRule(_localctx, 24, RULE_orderItems);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(248);
			match(ORDER_BY);
			setState(249);
			orderItem();
			setState(254);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(250);
				match(COMMA);
				setState(251);
				orderItem();
				}
				}
				setState(256);
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
		enterRule(_localctx, 26, RULE_orderItem);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(257);
			expression();
			setState(258);
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
		enterRule(_localctx, 28, RULE_limitNum);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(260);
			match(LIMIT);
			setState(261);
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
		enterRule(_localctx, 30, RULE_labels);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(263);
			match(COLON);
			setState(265); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(264);
				label();
				}
				}
				setState(267); 
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
		enterRule(_localctx, 32, RULE_label);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(269);
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
		enterRule(_localctx, 34, RULE_properties);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(271);
			match(LCURLY);
			setState(272);
			property();
			setState(277);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==COMMA) {
				{
				{
				setState(273);
				match(COMMA);
				setState(274);
				property();
				}
				}
				setState(279);
				_errHandler.sync(this);
				_la = _input.LA(1);
			}
			setState(280);
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
		enterRule(_localctx, 36, RULE_property);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(282);
			match(IDENTIFIER);
			setState(283);
			match(COLON);
			setState(284);
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
		int _startState = 38;
		enterRecursionRule(_localctx, 38, RULE_condition, _p);
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(345);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,39,_ctx) ) {
			case 1:
				{
				_localctx = new ConditionParenContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;

				setState(287);
				match(LPAREN);
				setState(288);
				condition(0);
				setState(289);
				match(RPAREN);
				}
				break;
			case 2:
				{
				_localctx = new ConditionNotContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(291);
				match(NOT);
				setState(292);
				condition(11);
				}
				break;
			case 3:
				{
				_localctx = new ConditionAllContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(293);
				match(ALL);
				setState(294);
				match(LPAREN);
				setState(295);
				variable();
				setState(296);
				match(IN);
				setState(297);
				expression();
				setState(298);
				match(WHERE);
				setState(299);
				condition(0);
				setState(300);
				match(RPAREN);
				}
				break;
			case 4:
				{
				_localctx = new ConditionAnyContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(302);
				match(ANY);
				setState(303);
				match(LPAREN);
				setState(304);
				variable();
				setState(305);
				match(IN);
				setState(306);
				expression();
				setState(307);
				match(WHERE);
				setState(308);
				condition(0);
				setState(309);
				match(RPAREN);
				}
				break;
			case 5:
				{
				_localctx = new ConditionNoneContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(311);
				match(NONE);
				setState(312);
				match(LPAREN);
				setState(313);
				variable();
				setState(314);
				match(IN);
				setState(315);
				expression();
				setState(316);
				match(WHERE);
				setState(317);
				condition(0);
				setState(318);
				match(RPAREN);
				}
				break;
			case 6:
				{
				_localctx = new ConditionSingleContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(320);
				match(SINGLE);
				setState(321);
				match(LPAREN);
				setState(322);
				variable();
				setState(323);
				match(IN);
				setState(324);
				expression();
				setState(325);
				match(WHERE);
				setState(326);
				condition(0);
				setState(327);
				match(RPAREN);
				}
				break;
			case 7:
				{
				_localctx = new ConditionEqualityContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(329);
				expression();
				setState(330);
				match(EQ);
				setState(331);
				value();
				}
				break;
			case 8:
				{
				_localctx = new ConditionNotEqualityContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(333);
				expression();
				setState(334);
				match(NEQ);
				setState(335);
				value();
				}
				break;
			case 9:
				{
				_localctx = new ConditionGreaterContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(337);
				expression();
				setState(338);
				match(RANGLE);
				setState(339);
				value();
				}
				break;
			case 10:
				{
				_localctx = new ConditionLessContext(_localctx);
				_ctx = _localctx;
				_prevctx = _localctx;
				setState(341);
				expression();
				setState(342);
				match(LANGLE);
				setState(343);
				value();
				}
				break;
			}
			_ctx.stop = _input.LT(-1);
			setState(355);
			_errHandler.sync(this);
			_alt = getInterpreter().adaptivePredict(_input,41,_ctx);
			while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER ) {
				if ( _alt==1 ) {
					if ( _parseListeners!=null ) triggerExitRuleEvent();
					_prevctx = _localctx;
					{
					setState(353);
					_errHandler.sync(this);
					switch ( getInterpreter().adaptivePredict(_input,40,_ctx) ) {
					case 1:
						{
						_localctx = new ConditionAndContext(new ConditionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_condition);
						setState(347);
						if (!(precpred(_ctx, 10))) throw new FailedPredicateException(this, "precpred(_ctx, 10)");
						setState(348);
						match(AND);
						setState(349);
						condition(11);
						}
						break;
					case 2:
						{
						_localctx = new ConditionOrContext(new ConditionContext(_parentctx, _parentState));
						pushNewRecursionContext(_localctx, _startState, RULE_condition);
						setState(350);
						if (!(precpred(_ctx, 9))) throw new FailedPredicateException(this, "precpred(_ctx, 9)");
						setState(351);
						match(OR);
						setState(352);
						condition(10);
						}
						break;
					}
					} 
				}
				setState(357);
				_errHandler.sync(this);
				_alt = getInterpreter().adaptivePredict(_input,41,_ctx);
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
		enterRule(_localctx, 40, RULE_variable);
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(358);
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
		enterRule(_localctx, 42, RULE_types);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(360);
			match(COLON);
			setState(361);
			match(IDENTIFIER);
			setState(366);
			_errHandler.sync(this);
			_la = _input.LA(1);
			while (_la==T__0) {
				{
				{
				setState(362);
				match(T__0);
				setState(363);
				match(IDENTIFIER);
				}
				}
				setState(368);
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
		enterRule(_localctx, 44, RULE_expression);
		int _la;
		try {
			setState(382);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,44,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(369);
				match(IDENTIFIER);
				setState(374);
				_errHandler.sync(this);
				_la = _input.LA(1);
				while (_la==DOT) {
					{
					{
					setState(370);
					match(DOT);
					setState(371);
					match(IDENTIFIER);
					}
					}
					setState(376);
					_errHandler.sync(this);
					_la = _input.LA(1);
				}
				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(377);
				match(IDENTIFIER);
				setState(378);
				match(LPAREN);
				setState(379);
				variable();
				setState(380);
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
		enterRule(_localctx, 46, RULE_value);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(384);
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
		enterRule(_localctx, 48, RULE_range);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(386);
			match(STAR);
			setState(388);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==DOUBLE_DOT || _la==NUMBER) {
				{
				setState(387);
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
		enterRule(_localctx, 50, RULE_rangeLiteral);
		int _la;
		try {
			setState(398);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,48,_ctx) ) {
			case 1:
				enterOuterAlt(_localctx, 1);
				{
				setState(391);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==NUMBER) {
					{
					setState(390);
					match(NUMBER);
					}
				}

				setState(393);
				match(DOUBLE_DOT);
				setState(395);
				_errHandler.sync(this);
				_la = _input.LA(1);
				if (_la==NUMBER) {
					{
					setState(394);
					match(NUMBER);
					}
				}

				}
				break;
			case 2:
				enterOuterAlt(_localctx, 2);
				{
				setState(397);
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
		case 19:
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
		"\u0004\u00017\u0191\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0002\u0004\u0007\u0004\u0002"+
		"\u0005\u0007\u0005\u0002\u0006\u0007\u0006\u0002\u0007\u0007\u0007\u0002"+
		"\b\u0007\b\u0002\t\u0007\t\u0002\n\u0007\n\u0002\u000b\u0007\u000b\u0002"+
		"\f\u0007\f\u0002\r\u0007\r\u0002\u000e\u0007\u000e\u0002\u000f\u0007\u000f"+
		"\u0002\u0010\u0007\u0010\u0002\u0011\u0007\u0011\u0002\u0012\u0007\u0012"+
		"\u0002\u0013\u0007\u0013\u0002\u0014\u0007\u0014\u0002\u0015\u0007\u0015"+
		"\u0002\u0016\u0007\u0016\u0002\u0017\u0007\u0017\u0002\u0018\u0007\u0018"+
		"\u0002\u0019\u0007\u0019\u0001\u0000\u0004\u00006\b\u0000\u000b\u0000"+
		"\f\u00007\u0001\u0000\u0001\u0000\u0001\u0001\u0001\u0001\u0003\u0001"+
		">\b\u0001\u0001\u0001\u0001\u0001\u0001\u0002\u0001\u0002\u0005\u0002"+
		"D\b\u0002\n\u0002\f\u0002G\t\u0002\u0001\u0002\u0001\u0002\u0001\u0003"+
		"\u0001\u0003\u0005\u0003M\b\u0003\n\u0003\f\u0003P\t\u0003\u0001\u0003"+
		"\u0001\u0003\u0003\u0003T\b\u0003\u0001\u0003\u0003\u0003W\b\u0003\u0001"+
		"\u0004\u0001\u0004\u0005\u0004[\b\u0004\n\u0004\f\u0004^\t\u0004\u0001"+
		"\u0004\u0001\u0004\u0001\u0005\u0001\u0005\u0005\u0005d\b\u0005\n\u0005"+
		"\f\u0005g\t\u0005\u0001\u0005\u0001\u0005\u0001\u0006\u0001\u0006\u0001"+
		"\u0006\u0001\u0007\u0001\u0007\u0001\u0007\u0003\u0007q\b\u0007\u0001"+
		"\u0007\u0001\u0007\u0001\u0007\u0001\u0007\u0005\u0007w\b\u0007\n\u0007"+
		"\f\u0007z\t\u0007\u0001\b\u0001\b\u0003\b~\b\b\u0001\b\u0003\b\u0081\b"+
		"\b\u0001\b\u0003\b\u0084\b\b\u0001\b\u0001\b\u0001\t\u0001\t\u0005\t\u008a"+
		"\b\t\n\t\f\t\u008d\t\t\u0001\t\u0001\t\u0003\t\u0091\b\t\u0001\t\u0003"+
		"\t\u0094\b\t\u0001\t\u0003\t\u0097\b\t\u0001\t\u0003\t\u009a\b\t\u0001"+
		"\t\u0001\t\u0005\t\u009e\b\t\n\t\f\t\u00a1\t\t\u0001\t\u0001\t\u0001\t"+
		"\u0005\t\u00a6\b\t\n\t\f\t\u00a9\t\t\u0001\t\u0001\t\u0003\t\u00ad\b\t"+
		"\u0001\t\u0003\t\u00b0\b\t\u0001\t\u0003\t\u00b3\b\t\u0001\t\u0003\t\u00b6"+
		"\b\t\u0001\t\u0001\t\u0005\t\u00ba\b\t\n\t\f\t\u00bd\t\t\u0001\t\u0001"+
		"\t\u0001\t\u0005\t\u00c2\b\t\n\t\f\t\u00c5\t\t\u0001\t\u0001\t\u0003\t"+
		"\u00c9\b\t\u0001\t\u0003\t\u00cc\b\t\u0001\t\u0003\t\u00cf\b\t\u0001\t"+
		"\u0003\t\u00d2\b\t\u0001\t\u0001\t\u0005\t\u00d6\b\t\n\t\f\t\u00d9\t\t"+
		"\u0001\t\u0003\t\u00dc\b\t\u0001\n\u0001\n\u0001\n\u0005\n\u00e1\b\n\n"+
		"\n\f\n\u00e4\t\n\u0001\u000b\u0001\u000b\u0001\u000b\u0003\u000b\u00e9"+
		"\b\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0001\u000b\u0005"+
		"\u000b\u00f0\b\u000b\n\u000b\f\u000b\u00f3\t\u000b\u0001\u000b\u0001\u000b"+
		"\u0003\u000b\u00f7\b\u000b\u0001\f\u0001\f\u0001\f\u0001\f\u0005\f\u00fd"+
		"\b\f\n\f\f\f\u0100\t\f\u0001\r\u0001\r\u0001\r\u0001\u000e\u0001\u000e"+
		"\u0001\u000e\u0001\u000f\u0001\u000f\u0004\u000f\u010a\b\u000f\u000b\u000f"+
		"\f\u000f\u010b\u0001\u0010\u0001\u0010\u0001\u0011\u0001\u0011\u0001\u0011"+
		"\u0001\u0011\u0005\u0011\u0114\b\u0011\n\u0011\f\u0011\u0117\t\u0011\u0001"+
		"\u0011\u0001\u0011\u0001\u0012\u0001\u0012\u0001\u0012\u0001\u0012\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0003\u0013\u015a"+
		"\b\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001\u0013\u0001"+
		"\u0013\u0005\u0013\u0162\b\u0013\n\u0013\f\u0013\u0165\t\u0013\u0001\u0014"+
		"\u0001\u0014\u0001\u0015\u0001\u0015\u0001\u0015\u0001\u0015\u0005\u0015"+
		"\u016d\b\u0015\n\u0015\f\u0015\u0170\t\u0015\u0001\u0016\u0001\u0016\u0001"+
		"\u0016\u0005\u0016\u0175\b\u0016\n\u0016\f\u0016\u0178\t\u0016\u0001\u0016"+
		"\u0001\u0016\u0001\u0016\u0001\u0016\u0001\u0016\u0003\u0016\u017f\b\u0016"+
		"\u0001\u0017\u0001\u0017\u0001\u0018\u0001\u0018\u0003\u0018\u0185\b\u0018"+
		"\u0001\u0019\u0003\u0019\u0188\b\u0019\u0001\u0019\u0001\u0019\u0003\u0019"+
		"\u018c\b\u0019\u0001\u0019\u0003\u0019\u018f\b\u0019\u0001\u0019\u0000"+
		"\u0001&\u001a\u0000\u0002\u0004\u0006\b\n\f\u000e\u0010\u0012\u0014\u0016"+
		"\u0018\u001a\u001c\u001e \"$&(*,.02\u0000\u0002\u0001\u0000\u001d\u001e"+
		"\u0001\u000045\u01b0\u00005\u0001\u0000\u0000\u0000\u0002;\u0001\u0000"+
		"\u0000\u0000\u0004A\u0001\u0000\u0000\u0000\u0006J\u0001\u0000\u0000\u0000"+
		"\bX\u0001\u0000\u0000\u0000\na\u0001\u0000\u0000\u0000\fj\u0001\u0000"+
		"\u0000\u0000\u000ep\u0001\u0000\u0000\u0000\u0010{\u0001\u0000\u0000\u0000"+
		"\u0012\u00db\u0001\u0000\u0000\u0000\u0014\u00dd\u0001\u0000\u0000\u0000"+
		"\u0016\u00f6\u0001\u0000\u0000\u0000\u0018\u00f8\u0001\u0000\u0000\u0000"+
		"\u001a\u0101\u0001\u0000\u0000\u0000\u001c\u0104\u0001\u0000\u0000\u0000"+
		"\u001e\u0107\u0001\u0000\u0000\u0000 \u010d\u0001\u0000\u0000\u0000\""+
		"\u010f\u0001\u0000\u0000\u0000$\u011a\u0001\u0000\u0000\u0000&\u0159\u0001"+
		"\u0000\u0000\u0000(\u0166\u0001\u0000\u0000\u0000*\u0168\u0001\u0000\u0000"+
		"\u0000,\u017e\u0001\u0000\u0000\u0000.\u0180\u0001\u0000\u0000\u00000"+
		"\u0182\u0001\u0000\u0000\u00002\u018e\u0001\u0000\u0000\u000046\u0003"+
		"\u0002\u0001\u000054\u0001\u0000\u0000\u000067\u0001\u0000\u0000\u0000"+
		"75\u0001\u0000\u0000\u000078\u0001\u0000\u0000\u000089\u0001\u0000\u0000"+
		"\u00009:\u0005\u0000\u0000\u0001:\u0001\u0001\u0000\u0000\u0000;=\u0003"+
		"\u0004\u0002\u0000<>\u0003\b\u0004\u0000=<\u0001\u0000\u0000\u0000=>\u0001"+
		"\u0000\u0000\u0000>?\u0001\u0000\u0000\u0000?@\u0003\u0006\u0003\u0000"+
		"@\u0003\u0001\u0000\u0000\u0000AE\u0005\u0002\u0000\u0000BD\u00057\u0000"+
		"\u0000CB\u0001\u0000\u0000\u0000DG\u0001\u0000\u0000\u0000EC\u0001\u0000"+
		"\u0000\u0000EF\u0001\u0000\u0000\u0000FH\u0001\u0000\u0000\u0000GE\u0001"+
		"\u0000\u0000\u0000HI\u0003\u000e\u0007\u0000I\u0005\u0001\u0000\u0000"+
		"\u0000JN\u0005\u0003\u0000\u0000KM\u00057\u0000\u0000LK\u0001\u0000\u0000"+
		"\u0000MP\u0001\u0000\u0000\u0000NL\u0001\u0000\u0000\u0000NO\u0001\u0000"+
		"\u0000\u0000OQ\u0001\u0000\u0000\u0000PN\u0001\u0000\u0000\u0000QS\u0003"+
		"\u0014\n\u0000RT\u0003\u0018\f\u0000SR\u0001\u0000\u0000\u0000ST\u0001"+
		"\u0000\u0000\u0000TV\u0001\u0000\u0000\u0000UW\u0003\u001c\u000e\u0000"+
		"VU\u0001\u0000\u0000\u0000VW\u0001\u0000\u0000\u0000W\u0007\u0001\u0000"+
		"\u0000\u0000X\\\u0005\u0004\u0000\u0000Y[\u00057\u0000\u0000ZY\u0001\u0000"+
		"\u0000\u0000[^\u0001\u0000\u0000\u0000\\Z\u0001\u0000\u0000\u0000\\]\u0001"+
		"\u0000\u0000\u0000]_\u0001\u0000\u0000\u0000^\\\u0001\u0000\u0000\u0000"+
		"_`\u0003&\u0013\u0000`\t\u0001\u0000\u0000\u0000ae\u00053\u0000\u0000"+
		"bd\u00057\u0000\u0000cb\u0001\u0000\u0000\u0000dg\u0001\u0000\u0000\u0000"+
		"ec\u0001\u0000\u0000\u0000ef\u0001\u0000\u0000\u0000fh\u0001\u0000\u0000"+
		"\u0000ge\u0001\u0000\u0000\u0000hi\u00054\u0000\u0000i\u000b\u0001\u0000"+
		"\u0000\u0000jk\u0005\u0006\u0000\u0000kl\u00056\u0000\u0000l\r\u0001\u0000"+
		"\u0000\u0000mn\u0003(\u0014\u0000no\u0005$\u0000\u0000oq\u0001\u0000\u0000"+
		"\u0000pm\u0001\u0000\u0000\u0000pq\u0001\u0000\u0000\u0000qr\u0001\u0000"+
		"\u0000\u0000rx\u0003\u0010\b\u0000st\u0003\u0012\t\u0000tu\u0003\u0010"+
		"\b\u0000uw\u0001\u0000\u0000\u0000vs\u0001\u0000\u0000\u0000wz\u0001\u0000"+
		"\u0000\u0000xv\u0001\u0000\u0000\u0000xy\u0001\u0000\u0000\u0000y\u000f"+
		"\u0001\u0000\u0000\u0000zx\u0001\u0000\u0000\u0000{}\u0005\u0010\u0000"+
		"\u0000|~\u0003(\u0014\u0000}|\u0001\u0000\u0000\u0000}~\u0001\u0000\u0000"+
		"\u0000~\u0080\u0001\u0000\u0000\u0000\u007f\u0081\u0003\u001e\u000f\u0000"+
		"\u0080\u007f\u0001\u0000\u0000\u0000\u0080\u0081\u0001\u0000\u0000\u0000"+
		"\u0081\u0083\u0001\u0000\u0000\u0000\u0082\u0084\u0003\"\u0011\u0000\u0083"+
		"\u0082\u0001\u0000\u0000\u0000\u0083\u0084\u0001\u0000\u0000\u0000\u0084"+
		"\u0085\u0001\u0000\u0000\u0000\u0085\u0086\u0005\u0011\u0000\u0000\u0086"+
		"\u0011\u0001\u0000\u0000\u0000\u0087\u008b\u0005\u0016\u0000\u0000\u0088"+
		"\u008a\u00057\u0000\u0000\u0089\u0088\u0001\u0000\u0000\u0000\u008a\u008d"+
		"\u0001\u0000\u0000\u0000\u008b\u0089\u0001\u0000\u0000\u0000\u008b\u008c"+
		"\u0001\u0000\u0000\u0000\u008c\u008e\u0001\u0000\u0000\u0000\u008d\u008b"+
		"\u0001\u0000\u0000\u0000\u008e\u0090\u0005\u0012\u0000\u0000\u008f\u0091"+
		"\u0003(\u0014\u0000\u0090\u008f\u0001\u0000\u0000\u0000\u0090\u0091\u0001"+
		"\u0000\u0000\u0000\u0091\u0093\u0001\u0000\u0000\u0000\u0092\u0094\u0003"+
		"*\u0015\u0000\u0093\u0092\u0001\u0000\u0000\u0000\u0093\u0094\u0001\u0000"+
		"\u0000\u0000\u0094\u0096\u0001\u0000\u0000\u0000\u0095\u0097\u00030\u0018"+
		"\u0000\u0096\u0095\u0001\u0000\u0000\u0000\u0096\u0097\u0001\u0000\u0000"+
		"\u0000\u0097\u0099\u0001\u0000\u0000\u0000\u0098\u009a\u0003\"\u0011\u0000"+
		"\u0099\u0098\u0001\u0000\u0000\u0000\u0099\u009a\u0001\u0000\u0000\u0000"+
		"\u009a\u009b\u0001\u0000\u0000\u0000\u009b\u009f\u0005\u0013\u0000\u0000"+
		"\u009c\u009e\u00057\u0000\u0000\u009d\u009c\u0001\u0000\u0000\u0000\u009e"+
		"\u00a1\u0001\u0000\u0000\u0000\u009f\u009d\u0001\u0000\u0000\u0000\u009f"+
		"\u00a0\u0001\u0000\u0000\u0000\u00a0\u00a2\u0001\u0000\u0000\u0000\u00a1"+
		"\u009f\u0001\u0000\u0000\u0000\u00a2\u00dc\u0005\u000b\u0000\u0000\u00a3"+
		"\u00a7\u0005\n\u0000\u0000\u00a4\u00a6\u00057\u0000\u0000\u00a5\u00a4"+
		"\u0001\u0000\u0000\u0000\u00a6\u00a9\u0001\u0000\u0000\u0000\u00a7\u00a5"+
		"\u0001\u0000\u0000\u0000\u00a7\u00a8\u0001\u0000\u0000\u0000\u00a8\u00aa"+
		"\u0001\u0000\u0000\u0000\u00a9\u00a7\u0001\u0000\u0000\u0000\u00aa\u00ac"+
		"\u0005\u0012\u0000\u0000\u00ab\u00ad\u0003(\u0014\u0000\u00ac\u00ab\u0001"+
		"\u0000\u0000\u0000\u00ac\u00ad\u0001\u0000\u0000\u0000\u00ad\u00af\u0001"+
		"\u0000\u0000\u0000\u00ae\u00b0\u0003*\u0015\u0000\u00af\u00ae\u0001\u0000"+
		"\u0000\u0000\u00af\u00b0\u0001\u0000\u0000\u0000\u00b0\u00b2\u0001\u0000"+
		"\u0000\u0000\u00b1\u00b3\u00030\u0018\u0000\u00b2\u00b1\u0001\u0000\u0000"+
		"\u0000\u00b2\u00b3\u0001\u0000\u0000\u0000\u00b3\u00b5\u0001\u0000\u0000"+
		"\u0000\u00b4\u00b6\u0003\"\u0011\u0000\u00b5\u00b4\u0001\u0000\u0000\u0000"+
		"\u00b5\u00b6\u0001\u0000\u0000\u0000\u00b6\u00b7\u0001\u0000\u0000\u0000"+
		"\u00b7\u00bb\u0005\u0013\u0000\u0000\u00b8\u00ba\u00057\u0000\u0000\u00b9"+
		"\u00b8\u0001\u0000\u0000\u0000\u00ba\u00bd\u0001\u0000\u0000\u0000\u00bb"+
		"\u00b9\u0001\u0000\u0000\u0000\u00bb\u00bc\u0001\u0000\u0000\u0000\u00bc"+
		"\u00be\u0001\u0000\u0000\u0000\u00bd\u00bb\u0001\u0000\u0000\u0000\u00be"+
		"\u00dc\u0005\u0016\u0000\u0000\u00bf\u00c3\u0005\u0016\u0000\u0000\u00c0"+
		"\u00c2\u00057\u0000\u0000\u00c1\u00c0\u0001\u0000\u0000\u0000\u00c2\u00c5"+
		"\u0001\u0000\u0000\u0000\u00c3\u00c1\u0001\u0000\u0000\u0000\u00c3\u00c4"+
		"\u0001\u0000\u0000\u0000\u00c4\u00c6\u0001\u0000\u0000\u0000\u00c5\u00c3"+
		"\u0001\u0000\u0000\u0000\u00c6\u00c8\u0005\u0012\u0000\u0000\u00c7\u00c9"+
		"\u0003(\u0014\u0000\u00c8\u00c7\u0001\u0000\u0000\u0000\u00c8\u00c9\u0001"+
		"\u0000\u0000\u0000\u00c9\u00cb\u0001\u0000\u0000\u0000\u00ca\u00cc\u0003"+
		"*\u0015\u0000\u00cb\u00ca\u0001\u0000\u0000\u0000\u00cb\u00cc\u0001\u0000"+
		"\u0000\u0000\u00cc\u00ce\u0001\u0000\u0000\u0000\u00cd\u00cf\u00030\u0018"+
		"\u0000\u00ce\u00cd\u0001\u0000\u0000\u0000\u00ce\u00cf\u0001\u0000\u0000"+
		"\u0000\u00cf\u00d1\u0001\u0000\u0000\u0000\u00d0\u00d2\u0003\"\u0011\u0000"+
		"\u00d1\u00d0\u0001\u0000\u0000\u0000\u00d1\u00d2\u0001\u0000\u0000\u0000"+
		"\u00d2\u00d3\u0001\u0000\u0000\u0000\u00d3\u00d7\u0005\u0013\u0000\u0000"+
		"\u00d4\u00d6\u00057\u0000\u0000\u00d5\u00d4\u0001\u0000\u0000\u0000\u00d6"+
		"\u00d9\u0001\u0000\u0000\u0000\u00d7\u00d5\u0001\u0000\u0000\u0000\u00d7"+
		"\u00d8\u0001\u0000\u0000\u0000\u00d8\u00da\u0001\u0000\u0000\u0000\u00d9"+
		"\u00d7\u0001\u0000\u0000\u0000\u00da\u00dc\u0005\u0016\u0000\u0000\u00db"+
		"\u0087\u0001\u0000\u0000\u0000\u00db\u00a3\u0001\u0000\u0000\u0000\u00db"+
		"\u00bf\u0001\u0000\u0000\u0000\u00dc\u0013\u0001\u0000\u0000\u0000\u00dd"+
		"\u00e2\u0003\u0016\u000b\u0000\u00de\u00df\u0005\u000f\u0000\u0000\u00df"+
		"\u00e1\u0003\u0016\u000b\u0000\u00e0\u00de\u0001\u0000\u0000\u0000\u00e1"+
		"\u00e4\u0001\u0000\u0000\u0000\u00e2\u00e0\u0001\u0000\u0000\u0000\u00e2"+
		"\u00e3\u0001\u0000\u0000\u0000\u00e3\u0015\u0001\u0000\u0000\u0000\u00e4"+
		"\u00e2\u0001\u0000\u0000\u0000\u00e5\u00e8\u0003,\u0016\u0000\u00e6\u00e7"+
		"\u0005\u0006\u0000\u0000\u00e7\u00e9\u0003(\u0014\u0000\u00e8\u00e6\u0001"+
		"\u0000\u0000\u0000\u00e8\u00e9\u0001\u0000\u0000\u0000\u00e9\u00f7\u0001"+
		"\u0000\u0000\u0000\u00ea\u00eb\u0005-\u0000\u0000\u00eb\u00ec\u0005\u0010"+
		"\u0000\u0000\u00ec\u00f1\u0003,\u0016\u0000\u00ed\u00ee\u0005\u000f\u0000"+
		"\u0000\u00ee\u00f0\u0003,\u0016\u0000\u00ef\u00ed\u0001\u0000\u0000\u0000"+
		"\u00f0\u00f3\u0001\u0000\u0000\u0000\u00f1\u00ef\u0001\u0000\u0000\u0000"+
		"\u00f1\u00f2\u0001\u0000\u0000\u0000\u00f2\u00f4\u0001\u0000\u0000\u0000"+
		"\u00f3\u00f1\u0001\u0000\u0000\u0000\u00f4\u00f5\u0005\u0011\u0000\u0000"+
		"\u00f5\u00f7\u0001\u0000\u0000\u0000\u00f6\u00e5\u0001\u0000\u0000\u0000"+
		"\u00f6\u00ea\u0001\u0000\u0000\u0000\u00f7\u0017\u0001\u0000\u0000\u0000"+
		"\u00f8\u00f9\u0005\u001c\u0000\u0000\u00f9\u00fe\u0003\u001a\r\u0000\u00fa"+
		"\u00fb\u0005\u000f\u0000\u0000\u00fb\u00fd\u0003\u001a\r\u0000\u00fc\u00fa"+
		"\u0001\u0000\u0000\u0000\u00fd\u0100\u0001\u0000\u0000\u0000\u00fe\u00fc"+
		"\u0001\u0000\u0000\u0000\u00fe\u00ff\u0001\u0000\u0000\u0000\u00ff\u0019"+
		"\u0001\u0000\u0000\u0000\u0100\u00fe\u0001\u0000\u0000\u0000\u0101\u0102"+
		"\u0003,\u0016\u0000\u0102\u0103\u0007\u0000\u0000\u0000\u0103\u001b\u0001"+
		"\u0000\u0000\u0000\u0104\u0105\u0005\u001f\u0000\u0000\u0105\u0106\u0005"+
		"5\u0000\u0000\u0106\u001d\u0001\u0000\u0000\u0000\u0107\u0109\u0005\u000e"+
		"\u0000\u0000\u0108\u010a\u0003 \u0010\u0000\u0109\u0108\u0001\u0000\u0000"+
		"\u0000\u010a\u010b\u0001\u0000\u0000\u0000\u010b\u0109\u0001\u0000\u0000"+
		"\u0000\u010b\u010c\u0001\u0000\u0000\u0000\u010c\u001f\u0001\u0000\u0000"+
		"\u0000\u010d\u010e\u00056\u0000\u0000\u010e!\u0001\u0000\u0000\u0000\u010f"+
		"\u0110\u0005\u0014\u0000\u0000\u0110\u0115\u0003$\u0012\u0000\u0111\u0112"+
		"\u0005\u000f\u0000\u0000\u0112\u0114\u0003$\u0012\u0000\u0113\u0111\u0001"+
		"\u0000\u0000\u0000\u0114\u0117\u0001\u0000\u0000\u0000\u0115\u0113\u0001"+
		"\u0000\u0000\u0000\u0115\u0116\u0001\u0000\u0000\u0000\u0116\u0118\u0001"+
		"\u0000\u0000\u0000\u0117\u0115\u0001\u0000\u0000\u0000\u0118\u0119\u0005"+
		"\u0015\u0000\u0000\u0119#\u0001\u0000\u0000\u0000\u011a\u011b\u00056\u0000"+
		"\u0000\u011b\u011c\u0005\u000e\u0000\u0000\u011c\u011d\u0003.\u0017\u0000"+
		"\u011d%\u0001\u0000\u0000\u0000\u011e\u011f\u0006\u0013\uffff\uffff\u0000"+
		"\u011f\u0120\u0005\u0010\u0000\u0000\u0120\u0121\u0003&\u0013\u0000\u0121"+
		"\u0122\u0005\u0011\u0000\u0000\u0122\u015a\u0001\u0000\u0000\u0000\u0123"+
		"\u0124\u0005\'\u0000\u0000\u0124\u015a\u0003&\u0013\u000b\u0125\u0126"+
		"\u0005/\u0000\u0000\u0126\u0127\u0005\u0010\u0000\u0000\u0127\u0128\u0003"+
		"(\u0014\u0000\u0128\u0129\u0005.\u0000\u0000\u0129\u012a\u0003,\u0016"+
		"\u0000\u012a\u012b\u0005\u0004\u0000\u0000\u012b\u012c\u0003&\u0013\u0000"+
		"\u012c\u012d\u0005\u0011\u0000\u0000\u012d\u015a\u0001\u0000\u0000\u0000"+
		"\u012e\u012f\u00050\u0000\u0000\u012f\u0130\u0005\u0010\u0000\u0000\u0130"+
		"\u0131\u0003(\u0014\u0000\u0131\u0132\u0005.\u0000\u0000\u0132\u0133\u0003"+
		",\u0016\u0000\u0133\u0134\u0005\u0004\u0000\u0000\u0134\u0135\u0003&\u0013"+
		"\u0000\u0135\u0136\u0005\u0011\u0000\u0000\u0136\u015a\u0001\u0000\u0000"+
		"\u0000\u0137\u0138\u00051\u0000\u0000\u0138\u0139\u0005\u0010\u0000\u0000"+
		"\u0139\u013a\u0003(\u0014\u0000\u013a\u013b\u0005.\u0000\u0000\u013b\u013c"+
		"\u0003,\u0016\u0000\u013c\u013d\u0005\u0004\u0000\u0000\u013d\u013e\u0003"+
		"&\u0013\u0000\u013e\u013f\u0005\u0011\u0000\u0000\u013f\u015a\u0001\u0000"+
		"\u0000\u0000\u0140\u0141\u00052\u0000\u0000\u0141\u0142\u0005\u0010\u0000"+
		"\u0000\u0142\u0143\u0003(\u0014\u0000\u0143\u0144\u0005.\u0000\u0000\u0144"+
		"\u0145\u0003,\u0016\u0000\u0145\u0146\u0005\u0004\u0000\u0000\u0146\u0147"+
		"\u0003&\u0013\u0000\u0147\u0148\u0005\u0011\u0000\u0000\u0148\u015a\u0001"+
		"\u0000\u0000\u0000\u0149\u014a\u0003,\u0016\u0000\u014a\u014b\u0005$\u0000"+
		"\u0000\u014b\u014c\u0003.\u0017\u0000\u014c\u015a\u0001\u0000\u0000\u0000"+
		"\u014d\u014e\u0003,\u0016\u0000\u014e\u014f\u0005\b\u0000\u0000\u014f"+
		"\u0150\u0003.\u0017\u0000\u0150\u015a\u0001\u0000\u0000\u0000\u0151\u0152"+
		"\u0003,\u0016\u0000\u0152\u0153\u0005\r\u0000\u0000\u0153\u0154\u0003"+
		".\u0017\u0000\u0154\u015a\u0001\u0000\u0000\u0000\u0155\u0156\u0003,\u0016"+
		"\u0000\u0156\u0157\u0005\f\u0000\u0000\u0157\u0158\u0003.\u0017\u0000"+
		"\u0158\u015a\u0001\u0000\u0000\u0000\u0159\u011e\u0001\u0000\u0000\u0000"+
		"\u0159\u0123\u0001\u0000\u0000\u0000\u0159\u0125\u0001\u0000\u0000\u0000"+
		"\u0159\u012e\u0001\u0000\u0000\u0000\u0159\u0137\u0001\u0000\u0000\u0000"+
		"\u0159\u0140\u0001\u0000\u0000\u0000\u0159\u0149\u0001\u0000\u0000\u0000"+
		"\u0159\u014d\u0001\u0000\u0000\u0000\u0159\u0151\u0001\u0000\u0000\u0000"+
		"\u0159\u0155\u0001\u0000\u0000\u0000\u015a\u0163\u0001\u0000\u0000\u0000"+
		"\u015b\u015c\n\n\u0000\u0000\u015c\u015d\u0005%\u0000\u0000\u015d\u0162"+
		"\u0003&\u0013\u000b\u015e\u015f\n\t\u0000\u0000\u015f\u0160\u0005&\u0000"+
		"\u0000\u0160\u0162\u0003&\u0013\n\u0161\u015b\u0001\u0000\u0000\u0000"+
		"\u0161\u015e\u0001\u0000\u0000\u0000\u0162\u0165\u0001\u0000\u0000\u0000"+
		"\u0163\u0161\u0001\u0000\u0000\u0000\u0163\u0164\u0001\u0000\u0000\u0000"+
		"\u0164\'\u0001\u0000\u0000\u0000\u0165\u0163\u0001\u0000\u0000\u0000\u0166"+
		"\u0167\u00056\u0000\u0000\u0167)\u0001\u0000\u0000\u0000\u0168\u0169\u0005"+
		"\u000e\u0000\u0000\u0169\u016e\u00056\u0000\u0000\u016a\u016b\u0005\u0001"+
		"\u0000\u0000\u016b\u016d\u00056\u0000\u0000\u016c\u016a\u0001\u0000\u0000"+
		"\u0000\u016d\u0170\u0001\u0000\u0000\u0000\u016e\u016c\u0001\u0000\u0000"+
		"\u0000\u016e\u016f\u0001\u0000\u0000\u0000\u016f+\u0001\u0000\u0000\u0000"+
		"\u0170\u016e\u0001\u0000\u0000\u0000\u0171\u0176\u00056\u0000\u0000\u0172"+
		"\u0173\u0005\t\u0000\u0000\u0173\u0175\u00056\u0000\u0000\u0174\u0172"+
		"\u0001\u0000\u0000\u0000\u0175\u0178\u0001\u0000\u0000\u0000\u0176\u0174"+
		"\u0001\u0000\u0000\u0000\u0176\u0177\u0001\u0000\u0000\u0000\u0177\u017f"+
		"\u0001\u0000\u0000\u0000\u0178\u0176\u0001\u0000\u0000\u0000\u0179\u017a"+
		"\u00056\u0000\u0000\u017a\u017b\u0005\u0010\u0000\u0000\u017b\u017c\u0003"+
		"(\u0014\u0000\u017c\u017d\u0005\u0011\u0000\u0000\u017d\u017f\u0001\u0000"+
		"\u0000\u0000\u017e\u0171\u0001\u0000\u0000\u0000\u017e\u0179\u0001\u0000"+
		"\u0000\u0000\u017f-\u0001\u0000\u0000\u0000\u0180\u0181\u0007\u0001\u0000"+
		"\u0000\u0181/\u0001\u0000\u0000\u0000\u0182\u0184\u0005\u0018\u0000\u0000"+
		"\u0183\u0185\u00032\u0019\u0000\u0184\u0183\u0001\u0000\u0000\u0000\u0184"+
		"\u0185\u0001\u0000\u0000\u0000\u01851\u0001\u0000\u0000\u0000\u0186\u0188"+
		"\u00055\u0000\u0000\u0187\u0186\u0001\u0000\u0000\u0000\u0187\u0188\u0001"+
		"\u0000\u0000\u0000\u0188\u0189\u0001\u0000\u0000\u0000\u0189\u018b\u0005"+
		"\u0019\u0000\u0000\u018a\u018c\u00055\u0000\u0000\u018b\u018a\u0001\u0000"+
		"\u0000\u0000\u018b\u018c\u0001\u0000\u0000\u0000\u018c\u018f\u0001\u0000"+
		"\u0000\u0000\u018d\u018f\u00055\u0000\u0000\u018e\u0187\u0001\u0000\u0000"+
		"\u0000\u018e\u018d\u0001\u0000\u0000\u0000\u018f3\u0001\u0000\u0000\u0000"+
		"17=ENSV\\epx}\u0080\u0083\u008b\u0090\u0093\u0096\u0099\u009f\u00a7\u00ac"+
		"\u00af\u00b2\u00b5\u00bb\u00c3\u00c8\u00cb\u00ce\u00d1\u00d7\u00db\u00e2"+
		"\u00e8\u00f1\u00f6\u00fe\u010b\u0115\u0159\u0161\u0163\u016e\u0176\u017e"+
		"\u0184\u0187\u018b\u018e";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}