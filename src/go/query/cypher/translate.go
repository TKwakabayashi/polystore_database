// Package cypher is the Cypher front end: it translates the query AST into a correct but naive logical plan.
// Each MATCH becomes an ir.Pattern (pattern graph and predicates as ir expressions); query/logical decides
// the start vertex, the edge order and anchored splits.
package cypher

import (
	"fmt"

	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
)

// VarKind is the kind of a bound variable.
type VarKind int

const (
	VarNode VarKind = iota
	VarRel
	VarPath
	VarValue
	VarSeg // variable-length segment (internal; not visible to users)
)

// varInfo describes a variable in scope.
type varInfo struct {
	kind   VarKind
	labels []string // node labels / relationship types
	elem   *varInfo // element info when the value column is a list of bindings (collect(n)); UNWIND turns elements back into bindings
}

// pathDef is a path variable's parts (left to right in the pattern).
type pathDef struct {
	elems []ir.PathElem
	nodes []string // bound nodes (endpoints included)
	rels  []string // fixed-length relationships
	segs  []string // variable-length segments
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ast.ErrUnsupported, fmt.Sprintf(format, args...))
}

// Translate translates a query AST into a logical plan whose MATCH clauses are unordered ir.Pattern nodes.
// Unknown variables wrap ast.ErrSemantic; unplannable shapes wrap ast.ErrUnsupported.
func Translate(q *ast.Query) (ir.Node, error) {
	b := &builder{
		scope: map[string]varInfo{},
		avail: map[string]bool{},
		paths: map[string]pathDef{},
		bound: map[string]bool{},
		group: map[string]bool{},
	}
	for _, c := range q.Clauses {
		var err error
		switch c := c.(type) {
		case *ast.MatchClause:
			err = b.match(c)
		case *ast.ProjectionClause:
			err = b.project(c)
		case *ast.UnwindClause:
			err = b.unwind(c)
		}
		if err != nil {
			return nil, err
		}
	}
	return b.plan, nil
}

type builder struct {
	plan  ir.Node
	scope map[string]varInfo
	avail map[string]bool // columns already in the row (value and path columns)
	paths map[string]pathDef
	bound map[string]bool // bound variables (nodes, relationships, segments, values)
	group map[string]bool // group variables of quantified path patterns (lists outside the group; not planned)
}
