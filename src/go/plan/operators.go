package plan

import (
	"fmt"
	"strconv"
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
	InputAlias []string // for Interim results

	InputSlot SlotTable

	Items      []ReturnItem
	OrderItems []OrderItem
	Limit      int
	Input      PlanNode

	Units []ProjectionUnit
}

func (p *Projection) String() string {
	// 1. Return Items の構築
	var items []string
	for _, item := range p.Items {
		items = append(items, item.Name)
	}

	// 2. Order Items の構築
	var orders []string
	for _, o := range p.OrderItems {
		dir := "ASC"
		if o.Direction == OrderDesc {
			dir = "DESC"
		}
		orders = append(orders, fmt.Sprintf("%s.%s %s", o.Alias, o.Prop, dir))
	}
	orderStr := "None"
	if len(orders) > 0 {
		orderStr = strings.Join(orders, ", ")
	}

	// 3. Fetch Plan (ProjectionUnit) の詳細
	var unitDetails []string
	for _, u := range p.Units {
		var fetchInfos []string
		for _, f := range u.Fetches {
			// 例: relational[id, name]
			fetchInfos = append(fetchInfos, fmt.Sprintf("%s%v", f.Store, f.Props))
		}
		unitDetails = append(unitDetails, fmt.Sprintf("%s(%s)", u.Alias, strings.Join(fetchInfos, ", ")))
	}

	// 4. Limit の処理
	limitStr := "None"
	if p.Limit > 0 {
		limitStr = strconv.Itoa(p.Limit)
	}

	return fmt.Sprintf("Projection(input:%s, slot:%v) [Return: %s, OrderBy: %s, FetchPlan: %s, Limit: %s]",
		strings.Join(p.InputAlias, ","),
		p.InputSlot.VarToSlot,
		strings.Join(items, ", "),
		orderStr,
		strings.Join(unitDetails, " | "),
		limitStr)
}

func (p *Projection) Children() []PlanNode { return []PlanNode{p.Input} }
