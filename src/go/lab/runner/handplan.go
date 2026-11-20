package runner

import (
	"fmt"
	"regexp"
	"slices"
	"time"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
	"polystore_database/src/go/query/physical"
	"polystore_database/src/go/stores"
)

// Hand-built plans (for experiments): the optimizer's plan rearranged in a fixed way per workload, without search.
// Used only through Tuning.Hand (query.PlanOptions.Rewrite).
//
// Traversal results are not weighted: one row per path. VP4 also puts the distinct into the traversal query.
//   - relational / document: the traversal result is shipped as a table and the rest (join, filter, aggregate,
//     sort, project) runs as one query
//   - columnar / widecolumn / kvs: no join with a table, so the rest stays on the coordinator
const handChunkRows = 20000

func handPlan(workload string, early bool) func(physical.Node, map[string]string) (physical.Node, error) {
	return func(root physical.Node, params map[string]string) (physical.Node, error) {
		g := findGraphQuery(root)
		if g == nil {
			if q, ok := root.(*physical.StoreQuery); ok && workload == "VP4" {
				q.TextRewrite = unwindStart // the traversal is not split from other stores (graph placement)
			}
			return root, nil
		}
		g.Weighted = false
		switch workload {
		case "VP1":
		case "VP2", "VP3":
			dropInputCols(g)
		case "VP4":
			dropCoordDistinct(root, g)
		default:
			return nil, fmt.Errorf("handplan: no plan for %s", workload)
		}
		s := g.Input.(*physical.StoreQuery).Store
		top := g
		if early {
			var err error
			if top, err = earlyFilter(workload, root, g); err != nil {
				return nil, err
			}
		}
		if s != kind.MySQL && s != kind.Mongo {
			return root, nil
		}
		top.Replace = true
		if s == kind.Mongo {
			return handMongo(workload, top, params, early)
		}
		return handSQL(workload, top, params, early)
	}
}

// unwindStart turns the start's `p.id IN $x` into UNWIND. Neo4j 5.26 drops a later WITH DISTINCT when a
// variable-length expansion starts from an IN index seek, so Cypher with DISTINCT takes its starts by UNWIND.
func unwindStart(text string) string {
	return unwindStartRe.ReplaceAllString(text, "UNWIND $1 AS _k MATCH (p:Person {id: _k})")
}

var unwindStartRe = regexp.MustCompile(`MATCH \(p:Person\) WHERE \(?p\.id IN (\[[^\]]*\])\)?`)

// findGraphQuery returns the graph query with input (the traversal).
func findGraphQuery(n physical.Node) *physical.StoreQuery {
	if q, ok := n.(*physical.StoreQuery); ok && q.Store == kind.Neo4j && !q.Leaf() {
		return q
	}
	for _, c := range n.Children() {
		if q := findGraphQuery(c); q != nil {
			return q
		}
	}
	return nil
}

// dropInputCols drops the input columns from the result.
func dropInputCols(g *physical.StoreQuery) {
	var out []string
	for _, c := range g.Outputs {
		keep := true
		for _, d := range g.Keys {
			keep = keep && c != d
		}
		if keep {
			out = append(out, c)
		}
	}
	g.Outputs = out
	g.Replace = true
}

// dropCoordDistinct removes the `Coord Distinct` above the traversal. The input goes in one query, so no duplicates span chunks.
func dropCoordDistinct(n physical.Node, g *physical.StoreQuery) {
	if q, ok := n.(*physical.StoreQuery); ok {
		if d, ok := q.Input.(*physical.Coord); ok && d.Inputs[0] == physical.Node(g) {
			q.Input = g
			return
		}
	}
	for _, c := range n.Children() {
		dropCoordDistinct(c, g)
	}
}

// earlyTarget is the column filtered first (a vertex in the middle of the traversal and its property).
var earlyTarget = map[string][2]string{
	"VP1": {"f", "birthday"},
	"VP2": {"m", "creationDate"},
	"VP3": {"f", "birthday"},
	"VP4": {"f", "birthday"},
}

const earlyMatch = "UNWIND $pIds AS _k MATCH (p:Person {uuid: _k}) MATCH (p)-[:KNOWS*1..2]-(f:Person)"

