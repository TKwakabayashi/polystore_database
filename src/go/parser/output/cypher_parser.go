// Code generated from Cypher.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Cypher
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type CypherParser struct {
	*antlr.BaseParser
}

var CypherParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func cypherParserInit() {
	staticData := &CypherParserStaticData
	staticData.LiteralNames = []string{
		"", "'|'", "", "", "", "", "", "'WITH'", "'<>'", "'<='", "'>='", "'.'",
		"'<-'", "'->'", "'<'", "'>'", "':'", "','", "'('", "')'", "'['", "']'",
		"'{'", "'}'", "'-'", "'+'", "'/'", "'%'", "'''", "'*'", "'..'", "'CREATE'",
		"'DELETE'", "'ORDER BY'", "'ASC'", "'DESC'", "'LIMIT'", "'OPTIONAL'",
		"'UNWIND'", "'FINISH'", "'SET'", "'='", "'AND'", "'OR'", "'NOT'", "'XOR'",
	}
	staticData.SymbolicNames = []string{
		"", "", "MATCH", "RETURN", "WHERE", "DISTINCT", "AS", "WITH", "NEQ",
		"LE", "GE", "DOT", "LARROW", "RARROW", "LANGLE", "RANGLE", "COLON",
		"COMMA", "LPAREN", "RPAREN", "LSQUARE", "RSQUARE", "LCURLY", "RCURLY",
		"MINUS", "PLUS", "SLASH", "PERCENT", "SQUOTE", "STAR", "DOUBLE_DOT",
		"CREATE", "DELETE", "ORDER_BY", "ASC", "DESC", "LIMIT", "OPTIONAL",
		"UNWIND", "FINISH", "SET", "EQ", "AND", "OR", "NOT", "XOR", "COUNT",
		"REDUCE", "SUM", "AVG", "MIN", "MAX", "COALESCE", "IN", "IS", "NULL",
		"ALL", "ANY", "NONE", "SINGLE", "CALL", "CASE", "WHEN", "THEN", "ELSE",
		"END", "STRING", "NUMBER", "IDENTIFIER", "WS",
	}
	staticData.RuleNames = []string{
		"cypher", "statement", "readingClause", "withClause", "matchClause",
		"returnClause", "whereClause", "callClause", "asClause", "pattern",
		"node", "relationship", "returnItems", "returnItem", "caseExpr", "caseResult",
		"valueExpr", "aggregateFunc", "aggArg", "orderItems", "orderItem", "limitNum",
		"labels", "label", "properties", "property", "condition", "list", "variable",
		"types", "expression", "value", "range", "rangeLiteral",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 69, 610, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 2, 33, 7, 33, 1, 0, 4, 0, 70, 8, 0, 11, 0, 12, 0, 71,
		1, 0, 1, 0, 1, 1, 4, 1, 77, 8, 1, 11, 1, 12, 1, 78, 1, 1, 1, 1, 1, 2, 1,
		2, 3, 2, 85, 8, 2, 1, 2, 3, 2, 88, 8, 2, 1, 3, 1, 3, 3, 3, 92, 8, 3, 1,
		3, 1, 3, 3, 3, 96, 8, 3, 1, 3, 3, 3, 99, 8, 3, 1, 3, 3, 3, 102, 8, 3, 1,
		4, 1, 4, 5, 4, 106, 8, 4, 10, 4, 12, 4, 109, 9, 4, 1, 4, 1, 4, 1, 4, 5,
		4, 114, 8, 4, 10, 4, 12, 4, 117, 9, 4, 1, 4, 5, 4, 120, 8, 4, 10, 4, 12,
		4, 123, 9, 4, 1, 5, 1, 5, 5, 5, 127, 8, 5, 10, 5, 12, 5, 130, 9, 5, 1,
		5, 1, 5, 3, 5, 134, 8, 5, 1, 5, 3, 5, 137, 8, 5, 1, 6, 1, 6, 5, 6, 141,
		8, 6, 10, 6, 12, 6, 144, 9, 6, 1, 6, 1, 6, 1, 7, 1, 7, 5, 7, 150, 8, 7,
		10, 7, 12, 7, 153, 9, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9,
		3, 9, 163, 8, 9, 1, 9, 1, 9, 1, 9, 1, 9, 5, 9, 169, 8, 9, 10, 9, 12, 9,
		172, 9, 9, 1, 10, 1, 10, 3, 10, 176, 8, 10, 1, 10, 3, 10, 179, 8, 10, 1,
		10, 3, 10, 182, 8, 10, 1, 10, 1, 10, 1, 11, 1, 11, 5, 11, 188, 8, 11, 10,
		11, 12, 11, 191, 9, 11, 1, 11, 1, 11, 3, 11, 195, 8, 11, 1, 11, 3, 11,
		198, 8, 11, 1, 11, 3, 11, 201, 8, 11, 1, 11, 3, 11, 204, 8, 11, 1, 11,
		1, 11, 5, 11, 208, 8, 11, 10, 11, 12, 11, 211, 9, 11, 1, 11, 1, 11, 1,
		11, 5, 11, 216, 8, 11, 10, 11, 12, 11, 219, 9, 11, 1, 11, 1, 11, 3, 11,
		223, 8, 11, 1, 11, 3, 11, 226, 8, 11, 1, 11, 3, 11, 229, 8, 11, 1, 11,
		3, 11, 232, 8, 11, 1, 11, 1, 11, 5, 11, 236, 8, 11, 10, 11, 12, 11, 239,
		9, 11, 1, 11, 1, 11, 1, 11, 5, 11, 244, 8, 11, 10, 11, 12, 11, 247, 9,
		11, 1, 11, 1, 11, 3, 11, 251, 8, 11, 1, 11, 3, 11, 254, 8, 11, 1, 11, 3,
		11, 257, 8, 11, 1, 11, 3, 11, 260, 8, 11, 1, 11, 1, 11, 5, 11, 264, 8,
		11, 10, 11, 12, 11, 267, 9, 11, 1, 11, 3, 11, 270, 8, 11, 1, 12, 1, 12,
		1, 12, 5, 12, 275, 8, 12, 10, 12, 12, 12, 278, 9, 12, 1, 13, 1, 13, 1,
		13, 3, 13, 283, 8, 13, 1, 13, 1, 13, 1, 13, 3, 13, 288, 8, 13, 1, 13, 1,
		13, 1, 13, 3, 13, 293, 8, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 5, 13,
		300, 8, 13, 10, 13, 12, 13, 303, 9, 13, 1, 13, 1, 13, 1, 13, 3, 13, 308,
		8, 13, 3, 13, 310, 8, 13, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 1, 14, 4,
		14, 318, 8, 14, 11, 14, 12, 14, 319, 1, 14, 1, 14, 3, 14, 324, 8, 14, 1,
		14, 1, 14, 1, 15, 1, 15, 3, 15, 330, 8, 15, 1, 16, 1, 16, 1, 16, 1, 16,
		1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 340, 8, 16, 10, 16, 12, 16, 343, 9,
		16, 3, 16, 345, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 351, 8, 16, 10,
		16, 12, 16, 354, 9, 16, 1, 16, 1, 16, 1, 16, 1, 16, 5, 16, 360, 8, 16,
		10, 16, 12, 16, 363, 9, 16, 3, 16, 365, 8, 16, 1, 16, 1, 16, 1, 16, 1,
		16, 1, 16, 1, 16, 3, 16, 373, 8, 16, 1, 16, 1, 16, 1, 16, 1, 16, 1, 16,
		1, 16, 5, 16, 381, 8, 16, 10, 16, 12, 16, 384, 9, 16, 1, 17, 1, 17, 1,
		17, 1, 17, 3, 17, 390, 8, 17, 1, 17, 3, 17, 393, 8, 17, 1, 17, 1, 17, 1,
		18, 1, 18, 1, 18, 3, 18, 400, 8, 18, 1, 19, 1, 19, 1, 19, 1, 19, 5, 19,
		406, 8, 19, 10, 19, 12, 19, 409, 9, 19, 1, 20, 1, 20, 1, 20, 1, 21, 1,
		21, 1, 21, 1, 22, 1, 22, 4, 22, 419, 8, 22, 11, 22, 12, 22, 420, 1, 23,
		1, 23, 1, 24, 1, 24, 1, 24, 1, 24, 5, 24, 429, 8, 24, 10, 24, 12, 24, 432,
		9, 24, 1, 24, 1, 24, 1, 25, 1, 25, 1, 25, 1, 25, 1, 26, 1, 26, 1, 26, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26,
		1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26,
		1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26,
		1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26,
		1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26,
		1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 1, 26, 3, 26, 544, 8, 26, 1, 26, 1,
		26, 1, 26, 1, 26, 1, 26, 1, 26, 5, 26, 552, 8, 26, 10, 26, 12, 26, 555,
		9, 26, 1, 27, 1, 27, 1, 27, 1, 27, 5, 27, 561, 8, 27, 10, 27, 12, 27, 564,
		9, 27, 1, 27, 1, 27, 1, 28, 1, 28, 1, 29, 1, 29, 1, 29, 1, 29, 5, 29, 574,
		8, 29, 10, 29, 12, 29, 577, 9, 29, 1, 30, 1, 30, 1, 30, 5, 30, 582, 8,
		30, 10, 30, 12, 30, 585, 9, 30, 1, 30, 1, 30, 1, 30, 1, 30, 1, 30, 3, 30,
		592, 8, 30, 1, 31, 1, 31, 1, 32, 1, 32, 3, 32, 598, 8, 32, 1, 33, 3, 33,
		601, 8, 33, 1, 33, 1, 33, 3, 33, 605, 8, 33, 1, 33, 3, 33, 608, 8, 33,
		1, 33, 0, 2, 32, 52, 34, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24,
		26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60,
		62, 64, 66, 0, 5, 1, 0, 66, 67, 2, 0, 26, 27, 29, 29, 1, 0, 24, 25, 2,
		0, 46, 46, 48, 51, 1, 0, 34, 35, 676, 0, 69, 1, 0, 0, 0, 2, 76, 1, 0, 0,
		0, 4, 87, 1, 0, 0, 0, 6, 89, 1, 0, 0, 0, 8, 103, 1, 0, 0, 0, 10, 124, 1,
		0, 0, 0, 12, 138, 1, 0, 0, 0, 14, 147, 1, 0, 0, 0, 16, 156, 1, 0, 0, 0,
		18, 162, 1, 0, 0, 0, 20, 173, 1, 0, 0, 0, 22, 269, 1, 0, 0, 0, 24, 271,
		1, 0, 0, 0, 26, 309, 1, 0, 0, 0, 28, 311, 1, 0, 0, 0, 30, 329, 1, 0, 0,
		0, 32, 372, 1, 0, 0, 0, 34, 385, 1, 0, 0, 0, 36, 396, 1, 0, 0, 0, 38, 401,
		1, 0, 0, 0, 40, 410, 1, 0, 0, 0, 42, 413, 1, 0, 0, 0, 44, 416, 1, 0, 0,
		0, 46, 422, 1, 0, 0, 0, 48, 424, 1, 0, 0, 0, 50, 435, 1, 0, 0, 0, 52, 543,
		1, 0, 0, 0, 54, 556, 1, 0, 0, 0, 56, 567, 1, 0, 0, 0, 58, 569, 1, 0, 0,
		0, 60, 591, 1, 0, 0, 0, 62, 593, 1, 0, 0, 0, 64, 595, 1, 0, 0, 0, 66, 607,
		1, 0, 0, 0, 68, 70, 3, 2, 1, 0, 69, 68, 1, 0, 0, 0, 70, 71, 1, 0, 0, 0,
		71, 69, 1, 0, 0, 0, 71, 72, 1, 0, 0, 0, 72, 73, 1, 0, 0, 0, 73, 74, 5,
		0, 0, 1, 74, 1, 1, 0, 0, 0, 75, 77, 3, 4, 2, 0, 76, 75, 1, 0, 0, 0, 77,
		78, 1, 0, 0, 0, 78, 76, 1, 0, 0, 0, 78, 79, 1, 0, 0, 0, 79, 80, 1, 0, 0,
		0, 80, 81, 3, 10, 5, 0, 81, 3, 1, 0, 0, 0, 82, 84, 3, 8, 4, 0, 83, 85,
		3, 12, 6, 0, 84, 83, 1, 0, 0, 0, 84, 85, 1, 0, 0, 0, 85, 88, 1, 0, 0, 0,
		86, 88, 3, 6, 3, 0, 87, 82, 1, 0, 0, 0, 87, 86, 1, 0, 0, 0, 88, 5, 1, 0,
		0, 0, 89, 91, 5, 7, 0, 0, 90, 92, 5, 5, 0, 0, 91, 90, 1, 0, 0, 0, 91, 92,
		1, 0, 0, 0, 92, 93, 1, 0, 0, 0, 93, 95, 3, 24, 12, 0, 94, 96, 3, 38, 19,
		0, 95, 94, 1, 0, 0, 0, 95, 96, 1, 0, 0, 0, 96, 98, 1, 0, 0, 0, 97, 99,
		3, 42, 21, 0, 98, 97, 1, 0, 0, 0, 98, 99, 1, 0, 0, 0, 99, 101, 1, 0, 0,
		0, 100, 102, 3, 12, 6, 0, 101, 100, 1, 0, 0, 0, 101, 102, 1, 0, 0, 0, 102,
		7, 1, 0, 0, 0, 103, 107, 5, 2, 0, 0, 104, 106, 5, 69, 0, 0, 105, 104, 1,
		0, 0, 0, 106, 109, 1, 0, 0, 0, 107, 105, 1, 0, 0, 0, 107, 108, 1, 0, 0,
		0, 108, 110, 1, 0, 0, 0, 109, 107, 1, 0, 0, 0, 110, 121, 3, 18, 9, 0, 111,
		115, 5, 17, 0, 0, 112, 114, 5, 69, 0, 0, 113, 112, 1, 0, 0, 0, 114, 117,
		1, 0, 0, 0, 115, 113, 1, 0, 0, 0, 115, 116, 1, 0, 0, 0, 116, 118, 1, 0,
		0, 0, 117, 115, 1, 0, 0, 0, 118, 120, 3, 18, 9, 0, 119, 111, 1, 0, 0, 0,
		120, 123, 1, 0, 0, 0, 121, 119, 1, 0, 0, 0, 121, 122, 1, 0, 0, 0, 122,
		9, 1, 0, 0, 0, 123, 121, 1, 0, 0, 0, 124, 128, 5, 3, 0, 0, 125, 127, 5,
		69, 0, 0, 126, 125, 1, 0, 0, 0, 127, 130, 1, 0, 0, 0, 128, 126, 1, 0, 0,
		0, 128, 129, 1, 0, 0, 0, 129, 131, 1, 0, 0, 0, 130, 128, 1, 0, 0, 0, 131,
		133, 3, 24, 12, 0, 132, 134, 3, 38, 19, 0, 133, 132, 1, 0, 0, 0, 133, 134,
		1, 0, 0, 0, 134, 136, 1, 0, 0, 0, 135, 137, 3, 42, 21, 0, 136, 135, 1,
		0, 0, 0, 136, 137, 1, 0, 0, 0, 137, 11, 1, 0, 0, 0, 138, 142, 5, 4, 0,
		0, 139, 141, 5, 69, 0, 0, 140, 139, 1, 0, 0, 0, 141, 144, 1, 0, 0, 0, 142,
		140, 1, 0, 0, 0, 142, 143, 1, 0, 0, 0, 143, 145, 1, 0, 0, 0, 144, 142,
		1, 0, 0, 0, 145, 146, 3, 52, 26, 0, 146, 13, 1, 0, 0, 0, 147, 151, 5, 60,
		0, 0, 148, 150, 5, 69, 0, 0, 149, 148, 1, 0, 0, 0, 150, 153, 1, 0, 0, 0,
		151, 149, 1, 0, 0, 0, 151, 152, 1, 0, 0, 0, 152, 154, 1, 0, 0, 0, 153,
		151, 1, 0, 0, 0, 154, 155, 5, 66, 0, 0, 155, 15, 1, 0, 0, 0, 156, 157,
		5, 6, 0, 0, 157, 158, 5, 68, 0, 0, 158, 17, 1, 0, 0, 0, 159, 160, 3, 56,
		28, 0, 160, 161, 5, 41, 0, 0, 161, 163, 1, 0, 0, 0, 162, 159, 1, 0, 0,
		0, 162, 163, 1, 0, 0, 0, 163, 164, 1, 0, 0, 0, 164, 170, 3, 20, 10, 0,
		165, 166, 3, 22, 11, 0, 166, 167, 3, 20, 10, 0, 167, 169, 1, 0, 0, 0, 168,
		165, 1, 0, 0, 0, 169, 172, 1, 0, 0, 0, 170, 168, 1, 0, 0, 0, 170, 171,
		1, 0, 0, 0, 171, 19, 1, 0, 0, 0, 172, 170, 1, 0, 0, 0, 173, 175, 5, 18,
		0, 0, 174, 176, 3, 56, 28, 0, 175, 174, 1, 0, 0, 0, 175, 176, 1, 0, 0,
		0, 176, 178, 1, 0, 0, 0, 177, 179, 3, 44, 22, 0, 178, 177, 1, 0, 0, 0,
		178, 179, 1, 0, 0, 0, 179, 181, 1, 0, 0, 0, 180, 182, 3, 48, 24, 0, 181,
		180, 1, 0, 0, 0, 181, 182, 1, 0, 0, 0, 182, 183, 1, 0, 0, 0, 183, 184,
		5, 19, 0, 0, 184, 21, 1, 0, 0, 0, 185, 189, 5, 24, 0, 0, 186, 188, 5, 69,
		0, 0, 187, 186, 1, 0, 0, 0, 188, 191, 1, 0, 0, 0, 189, 187, 1, 0, 0, 0,
		189, 190, 1, 0, 0, 0, 190, 192, 1, 0, 0, 0, 191, 189, 1, 0, 0, 0, 192,
		194, 5, 20, 0, 0, 193, 195, 3, 56, 28, 0, 194, 193, 1, 0, 0, 0, 194, 195,
		1, 0, 0, 0, 195, 197, 1, 0, 0, 0, 196, 198, 3, 58, 29, 0, 197, 196, 1,
		0, 0, 0, 197, 198, 1, 0, 0, 0, 198, 200, 1, 0, 0, 0, 199, 201, 3, 64, 32,
		0, 200, 199, 1, 0, 0, 0, 200, 201, 1, 0, 0, 0, 201, 203, 1, 0, 0, 0, 202,
		204, 3, 48, 24, 0, 203, 202, 1, 0, 0, 0, 203, 204, 1, 0, 0, 0, 204, 205,
		1, 0, 0, 0, 205, 209, 5, 21, 0, 0, 206, 208, 5, 69, 0, 0, 207, 206, 1,
		0, 0, 0, 208, 211, 1, 0, 0, 0, 209, 207, 1, 0, 0, 0, 209, 210, 1, 0, 0,
		0, 210, 212, 1, 0, 0, 0, 211, 209, 1, 0, 0, 0, 212, 270, 5, 13, 0, 0, 213,
		217, 5, 12, 0, 0, 214, 216, 5, 69, 0, 0, 215, 214, 1, 0, 0, 0, 216, 219,
		1, 0, 0, 0, 217, 215, 1, 0, 0, 0, 217, 218, 1, 0, 0, 0, 218, 220, 1, 0,
		0, 0, 219, 217, 1, 0, 0, 0, 220, 222, 5, 20, 0, 0, 221, 223, 3, 56, 28,
		0, 222, 221, 1, 0, 0, 0, 222, 223, 1, 0, 0, 0, 223, 225, 1, 0, 0, 0, 224,
		226, 3, 58, 29, 0, 225, 224, 1, 0, 0, 0, 225, 226, 1, 0, 0, 0, 226, 228,
		1, 0, 0, 0, 227, 229, 3, 64, 32, 0, 228, 227, 1, 0, 0, 0, 228, 229, 1,
		0, 0, 0, 229, 231, 1, 0, 0, 0, 230, 232, 3, 48, 24, 0, 231, 230, 1, 0,
		0, 0, 231, 232, 1, 0, 0, 0, 232, 233, 1, 0, 0, 0, 233, 237, 5, 21, 0, 0,
		234, 236, 5, 69, 0, 0, 235, 234, 1, 0, 0, 0, 236, 239, 1, 0, 0, 0, 237,
		235, 1, 0, 0, 0, 237, 238, 1, 0, 0, 0, 238, 240, 1, 0, 0, 0, 239, 237,
		1, 0, 0, 0, 240, 270, 5, 24, 0, 0, 241, 245, 5, 24, 0, 0, 242, 244, 5,
		69, 0, 0, 243, 242, 1, 0, 0, 0, 244, 247, 1, 0, 0, 0, 245, 243, 1, 0, 0,
		0, 245, 246, 1, 0, 0, 0, 246, 248, 1, 0, 0, 0, 247, 245, 1, 0, 0, 0, 248,
		250, 5, 20, 0, 0, 249, 251, 3, 56, 28, 0, 250, 249, 1, 0, 0, 0, 250, 251,
		1, 0, 0, 0, 251, 253, 1, 0, 0, 0, 252, 254, 3, 58, 29, 0, 253, 252, 1,
		0, 0, 0, 253, 254, 1, 0, 0, 0, 254, 256, 1, 0, 0, 0, 255, 257, 3, 64, 32,
		0, 256, 255, 1, 0, 0, 0, 256, 257, 1, 0, 0, 0, 257, 259, 1, 0, 0, 0, 258,
		260, 3, 48, 24, 0, 259, 258, 1, 0, 0, 0, 259, 260, 1, 0, 0, 0, 260, 261,
		1, 0, 0, 0, 261, 265, 5, 21, 0, 0, 262, 264, 5, 69, 0, 0, 263, 262, 1,
		0, 0, 0, 264, 267, 1, 0, 0, 0, 265, 263, 1, 0, 0, 0, 265, 266, 1, 0, 0,
		0, 266, 268, 1, 0, 0, 0, 267, 265, 1, 0, 0, 0, 268, 270, 5, 24, 0, 0, 269,
		185, 1, 0, 0, 0, 269, 213, 1, 0, 0, 0, 269, 241, 1, 0, 0, 0, 270, 23, 1,
		0, 0, 0, 271, 276, 3, 26, 13, 0, 272, 273, 5, 17, 0, 0, 273, 275, 3, 26,
		13, 0, 274, 272, 1, 0, 0, 0, 275, 278, 1, 0, 0, 0, 276, 274, 1, 0, 0, 0,
		276, 277, 1, 0, 0, 0, 277, 25, 1, 0, 0, 0, 278, 276, 1, 0, 0, 0, 279, 282,
		3, 34, 17, 0, 280, 281, 5, 6, 0, 0, 281, 283, 3, 56, 28, 0, 282, 280, 1,
		0, 0, 0, 282, 283, 1, 0, 0, 0, 283, 310, 1, 0, 0, 0, 284, 287, 3, 28, 14,
		0, 285, 286, 5, 6, 0, 0, 286, 288, 3, 56, 28, 0, 287, 285, 1, 0, 0, 0,
		287, 288, 1, 0, 0, 0, 288, 310, 1, 0, 0, 0, 289, 292, 3, 32, 16, 0, 290,
		291, 5, 6, 0, 0, 291, 293, 3, 56, 28, 0, 292, 290, 1, 0, 0, 0, 292, 293,
		1, 0, 0, 0, 293, 310, 1, 0, 0, 0, 294, 295, 5, 52, 0, 0, 295, 296, 5, 18,
		0, 0, 296, 301, 3, 60, 30, 0, 297, 298, 5, 17, 0, 0, 298, 300, 3, 60, 30,
		0, 299, 297, 1, 0, 0, 0, 300, 303, 1, 0, 0, 0, 301, 299, 1, 0, 0, 0, 301,
		302, 1, 0, 0, 0, 302, 304, 1, 0, 0, 0, 303, 301, 1, 0, 0, 0, 304, 307,
		5, 19, 0, 0, 305, 306, 5, 6, 0, 0, 306, 308, 3, 56, 28, 0, 307, 305, 1,
		0, 0, 0, 307, 308, 1, 0, 0, 0, 308, 310, 1, 0, 0, 0, 309, 279, 1, 0, 0,
		0, 309, 284, 1, 0, 0, 0, 309, 289, 1, 0, 0, 0, 309, 294, 1, 0, 0, 0, 310,
		27, 1, 0, 0, 0, 311, 317, 5, 61, 0, 0, 312, 313, 5, 62, 0, 0, 313, 314,
		3, 52, 26, 0, 314, 315, 5, 63, 0, 0, 315, 316, 3, 30, 15, 0, 316, 318,
		1, 0, 0, 0, 317, 312, 1, 0, 0, 0, 318, 319, 1, 0, 0, 0, 319, 317, 1, 0,
		0, 0, 319, 320, 1, 0, 0, 0, 320, 323, 1, 0, 0, 0, 321, 322, 5, 64, 0, 0,
		322, 324, 3, 30, 15, 0, 323, 321, 1, 0, 0, 0, 323, 324, 1, 0, 0, 0, 324,
		325, 1, 0, 0, 0, 325, 326, 5, 65, 0, 0, 326, 29, 1, 0, 0, 0, 327, 330,
		3, 62, 31, 0, 328, 330, 3, 60, 30, 0, 329, 327, 1, 0, 0, 0, 329, 328, 1,
		0, 0, 0, 330, 31, 1, 0, 0, 0, 331, 332, 6, 16, -1, 0, 332, 333, 5, 24,
		0, 0, 333, 373, 3, 32, 16, 8, 334, 335, 5, 68, 0, 0, 335, 344, 5, 18, 0,
		0, 336, 341, 3, 32, 16, 0, 337, 338, 5, 17, 0, 0, 338, 340, 3, 32, 16,
		0, 339, 337, 1, 0, 0, 0, 340, 343, 1, 0, 0, 0, 341, 339, 1, 0, 0, 0, 341,
		342, 1, 0, 0, 0, 342, 345, 1, 0, 0, 0, 343, 341, 1, 0, 0, 0, 344, 336,
		1, 0, 0, 0, 344, 345, 1, 0, 0, 0, 345, 346, 1, 0, 0, 0, 346, 373, 5, 19,
		0, 0, 347, 352, 5, 68, 0, 0, 348, 349, 5, 11, 0, 0, 349, 351, 5, 68, 0,
		0, 350, 348, 1, 0, 0, 0, 351, 354, 1, 0, 0, 0, 352, 350, 1, 0, 0, 0, 352,
		353, 1, 0, 0, 0, 353, 373, 1, 0, 0, 0, 354, 352, 1, 0, 0, 0, 355, 364,
		5, 20, 0, 0, 356, 361, 3, 32, 16, 0, 357, 358, 5, 17, 0, 0, 358, 360, 3,
		32, 16, 0, 359, 357, 1, 0, 0, 0, 360, 363, 1, 0, 0, 0, 361, 359, 1, 0,
		0, 0, 361, 362, 1, 0, 0, 0, 362, 365, 1, 0, 0, 0, 363, 361, 1, 0, 0, 0,
		364, 356, 1, 0, 0, 0, 364, 365, 1, 0, 0, 0, 365, 366, 1, 0, 0, 0, 366,
		373, 5, 21, 0, 0, 367, 373, 7, 0, 0, 0, 368, 369, 5, 18, 0, 0, 369, 370,
		3, 32, 16, 0, 370, 371, 5, 19, 0, 0, 371, 373, 1, 0, 0, 0, 372, 331, 1,
		0, 0, 0, 372, 334, 1, 0, 0, 0, 372, 347, 1, 0, 0, 0, 372, 355, 1, 0, 0,
		0, 372, 367, 1, 0, 0, 0, 372, 368, 1, 0, 0, 0, 373, 382, 1, 0, 0, 0, 374,
		375, 10, 7, 0, 0, 375, 376, 7, 1, 0, 0, 376, 381, 3, 32, 16, 8, 377, 378,
		10, 6, 0, 0, 378, 379, 7, 2, 0, 0, 379, 381, 3, 32, 16, 7, 380, 374, 1,
		0, 0, 0, 380, 377, 1, 0, 0, 0, 381, 384, 1, 0, 0, 0, 382, 380, 1, 0, 0,
		0, 382, 383, 1, 0, 0, 0, 383, 33, 1, 0, 0, 0, 384, 382, 1, 0, 0, 0, 385,
		386, 7, 3, 0, 0, 386, 392, 5, 18, 0, 0, 387, 393, 5, 29, 0, 0, 388, 390,
		5, 5, 0, 0, 389, 388, 1, 0, 0, 0, 389, 390, 1, 0, 0, 0, 390, 391, 1, 0,
		0, 0, 391, 393, 3, 36, 18, 0, 392, 387, 1, 0, 0, 0, 392, 389, 1, 0, 0,
		0, 393, 394, 1, 0, 0, 0, 394, 395, 5, 19, 0, 0, 395, 35, 1, 0, 0, 0, 396,
		399, 5, 68, 0, 0, 397, 398, 5, 11, 0, 0, 398, 400, 5, 68, 0, 0, 399, 397,
		1, 0, 0, 0, 399, 400, 1, 0, 0, 0, 400, 37, 1, 0, 0, 0, 401, 402, 5, 33,
		0, 0, 402, 407, 3, 40, 20, 0, 403, 404, 5, 17, 0, 0, 404, 406, 3, 40, 20,
		0, 405, 403, 1, 0, 0, 0, 406, 409, 1, 0, 0, 0, 407, 405, 1, 0, 0, 0, 407,
		408, 1, 0, 0, 0, 408, 39, 1, 0, 0, 0, 409, 407, 1, 0, 0, 0, 410, 411, 3,
		60, 30, 0, 411, 412, 7, 4, 0, 0, 412, 41, 1, 0, 0, 0, 413, 414, 5, 36,
		0, 0, 414, 415, 5, 67, 0, 0, 415, 43, 1, 0, 0, 0, 416, 418, 5, 16, 0, 0,
		417, 419, 3, 46, 23, 0, 418, 417, 1, 0, 0, 0, 419, 420, 1, 0, 0, 0, 420,
		418, 1, 0, 0, 0, 420, 421, 1, 0, 0, 0, 421, 45, 1, 0, 0, 0, 422, 423, 5,
		68, 0, 0, 423, 47, 1, 0, 0, 0, 424, 425, 5, 22, 0, 0, 425, 430, 3, 50,
		25, 0, 426, 427, 5, 17, 0, 0, 427, 429, 3, 50, 25, 0, 428, 426, 1, 0, 0,
		0, 429, 432, 1, 0, 0, 0, 430, 428, 1, 0, 0, 0, 430, 431, 1, 0, 0, 0, 431,
		433, 1, 0, 0, 0, 432, 430, 1, 0, 0, 0, 433, 434, 5, 23, 0, 0, 434, 49,
		1, 0, 0, 0, 435, 436, 5, 68, 0, 0, 436, 437, 5, 16, 0, 0, 437, 438, 3,
		62, 31, 0, 438, 51, 1, 0, 0, 0, 439, 440, 6, 26, -1, 0, 440, 441, 5, 18,
		0, 0, 441, 442, 3, 52, 26, 0, 442, 443, 5, 19, 0, 0, 443, 544, 1, 0, 0,
		0, 444, 445, 5, 44, 0, 0, 445, 544, 3, 52, 26, 22, 446, 447, 5, 56, 0,
		0, 447, 448, 5, 18, 0, 0, 448, 449, 3, 56, 28, 0, 449, 450, 5, 53, 0, 0,
		450, 451, 3, 60, 30, 0, 451, 452, 5, 4, 0, 0, 452, 453, 3, 52, 26, 0, 453,
		454, 5, 19, 0, 0, 454, 544, 1, 0, 0, 0, 455, 456, 5, 57, 0, 0, 456, 457,
		5, 18, 0, 0, 457, 458, 3, 56, 28, 0, 458, 459, 5, 53, 0, 0, 459, 460, 3,
		60, 30, 0, 460, 461, 5, 4, 0, 0, 461, 462, 3, 52, 26, 0, 462, 463, 5, 19,
		0, 0, 463, 544, 1, 0, 0, 0, 464, 465, 5, 58, 0, 0, 465, 466, 5, 18, 0,
		0, 466, 467, 3, 56, 28, 0, 467, 468, 5, 53, 0, 0, 468, 469, 3, 60, 30,
		0, 469, 470, 5, 4, 0, 0, 470, 471, 3, 52, 26, 0, 471, 472, 5, 19, 0, 0,
		472, 544, 1, 0, 0, 0, 473, 474, 5, 59, 0, 0, 474, 475, 5, 18, 0, 0, 475,
		476, 3, 56, 28, 0, 476, 477, 5, 53, 0, 0, 477, 478, 3, 60, 30, 0, 478,
		479, 5, 4, 0, 0, 479, 480, 3, 52, 26, 0, 480, 481, 5, 19, 0, 0, 481, 544,
		1, 0, 0, 0, 482, 483, 3, 60, 30, 0, 483, 484, 5, 54, 0, 0, 484, 485, 5,
		55, 0, 0, 485, 544, 1, 0, 0, 0, 486, 487, 3, 60, 30, 0, 487, 488, 5, 54,
		0, 0, 488, 489, 5, 44, 0, 0, 489, 490, 5, 55, 0, 0, 490, 544, 1, 0, 0,
		0, 491, 492, 3, 60, 30, 0, 492, 493, 5, 53, 0, 0, 493, 494, 3, 54, 27,
		0, 494, 544, 1, 0, 0, 0, 495, 496, 3, 60, 30, 0, 496, 497, 5, 41, 0, 0,
		497, 498, 3, 62, 31, 0, 498, 544, 1, 0, 0, 0, 499, 500, 3, 60, 30, 0, 500,
		501, 5, 8, 0, 0, 501, 502, 3, 62, 31, 0, 502, 544, 1, 0, 0, 0, 503, 504,
		3, 60, 30, 0, 504, 505, 5, 15, 0, 0, 505, 506, 3, 62, 31, 0, 506, 544,
		1, 0, 0, 0, 507, 508, 3, 60, 30, 0, 508, 509, 5, 14, 0, 0, 509, 510, 3,
		62, 31, 0, 510, 544, 1, 0, 0, 0, 511, 512, 3, 60, 30, 0, 512, 513, 5, 10,
		0, 0, 513, 514, 3, 62, 31, 0, 514, 544, 1, 0, 0, 0, 515, 516, 3, 60, 30,
		0, 516, 517, 5, 9, 0, 0, 517, 518, 3, 62, 31, 0, 518, 544, 1, 0, 0, 0,
		519, 520, 3, 60, 30, 0, 520, 521, 5, 41, 0, 0, 521, 522, 3, 60, 30, 0,
		522, 544, 1, 0, 0, 0, 523, 524, 3, 60, 30, 0, 524, 525, 5, 8, 0, 0, 525,
		526, 3, 60, 30, 0, 526, 544, 1, 0, 0, 0, 527, 528, 3, 60, 30, 0, 528, 529,
		5, 15, 0, 0, 529, 530, 3, 60, 30, 0, 530, 544, 1, 0, 0, 0, 531, 532, 3,
		60, 30, 0, 532, 533, 5, 14, 0, 0, 533, 534, 3, 60, 30, 0, 534, 544, 1,
		0, 0, 0, 535, 536, 3, 60, 30, 0, 536, 537, 5, 10, 0, 0, 537, 538, 3, 60,
		30, 0, 538, 544, 1, 0, 0, 0, 539, 540, 3, 60, 30, 0, 540, 541, 5, 9, 0,
		0, 541, 542, 3, 60, 30, 0, 542, 544, 1, 0, 0, 0, 543, 439, 1, 0, 0, 0,
		543, 444, 1, 0, 0, 0, 543, 446, 1, 0, 0, 0, 543, 455, 1, 0, 0, 0, 543,
		464, 1, 0, 0, 0, 543, 473, 1, 0, 0, 0, 543, 482, 1, 0, 0, 0, 543, 486,
		1, 0, 0, 0, 543, 491, 1, 0, 0, 0, 543, 495, 1, 0, 0, 0, 543, 499, 1, 0,
		0, 0, 543, 503, 1, 0, 0, 0, 543, 507, 1, 0, 0, 0, 543, 511, 1, 0, 0, 0,
		543, 515, 1, 0, 0, 0, 543, 519, 1, 0, 0, 0, 543, 523, 1, 0, 0, 0, 543,
		527, 1, 0, 0, 0, 543, 531, 1, 0, 0, 0, 543, 535, 1, 0, 0, 0, 543, 539,
		1, 0, 0, 0, 544, 553, 1, 0, 0, 0, 545, 546, 10, 21, 0, 0, 546, 547, 5,
		42, 0, 0, 547, 552, 3, 52, 26, 22, 548, 549, 10, 20, 0, 0, 549, 550, 5,
		43, 0, 0, 550, 552, 3, 52, 26, 21, 551, 545, 1, 0, 0, 0, 551, 548, 1, 0,
		0, 0, 552, 555, 1, 0, 0, 0, 553, 551, 1, 0, 0, 0, 553, 554, 1, 0, 0, 0,
		554, 53, 1, 0, 0, 0, 555, 553, 1, 0, 0, 0, 556, 557, 5, 20, 0, 0, 557,
		562, 3, 62, 31, 0, 558, 559, 5, 17, 0, 0, 559, 561, 3, 62, 31, 0, 560,
		558, 1, 0, 0, 0, 561, 564, 1, 0, 0, 0, 562, 560, 1, 0, 0, 0, 562, 563,
		1, 0, 0, 0, 563, 565, 1, 0, 0, 0, 564, 562, 1, 0, 0, 0, 565, 566, 5, 21,
		0, 0, 566, 55, 1, 0, 0, 0, 567, 568, 5, 68, 0, 0, 568, 57, 1, 0, 0, 0,
		569, 570, 5, 16, 0, 0, 570, 575, 5, 68, 0, 0, 571, 572, 5, 1, 0, 0, 572,
		574, 5, 68, 0, 0, 573, 571, 1, 0, 0, 0, 574, 577, 1, 0, 0, 0, 575, 573,
		1, 0, 0, 0, 575, 576, 1, 0, 0, 0, 576, 59, 1, 0, 0, 0, 577, 575, 1, 0,
		0, 0, 578, 583, 5, 68, 0, 0, 579, 580, 5, 11, 0, 0, 580, 582, 5, 68, 0,
		0, 581, 579, 1, 0, 0, 0, 582, 585, 1, 0, 0, 0, 583, 581, 1, 0, 0, 0, 583,
		584, 1, 0, 0, 0, 584, 592, 1, 0, 0, 0, 585, 583, 1, 0, 0, 0, 586, 587,
		5, 68, 0, 0, 587, 588, 5, 18, 0, 0, 588, 589, 3, 56, 28, 0, 589, 590, 5,
		19, 0, 0, 590, 592, 1, 0, 0, 0, 591, 578, 1, 0, 0, 0, 591, 586, 1, 0, 0,
		0, 592, 61, 1, 0, 0, 0, 593, 594, 7, 0, 0, 0, 594, 63, 1, 0, 0, 0, 595,
		597, 5, 29, 0, 0, 596, 598, 3, 66, 33, 0, 597, 596, 1, 0, 0, 0, 597, 598,
		1, 0, 0, 0, 598, 65, 1, 0, 0, 0, 599, 601, 5, 67, 0, 0, 600, 599, 1, 0,
		0, 0, 600, 601, 1, 0, 0, 0, 601, 602, 1, 0, 0, 0, 602, 604, 5, 30, 0, 0,
		603, 605, 5, 67, 0, 0, 604, 603, 1, 0, 0, 0, 604, 605, 1, 0, 0, 0, 605,
		608, 1, 0, 0, 0, 606, 608, 5, 67, 0, 0, 607, 600, 1, 0, 0, 0, 607, 606,
		1, 0, 0, 0, 608, 67, 1, 0, 0, 0, 75, 71, 78, 84, 87, 91, 95, 98, 101, 107,
		115, 121, 128, 133, 136, 142, 151, 162, 170, 175, 178, 181, 189, 194, 197,
		200, 203, 209, 217, 222, 225, 228, 231, 237, 245, 250, 253, 256, 259, 265,
		269, 276, 282, 287, 292, 301, 307, 309, 319, 323, 329, 341, 344, 352, 361,
		364, 372, 380, 382, 389, 392, 399, 407, 420, 430, 543, 551, 553, 562, 575,
		583, 591, 597, 600, 604, 607,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// CypherParserInit initializes any static state used to implement CypherParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewCypherParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func CypherParserInit() {
	staticData := &CypherParserStaticData
	staticData.once.Do(cypherParserInit)
}

// NewCypherParser produces a new parser instance for the optional input antlr.TokenStream.
func NewCypherParser(input antlr.TokenStream) *CypherParser {
	CypherParserInit()
	this := new(CypherParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &CypherParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Cypher.g4"

	return this
}

// CypherParser tokens.
const (
	CypherParserEOF        = antlr.TokenEOF
	CypherParserT__0       = 1
	CypherParserMATCH      = 2
	CypherParserRETURN     = 3
	CypherParserWHERE      = 4
	CypherParserDISTINCT   = 5
	CypherParserAS         = 6
	CypherParserWITH       = 7
	CypherParserNEQ        = 8
	CypherParserLE         = 9
	CypherParserGE         = 10
	CypherParserDOT        = 11
	CypherParserLARROW     = 12
	CypherParserRARROW     = 13
	CypherParserLANGLE     = 14
	CypherParserRANGLE     = 15
	CypherParserCOLON      = 16
	CypherParserCOMMA      = 17
	CypherParserLPAREN     = 18
	CypherParserRPAREN     = 19
	CypherParserLSQUARE    = 20
	CypherParserRSQUARE    = 21
	CypherParserLCURLY     = 22
	CypherParserRCURLY     = 23
	CypherParserMINUS      = 24
	CypherParserPLUS       = 25
	CypherParserSLASH      = 26
	CypherParserPERCENT    = 27
	CypherParserSQUOTE     = 28
	CypherParserSTAR       = 29
	CypherParserDOUBLE_DOT = 30
	CypherParserCREATE     = 31
	CypherParserDELETE     = 32
	CypherParserORDER_BY   = 33
	CypherParserASC        = 34
	CypherParserDESC       = 35
	CypherParserLIMIT      = 36
	CypherParserOPTIONAL   = 37
	CypherParserUNWIND     = 38
	CypherParserFINISH     = 39
	CypherParserSET        = 40
	CypherParserEQ         = 41
	CypherParserAND        = 42
	CypherParserOR         = 43
	CypherParserNOT        = 44
	CypherParserXOR        = 45
	CypherParserCOUNT      = 46
	CypherParserREDUCE     = 47
	CypherParserSUM        = 48
	CypherParserAVG        = 49
	CypherParserMIN        = 50
	CypherParserMAX        = 51
	CypherParserCOALESCE   = 52
	CypherParserIN         = 53
	CypherParserIS         = 54
	CypherParserNULL       = 55
	CypherParserALL        = 56
	CypherParserANY        = 57
	CypherParserNONE       = 58
	CypherParserSINGLE     = 59
	CypherParserCALL       = 60
	CypherParserCASE       = 61
	CypherParserWHEN       = 62
	CypherParserTHEN       = 63
	CypherParserELSE       = 64
	CypherParserEND        = 65
	CypherParserSTRING     = 66
	CypherParserNUMBER     = 67
	CypherParserIDENTIFIER = 68
	CypherParserWS         = 69
)

// CypherParser rules.
const (
	CypherParserRULE_cypher        = 0
	CypherParserRULE_statement     = 1
	CypherParserRULE_readingClause = 2
	CypherParserRULE_withClause    = 3
	CypherParserRULE_matchClause   = 4
	CypherParserRULE_returnClause  = 5
	CypherParserRULE_whereClause   = 6
	CypherParserRULE_callClause    = 7
	CypherParserRULE_asClause      = 8
	CypherParserRULE_pattern       = 9
	CypherParserRULE_node          = 10
	CypherParserRULE_relationship  = 11
	CypherParserRULE_returnItems   = 12
	CypherParserRULE_returnItem    = 13
	CypherParserRULE_caseExpr      = 14
	CypherParserRULE_caseResult    = 15
	CypherParserRULE_valueExpr     = 16
	CypherParserRULE_aggregateFunc = 17
	CypherParserRULE_aggArg        = 18
	CypherParserRULE_orderItems    = 19
	CypherParserRULE_orderItem     = 20
	CypherParserRULE_limitNum      = 21
	CypherParserRULE_labels        = 22
	CypherParserRULE_label         = 23
	CypherParserRULE_properties    = 24
	CypherParserRULE_property      = 25
	CypherParserRULE_condition     = 26
	CypherParserRULE_list          = 27
	CypherParserRULE_variable      = 28
	CypherParserRULE_types         = 29
	CypherParserRULE_expression    = 30
	CypherParserRULE_value         = 31
	CypherParserRULE_range         = 32
	CypherParserRULE_rangeLiteral  = 33
)

// ICypherContext is an interface to support dynamic dispatch.
type ICypherContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllStatement() []IStatementContext
	Statement(i int) IStatementContext

	// IsCypherContext differentiates from other interfaces.
	IsCypherContext()
}

type CypherContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCypherContext() *CypherContext {
	var p = new(CypherContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_cypher
	return p
}

func InitEmptyCypherContext(p *CypherContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_cypher
}

func (*CypherContext) IsCypherContext() {}

func NewCypherContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CypherContext {
	var p = new(CypherContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_cypher

	return p
}

func (s *CypherContext) GetParser() antlr.Parser { return s.parser }

func (s *CypherContext) EOF() antlr.TerminalNode {
	return s.GetToken(CypherParserEOF, 0)
}

func (s *CypherContext) AllStatement() []IStatementContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IStatementContext); ok {
			len++
		}
	}

	tst := make([]IStatementContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IStatementContext); ok {
			tst[i] = t.(IStatementContext)
			i++
		}
	}

	return tst
}

