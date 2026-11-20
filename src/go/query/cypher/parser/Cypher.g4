grammar Cypher;
options { caseInsensitive = true; }

// lexer rule

MATCH: 'MATCH';
RETURN: 'RETURN';
WHERE: 'WHERE';
DISTINCT: 'DISTINCT';
AS: 'AS';
WITH: 'WITH';

NEQ: '<>';
LE: '<=';
GE: '>=';
EQ: '=';

LARROW: '<-';
RARROW: '->';
LANGLE: '<';
RANGLE: '>';
COLON: ':';
DOT: '.';
COMMA: ',';

LPAREN: '(';
RPAREN: ')';
LSQUARE: '[';
RSQUARE: ']';
LCURLY: '{';
RCURLY: '}';

PLUS: '+';
MINUS: '-';
SLASH: '/';
PERCENT: '%';
SQUOTE: '\'';
STAR: '*';
DOUBLE_DOT: '..';

ORDER_BY:'ORDER BY'; 
ASC: 'ASC';
DESC: 'DESC';
LIMIT: 'LIMIT';
OPTIONAL: 'OPTIONAL';
UNWIND: 'UNWIND';
FINISH: 'FINISH';

// operator
AND: 'AND';
OR: 'OR';
NOT: 'NOT';
XOR: 'XOR';

// function
COUNT: 'COUNT';
REDUCE: 'REDUCE';
SUM: 'SUM';
AVG: 'AVG';
MIN: 'MIN';
MAX: 'MAX';
COLLECT: 'COLLECT';
COALESCE: 'COALESCE';

IN: 'IN';
IS: 'IS';
NULL: 'NULL';
TRUE: 'TRUE';
FALSE: 'FALSE';

// predicates
ALL: 'ALL';
ANY: 'ANY';
NONE: 'NONE';
SINGLE: 'SINGLE';

// CUD TODO: implementation
CREATE: 'CREATE';
REMOVE: 'REMOVE';
SET: 'SET';
MERGE: 'MERGE';
MIGRATE: 'MIGRATE';

// additional function
CALL: 'CALL';

// CASE expressions
CASE: 'CASE';
WHEN: 'WHEN';
THEN: 'THEN';
ELSE: 'ELSE';
END: 'END';

// Define last to avoid duplication
STRING: '"' (~["\\] | '\\' .)* '"' | '\'' (~['\\] | '\\' .)* '\'';
NUMBER: [0-9]+ ('.' [0-9]+)?;
IDENTIFIER: [a-z_][a-z_0-9]*;

// Whitespace and comments
WS: [ \t\r\n]+ -> skip;

// parser rule

cypher: statement + EOF;
// A statement is one or more reading clauses (MATCH [WHERE] / WITH) followed by RETURN.
// A single-stage query (MATCH [WHERE] RETURN) is one readingClause plus a returnClause.
statement: readingClause+ returnClause;
readingClause: OPTIONAL? matchClause whereClause? | withClause | unwindClause;
unwindClause: UNWIND valueExpr AS variable;
withClause: WITH DISTINCT? returnItems orderItems? limitNum? whereClause?;
matchClause: MATCH pattern (COMMA pattern)*;
whereClause: WHERE condition;
returnClause: RETURN returnItems orderItems? limitNum?;

callClause: CALL STRING;
asClause: AS IDENTIFIER;

pattern: (variable EQ)? node (patternStep node)*;
// A step is a relationship (quantified with {m,n} / + / * as a shorthand of a one-relationship group)
// or a quantified path pattern ((x)-[r]->(y) WHERE c){m,n}.
patternStep: relationship quantifier? | quantifiedPath;
quantifiedPath: LPAREN node (relationship node)+ whereClause? RPAREN quantifier;
quantifier:
      LCURLY NUMBER RCURLY
    | LCURLY NUMBER? COMMA NUMBER? RCURLY
    | PLUS
    | STAR;
node: LPAREN variable? labels? properties? (WHERE condition)? RPAREN;
relationship:
    MINUS LSQUARE variable? types? range? properties? (WHERE condition)? RSQUARE RARROW
    | LARROW LSQUARE variable? types? range? properties? (WHERE condition)? RSQUARE MINUS
    | MINUS LSQUARE variable? types? range? properties? (WHERE condition)? RSQUARE MINUS;

returnItems: returnItem (COMMA returnItem)*;
returnItem:
      aggregateFunc (AS variable)?
    | caseExpr (AS variable)?
    | valueExpr (AS variable)?
    | COALESCE LPAREN expression (COMMA expression)* RPAREN (AS variable)?;
caseExpr:
      CASE (WHEN condition THEN caseResult)+ (ELSE caseResult)? END              #caseSearched
    | CASE valueExpr (WHEN caseKey THEN caseResult)+ (ELSE caseResult)? END    #caseSimple;
// The value of a simple CASE (compared with `=` as in Neo4j, so WHEN null never matches).
caseKey: value | NULL | TRUE | FALSE | expression;
caseResult: value | NULL | TRUE | FALSE | expression;

// valueExpr is a general RETURN item expression (arithmetic, function calls, list literals).
// A plain property reference (a.b) stays a materialized column; only compound expressions are evaluated as Expr.
valueExpr:
      MINUS valueExpr                                            #veNeg
    | valueExpr op=(STAR|SLASH|PERCENT) valueExpr                #veMulDiv
    | valueExpr op=(PLUS|MINUS) valueExpr                        #veAddSub
    | IDENTIFIER LPAREN (valueExpr (COMMA valueExpr)*)? RPAREN   #veCall
    | IDENTIFIER (DOT IDENTIFIER)*                               #vePropRef
    | LSQUARE (valueExpr (COMMA valueExpr)*)? RSQUARE            #veList
    | (STRING | NUMBER)                                          #veLit
    | LPAREN valueExpr RPAREN                                    #veParen;
aggregateFunc: (COUNT | SUM | AVG | MIN | MAX | COLLECT) LPAREN (STAR | DISTINCT? aggArg) RPAREN;
aggArg: IDENTIFIER (DOT IDENTIFIER)?;

orderItems: ORDER_BY orderItem (COMMA orderItem)*;
orderItem: expression (ASC|DESC)?;
limitNum: LIMIT NUMBER;

labels: COLON label+;
label: IDENTIFIER;
properties: LCURLY property (COMMA property)* RCURLY;
property: IDENTIFIER COLON (value | expression);

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
    | expression IN expression #ConditionInExpr
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