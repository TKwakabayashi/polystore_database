// Package query is the entry to the query engine: it plans a statement through the language
// front end, the logical and physical planners, and runs the plan with the executor.
package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/catalog"
	"polystore_database/src/go/catalog/stats"
	"polystore_database/src/go/query/cypher"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/exec"
	"polystore_database/src/go/query/logical"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/query/result"
)

// PlanOptions are the planning switches for one query.
type PlanOptions struct {
	// CostBased uses the mapping's statistics to choose anchors, splits and edge order.
	CostBased bool
	Logical   logical.Options
	Physical  physical.Options
	// Rewrite, when set, rewrites the physical plan before execution (hand-built experiment plans).
	Rewrite func(physical.Node, map[string]string) (physical.Node, error)
}

// Plan turns Cypher into a physical plan. Syntax and semantic errors wrap ast.ErrSyntax / ErrSemantic;
// a valid query that cannot be planned wraps ast.ErrUnsupported.
func Plan(stmt string, params map[string]string, md *catalog.MappingDictionary, opt PlanOptions) (physical.Node, *physical.Placement, error) {
	a, err := ast.Parse(stmt, params)
	if err != nil {
		return nil, nil, err
	}
	var st *stats.Stats
	if opt.CostBased && md != nil {
		st = md.Stats
	}
	// Unique keys are schema knowledge from the mapping, not a tuning switch.
	keys := md.UniqueKeys()
	opt.Logical.UniqueKeys, opt.Physical.UniqueKeys = keys, keys
	l, err := cypher.Translate(a)
	if err != nil {
		return nil, nil, err
	}
	cat := physical.NewPlacement(md)
	p, err := physical.Build(logical.Rewrite(logical.PlanPatterns(l, st, opt.Logical), opt.Logical), cat, opt.Physical)
	if err != nil {
		return nil, nil, err
	}
	if opt.Rewrite != nil {
		if p, err = opt.Rewrite(p, params); err != nil {
			return nil, nil, err
		}
	}
	// Check that every store query compiles, so an uncompilable shape is delegated before execution.
	if err := checkLowerable(p, cat); err != nil {
		return nil, nil, err
	}
	return p, cat, nil
}

// RenderPlan prints the physical plan and the native query sent to each store (for experiment records).
func RenderPlan(p physical.Node, cat *physical.Placement) string {
	var b strings.Builder
	b.WriteString(physical.Dump(p))
	var walk func(physical.Node)
	walk = func(n physical.Node) {
		if n == nil {
			return
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
		switch x := n.(type) {
		case *physical.StoreQuery:
			if nq, err := exec.CompileStoreQuery(x, cat.TypeOf); err == nil {
				b.WriteString(nq.Render() + "\n")
			} else {
				fmt.Fprintf(&b, "(lower: %v)\n", err)
			}
		case *physical.ListFetch:
			qs, err := exec.CompileList(x, cat.TypeOf)
			if err != nil {
				fmt.Fprintf(&b, "(lower: %v)\n", err)
			}
			for _, q := range qs {
				b.WriteString(q.Render() + "\n")
			}
		case *physical.StepExpand:
			walk(x.Hop)
			for _, f := range x.Fetches {
				walk(f)
			}
		}
	}
	walk(p)
	return b.String()
}

func checkLowerable(n physical.Node, cat *physical.Placement) error {
	switch x := n.(type) {
	case *physical.StoreQuery:
		if _, err := exec.CompileStoreQuery(x, cat.TypeOf); err != nil {
			return err
		}
	case *physical.ListFetch:
		if _, err := exec.CompileList(x, cat.TypeOf); err != nil {
			return err
		}
	case *physical.StepExpand:
		if _, err := exec.CompileStoreQuery(x.Hop, cat.TypeOf); err != nil {
			return err
		}
		for _, f := range x.Fetches {
			if _, err := exec.CompileList(f, cat.TypeOf); err != nil {
				return err
			}
		}
	}
	for _, c := range n.Children() {
		if err := checkLowerable(c, cat); err != nil {
			return err
		}
	}
	return nil
}

// Run plans and executes Cypher. Syntax and semantic errors are returned; a query that cannot be
// planned is delegated whole to Neo4j (the whole-statement case of partial delegation).
func Run(ctx context.Context, be exec.Backend, md *catalog.MappingDictionary, cypher string, params map[string]string, popt PlanOptions, eopt exec.ExecOptions) (result.Result, error) {
	model := eopt.Model
	if model == "" {
		model = exec.ModelBulk
		eopt.Model = model
	}
	t0 := time.Now()
	p, cat, err := Plan(cypher, params, md, popt)
	planT := time.Since(t0)
	if errors.Is(err, ast.ErrUnsupported) {
		start := time.Now()
		rows, verr := be.Verbatim(ctx, cypher, typedParams(params))
		if verr != nil {
			return result.Result{}, verr
		}
		for _, row := range rows { // same value types as the planned path (Neo4j Date -> time.Time, ...)
			for k, v := range row {
				row[k] = eval.Normalize(v)
			}
		}
		return result.Result{Rows: rows, PlanTime: planT, ExecTime: time.Since(start), Engine: "redesign-" + string(model) + "(verbatim)", RoundTrips: 1,
			StoreTrips: map[string]int64{kind.Neo4j.String(): 1}, ExecKind: result.ExecVerbatim}, nil
	}
	if err != nil {
		return result.Result{}, err
	}
	ex := exec.New(ctx, be, cat.TypeOf, eopt)
	start := time.Now()
	rows, err := ex.ExecuteModel(p)
	if err != nil {
		// On failure or timeout, still return the metrics of the operators that finished (for diagnosis).
		return result.Result{PlanTime: planT, ExecTime: time.Since(start), Steps: ex.Steps(),
			RoundTrips: ex.RoundTrips(), StoreTrips: ex.StoreTrips(), Engine: "redesign-" + string(model), ExecKind: result.ExecStructured}, err
	}
	res := result.Result{Rows: exec.OutputRows(p, rows), PlanTime: planT, ExecTime: time.Since(start),
		Steps: ex.Steps(), RoundTrips: ex.RoundTrips(), StoreTrips: ex.StoreTrips(), Engine: "redesign-" + string(model), ExecKind: result.ExecStructured}
	if model == exec.ModelVectorized || model == exec.ModelVecStream || model == exec.ModelStream {
		res.VectorWidth = eopt.VectorWidth()
	}
	return res, nil
}

// ParseListParam reads a JSON list parameter. Integers stay int64 (as float64 Neo4j would return other
// value types, and ids above 2^53 would change).
func ParseListParam(v string) ([]any, bool) {
	if !strings.HasPrefix(v, "[") {
		return nil, false
	}
	dec := json.NewDecoder(strings.NewReader(v))
	dec.UseNumber()
	var list []any
	if dec.Decode(&list) != nil {
		return nil, false
	}
	for i, e := range list {
		n, ok := e.(json.Number)
		if !ok {
			continue
		}
		if x, err := n.Int64(); err == nil {
			list[i] = x
		} else if f, err := n.Float64(); err == nil {
			list[i] = f
		}
	}
	return list, true
}

// typedParams types the string parameters for delegation to Neo4j (the same rules as runner.RunNeo4j).
func typedParams(params map[string]string) map[string]any {
	out := make(map[string]any, len(params))
	for k, v := range params {
		if n, err := strconv.Atoi(v); err == nil {
			out[k] = n
		} else if list, ok := ParseListParam(v); ok {
			out[k] = list
		} else if t, err := time.Parse(time.RFC3339, v); err == nil {
			out[k] = t
		} else {
			out[k] = v
		}
	}
	return out
}