func (s *CypherContext) Statement(i int) IStatementContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IStatementContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IStatementContext)
}

func (s *CypherContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CypherContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CypherContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterCypher(s)
	}
}

func (s *CypherContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitCypher(s)
	}
}

func (p *CypherParser) Cypher() (localctx ICypherContext) {
	localctx = NewCypherContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, CypherParserRULE_cypher)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(69)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == CypherParserMATCH || _la == CypherParserWITH {
		{
			p.SetState(68)
			p.Statement()
		}

		p.SetState(71)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(73)
		p.Match(CypherParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IStatementContext is an interface to support dynamic dispatch.
type IStatementContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ReturnClause() IReturnClauseContext
	AllReadingClause() []IReadingClauseContext
	ReadingClause(i int) IReadingClauseContext

	// IsStatementContext differentiates from other interfaces.
	IsStatementContext()
}

type StatementContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyStatementContext() *StatementContext {
	var p = new(StatementContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_statement
	return p
}

func InitEmptyStatementContext(p *StatementContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_statement
}

func (*StatementContext) IsStatementContext() {}

func NewStatementContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *StatementContext {
	var p = new(StatementContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_statement

	return p
}

func (s *StatementContext) GetParser() antlr.Parser { return s.parser }

func (s *StatementContext) ReturnClause() IReturnClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturnClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturnClauseContext)
}

func (s *StatementContext) AllReadingClause() []IReadingClauseContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IReadingClauseContext); ok {
			len++
		}
	}

	tst := make([]IReadingClauseContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IReadingClauseContext); ok {
			tst[i] = t.(IReadingClauseContext)
			i++
		}
	}

	return tst
}

