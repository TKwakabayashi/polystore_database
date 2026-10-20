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
		"", "'FINISH'", "'SET'", "'='", "'AND'", "'OR'", "'NOT'", "'XOR'",
	}
	staticData.SymbolicNames = []string{
		"", "", "MATCH", "RETURN", "WHERE", "DISTINCT", "AS", "WITH", "NEQ",
		"LE", "GE", "DOT", "LARROW", "RARROW", "LANGLE", "RANGLE", "COLON",
		"COMMA", "LPAREN", "RPAREN", "LSQUARE", "RSQUARE", "LCURLY", "RCURLY",
		"MINUS", "PLUS", "SLASH", "PERCENT", "SQUOTE", "STAR", "DOUBLE_DOT",
		"CREATE", "DELETE", "ORDER_BY", "ASC", "DESC", "LIMIT", "OPTIONAL",
		"UNWIND", "FINISH", "SET", "EQ", "AND", "OR", "NOT", "XOR", "COUNT",
		"REDUCE", "SUM", "AVG", "MIN", "MAX", "COLLECT", "COALESCE", "IN", "IS",
		"NULL", "TRUE", "FALSE", "ALL", "ANY", "NONE", "SINGLE", "CALL", "CASE",
		"WHEN", "THEN", "ELSE", "END", "STRING", "NUMBER", "IDENTIFIER", "WS",
	}
	staticData.RuleNames = []string{
		"cypher", "statement", "readingClause", "unwindClause", "withClause",
		"matchClause", "returnClause", "whereClause", "callClause", "asClause",
		"pattern", "node", "relationship", "returnItems", "returnItem", "caseExpr",
		"caseKey", "caseResult", "valueExpr", "aggregateFunc", "aggArg", "orderItems",
		"orderItem", "limitNum", "labels", "label", "properties", "property",
		"condition", "list", "variable", "types", "expression", "value", "range",
		"rangeLiteral",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 72, 659, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30, 7, 30, 2, 31, 7,
		31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 2, 35, 7, 35, 1, 0, 4, 0,
		74, 8, 0, 11, 0, 12, 0, 75, 1, 0, 1, 0, 1, 1, 4, 1, 81, 8, 1, 11, 1, 12,
		1, 82, 1, 1, 1, 1, 1, 2, 3, 2, 88, 8, 2, 1, 2, 1, 2, 3, 2, 92, 8, 2, 1,
		2, 1, 2, 3, 2, 96, 8, 2, 1, 3, 1, 3, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 3, 4,
		105, 8, 4, 1, 4, 1, 4, 3, 4, 109, 8, 4, 1, 4, 3, 4, 112, 8, 4, 1, 4, 3,
		4, 115, 8, 4, 1, 5, 1, 5, 5, 5, 119, 8, 5, 10, 5, 12, 5, 122, 9, 5, 1,
		5, 1, 5, 1, 5, 5, 5, 127, 8, 5, 10, 5, 12, 5, 130, 9, 5, 1, 5, 5, 5, 133,
		8, 5, 10, 5, 12, 5, 136, 9, 5, 1, 6, 1, 6, 5, 6, 140, 8, 6, 10, 6, 12,
		6, 143, 9, 6, 1, 6, 1, 6, 3, 6, 147, 8, 6, 1, 6, 3, 6, 150, 8, 6, 1, 7,
		1, 7, 5, 7, 154, 8, 7, 10, 7, 12, 7, 157, 9, 7, 1, 7, 1, 7, 1, 8, 1, 8,
		5, 8, 163, 8, 8, 10, 8, 12, 8, 166, 9, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9,
		1, 10, 1, 10, 1, 10, 3, 10, 176, 8, 10, 1, 10, 1, 10, 1, 10, 1, 10, 5,
		10, 182, 8, 10, 10, 10, 12, 10, 185, 9, 10, 1, 11, 1, 11, 3, 11, 189, 8,
		11, 1, 11, 3, 11, 192, 8, 11, 1, 11, 3, 11, 195, 8, 11, 1, 11, 1, 11, 1,
		12, 1, 12, 5, 12, 201, 8, 12, 10, 12, 12, 12, 204, 9, 12, 1, 12, 1, 12,
		3, 12, 208, 8, 12, 1, 12, 3, 12, 211, 8, 12, 1, 12, 3, 12, 214, 8, 12,
		1, 12, 3, 12, 217, 8, 12, 1, 12, 1, 12, 5, 12, 221, 8, 12, 10, 12, 12,
		12, 224, 9, 12, 1, 12, 1, 12, 1, 12, 5, 12, 229, 8, 12, 10, 12, 12, 12,
		232, 9, 12, 1, 12, 1, 12, 3, 12, 236, 8, 12, 1, 12, 3, 12, 239, 8, 12,
		1, 12, 3, 12, 242, 8, 12, 1, 12, 3, 12, 245, 8, 12, 1, 12, 1, 12, 5, 12,
		249, 8, 12, 10, 12, 12, 12, 252, 9, 12, 1, 12, 1, 12, 1, 12, 5, 12, 257,
		8, 12, 10, 12, 12, 12, 260, 9, 12, 1, 12, 1, 12, 3, 12, 264, 8, 12, 1,
		12, 3, 12, 267, 8, 12, 1, 12, 3, 12, 270, 8, 12, 1, 12, 3, 12, 273, 8,
		12, 1, 12, 1, 12, 5, 12, 277, 8, 12, 10, 12, 12, 12, 280, 9, 12, 1, 12,
		3, 12, 283, 8, 12, 1, 13, 1, 13, 1, 13, 5, 13, 288, 8, 13, 10, 13, 12,
		13, 291, 9, 13, 1, 14, 1, 14, 1, 14, 3, 14, 296, 8, 14, 1, 14, 1, 14, 1,
		14, 3, 14, 301, 8, 14, 1, 14, 1, 14, 1, 14, 3, 14, 306, 8, 14, 1, 14, 1,
		14, 1, 14, 1, 14, 1, 14, 5, 14, 313, 8, 14, 10, 14, 12, 14, 316, 9, 14,
		1, 14, 1, 14, 1, 14, 3, 14, 321, 8, 14, 3, 14, 323, 8, 14, 1, 15, 1, 15,
		1, 15, 1, 15, 1, 15, 1, 15, 4, 15, 331, 8, 15, 11, 15, 12, 15, 332, 1,
		15, 1, 15, 3, 15, 337, 8, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15, 1, 15,
		1, 15, 1, 15, 1, 15, 4, 15, 348, 8, 15, 11, 15, 12, 15, 349, 1, 15, 1,
		15, 3, 15, 354, 8, 15, 1, 15, 1, 15, 3, 15, 358, 8, 15, 1, 16, 1, 16, 1,
		16, 1, 16, 1, 16, 3, 16, 365, 8, 16, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17,
		3, 17, 372, 8, 17, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1,
		18, 5, 18, 382, 8, 18, 10, 18, 12, 18, 385, 9, 18, 3, 18, 387, 8, 18, 1,
		18, 1, 18, 1, 18, 1, 18, 5, 18, 393, 8, 18, 10, 18, 12, 18, 396, 9, 18,
		1, 18, 1, 18, 1, 18, 1, 18, 5, 18, 402, 8, 18, 10, 18, 12, 18, 405, 9,
		18, 3, 18, 407, 8, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 3, 18,
		415, 8, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 1, 18, 5, 18, 423, 8, 18,
		10, 18, 12, 18, 426, 9, 18, 1, 19, 1, 19, 1, 19, 1, 19, 3, 19, 432, 8,
		19, 1, 19, 3, 19, 435, 8, 19, 1, 19, 1, 19, 1, 20, 1, 20, 1, 20, 3, 20,
		442, 8, 20, 1, 21, 1, 21, 1, 21, 1, 21, 5, 21, 448, 8, 21, 10, 21, 12,
		21, 451, 9, 21, 1, 22, 1, 22, 3, 22, 455, 8, 22, 1, 23, 1, 23, 1, 23, 1,
		24, 1, 24, 4, 24, 462, 8, 24, 11, 24, 12, 24, 463, 1, 25, 1, 25, 1, 26,
		1, 26, 1, 26, 1, 26, 5, 26, 472, 8, 26, 10, 26, 12, 26, 475, 9, 26, 1,
		26, 1, 26, 1, 27, 1, 27, 1, 27, 1, 27, 3, 27, 483, 8, 27, 1, 28, 1, 28,
		1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1,
		28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28,
		1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1,
		28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28,
		1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1,
		28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28,
		1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1,
		28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28,
		1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1,
		28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28,
		1, 28, 3, 28, 593, 8, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 1, 28, 5,
		28, 601, 8, 28, 10, 28, 12, 28, 604, 9, 28, 1, 29, 1, 29, 1, 29, 1, 29,
		5, 29, 610, 8, 29, 10, 29, 12, 29, 613, 9, 29, 1, 29, 1, 29, 1, 30, 1,
		30, 1, 31, 1, 31, 1, 31, 1, 31, 5, 31, 623, 8, 31, 10, 31, 12, 31, 626,
		9, 31, 1, 32, 1, 32, 1, 32, 5, 32, 631, 8, 32, 10, 32, 12, 32, 634, 9,
		32, 1, 32, 1, 32, 1, 32, 1, 32, 1, 32, 3, 32, 641, 8, 32, 1, 33, 1, 33,
		1, 34, 1, 34, 3, 34, 647, 8, 34, 1, 35, 3, 35, 650, 8, 35, 1, 35, 1, 35,
		3, 35, 654, 8, 35, 1, 35, 3, 35, 657, 8, 35, 1, 35, 0, 2, 36, 56, 36, 0,
		2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38,
		40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 0, 5, 1,
		0, 69, 70, 2, 0, 26, 27, 29, 29, 1, 0, 24, 25, 2, 0, 46, 46, 48, 52, 1,
		0, 34, 35, 738, 0, 73, 1, 0, 0, 0, 2, 80, 1, 0, 0, 0, 4, 95, 1, 0, 0, 0,
		6, 97, 1, 0, 0, 0, 8, 102, 1, 0, 0, 0, 10, 116, 1, 0, 0, 0, 12, 137, 1,
		0, 0, 0, 14, 151, 1, 0, 0, 0, 16, 160, 1, 0, 0, 0, 18, 169, 1, 0, 0, 0,
		20, 175, 1, 0, 0, 0, 22, 186, 1, 0, 0, 0, 24, 282, 1, 0, 0, 0, 26, 284,
		1, 0, 0, 0, 28, 322, 1, 0, 0, 0, 30, 357, 1, 0, 0, 0, 32, 364, 1, 0, 0,
		0, 34, 371, 1, 0, 0, 0, 36, 414, 1, 0, 0, 0, 38, 427, 1, 0, 0, 0, 40, 438,
		1, 0, 0, 0, 42, 443, 1, 0, 0, 0, 44, 452, 1, 0, 0, 0, 46, 456, 1, 0, 0,
		0, 48, 459, 1, 0, 0, 0, 50, 465, 1, 0, 0, 0, 52, 467, 1, 0, 0, 0, 54, 478,
		1, 0, 0, 0, 56, 592, 1, 0, 0, 0, 58, 605, 1, 0, 0, 0, 60, 616, 1, 0, 0,
		0, 62, 618, 1, 0, 0, 0, 64, 640, 1, 0, 0, 0, 66, 642, 1, 0, 0, 0, 68, 644,
		1, 0, 0, 0, 70, 656, 1, 0, 0, 0, 72, 74, 3, 2, 1, 0, 73, 72, 1, 0, 0, 0,
		74, 75, 1, 0, 0, 0, 75, 73, 1, 0, 0, 0, 75, 76, 1, 0, 0, 0, 76, 77, 1,
		0, 0, 0, 77, 78, 5, 0, 0, 1, 78, 1, 1, 0, 0, 0, 79, 81, 3, 4, 2, 0, 80,
		79, 1, 0, 0, 0, 81, 82, 1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 82, 83, 1, 0, 0,
		0, 83, 84, 1, 0, 0, 0, 84, 85, 3, 12, 6, 0, 85, 3, 1, 0, 0, 0, 86, 88,
		5, 37, 0, 0, 87, 86, 1, 0, 0, 0, 87, 88, 1, 0, 0, 0, 88, 89, 1, 0, 0, 0,
		89, 91, 3, 10, 5, 0, 90, 92, 3, 14, 7, 0, 91, 90, 1, 0, 0, 0, 91, 92, 1,
		0, 0, 0, 92, 96, 1, 0, 0, 0, 93, 96, 3, 8, 4, 0, 94, 96, 3, 6, 3, 0, 95,
		87, 1, 0, 0, 0, 95, 93, 1, 0, 0, 0, 95, 94, 1, 0, 0, 0, 96, 5, 1, 0, 0,
		0, 97, 98, 5, 38, 0, 0, 98, 99, 3, 36, 18, 0, 99, 100, 5, 6, 0, 0, 100,
		101, 3, 60, 30, 0, 101, 7, 1, 0, 0, 0, 102, 104, 5, 7, 0, 0, 103, 105,
		5, 5, 0, 0, 104, 103, 1, 0, 0, 0, 104, 105, 1, 0, 0, 0, 105, 106, 1, 0,
		0, 0, 106, 108, 3, 26, 13, 0, 107, 109, 3, 42, 21, 0, 108, 107, 1, 0, 0,
		0, 108, 109, 1, 0, 0, 0, 109, 111, 1, 0, 0, 0, 110, 112, 3, 46, 23, 0,
		111, 110, 1, 0, 0, 0, 111, 112, 1, 0, 0, 0, 112, 114, 1, 0, 0, 0, 113,
		115, 3, 14, 7, 0, 114, 113, 1, 0, 0, 0, 114, 115, 1, 0, 0, 0, 115, 9, 1,
		0, 0, 0, 116, 120, 5, 2, 0, 0, 117, 119, 5, 72, 0, 0, 118, 117, 1, 0, 0,
		0, 119, 122, 1, 0, 0, 0, 120, 118, 1, 0, 0, 0, 120, 121, 1, 0, 0, 0, 121,
		123, 1, 0, 0, 0, 122, 120, 1, 0, 0, 0, 123, 134, 3, 20, 10, 0, 124, 128,
		5, 17, 0, 0, 125, 127, 5, 72, 0, 0, 126, 125, 1, 0, 0, 0, 127, 130, 1,
		0, 0, 0, 128, 126, 1, 0, 0, 0, 128, 129, 1, 0, 0, 0, 129, 131, 1, 0, 0,
		0, 130, 128, 1, 0, 0, 0, 131, 133, 3, 20, 10, 0, 132, 124, 1, 0, 0, 0,
		133, 136, 1, 0, 0, 0, 134, 132, 1, 0, 0, 0, 134, 135, 1, 0, 0, 0, 135,
		11, 1, 0, 0, 0, 136, 134, 1, 0, 0, 0, 137, 141, 5, 3, 0, 0, 138, 140, 5,
		72, 0, 0, 139, 138, 1, 0, 0, 0, 140, 143, 1, 0, 0, 0, 141, 139, 1, 0, 0,
		0, 141, 142, 1, 0, 0, 0, 142, 144, 1, 0, 0, 0, 143, 141, 1, 0, 0, 0, 144,
		146, 3, 26, 13, 0, 145, 147, 3, 42, 21, 0, 146, 145, 1, 0, 0, 0, 146, 147,
		1, 0, 0, 0, 147, 149, 1, 0, 0, 0, 148, 150, 3, 46, 23, 0, 149, 148, 1,
		0, 0, 0, 149, 150, 1, 0, 0, 0, 150, 13, 1, 0, 0, 0, 151, 155, 5, 4, 0,
		0, 152, 154, 5, 72, 0, 0, 153, 152, 1, 0, 0, 0, 154, 157, 1, 0, 0, 0, 155,
		153, 1, 0, 0, 0, 155, 156, 1, 0, 0, 0, 156, 158, 1, 0, 0, 0, 157, 155,
		1, 0, 0, 0, 158, 159, 3, 56, 28, 0, 159, 15, 1, 0, 0, 0, 160, 164, 5, 63,
		0, 0, 161, 163, 5, 72, 0, 0, 162, 161, 1, 0, 0, 0, 163, 166, 1, 0, 0, 0,
		164, 162, 1, 0, 0, 0, 164, 165, 1, 0, 0, 0, 165, 167, 1, 0, 0, 0, 166,
		164, 1, 0, 0, 0, 167, 168, 5, 69, 0, 0, 168, 17, 1, 0, 0, 0, 169, 170,
		5, 6, 0, 0, 170, 171, 5, 71, 0, 0, 171, 19, 1, 0, 0, 0, 172, 173, 3, 60,
		30, 0, 173, 174, 5, 41, 0, 0, 174, 176, 1, 0, 0, 0, 175, 172, 1, 0, 0,
		0, 175, 176, 1, 0, 0, 0, 176, 177, 1, 0, 0, 0, 177, 183, 3, 22, 11, 0,
		178, 179, 3, 24, 12, 0, 179, 180, 3, 22, 11, 0, 180, 182, 1, 0, 0, 0, 181,
		178, 1, 0, 0, 0, 182, 185, 1, 0, 0, 0, 183, 181, 1, 0, 0, 0, 183, 184,
		1, 0, 0, 0, 184, 21, 1, 0, 0, 0, 185, 183, 1, 0, 0, 0, 186, 188, 5, 18,
		0, 0, 187, 189, 3, 60, 30, 0, 188, 187, 1, 0, 0, 0, 188, 189, 1, 0, 0,
		0, 189, 191, 1, 0, 0, 0, 190, 192, 3, 48, 24, 0, 191, 190, 1, 0, 0, 0,
		191, 192, 1, 0, 0, 0, 192, 194, 1, 0, 0, 0, 193, 195, 3, 52, 26, 0, 194,
		193, 1, 0, 0, 0, 194, 195, 1, 0, 0, 0, 195, 196, 1, 0, 0, 0, 196, 197,
		5, 19, 0, 0, 197, 23, 1, 0, 0, 0, 198, 202, 5, 24, 0, 0, 199, 201, 5, 72,
		0, 0, 200, 199, 1, 0, 0, 0, 201, 204, 1, 0, 0, 0, 202, 200, 1, 0, 0, 0,
		202, 203, 1, 0, 0, 0, 203, 205, 1, 0, 0, 0, 204, 202, 1, 0, 0, 0, 205,
		207, 5, 20, 0, 0, 206, 208, 3, 60, 30, 0, 207, 206, 1, 0, 0, 0, 207, 208,
		1, 0, 0, 0, 208, 210, 1, 0, 0, 0, 209, 211, 3, 62, 31, 0, 210, 209, 1,
		0, 0, 0, 210, 211, 1, 0, 0, 0, 211, 213, 1, 0, 0, 0, 212, 214, 3, 68, 34,
		0, 213, 212, 1, 0, 0, 0, 213, 214, 1, 0, 0, 0, 214, 216, 1, 0, 0, 0, 215,
		217, 3, 52, 26, 0, 216, 215, 1, 0, 0, 0, 216, 217, 1, 0, 0, 0, 217, 218,
		1, 0, 0, 0, 218, 222, 5, 21, 0, 0, 219, 221, 5, 72, 0, 0, 220, 219, 1,
		0, 0, 0, 221, 224, 1, 0, 0, 0, 222, 220, 1, 0, 0, 0, 222, 223, 1, 0, 0,
		0, 223, 225, 1, 0, 0, 0, 224, 222, 1, 0, 0, 0, 225, 283, 5, 13, 0, 0, 226,
		230, 5, 12, 0, 0, 227, 229, 5, 72, 0, 0, 228, 227, 1, 0, 0, 0, 229, 232,
		1, 0, 0, 0, 230, 228, 1, 0, 0, 0, 230, 231, 1, 0, 0, 0, 231, 233, 1, 0,
		0, 0, 232, 230, 1, 0, 0, 0, 233, 235, 5, 20, 0, 0, 234, 236, 3, 60, 30,
		0, 235, 234, 1, 0, 0, 0, 235, 236, 1, 0, 0, 0, 236, 238, 1, 0, 0, 0, 237,
		239, 3, 62, 31, 0, 238, 237, 1, 0, 0, 0, 238, 239, 1, 0, 0, 0, 239, 241,
		1, 0, 0, 0, 240, 242, 3, 68, 34, 0, 241, 240, 1, 0, 0, 0, 241, 242, 1,
		0, 0, 0, 242, 244, 1, 0, 0, 0, 243, 245, 3, 52, 26, 0, 244, 243, 1, 0,
		0, 0, 244, 245, 1, 0, 0, 0, 245, 246, 1, 0, 0, 0, 246, 250, 5, 21, 0, 0,
		247, 249, 5, 72, 0, 0, 248, 247, 1, 0, 0, 0, 249, 252, 1, 0, 0, 0, 250,
		248, 1, 0, 0, 0, 250, 251, 1, 0, 0, 0, 251, 253, 1, 0, 0, 0, 252, 250,
		1, 0, 0, 0, 253, 283, 5, 24, 0, 0, 254, 258, 5, 24, 0, 0, 255, 257, 5,
		72, 0, 0, 256, 255, 1, 0, 0, 0, 257, 260, 1, 0, 0, 0, 258, 256, 1, 0, 0,
		0, 258, 259, 1, 0, 0, 0, 259, 261, 1, 0, 0, 0, 260, 258, 1, 0, 0, 0, 261,
		263, 5, 20, 0, 0, 262, 264, 3, 60, 30, 0, 263, 262, 1, 0, 0, 0, 263, 264,
		1, 0, 0, 0, 264, 266, 1, 0, 0, 0, 265, 267, 3, 62, 31, 0, 266, 265, 1,
		0, 0, 0, 266, 267, 1, 0, 0, 0, 267, 269, 1, 0, 0, 0, 268, 270, 3, 68, 34,
		0, 269, 268, 1, 0, 0, 0, 269, 270, 1, 0, 0, 0, 270, 272, 1, 0, 0, 0, 271,
		273, 3, 52, 26, 0, 272, 271, 1, 0, 0, 0, 272, 273, 1, 0, 0, 0, 273, 274,
		1, 0, 0, 0, 274, 278, 5, 21, 0, 0, 275, 277, 5, 72, 0, 0, 276, 275, 1,
		0, 0, 0, 277, 280, 1, 0, 0, 0, 278, 276, 1, 0, 0, 0, 278, 279, 1, 0, 0,
		0, 279, 281, 1, 0, 0, 0, 280, 278, 1, 0, 0, 0, 281, 283, 5, 24, 0, 0, 282,
		198, 1, 0, 0, 0, 282, 226, 1, 0, 0, 0, 282, 254, 1, 0, 0, 0, 283, 25, 1,
		0, 0, 0, 284, 289, 3, 28, 14, 0, 285, 286, 5, 17, 0, 0, 286, 288, 3, 28,
		14, 0, 287, 285, 1, 0, 0, 0, 288, 291, 1, 0, 0, 0, 289, 287, 1, 0, 0, 0,
		289, 290, 1, 0, 0, 0, 290, 27, 1, 0, 0, 0, 291, 289, 1, 0, 0, 0, 292, 295,
		3, 38, 19, 0, 293, 294, 5, 6, 0, 0, 294, 296, 3, 60, 30, 0, 295, 293, 1,
		0, 0, 0, 295, 296, 1, 0, 0, 0, 296, 323, 1, 0, 0, 0, 297, 300, 3, 30, 15,
		0, 298, 299, 5, 6, 0, 0, 299, 301, 3, 60, 30, 0, 300, 298, 1, 0, 0, 0,
		300, 301, 1, 0, 0, 0, 301, 323, 1, 0, 0, 0, 302, 305, 3, 36, 18, 0, 303,
		304, 5, 6, 0, 0, 304, 306, 3, 60, 30, 0, 305, 303, 1, 0, 0, 0, 305, 306,
		1, 0, 0, 0, 306, 323, 1, 0, 0, 0, 307, 308, 5, 53, 0, 0, 308, 309, 5, 18,
		0, 0, 309, 314, 3, 64, 32, 0, 310, 311, 5, 17, 0, 0, 311, 313, 3, 64, 32,
		0, 312, 310, 1, 0, 0, 0, 313, 316, 1, 0, 0, 0, 314, 312, 1, 0, 0, 0, 314,
		315, 1, 0, 0, 0, 315, 317, 1, 0, 0, 0, 316, 314, 1, 0, 0, 0, 317, 320,
		5, 19, 0, 0, 318, 319, 5, 6, 0, 0, 319, 321, 3, 60, 30, 0, 320, 318, 1,
		0, 0, 0, 320, 321, 1, 0, 0, 0, 321, 323, 1, 0, 0, 0, 322, 292, 1, 0, 0,
		0, 322, 297, 1, 0, 0, 0, 322, 302, 1, 0, 0, 0, 322, 307, 1, 0, 0, 0, 323,
		29, 1, 0, 0, 0, 324, 330, 5, 64, 0, 0, 325, 326, 5, 65, 0, 0, 326, 327,
		3, 56, 28, 0, 327, 328, 5, 66, 0, 0, 328, 329, 3, 34, 17, 0, 329, 331,
		1, 0, 0, 0, 330, 325, 1, 0, 0, 0, 331, 332, 1, 0, 0, 0, 332, 330, 1, 0,
		0, 0, 332, 333, 1, 0, 0, 0, 333, 336, 1, 0, 0, 0, 334, 335, 5, 67, 0, 0,
		335, 337, 3, 34, 17, 0, 336, 334, 1, 0, 0, 0, 336, 337, 1, 0, 0, 0, 337,
		338, 1, 0, 0, 0, 338, 339, 5, 68, 0, 0, 339, 358, 1, 0, 0, 0, 340, 341,
		5, 64, 0, 0, 341, 347, 3, 36, 18, 0, 342, 343, 5, 65, 0, 0, 343, 344, 3,
		32, 16, 0, 344, 345, 5, 66, 0, 0, 345, 346, 3, 34, 17, 0, 346, 348, 1,
		0, 0, 0, 347, 342, 1, 0, 0, 0, 348, 349, 1, 0, 0, 0, 349, 347, 1, 0, 0,
		0, 349, 350, 1, 0, 0, 0, 350, 353, 1, 0, 0, 0, 351, 352, 5, 67, 0, 0, 352,
		354, 3, 34, 17, 0, 353, 351, 1, 0, 0, 0, 353, 354, 1, 0, 0, 0, 354, 355,
		1, 0, 0, 0, 355, 356, 5, 68, 0, 0, 356, 358, 1, 0, 0, 0, 357, 324, 1, 0,
		0, 0, 357, 340, 1, 0, 0, 0, 358, 31, 1, 0, 0, 0, 359, 365, 3, 66, 33, 0,
		360, 365, 5, 56, 0, 0, 361, 365, 5, 57, 0, 0, 362, 365, 5, 58, 0, 0, 363,
		365, 3, 64, 32, 0, 364, 359, 1, 0, 0, 0, 364, 360, 1, 0, 0, 0, 364, 361,
		1, 0, 0, 0, 364, 362, 1, 0, 0, 0, 364, 363, 1, 0, 0, 0, 365, 33, 1, 0,
		0, 0, 366, 372, 3, 66, 33, 0, 367, 372, 5, 56, 0, 0, 368, 372, 5, 57, 0,
		0, 369, 372, 5, 58, 0, 0, 370, 372, 3, 64, 32, 0, 371, 366, 1, 0, 0, 0,
		371, 367, 1, 0, 0, 0, 371, 368, 1, 0, 0, 0, 371, 369, 1, 0, 0, 0, 371,
		370, 1, 0, 0, 0, 372, 35, 1, 0, 0, 0, 373, 374, 6, 18, -1, 0, 374, 375,
		5, 24, 0, 0, 375, 415, 3, 36, 18, 8, 376, 377, 5, 71, 0, 0, 377, 386, 5,
		18, 0, 0, 378, 383, 3, 36, 18, 0, 379, 380, 5, 17, 0, 0, 380, 382, 3, 36,
		18, 0, 381, 379, 1, 0, 0, 0, 382, 385, 1, 0, 0, 0, 383, 381, 1, 0, 0, 0,
		383, 384, 1, 0, 0, 0, 384, 387, 1, 0, 0, 0, 385, 383, 1, 0, 0, 0, 386,
		378, 1, 0, 0, 0, 386, 387, 1, 0, 0, 0, 387, 388, 1, 0, 0, 0, 388, 415,
		5, 19, 0, 0, 389, 394, 5, 71, 0, 0, 390, 391, 5, 11, 0, 0, 391, 393, 5,
		71, 0, 0, 392, 390, 1, 0, 0, 0, 393, 396, 1, 0, 0, 0, 394, 392, 1, 0, 0,
		0, 394, 395, 1, 0, 0, 0, 395, 415, 1, 0, 0, 0, 396, 394, 1, 0, 0, 0, 397,
		406, 5, 20, 0, 0, 398, 403, 3, 36, 18, 0, 399, 400, 5, 17, 0, 0, 400, 402,
		3, 36, 18, 0, 401, 399, 1, 0, 0, 0, 402, 405, 1, 0, 0, 0, 403, 401, 1,
		0, 0, 0, 403, 404, 1, 0, 0, 0, 404, 407, 1, 0, 0, 0, 405, 403, 1, 0, 0,
		0, 406, 398, 1, 0, 0, 0, 406, 407, 1, 0, 0, 0, 407, 408, 1, 0, 0, 0, 408,
		415, 5, 21, 0, 0, 409, 415, 7, 0, 0, 0, 410, 411, 5, 18, 0, 0, 411, 412,
		3, 36, 18, 0, 412, 413, 5, 19, 0, 0, 413, 415, 1, 0, 0, 0, 414, 373, 1,
		0, 0, 0, 414, 376, 1, 0, 0, 0, 414, 389, 1, 0, 0, 0, 414, 397, 1, 0, 0,
		0, 414, 409, 1, 0, 0, 0, 414, 410, 1, 0, 0, 0, 415, 424, 1, 0, 0, 0, 416,
		417, 10, 7, 0, 0, 417, 418, 7, 1, 0, 0, 418, 423, 3, 36, 18, 8, 419, 420,
		10, 6, 0, 0, 420, 421, 7, 2, 0, 0, 421, 423, 3, 36, 18, 7, 422, 416, 1,
		0, 0, 0, 422, 419, 1, 0, 0, 0, 423, 426, 1, 0, 0, 0, 424, 422, 1, 0, 0,
		0, 424, 425, 1, 0, 0, 0, 425, 37, 1, 0, 0, 0, 426, 424, 1, 0, 0, 0, 427,
		428, 7, 3, 0, 0, 428, 434, 5, 18, 0, 0, 429, 435, 5, 29, 0, 0, 430, 432,
		5, 5, 0, 0, 431, 430, 1, 0, 0, 0, 431, 432, 1, 0, 0, 0, 432, 433, 1, 0,
		0, 0, 433, 435, 3, 40, 20, 0, 434, 429, 1, 0, 0, 0, 434, 431, 1, 0, 0,
		0, 435, 436, 1, 0, 0, 0, 436, 437, 5, 19, 0, 0, 437, 39, 1, 0, 0, 0, 438,
		441, 5, 71, 0, 0, 439, 440, 5, 11, 0, 0, 440, 442, 5, 71, 0, 0, 441, 439,
		1, 0, 0, 0, 441, 442, 1, 0, 0, 0, 442, 41, 1, 0, 0, 0, 443, 444, 5, 33,
		0, 0, 444, 449, 3, 44, 22, 0, 445, 446, 5, 17, 0, 0, 446, 448, 3, 44, 22,
		0, 447, 445, 1, 0, 0, 0, 448, 451, 1, 0, 0, 0, 449, 447, 1, 0, 0, 0, 449,
		450, 1, 0, 0, 0, 450, 43, 1, 0, 0, 0, 451, 449, 1, 0, 0, 0, 452, 454, 3,
		64, 32, 0, 453, 455, 7, 4, 0, 0, 454, 453, 1, 0, 0, 0, 454, 455, 1, 0,
		0, 0, 455, 45, 1, 0, 0, 0, 456, 457, 5, 36, 0, 0, 457, 458, 5, 70, 0, 0,
		458, 47, 1, 0, 0, 0, 459, 461, 5, 16, 0, 0, 460, 462, 3, 50, 25, 0, 461,
		460, 1, 0, 0, 0, 462, 463, 1, 0, 0, 0, 463, 461, 1, 0, 0, 0, 463, 464,
		1, 0, 0, 0, 464, 49, 1, 0, 0, 0, 465, 466, 5, 71, 0, 0, 466, 51, 1, 0,
		0, 0, 467, 468, 5, 22, 0, 0, 468, 473, 3, 54, 27, 0, 469, 470, 5, 17, 0,
		0, 470, 472, 3, 54, 27, 0, 471, 469, 1, 0, 0, 0, 472, 475, 1, 0, 0, 0,
		473, 471, 1, 0, 0, 0, 473, 474, 1, 0, 0, 0, 474, 476, 1, 0, 0, 0, 475,
		473, 1, 0, 0, 0, 476, 477, 5, 23, 0, 0, 477, 53, 1, 0, 0, 0, 478, 479,
		5, 71, 0, 0, 479, 482, 5, 16, 0, 0, 480, 483, 3, 66, 33, 0, 481, 483, 3,
		64, 32, 0, 482, 480, 1, 0, 0, 0, 482, 481, 1, 0, 0, 0, 483, 55, 1, 0, 0,
		0, 484, 485, 6, 28, -1, 0, 485, 486, 5, 18, 0, 0, 486, 487, 3, 56, 28,
		0, 487, 488, 5, 19, 0, 0, 488, 593, 1, 0, 0, 0, 489, 490, 5, 44, 0, 0,
		490, 593, 3, 56, 28, 23, 491, 492, 5, 59, 0, 0, 492, 493, 5, 18, 0, 0,
		493, 494, 3, 60, 30, 0, 494, 495, 5, 54, 0, 0, 495, 496, 3, 64, 32, 0,
		496, 497, 5, 4, 0, 0, 497, 498, 3, 56, 28, 0, 498, 499, 5, 19, 0, 0, 499,
		593, 1, 0, 0, 0, 500, 501, 5, 60, 0, 0, 501, 502, 5, 18, 0, 0, 502, 503,
		3, 60, 30, 0, 503, 504, 5, 54, 0, 0, 504, 505, 3, 64, 32, 0, 505, 506,
		5, 4, 0, 0, 506, 507, 3, 56, 28, 0, 507, 508, 5, 19, 0, 0, 508, 593, 1,
		0, 0, 0, 509, 510, 5, 61, 0, 0, 510, 511, 5, 18, 0, 0, 511, 512, 3, 60,
		30, 0, 512, 513, 5, 54, 0, 0, 513, 514, 3, 64, 32, 0, 514, 515, 5, 4, 0,
		0, 515, 516, 3, 56, 28, 0, 516, 517, 5, 19, 0, 0, 517, 593, 1, 0, 0, 0,
		518, 519, 5, 62, 0, 0, 519, 520, 5, 18, 0, 0, 520, 521, 3, 60, 30, 0, 521,
		522, 5, 54, 0, 0, 522, 523, 3, 64, 32, 0, 523, 524, 5, 4, 0, 0, 524, 525,
		3, 56, 28, 0, 525, 526, 5, 19, 0, 0, 526, 593, 1, 0, 0, 0, 527, 528, 3,
		64, 32, 0, 528, 529, 5, 55, 0, 0, 529, 530, 5, 56, 0, 0, 530, 593, 1, 0,
		0, 0, 531, 532, 3, 64, 32, 0, 532, 533, 5, 55, 0, 0, 533, 534, 5, 44, 0,
		0, 534, 535, 5, 56, 0, 0, 535, 593, 1, 0, 0, 0, 536, 537, 3, 64, 32, 0,
		537, 538, 5, 54, 0, 0, 538, 539, 3, 58, 29, 0, 539, 593, 1, 0, 0, 0, 540,
		541, 3, 64, 32, 0, 541, 542, 5, 54, 0, 0, 542, 543, 3, 64, 32, 0, 543,
		593, 1, 0, 0, 0, 544, 545, 3, 64, 32, 0, 545, 546, 5, 41, 0, 0, 546, 547,
		3, 66, 33, 0, 547, 593, 1, 0, 0, 0, 548, 549, 3, 64, 32, 0, 549, 550, 5,
		8, 0, 0, 550, 551, 3, 66, 33, 0, 551, 593, 1, 0, 0, 0, 552, 553, 3, 64,
		32, 0, 553, 554, 5, 15, 0, 0, 554, 555, 3, 66, 33, 0, 555, 593, 1, 0, 0,
		0, 556, 557, 3, 64, 32, 0, 557, 558, 5, 14, 0, 0, 558, 559, 3, 66, 33,
		0, 559, 593, 1, 0, 0, 0, 560, 561, 3, 64, 32, 0, 561, 562, 5, 10, 0, 0,
		562, 563, 3, 66, 33, 0, 563, 593, 1, 0, 0, 0, 564, 565, 3, 64, 32, 0, 565,
		566, 5, 9, 0, 0, 566, 567, 3, 66, 33, 0, 567, 593, 1, 0, 0, 0, 568, 569,
		3, 64, 32, 0, 569, 570, 5, 41, 0, 0, 570, 571, 3, 64, 32, 0, 571, 593,
		1, 0, 0, 0, 572, 573, 3, 64, 32, 0, 573, 574, 5, 8, 0, 0, 574, 575, 3,
		64, 32, 0, 575, 593, 1, 0, 0, 0, 576, 577, 3, 64, 32, 0, 577, 578, 5, 15,
		0, 0, 578, 579, 3, 64, 32, 0, 579, 593, 1, 0, 0, 0, 580, 581, 3, 64, 32,
		0, 581, 582, 5, 14, 0, 0, 582, 583, 3, 64, 32, 0, 583, 593, 1, 0, 0, 0,
		584, 585, 3, 64, 32, 0, 585, 586, 5, 10, 0, 0, 586, 587, 3, 64, 32, 0,
		587, 593, 1, 0, 0, 0, 588, 589, 3, 64, 32, 0, 589, 590, 5, 9, 0, 0, 590,
		591, 3, 64, 32, 0, 591, 593, 1, 0, 0, 0, 592, 484, 1, 0, 0, 0, 592, 489,
		1, 0, 0, 0, 592, 491, 1, 0, 0, 0, 592, 500, 1, 0, 0, 0, 592, 509, 1, 0,
		0, 0, 592, 518, 1, 0, 0, 0, 592, 527, 1, 0, 0, 0, 592, 531, 1, 0, 0, 0,
		592, 536, 1, 0, 0, 0, 592, 540, 1, 0, 0, 0, 592, 544, 1, 0, 0, 0, 592,
		548, 1, 0, 0, 0, 592, 552, 1, 0, 0, 0, 592, 556, 1, 0, 0, 0, 592, 560,
		1, 0, 0, 0, 592, 564, 1, 0, 0, 0, 592, 568, 1, 0, 0, 0, 592, 572, 1, 0,
		0, 0, 592, 576, 1, 0, 0, 0, 592, 580, 1, 0, 0, 0, 592, 584, 1, 0, 0, 0,
		592, 588, 1, 0, 0, 0, 593, 602, 1, 0, 0, 0, 594, 595, 10, 22, 0, 0, 595,
		596, 5, 42, 0, 0, 596, 601, 3, 56, 28, 23, 597, 598, 10, 21, 0, 0, 598,
		599, 5, 43, 0, 0, 599, 601, 3, 56, 28, 22, 600, 594, 1, 0, 0, 0, 600, 597,
		1, 0, 0, 0, 601, 604, 1, 0, 0, 0, 602, 600, 1, 0, 0, 0, 602, 603, 1, 0,
		0, 0, 603, 57, 1, 0, 0, 0, 604, 602, 1, 0, 0, 0, 605, 606, 5, 20, 0, 0,
		606, 611, 3, 66, 33, 0, 607, 608, 5, 17, 0, 0, 608, 610, 3, 66, 33, 0,
		609, 607, 1, 0, 0, 0, 610, 613, 1, 0, 0, 0, 611, 609, 1, 0, 0, 0, 611,
		612, 1, 0, 0, 0, 612, 614, 1, 0, 0, 0, 613, 611, 1, 0, 0, 0, 614, 615,
		5, 21, 0, 0, 615, 59, 1, 0, 0, 0, 616, 617, 5, 71, 0, 0, 617, 61, 1, 0,
		0, 0, 618, 619, 5, 16, 0, 0, 619, 624, 5, 71, 0, 0, 620, 621, 5, 1, 0,
		0, 621, 623, 5, 71, 0, 0, 622, 620, 1, 0, 0, 0, 623, 626, 1, 0, 0, 0, 624,
		622, 1, 0, 0, 0, 624, 625, 1, 0, 0, 0, 625, 63, 1, 0, 0, 0, 626, 624, 1,
		0, 0, 0, 627, 632, 5, 71, 0, 0, 628, 629, 5, 11, 0, 0, 629, 631, 5, 71,
		0, 0, 630, 628, 1, 0, 0, 0, 631, 634, 1, 0, 0, 0, 632, 630, 1, 0, 0, 0,
		632, 633, 1, 0, 0, 0, 633, 641, 1, 0, 0, 0, 634, 632, 1, 0, 0, 0, 635,
		636, 5, 71, 0, 0, 636, 637, 5, 18, 0, 0, 637, 638, 3, 60, 30, 0, 638, 639,
		5, 19, 0, 0, 639, 641, 1, 0, 0, 0, 640, 627, 1, 0, 0, 0, 640, 635, 1, 0,
		0, 0, 641, 65, 1, 0, 0, 0, 642, 643, 7, 0, 0, 0, 643, 67, 1, 0, 0, 0, 644,
		646, 5, 29, 0, 0, 645, 647, 3, 70, 35, 0, 646, 645, 1, 0, 0, 0, 646, 647,
		1, 0, 0, 0, 647, 69, 1, 0, 0, 0, 648, 650, 5, 70, 0, 0, 649, 648, 1, 0,
		0, 0, 649, 650, 1, 0, 0, 0, 650, 651, 1, 0, 0, 0, 651, 653, 5, 30, 0, 0,
		652, 654, 5, 70, 0, 0, 653, 652, 1, 0, 0, 0, 653, 654, 1, 0, 0, 0, 654,
		657, 1, 0, 0, 0, 655, 657, 5, 70, 0, 0, 656, 649, 1, 0, 0, 0, 656, 655,
		1, 0, 0, 0, 657, 71, 1, 0, 0, 0, 82, 75, 82, 87, 91, 95, 104, 108, 111,
		114, 120, 128, 134, 141, 146, 149, 155, 164, 175, 183, 188, 191, 194, 202,
		207, 210, 213, 216, 222, 230, 235, 238, 241, 244, 250, 258, 263, 266, 269,
		272, 278, 282, 289, 295, 300, 305, 314, 320, 322, 332, 336, 349, 353, 357,
		364, 371, 383, 386, 394, 403, 406, 414, 422, 424, 431, 434, 441, 449, 454,
		463, 473, 482, 592, 600, 602, 611, 624, 632, 640, 646, 649, 653, 656,
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
	CypherParserCOLLECT    = 52
	CypherParserCOALESCE   = 53
	CypherParserIN         = 54
	CypherParserIS         = 55
	CypherParserNULL       = 56
	CypherParserTRUE       = 57
	CypherParserFALSE      = 58
	CypherParserALL        = 59
	CypherParserANY        = 60
	CypherParserNONE       = 61
	CypherParserSINGLE     = 62
	CypherParserCALL       = 63
	CypherParserCASE       = 64
	CypherParserWHEN       = 65
	CypherParserTHEN       = 66
	CypherParserELSE       = 67
	CypherParserEND        = 68
	CypherParserSTRING     = 69
	CypherParserNUMBER     = 70
	CypherParserIDENTIFIER = 71
	CypherParserWS         = 72
)

