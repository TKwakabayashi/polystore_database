package plan

type ReturnItem struct {
	Name       string
	Alias      string
	Props      []string
	IsCoalesce bool

	IsAggregate bool
	Agg         *AggregateItem
}

type ProjectionUnit struct {
	Alias   string
	ObjType ObjectType
	Labels  []string

	Fetches []FetchPlan
}

type FetchPlan struct {
	Store   string
	Props   []string
	TypeMap map[string]string
}

// AggregateItem は 1 つの集約式 (count/sum/avg/min/max) を表す。
// Alias/Prop は集約対象の列。count(*) は Alias=="" && Prop=="" で表す。
type AggregateItem struct {
	Func     AggFunc
	Alias    string
	Prop     string
	Distinct bool
	OutName  string // AS 別名、なければ "count(x)" 等の関数表記
}

// GroupKey は暗黙 GROUP BY のキー列（非集約の RETURN 項目）を表す。
type GroupKey struct {
	Alias   string
	Prop    string
	OutName string
}

type OrderItem struct {
	Alias     string
	Prop      string
	Direction OrderDir

	// Key は wide row から並べ替えキーを引くための正規化キー。
	// 通常は "alias.prop"、集約別名で並べる場合はその別名。
	Key string
}