func (s *StatementContext) ReadingClause(i int) IReadingClauseContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReadingClauseContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReadingClauseContext)
}

func (s *StatementContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StatementContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *StatementContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterStatement(s)
	}
}

func (s *StatementContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitStatement(s)
	}
}

func (p *CypherParser) Statement() (localctx IStatementContext) {
	localctx = NewStatementContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, CypherParserRULE_statement)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(76)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == CypherParserMATCH || _la == CypherParserWITH {
		{
			p.SetState(75)
			p.ReadingClause()
		}

		p.SetState(78)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(80)
		p.ReturnClause()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReadingClauseContext is an interface to support dynamic dispatch.
type IReadingClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MatchClause() IMatchClauseContext
	WhereClause() IWhereClauseContext
	WithClause() IWithClauseContext

	// IsReadingClauseContext differentiates from other interfaces.
	IsReadingClauseContext()
}

type ReadingClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReadingClauseContext() *ReadingClauseContext {
	var p = new(ReadingClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_readingClause
	return p
}

func InitEmptyReadingClauseContext(p *ReadingClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_readingClause
}

func (*ReadingClauseContext) IsReadingClauseContext() {}

func NewReadingClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ReadingClauseContext {
	var p = new(ReadingClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_readingClause

	return p
}

func (s *ReadingClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *ReadingClauseContext) MatchClause() IMatchClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IMatchClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IMatchClauseContext)
}

