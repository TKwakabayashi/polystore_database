grammar Cypher;

// lexer rule

MATCH: [mM][aA][tT][cC][hH];
RETURN: [rR][eE][tT][uU][rR][nN];
WHERE: [wW][hH][eE][rR][eE];
DISTINCT: [dD][iI][sS][tT][iI][nN][cC][tT];
AS: [aA][sS];
WITH: 'WITH';

NEQ: '<>';
LE: '<=';
GE: '>=';

DOT: '.';
LARROW: '<-';
RARROW: '->';
LANGLE: '<';
RANGLE: '>';
COLON: ':';
COMMA: ',';
LPAREN: '(';
RPAREN: ')';
LSQUARE: '[';
RSQUARE: ']';
LCURLY: '{';
RCURLY: '}';
MINUS: '-';
PLUS: '+';
SLASH: '/';
PERCENT: '%';
SQUOTE: '\'';
STAR: '*';
DOUBLE_DOT: '..';

CREATE: 'CREATE';
DELETE: 'DELETE';
ORDER_BY:'ORDER BY'; 
ASC: 'ASC';
DESC: 'DESC';
LIMIT: 'LIMIT';
OPTIONAL: 'OPTIONAL';
UNWIND: 'UNWIND';
FINISH: 'FINISH';
SET: 'SET';

// operator
EQ: '=';
AND: 'AND';
OR: 'OR';
NOT: 'NOT';
XOR: 'XOR';

// function
COUNT: [cC][oO][uU][nN][tT];
REDUCE: [rR][eE][dD][uU][cC][eE];
SUM: [sS][uU][mM];
AVG: [aA][vV][gG];
MIN: [mM][iI][nN];
MAX: [mM][aA][xX];
COALESCE: [cC][oO][aA][lL][eE][sS][cC][eE];

IN: [iI][nN];
IS: [iI][sS];
NULL: [nN][uU][lL][lL];

// predicates
ALL: [aA][lL][lL];
ANY: [aA][nN][yY];
NONE: [nN][oO][nN][eE];
SINGLE: [sS][iI][nN][gG][lL][eE];

// additional function
CALL: [cC][aA][lL][lL];

// CASE 式
CASE: [cC][aA][sS][eE];
WHEN: [wW][hH][eE][nN];
THEN: [tT][hH][eE][nN];
ELSE: [eE][lL][sS][eE];
END: [eE][nN][dD];

// Define last to avoid duplication
STRING: '"' (~["\\] | '\\' .)* '"' | '\'' (~['\\] | '\\' .)* '\'';
NUMBER: [0-9]+ ('.' [0-9]+)?;
IDENTIFIER: [a-zA-Z_][a-zA-Z_0-9]*;

// Whitespace and comments
WS: [ \t\r\n]+ -> skip;

// parser rule

cypher: statement+ EOF;
// statement は 1 つ以上の reading 節（MATCH[WHERE] / WITH）に続く RETURN。
// 単段クエリ（MATCH [WHERE] RETURN）は readingClause 1 個 + returnClause として後方互換。
statement: readingClause+ returnClause;
readingClause: matchClause whereClause? | withClause;
withClause: WITH DISTINCT? returnItems orderItems? limitNum? whereClause?;
matchClause: MATCH WS* pattern (COMMA WS* pattern)*;
returnClause: RETURN WS* returnItems orderItems? limitNum?;
whereClause: WHERE WS* condition;

callClause: CALL WS* STRING;
asClause: AS IDENTIFIER;

pattern: (variable EQ)? node (relationship node)*;
node: LPAREN variable? labels? properties? RPAREN;
relationship:
    MINUS WS* LSQUARE variable? types? range? properties? RSQUARE WS* RARROW
    | LARROW WS* LSQUARE variable? types? range? properties? RSQUARE WS* MINUS
    | MINUS WS* LSQUARE variable? types? range? properties? RSQUARE WS* MINUS;

returnItems: returnItem (COMMA returnItem)*;
returnItem:
      aggregateFunc (AS variable)?
    | caseExpr (AS variable)?
    | valueExpr (AS variable)?
    | COALESCE LPAREN expression (COMMA expression)* RPAREN (AS variable)?;
caseExpr: CASE (WHEN condition THEN caseResult)+ (ELSE caseResult)? END;
caseResult: value | expression;

// valueExpr は RETURN 項目の汎用計算式（算術・関数呼び・list リテラル）。
// 単純なプロパティ参照（a.b）は従来どおり材料化列として扱い、複合式のみ Expr として評価する。
valueExpr:
      MINUS valueExpr                                            #veNeg
    | valueExpr op=(STAR|SLASH|PERCENT) valueExpr                #veMulDiv
    | valueExpr op=(PLUS|MINUS) valueExpr                        #veAddSub
    | IDENTIFIER LPAREN (valueExpr (COMMA valueExpr)*)? RPAREN   #veCall
    | IDENTIFIER (DOT IDENTIFIER)*                               #vePropRef
    | LSQUARE (valueExpr (COMMA valueExpr)*)? RSQUARE            #veList
    | (STRING | NUMBER)                                          #veLit
    | LPAREN valueExpr RPAREN                                    #veParen;
aggregateFunc: (COUNT | SUM | AVG | MIN | MAX) LPAREN (STAR | DISTINCT? aggArg) RPAREN;
aggArg: IDENTIFIER (DOT IDENTIFIER)?;

orderItems: ORDER_BY orderItem (COMMA orderItem)*;
orderItem: expression (ASC|DESC);
limitNum: LIMIT NUMBER;

labels: COLON label+;
label: IDENTIFIER;
properties: LCURLY property (COMMA property)* RCURLY;
property: IDENTIFIER COLON value;

condition:
    LPAREN condition RPAREN #conditionParen
    | NOT condition #ConditionNot
    | condition AND condition #ConditionAnd
    | condition OR condition #ConditionOr
    | ALL LPAREN variable IN expression WHERE condition RPAREN    #ConditionAll
    | ANY LPAREN variable IN expression WHERE condition RPAREN    #ConditionAny
    | NONE LPAREN variable IN expression WHERE condition RPAREN   #ConditionNone
    | SINGLE LPAREN variable IN expression WHERE condition RPAREN #ConditionSingle
    | expression IS NULL #ConditionIsNull
    | expression IS NOT NULL #ConditionIsNotNull
    | expression IN list #ConditionIn
    | expression EQ value # ConditionEquality
    | expression NEQ value # ConditionNotEquality
    | expression RANGLE value # ConditionGreater
    | expression LANGLE value # ConditionLess
    | expression GE value # ConditionGreaterEqual
    | expression LE value # ConditionLessEqual
    | expression EQ expression # ConditionVarEq
    | expression NEQ expression # ConditionVarNeq
    | expression RANGLE expression # ConditionVarGreater
    | expression LANGLE expression # ConditionVarLess
    | expression GE expression # ConditionVarGreaterEqual
    | expression LE expression # ConditionVarLessEqual;

list: LSQUARE value (COMMA value)* RSQUARE;


variable: IDENTIFIER;
types: COLON IDENTIFIER ( '|' IDENTIFIER)*;
expression: IDENTIFIER (DOT IDENTIFIER)*
          | IDENTIFIER LPAREN variable RPAREN;
value: STRING | NUMBER;

// range option for multihop
range:
    STAR rangeLiteral?; 

rangeLiteral: NUMBER? DOUBLE_DOT NUMBER? | NUMBER;