// CypherParser rules.
const (
	CypherParserRULE_cypher        = 0
	CypherParserRULE_statement     = 1
	CypherParserRULE_readingClause = 2
	CypherParserRULE_unwindClause  = 3
	CypherParserRULE_withClause    = 4
	CypherParserRULE_matchClause   = 5
	CypherParserRULE_returnClause  = 6
	CypherParserRULE_whereClause   = 7
	CypherParserRULE_callClause    = 8
	CypherParserRULE_asClause      = 9
	CypherParserRULE_pattern       = 10
	CypherParserRULE_node          = 11
	CypherParserRULE_relationship  = 12
	CypherParserRULE_returnItems   = 13
	CypherParserRULE_returnItem    = 14
	CypherParserRULE_caseExpr      = 15
	CypherParserRULE_caseKey       = 16
	CypherParserRULE_caseResult    = 17
	CypherParserRULE_valueExpr     = 18
	CypherParserRULE_aggregateFunc = 19
	CypherParserRULE_aggArg        = 20
	CypherParserRULE_orderItems    = 21
	CypherParserRULE_orderItem     = 22
	CypherParserRULE_limitNum      = 23
	CypherParserRULE_labels        = 24
	CypherParserRULE_label         = 25
	CypherParserRULE_properties    = 26
	CypherParserRULE_property      = 27
	CypherParserRULE_condition     = 28
	CypherParserRULE_list          = 29
	CypherParserRULE_variable      = 30
	CypherParserRULE_types         = 31
	CypherParserRULE_expression    = 32
	CypherParserRULE_value         = 33
	CypherParserRULE_range         = 34
	CypherParserRULE_rangeLiteral  = 35
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
	p.SetState(73)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&412316860548) != 0) {
		{
			p.SetState(72)
			p.Statement()
		}

		p.SetState(75)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(77)
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
	p.SetState(80)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&412316860548) != 0) {
		{
			p.SetState(79)
			p.ReadingClause()
		}

		p.SetState(82)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(84)
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
	OPTIONAL() antlr.TerminalNode
	WhereClause() IWhereClauseContext
	WithClause() IWithClauseContext
	UnwindClause() IUnwindClauseContext

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

func (s *ReadingClauseContext) OPTIONAL() antlr.TerminalNode {
	return s.GetToken(CypherParserOPTIONAL, 0)
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

func (s *ReadingClauseContext) UnwindClause() IUnwindClauseContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnwindClauseContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnwindClauseContext)
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

	p.SetState(95)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserMATCH, CypherParserOPTIONAL:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(87)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserOPTIONAL {
			{
				p.SetState(86)
				p.Match(CypherParserOPTIONAL)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(89)
			p.MatchClause()
		}
		p.SetState(91)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserWHERE {
			{
				p.SetState(90)
				p.WhereClause()
			}

		}

	case CypherParserWITH:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(93)
			p.WithClause()
		}

	case CypherParserUNWIND:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(94)
			p.UnwindClause()
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

// IUnwindClauseContext is an interface to support dynamic dispatch.
type IUnwindClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	UNWIND() antlr.TerminalNode
	ValueExpr() IValueExprContext
	AS() antlr.TerminalNode
	Variable() IVariableContext

	// IsUnwindClauseContext differentiates from other interfaces.
	IsUnwindClauseContext()
}

type UnwindClauseContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnwindClauseContext() *UnwindClauseContext {
	var p = new(UnwindClauseContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_unwindClause
	return p
}

func InitEmptyUnwindClauseContext(p *UnwindClauseContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_unwindClause
}

func (*UnwindClauseContext) IsUnwindClauseContext() {}

func NewUnwindClauseContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UnwindClauseContext {
	var p = new(UnwindClauseContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_unwindClause

	return p
}

func (s *UnwindClauseContext) GetParser() antlr.Parser { return s.parser }

func (s *UnwindClauseContext) UNWIND() antlr.TerminalNode {
	return s.GetToken(CypherParserUNWIND, 0)
}

func (s *UnwindClauseContext) ValueExpr() IValueExprContext {
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

func (s *UnwindClauseContext) AS() antlr.TerminalNode {
	return s.GetToken(CypherParserAS, 0)
}

func (s *UnwindClauseContext) Variable() IVariableContext {
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

func (s *UnwindClauseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnwindClauseContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UnwindClauseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterUnwindClause(s)
	}
}

func (s *UnwindClauseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitUnwindClause(s)
	}
}

func (p *CypherParser) UnwindClause() (localctx IUnwindClauseContext) {
	localctx = NewUnwindClauseContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, CypherParserRULE_unwindClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(97)
		p.Match(CypherParserUNWIND)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(98)
		p.valueExpr(0)
	}
	{
		p.SetState(99)
		p.Match(CypherParserAS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(100)
		p.Variable()
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
	p.EnterRule(localctx, 8, CypherParserRULE_withClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(102)
		p.Match(CypherParserWITH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(104)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDISTINCT {
		{
			p.SetState(103)
			p.Match(CypherParserDISTINCT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(106)
		p.ReturnItems()
	}
	p.SetState(108)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserORDER_BY {
		{
			p.SetState(107)
			p.OrderItems()
		}

	}
	p.SetState(111)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLIMIT {
		{
			p.SetState(110)
			p.LimitNum()
		}

	}
	p.SetState(114)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserWHERE {
		{
			p.SetState(113)
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
	p.EnterRule(localctx, 10, CypherParserRULE_matchClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(116)
		p.Match(CypherParserMATCH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(120)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(117)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(122)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(123)
		p.Pattern()
	}
	p.SetState(134)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(124)
			p.Match(CypherParserCOMMA)
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
			p.Pattern()
		}

		p.SetState(136)
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
	p.EnterRule(localctx, 12, CypherParserRULE_returnClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(137)
		p.Match(CypherParserRETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(141)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(138)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(143)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(144)
		p.ReturnItems()
	}
	p.SetState(146)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserORDER_BY {
		{
			p.SetState(145)
			p.OrderItems()
		}

	}
	p.SetState(149)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLIMIT {
		{
			p.SetState(148)
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
	p.EnterRule(localctx, 14, CypherParserRULE_whereClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(151)
		p.Match(CypherParserWHERE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(155)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(152)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(157)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(158)
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
	p.EnterRule(localctx, 16, CypherParserRULE_callClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(160)
		p.Match(CypherParserCALL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(164)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(161)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(166)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(167)
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
	p.EnterRule(localctx, 18, CypherParserRULE_asClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(169)
		p.Match(CypherParserAS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(170)
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
	p.EnterRule(localctx, 20, CypherParserRULE_pattern)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(175)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserIDENTIFIER {
		{
			p.SetState(172)
			p.Variable()
		}
		{
			p.SetState(173)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(177)
		p.Node()
	}
	p.SetState(183)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserLARROW || _la == CypherParserMINUS {
		{
			p.SetState(178)
			p.Relationship()
		}
		{
			p.SetState(179)
			p.Node()
		}

		p.SetState(185)
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
	p.EnterRule(localctx, 22, CypherParserRULE_node)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(186)
		p.Match(CypherParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(188)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserIDENTIFIER {
		{
			p.SetState(187)
			p.Variable()
		}

	}
	p.SetState(191)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserCOLON {
		{
			p.SetState(190)
			p.Labels()
		}

	}
	p.SetState(194)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLCURLY {
		{
			p.SetState(193)
			p.Properties()
		}

	}
	{
		p.SetState(196)
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
	p.EnterRule(localctx, 24, CypherParserRULE_relationship)
	var _la int

	p.SetState(282)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 40, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(198)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(202)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(199)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(204)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(205)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(207)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(206)
				p.Variable()
			}

		}
		p.SetState(210)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(209)
				p.Types()
			}

		}
		p.SetState(213)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(212)
				p.Range_()
			}

		}
		p.SetState(216)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(215)
				p.Properties()
			}

		}
		{
			p.SetState(218)
			p.Match(CypherParserRSQUARE)
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

		for _la == CypherParserWS {
			{
				p.SetState(219)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(224)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(225)
			p.Match(CypherParserRARROW)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(226)
			p.Match(CypherParserLARROW)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(230)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(227)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(232)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(233)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(235)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(234)
				p.Variable()
			}

		}
		p.SetState(238)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(237)
				p.Types()
			}

		}
		p.SetState(241)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(240)
				p.Range_()
			}

		}
		p.SetState(244)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(243)
				p.Properties()
			}

		}
		{
			p.SetState(246)
			p.Match(CypherParserRSQUARE)
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

		for _la == CypherParserWS {
			{
				p.SetState(247)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(252)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(253)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(254)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(258)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(255)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(260)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(261)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(263)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(262)
				p.Variable()
			}

		}
		p.SetState(266)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(265)
				p.Types()
			}

		}
		p.SetState(269)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(268)
				p.Range_()
			}

		}
		p.SetState(272)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(271)
				p.Properties()
			}

		}
		{
			p.SetState(274)
			p.Match(CypherParserRSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(278)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(275)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(280)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(281)
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
	p.EnterRule(localctx, 26, CypherParserRULE_returnItems)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(284)
		p.ReturnItem()
	}
	p.SetState(289)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(285)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(286)
			p.ReturnItem()
		}

		p.SetState(291)
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
	p.EnterRule(localctx, 28, CypherParserRULE_returnItem)
	var _la int

	p.SetState(322)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserCOUNT, CypherParserSUM, CypherParserAVG, CypherParserMIN, CypherParserMAX, CypherParserCOLLECT:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(292)
			p.AggregateFunc()
		}
		p.SetState(295)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(293)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(294)
				p.Variable()
			}

		}

	case CypherParserCASE:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(297)
			p.CaseExpr()
		}
		p.SetState(300)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(298)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(299)
				p.Variable()
			}

		}

	case CypherParserLPAREN, CypherParserLSQUARE, CypherParserMINUS, CypherParserSTRING, CypherParserNUMBER, CypherParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(302)
			p.valueExpr(0)
		}
		p.SetState(305)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(303)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(304)
				p.Variable()
			}

		}

	case CypherParserCOALESCE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(307)
			p.Match(CypherParserCOALESCE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(308)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(309)
			p.Expression()
		}
		p.SetState(314)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserCOMMA {
			{
				p.SetState(310)
				p.Match(CypherParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(311)
				p.Expression()
			}

			p.SetState(316)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(317)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(320)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(318)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(319)
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

func (s *CaseExprContext) CopyAll(ctx *CaseExprContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *CaseExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CaseExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type CaseSimpleContext struct {
	CaseExprContext
}

func NewCaseSimpleContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CaseSimpleContext {
	var p = new(CaseSimpleContext)

	InitEmptyCaseExprContext(&p.CaseExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*CaseExprContext))

	return p
}

func (s *CaseSimpleContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CaseSimpleContext) CASE() antlr.TerminalNode {
	return s.GetToken(CypherParserCASE, 0)
}

func (s *CaseSimpleContext) ValueExpr() IValueExprContext {
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

func (s *CaseSimpleContext) END() antlr.TerminalNode {
	return s.GetToken(CypherParserEND, 0)
}

func (s *CaseSimpleContext) AllWHEN() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWHEN)
}

func (s *CaseSimpleContext) WHEN(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWHEN, i)
}