func (s *ReadingClauseContext) WhereClause() IWhereClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhereClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWhereClauseContext)
}

func (s *ReadingClauseContext) WithClause() IWithClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWithClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWithClauseContext)
}

func (s *ReadingClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ReadingClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ReadingClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterReadingClause(s)
	}
}

func (s *ReadingClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitReadingClause(s)
	}
}

func (p *CypherParser) ReadingClause() (localctx IReadingClauseContext) {
	localctx = NewReadingClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, CypherParserRULE_readingClause)
	var _la int

	p.SetState(87)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserMATCH:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(82)
			p.MatchClause()
		}
		p.SetState(84)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserWHERE {
			{
				p.SetState(83)
				p.WhereClause()
			}

		}

	case CypherParserWITH:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(86)
			p.WithClause()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IWithClauseContext is an interface to support dynamic dispatch.
type IWithClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	WITH() antlr.TerminalNode
	ReturnItems() IReturnItemsContext
	DISTINCT() antlr.TerminalNode
	OrderItems() IOrderItemsContext
	LimitNum() ILimitNumContext
	WhereClause() IWhereClauseContext

	// IsWithClauseContext differentiates from other interfaces.
	IsWithClauseContext()
}

type WithClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWithClauseContext() *WithClauseContext {
	var p = new(WithClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_withClause
	return p
}

func InitEmptyWithClauseContext(p *WithClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_withClause
}

func (*WithClauseContext) IsWithClauseContext() {}

func NewWithClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WithClauseContext {
	var p = new(WithClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_withClause

	return p
}

func (s *WithClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *WithClauseContext) WITH() antlr.TerminalNode {
	return s.GetToken(CypherParserWITH, 0)
}

func (s *WithClauseContext) ReturnItems() IReturnItemsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturnItemsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturnItemsContext)
}

func (s *WithClauseContext) DISTINCT() antlr.TerminalNode {
	return s.GetToken(CypherParserDISTINCT, 0)
}

func (s *WithClauseContext) OrderItems() IOrderItemsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrderItemsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrderItemsContext)
}

func (s *WithClauseContext) LimitNum() ILimitNumContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILimitNumContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILimitNumContext)
}

func (s *WithClauseContext) WhereClause() IWhereClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IWhereClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IWhereClauseContext)
}

func (s *WithClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WithClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WithClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterWithClause(s)
	}
}

func (s *WithClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitWithClause(s)
	}
}

func (p *CypherParser) WithClause() (localctx IWithClauseContext) {
	localctx = NewWithClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, CypherParserRULE_withClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(89)
		p.Match(CypherParserWITH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(91)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDISTINCT {
		{
			p.SetState(90)
			p.Match(CypherParserDISTINCT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(93)
		p.ReturnItems()
	}
	p.SetState(95)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserORDER_BY {
		{
			p.SetState(94)
			p.OrderItems()
		}

	}
	p.SetState(98)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLIMIT {
		{
			p.SetState(97)
			p.LimitNum()
		}

	}
	p.SetState(101)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserWHERE {
		{
			p.SetState(100)
			p.WhereClause()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IMatchClauseContext is an interface to support dynamic dispatch.
type IMatchClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MATCH() antlr.TerminalNode
	AllPattern() []IPatternContext
	Pattern(i int) IPatternContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsMatchClauseContext differentiates from other interfaces.
	IsMatchClauseContext()
}

type MatchClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyMatchClauseContext() *MatchClauseContext {
	var p = new(MatchClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_matchClause
	return p
}

func InitEmptyMatchClauseContext(p *MatchClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_matchClause
}

func (*MatchClauseContext) IsMatchClauseContext() {}

func NewMatchClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *MatchClauseContext {
	var p = new(MatchClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_matchClause

	return p
}

func (s *MatchClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *MatchClauseContext) MATCH() antlr.TerminalNode {
	return s.GetToken(CypherParserMATCH, 0)
}

func (s *MatchClauseContext) AllPattern() []IPatternContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPatternContext); ok {
			len++
		}
	}

	tst := make([]IPatternContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPatternContext); ok {
			tst[i] = t.(IPatternContext)
			i++
		}
	}

	return tst
}

func (s *MatchClauseContext) Pattern(i int) IPatternContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPatternContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPatternContext)
}

func (s *MatchClauseContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWS)
}

func (s *MatchClauseContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWS, i)
}

func (s *MatchClauseContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *MatchClauseContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *MatchClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MatchClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *MatchClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterMatchClause(s)
	}
}

func (s *MatchClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitMatchClause(s)
	}
}

func (p *CypherParser) MatchClause() (localctx IMatchClauseContext) {
	localctx = NewMatchClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, CypherParserRULE_matchClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(103)
		p.Match(CypherParserMATCH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(107)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(104)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(109)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(110)
		p.Pattern()
	}
	p.SetState(121)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(111)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(115)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(112)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(117)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(118)
			p.Pattern()
		}

		p.SetState(123)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReturnClauseContext is an interface to support dynamic dispatch.
type IReturnClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	RETURN() antlr.TerminalNode
	ReturnItems() IReturnItemsContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode
	OrderItems() IOrderItemsContext
	LimitNum() ILimitNumContext

	// IsReturnClauseContext differentiates from other interfaces.
	IsReturnClauseContext()
}

type ReturnClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReturnClauseContext() *ReturnClauseContext {
	var p = new(ReturnClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_returnClause
	return p
}

func InitEmptyReturnClauseContext(p *ReturnClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_returnClause
}

func (*ReturnClauseContext) IsReturnClauseContext() {}

func NewReturnClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ReturnClauseContext {
	var p = new(ReturnClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_returnClause

	return p
}

func (s *ReturnClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *ReturnClauseContext) RETURN() antlr.TerminalNode {
	return s.GetToken(CypherParserRETURN, 0)
}

func (s *ReturnClauseContext) ReturnItems() IReturnItemsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturnItemsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturnItemsContext)
}

func (s *ReturnClauseContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWS)
}

func (s *ReturnClauseContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWS, i)
}

func (s *ReturnClauseContext) OrderItems() IOrderItemsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrderItemsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrderItemsContext)
}

func (s *ReturnClauseContext) LimitNum() ILimitNumContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILimitNumContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILimitNumContext)
}

func (s *ReturnClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ReturnClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ReturnClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterReturnClause(s)
	}
}

func (s *ReturnClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitReturnClause(s)
	}
}

func (p *CypherParser) ReturnClause() (localctx IReturnClauseContext) {
	localctx = NewReturnClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, CypherParserRULE_returnClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(124)
		p.Match(CypherParserRETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(128)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(125)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(130)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(131)
		p.ReturnItems()
	}
	p.SetState(133)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserORDER_BY {
		{
			p.SetState(132)
			p.OrderItems()
		}

	}
	p.SetState(136)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLIMIT {
		{
			p.SetState(135)
			p.LimitNum()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IWhereClauseContext is an interface to support dynamic dispatch.
type IWhereClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	WHERE() antlr.TerminalNode
	Condition() IConditionContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsWhereClauseContext differentiates from other interfaces.
	IsWhereClauseContext()
}

type WhereClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyWhereClauseContext() *WhereClauseContext {
	var p = new(WhereClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_whereClause
	return p
}

func InitEmptyWhereClauseContext(p *WhereClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_whereClause
}

func (*WhereClauseContext) IsWhereClauseContext() {}

func NewWhereClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *WhereClauseContext {
	var p = new(WhereClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_whereClause

	return p
}

func (s *WhereClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *WhereClauseContext) WHERE() antlr.TerminalNode {
	return s.GetToken(CypherParserWHERE, 0)
}

func (s *WhereClauseContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *WhereClauseContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWS)
}

func (s *WhereClauseContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWS, i)
}

func (s *WhereClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WhereClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *WhereClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterWhereClause(s)
	}
}

func (s *WhereClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitWhereClause(s)
	}
}

func (p *CypherParser) WhereClause() (localctx IWhereClauseContext) {
	localctx = NewWhereClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, CypherParserRULE_whereClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(138)
		p.Match(CypherParserWHERE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(142)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(139)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(144)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(145)
		p.condition(0)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICallClauseContext is an interface to support dynamic dispatch.
type ICallClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CALL() antlr.TerminalNode
	STRING() antlr.TerminalNode
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsCallClauseContext differentiates from other interfaces.
	IsCallClauseContext()
}

type CallClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCallClauseContext() *CallClauseContext {
	var p = new(CallClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_callClause
	return p
}

func InitEmptyCallClauseContext(p *CallClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_callClause
}

func (*CallClauseContext) IsCallClauseContext() {}

func NewCallClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CallClauseContext {
	var p = new(CallClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_callClause

	return p
}

func (s *CallClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *CallClauseContext) CALL() antlr.TerminalNode {
	return s.GetToken(CypherParserCALL, 0)
}

func (s *CallClauseContext) STRING() antlr.TerminalNode {
	return s.GetToken(CypherParserSTRING, 0)
}

func (s *CallClauseContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWS)
}

func (s *CallClauseContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWS, i)
}

func (s *CallClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CallClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CallClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterCallClause(s)
	}
}

func (s *CallClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitCallClause(s)
	}
}

func (p *CypherParser) CallClause() (localctx ICallClauseContext) {
	localctx = NewCallClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, CypherParserRULE_callClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(147)
		p.Match(CypherParserCALL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(151)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(148)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(153)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(154)
		p.Match(CypherParserSTRING)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAsClauseContext is an interface to support dynamic dispatch.
type IAsClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AS() antlr.TerminalNode
	IDENTIFIER() antlr.TerminalNode

	// IsAsClauseContext differentiates from other interfaces.
	IsAsClauseContext()
}

type AsClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAsClauseContext() *AsClauseContext {
	var p = new(AsClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_asClause
	return p
}

func InitEmptyAsClauseContext(p *AsClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_asClause
}

func (*AsClauseContext) IsAsClauseContext() {}

func NewAsClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AsClauseContext {
	var p = new(AsClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_asClause

	return p
}

func (s *AsClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *AsClauseContext) AS() antlr.TerminalNode {
	return s.GetToken(CypherParserAS, 0)
}

func (s *AsClauseContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, 0)
}

func (s *AsClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AsClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AsClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterAsClause(s)
	}
}

func (s *AsClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitAsClause(s)
	}
}

func (p *CypherParser) AsClause() (localctx IAsClauseContext) {
	localctx = NewAsClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, CypherParserRULE_asClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(156)
		p.Match(CypherParserAS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(157)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPatternContext is an interface to support dynamic dispatch.
type IPatternContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNode() []INodeContext
	Node(i int) INodeContext
	Variable() IVariableContext
	EQ() antlr.TerminalNode
	AllRelationship() []IRelationshipContext
	Relationship(i int) IRelationshipContext

	// IsPatternContext differentiates from other interfaces.
	IsPatternContext()
}

type PatternContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPatternContext() *PatternContext {
	var p = new(PatternContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_pattern
	return p
}

func InitEmptyPatternContext(p *PatternContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_pattern
}

func (*PatternContext) IsPatternContext() {}

func NewPatternContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PatternContext {
	var p = new(PatternContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_pattern

	return p
}

func (s *PatternContext) GetParser() antlr.Parser { return s.parser }

func (s *PatternContext) AllNode() []INodeContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(INodeContext); ok {
			len++
		}
	}

	tst := make([]INodeContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(INodeContext); ok {
			tst[i] = t.(INodeContext)
			i++
		}
	}

	return tst
}

func (s *PatternContext) Node(i int) INodeContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(INodeContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(INodeContext)
}

func (s *PatternContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *PatternContext) EQ() antlr.TerminalNode {
	return s.GetToken(CypherParserEQ, 0)
}

func (s *PatternContext) AllRelationship() []IRelationshipContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRelationshipContext); ok {
			len++
		}
	}

	tst := make([]IRelationshipContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRelationshipContext); ok {
			tst[i] = t.(IRelationshipContext)
			i++
		}
	}

	return tst
}

func (s *PatternContext) Relationship(i int) IRelationshipContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRelationshipContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRelationshipContext)
}

func (s *PatternContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PatternContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PatternContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterPattern(s)
	}
}

func (s *PatternContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitPattern(s)
	}
}

func (p *CypherParser) Pattern() (localctx IPatternContext) {
	localctx = NewPatternContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, CypherParserRULE_pattern)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(162)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserIDENTIFIER {
		{
			p.SetState(159)
			p.Variable()
		}
		{
			p.SetState(160)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(164)
		p.Node()
	}
	p.SetState(170)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserLARROW || _la == CypherParserMINUS {
		{
			p.SetState(165)
			p.Relationship()
		}
		{
			p.SetState(166)
			p.Node()
		}

		p.SetState(172)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// INodeContext is an interface to support dynamic dispatch.
type INodeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	Variable() IVariableContext
	Labels() ILabelsContext
	Properties() IPropertiesContext

	// IsNodeContext differentiates from other interfaces.
	IsNodeContext()
}

type NodeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyNodeContext() *NodeContext {
	var p = new(NodeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_node
	return p
}

func InitEmptyNodeContext(p *NodeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_node
}

func (*NodeContext) IsNodeContext() {}

func NewNodeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *NodeContext {
	var p = new(NodeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_node

	return p
}

func (s *NodeContext) GetParser() antlr.Parser { return s.parser }

func (s *NodeContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *NodeContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *NodeContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *NodeContext) Labels() ILabelsContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILabelsContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILabelsContext)
}

func (s *NodeContext) Properties() IPropertiesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertiesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertiesContext)
}

func (s *NodeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NodeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *NodeContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterNode(s)
	}
}

func (s *NodeContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitNode(s)
	}
}

