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
		"'{'", "'}'", "'-'", "'''", "'*'", "'..'", "'CREATE'", "'DELETE'", "'ORDER BY'",
		"'ASC'", "'DESC'", "'LIMIT'", "'OPTIONAL'", "'UNWIND'", "'FINISH'",
		"'SET'", "'='", "'AND'", "'OR'", "'NOT'", "'XOR'",
	}
	staticData.SymbolicNames = []string{
		"", "", "MATCH", "RETURN", "WHERE", "DISTINCT", "AS", "WITH", "NEQ",
		"LE", "GE", "DOT", "LARROW", "RARROW", "LANGLE", "RANGLE", "COLON",
		"COMMA", "LPAREN", "RPAREN", "LSQUARE", "RSQUARE", "LCURLY", "RCURLY",
		"MINUS", "SQUOTE", "STAR", "DOUBLE_DOT", "CREATE", "DELETE", "ORDER_BY",
		"ASC", "DESC", "LIMIT", "OPTIONAL", "UNWIND", "FINISH", "SET", "EQ",
		"AND", "OR", "NOT", "XOR", "COUNT", "REDUCE", "SUM", "AVG", "MIN", "MAX",
		"COALESCE", "IN", "ALL", "ANY", "NONE", "SINGLE", "CALL", "STRING",
		"NUMBER", "IDENTIFIER", "WS",
	}
	staticData.RuleNames = []string{
		"cypher", "statement", "matchClause", "returnClause", "whereClause",
		"callClause", "asClause", "pattern", "node", "relationship", "returnItems",
		"returnItem", "aggregateFunc", "aggArg", "orderItems", "orderItem",
		"limitNum", "labels", "label", "properties", "property", "condition",
		"variable", "types", "expression", "value", "range", "rangeLiteral",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 59, 437, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15, 7, 15,
		2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7, 20, 2,
		21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25, 2, 26,
		7, 26, 2, 27, 7, 27, 1, 0, 4, 0, 58, 8, 0, 11, 0, 12, 0, 59, 1, 0, 1, 0,
		1, 1, 1, 1, 3, 1, 66, 8, 1, 1, 1, 1, 1, 1, 2, 1, 2, 5, 2, 72, 8, 2, 10,
		2, 12, 2, 75, 9, 2, 1, 2, 1, 2, 1, 3, 1, 3, 5, 3, 81, 8, 3, 10, 3, 12,
		3, 84, 9, 3, 1, 3, 1, 3, 3, 3, 88, 8, 3, 1, 3, 3, 3, 91, 8, 3, 1, 4, 1,
		4, 5, 4, 95, 8, 4, 10, 4, 12, 4, 98, 9, 4, 1, 4, 1, 4, 1, 5, 1, 5, 5, 5,
		104, 8, 5, 10, 5, 12, 5, 107, 9, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 7,
		1, 7, 1, 7, 3, 7, 117, 8, 7, 1, 7, 1, 7, 1, 7, 1, 7, 5, 7, 123, 8, 7, 10,
		7, 12, 7, 126, 9, 7, 1, 8, 1, 8, 3, 8, 130, 8, 8, 1, 8, 3, 8, 133, 8, 8,
		1, 8, 3, 8, 136, 8, 8, 1, 8, 1, 8, 1, 9, 1, 9, 5, 9, 142, 8, 9, 10, 9,
		12, 9, 145, 9, 9, 1, 9, 1, 9, 3, 9, 149, 8, 9, 1, 9, 3, 9, 152, 8, 9, 1,
		9, 3, 9, 155, 8, 9, 1, 9, 3, 9, 158, 8, 9, 1, 9, 1, 9, 5, 9, 162, 8, 9,
		10, 9, 12, 9, 165, 9, 9, 1, 9, 1, 9, 1, 9, 5, 9, 170, 8, 9, 10, 9, 12,
		9, 173, 9, 9, 1, 9, 1, 9, 3, 9, 177, 8, 9, 1, 9, 3, 9, 180, 8, 9, 1, 9,
		3, 9, 183, 8, 9, 1, 9, 3, 9, 186, 8, 9, 1, 9, 1, 9, 5, 9, 190, 8, 9, 10,
		9, 12, 9, 193, 9, 9, 1, 9, 1, 9, 1, 9, 5, 9, 198, 8, 9, 10, 9, 12, 9, 201,
		9, 9, 1, 9, 1, 9, 3, 9, 205, 8, 9, 1, 9, 3, 9, 208, 8, 9, 1, 9, 3, 9, 211,
		8, 9, 1, 9, 3, 9, 214, 8, 9, 1, 9, 1, 9, 5, 9, 218, 8, 9, 10, 9, 12, 9,
		221, 9, 9, 1, 9, 3, 9, 224, 8, 9, 1, 10, 1, 10, 1, 10, 5, 10, 229, 8, 10,
		10, 10, 12, 10, 232, 9, 10, 1, 11, 1, 11, 1, 11, 3, 11, 237, 8, 11, 1,
		11, 1, 11, 1, 11, 3, 11, 242, 8, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11,
		5, 11, 249, 8, 11, 10, 11, 12, 11, 252, 9, 11, 1, 11, 1, 11, 1, 11, 3,
		11, 257, 8, 11, 3, 11, 259, 8, 11, 1, 12, 1, 12, 1, 12, 1, 12, 3, 12, 265,
		8, 12, 1, 12, 3, 12, 268, 8, 12, 1, 12, 1, 12, 1, 13, 1, 13, 1, 13, 3,
		13, 275, 8, 13, 1, 14, 1, 14, 1, 14, 1, 14, 5, 14, 281, 8, 14, 10, 14,
		12, 14, 284, 9, 14, 1, 15, 1, 15, 1, 15, 1, 16, 1, 16, 1, 16, 1, 17, 1,
		17, 4, 17, 294, 8, 17, 11, 17, 12, 17, 295, 1, 18, 1, 18, 1, 19, 1, 19,
		1, 19, 1, 19, 5, 19, 304, 8, 19, 10, 19, 12, 19, 307, 9, 19, 1, 19, 1,
		19, 1, 20, 1, 20, 1, 20, 1, 20, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21,
		1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1,
		21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21,
		1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1,
		21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21,
		1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1,
		21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 3, 21, 382,
		8, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 1, 21, 5, 21, 390, 8, 21, 10,
		21, 12, 21, 393, 9, 21, 1, 22, 1, 22, 1, 23, 1, 23, 1, 23, 1, 23, 5, 23,
		401, 8, 23, 10, 23, 12, 23, 404, 9, 23, 1, 24, 1, 24, 1, 24, 5, 24, 409,
		8, 24, 10, 24, 12, 24, 412, 9, 24, 1, 24, 1, 24, 1, 24, 1, 24, 1, 24, 3,
		24, 419, 8, 24, 1, 25, 1, 25, 1, 26, 1, 26, 3, 26, 425, 8, 26, 1, 27, 3,
		27, 428, 8, 27, 1, 27, 1, 27, 3, 27, 432, 8, 27, 1, 27, 3, 27, 435, 8,
		27, 1, 27, 0, 1, 42, 28, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24,
		26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 0, 3, 2, 0,
		43, 43, 45, 48, 1, 0, 31, 32, 1, 0, 56, 57, 474, 0, 57, 1, 0, 0, 0, 2,
		63, 1, 0, 0, 0, 4, 69, 1, 0, 0, 0, 6, 78, 1, 0, 0, 0, 8, 92, 1, 0, 0, 0,
		10, 101, 1, 0, 0, 0, 12, 110, 1, 0, 0, 0, 14, 116, 1, 0, 0, 0, 16, 127,
		1, 0, 0, 0, 18, 223, 1, 0, 0, 0, 20, 225, 1, 0, 0, 0, 22, 258, 1, 0, 0,
		0, 24, 260, 1, 0, 0, 0, 26, 271, 1, 0, 0, 0, 28, 276, 1, 0, 0, 0, 30, 285,
		1, 0, 0, 0, 32, 288, 1, 0, 0, 0, 34, 291, 1, 0, 0, 0, 36, 297, 1, 0, 0,
		0, 38, 299, 1, 0, 0, 0, 40, 310, 1, 0, 0, 0, 42, 381, 1, 0, 0, 0, 44, 394,
		1, 0, 0, 0, 46, 396, 1, 0, 0, 0, 48, 418, 1, 0, 0, 0, 50, 420, 1, 0, 0,
		0, 52, 422, 1, 0, 0, 0, 54, 434, 1, 0, 0, 0, 56, 58, 3, 2, 1, 0, 57, 56,
		1, 0, 0, 0, 58, 59, 1, 0, 0, 0, 59, 57, 1, 0, 0, 0, 59, 60, 1, 0, 0, 0,
		60, 61, 1, 0, 0, 0, 61, 62, 5, 0, 0, 1, 62, 1, 1, 0, 0, 0, 63, 65, 3, 4,
		2, 0, 64, 66, 3, 8, 4, 0, 65, 64, 1, 0, 0, 0, 65, 66, 1, 0, 0, 0, 66, 67,
		1, 0, 0, 0, 67, 68, 3, 6, 3, 0, 68, 3, 1, 0, 0, 0, 69, 73, 5, 2, 0, 0,
		70, 72, 5, 59, 0, 0, 71, 70, 1, 0, 0, 0, 72, 75, 1, 0, 0, 0, 73, 71, 1,
		0, 0, 0, 73, 74, 1, 0, 0, 0, 74, 76, 1, 0, 0, 0, 75, 73, 1, 0, 0, 0, 76,
		77, 3, 14, 7, 0, 77, 5, 1, 0, 0, 0, 78, 82, 5, 3, 0, 0, 79, 81, 5, 59,
		0, 0, 80, 79, 1, 0, 0, 0, 81, 84, 1, 0, 0, 0, 82, 80, 1, 0, 0, 0, 82, 83,
		1, 0, 0, 0, 83, 85, 1, 0, 0, 0, 84, 82, 1, 0, 0, 0, 85, 87, 3, 20, 10,
		0, 86, 88, 3, 28, 14, 0, 87, 86, 1, 0, 0, 0, 87, 88, 1, 0, 0, 0, 88, 90,
		1, 0, 0, 0, 89, 91, 3, 32, 16, 0, 90, 89, 1, 0, 0, 0, 90, 91, 1, 0, 0,
		0, 91, 7, 1, 0, 0, 0, 92, 96, 5, 4, 0, 0, 93, 95, 5, 59, 0, 0, 94, 93,
		1, 0, 0, 0, 95, 98, 1, 0, 0, 0, 96, 94, 1, 0, 0, 0, 96, 97, 1, 0, 0, 0,
		97, 99, 1, 0, 0, 0, 98, 96, 1, 0, 0, 0, 99, 100, 3, 42, 21, 0, 100, 9,
		1, 0, 0, 0, 101, 105, 5, 55, 0, 0, 102, 104, 5, 59, 0, 0, 103, 102, 1,
		0, 0, 0, 104, 107, 1, 0, 0, 0, 105, 103, 1, 0, 0, 0, 105, 106, 1, 0, 0,
		0, 106, 108, 1, 0, 0, 0, 107, 105, 1, 0, 0, 0, 108, 109, 5, 56, 0, 0, 109,
		11, 1, 0, 0, 0, 110, 111, 5, 6, 0, 0, 111, 112, 5, 58, 0, 0, 112, 13, 1,
		0, 0, 0, 113, 114, 3, 44, 22, 0, 114, 115, 5, 38, 0, 0, 115, 117, 1, 0,
		0, 0, 116, 113, 1, 0, 0, 0, 116, 117, 1, 0, 0, 0, 117, 118, 1, 0, 0, 0,
		118, 124, 3, 16, 8, 0, 119, 120, 3, 18, 9, 0, 120, 121, 3, 16, 8, 0, 121,
		123, 1, 0, 0, 0, 122, 119, 1, 0, 0, 0, 123, 126, 1, 0, 0, 0, 124, 122,
		1, 0, 0, 0, 124, 125, 1, 0, 0, 0, 125, 15, 1, 0, 0, 0, 126, 124, 1, 0,
		0, 0, 127, 129, 5, 18, 0, 0, 128, 130, 3, 44, 22, 0, 129, 128, 1, 0, 0,
		0, 129, 130, 1, 0, 0, 0, 130, 132, 1, 0, 0, 0, 131, 133, 3, 34, 17, 0,
		132, 131, 1, 0, 0, 0, 132, 133, 1, 0, 0, 0, 133, 135, 1, 0, 0, 0, 134,
		136, 3, 38, 19, 0, 135, 134, 1, 0, 0, 0, 135, 136, 1, 0, 0, 0, 136, 137,
		1, 0, 0, 0, 137, 138, 5, 19, 0, 0, 138, 17, 1, 0, 0, 0, 139, 143, 5, 24,
		0, 0, 140, 142, 5, 59, 0, 0, 141, 140, 1, 0, 0, 0, 142, 145, 1, 0, 0, 0,
		143, 141, 1, 0, 0, 0, 143, 144, 1, 0, 0, 0, 144, 146, 1, 0, 0, 0, 145,
		143, 1, 0, 0, 0, 146, 148, 5, 20, 0, 0, 147, 149, 3, 44, 22, 0, 148, 147,
		1, 0, 0, 0, 148, 149, 1, 0, 0, 0, 149, 151, 1, 0, 0, 0, 150, 152, 3, 46,
		23, 0, 151, 150, 1, 0, 0, 0, 151, 152, 1, 0, 0, 0, 152, 154, 1, 0, 0, 0,
		153, 155, 3, 52, 26, 0, 154, 153, 1, 0, 0, 0, 154, 155, 1, 0, 0, 0, 155,
		157, 1, 0, 0, 0, 156, 158, 3, 38, 19, 0, 157, 156, 1, 0, 0, 0, 157, 158,
		1, 0, 0, 0, 158, 159, 1, 0, 0, 0, 159, 163, 5, 21, 0, 0, 160, 162, 5, 59,
		0, 0, 161, 160, 1, 0, 0, 0, 162, 165, 1, 0, 0, 0, 163, 161, 1, 0, 0, 0,
		163, 164, 1, 0, 0, 0, 164, 166, 1, 0, 0, 0, 165, 163, 1, 0, 0, 0, 166,
		224, 5, 13, 0, 0, 167, 171, 5, 12, 0, 0, 168, 170, 5, 59, 0, 0, 169, 168,
		1, 0, 0, 0, 170, 173, 1, 0, 0, 0, 171, 169, 1, 0, 0, 0, 171, 172, 1, 0,
		0, 0, 172, 174, 1, 0, 0, 0, 173, 171, 1, 0, 0, 0, 174, 176, 5, 20, 0, 0,
		175, 177, 3, 44, 22, 0, 176, 175, 1, 0, 0, 0, 176, 177, 1, 0, 0, 0, 177,
		179, 1, 0, 0, 0, 178, 180, 3, 46, 23, 0, 179, 178, 1, 0, 0, 0, 179, 180,
		1, 0, 0, 0, 180, 182, 1, 0, 0, 0, 181, 183, 3, 52, 26, 0, 182, 181, 1,
		0, 0, 0, 182, 183, 1, 0, 0, 0, 183, 185, 1, 0, 0, 0, 184, 186, 3, 38, 19,
		0, 185, 184, 1, 0, 0, 0, 185, 186, 1, 0, 0, 0, 186, 187, 1, 0, 0, 0, 187,
		191, 5, 21, 0, 0, 188, 190, 5, 59, 0, 0, 189, 188, 1, 0, 0, 0, 190, 193,
		1, 0, 0, 0, 191, 189, 1, 0, 0, 0, 191, 192, 1, 0, 0, 0, 192, 194, 1, 0,
		0, 0, 193, 191, 1, 0, 0, 0, 194, 224, 5, 24, 0, 0, 195, 199, 5, 24, 0,
		0, 196, 198, 5, 59, 0, 0, 197, 196, 1, 0, 0, 0, 198, 201, 1, 0, 0, 0, 199,
		197, 1, 0, 0, 0, 199, 200, 1, 0, 0, 0, 200, 202, 1, 0, 0, 0, 201, 199,
		1, 0, 0, 0, 202, 204, 5, 20, 0, 0, 203, 205, 3, 44, 22, 0, 204, 203, 1,
		0, 0, 0, 204, 205, 1, 0, 0, 0, 205, 207, 1, 0, 0, 0, 206, 208, 3, 46, 23,
		0, 207, 206, 1, 0, 0, 0, 207, 208, 1, 0, 0, 0, 208, 210, 1, 0, 0, 0, 209,
		211, 3, 52, 26, 0, 210, 209, 1, 0, 0, 0, 210, 211, 1, 0, 0, 0, 211, 213,
		1, 0, 0, 0, 212, 214, 3, 38, 19, 0, 213, 212, 1, 0, 0, 0, 213, 214, 1,
		0, 0, 0, 214, 215, 1, 0, 0, 0, 215, 219, 5, 21, 0, 0, 216, 218, 5, 59,
		0, 0, 217, 216, 1, 0, 0, 0, 218, 221, 1, 0, 0, 0, 219, 217, 1, 0, 0, 0,
		219, 220, 1, 0, 0, 0, 220, 222, 1, 0, 0, 0, 221, 219, 1, 0, 0, 0, 222,
		224, 5, 24, 0, 0, 223, 139, 1, 0, 0, 0, 223, 167, 1, 0, 0, 0, 223, 195,
		1, 0, 0, 0, 224, 19, 1, 0, 0, 0, 225, 230, 3, 22, 11, 0, 226, 227, 5, 17,
		0, 0, 227, 229, 3, 22, 11, 0, 228, 226, 1, 0, 0, 0, 229, 232, 1, 0, 0,
		0, 230, 228, 1, 0, 0, 0, 230, 231, 1, 0, 0, 0, 231, 21, 1, 0, 0, 0, 232,
		230, 1, 0, 0, 0, 233, 236, 3, 24, 12, 0, 234, 235, 5, 6, 0, 0, 235, 237,
		3, 44, 22, 0, 236, 234, 1, 0, 0, 0, 236, 237, 1, 0, 0, 0, 237, 259, 1,
		0, 0, 0, 238, 241, 3, 48, 24, 0, 239, 240, 5, 6, 0, 0, 240, 242, 3, 44,
		22, 0, 241, 239, 1, 0, 0, 0, 241, 242, 1, 0, 0, 0, 242, 259, 1, 0, 0, 0,
		243, 244, 5, 49, 0, 0, 244, 245, 5, 18, 0, 0, 245, 250, 3, 48, 24, 0, 246,
		247, 5, 17, 0, 0, 247, 249, 3, 48, 24, 0, 248, 246, 1, 0, 0, 0, 249, 252,
		1, 0, 0, 0, 250, 248, 1, 0, 0, 0, 250, 251, 1, 0, 0, 0, 251, 253, 1, 0,
		0, 0, 252, 250, 1, 0, 0, 0, 253, 256, 5, 19, 0, 0, 254, 255, 5, 6, 0, 0,
		255, 257, 3, 44, 22, 0, 256, 254, 1, 0, 0, 0, 256, 257, 1, 0, 0, 0, 257,
		259, 1, 0, 0, 0, 258, 233, 1, 0, 0, 0, 258, 238, 1, 0, 0, 0, 258, 243,
		1, 0, 0, 0, 259, 23, 1, 0, 0, 0, 260, 261, 7, 0, 0, 0, 261, 267, 5, 18,
		0, 0, 262, 268, 5, 26, 0, 0, 263, 265, 5, 5, 0, 0, 264, 263, 1, 0, 0, 0,
		264, 265, 1, 0, 0, 0, 265, 266, 1, 0, 0, 0, 266, 268, 3, 26, 13, 0, 267,
		262, 1, 0, 0, 0, 267, 264, 1, 0, 0, 0, 268, 269, 1, 0, 0, 0, 269, 270,
		5, 19, 0, 0, 270, 25, 1, 0, 0, 0, 271, 274, 5, 58, 0, 0, 272, 273, 5, 11,
		0, 0, 273, 275, 5, 58, 0, 0, 274, 272, 1, 0, 0, 0, 274, 275, 1, 0, 0, 0,
		275, 27, 1, 0, 0, 0, 276, 277, 5, 30, 0, 0, 277, 282, 3, 30, 15, 0, 278,
		279, 5, 17, 0, 0, 279, 281, 3, 30, 15, 0, 280, 278, 1, 0, 0, 0, 281, 284,
		1, 0, 0, 0, 282, 280, 1, 0, 0, 0, 282, 283, 1, 0, 0, 0, 283, 29, 1, 0,
		0, 0, 284, 282, 1, 0, 0, 0, 285, 286, 3, 48, 24, 0, 286, 287, 7, 1, 0,
		0, 287, 31, 1, 0, 0, 0, 288, 289, 5, 33, 0, 0, 289, 290, 5, 57, 0, 0, 290,
		33, 1, 0, 0, 0, 291, 293, 5, 16, 0, 0, 292, 294, 3, 36, 18, 0, 293, 292,
		1, 0, 0, 0, 294, 295, 1, 0, 0, 0, 295, 293, 1, 0, 0, 0, 295, 296, 1, 0,
		0, 0, 296, 35, 1, 0, 0, 0, 297, 298, 5, 58, 0, 0, 298, 37, 1, 0, 0, 0,
		299, 300, 5, 22, 0, 0, 300, 305, 3, 40, 20, 0, 301, 302, 5, 17, 0, 0, 302,
		304, 3, 40, 20, 0, 303, 301, 1, 0, 0, 0, 304, 307, 1, 0, 0, 0, 305, 303,
		1, 0, 0, 0, 305, 306, 1, 0, 0, 0, 306, 308, 1, 0, 0, 0, 307, 305, 1, 0,
		0, 0, 308, 309, 5, 23, 0, 0, 309, 39, 1, 0, 0, 0, 310, 311, 5, 58, 0, 0,
		311, 312, 5, 16, 0, 0, 312, 313, 3, 50, 25, 0, 313, 41, 1, 0, 0, 0, 314,
		315, 6, 21, -1, 0, 315, 316, 5, 18, 0, 0, 316, 317, 3, 42, 21, 0, 317,
		318, 5, 19, 0, 0, 318, 382, 1, 0, 0, 0, 319, 320, 5, 41, 0, 0, 320, 382,
		3, 42, 21, 13, 321, 322, 5, 51, 0, 0, 322, 323, 5, 18, 0, 0, 323, 324,
		3, 44, 22, 0, 324, 325, 5, 50, 0, 0, 325, 326, 3, 48, 24, 0, 326, 327,
		5, 4, 0, 0, 327, 328, 3, 42, 21, 0, 328, 329, 5, 19, 0, 0, 329, 382, 1,
		0, 0, 0, 330, 331, 5, 52, 0, 0, 331, 332, 5, 18, 0, 0, 332, 333, 3, 44,
		22, 0, 333, 334, 5, 50, 0, 0, 334, 335, 3, 48, 24, 0, 335, 336, 5, 4, 0,
		0, 336, 337, 3, 42, 21, 0, 337, 338, 5, 19, 0, 0, 338, 382, 1, 0, 0, 0,
		339, 340, 5, 53, 0, 0, 340, 341, 5, 18, 0, 0, 341, 342, 3, 44, 22, 0, 342,
		343, 5, 50, 0, 0, 343, 344, 3, 48, 24, 0, 344, 345, 5, 4, 0, 0, 345, 346,
		3, 42, 21, 0, 346, 347, 5, 19, 0, 0, 347, 382, 1, 0, 0, 0, 348, 349, 5,
		54, 0, 0, 349, 350, 5, 18, 0, 0, 350, 351, 3, 44, 22, 0, 351, 352, 5, 50,
		0, 0, 352, 353, 3, 48, 24, 0, 353, 354, 5, 4, 0, 0, 354, 355, 3, 42, 21,
		0, 355, 356, 5, 19, 0, 0, 356, 382, 1, 0, 0, 0, 357, 358, 3, 48, 24, 0,
		358, 359, 5, 38, 0, 0, 359, 360, 3, 50, 25, 0, 360, 382, 1, 0, 0, 0, 361,
		362, 3, 48, 24, 0, 362, 363, 5, 8, 0, 0, 363, 364, 3, 50, 25, 0, 364, 382,
		1, 0, 0, 0, 365, 366, 3, 48, 24, 0, 366, 367, 5, 15, 0, 0, 367, 368, 3,
		50, 25, 0, 368, 382, 1, 0, 0, 0, 369, 370, 3, 48, 24, 0, 370, 371, 5, 14,
		0, 0, 371, 372, 3, 50, 25, 0, 372, 382, 1, 0, 0, 0, 373, 374, 3, 48, 24,
		0, 374, 375, 5, 10, 0, 0, 375, 376, 3, 50, 25, 0, 376, 382, 1, 0, 0, 0,
		377, 378, 3, 48, 24, 0, 378, 379, 5, 9, 0, 0, 379, 380, 3, 50, 25, 0, 380,
		382, 1, 0, 0, 0, 381, 314, 1, 0, 0, 0, 381, 319, 1, 0, 0, 0, 381, 321,
		1, 0, 0, 0, 381, 330, 1, 0, 0, 0, 381, 339, 1, 0, 0, 0, 381, 348, 1, 0,
		0, 0, 381, 357, 1, 0, 0, 0, 381, 361, 1, 0, 0, 0, 381, 365, 1, 0, 0, 0,
		381, 369, 1, 0, 0, 0, 381, 373, 1, 0, 0, 0, 381, 377, 1, 0, 0, 0, 382,
		391, 1, 0, 0, 0, 383, 384, 10, 12, 0, 0, 384, 385, 5, 39, 0, 0, 385, 390,
		3, 42, 21, 13, 386, 387, 10, 11, 0, 0, 387, 388, 5, 40, 0, 0, 388, 390,
		3, 42, 21, 12, 389, 383, 1, 0, 0, 0, 389, 386, 1, 0, 0, 0, 390, 393, 1,
		0, 0, 0, 391, 389, 1, 0, 0, 0, 391, 392, 1, 0, 0, 0, 392, 43, 1, 0, 0,
		0, 393, 391, 1, 0, 0, 0, 394, 395, 5, 58, 0, 0, 395, 45, 1, 0, 0, 0, 396,
		397, 5, 16, 0, 0, 397, 402, 5, 58, 0, 0, 398, 399, 5, 1, 0, 0, 399, 401,
		5, 58, 0, 0, 400, 398, 1, 0, 0, 0, 401, 404, 1, 0, 0, 0, 402, 400, 1, 0,
		0, 0, 402, 403, 1, 0, 0, 0, 403, 47, 1, 0, 0, 0, 404, 402, 1, 0, 0, 0,
		405, 410, 5, 58, 0, 0, 406, 407, 5, 11, 0, 0, 407, 409, 5, 58, 0, 0, 408,
		406, 1, 0, 0, 0, 409, 412, 1, 0, 0, 0, 410, 408, 1, 0, 0, 0, 410, 411,
		1, 0, 0, 0, 411, 419, 1, 0, 0, 0, 412, 410, 1, 0, 0, 0, 413, 414, 5, 58,
		0, 0, 414, 415, 5, 18, 0, 0, 415, 416, 3, 44, 22, 0, 416, 417, 5, 19, 0,
		0, 417, 419, 1, 0, 0, 0, 418, 405, 1, 0, 0, 0, 418, 413, 1, 0, 0, 0, 419,
		49, 1, 0, 0, 0, 420, 421, 7, 2, 0, 0, 421, 51, 1, 0, 0, 0, 422, 424, 5,
		26, 0, 0, 423, 425, 3, 54, 27, 0, 424, 423, 1, 0, 0, 0, 424, 425, 1, 0,
		0, 0, 425, 53, 1, 0, 0, 0, 426, 428, 5, 57, 0, 0, 427, 426, 1, 0, 0, 0,
		427, 428, 1, 0, 0, 0, 428, 429, 1, 0, 0, 0, 429, 431, 5, 27, 0, 0, 430,
		432, 5, 57, 0, 0, 431, 430, 1, 0, 0, 0, 431, 432, 1, 0, 0, 0, 432, 435,
		1, 0, 0, 0, 433, 435, 5, 57, 0, 0, 434, 427, 1, 0, 0, 0, 434, 433, 1, 0,
		0, 0, 435, 55, 1, 0, 0, 0, 54, 59, 65, 73, 82, 87, 90, 96, 105, 116, 124,
		129, 132, 135, 143, 148, 151, 154, 157, 163, 171, 176, 179, 182, 185, 191,
		199, 204, 207, 210, 213, 219, 223, 230, 236, 241, 250, 256, 258, 264, 267,
		274, 282, 295, 305, 381, 389, 391, 402, 410, 418, 424, 427, 431, 434,
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
	CypherParserSQUOTE     = 25
	CypherParserSTAR       = 26
	CypherParserDOUBLE_DOT = 27
	CypherParserCREATE     = 28
	CypherParserDELETE     = 29
	CypherParserORDER_BY   = 30
	CypherParserASC        = 31
	CypherParserDESC       = 32
	CypherParserLIMIT      = 33
	CypherParserOPTIONAL   = 34
	CypherParserUNWIND     = 35
	CypherParserFINISH     = 36
	CypherParserSET        = 37
	CypherParserEQ         = 38
	CypherParserAND        = 39
	CypherParserOR         = 40
	CypherParserNOT        = 41
	CypherParserXOR        = 42
	CypherParserCOUNT      = 43
	CypherParserREDUCE     = 44
	CypherParserSUM        = 45
	CypherParserAVG        = 46
	CypherParserMIN        = 47
	CypherParserMAX        = 48
	CypherParserCOALESCE   = 49
	CypherParserIN         = 50
	CypherParserALL        = 51
	CypherParserANY        = 52
	CypherParserNONE       = 53
	CypherParserSINGLE     = 54
	CypherParserCALL       = 55
	CypherParserSTRING     = 56
	CypherParserNUMBER     = 57
	CypherParserIDENTIFIER = 58
	CypherParserWS         = 59
)

