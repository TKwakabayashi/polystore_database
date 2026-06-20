grammar Cypher;

// lexer rule

MATCH: [mM][aA][tT][cC][hH];
RETURN: [rR][eE][tT][uU][rR][nN];
WHERE: [wW][hH][eE][rR][eE];
DISTINCT: [dD][iI][sS][tT][iI][nN][cC][tT];
AS: [aA][sS];
WITH: 'WITH';

NEQ: '<>';

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
COALESCE: [cC][oO][aA][lL][eE][sS][cC][eE];

IN: [iI][nN];

// predicates
ALL: [aA][lL][lL];
ANY: [aA][nN][yY];
NONE: [nN][oO][nN][eE];
SINGLE: [sS][iI][nN][gG][lL][eE];

// additional function
CALL: [cC][aA][lL][lL];

// Define last to avoid duplication
STRING: '"' (~["\\] | '\\' .)* '"' | '\'' (~['\\] | '\\' .)* '\'';
NUMBER: [0-9]+ ('.' [0-9]+)?;
IDENTIFIER: [a-zA-Z_][a-zA-Z_0-9]*;

// Whitespace and comments
WS: [ \t\r\n]+ -> skip;

// parser rule

cypher: statement+ EOF;
statement: matchClause whereClause? returnClause;
matchClause: MATCH WS* pattern;
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
returnItem: expression (AS variable)? | COALESCE LPAREN expression (COMMA expression)* RPAREN;

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
    | expression EQ value # ConditionEquality 
    | expression NEQ value # ConditionNotEquality 
    | expression RANGLE value # ConditionGreater 
    | expression LANGLE value # ConditionLess;


variable: IDENTIFIER;
types: COLON IDENTIFIER ( '|' IDENTIFIER)*;
expression: IDENTIFIER (DOT IDENTIFIER)*
          | IDENTIFIER LPAREN variable RPAREN;
value: STRING | NUMBER;

// range option for multihop
range:
    STAR rangeLiteral?; 

rangeLiteral: NUMBER? DOUBLE_DOT NUMBER? | NUMBER;