func (s *CaseSimpleContext) AllCaseKey() []ICaseKeyContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICaseKeyContext); ok {
			len++
		}
	}

	tst := make([]ICaseKeyContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICaseKeyContext); ok {
			tst[i] = t.(ICaseKeyContext)
			i++
		}
	}

	return tst
}

func (s *CaseSimpleContext) CaseKey(i int) ICaseKeyContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICaseKeyContext); ok {
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

	return t.(ICaseKeyContext)
}

func (s *CaseSimpleContext) AllTHEN() []antlr.TerminalNode {
	return s.GetTokens(CypherParserTHEN)
}

func (s *CaseSimpleContext) THEN(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserTHEN, i)
}

func (s *CaseSimpleContext) AllCaseResult() []ICaseResultContext {
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

func (s *CaseSimpleContext) CaseResult(i int) ICaseResultContext {
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

func (s *CaseSimpleContext) ELSE() antlr.TerminalNode {
	return s.GetToken(CypherParserELSE, 0)
}

func (s *CaseSimpleContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterCaseSimple(s)
	}
}

func (s *CaseSimpleContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitCaseSimple(s)
	}
}

type CaseSearchedContext struct {
	CaseExprContext
}

func NewCaseSearchedContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CaseSearchedContext {
	var p = new(CaseSearchedContext)

	InitEmptyCaseExprContext(&p.CaseExprContext)
	p.parser = parser
	p.CopyAll(ctx.(*CaseExprContext))

	return p
}

