package plan

type PlanNode interface {
	Children() []PlanNode
	String() string
}
