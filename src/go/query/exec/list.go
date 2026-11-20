package exec

// list.go: values of the elements of segment and path lists.

import (
	"polystore_database/src/go/base/id"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
)

// listFetch reads values for the elements of a list column (the union over all rows) and makes a value list column.
func (e *Executor) listFetch(f *physical.ListFetch, in []Row) ([]Row, error) {
	elems := func(r Row) []id.UUID {
		lv, _ := r[f.Source].(*ListVal)
		if lv == nil {
			return nil
		}
		if f.Part == ir.SegRels {
			return lv.Rels
		}
		return lv.Nodes
	}
	var keys []Key
	seen := map[id.UUID]bool{}
	for _, r := range in {
		for _, u := range elems(r) {
			if !u.Empty() && !seen[u] {
				seen[u] = true
				keys = append(keys, Key{u})
			}
		}
	}
	values, err := e.fetchValues(f, f, keys)
	if err != nil {
		return nil, err
	}
	key := f.Key()
	out := make([]Row, 0, len(in))
	for _, r := range in {
		es := elems(r)
		list := make([]any, len(es))
		for i, u := range es {
			list[i] = values[u]
		}
		nr := r.Clone()
		nr[key] = list
		out = append(out, nr)
	}
	return out, nil
}

// fetchValues reads f's property for the element ids keys (one query per element label); round trips count for owner.
func (e *Executor) fetchValues(owner physical.Node, f *physical.ListFetch, keys []Key) (map[id.UUID]any, error) {
	values := map[id.UUID]any{}
	if len(keys) == 0 {
		return values, nil
	}
	qs, err := e.lowerList(f)
	if err != nil {
		return nil, err
	}
	for _, q := range qs {
		tuples, err := e.run(owner, q, keys)
		if err != nil {
			return nil, err
		}
		for _, t := range tuples {
			u, _ := t[0].(id.UUID)
			values[u] = t[1]
		}
	}
	return values, nil
}
