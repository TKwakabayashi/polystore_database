// Package lower holds the helpers shared by the databases' query compilers.
package lower

import (
	"fmt"
	"polystore_database/src/go/stores"
	"strings"

	"polystore_database/src/go/base/codec"
	"polystore_database/src/go/query/cypher/ast"
	"polystore_database/src/go/query/ir"
)

func Unsupported(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ast.ErrUnsupported, fmt.Sprintf(format, args...))
}

// Chain lists a chain of unary operators from leaf to root (nil if there is a binary operator).
func Chain(n ir.Node) []ir.Node {
	var rev []ir.Node
	for n != nil {
		rev = append(rev, n)
		cs := n.Children()
		if len(cs) == 0 {
			break
		}
		if len(cs) > 1 {
			return nil
		}
		n = cs[0]
	}
	out := make([]ir.Node, len(rev))
	for i, x := range rev {
		out[len(rev)-1-i] = x
	}
	return out
}

// Owner is the kind and labels of an alias's owner (collected from GetProperty).
type Owner struct {
	Kind   ir.ObjKind
	Labels []string
}

func Owners(n ir.Node) map[string]Owner {
	out := map[string]Owner{}
	var walk func(ir.Node)
	walk = func(n ir.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *ir.GetProperty:
			if _, ok := out[x.Alias]; !ok {
				out[x.Alias] = Owner{Kind: x.Kind, Labels: x.Labels}
			}
		case *ir.EntityScan:
			if _, ok := out[x.Alias]; !ok {
				out[x.Alias] = Owner{Kind: ir.NodeObj, Labels: x.Labels}
			}
		case *ir.Expand:
			out[x.Rel] = Owner{Kind: ir.RelObj, Labels: x.Types}
		case *ir.ExpandInto:
			out[x.Rel] = Owner{Kind: ir.RelObj, Labels: x.Types}
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(n)
	return out
}

// ColType returns the type name of value column "alias.prop".
func ColType(key string, own map[string]Owner, types stores.TypeFn) string {
	i := strings.LastIndex(key, ".")
	if i < 0 || types == nil {
		return ""
	}
	o, ok := own[key[:i]]
	if !ok {
		return ""
	}
	return types(o.Kind, o.Labels, key[i+1:])
}

// OutTypes returns the type names of the columns the subtree produces (including aggregates), used to
// decode results (drivers alone cannot tell: MySQL returns aggregates as strings, for example).
func OutTypes(n ir.Node, own map[string]Owner, types stores.TypeFn) map[string]string {
	out := map[string]string{}
	var walk func(ir.Node)
	walk = func(n ir.Node) {
		if n == nil {
			return
		}
		for _, c := range n.Children() {
			walk(c)
		}
		typeOf := func(e ir.Expr) string {
			if c, ok := e.(*ir.Column); ok {
				if t, ok := out[c.Key]; ok {
					return t
				}
				return ColType(c.Key, own, types)
			}
			return ""
		}
		switch x := n.(type) {
		case *ir.GetProperty:
			out[x.Key()] = ColType(x.Key(), own, types)
		case *ir.Projection:
			for _, it := range x.Items {
				if t := typeOf(it.Expr); t != "" {
					out[it.Name] = t
				}
			}
		case *ir.Aggregation:
			for _, it := range x.Keys {
				if t := typeOf(it.Expr); t != "" {
					out[it.Name] = t
				}
			}
			for _, a := range x.Aggs {
				switch a.Fn {
				case "count":
					out[a.Name] = "long"
				case "avg":
					out[a.Name] = "double"
				case "sum", "min", "max":
					if t := typeOf(a.Arg); t != "" {
						out[a.Name] = t
					}
				}
			}
		}
	}
	walk(n)
	return out
}

// Typed converts a literal to the column's type (unchanged if that fails or the type is unknown).
func Typed(v any, typ string) any {
	if typ == "" || v == nil {
		return v
	}
	out, err := codec.ConvertToNativeType(v, typ)
	if err != nil {
		return v
	}
	return out
}

// LitType returns the column's type name when the other side of a comparison or IN is a column.
func LitType(other ir.Expr, own map[string]Owner, types stores.TypeFn) string {
	if c, ok := other.(*ir.Column); ok {
		return ColType(c.Key, own, types)
	}
	return ""
}

func ColName(i int) string { return fmt.Sprintf("c%d", i) }

// Finish fills the result column types and checks that every column kind is set.
func Finish(q *ir.StoreQuery, out *stores.Query, types stores.TypeFn) (*stores.Query, error) {
	own := Owners(q.Ops)
	ot := OutTypes(q.Ops, own, types)
	out.ColTypes = make([]string, len(out.Columns))
	for i, c := range out.Columns {
		if c == ir.WeightCol {
			out.ColTypes[i] = "long"
			continue
		}
		if t, ok := ot[c]; ok {
			out.ColTypes[i] = t
		} else {
			out.ColTypes[i] = ColType(c, own, types)
		}
	}
	if len(out.ColKinds) != len(out.Columns) {
		return nil, Unsupported("column kinds not set for %s query", q.Store)
	}
	return out, nil
}

func Contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func PropOf(key string) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '.' {
			return key[i+1:]
		}
	}
	return key
}

func AppendNew(xs []string, s string) []string {
	for _, x := range xs {
		if x == s {
			return xs
		}
	}
	return append(xs, s)
}
