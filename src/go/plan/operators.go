package plan

import (
	"fmt"
	"strings"
)

// ================================
// Structure for Operator
// ================================

type EntityScan struct {
	OutputAlias []string

	OutputSlot SlotTable

	Alias  string
	Labels []string
	Filter []*ConditionNode

	DataStore string
}

func (e *EntityScan) String() string {
	conditions := ConnectCondNode(e.Filter)
	condStr := "None"
	if conditions != nil {
		condStr = conditions.Print()
	}
	return fmt.Sprintf("EntityScan(output:%s, slot:%v) [Alias: %s, Label: %s, Store: %s, Condition: {%s}]",
		strings.Join(e.OutputAlias, ","), e.OutputSlot.VarToSlot, e.Alias, strings.Join(e.Labels, "|"), e.DataStore, condStr)
}

func (e *EntityScan) Children() []PlanNode { return nil }

// あるデータベースの、あるaliasが保持しているプロパティに対してfilterを行う
// 現行の実装ではaliasごとに行うためaliasを保持する
// 今後の実装では複数のaliasに対して、または複数にデータベースに対して行う必要がある
// OR演算にも対応するためにconditionNodeをfilter処理で操作する必要がある
type Filter struct {
	InputAlias  []string // for Interim results
	OutputAlias []string

	InputSlot  SlotTable
	OutputSlot SlotTable

	Filter []*ConditionNode
	Input  PlanNode

	Alias     string
	Labels    []string
	ObjType   ObjectType
	DataStore string
}

func (f *Filter) String() string {
	conditions := ConnectCondNode(f.Filter)
	condStr := "None"
	if conditions != nil {
		condStr = conditions.Print()
	}
	return fmt.Sprintf("Filter(input:%s, slot:%v) (output:%s, slot:%v) [Store: %s, Alias: %s, Label: %s, Condition: {%s}]",
		strings.Join(f.InputAlias, ","),
		f.InputSlot.VarToSlot,
		strings.Join(f.OutputAlias, ","),
		f.OutputSlot.VarToSlot,
		f.DataStore,

		f.Alias,
		strings.Join(f.Labels, "|"),
		condStr,
	)
}
func (f *Filter) Children() []PlanNode { return []PlanNode{f.Input} }

// Expandは１段階のみの固定長であるため、これに対するフィルターはtargetと一緒に行う
type Expand struct {
	InputAlias  []string // for Interim results
	OutputAlias []string

	InputSlot  SlotTable
	OutputSlot SlotTable

	RelLabel     string
	Alias        string
	Dir          Direction
	SourceEntity string
	TargetEntity string
	TargetLabels []string
	Input        PlanNode
}

func (e *Expand) String() string {
	var pattern string
	switch e.Dir {
	case Outgoing:
		pattern = fmt.Sprintf("-[%s:%s]->", e.Alias, e.RelLabel)
	case Incoming:
		pattern = fmt.Sprintf("<-[%s:%s]-", e.Alias, e.RelLabel)
	default:
		pattern = fmt.Sprintf("-[%s:%s]-", e.Alias, e.RelLabel)
	}

	target := e.TargetEntity
	if len(e.TargetLabels) > 0 {
		target = fmt.Sprintf("%s:%s", e.TargetEntity, strings.Join(e.TargetLabels, "|"))
	}

	return fmt.Sprintf("Expand(input:%s, slot:%v) (output:%s, slot:%v) [(%s)%s(%s)]",
		strings.Join(e.InputAlias, ","), e.InputSlot.VarToSlot, strings.Join(e.OutputAlias, ","),
		e.OutputSlot.VarToSlot, e.SourceEntity, pattern, target)
}
func (e *Expand) Children() []PlanNode { return []PlanNode{e.Input} }

type VarLengthExpand struct {
	InputAlias  []string // for Interim results
	OutputAlias []string

	InputSlot  SlotTable
	OutputSlot SlotTable

	RelLabel     string
	Alias        string
	Dir          Direction
	SourceEntity string
	TargetEntity string
	TargetLabels []string
	Input        PlanNode

	Filters []VarLengthFilter // このconditionを使って中間結果を削減する

	MinHops int
	MaxHops int
}

func (e *VarLengthExpand) String() string {
	var rangeLit string
	var pattern string

	if e.MinHops > e.MaxHops {
		panic("不正なホップ数")
	} else if e.MinHops < 0 && e.MaxHops < 0 {
		rangeLit = "*"
	} else if e.MaxHops < 0 {
		rangeLit = fmt.Sprintf("*%d..", e.MinHops)
	} else if e.MinHops < 0 {
		rangeLit = fmt.Sprintf("*..%d", e.MaxHops)
	} else if e.MinHops == e.MaxHops {
		rangeLit = fmt.Sprintf("*%d", e.MinHops)
	} else { // minHop > 0 && maxHop > 0
		rangeLit = fmt.Sprintf("*%d..%d", e.MinHops, e.MaxHops)
	}

	switch e.Dir {
	case Outgoing:
		pattern = fmt.Sprintf("-[%s:%s%s]->", e.Alias, e.RelLabel, rangeLit)
	case Incoming:
		pattern = fmt.Sprintf("<-[%s:%s%s]-", e.Alias, e.RelLabel, rangeLit)
	case Bidirectional:
		pattern = fmt.Sprintf("-[%s:%s%s]-", e.Alias, e.RelLabel, rangeLit)
	}

	target := e.TargetEntity
	if len(e.TargetLabels) > 0 {
		target = fmt.Sprintf("%s:%s", e.TargetEntity, strings.Join(e.TargetLabels, "|"))
	}

	return fmt.Sprintf("VarLengthExpand(input:%s, slot:%v) (output:%s, slot:%v) [(%s)%s(%s), Condition:{}]",
		strings.Join(e.InputAlias, ","), e.InputSlot.VarToSlot, strings.Join(e.OutputAlias, ","),
		e.OutputSlot.VarToSlot, e.SourceEntity, pattern, target)
}

