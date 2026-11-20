// Package stores is the boundary to the databases: the compiled query that a database runs,
// the input it receives, and the database interfaces implemented by stores/<db>.
package stores

import (
	"fmt"
	"sort"
	"strings"

	"polystore_database/src/go/base/kind"
	"polystore_database/src/go/query/ir"
)

// InParam is the parameter that carries the boundary ids (Cypher $in; the expansion point in SQL / CQL).
const InParam = "in"

// TypeFn returns a property's type name (the mapping's type), used to type literals.
type TypeFn func(kind ir.ObjKind, labels []string, prop string) string

// Query is one native query.
type Query struct {
	Store    kind.Kind
	Columns  []string  // plan column key of result column ci
	ColTypes []string  // type name per column (property columns only; used to decode values)
	ColKinds []ColKind // kind per column (id / value / segment or path)
	InKeys   []string  // id columns received from the boundary (empty for a leaf)

	// Body is the database-specific query: *TextBody (Neo4j, MySQL, ClickHouse, Cassandra),
	// *PipelineBody (MongoDB) or *KVSQuery (LevelDB, RocksDB).
	Body Body

	// Shipped boundary (ShipMode other than none). ShipCols is the order of the input table's columns
	// (those of Keys, Ship and the weight / row-number / position columns), passed as a Shipment. SQL takes
	// the table as one JSON argument at InArg; MongoDB puts it in the leading $documents of a database
	// aggregation (OnDatabase). ShipEmulated marks a query whose table join the executor emulates (the
	// query is the id-boundary query under the aggregation).
	ShipEmulated bool
	ShipCols     []string
	ShipNames    []string // table column names (order of ShipCols; field names in MongoDB's $documents)
	ShipMode     ir.ShipMode
}

// Body is a database's query text or plan. Render prints it for plan records and goldens.
type Body interface {
	Render(b *strings.Builder)
}

// TextBody is a query in text: Cypher, SQL or CQL.
type TextBody struct {
	Text   string
	Params map[string]any // named Cypher parameters
	Args   []any          // positional SQL / CQL parameters (input uuids expand at InArg)
	InArg  int            // position in Args of the input uuids (-1 if none)
}

// PipelineBody is a MongoDB aggregation.
type PipelineBody struct {
	Collection string
	Stages     []Stage // at run time `uuid: {$in: ids}` is added to the first $match (when InKeys is set)
	OnDatabase bool    // a database aggregation (starting with $documents) instead of a collection's
}

// TextBody returns q's body as text (nil for other bodies).
func (q *Query) TextBody() *TextBody { b, _ := q.Body.(*TextBody); return b }

// PipelineBody returns q's MongoDB pipeline (nil for other bodies).
func (q *Query) PipelineBody() *PipelineBody { b, _ := q.Body.(*PipelineBody); return b }

// KVPlan returns q's key-value fetch plan (nil for other bodies).
func (q *Query) KVPlan() *KVSQuery { b, _ := q.Body.(*KVSQuery); return b }

// ColKind is a result column's kind. Id columns return uuids; segment and path columns {nodes, rels} uuid lists.
type ColKind int

const (
	KindValue ColKind = iota
	KindIdent
	KindList
)

// Stage is one MongoDB pipeline stage (operator -> body), ordered for display.
type Stage struct {
	Op   string
	Body Doc
}

// Doc is an ordered MongoDB document.
type Doc []Field

// Field is one element of a Doc.
type Field struct {
	Key string
	Val any
}

// KVSQuery is a fetch plan on the key-value store:
// - Scan: uuids of all nodes of the label (a reverse-index prefix scan with PrimaryEq)
// - with InKeys: point reads for the input uuids
// - Props are read with BuildEntityKey(Label, uuid, prop) and Filters are evaluated in memory
type KVSQuery struct {
	Label     string
	Scan      bool
	PrimaryEq *KVSEq
	Props     []string
	Types     map[string]string
	Filters   []ir.Expr // column keys are "alias.prop"
	Alias     string
}

// KVSEq is an equality the reverse index can answer.
type KVSEq struct {
	Prop  string
	Value any
}

// Render prints the native query deterministically (for goldens).
func (q *Query) Render() string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s] columns=(%s)", q.Store, strings.Join(q.Columns, ", "))
	if len(q.InKeys) > 0 {
		fmt.Fprintf(&b, " in=(%s)", strings.Join(q.InKeys, ", "))
	}
	switch {
	case q.ShipEmulated:
		fmt.Fprintf(&b, " shipped=%s(emulated: executor joins the table and aggregates)", q.ShipMode)
	case q.ShipMode != ir.ShipNone:
		fmt.Fprintf(&b, " shipped=%s table=(%s)", q.ShipMode, strings.Join(q.ShipNames, ", "))
	}
	b.WriteByte('\n')
	if q.Body != nil {
		q.Body.Render(&b)
	}
	return b.String()
}

// show prints a value with its type (so goldens catch type mix-ups).
func show(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return fmt.Sprintf("%q", x)
	case Doc:
		parts := make([]string, len(x))
		for i, f := range x {
			parts[i] = f.Key + ": " + show(f.Val)
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = show(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case inMarker:
		return "<in>"
	default:
		return fmt.Sprintf("%v(%T)", x, x)
	}
}

// inMarker marks where the input uuids go in positional parameters and MongoDB pipelines.
type inMarker struct{}

// InMarker is the input uuid placeholder in Args / Pipeline.
var InMarker = inMarker{}

// Render prints a key-value fetch plan.
func (k *KVSQuery) Render(b *strings.Builder) {
	src := "ids"
	if k.Scan {
		src = "scan"
		if k.PrimaryEq != nil {
			src = fmt.Sprintf("index %s=%s", k.PrimaryEq.Prop, show(k.PrimaryEq.Value))
		}
	}
	fmt.Fprintf(b, "  KVS %s %s get=(%s)", k.Label, src, strings.Join(k.Props, ", "))
	for _, f := range k.Filters {
		fmt.Fprintf(b, " filter %s", f)
	}
	b.WriteByte('\n')
}

// Render prints a MongoDB aggregation.
func (p *PipelineBody) Render(b *strings.Builder) {
	if p.OnDatabase {
		b.WriteString("  db.aggregate([\n")
	} else {
		fmt.Fprintf(b, "  db.%s.aggregate([\n", p.Collection)
	}
	for _, s := range p.Stages {
		fmt.Fprintf(b, "    {%s: %s}\n", s.Op, show(s.Body))
	}
	b.WriteString("  ])\n")
}

// Render prints a text query with its parameters.
func (t *TextBody) Render(b *strings.Builder) {
	fmt.Fprintf(b, "  %s\n", t.Text)
	if len(t.Params) > 0 {
		keys := make([]string, 0, len(t.Params))
		for k := range t.Params {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + "=" + show(t.Params[k])
		}
		fmt.Fprintf(b, "  params: %s\n", strings.Join(parts, ", "))
	}
	if len(t.Args) > 0 {
		parts := make([]string, len(t.Args))
		for i, a := range t.Args {
			parts[i] = show(a)
		}
		fmt.Fprintf(b, "  args: %s\n", strings.Join(parts, ", "))
	}
}