func (p *CypherParser) Node() (localctx INodeContext) {
	localctx = NewNodeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, CypherParserRULE_node)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(173)
		p.Match(CypherParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(175)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserIDENTIFIER {
		{
			p.SetState(174)
			p.Variable()
		}

	}
	p.SetState(178)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserCOLON {
		{
			p.SetState(177)
			p.Labels()
		}

	}
	p.SetState(181)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLCURLY {
		{
			p.SetState(180)
			p.Properties()
		}

	}
	{
		p.SetState(183)
		p.Match(CypherParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRelationshipContext is an interface to support dynamic dispatch.
type IRelationshipContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllMINUS() []antlr.TerminalNode
	MINUS(i int) antlr.TerminalNode
	LSQUARE() antlr.TerminalNode
	RSQUARE() antlr.TerminalNode
	RARROW() antlr.TerminalNode
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode
	Variable() IVariableContext
	Types() ITypesContext
	Range_() IRangeContext
	Properties() IPropertiesContext
	LARROW() antlr.TerminalNode

	// IsRelationshipContext differentiates from other interfaces.
	IsRelationshipContext()
}

type RelationshipContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRelationshipContext() *RelationshipContext {
	var p = new(RelationshipContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_relationship
	return p
}

func InitEmptyRelationshipContext(p *RelationshipContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_relationship
}

func (*RelationshipContext) IsRelationshipContext() {}

func NewRelationshipContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RelationshipContext {
	var p = new(RelationshipContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_relationship

	return p
}

func (s *RelationshipContext) GetParser() antlr.Parser { return s.parser }

func (s *RelationshipContext) AllMINUS() []antlr.TerminalNode {
	return s.GetTokens(CypherParserMINUS)
}

func (s *RelationshipContext) MINUS(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserMINUS, i)
}

func (s *RelationshipContext) LSQUARE() antlr.TerminalNode {
	return s.GetToken(CypherParserLSQUARE, 0)
}

func (s *RelationshipContext) RSQUARE() antlr.TerminalNode {
	return s.GetToken(CypherParserRSQUARE, 0)
}

func (s *RelationshipContext) RARROW() antlr.TerminalNode {
	return s.GetToken(CypherParserRARROW, 0)
}

func (s *RelationshipContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWS)
}

func (s *RelationshipContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWS, i)
}

func (s *RelationshipContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *RelationshipContext) Types() ITypesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypesContext)
}

func (s *RelationshipContext) Range_() IRangeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRangeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRangeContext)
}

func (s *RelationshipContext) Properties() IPropertiesContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertiesContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertiesContext)
}

func (s *RelationshipContext) LARROW() antlr.TerminalNode {
	return s.GetToken(CypherParserLARROW, 0)
}

func (s *RelationshipContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RelationshipContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RelationshipContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterRelationship(s)
	}
}

func (s *RelationshipContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitRelationship(s)
	}
}

func (p *CypherParser) Relationship() (localctx IRelationshipContext) {
	localctx = NewRelationshipContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, CypherParserRULE_relationship)
	var _la int

	p.SetState(269)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 39, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(185)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(189)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(186)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(191)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(192)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(194)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(193)
				p.Variable()
			}

		}
		p.SetState(197)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(196)
				p.Types()
			}

		}
		p.SetState(200)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(199)
				p.Range_()
			}

		}
		p.SetState(203)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(202)
				p.Properties()
			}

		}
		{
			p.SetState(205)
			p.Match(CypherParserRSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(209)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(206)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(211)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(212)
			p.Match(CypherParserRARROW)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(213)
			p.Match(CypherParserLARROW)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(217)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(214)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(219)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(220)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(222)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(221)
				p.Variable()
			}

		}
		p.SetState(225)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(224)
				p.Types()
			}

		}
		p.SetState(228)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(227)
				p.Range_()
			}

		}
		p.SetState(231)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(230)
				p.Properties()
			}

		}
		{
			p.SetState(233)
			p.Match(CypherParserRSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(237)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(234)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(239)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(240)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(241)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(245)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(242)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(247)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(248)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(250)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(249)
				p.Variable()
			}

		}
		p.SetState(253)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(252)
				p.Types()
			}

		}
		p.SetState(256)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(255)
				p.Range_()
			}

		}
		p.SetState(259)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(258)
				p.Properties()
			}

		}
		{
			p.SetState(261)
			p.Match(CypherParserRSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(265)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(262)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(267)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(268)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReturnItemsContext is an interface to support dynamic dispatch.
type IReturnItemsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllReturnItem() []IReturnItemContext
	ReturnItem(i int) IReturnItemContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsReturnItemsContext differentiates from other interfaces.
	IsReturnItemsContext()
}

type ReturnItemsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReturnItemsContext() *ReturnItemsContext {
	var p = new(ReturnItemsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_returnItems
	return p
}

func InitEmptyReturnItemsContext(p *ReturnItemsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_returnItems
}

func (*ReturnItemsContext) IsReturnItemsContext() {}

func NewReturnItemsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ReturnItemsContext {
	var p = new(ReturnItemsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_returnItems

	return p
}

func (s *ReturnItemsContext) GetParser() antlr.Parser { return s.parser }

func (s *ReturnItemsContext) AllReturnItem() []IReturnItemContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IReturnItemContext); ok {
			len++
		}
	}

	tst := make([]IReturnItemContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IReturnItemContext); ok {
			tst[i] = t.(IReturnItemContext)
			i++
		}
	}

	return tst
}

func (s *ReturnItemsContext) ReturnItem(i int) IReturnItemContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReturnItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReturnItemContext)
}

func (s *ReturnItemsContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *ReturnItemsContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *ReturnItemsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ReturnItemsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ReturnItemsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterReturnItems(s)
	}
}

func (s *ReturnItemsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitReturnItems(s)
	}
}

func (p *CypherParser) ReturnItems() (localctx IReturnItemsContext) {
	localctx = NewReturnItemsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, CypherParserRULE_returnItems)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(271)
		p.ReturnItem()
	}
	p.SetState(276)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(272)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(273)
			p.ReturnItem()
		}

		p.SetState(278)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReturnItemContext is an interface to support dynamic dispatch.
type IReturnItemContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AggregateFunc() IAggregateFuncContext
	AS() antlr.TerminalNode
	Variable() IVariableContext
	CaseExpr() ICaseExprContext
	ValueExpr() IValueExprContext
	COALESCE() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	RPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsReturnItemContext differentiates from other interfaces.
	IsReturnItemContext()
}

type ReturnItemContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReturnItemContext() *ReturnItemContext {
	var p = new(ReturnItemContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_returnItem
	return p
}

func InitEmptyReturnItemContext(p *ReturnItemContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_returnItem
}

func (*ReturnItemContext) IsReturnItemContext() {}

func NewReturnItemContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ReturnItemContext {
	var p = new(ReturnItemContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_returnItem

	return p
}

func (s *ReturnItemContext) GetParser() antlr.Parser { return s.parser }

func (s *ReturnItemContext) AggregateFunc() IAggregateFuncContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAggregateFuncContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAggregateFuncContext)
}

func (s *ReturnItemContext) AS() antlr.TerminalNode {
	return s.GetToken(CypherParserAS, 0)
}

func (s *ReturnItemContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *ReturnItemContext) CaseExpr() ICaseExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICaseExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICaseExprContext)
}

func (s *ReturnItemContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *ReturnItemContext) COALESCE() antlr.TerminalNode {
	return s.GetToken(CypherParserCOALESCE, 0)
}

func (s *ReturnItemContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *ReturnItemContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ReturnItemContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ReturnItemContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *ReturnItemContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *ReturnItemContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *ReturnItemContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ReturnItemContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ReturnItemContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterReturnItem(s)
	}
}

func (s *ReturnItemContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitReturnItem(s)
	}
}

func (p *CypherParser) ReturnItem() (localctx IReturnItemContext) {
	localctx = NewReturnItemContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, CypherParserRULE_returnItem)
	var _la int

	p.SetState(309)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserCOUNT, CypherParserSUM, CypherParserAVG, CypherParserMIN, CypherParserMAX:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(279)
			p.AggregateFunc()
		}
		p.SetState(282)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(280)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(281)
				p.Variable()
			}

		}

	case CypherParserCASE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(284)
			p.CaseExpr()
		}
		p.SetState(287)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(285)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(286)
				p.Variable()
			}

		}

	case CypherParserLPAREN, CypherParserLSQUARE, CypherParserMINUS, CypherParserSTRING, CypherParserNUMBER, CypherParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(289)
			p.valueExpr(0)
		}
		p.SetState(292)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(290)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(291)
				p.Variable()
			}

		}

	case CypherParserCOALESCE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(294)
			p.Match(CypherParserCOALESCE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(295)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(296)
			p.Expression()
		}
		p.SetState(301)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserCOMMA {
			{
				p.SetState(297)
				p.Match(CypherParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(298)
				p.Expression()
			}

			p.SetState(303)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(304)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(307)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(305)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(306)
				p.Variable()
			}

		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICaseExprContext is an interface to support dynamic dispatch.
type ICaseExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	CASE() antlr.TerminalNode
	END() antlr.TerminalNode
	AllWHEN() []antlr.TerminalNode
	WHEN(i int) antlr.TerminalNode
	AllCondition() []IConditionContext
	Condition(i int) IConditionContext
	AllTHEN() []antlr.TerminalNode
	THEN(i int) antlr.TerminalNode
	AllCaseResult() []ICaseResultContext
	CaseResult(i int) ICaseResultContext
	ELSE() antlr.TerminalNode

	// IsCaseExprContext differentiates from other interfaces.
	IsCaseExprContext()
}

type CaseExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCaseExprContext() *CaseExprContext {
	var p = new(CaseExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_caseExpr
	return p
}

func InitEmptyCaseExprContext(p *CaseExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_caseExpr
}

func (*CaseExprContext) IsCaseExprContext() {}

func NewCaseExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CaseExprContext {
	var p = new(CaseExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_caseExpr

	return p
}

func (s *CaseExprContext) GetParser() antlr.Parser { return s.parser }

func (s *CaseExprContext) CASE() antlr.TerminalNode {
	return s.GetToken(CypherParserCASE, 0)
}

func (s *CaseExprContext) END() antlr.TerminalNode {
	return s.GetToken(CypherParserEND, 0)
}

func (s *CaseExprContext) AllWHEN() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWHEN)
}

func (s *CaseExprContext) WHEN(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWHEN, i)
}

func (s *CaseExprContext) AllCondition() []IConditionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IConditionContext); ok {
			len++
		}
	}

	tst := make([]IConditionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IConditionContext); ok {
			tst[i] = t.(IConditionContext)
			i++
		}
	}

	return tst
}

func (s *CaseExprContext) Condition(i int) IConditionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *CaseExprContext) AllTHEN() []antlr.TerminalNode {
	return s.GetTokens(CypherParserTHEN)
}

func (s *CaseExprContext) THEN(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserTHEN, i)
}

func (s *CaseExprContext) AllCaseResult() []ICaseResultContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICaseResultContext); ok {
			len++
		}
	}

	tst := make([]ICaseResultContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICaseResultContext); ok {
			tst[i] = t.(ICaseResultContext)
			i++
		}
	}

	return tst
}

func (s *CaseExprContext) CaseResult(i int) ICaseResultContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICaseResultContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICaseResultContext)
}

func (s *CaseExprContext) ELSE() antlr.TerminalNode {
	return s.GetToken(CypherParserELSE, 0)
}

func (s *CaseExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CaseExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CaseExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterCaseExpr(s)
	}
}

func (s *CaseExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitCaseExpr(s)
	}
}

func (p *CypherParser) CaseExpr() (localctx ICaseExprContext) {
	localctx = NewCaseExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, CypherParserRULE_caseExpr)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(311)
		p.Match(CypherParserCASE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(317)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == CypherParserWHEN {
		{
			p.SetState(312)
			p.Match(CypherParserWHEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(313)
			p.condition(0)
		}
		{
			p.SetState(314)
			p.Match(CypherParserTHEN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(315)
			p.CaseResult()
		}

		p.SetState(319)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(323)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserELSE {
		{
			p.SetState(321)
			p.Match(CypherParserELSE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(322)
			p.CaseResult()
		}

	}
	{
		p.SetState(325)
		p.Match(CypherParserEND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICaseResultContext is an interface to support dynamic dispatch.
type ICaseResultContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Value() IValueContext
	Expression() IExpressionContext

	// IsCaseResultContext differentiates from other interfaces.
	IsCaseResultContext()
}

type CaseResultContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCaseResultContext() *CaseResultContext {
	var p = new(CaseResultContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_caseResult
	return p
}

func InitEmptyCaseResultContext(p *CaseResultContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_caseResult
}

func (*CaseResultContext) IsCaseResultContext() {}

func NewCaseResultContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CaseResultContext {
	var p = new(CaseResultContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_caseResult

	return p
}

func (s *CaseResultContext) GetParser() antlr.Parser { return s.parser }

func (s *CaseResultContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *CaseResultContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *CaseResultContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CaseResultContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CaseResultContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterCaseResult(s)
	}
}

func (s *CaseResultContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitCaseResult(s)
	}
}

func (p *CypherParser) CaseResult() (localctx ICaseResultContext) {
	localctx = NewCaseResultContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, CypherParserRULE_caseResult)
	p.SetState(329)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserSTRING, CypherParserNUMBER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(327)
			p.Value()
		}

	case CypherParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(328)
			p.Expression()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IValueExprContext is an interface to support dynamic dispatch.
type IValueExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsValueExprContext differentiates from other interfaces.
	IsValueExprContext()
}

type ValueExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueExprContext() *ValueExprContext {
	var p = new(ValueExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_valueExpr
	return p
}

func InitEmptyValueExprContext(p *ValueExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_valueExpr
}

func (*ValueExprContext) IsValueExprContext() {}

func NewValueExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueExprContext {
	var p = new(ValueExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_valueExpr

	return p
}

func (s *ValueExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueExprContext) CopyAll(ctx *ValueExprContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ValueExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type VeListContext struct {
	ValueExprContext
}

func NewVeListContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VeListContext {
	var p = new(VeListContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VeListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VeListContext) LSQUARE() antlr.TerminalNode {
	return s.GetToken(CypherParserLSQUARE, 0)
}

func (s *VeListContext) RSQUARE() antlr.TerminalNode {
	return s.GetToken(CypherParserRSQUARE, 0)
}

func (s *VeListContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *VeListContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *VeListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *VeListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *VeListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVeList(s)
	}
}

func (s *VeListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVeList(s)
	}
}

type VeMulDivContext struct {
	ValueExprContext
	op antlr.Token
}

func NewVeMulDivContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VeMulDivContext {
	var p = new(VeMulDivContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VeMulDivContext) GetOp() antlr.Token { return s.op }

func (s *VeMulDivContext) SetOp(v antlr.Token) { s.op = v }

func (s *VeMulDivContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VeMulDivContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *VeMulDivContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *VeMulDivContext) STAR() antlr.TerminalNode {
	return s.GetToken(CypherParserSTAR, 0)
}

func (s *VeMulDivContext) SLASH() antlr.TerminalNode {
	return s.GetToken(CypherParserSLASH, 0)
}

func (s *VeMulDivContext) PERCENT() antlr.TerminalNode {
	return s.GetToken(CypherParserPERCENT, 0)
}

func (s *VeMulDivContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVeMulDiv(s)
	}
}

func (s *VeMulDivContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVeMulDiv(s)
	}
}

type VeAddSubContext struct {
	ValueExprContext
	op antlr.Token
}

func NewVeAddSubContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VeAddSubContext {
	var p = new(VeAddSubContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VeAddSubContext) GetOp() antlr.Token { return s.op }

func (s *VeAddSubContext) SetOp(v antlr.Token) { s.op = v }

func (s *VeAddSubContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VeAddSubContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *VeAddSubContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *VeAddSubContext) PLUS() antlr.TerminalNode {
	return s.GetToken(CypherParserPLUS, 0)
}

func (s *VeAddSubContext) MINUS() antlr.TerminalNode {
	return s.GetToken(CypherParserMINUS, 0)
}

func (s *VeAddSubContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVeAddSub(s)
	}
}

func (s *VeAddSubContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVeAddSub(s)
	}
}

type VeLitContext struct {
	ValueExprContext
}

func NewVeLitContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VeLitContext {
	var p = new(VeLitContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VeLitContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VeLitContext) STRING() antlr.TerminalNode {
	return s.GetToken(CypherParserSTRING, 0)
}

func (s *VeLitContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(CypherParserNUMBER, 0)
}

func (s *VeLitContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVeLit(s)
	}
}

func (s *VeLitContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVeLit(s)
	}
}

type VeCallContext struct {
	ValueExprContext
}

func NewVeCallContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VeCallContext {
	var p = new(VeCallContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VeCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VeCallContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, 0)
}

func (s *VeCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *VeCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *VeCallContext) AllValueExpr() []IValueExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueExprContext); ok {
			len++
		}
	}

	tst := make([]IValueExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueExprContext); ok {
			tst[i] = t.(IValueExprContext)
			i++
		}
	}

	return tst
}