func (e *VarLengthExpand) Children() []PlanNode { return []PlanNode{e.Input} }

type Projection struct {
	InputAlias []string
	InputSlot  SlotTable

	Input PlanNode
	Units []ProjectionUnit
}

func (p *Projection) String() string {
	var unitDetails []string
	for _, u := range p.Units {
		var fetchInfos []string
		for _, f := range u.Fetches {
			fetchInfos = append(fetchInfos, fmt.Sprintf("%s%v", f.Store, f.Props))
		}
		unitDetails = append(unitDetails, fmt.Sprintf("%s(%s)", u.Alias, strings.Join(fetchInfos, ", ")))
	}
	return fmt.Sprintf("Projection(input:%s, slot:%v) [Materialize: %s]",
		strings.Join(p.InputAlias, ","), p.InputSlot.VarToSlot, strings.Join(unitDetails, " | "))
}

func (p *Projection) Children() []PlanNode { return []PlanNode{p.Input} }

type Aggregate struct {
	InputAlias []string
	InputSlot  SlotTable

	GroupKeys []GroupKey
	Aggs      []AggregateItem
	Input     PlanNode
}

func (a *Aggregate) Children() []PlanNode { return []PlanNode{a.Input} }

func (a *Aggregate) String() string {
	var groups []string
	for _, g := range a.GroupKeys {
		groups = append(groups, g.Alias+"."+g.Prop)
	}
	groupStr := "∅"
	if len(groups) > 0 {
		groupStr = strings.Join(groups, ", ")
	}
	var aggs []string
	for _, ag := range a.Aggs {
		arg := "*"
		if ag.Alias != "" {
			arg = ag.Alias
			if ag.Prop != "" {
				arg += "." + ag.Prop
			}
		}
		dist := ""
		if ag.Distinct {
			dist = "DISTINCT "
		}
		aggs = append(aggs, fmt.Sprintf("%s(%s%s) AS %s", ag.Func, dist, arg, ag.OutName))
	}
	return fmt.Sprintf("Aggregate[group: %s | %s]", groupStr, strings.Join(aggs, ", "))
}

type Sort struct {
	OrderItems []OrderItem
	Input      PlanNode
}

func (s *Sort) Children() []PlanNode { return []PlanNode{s.Input} }

func (s *Sort) String() string {
	var orders []string
	for _, o := range s.OrderItems {
		dir := "ASC"
		if o.Direction == OrderDesc {
			dir = "DESC"
		}
		key := o.Key
		if key == "" {
			key = o.Alias + "." + o.Prop
		}
		orders = append(orders, fmt.Sprintf("%s %s", key, dir))
	}
	return fmt.Sprintf("Sort[%s]", strings.Join(orders, ", "))
}

type Limit struct {
	Count int
	Input PlanNode
}

func (l *Limit) Children() []PlanNode { return []PlanNode{l.Input} }
func (l *Limit) String() string       { return fmt.Sprintf("Limit[%d]", l.Count) }

type Return struct {
	Items []ReturnItem
	Input PlanNode
}

func (r *Return) Children() []PlanNode { return []PlanNode{r.Input} }

func (r *Return) String() string {
	var items []string
	for _, it := range r.Items {
		items = append(items, it.Name)
	}
	return fmt.Sprintf("Return[%s]", strings.Join(items, ", "))
}

// StorePushdown は集約(＋group/order/limit)を単一ストアへ委譲する物理演算子。
// 参照する全プロパティが1つのストアに解決できる場合にプランナが生成する。
//   - graph: パラメータ埋め込み済みの Cypher をそのまま Neo4j で実行する。
//   - 非graph: Table/Filters/GroupKeys/Aggs/... からネイティブ集約クエリを生成する（拡張点）。
type StorePushdown struct {
	Store  string            // "graph","relational","columnar","document","kvs"
	Query  string            // graph 用: 原クエリ（$param 付き。baseline と同一発行）
	Params map[string]string // graph 用: パラメータ（実行時に型付けして渡す）

	// 非graph 生成用（traversal 無しの単一スキャン集約）
	Table      string
	Filters    []*ConditionNode
	GroupKeys  []GroupKey
	Aggs       []AggregateItem
	OrderItems []OrderItem
	Limit      int
	Items      []ReturnItem
}

func (s *StorePushdown) Children() []PlanNode { return nil }

func (s *StorePushdown) String() string {
	if s.Store == "graph" {
		return fmt.Sprintf("StorePushdown[graph]: %s", strings.Join(strings.Fields(s.Query), " "))
	}
	return fmt.Sprintf("StorePushdown[%s] table=%s groups=%d aggs=%d limit=%d",
		s.Store, s.Table, len(s.GroupKeys), len(s.Aggs), s.Limit)
}