// CypherParser rules.
const (
	CypherParserRULE_cypher        = 0
	CypherParserRULE_statement     = 1
	CypherParserRULE_matchClause   = 2
	CypherParserRULE_returnClause  = 3
	CypherParserRULE_whereClause   = 4
	CypherParserRULE_callClause    = 5
	CypherParserRULE_asClause      = 6
	CypherParserRULE_pattern       = 7
	CypherParserRULE_node          = 8
	CypherParserRULE_relationship  = 9
	CypherParserRULE_returnItems   = 10
	CypherParserRULE_returnItem    = 11
	CypherParserRULE_aggregateFunc = 12
	CypherParserRULE_aggArg        = 13
	CypherParserRULE_orderItems    = 14
	CypherParserRULE_orderItem     = 15
	CypherParserRULE_limitNum      = 16
	CypherParserRULE_labels        = 17
	CypherParserRULE_label         = 18
	CypherParserRULE_properties    = 19
	CypherParserRULE_property      = 20
	CypherParserRULE_condition     = 21
	CypherParserRULE_variable      = 22
	CypherParserRULE_types         = 23
	CypherParserRULE_expression    = 24
	CypherParserRULE_value         = 25
	CypherParserRULE_range         = 26
	CypherParserRULE_rangeLiteral  = 27
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
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == CypherParserMATCH {
		{
			p.SetState(56)
			p.Statement()
		}

		p.SetState(59)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(61)
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
	MatchClause() IMatchClauseContext
	ReturnClause() IReturnClauseContext
	WhereClause() IWhereClauseContext

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

func (s *StatementContext) MatchClause() IMatchClauseContext {
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

func (s *StatementContext) WhereClause() IWhereClauseContext {
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
	{
		p.SetState(63)
		p.MatchClause()
	}
	p.SetState(65)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserWHERE {
		{
			p.SetState(64)
			p.WhereClause()
		}

	}
	{
		p.SetState(67)
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

// IMatchClauseContext is an interface to support dynamic dispatch.
type IMatchClauseContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	MATCH() antlr.TerminalNode
	Pattern() IPatternContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

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

func (s *MatchClauseContext) Pattern() IPatternContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPatternContext); ok {
			t = ctx.(antlr.RuleContext)
			break
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
	p.EnterRule(localctx, 4, CypherParserRULE_matchClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(69)
		p.Match(CypherParserMATCH)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(73)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(70)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(75)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(76)
		p.Pattern()
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
	p.EnterRule(localctx, 6, CypherParserRULE_returnClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(78)
		p.Match(CypherParserRETURN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(82)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(79)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(84)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(85)
		p.ReturnItems()
	}
	p.SetState(87)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserORDER_BY {
		{
			p.SetState(86)
			p.OrderItems()
		}

	}
	p.SetState(90)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLIMIT {
		{
			p.SetState(89)
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
	p.EnterRule(localctx, 8, CypherParserRULE_whereClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(92)
		p.Match(CypherParserWHERE)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(96)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(93)
			p.Match(CypherParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(98)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(99)
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
	p.EnterRule(localctx, 10, CypherParserRULE_callClause)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(101)
		p.Match(CypherParserCALL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(105)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserWS {
		{
			p.SetState(102)
			p.Match(CypherParserWS)
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
	}
	{
		p.SetState(108)
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
	p.EnterRule(localctx, 12, CypherParserRULE_asClause)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(110)
		p.Match(CypherParserAS)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(111)
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
	p.EnterRule(localctx, 14, CypherParserRULE_pattern)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(116)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserIDENTIFIER {
		{
			p.SetState(113)
			p.Variable()
		}
		{
			p.SetState(114)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(118)
		p.Node()
	}
	p.SetState(124)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserLARROW || _la == CypherParserMINUS {
		{
			p.SetState(119)
			p.Relationship()
		}
		{
			p.SetState(120)
			p.Node()
		}

		p.SetState(126)
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
	p.EnterRule(localctx, 16, CypherParserRULE_node)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(127)
		p.Match(CypherParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(129)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserIDENTIFIER {
		{
			p.SetState(128)
			p.Variable()
		}

	}
	p.SetState(132)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserCOLON {
		{
			p.SetState(131)
			p.Labels()
		}

	}
	p.SetState(135)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserLCURLY {
		{
			p.SetState(134)
			p.Properties()
		}

	}
	{
		p.SetState(137)
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
	p.EnterRule(localctx, 18, CypherParserRULE_relationship)
	var _la int

	p.SetState(223)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 31, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(139)
			p.Match(CypherParserMINUS)
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

		for _la == CypherParserWS {
			{
				p.SetState(140)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(145)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(146)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(148)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(147)
				p.Variable()
			}

		}
		p.SetState(151)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(150)
				p.Types()
			}

		}
		p.SetState(154)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(153)
				p.Range_()
			}

		}
		p.SetState(157)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(156)
				p.Properties()
			}

		}
		{
			p.SetState(159)
			p.Match(CypherParserRSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(163)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(160)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(165)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(166)
			p.Match(CypherParserRARROW)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(167)
			p.Match(CypherParserLARROW)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(171)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(168)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(173)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(174)
			p.Match(CypherParserLSQUARE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(176)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(175)
				p.Variable()
			}

		}
		p.SetState(179)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(178)
				p.Types()
			}

		}
		p.SetState(182)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(181)
				p.Range_()
			}

		}
		p.SetState(185)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(184)
				p.Properties()
			}

		}
		{
			p.SetState(187)
			p.Match(CypherParserRSQUARE)
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

		for _la == CypherParserWS {
			{
				p.SetState(188)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(193)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(194)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(195)
			p.Match(CypherParserMINUS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(199)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserWS {
			{
				p.SetState(196)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(201)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(202)
			p.Match(CypherParserLSQUARE)
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

		if _la == CypherParserIDENTIFIER {
			{
				p.SetState(203)
				p.Variable()
			}

		}
		p.SetState(207)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserCOLON {
			{
				p.SetState(206)
				p.Types()
			}

		}
		p.SetState(210)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserSTAR {
			{
				p.SetState(209)
				p.Range_()
			}

		}
		p.SetState(213)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserLCURLY {
			{
				p.SetState(212)
				p.Properties()
			}

		}
		{
			p.SetState(215)
			p.Match(CypherParserRSQUARE)
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

		for _la == CypherParserWS {
			{
				p.SetState(216)
				p.Match(CypherParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(221)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(222)
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
	p.EnterRule(localctx, 20, CypherParserRULE_returnItems)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(225)
		p.ReturnItem()
	}
	p.SetState(230)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(226)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(227)
			p.ReturnItem()
		}

		p.SetState(232)
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
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	COALESCE() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
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

func (s *ReturnItemContext) COALESCE() antlr.TerminalNode {
	return s.GetToken(CypherParserCOALESCE, 0)
}

func (s *ReturnItemContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(CypherParserLPAREN, 0)
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
	p.EnterRule(localctx, 22, CypherParserRULE_returnItem)
	var _la int

	p.SetState(258)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case CypherParserCOUNT, CypherParserSUM, CypherParserAVG, CypherParserMIN, CypherParserMAX:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(233)
			p.AggregateFunc()
		}
		p.SetState(236)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(234)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(235)
				p.Variable()
			}

		}

	case CypherParserIDENTIFIER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(238)
			p.Expression()
		}
		p.SetState(241)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(239)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(240)
				p.Variable()
			}

		}

	case CypherParserCOALESCE:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(243)
			p.Match(CypherParserCOALESCE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(244)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(245)
			p.Expression()
		}
		p.SetState(250)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserCOMMA {
			{
				p.SetState(246)
				p.Match(CypherParserCOMMA)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(247)
				p.Expression()
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
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(256)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserAS {
			{
				p.SetState(254)
				p.Match(CypherParserAS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(255)
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
	p.EnterRule(localctx, 24, CypherParserRULE_aggregateFunc)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(260)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&536561674354688) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(261)
		p.Match(CypherParserLPAREN)
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

	switch p.GetTokenStream().LA(1) {
	case CypherParserSTAR:
		{
			p.SetState(262)
			p.Match(CypherParserSTAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case CypherParserDISTINCT, CypherParserIDENTIFIER:
		p.SetState(264)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserDISTINCT {
			{
				p.SetState(263)
				p.Match(CypherParserDISTINCT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(266)
			p.AggArg()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(269)
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
	p.EnterRule(localctx, 26, CypherParserRULE_aggArg)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(271)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(274)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDOT {
		{
			p.SetState(272)
			p.Match(CypherParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(273)
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
	p.EnterRule(localctx, 28, CypherParserRULE_orderItems)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(276)
		p.Match(CypherParserORDER_BY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(277)
		p.OrderItem()
	}
	p.SetState(282)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(278)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(279)
			p.OrderItem()
		}

		p.SetState(284)
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
	p.EnterRule(localctx, 30, CypherParserRULE_orderItem)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(285)
		p.Expression()
	}
	{
		p.SetState(286)
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
	p.EnterRule(localctx, 32, CypherParserRULE_limitNum)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(288)
		p.Match(CypherParserLIMIT)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(289)
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
	p.EnterRule(localctx, 34, CypherParserRULE_labels)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(291)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(293)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == CypherParserIDENTIFIER {
		{
			p.SetState(292)
			p.Label()
		}

		p.SetState(295)
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
	p.EnterRule(localctx, 36, CypherParserRULE_label)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(297)
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
	p.EnterRule(localctx, 38, CypherParserRULE_properties)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(299)
		p.Match(CypherParserLCURLY)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(300)
		p.Property()
	}
	p.SetState(305)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserCOMMA {
		{
			p.SetState(301)
			p.Match(CypherParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(302)
			p.Property()
		}

		p.SetState(307)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(308)
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
	p.EnterRule(localctx, 40, CypherParserRULE_property)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(310)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(311)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(312)
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
	_startState := 42
	p.EnterRecursionRule(localctx, 42, CypherParserRULE_condition, _p)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(381)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 44, p.GetParserRuleContext()) {
	case 1:
		localctx = NewConditionParenContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(315)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(316)
			p.condition(0)
		}
		{
			p.SetState(317)
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
			p.SetState(319)
			p.Match(CypherParserNOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(320)
			p.condition(13)
		}

	case 3:
		localctx = NewConditionAllContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(321)
			p.Match(CypherParserALL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(322)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(323)
			p.Variable()
		}
		{
			p.SetState(324)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(325)
			p.Expression()
		}
		{
			p.SetState(326)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(327)
			p.condition(0)
		}
		{
			p.SetState(328)
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
			p.SetState(330)
			p.Match(CypherParserANY)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(331)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(332)
			p.Variable()
		}
		{
			p.SetState(333)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(334)
			p.Expression()
		}
		{
			p.SetState(335)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(336)
			p.condition(0)
		}
		{
			p.SetState(337)
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
			p.SetState(339)
			p.Match(CypherParserNONE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(340)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(341)
			p.Variable()
		}
		{
			p.SetState(342)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(343)
			p.Expression()
		}
		{
			p.SetState(344)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(345)
			p.condition(0)
		}
		{
			p.SetState(346)
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
			p.SetState(348)
			p.Match(CypherParserSINGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(349)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(350)
			p.Variable()
		}
		{
			p.SetState(351)
			p.Match(CypherParserIN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(352)
			p.Expression()
		}
		{
			p.SetState(353)
			p.Match(CypherParserWHERE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(354)
			p.condition(0)
		}
		{
			p.SetState(355)
			p.Match(CypherParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewConditionEqualityContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(357)
			p.Expression()
		}
		{
			p.SetState(358)
			p.Match(CypherParserEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(359)
			p.Value()
		}

	case 8:
		localctx = NewConditionNotEqualityContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(361)
			p.Expression()
		}
		{
			p.SetState(362)
			p.Match(CypherParserNEQ)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(363)
			p.Value()
		}

	case 9:
		localctx = NewConditionGreaterContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(365)
			p.Expression()
		}
		{
			p.SetState(366)
			p.Match(CypherParserRANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(367)
			p.Value()
		}

	case 10:
		localctx = NewConditionLessContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(369)
			p.Expression()
		}
		{
			p.SetState(370)
			p.Match(CypherParserLANGLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(371)
			p.Value()
		}

	case 11:
		localctx = NewConditionGreaterEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(373)
			p.Expression()
		}
		{
			p.SetState(374)
			p.Match(CypherParserGE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(375)
			p.Value()
		}

	case 12:
		localctx = NewConditionLessEqualContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(377)
			p.Expression()
		}
		{
			p.SetState(378)
			p.Match(CypherParserLE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(379)
			p.Value()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(391)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 46, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(389)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 45, p.GetParserRuleContext()) {
			case 1:
				localctx = NewConditionAndContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_condition)
				p.SetState(383)

				if !(p.Precpred(p.GetParserRuleContext(), 12)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 12)", ""))
					goto errorExit
				}
				{
					p.SetState(384)
					p.Match(CypherParserAND)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(385)
					p.condition(13)
				}

			case 2:
				localctx = NewConditionOrContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, CypherParserRULE_condition)
				p.SetState(386)

				if !(p.Precpred(p.GetParserRuleContext(), 11)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 11)", ""))
					goto errorExit
				}
				{
					p.SetState(387)
					p.Match(CypherParserOR)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(388)
					p.condition(12)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(393)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 46, p.GetParserRuleContext())
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
	p.EnterRule(localctx, 44, CypherParserRULE_variable)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(394)
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
	p.EnterRule(localctx, 46, CypherParserRULE_types)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(396)
		p.Match(CypherParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(397)
		p.Match(CypherParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(402)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == CypherParserT__0 {
		{
			p.SetState(398)
			p.Match(CypherParserT__0)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(399)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(404)
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
	p.EnterRule(localctx, 48, CypherParserRULE_expression)
	var _la int

	p.SetState(418)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 49, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(405)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(410)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == CypherParserDOT {
			{
				p.SetState(406)
				p.Match(CypherParserDOT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}
			{
				p.SetState(407)
				p.Match(CypherParserIDENTIFIER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(412)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(413)
			p.Match(CypherParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(414)
			p.Match(CypherParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(415)
			p.Variable()
		}
		{
			p.SetState(416)
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
	p.EnterRule(localctx, 50, CypherParserRULE_value)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(420)
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
	p.EnterRule(localctx, 52, CypherParserRULE_range)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(422)
		p.Match(CypherParserSTAR)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(424)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == CypherParserDOUBLE_DOT || _la == CypherParserNUMBER {
		{
			p.SetState(423)
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
	p.EnterRule(localctx, 54, CypherParserRULE_rangeLiteral)
	var _la int

	p.SetState(434)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 53, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		p.SetState(427)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserNUMBER {
			{
				p.SetState(426)
				p.Match(CypherParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(429)
			p.Match(CypherParserDOUBLE_DOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(431)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == CypherParserNUMBER {
			{
				p.SetState(430)
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
			p.SetState(433)
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
	case 21:
		var t *ConditionContext = nil
		if localctx != nil {
			t = localctx.(*ConditionContext)
		}
		return p.Condition_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *CypherParser) Condition_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 12)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 11)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