func (s *VeCallContext) ValueExpr(i int) IValueExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *VeCallContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *VeCallContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *VeCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVeCall(s)
	}
}

func (s *VeCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVeCall(s)
	}
}

type VeParenContext struct {
	ValueExprContext
}

func NewVeParenContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VeParenContext {
	var p = new(VeParenContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VeParenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VeParenContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *VeParenContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *VeParenContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *VeParenContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVeParen(s)
	}
}

func (s *VeParenContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVeParen(s)
	}
}

type VeNegContext struct {
	ValueExprContext
}

func NewVeNegContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VeNegContext {
	var p = new(VeNegContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VeNegContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VeNegContext) MINUS() antlr.TerminalNode {
	return s.GetToken(CypherParserMINUS, 0)
}

func (s *VeNegContext) ValueExpr() IValueExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueExprContext)
}

func (s *VeNegContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVeNeg(s)
	}
}

func (s *VeNegContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVeNeg(s)
	}
}

type VePropRefContext struct {
	ValueExprContext
}

func NewVePropRefContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VePropRefContext {
	var p = new(VePropRefContext)

	InitEmptyValueExprContext(&p.ValueExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*ValueExprContext))

	return p
}

func (s *VePropRefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VePropRefContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(CypherParserIDENTIFIER)
}

func (s *VePropRefContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, i)
}

func (s *VePropRefContext) AllDOT() []antlr.TerminalNode {
	return s.GetTokens(CypherParserDOT)
}

func (s *VePropRefContext) DOT(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserDOT, i)
}

func (s *VePropRefContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVePropRef(s)
	}
}

func (s *VePropRefContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVePropRef(s)
	}
}

func (p *CypherParser) ValueExpr() (localctx IValueExprContext) {
	return p.valueExpr(0)
}

func (p *CypherParser) valueExpr(_p int) (localctx IValueExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewValueExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IValueExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 32
	p.EnterRecursionRule(localctx, 32, CypherParserRULE_valueExpr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(372)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 55, p.GetParserRuleContext()) {
	case 1:
		localctx = NewVeNegContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(332)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(333)
			p.valueExpr(8)
		}

	case 2:
		localctx = NewVeCallContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(334)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(335)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(344)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64((_la-18)) & ^0x3f) == 0 && ((int64(1)<<(_la-18))&1970324836974661) != 0 {
			{
				p.SetState(336)
				p.valueExpr(0)
			}
			p.SetState(341)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == CypherParserCOMMA {
				{
					p.SetState(337)
					p.Match(CypherParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(338)
					p.valueExpr(0)
				}

				p.SetState(343)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(346)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewVePropRefContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(347)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(352)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 52, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(348)
					p.Match(CypherParserDOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(349)
					p.Match(CypherParserIDENTIFIER)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(354)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 52, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 4:
		localctx = NewVeListContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(355)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(364)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64((_la-18)) & ^0x3f) == 0 && ((int64(1)<<(_la-18))&1970324836974661) != 0 {
			{
				p.SetState(356)
				p.valueExpr(0)
			}
			p.SetState(361)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == CypherParserCOMMA {
				{
					p.SetState(357)
					p.Match(CypherParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(358)
					p.valueExpr(0)
				}

				p.SetState(363)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(366)
			p.Match(CypherParserRSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		localctx = NewVeLitContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(367)
			_la = p.GetTokenStream().LA(1)

			if !(_la == CypherParserSTRING || _la == CypherParserNUMBER) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	case 6:
		localctx = NewVeParenContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(368)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(369)
			p.valueExpr(0)
		}
		{
			p.SetState(370)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(382)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 57, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(380)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 56, p.GetParserRuleContext()) {
			case 1:
				localctx = NewVeMulDivContext(p, NewValueExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_valueExpr)
				p.SetState(374)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(375)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*VeMulDivContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&738197504) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*VeMulDivContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(376)
					p.valueExpr(8)
				}

			case 2:
				localctx = NewVeAddSubContext(p, NewValueExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_valueExpr)
				p.SetState(377)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(378)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*VeAddSubContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == CypherParserMINUS || _la == CypherParserPLUS) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*VeAddSubContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(379)
					p.valueExpr(7)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(384)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 57, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAggregateFuncContext is an interface to support dynamic dispatch.
type IAggregateFuncContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	COUNT() antlr.TerminalNode
	SUM() antlr.TerminalNode
	AVG() antlr.TerminalNode
	MIN() antlr.TerminalNode
	MAX() antlr.TerminalNode
	STAR() antlr.TerminalNode
	AggArg() IAggArgContext
	DISTINCT() antlr.TerminalNode

	// IsAggregateFuncContext differentiates from other interfaces.
	IsAggregateFuncContext()
}

type AggregateFuncContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAggregateFuncContext() *AggregateFuncContext {
	var p = new(AggregateFuncContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_aggregateFunc
	return p
}

func InitEmptyAggregateFuncContext(p *AggregateFuncContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_aggregateFunc
}

func (*AggregateFuncContext) IsAggregateFuncContext() {}

func NewAggregateFuncContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AggregateFuncContext {
	var p = new(AggregateFuncContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_aggregateFunc

	return p
}

func (s *AggregateFuncContext) GetParser() antlr.Parser { return s.parser }

func (s *AggregateFuncContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *AggregateFuncContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *AggregateFuncContext) COUNT() antlr.TerminalNode {
	return s.GetToken(CypherParserCOUNT, 0)
}

func (s *AggregateFuncContext) SUM() antlr.TerminalNode {
	return s.GetToken(CypherParserSUM, 0)
}

func (s *AggregateFuncContext) AVG() antlr.TerminalNode {
	return s.GetToken(CypherParserAVG, 0)
}

func (s *AggregateFuncContext) MIN() antlr.TerminalNode {
	return s.GetToken(CypherParserMIN, 0)
}

func (s *AggregateFuncContext) MAX() antlr.TerminalNode {
	return s.GetToken(CypherParserMAX, 0)
}

func (s *AggregateFuncContext) STAR() antlr.TerminalNode {
	return s.GetToken(CypherParserSTAR, 0)
}

func (s *AggregateFuncContext) AggArg() IAggArgContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAggArgContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAggArgContext)
}

func (s *AggregateFuncContext) DISTINCT() antlr.TerminalNode {
	return s.GetToken(CypherParserDISTINCT, 0)
}

func (s *AggregateFuncContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AggregateFuncContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AggregateFuncContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterAggregateFunc(s)
	}
}

func (s *AggregateFuncContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitAggregateFunc(s)
	}
}

func (p *CypherParser) AggregateFunc() (localctx IAggregateFuncContext) {
	localctx = NewAggregateFuncContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 34, CypherParserRULE_aggregateFunc)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(385)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&4292493394837504) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(386)
		p.Match(CypherParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(392)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserSTAR:
		{
			p.SetState(387)
			p.Match(CypherParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserDISTINCT, CypherParserIDENTIFIER:
		p.SetState(389)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserDISTINCT {
			{
				p.SetState(388)
				p.Match(CypherParserDISTINCT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(391)
			p.AggArg()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(394)
		p.Match(CypherParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAggArgContext is an interface to support dynamic dispatch.
type IAggArgContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	DOT() antlr.TerminalNode

	// IsAggArgContext differentiates from other interfaces.
	IsAggArgContext()
}

type AggArgContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAggArgContext() *AggArgContext {
	var p = new(AggArgContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_aggArg
	return p
}

func InitEmptyAggArgContext(p *AggArgContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_aggArg
}

func (*AggArgContext) IsAggArgContext() {}

func NewAggArgContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AggArgContext {
	var p = new(AggArgContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_aggArg

	return p
}

func (s *AggArgContext) GetParser() antlr.Parser { return s.parser }

func (s *AggArgContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(CypherParserIDENTIFIER)
}

func (s *AggArgContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, i)
}

func (s *AggArgContext) DOT() antlr.TerminalNode {
	return s.GetToken(CypherParserDOT, 0)
}

func (s *AggArgContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AggArgContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AggArgContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterAggArg(s)
	}
}

func (s *AggArgContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitAggArg(s)
	}
}

func (p *CypherParser) AggArg() (localctx IAggArgContext) {
	localctx = NewAggArgContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 36, CypherParserRULE_aggArg)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(396)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(399)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDOT {
		{
			p.SetState(397)
			p.Match(CypherParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(398)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrderItemsContext is an interface to support dynamic dispatch.
type IOrderItemsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ORDER_BY() antlr.TerminalNode
	AllOrderItem() []IOrderItemContext
	OrderItem(i int) IOrderItemContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsOrderItemsContext differentiates from other interfaces.
	IsOrderItemsContext()
}

type OrderItemsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrderItemsContext() *OrderItemsContext {
	var p = new(OrderItemsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_orderItems
	return p
}

func InitEmptyOrderItemsContext(p *OrderItemsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_orderItems
}

func (*OrderItemsContext) IsOrderItemsContext() {}

func NewOrderItemsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrderItemsContext {
	var p = new(OrderItemsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_orderItems

	return p
}

func (s *OrderItemsContext) GetParser() antlr.Parser { return s.parser }

func (s *OrderItemsContext) ORDER_BY() antlr.TerminalNode {
	return s.GetToken(CypherParserORDER_BY, 0)
}

func (s *OrderItemsContext) AllOrderItem() []IOrderItemContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IOrderItemContext); ok {
			len++
		}
	}

	tst := make([]IOrderItemContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IOrderItemContext); ok {
			tst[i] = t.(IOrderItemContext)
			i++
		}
	}

	return tst
}

func (s *OrderItemsContext) OrderItem(i int) IOrderItemContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOrderItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOrderItemContext)
}

func (s *OrderItemsContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *OrderItemsContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *OrderItemsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrderItemsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrderItemsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterOrderItems(s)
	}
}

func (s *OrderItemsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitOrderItems(s)
	}
}

func (p *CypherParser) OrderItems() (localctx IOrderItemsContext) {
	localctx = NewOrderItemsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 38, CypherParserRULE_orderItems)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(401)
		p.Match(CypherParserORDER_BY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(402)
		p.OrderItem()
	}
	p.SetState(407)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(403)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(404)
			p.OrderItem()
		}

		p.SetState(409)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOrderItemContext is an interface to support dynamic dispatch.
type IOrderItemContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext
	ASC() antlr.TerminalNode
	DESC() antlr.TerminalNode

	// IsOrderItemContext differentiates from other interfaces.
	IsOrderItemContext()
}

type OrderItemContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOrderItemContext() *OrderItemContext {
	var p = new(OrderItemContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_orderItem
	return p
}

func InitEmptyOrderItemContext(p *OrderItemContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_orderItem
}

func (*OrderItemContext) IsOrderItemContext() {}

func NewOrderItemContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OrderItemContext {
	var p = new(OrderItemContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_orderItem

	return p
}

func (s *OrderItemContext) GetParser() antlr.Parser { return s.parser }

func (s *OrderItemContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *OrderItemContext) ASC() antlr.TerminalNode {
	return s.GetToken(CypherParserASC, 0)
}

func (s *OrderItemContext) DESC() antlr.TerminalNode {
	return s.GetToken(CypherParserDESC, 0)
}

func (s *OrderItemContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OrderItemContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OrderItemContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterOrderItem(s)
	}
}

func (s *OrderItemContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitOrderItem(s)
	}
}

func (p *CypherParser) OrderItem() (localctx IOrderItemContext) {
	localctx = NewOrderItemContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 40, CypherParserRULE_orderItem)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(410)
		p.Expression()
	}
	{
		p.SetState(411)
		_la = p.GetTokenStream().LA(1)

		if !(_la == CypherParserASC || _la == CypherParserDESC) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILimitNumContext is an interface to support dynamic dispatch.
type ILimitNumContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LIMIT() antlr.TerminalNode
	NUMBER() antlr.TerminalNode

	// IsLimitNumContext differentiates from other interfaces.
	IsLimitNumContext()
}

type LimitNumContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLimitNumContext() *LimitNumContext {
	var p = new(LimitNumContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_limitNum
	return p
}

func InitEmptyLimitNumContext(p *LimitNumContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_limitNum
}

func (*LimitNumContext) IsLimitNumContext() {}

func NewLimitNumContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LimitNumContext {
	var p = new(LimitNumContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_limitNum

	return p
}

func (s *LimitNumContext) GetParser() antlr.Parser { return s.parser }

func (s *LimitNumContext) LIMIT() antlr.TerminalNode {
	return s.GetToken(CypherParserLIMIT, 0)
}

func (s *LimitNumContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(CypherParserNUMBER, 0)
}

func (s *LimitNumContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LimitNumContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LimitNumContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterLimitNum(s)
	}
}

func (s *LimitNumContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitLimitNum(s)
	}
}

func (p *CypherParser) LimitNum() (localctx ILimitNumContext) {
	localctx = NewLimitNumContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 42, CypherParserRULE_limitNum)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(413)
		p.Match(CypherParserLIMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(414)
		p.Match(CypherParserNUMBER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILabelsContext is an interface to support dynamic dispatch.
type ILabelsContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	COLON() antlr.TerminalNode
	AllLabel() []ILabelContext
	Label(i int) ILabelContext

	// IsLabelsContext differentiates from other interfaces.
	IsLabelsContext()
}

type LabelsContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLabelsContext() *LabelsContext {
	var p = new(LabelsContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_labels
	return p
}

func InitEmptyLabelsContext(p *LabelsContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_labels
}

func (*LabelsContext) IsLabelsContext() {}

func NewLabelsContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LabelsContext {
	var p = new(LabelsContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_labels

	return p
}

func (s *LabelsContext) GetParser() antlr.Parser { return s.parser }

func (s *LabelsContext) COLON() antlr.TerminalNode {
	return s.GetToken(CypherParserCOLON, 0)
}

func (s *LabelsContext) AllLabel() []ILabelContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ILabelContext); ok {
			len++
		}
	}

	tst := make([]ILabelContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ILabelContext); ok {
			tst[i] = t.(ILabelContext)
			i++
		}
	}

	return tst
}

func (s *LabelsContext) Label(i int) ILabelContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILabelContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILabelContext)
}

func (s *LabelsContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LabelsContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LabelsContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterLabels(s)
	}
}

func (s *LabelsContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitLabels(s)
	}
}

func (p *CypherParser) Labels() (localctx ILabelsContext) {
	localctx = NewLabelsContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 44, CypherParserRULE_labels)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(416)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(418)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == CypherParserIDENTIFIER {
		{
			p.SetState(417)
			p.Label()
		}

		p.SetState(420)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILabelContext is an interface to support dynamic dispatch.
type ILabelContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsLabelContext differentiates from other interfaces.
	IsLabelContext()
}

type LabelContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLabelContext() *LabelContext {
	var p = new(LabelContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_label
	return p
}

func InitEmptyLabelContext(p *LabelContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_label
}

func (*LabelContext) IsLabelContext() {}

func NewLabelContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LabelContext {
	var p = new(LabelContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_label

	return p
}

func (s *LabelContext) GetParser() antlr.Parser { return s.parser }

func (s *LabelContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, 0)
}

func (s *LabelContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LabelContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LabelContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterLabel(s)
	}
}

