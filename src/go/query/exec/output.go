package exec

import (
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/query/eval"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
)

// OutputRows turns the final rows into maps of the output columns (ids as strings, segments and paths as uuid string lists).
func OutputRows(p physical.Node, rows []Row) []map[string]any {
	cols := outputColumns(p)
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		m := make(map[string]any, len(cols))
		for _, c := range cols {
			m[c] = external(r[c])
		}
		// Weighted rows are expanded to their original count only here.
		for w := eval.WeightOf(r); w > 0; w-- {
			out = append(out, m)
		}
	}
	return out
}

func outputColumns(p physical.Node) []string {
	switch x := p.(type) {
	case *physical.WithSets:
		return outputColumns(x.Input)
	case *physical.StoreQuery:
		return x.Outputs
	case *physical.Coord:
		switch op := x.Op.(type) {
		case *ir.Projection:
			names := make([]string, len(op.Items))
			for i, it := range op.Items {
				names[i] = it.Name
			}
			return names
		case *ir.Aggregation:
			var names []string
			for _, k := range op.Keys {
				names = append(names, k.Name)
			}
			for _, a := range op.Aggs {
				names = append(names, a.Name)
			}
			return names
		}
		if len(x.Inputs) > 0 {
			return outputColumns(x.Inputs[0])
		}
	}
	return nil
}

func external(v any) any {
	switch x := v.(type) {
	case id.UUID:
		return x.String()
	case *eval.ListVal:
		nodes := make([]any, len(x.Nodes))
		for i, u := range x.Nodes {
			nodes[i] = u.String()
		}
		rels := make([]any, len(x.Rels))
		for i, u := range x.Rels {
			rels[i] = u.String()
		}
		return map[string]any{"nodes": nodes, "rels": rels}
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = external(e)
		}
		return out
	}
	return v
}
