package ir

import (
	"fmt"
	"strings"
)

// Pattern is one MATCH clause's pattern and predicates on Input, before the pattern is ordered. The language
// front end makes it; logical planning replaces it with scans, expansions, joins and filters (start vertex,
// edge order and anchored splits are its choice). It never reaches physical planning.
type Pattern struct {
	Input    Node     // rows so far (nil: none; an Argument inside an Optional)
	Bound    []string // variables bound before the pattern
	Vertices []PatternVertex
	Edges    []PatternEdge
	Preds    []PatternPred
	// UniqueRels / UniqueSegs are the relationships and segments subject to relationship uniqueness.
	UniqueRels []string
	UniqueSegs []string
	// Labels are the labels (or types) of every node and relationship variable in scope, for estimates.
	Labels map[string][]string
}

// PatternVertex is a node of the pattern (labels include those of an earlier binding).
type PatternVertex struct {
	Var    string
	Labels []string
}

// PatternEdge is a relationship of the pattern from Left to Right (in Dir), in pattern order.
type PatternEdge struct {
	Var         string
	Types       []string
	Left, Right string
	Dir         Direction
	Range       *HopRange // variable length (nil: one hop)
	SegPreds    []SegPred
	InnerLabels []string // labels every inner node of a variable-length edge has (quantified path patterns)
}

// PatternPred is a predicate of the pattern: placed once Deps (nodes, relationships and segments) are bound,
// after Fetches (GetProperty / GetListProperty / ProjectPath without input) add the columns it needs that
// the input does not have.
type PatternPred struct {
	Pred    Expr
	Deps    []string
	Fetches []Node
}

func (n *Pattern) Children() []Node { return []Node{n.Input} }

func (n *Pattern) Describe() string {
	var parts []string
	for _, e := range n.Edges {
		r := e.Var + labelSuffix(e.Types)
		if e.Range != nil {
			r += e.Range.Cypher()
		}
		parts = append(parts, fmt.Sprintf("(%s)%s(%s)", e.Left, e.Dir.Pattern(r), e.Right))
	}
	for _, v := range n.Vertices {
		parts = append(parts, "("+v.Var+labelSuffix(v.Labels)+")")
	}
	for _, p := range n.Preds {
		parts = append(parts, p.Pred.String())
	}
	return fmt.Sprintf("Pattern[%s]", strings.Join(parts, ", "))
}

// FetchKey is the column a fetch adds (GetProperty, GetListProperty or ProjectPath).
func FetchKey(n Node) string {
	switch x := n.(type) {
	case *GetProperty:
		return x.Key()
	case *GetListProperty:
		return x.Key()
	case *ProjectPath:
		return x.Path
	}
	panic(fmt.Sprintf("ir.FetchKey: %T is not a fetch", n))
}