func (s *LabelContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitLabel(s)
	}
}

func (p *CypherParser) Label() (localctx ILabelContext) {
	localctx = NewLabelContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 46, CypherParserRULE_label)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(422)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPropertiesContext is an interface to support dynamic dispatch.
type IPropertiesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LCURLY() antlr.TerminalNode
	AllProperty() []IPropertyContext
	Property(i int) IPropertyContext
	RCURLY() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsPropertiesContext differentiates from other interfaces.
	IsPropertiesContext()
}

type PropertiesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPropertiesContext() *PropertiesContext {
	var p = new(PropertiesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_properties
	return p
}

func InitEmptyPropertiesContext(p *PropertiesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_properties
}

func (*PropertiesContext) IsPropertiesContext() {}

func NewPropertiesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PropertiesContext {
	var p = new(PropertiesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_properties

	return p
}

func (s *PropertiesContext) GetParser() antlr.Parser { return s.parser }

func (s *PropertiesContext) LCURLY() antlr.TerminalNode {
	return s.GetToken(CypherParserLCURLY, 0)
}

func (s *PropertiesContext) AllProperty() []IPropertyContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPropertyContext); ok {
			len++
		}
	}

	tst := make([]IPropertyContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPropertyContext); ok {
			tst[i] = t.(IPropertyContext)
			i++
		}
	}

	return tst
}

func (s *PropertiesContext) Property(i int) IPropertyContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPropertyContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPropertyContext)
}

func (s *PropertiesContext) RCURLY() antlr.TerminalNode {
	return s.GetToken(CypherParserRCURLY, 0)
}

func (s *PropertiesContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *PropertiesContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *PropertiesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PropertiesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PropertiesContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterProperties(s)
	}
}

func (s *PropertiesContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitProperties(s)
	}
}

func (p *CypherParser) Properties() (localctx IPropertiesContext) {
	localctx = NewPropertiesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 48, CypherParserRULE_properties)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(424)
		p.Match(CypherParserLCURLY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(425)
		p.Property()
	}
	p.SetState(430)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(426)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(427)
			p.Property()
		}

		p.SetState(432)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(433)
		p.Match(CypherParserRCURLY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPropertyContext is an interface to support dynamic dispatch.
type IPropertyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode
	COLON() antlr.TerminalNode
	Value() IValueContext

	// IsPropertyContext differentiates from other interfaces.
	IsPropertyContext()
}

type PropertyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPropertyContext() *PropertyContext {
	var p = new(PropertyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_property
	return p
}

func InitEmptyPropertyContext(p *PropertyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_property
}

func (*PropertyContext) IsPropertyContext() {}

func NewPropertyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PropertyContext {
	var p = new(PropertyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_property

	return p
}

func (s *PropertyContext) GetParser() antlr.Parser { return s.parser }

func (s *PropertyContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, 0)
}

func (s *PropertyContext) COLON() antlr.TerminalNode {
	return s.GetToken(CypherParserCOLON, 0)
}

func (s *PropertyContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *PropertyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PropertyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PropertyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterProperty(s)
	}
}

func (s *PropertyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitProperty(s)
	}
}

func (p *CypherParser) Property() (localctx IPropertyContext) {
	localctx = NewPropertyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 50, CypherParserRULE_property)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(435)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(436)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(437)
		p.Value()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IConditionContext is an interface to support dynamic dispatch.
type IConditionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsConditionContext differentiates from other interfaces.
	IsConditionContext()
}

type ConditionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConditionContext() *ConditionContext {
	var p = new(ConditionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_condition
	return p
}

func InitEmptyConditionContext(p *ConditionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_condition
}

func (*ConditionContext) IsConditionContext() {}

func NewConditionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConditionContext {
	var p = new(ConditionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_condition

	return p
}

func (s *ConditionContext) GetParser() antlr.Parser { return s.parser }

func (s *ConditionContext) CopyAll(ctx *ConditionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ConditionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ConditionAndContext struct {
	ConditionContext
}

func NewConditionAndContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionAndContext {
	var p = new(ConditionAndContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionAndContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionAndContext) AllCondition() []IConditionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IConditionContext); ok {
			len++
		}
	}

	tst := make([]IConditionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IConditionContext); ok {
			tst[i] = t.(IConditionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionAndContext) Condition(i int) IConditionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionAndContext) AND() antlr.TerminalNode {
	return s.GetToken(CypherParserAND, 0)
}

func (s *ConditionAndContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionAnd(s)
	}
}

func (s *ConditionAndContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionAnd(s)
	}
}

type ConditionNotContext struct {
	ConditionContext
}

func NewConditionNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionNotContext {
	var p = new(ConditionNotContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionNotContext) NOT() antlr.TerminalNode {
	return s.GetToken(CypherParserNOT, 0)
}

func (s *ConditionNotContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionNotContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionNot(s)
	}
}

func (s *ConditionNotContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionNot(s)
	}
}

type ConditionVarNeqContext struct {
	ConditionContext
}

func NewConditionVarNeqContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionVarNeqContext {
	var p = new(ConditionVarNeqContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionVarNeqContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionVarNeqContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionVarNeqContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionVarNeqContext) NEQ() antlr.TerminalNode {
	return s.GetToken(CypherParserNEQ, 0)
}

func (s *ConditionVarNeqContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionVarNeq(s)
	}
}

func (s *ConditionVarNeqContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionVarNeq(s)
	}
}

type ConditionVarGreaterEqualContext struct {
	ConditionContext
}

func NewConditionVarGreaterEqualContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionVarGreaterEqualContext {
	var p = new(ConditionVarGreaterEqualContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionVarGreaterEqualContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionVarGreaterEqualContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionVarGreaterEqualContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionVarGreaterEqualContext) GE() antlr.TerminalNode {
	return s.GetToken(CypherParserGE, 0)
}

func (s *ConditionVarGreaterEqualContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionVarGreaterEqual(s)
	}
}

func (s *ConditionVarGreaterEqualContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionVarGreaterEqual(s)
	}
}

type ConditionIsNullContext struct {
	ConditionContext
}

func NewConditionIsNullContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionIsNullContext {
	var p = new(ConditionIsNullContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionIsNullContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionIsNullContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionIsNullContext) IS() antlr.TerminalNode {
	return s.GetToken(CypherParserIS, 0)
}

func (s *ConditionIsNullContext) NULL() antlr.TerminalNode {
	return s.GetToken(CypherParserNULL, 0)
}

func (s *ConditionIsNullContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionIsNull(s)
	}
}

func (s *ConditionIsNullContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionIsNull(s)
	}
}

type ConditionGreaterContext struct {
	ConditionContext
}

func NewConditionGreaterContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionGreaterContext {
	var p = new(ConditionGreaterContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionGreaterContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionGreaterContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionGreaterContext) RANGLE() antlr.TerminalNode {
	return s.GetToken(CypherParserRANGLE, 0)
}

func (s *ConditionGreaterContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ConditionGreaterContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionGreater(s)
	}
}

func (s *ConditionGreaterContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionGreater(s)
	}
}

type ConditionInContext struct {
	ConditionContext
}

func NewConditionInContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionInContext {
	var p = new(ConditionInContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionInContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionInContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionInContext) IN() antlr.TerminalNode {
	return s.GetToken(CypherParserIN, 0)
}

func (s *ConditionInContext) List() IListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IListContext)
}

func (s *ConditionInContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionIn(s)
	}
}

func (s *ConditionInContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionIn(s)
	}
}

type ConditionAnyContext struct {
	ConditionContext
}

func NewConditionAnyContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionAnyContext {
	var p = new(ConditionAnyContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionAnyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionAnyContext) ANY() antlr.TerminalNode {
	return s.GetToken(CypherParserANY, 0)
}

func (s *ConditionAnyContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *ConditionAnyContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *ConditionAnyContext) IN() antlr.TerminalNode {
	return s.GetToken(CypherParserIN, 0)
}

func (s *ConditionAnyContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionAnyContext) WHERE() antlr.TerminalNode {
	return s.GetToken(CypherParserWHERE, 0)
}

func (s *ConditionAnyContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionAnyContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *ConditionAnyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionAny(s)
	}
}

func (s *ConditionAnyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionAny(s)
	}
}

type ConditionVarEqContext struct {
	ConditionContext
}

func NewConditionVarEqContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionVarEqContext {
	var p = new(ConditionVarEqContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionVarEqContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionVarEqContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionVarEqContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionVarEqContext) EQ() antlr.TerminalNode {
	return s.GetToken(CypherParserEQ, 0)
}

func (s *ConditionVarEqContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionVarEq(s)
	}
}

func (s *ConditionVarEqContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionVarEq(s)
	}
}

type ConditionVarLessEqualContext struct {
	ConditionContext
}

func NewConditionVarLessEqualContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionVarLessEqualContext {
	var p = new(ConditionVarLessEqualContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionVarLessEqualContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionVarLessEqualContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionVarLessEqualContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionVarLessEqualContext) LE() antlr.TerminalNode {
	return s.GetToken(CypherParserLE, 0)
}

func (s *ConditionVarLessEqualContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionVarLessEqual(s)
	}
}

func (s *ConditionVarLessEqualContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionVarLessEqual(s)
	}
}

type ConditionOrContext struct {
	ConditionContext
}

func NewConditionOrContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionOrContext {
	var p = new(ConditionOrContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionOrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionOrContext) AllCondition() []IConditionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IConditionContext); ok {
			len++
		}
	}

	tst := make([]IConditionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IConditionContext); ok {
			tst[i] = t.(IConditionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionOrContext) Condition(i int) IConditionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionOrContext) OR() antlr.TerminalNode {
	return s.GetToken(CypherParserOR, 0)
}

func (s *ConditionOrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionOr(s)
	}
}

func (s *ConditionOrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionOr(s)
	}
}

type ConditionIsNotNullContext struct {
	ConditionContext
}

func NewConditionIsNotNullContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionIsNotNullContext {
	var p = new(ConditionIsNotNullContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionIsNotNullContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionIsNotNullContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionIsNotNullContext) IS() antlr.TerminalNode {
	return s.GetToken(CypherParserIS, 0)
}

func (s *ConditionIsNotNullContext) NOT() antlr.TerminalNode {
	return s.GetToken(CypherParserNOT, 0)
}

func (s *ConditionIsNotNullContext) NULL() antlr.TerminalNode {
	return s.GetToken(CypherParserNULL, 0)
}

func (s *ConditionIsNotNullContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionIsNotNull(s)
	}
}

func (s *ConditionIsNotNullContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionIsNotNull(s)
	}
}

type ConditionParenContext struct {
	ConditionContext
}

func NewConditionParenContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionParenContext {
	var p = new(ConditionParenContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionParenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionParenContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *ConditionParenContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionParenContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *ConditionParenContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionParen(s)
	}
}

func (s *ConditionParenContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionParen(s)
	}
}

type ConditionLessEqualContext struct {
	ConditionContext
}

func NewConditionLessEqualContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionLessEqualContext {
	var p = new(ConditionLessEqualContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionLessEqualContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionLessEqualContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionLessEqualContext) LE() antlr.TerminalNode {
	return s.GetToken(CypherParserLE, 0)
}

func (s *ConditionLessEqualContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ConditionLessEqualContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionLessEqual(s)
	}
}

func (s *ConditionLessEqualContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionLessEqual(s)
	}
}

type ConditionNoneContext struct {
	ConditionContext
}

func NewConditionNoneContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionNoneContext {
	var p = new(ConditionNoneContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionNoneContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionNoneContext) NONE() antlr.TerminalNode {
	return s.GetToken(CypherParserNONE, 0)
}

func (s *ConditionNoneContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *ConditionNoneContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *ConditionNoneContext) IN() antlr.TerminalNode {
	return s.GetToken(CypherParserIN, 0)
}

func (s *ConditionNoneContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionNoneContext) WHERE() antlr.TerminalNode {
	return s.GetToken(CypherParserWHERE, 0)
}

func (s *ConditionNoneContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionNoneContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *ConditionNoneContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionNone(s)
	}
}

func (s *ConditionNoneContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionNone(s)
	}
}

type ConditionGreaterEqualContext struct {
	ConditionContext
}

func NewConditionGreaterEqualContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionGreaterEqualContext {
	var p = new(ConditionGreaterEqualContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionGreaterEqualContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionGreaterEqualContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionGreaterEqualContext) GE() antlr.TerminalNode {
	return s.GetToken(CypherParserGE, 0)
}

func (s *ConditionGreaterEqualContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ConditionGreaterEqualContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionGreaterEqual(s)
	}
}

func (s *ConditionGreaterEqualContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionGreaterEqual(s)
	}
}

type ConditionVarGreaterContext struct {
	ConditionContext
}

func NewConditionVarGreaterContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionVarGreaterContext {
	var p = new(ConditionVarGreaterContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionVarGreaterContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionVarGreaterContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionVarGreaterContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionVarGreaterContext) RANGLE() antlr.TerminalNode {
	return s.GetToken(CypherParserRANGLE, 0)
}

func (s *ConditionVarGreaterContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionVarGreater(s)
	}
}

func (s *ConditionVarGreaterContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionVarGreater(s)
	}
}

type ConditionAllContext struct {
	ConditionContext
}

func NewConditionAllContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionAllContext {
	var p = new(ConditionAllContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionAllContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionAllContext) ALL() antlr.TerminalNode {
	return s.GetToken(CypherParserALL, 0)
}

func (s *ConditionAllContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *ConditionAllContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *ConditionAllContext) IN() antlr.TerminalNode {
	return s.GetToken(CypherParserIN, 0)
}

func (s *ConditionAllContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionAllContext) WHERE() antlr.TerminalNode {
	return s.GetToken(CypherParserWHERE, 0)
}

func (s *ConditionAllContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionAllContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *ConditionAllContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionAll(s)
	}
}

func (s *ConditionAllContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionAll(s)
	}
}

type ConditionVarLessContext struct {
	ConditionContext
}

func NewConditionVarLessContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionVarLessContext {
	var p = new(ConditionVarLessContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionVarLessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionVarLessContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ConditionVarLessContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionVarLessContext) LANGLE() antlr.TerminalNode {
	return s.GetToken(CypherParserLANGLE, 0)
}

func (s *ConditionVarLessContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionVarLess(s)
	}
}

func (s *ConditionVarLessContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionVarLess(s)
	}
}

type ConditionNotEqualityContext struct {
	ConditionContext
}

func NewConditionNotEqualityContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionNotEqualityContext {
	var p = new(ConditionNotEqualityContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionNotEqualityContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionNotEqualityContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionNotEqualityContext) NEQ() antlr.TerminalNode {
	return s.GetToken(CypherParserNEQ, 0)
}

func (s *ConditionNotEqualityContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ConditionNotEqualityContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionNotEquality(s)
	}
}

func (s *ConditionNotEqualityContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionNotEquality(s)
	}
}

type ConditionLessContext struct {
	ConditionContext
}

func NewConditionLessContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionLessContext {
	var p = new(ConditionLessContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionLessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionLessContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionLessContext) LANGLE() antlr.TerminalNode {
	return s.GetToken(CypherParserLANGLE, 0)
}

func (s *ConditionLessContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ConditionLessContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionLess(s)
	}
}

func (s *ConditionLessContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionLess(s)
	}
}

type ConditionSingleContext struct {
	ConditionContext
}

func NewConditionSingleContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionSingleContext {
	var p = new(ConditionSingleContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionSingleContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionSingleContext) SINGLE() antlr.TerminalNode {
	return s.GetToken(CypherParserSINGLE, 0)
}

func (s *ConditionSingleContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *ConditionSingleContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *ConditionSingleContext) IN() antlr.TerminalNode {
	return s.GetToken(CypherParserIN, 0)
}

func (s *ConditionSingleContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionSingleContext) WHERE() antlr.TerminalNode {
	return s.GetToken(CypherParserWHERE, 0)
}

func (s *ConditionSingleContext) Condition() IConditionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IConditionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ConditionSingleContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *ConditionSingleContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionSingle(s)
	}
}

func (s *ConditionSingleContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionSingle(s)
	}
}

type ConditionEqualityContext struct {
	ConditionContext
}

func NewConditionEqualityContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionEqualityContext {
	var p = new(ConditionEqualityContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionEqualityContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionEqualityContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConditionEqualityContext) EQ() antlr.TerminalNode {
	return s.GetToken(CypherParserEQ, 0)
}

func (s *ConditionEqualityContext) Value() IValueContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ConditionEqualityContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionEquality(s)
	}
}

func (s *ConditionEqualityContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionEquality(s)
	}
}

func (p *CypherParser) Condition() (localctx IConditionContext) {
	return p.condition(0)
}

func (p *CypherParser) condition(_p int) (localctx IConditionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewConditionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IConditionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 52
	p.EnterRecursionRule(localctx, 52, CypherParserRULE_condition, _p)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(543)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 64, p.GetParserRuleContext()) {
	case 1:
		localctx = NewConditionParenContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(440)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(441)
			p.condition(0)
		}
		{
			p.SetState(442)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewConditionNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(444)
			p.Match(CypherParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(445)
			p.condition(22)
		}

	case 3:
		localctx = NewConditionAllContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(446)
			p.Match(CypherParserALL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(447)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(448)
			p.Variable()
		}
		{
			p.SetState(449)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(450)
			p.Expression()
		}
		{
			p.SetState(451)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(452)
			p.condition(0)
		}
		{
			p.SetState(453)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 4:
		localctx = NewConditionAnyContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(455)
			p.Match(CypherParserANY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(456)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(457)
			p.Variable()
		}
		{
			p.SetState(458)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(459)
			p.Expression()
		}
		{
			p.SetState(460)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(461)
			p.condition(0)
		}
		{
			p.SetState(462)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 5:
		localctx = NewConditionNoneContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(464)
			p.Match(CypherParserNONE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(465)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(466)
			p.Variable()
		}
		{
			p.SetState(467)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(468)
			p.Expression()
		}
		{
			p.SetState(469)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(470)
			p.condition(0)
		}
		{
			p.SetState(471)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewConditionSingleContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(473)
			p.Match(CypherParserSINGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(474)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(475)
			p.Variable()
		}
		{
			p.SetState(476)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(477)
			p.Expression()
		}
		{
			p.SetState(478)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(479)
			p.condition(0)
		}
		{
			p.SetState(480)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewConditionIsNullContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(482)
			p.Expression()
		}
		{
			p.SetState(483)
			p.Match(CypherParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(484)
			p.Match(CypherParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 8:
		localctx = NewConditionIsNotNullContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(486)
			p.Expression()
		}
		{
			p.SetState(487)
			p.Match(CypherParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(488)
			p.Match(CypherParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(489)
			p.Match(CypherParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 9:
		localctx = NewConditionInContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(491)
			p.Expression()
		}
		{
			p.SetState(492)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(493)
			p.List()
		}

	case 10:
		localctx = NewConditionEqualityContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(495)
			p.Expression()
		}
		{
			p.SetState(496)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(497)
			p.Value()
		}

	case 11:
		localctx = NewConditionNotEqualityContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(499)
			p.Expression()
		}
		{
			p.SetState(500)
			p.Match(CypherParserNEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(501)
			p.Value()
		}

	case 12:
		localctx = NewConditionGreaterContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(503)
			p.Expression()
		}
		{
			p.SetState(504)
			p.Match(CypherParserRANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(505)
			p.Value()
		}

	case 13:
		localctx = NewConditionLessContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(507)
			p.Expression()
		}
		{
			p.SetState(508)
			p.Match(CypherParserLANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(509)
			p.Value()
		}

	case 14:
		localctx = NewConditionGreaterEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(511)
			p.Expression()
		}
		{
			p.SetState(512)
			p.Match(CypherParserGE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(513)
			p.Value()
		}

	case 15:
		localctx = NewConditionLessEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(515)
			p.Expression()
		}
		{
			p.SetState(516)
			p.Match(CypherParserLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(517)
			p.Value()
		}

	case 16:
		localctx = NewConditionVarEqContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(519)
			p.Expression()
		}
		{
			p.SetState(520)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(521)
			p.Expression()
		}

	case 17:
		localctx = NewConditionVarNeqContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(523)
			p.Expression()
		}
		{
			p.SetState(524)
			p.Match(CypherParserNEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(525)
			p.Expression()
		}

	case 18:
		localctx = NewConditionVarGreaterContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(527)
			p.Expression()
		}
		{
			p.SetState(528)
			p.Match(CypherParserRANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(529)
			p.Expression()
		}

	case 19:
		localctx = NewConditionVarLessContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(531)
			p.Expression()
		}
		{
			p.SetState(532)
			p.Match(CypherParserLANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(533)
			p.Expression()
		}

	case 20:
		localctx = NewConditionVarGreaterEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(535)
			p.Expression()
		}
		{
			p.SetState(536)
			p.Match(CypherParserGE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(537)
			p.Expression()
		}

	case 21:
		localctx = NewConditionVarLessEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(539)
			p.Expression()
		}
		{
			p.SetState(540)
			p.Match(CypherParserLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(541)
			p.Expression()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(553)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 66, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(551)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 65, p.GetParserRuleContext()) {
			case 1:
				localctx = NewConditionAndContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_condition)
				p.SetState(545)

				if !(p.Precpred(p.GetParserRuleContext(), 21)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 21)", ""))
					goto errorExit
				}
				{
					p.SetState(546)
					p.Match(CypherParserAND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(547)
					p.condition(22)
				}

			case 2:
				localctx = NewConditionOrContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_condition)
				p.SetState(548)

				if !(p.Precpred(p.GetParserRuleContext(), 20)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 20)", ""))
					goto errorExit
				}
				{
					p.SetState(549)
					p.Match(CypherParserOR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(550)
					p.condition(21)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(555)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 66, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IListContext is an interface to support dynamic dispatch.
type IListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LSQUARE() antlr.TerminalNode
	AllValue() []IValueContext
	Value(i int) IValueContext
	RSQUARE() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsListContext differentiates from other interfaces.
	IsListContext()
}

type ListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyListContext() *ListContext {
	var p = new(ListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_list
	return p
}

func InitEmptyListContext(p *ListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_list
}

func (*ListContext) IsListContext() {}

func NewListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ListContext {
	var p = new(ListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_list

	return p
}

func (s *ListContext) GetParser() antlr.Parser { return s.parser }

func (s *ListContext) LSQUARE() antlr.TerminalNode {
	return s.GetToken(CypherParserLSQUARE, 0)
}

func (s *ListContext) AllValue() []IValueContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IValueContext); ok {
			len++
		}
	}

	tst := make([]IValueContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IValueContext); ok {
			tst[i] = t.(IValueContext)
			i++
		}
	}

	return tst
}

func (s *ListContext) Value(i int) IValueContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IValueContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IValueContext)
}

func (s *ListContext) RSQUARE() antlr.TerminalNode {
	return s.GetToken(CypherParserRSQUARE, 0)
}

func (s *ListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(CypherParserCOMMA)
}

func (s *ListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserCOMMA, i)
}

func (s *ListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterList(s)
	}
}

func (s *ListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitList(s)
	}
}

func (p *CypherParser) List() (localctx IListContext) {
	localctx = NewListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 54, CypherParserRULE_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(556)
		p.Match(CypherParserLSQUARE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(557)
		p.Value()
	}
	p.SetState(562)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(558)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(559)
			p.Value()
		}

		p.SetState(564)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(565)
		p.Match(CypherParserRSQUARE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IVariableContext is an interface to support dynamic dispatch.
type IVariableContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsVariableContext differentiates from other interfaces.
	IsVariableContext()
}

type VariableContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVariableContext() *VariableContext {
	var p = new(VariableContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_variable
	return p
}

func InitEmptyVariableContext(p *VariableContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_variable
}

func (*VariableContext) IsVariableContext() {}

func NewVariableContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VariableContext {
	var p = new(VariableContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_variable

	return p
}

func (s *VariableContext) GetParser() antlr.Parser { return s.parser }

func (s *VariableContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, 0)
}

func (s *VariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *VariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterVariable(s)
	}
}

func (s *VariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitVariable(s)
	}
}

func (p *CypherParser) Variable() (localctx IVariableContext) {
	localctx = NewVariableContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 56, CypherParserRULE_variable)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(567)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITypesContext is an interface to support dynamic dispatch.
type ITypesContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	COLON() antlr.TerminalNode
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode

	// IsTypesContext differentiates from other interfaces.
	IsTypesContext()
}

type TypesContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTypesContext() *TypesContext {
	var p = new(TypesContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_types
	return p
}

func InitEmptyTypesContext(p *TypesContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_types
}

func (*TypesContext) IsTypesContext() {}

func NewTypesContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TypesContext {
	var p = new(TypesContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_types

	return p
}

func (s *TypesContext) GetParser() antlr.Parser { return s.parser }

func (s *TypesContext) COLON() antlr.TerminalNode {
	return s.GetToken(CypherParserCOLON, 0)
}

func (s *TypesContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(CypherParserIDENTIFIER)
}

func (s *TypesContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, i)
}

func (s *TypesContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TypesContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TypesContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterTypes(s)
	}
}

func (s *TypesContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitTypes(s)
	}
}

func (p *CypherParser) Types() (localctx ITypesContext) {
	localctx = NewTypesContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 58, CypherParserRULE_types)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(569)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(570)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(575)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserT__0 {
		{
			p.SetState(571)
			p.Match(CypherParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(572)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(577)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	AllDOT() []antlr.TerminalNode
	DOT(i int) antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	Variable() IVariableContext
	RPAREN() antlr.TerminalNode

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(CypherParserIDENTIFIER)
}

func (s *ExpressionContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserIDENTIFIER, i)
}

func (s *ExpressionContext) AllDOT() []antlr.TerminalNode {
	return s.GetTokens(CypherParserDOT)
}

func (s *ExpressionContext) DOT(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserDOT, i)
}

func (s *ExpressionContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
}

func (s *ExpressionContext) Variable() IVariableContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IVariableContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IVariableContext)
}

func (s *ExpressionContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserRPAREN, 0)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitExpression(s)
	}
}

func (p *CypherParser) Expression() (localctx IExpressionContext) {
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 60, CypherParserRULE_expression)
	var _alt int

	p.SetState(591)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 70, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(578)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(583)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 69, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(579)
					p.Match(CypherParserDOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(580)
					p.Match(CypherParserIDENTIFIER)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(585)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 69, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(586)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(587)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(588)
			p.Variable()
		}
		{
			p.SetState(589)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IValueContext is an interface to support dynamic dispatch.
type IValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	NUMBER() antlr.TerminalNode

	// IsValueContext differentiates from other interfaces.
	IsValueContext()
}

type ValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyValueContext() *ValueContext {
	var p = new(ValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_value
	return p
}

func InitEmptyValueContext(p *ValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_value
}

func (*ValueContext) IsValueContext() {}

func NewValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ValueContext {
	var p = new(ValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_value

	return p
}

func (s *ValueContext) GetParser() antlr.Parser { return s.parser }

func (s *ValueContext) STRING() antlr.TerminalNode {
	return s.GetToken(CypherParserSTRING, 0)
}

func (s *ValueContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(CypherParserNUMBER, 0)
}

func (s *ValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterValue(s)
	}
}

func (s *ValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitValue(s)
	}
}

func (p *CypherParser) Value() (localctx IValueContext) {
	localctx = NewValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 62, CypherParserRULE_value)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(593)
		_la = p.GetTokenStream().LA(1)

		if !(_la == CypherParserSTRING || _la == CypherParserNUMBER) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRangeContext is an interface to support dynamic dispatch.
type IRangeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STAR() antlr.TerminalNode
	RangeLiteral() IRangeLiteralContext

	// IsRangeContext differentiates from other interfaces.
	IsRangeContext()
}

type RangeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRangeContext() *RangeContext {
	var p = new(RangeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_range
	return p
}

func InitEmptyRangeContext(p *RangeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_range
}

func (*RangeContext) IsRangeContext() {}

func NewRangeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RangeContext {
	var p = new(RangeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_range

	return p
}

func (s *RangeContext) GetParser() antlr.Parser { return s.parser }

func (s *RangeContext) STAR() antlr.TerminalNode {
	return s.GetToken(CypherParserSTAR, 0)
}

func (s *RangeContext) RangeLiteral() IRangeLiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRangeLiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRangeLiteralContext)
}

func (s *RangeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RangeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RangeContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterRange(s)
	}
}

func (s *RangeContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitRange(s)
	}
}

func (p *CypherParser) Range_() (localctx IRangeContext) {
	localctx = NewRangeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 64, CypherParserRULE_range)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(595)
		p.Match(CypherParserSTAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(597)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDOUBLE_DOT || _la == CypherParserNUMBER {
		{
			p.SetState(596)
			p.RangeLiteral()
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRangeLiteralContext is an interface to support dynamic dispatch.
type IRangeLiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	DOUBLE_DOT() antlr.TerminalNode
	AllNUMBER() []antlr.TerminalNode
	NUMBER(i int) antlr.TerminalNode

	// IsRangeLiteralContext differentiates from other interfaces.
	IsRangeLiteralContext()
}

type RangeLiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRangeLiteralContext() *RangeLiteralContext {
	var p = new(RangeLiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_rangeLiteral
	return p
}

func InitEmptyRangeLiteralContext(p *RangeLiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_rangeLiteral
}

func (*RangeLiteralContext) IsRangeLiteralContext() {}

func NewRangeLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RangeLiteralContext {
	var p = new(RangeLiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_rangeLiteral

	return p
}

func (s *RangeLiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *RangeLiteralContext) DOUBLE_DOT() antlr.TerminalNode {
	return s.GetToken(CypherParserDOUBLE_DOT, 0)
}

func (s *RangeLiteralContext) AllNUMBER() []antlr.TerminalNode {
	return s.GetTokens(CypherParserNUMBER)
}

func (s *RangeLiteralContext) NUMBER(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserNUMBER, i)
}

func (s *RangeLiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RangeLiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RangeLiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterRangeLiteral(s)
	}
}

func (s *RangeLiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitRangeLiteral(s)
	}
}

func (p *CypherParser) RangeLiteral() (localctx IRangeLiteralContext) {
	localctx = NewRangeLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 66, CypherParserRULE_rangeLiteral)
	var _la int

	p.SetState(607)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 74, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(600)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserNUMBER {
			{
				p.SetState(599)
				p.Match(CypherParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(602)
			p.Match(CypherParserDOUBLE_DOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(604)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserNUMBER {
			{
				p.SetState(603)
				p.Match(CypherParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(606)
			p.Match(CypherParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *CypherParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 16:
		var t *ValueExprContext = nil
		if localctx != nil {
			t = localctx.(*ValueExprContext)
		}
		return p.ValueExpr_Sempred(t, predIndex)

	case 26:
		var t *ConditionContext = nil
		if localctx != nil {
			t = localctx.(*ConditionContext)
		}
		return p.Condition_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *CypherParser) ValueExpr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 7)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 6)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *CypherParser) Condition_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 2:
		return p.Precpred(p.GetParserRuleContext(), 21)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 20)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