// earlyTraversal is the traversal Cypher with the filter results as uuid conditions: $pIds are the starts, $fIds / $mIds the filter results.
var earlyTraversal = map[string]string{
	"VP1": earlyMatch + " WHERE f.uuid IN $fIds RETURN p.uuid AS c0, f.uuid AS c1",
	"VP2": earlyMatch + " MATCH (f)<-[:HAS_CREATOR]-(m:Message) WHERE m.uuid IN $mIds RETURN f.uuid AS c0, m.uuid AS c1",
	"VP3": earlyMatch + " WHERE f.uuid IN $fIds RETURN f.uuid AS c0",
	"VP4": earlyMatch + " WHERE f.uuid IN $fIds WITH DISTINCT f RETURN f.uuid AS c0",
}

// earlyFilter moves the filter above the traversal to a filter over all rows before it (a leaf query) and passes the
// resulting uuids to the traversal: start σ -> all-rows σ -> traversal -> rest. The rest loses that Filter. It returns the new traversal query.
func earlyFilter(workload string, root physical.Node, g *physical.StoreQuery) (*physical.StoreQuery, error) {
	t := earlyTarget[workload]
	alias, prop := t[0], t[1]
	var filters []*ir.Filter
	var get *ir.GetProperty
	scan := func(n ir.Node) {
		for ; n != nil; n = ir.Child(n) {
			switch x := n.(type) {
			case *ir.GetProperty:
				if x.Alias == alias && x.Prop == prop {
					get = x
				}
			case *ir.Filter:
				if slices.Contains(ir.Refs(x), alias+"."+prop) {
					filters = append(filters, x)
				}
			}
		}
	}
	var walk func(n physical.Node)
	walk = func(n physical.Node) {
		if n == physical.Node(g) {
			return
		}
		switch x := n.(type) {
		case *physical.StoreQuery:
			scan(x.Ops)
		case *physical.Coord:
			scan(x.Op)
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(root)
	if get == nil || len(filters) == 0 {
		return nil, fmt.Errorf("handplan: %s: filter %s.%s not found", workload, alias, prop)
	}
	var ops ir.Node = ir.WithInput(get, &ir.EntityScan{Alias: alias, Labels: get.Labels})
	for _, f := range filters {
		ops = ir.WithInput(f, ops)
	}
	src := g.Input.(*physical.StoreQuery)
	sigma := &physical.StoreQuery{StoreQuery: ir.StoreQuery{Store: src.Store, Ops: ops, Outputs: []string{alias},
		ShipMode: ir.ShipHand}, Input: src, HandScan: true}

	cols := append(append([]string{}, g.Keys...), alias)
	nq := &stores.Query{Store: kind.Neo4j, Body: &stores.TextBody{Text: earlyTraversal[workload], InArg: -1}, ShipCols: cols, ShipMode: ir.ShipHand}
	for _, c := range g.Outputs {
		nq.Columns = append(nq.Columns, c)
		nq.ColTypes = append(nq.ColTypes, "")
		nq.ColKinds = append(nq.ColKinds, stores.KindIdent)
	}
	t2 := &physical.StoreQuery{StoreQuery: ir.StoreQuery{Store: kind.Neo4j, Ops: &ir.Boundary{Cols: cols}, Keys: g.Keys,
		Outputs: g.Outputs, ShipMode: ir.ShipHand}, Input: sigma, Prelowered: nq}
	swapInput(root, g, t2)
	stripFilters(root, t2, alias+"."+prop)
	return t2, nil
}

// swapInput replaces g under n with t.
func swapInput(n physical.Node, g, t *physical.StoreQuery) {
	switch x := n.(type) {
	case *physical.StoreQuery:
		if x.Input == physical.Node(g) {
			x.Input = t
			return
		}
	case *physical.Coord:
		for i, c := range x.Inputs {
			if c == physical.Node(g) {
				x.Inputs[i] = t
				return
			}
		}
	}
	for _, c := range n.Children() {
		swapInput(c, g, t)
	}
}

// stripFilters removes the Filters on col from the queries above t (they are already applied).
func stripFilters(n physical.Node, t *physical.StoreQuery, col string) {
	if n == physical.Node(t) {
		return
	}
	if q, ok := n.(*physical.StoreQuery); ok && q.Prelowered == nil {
		q.Ops = withoutFilter(q.Ops, col)
	}
	for _, c := range n.Children() {
		stripFilters(c, t, col)
	}
}

func withoutFilter(n ir.Node, col string) ir.Node {
	c := ir.Child(n)
	if c == nil {
		return n
	}
	if f, ok := n.(*ir.Filter); ok && slices.Contains(ir.Refs(f), col) {
		return withoutFilter(c, col)
	}
	return ir.WithInput(n, withoutFilter(c, col))
}

// handOutputs are the query's result columns and types.
type handCol struct{ name, typ string }

var handOutputs = map[string][]handCol{
	"VP1": {{"p.id", "long"}, {"f.id", "long"}, {"f.firstName", "string"}, {"f.lastName", "string"}},
	"VP2": {{"f.id", "long"}, {"f.lastName", "string"}, {"n", "long"}, {"avgLen", "double"}},
	"VP3": {{"f.id", "long"}, {"f.firstName", "string"}, {"f.lastName", "string"}},
	"VP4": {{"f.id", "long"}, {"f.firstName", "string"}, {"f.lastName", "string"}},
}

// handTop makes the one query over the traversal query that receives it as a table.
func handTop(workload string, g *physical.StoreQuery, s kind.Kind, nq *stores.Query, group string, chunk int) *physical.StoreQuery {
	cols := handOutputs[workload]
	nq.Store = s
	nq.ColKinds, nq.ColTypes, nq.Columns = nil, nil, nil
	for _, c := range cols {
		nq.Columns = append(nq.Columns, c.name)
		nq.ColTypes = append(nq.ColTypes, c.typ)
		nq.ColKinds = append(nq.ColKinds, stores.KindValue)
	}
	nq.ShipCols = g.Outputs
	nq.ShipMode = ir.ShipHand
	return &physical.StoreQuery{StoreQuery: ir.StoreQuery{Store: s, Ops: &ir.Boundary{Cols: g.Outputs}, Outputs: nq.Columns,
		ShipMode: ir.ShipHand}, Input: g, Prelowered: nq, HandGroup: group, HandChunk: chunk}
}

// handRange are the date range parameters.
func handRange(params map[string]string) (start, end time.Time, err error) {
	parse := func(s string) (time.Time, error) {
		if t, e := time.Parse(time.DateOnly, s); e == nil {
			return t, nil
		}
		return time.Parse(time.RFC3339, s)
	}
	if start, err = parse(params["start"]); err != nil {
		return
	}
	end, err = parse(params["end"])
	return
}

func handSQL(workload string, g *physical.StoreQuery, params map[string]string, early bool) (physical.Node, error) {
	uuidCol := "VARCHAR(38) COLLATE utf8mb4_bin PATH"
	table := func(cols string) string {
		return "JSON_TABLE(CAST(? AS JSON), '$[*]' COLUMNS(" + cols + ")) AS t"
	}
	var text, group string
	chunk := handChunkRows
	names := []string{"k0", "k1"}
	start, end, err := handRange(params)
	if err != nil {
		return nil, err
	}
	args := []any{nil, start, end}
	if early {
		args = args[:1]
	}
	// where is empty when early (the filter ran before the traversal).
	where := func(cond string) string {
		if early {
			return ""
		}
		return cond
	}
	switch workload {
	case "VP1":
		text = "SELECT p.`id` AS c0, f.`id` AS c1, f.`firstName` AS c2, f.`lastName` AS c3 FROM " +
			table("`k0` "+uuidCol+" '$[0]', `k1` "+uuidCol+" '$[1]'") +
			" JOIN `Person` AS f ON f.`uuid` = t.`k1` JOIN `Person` AS p ON p.`uuid` = t.`k0`" +
			where(" WHERE f.`birthday` >= ? AND f.`birthday` < ?")
	case "VP2":
		group = "f"
		text = "SELECT f.`id` AS c0, f.`lastName` AS c1, COUNT(*) AS c2, AVG(CAST(m.`length` AS DOUBLE)) AS c3 FROM " +
			table("`k0` "+uuidCol+" '$[0]', `k1` "+uuidCol+" '$[1]'") +
			" JOIN `Message` AS m ON m.`uuid` = t.`k1` JOIN `Person` AS f ON f.`uuid` = t.`k0`" +
			where(" WHERE m.`creationDate` >= ? AND m.`creationDate` < ?") + " GROUP BY f.`id`, f.`lastName`"
	case "VP3":
		chunk, names = 0, []string{"k0"}
		text = "SELECT f.`id` AS c0, f.`firstName` AS c1, f.`lastName` AS c2 FROM " +
			table("`k0` "+uuidCol+" '$[0]'") + " JOIN `Person` AS f ON f.`uuid` = t.`k0`" +
			where(" WHERE f.`birthday` >= ? AND f.`birthday` < ?") + " ORDER BY f.`birthday` DESC, f.`id` ASC LIMIT 100"
	case "VP4":
		chunk, names = 0, []string{"k0"}
		text = "SELECT f.`id` AS c0, f.`firstName` AS c1, f.`lastName` AS c2 FROM " +
			table("`k0` "+uuidCol+" '$[0]'") + " JOIN `Person` AS f ON f.`uuid` = t.`k0`" +
			where(" WHERE f.`birthday` >= ? AND f.`birthday` < ?")
	}
	nq := &stores.Query{Body: &stores.TextBody{Text: text, Args: args, InArg: 0}, ShipNames: names}
	return handTop(workload, g, kind.MySQL, nq, group, chunk), nil
}

func handMongo(workload string, g *physical.StoreQuery, params map[string]string, early bool) (physical.Node, error) {
	type doc = stores.Doc
	type fld = stores.Field
	st := func(op string, body doc) stores.Stage { return stores.Stage{Op: op, Body: body} }
	lookup := func(from, local, as string) []stores.Stage {
		return []stores.Stage{
			st("$lookup", doc{{Key: "from", Val: from}, {Key: "localField", Val: local}, {Key: "foreignField", Val: "uuid"}, {Key: "as", Val: as}}),
			st("$unwind", doc{{Key: "path", Val: "$" + as}}),
		}
	}
	start, end, err := handRange(params)
	if err != nil {
		return nil, err
	}
	between := func(col string) stores.Stage {
		if early { // the filter ran before the traversal
			return st("$match", doc{})
		}
		return st("$match", doc{{Key: col, Val: doc{{Key: "$gte", Val: start}, {Key: "$lt", Val: end}}}})
	}
	project := func(fs ...string) stores.Stage {
		body := doc{{Key: "_id", Val: 0}}
		for i, f := range fs {
			body = append(body, fld{Key: fmt.Sprintf("c%d", i), Val: f})
		}
		return st("$project", body)
	}
	p := []stores.Stage{st("$documents", doc{{Key: "rows", Val: nil}})}
	var group string
	chunk := handChunkRows
	names := []string{"k0", "k1"}
	switch workload {
	case "VP1":
		p = append(p, lookup("Person", "k1", "f")...)
		p = append(p, between("f.birthday"))
		p = append(p, lookup("Person", "k0", "p")...)
		p = append(p, project("$p.id", "$f.id", "$f.firstName", "$f.lastName"))
	case "VP2":
		group = "f"
		p = append(p, lookup("Message", "k1", "m")...)
		p = append(p, between("m.creationDate"))
		p = append(p, st("$group", doc{{Key: "_id", Val: "$k0"}, {Key: "n", Val: doc{{Key: "$sum", Val: 1}}}, {Key: "a", Val: doc{{Key: "$avg", Val: "$m.length"}}}}))
		p = append(p, lookup("Person", "_id", "f")...)
		p = append(p, project("$f.id", "$f.lastName", "$n", "$a"))
	case "VP3":
		chunk, names = 0, []string{"k0"}
		p = append(p, lookup("Person", "k0", "f")...)
		p = append(p, between("f.birthday"))
		p = append(p, st("$sort", doc{{Key: "f.birthday", Val: -1}, {Key: "f.id", Val: 1}}), st("$limit", nil))
		p[len(p)-1].Body = doc{{Key: "n", Val: int64(100)}}
		p = append(p, project("$f.id", "$f.firstName", "$f.lastName"))
	case "VP4":
		chunk, names = 0, []string{"k0"}
		p = append(p, lookup("Person", "k0", "f")...)
		p = append(p, between("f.birthday"))
		p = append(p, project("$f.id", "$f.firstName", "$f.lastName"))
	}
	nq := &stores.Query{Body: &stores.PipelineBody{Stages: p, OnDatabase: true}, ShipNames: names}
	return handTop(workload, g, kind.Mongo, nq, group, chunk), nil
}