func (s *CaseSearchedContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CaseSearchedContext) CASE() antlr.TerminalNode {
	return s.GetToken(CypherParserCASE, 0)
}

func (s *CaseSearchedContext) END() antlr.TerminalNode {
	return s.GetToken(CypherParserEND, 0)
}

func (s *CaseSearchedContext) AllWHEN() []antlr.TerminalNode {
	return s.GetTokens(CypherParserWHEN)
}

func (s *CaseSearchedContext) WHEN(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserWHEN, i)
}

func (s *CaseSearchedContext) AllCondition() []IConditionContext {
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

func (s *CaseSearchedContext) Condition(i int) IConditionContext {
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

func (s *CaseSearchedContext) AllTHEN() []antlr.TerminalNode {
	return s.GetTokens(CypherParserTHEN)
}

func (s *CaseSearchedContext) THEN(i int) antlr.TerminalNode {
	return s.GetToken(CypherParserTHEN, i)
}

func (s *CaseSearchedContext) AllCaseResult() []ICaseResultContext {
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

func (s *CaseSearchedContext) CaseResult(i int) ICaseResultContext {
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

func (s *CaseSearchedContext) ELSE() antlr.TerminalNode {
	return s.GetToken(CypherParserELSE, 0)
}

func (s *CaseSearchedContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterCaseSearched(s)
	}
}

func (s *CaseSearchedContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitCaseSearched(s)
	}
}

func (p *CypherParser) CaseExpr() (localctx ICaseExprContext) {
	localctx = NewCaseExprContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 30, CypherParserRULE_caseExpr)
	var _la int

	p.SetState(357)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 52, p.GetParserRuleContext()) {
	case 1:
		localctx = NewCaseSearchedContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(324)
			p.Match(CypherParserCASE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(330)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == CypherParserWHEN {
			{
				p.SetState(325)
				p.Match(CypherParserWHEN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(326)
				p.condition(0)
			}
			{
				p.SetState(327)
				p.Match(CypherParserTHEN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(328)
				p.CaseResult()
			}

			p.SetState(332)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(336)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserELSE {
			{
				p.SetState(334)
				p.Match(CypherParserELSE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(335)
				p.CaseResult()
			}

		}
		{
			p.SetState(338)
			p.Match(CypherParserEND)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewCaseSimpleContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(340)
			p.Match(CypherParserCASE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(341)
			p.valueExpr(0)
		}
		p.SetState(347)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for ok := true; ok; ok = _la == CypherParserWHEN {
			{
				p.SetState(342)
				p.Match(CypherParserWHEN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(343)
				p.CaseKey()
			}
			{
				p.SetState(344)
				p.Match(CypherParserTHEN)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(345)
				p.CaseResult()
			}

			p.SetState(349)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		p.SetState(353)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserELSE {
			{
				p.SetState(351)
				p.Match(CypherParserELSE)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(352)
				p.CaseResult()
			}

		}
		{
			p.SetState(355)
			p.Match(CypherParserEND)
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

// ICaseKeyContext is an interface to support dynamic dispatch.
type ICaseKeyContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Value() IValueContext
	NULL() antlr.TerminalNode
	TRUE() antlr.TerminalNode
	FALSE() antlr.TerminalNode
	Expression() IExpressionContext

	// IsCaseKeyContext differentiates from other interfaces.
	IsCaseKeyContext()
}

type CaseKeyContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCaseKeyContext() *CaseKeyContext {
	var p = new(CaseKeyContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_caseKey
	return p
}

func InitEmptyCaseKeyContext(p *CaseKeyContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = CypherParserRULE_caseKey
}

func (*CaseKeyContext) IsCaseKeyContext() {}

func NewCaseKeyContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CaseKeyContext {
	var p = new(CaseKeyContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = CypherParserRULE_caseKey

	return p
}

func (s *CaseKeyContext) GetParser() antlr.Parser { return s.parser }

func (s *CaseKeyContext) Value() IValueContext {
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

func (s *CaseKeyContext) NULL() antlr.TerminalNode {
	return s.GetToken(CypherParserNULL, 0)
}

func (s *CaseKeyContext) TRUE() antlr.TerminalNode {
	return s.GetToken(CypherParserTRUE, 0)
}

func (s *CaseKeyContext) FALSE() antlr.TerminalNode {
	return s.GetToken(CypherParserFALSE, 0)
}

func (s *CaseKeyContext) Expression() IExpressionContext {
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

func (s *CaseKeyContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CaseKeyContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CaseKeyContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterCaseKey(s)
	}
}

func (s *CaseKeyContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitCaseKey(s)
	}
}

func (p *CypherParser) CaseKey() (localctx ICaseKeyContext) {
	localctx = NewCaseKeyContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 32, CypherParserRULE_caseKey)
	p.SetState(364)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserSTRING, CypherParserNUMBER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(359)
			p.Value()
		}

	case CypherParserNULL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(360)
			p.Match(CypherParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserTRUE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(361)
			p.Match(CypherParserTRUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserFALSE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(362)
			p.Match(CypherParserFALSE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(363)
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

// ICaseResultContext is an interface to support dynamic dispatch.
type ICaseResultContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Value() IValueContext
	NULL() antlr.TerminalNode
	TRUE() antlr.TerminalNode
	FALSE() antlr.TerminalNode
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

func (s *CaseResultContext) NULL() antlr.TerminalNode {
	return s.GetToken(CypherParserNULL, 0)
}

func (s *CaseResultContext) TRUE() antlr.TerminalNode {
	return s.GetToken(CypherParserTRUE, 0)
}

func (s *CaseResultContext) FALSE() antlr.TerminalNode {
	return s.GetToken(CypherParserFALSE, 0)
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
	p.EnterRule(localctx, 34, CypherParserRULE_caseResult)
	p.SetState(371)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserSTRING, CypherParserNUMBER:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(366)
			p.Value()
		}

	case CypherParserNULL:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(367)
			p.Match(CypherParserNULL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserTRUE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(368)
			p.Match(CypherParserTRUE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserFALSE:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(369)
			p.Match(CypherParserFALSE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(370)
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
	_startState := 36
	p.EnterRecursionRule(localctx, 36, CypherParserRULE_valueExpr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(414)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 60, p.GetParserRuleContext()) {
	case 1:
		localctx = NewVeNegContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(374)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(375)
			p.valueExpr(8)
		}

	case 2:
		localctx = NewVeCallContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(376)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(377)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(386)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64((_la-18)) & ^0x3f) == 0 && ((int64(1)<<(_la-18))&15762598695796805) != 0 {
			{
				p.SetState(378)
				p.valueExpr(0)
			}
			p.SetState(383)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == CypherParserCOMMA {
				{
					p.SetState(379)
					p.Match(CypherParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(380)
					p.valueExpr(0)
				}

				p.SetState(385)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(388)
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
			p.SetState(389)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(394)
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
				{
					p.SetState(390)
					p.Match(CypherParserDOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(391)
					p.Match(CypherParserIDENTIFIER)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(396)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 57, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 4:
		localctx = NewVeListContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(397)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(406)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64((_la-18)) & ^0x3f) == 0 && ((int64(1)<<(_la-18))&15762598695796805) != 0 {
			{
				p.SetState(398)
				p.valueExpr(0)
			}
			p.SetState(403)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == CypherParserCOMMA {
				{
					p.SetState(399)
					p.Match(CypherParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(400)
					p.valueExpr(0)
				}

				p.SetState(405)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(408)
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
			p.SetState(409)
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
			p.SetState(410)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(411)
			p.valueExpr(0)
		}
		{
			p.SetState(412)
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
	p.SetState(424)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 62, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(422)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 61, p.GetParserRuleContext()) {
			case 1:
				localctx = NewVeMulDivContext(p, NewValueExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_valueExpr)
				p.SetState(416)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(417)

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
					p.SetState(418)
					p.valueExpr(8)
				}

			case 2:
				localctx = NewVeAddSubContext(p, NewValueExprContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_valueExpr)
				p.SetState(419)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
					goto errorExit
				}
				{
					p.SetState(420)

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
					p.SetState(421)
					p.valueExpr(7)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(426)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 62, p.GetParserRuleContext())
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
	COLLECT() antlr.TerminalNode
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

func (s *AggregateFuncContext) COLLECT() antlr.TerminalNode {
	return s.GetToken(CypherParserCOLLECT, 0)
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
	p.EnterRule(localctx, 38, CypherParserRULE_aggregateFunc)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(427)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&8796093022208000) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(428)
		p.Match(CypherParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(434)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserSTAR:
		{
			p.SetState(429)
			p.Match(CypherParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserDISTINCT, CypherParserIDENTIFIER:
		p.SetState(431)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserDISTINCT {
			{
				p.SetState(430)
				p.Match(CypherParserDISTINCT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(433)
			p.AggArg()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(436)
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
	p.EnterRule(localctx, 40, CypherParserRULE_aggArg)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(438)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(441)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDOT {
		{
			p.SetState(439)
			p.Match(CypherParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(440)
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
	p.EnterRule(localctx, 42, CypherParserRULE_orderItems)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(443)
		p.Match(CypherParserORDER_BY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(444)
		p.OrderItem()
	}
	p.SetState(449)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(445)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(446)
			p.OrderItem()
		}

		p.SetState(451)
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
	p.EnterRule(localctx, 44, CypherParserRULE_orderItem)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(452)
		p.Expression()
	}
	p.SetState(454)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserASC || _la == CypherParserDESC {
		{
			p.SetState(453)
			_la = p.GetTokenStream().LA(1)

			if !(_la == CypherParserASC || _la == CypherParserDESC) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
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
	p.EnterRule(localctx, 46, CypherParserRULE_limitNum)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(456)
		p.Match(CypherParserLIMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(457)
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
	p.EnterRule(localctx, 48, CypherParserRULE_labels)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(459)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(461)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == CypherParserIDENTIFIER {
		{
			p.SetState(460)
			p.Label()
		}

		p.SetState(463)
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
	p.EnterRule(localctx, 50, CypherParserRULE_label)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(465)
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
	p.EnterRule(localctx, 52, CypherParserRULE_properties)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(467)
		p.Match(CypherParserLCURLY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(468)
		p.Property()
	}
	p.SetState(473)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(469)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(470)
			p.Property()
		}

		p.SetState(475)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(476)
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
	Expression() IExpressionContext

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

func (s *PropertyContext) Expression() IExpressionContext {
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
	p.EnterRule(localctx, 54, CypherParserRULE_property)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(478)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(479)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(482)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserSTRING, CypherParserNUMBER:
		{
			p.SetState(480)
			p.Value()
		}

	case CypherParserIDENTIFIER:
		{
			p.SetState(481)
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

type ConditionInExprContext struct {
	ConditionContext
}

func NewConditionInExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConditionInExprContext {
	var p = new(ConditionInExprContext)

	InitEmptyConditionContext(&p.ConditionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ConditionContext))

	return p
}

func (s *ConditionInExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionInExprContext) AllExpression() []IExpressionContext {
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

func (s *ConditionInExprContext) Expression(i int) IExpressionContext {
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

func (s *ConditionInExprContext) IN() antlr.TerminalNode {
	return s.GetToken(CypherParserIN, 0)
}

func (s *ConditionInExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.EnterConditionInExpr(s)
	}
}

func (s *ConditionInExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(CypherListener); ok {
		listenerT.ExitConditionInExpr(s)
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
	_startState := 56
	p.EnterRecursionRule(localctx, 56, CypherParserRULE_condition, _p)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(592)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 71, p.GetParserRuleContext()) {
	case 1:
		localctx = NewConditionParenContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(485)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(486)
			p.condition(0)
		}
		{
			p.SetState(487)
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
			p.SetState(489)
			p.Match(CypherParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(490)
			p.condition(23)
		}

	case 3:
		localctx = NewConditionAllContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(491)
			p.Match(CypherParserALL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(492)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(493)
			p.Variable()
		}
		{
			p.SetState(494)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(495)
			p.Expression()
		}
		{
			p.SetState(496)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(497)
			p.condition(0)
		}
		{
			p.SetState(498)
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
			p.SetState(500)
			p.Match(CypherParserANY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(501)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(502)
			p.Variable()
		}
		{
			p.SetState(503)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(504)
			p.Expression()
		}
		{
			p.SetState(505)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(506)
			p.condition(0)
		}
		{
			p.SetState(507)
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
			p.SetState(509)
			p.Match(CypherParserNONE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(510)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(511)
			p.Variable()
		}
		{
			p.SetState(512)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(513)
			p.Expression()
		}
		{
			p.SetState(514)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(515)
			p.condition(0)
		}
		{
			p.SetState(516)
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
			p.SetState(518)
			p.Match(CypherParserSINGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(519)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(520)
			p.Variable()
		}
		{
			p.SetState(521)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(522)
			p.Expression()
		}
		{
			p.SetState(523)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(524)
			p.condition(0)
		}
		{
			p.SetState(525)
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
			p.SetState(527)
			p.Expression()
		}
		{
			p.SetState(528)
			p.Match(CypherParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(529)
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
			p.SetState(531)
			p.Expression()
		}
		{
			p.SetState(532)
			p.Match(CypherParserIS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(533)
			p.Match(CypherParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(534)
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
			p.SetState(536)
			p.Expression()
		}
		{
			p.SetState(537)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(538)
			p.List()
		}

	case 10:
		localctx = NewConditionInExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(540)
			p.Expression()
		}
		{
			p.SetState(541)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(542)
			p.Expression()
		}

	case 11:
		localctx = NewConditionEqualityContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(544)
			p.Expression()
		}
		{
			p.SetState(545)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(546)
			p.Value()
		}

	case 12:
		localctx = NewConditionNotEqualityContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(548)
			p.Expression()
		}
		{
			p.SetState(549)
			p.Match(CypherParserNEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(550)
			p.Value()
		}

	case 13:
		localctx = NewConditionGreaterContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(552)
			p.Expression()
		}
		{
			p.SetState(553)
			p.Match(CypherParserRANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(554)
			p.Value()
		}

	case 14:
		localctx = NewConditionLessContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(556)
			p.Expression()
		}
		{
			p.SetState(557)
			p.Match(CypherParserLANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(558)
			p.Value()
		}

	case 15:
		localctx = NewConditionGreaterEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(560)
			p.Expression()
		}
		{
			p.SetState(561)
			p.Match(CypherParserGE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(562)
			p.Value()
		}

	case 16:
		localctx = NewConditionLessEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(564)
			p.Expression()
		}
		{
			p.SetState(565)
			p.Match(CypherParserLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(566)
			p.Value()
		}

	case 17:
		localctx = NewConditionVarEqContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(568)
			p.Expression()
		}
		{
			p.SetState(569)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(570)
			p.Expression()
		}

	case 18:
		localctx = NewConditionVarNeqContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(572)
			p.Expression()
		}
		{
			p.SetState(573)
			p.Match(CypherParserNEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(574)
			p.Expression()
		}

	case 19:
		localctx = NewConditionVarGreaterContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(576)
			p.Expression()
		}
		{
			p.SetState(577)
			p.Match(CypherParserRANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(578)
			p.Expression()
		}

	case 20:
		localctx = NewConditionVarLessContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(580)
			p.Expression()
		}
		{
			p.SetState(581)
			p.Match(CypherParserLANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(582)
			p.Expression()
		}

	case 21:
		localctx = NewConditionVarGreaterEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(584)
			p.Expression()
		}
		{
			p.SetState(585)
			p.Match(CypherParserGE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(586)
			p.Expression()
		}

	case 22:
		localctx = NewConditionVarLessEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(588)
			p.Expression()
		}
		{
			p.SetState(589)
			p.Match(CypherParserLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(590)
			p.Expression()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(602)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 73, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(600)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 72, p.GetParserRuleContext()) {
			case 1:
				localctx = NewConditionAndContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_condition)
				p.SetState(594)

				if !(p.Precpred(p.GetParserRuleContext(), 22)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 22)", ""))
					goto errorExit
				}
				{
					p.SetState(595)
					p.Match(CypherParserAND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(596)
					p.condition(23)
				}

			case 2:
				localctx = NewConditionOrContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_condition)
				p.SetState(597)

				if !(p.Precpred(p.GetParserRuleContext(), 21)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 21)", ""))
					goto errorExit
				}
				{
					p.SetState(598)
					p.Match(CypherParserOR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(599)
					p.condition(22)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(604)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 73, p.GetParserRuleContext())
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
	p.EnterRule(localctx, 58, CypherParserRULE_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(605)
		p.Match(CypherParserLSQUARE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(606)
		p.Value()
	}
	p.SetState(611)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(607)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(608)
			p.Value()
		}

		p.SetState(613)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(614)
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
	p.EnterRule(localctx, 60, CypherParserRULE_variable)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(616)
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
	p.EnterRule(localctx, 62, CypherParserRULE_types)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(618)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(619)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(624)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserT__0 {
		{
			p.SetState(620)
			p.Match(CypherParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(621)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(626)
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
	p.EnterRule(localctx, 64, CypherParserRULE_expression)
	var _alt int

	p.SetState(640)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 77, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(627)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(632)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 76, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(628)
					p.Match(CypherParserDOT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(629)
					p.Match(CypherParserIDENTIFIER)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(634)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 76, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(635)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(636)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(637)
			p.Variable()
		}
		{
			p.SetState(638)
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
	p.EnterRule(localctx, 66, CypherParserRULE_value)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(642)
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
	p.EnterRule(localctx, 68, CypherParserRULE_range)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(644)
		p.Match(CypherParserSTAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(646)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDOUBLE_DOT || _la == CypherParserNUMBER {
		{
			p.SetState(645)
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
	p.EnterRule(localctx, 70, CypherParserRULE_rangeLiteral)
	var _la int

	p.SetState(656)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 81, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(649)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserNUMBER {
			{
				p.SetState(648)
				p.Match(CypherParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(651)
			p.Match(CypherParserDOUBLE_DOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(653)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserNUMBER {
			{
				p.SetState(652)
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
			p.SetState(655)
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
	case 18:
		var t *ValueExprContext = nil
		if localctx != nil {
			t = localctx.(*ValueExprContext)
		}
		return p.ValueExpr_Sempred(t, predIndex)

	case 28:
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
		return p.Precpred(p.GetParserRuleContext(), 22)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 21)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
