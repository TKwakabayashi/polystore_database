package plan

type ReturnItem struct {
	Name       string
	Alias      string
	Props      []string
	IsCoalesce bool
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

type OrderItem struct {
	Alias     string
	Prop      string
	Direction OrderDir
}
