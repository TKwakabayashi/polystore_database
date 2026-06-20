package logical_plan

import (
	"fmt"
	collections "polystore_database/src/go/collections"
	parser "polystore_database/src/go/parser/output"
	plan "polystore_database/src/go/plan"
	schema "polystore_database/src/go/schema"
	"regexp"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
)

type EntityInfo struct {
	alias   string
	labels  []string
	filters []*plan.ConditionNode
}

type RelInfo struct {
	label   string // relにマルチラベルは対応していない
	alias   string
	dir     plan.Direction
	source  string
	target  string
	filters []*plan.ConditionNode

	isVarLength bool
	minHops     int
	maxHops     int
}

// ================================
// plannnerのmain関数
// ================================
type QueryPlannerListener struct {
	*parser.BaseCypherListener

	entityInfo []*EntityInfo

	relInfo []*RelInfo

	lastNodeAlias string

	stackCond         []*plan.ConditionNode
	symbolCondMapping map[string][]*plan.ConditionNode

	symbolEntTable   map[string]int   // alias to its Entity number
	sorceEntMapping  map[string][]int // alias of source entity to relationships number
	targetEntMapping map[string][]int // alias of target entity to relationships number/

	symbolRelTable map[string]int // alias to its Entity number

	varEntCounter int // number of dummy entity variables for planner
	varRelCounter int // number of dummy Relationship variables for planner

	mappingDictionary *schema.MappingDictionary

	returnItems []plan.ReturnItem
	orderItems  []plan.OrderItem
	limitNum    int
	planRoot    plan.PlanNode
}

func NewQueryPlannerListener(mappingPath string) (*QueryPlannerListener, error) {
	mappingDict, err := schema.LoadMappingDictionary(mappingPath)
	if err != nil {
		return nil, err
	}

	return &QueryPlannerListener{
		stackCond:         make([]*plan.ConditionNode, 0),
		symbolCondMapping: make(map[string][]*plan.ConditionNode),

		symbolEntTable:   make(map[string]int),
		sorceEntMapping:  make(map[string][]int),
		targetEntMapping: make(map[string][]int),

		symbolRelTable: make(map[string]int),

		mappingDictionary: mappingDict,
	}, nil
}

func (l *QueryPlannerListener) BuildPlan() error {

	for alias := range l.symbolCondMapping {
		for _, cond := range l.symbolCondMapping[alias] {
			l.fillConditionMetadata(cond)
		}
	}

	startEnt := l.getStartEntity()
	if startEnt == nil {
		return nil
	}

	visitedEnt := make(map[string]bool)
	visitedRel := make(map[string]bool)

	// 2. 開始ノードのフィルタをデータストアごとに分類
	startFilters := l.symbolCondMapping[startEnt.alias]
	storesMap := make(map[string][]*plan.ConditionNode)
	for _, cond := range startFilters {
		store := cond.DataStore
		if store == "" {
			store = "unknown"
		}
		storesMap[store] = append(storesMap[store], cond)
	}

	// 3. 最も条件が多いデータストアを特定
	bestStore := "graph"
	maxConds := -1
	for store, conds := range storesMap {
		if len(conds) > maxConds {
			maxConds = len(conds)
			bestStore = store
		}
	}

	// 4. EntityScan の作成（最も条件が多いストアを割り当て）
	var root plan.PlanNode = &plan.EntityScan{
		Alias:     startEnt.alias,
		Labels:    startEnt.labels,
		DataStore: bestStore,
		Filter:    storesMap[bestStore],
	}
	visitedEnt[startEnt.alias] = true

	// 5. 残りのデータストアのフィルタを Filter オペレータとして追加
	for store, conds := range storesMap {
		if store == bestStore {
			continue
		}
		root = &plan.Filter{
			Input:     root,
			Filter:    conds,
			Alias:     startEnt.alias,
			ObjType:   plan.Entity,
			DataStore: store,
		}
	}

	// 2. 再帰的に枝分かれを探索して Operator を積み上げる
	root = l.buildOperatorTree(startEnt.alias, root, visitedEnt, visitedRel)

	// --- ProjectionUnit の構築 ---
	unitMap := make(map[string]*plan.ProjectionUnit)

	for _, item := range l.returnItems {
		unit, ok := unitMap[item.Alias]
		if !ok {
			var labels []string
			var objType plan.ObjectType
			if idx, exists := l.symbolEntTable[item.Alias]; exists {
				labels = l.entityInfo[idx].labels
				objType = plan.Entity
			} else if idx, exists := l.symbolRelTable[item.Alias]; exists {
				labels = []string{l.relInfo[idx].label}
				objType = plan.Relationship
			}

			unit = &plan.ProjectionUnit{
				Alias:   item.Alias,
				ObjType: objType,
				Labels:  labels,
				Fetches: []plan.FetchPlan{},
			}
			unitMap[item.Alias] = unit
		}

		// itemが持つ各プロパティに対して FetchPlan を構成
		for _, propName := range item.Props {

			targetLabel := ""
			if len(unit.Labels) > 0 {
				targetLabel = unit.Labels[0]
			}

			store, dType, err := l.mappingDictionary.LookupMappingDictionary(unit.ObjType, targetLabel, propName)
			if err != nil {

			}

			foundStore := false
			for i := range unit.Fetches {
				if unit.Fetches[i].Store == store {
					unit.Fetches[i].Props = append(unit.Fetches[i].Props, propName)
					unit.Fetches[i].TypeMap[propName] = dType
					foundStore = true
					break
				}
			}

			if !foundStore {
				unit.Fetches = append(unit.Fetches, plan.FetchPlan{
					Store:   store,
					Props:   []string{propName},
					TypeMap: map[string]string{propName: dType},
				})
			}
		}
	}

	// マップからスライスに変換
	projectionUnits := make([]plan.ProjectionUnit, 0, len(unitMap))
	for _, u := range unitMap {
		projectionUnits = append(projectionUnits, *u)
	}

	// 3. 最終的な Projection を適用
	l.planRoot = &plan.Projection{
		Input:      root,
		Items:      l.returnItems,
		OrderItems: l.orderItems,
		Limit:      l.limitNum,

		Units: projectionUnits,
	}

	return nil
}

func (l *QueryPlannerListener) fillConditionMetadata(node *plan.ConditionNode) {
	if node == nil {
		return
	}

	switch node.Type {
	case plan.CondAnd, plan.CondOr:
		l.fillConditionMetadata(node.Left)
		l.fillConditionMetadata(node.Right)
	case plan.CondNot, plan.CondParen, plan.CondAll, plan.CondAny, plan.CondSingle, plan.CondNone:
		l.fillConditionMetadata(node.Child)
	case plan.CondEq, plan.CondNeq, plan.CondLess, plan.CondGreater:
		break
	default:
		fmt.Println("Unknown Condition Type")
	}

	// 1. 葉ノード（具体的な比較条件）の場合にメタデータを埋める
	if node.Alias != "" && node.Property != "" {
		var objType plan.ObjectType
		var label string

		// エイリアスから Entity か Relationship かを判定
		if idx, ok := l.symbolEntTable[node.Alias]; ok {
			objType = plan.Entity
			node.ObjType = objType
			if len(l.entityInfo[idx].labels) > 0 {
				label = l.entityInfo[idx].labels[0]
				node.Labels = l.entityInfo[idx].labels
			}
		} else if idx, ok := l.symbolRelTable[node.Alias]; ok {
			objType = plan.Relationship
			node.ObjType = objType
			label = l.relInfo[idx].label
			node.Labels = []string{label}
		}

		// 辞書から DataStore と DataType を取得
		store, dType, err := l.mappingDictionary.LookupMappingDictionary(objType, label, node.Property)
		if err == nil {
			node.DataStore = store
			node.DataType = dType
		}
	}

}

func (l *QueryPlannerListener) getStartEntity() *EntityInfo {
	if len(l.entityInfo) == 0 {
		return nil
	}
	return l.entityInfo[0]
}

func (l *QueryPlannerListener) buildOperatorTree(currentAlias string, input plan.PlanNode, vEnt map[string]bool, vRel map[string]bool) plan.PlanNode {
	currentOp := input

	// 現在のノード (currentAlias) から伸びるリレーションを取得
	relIndices := l.sorceEntMapping[currentAlias]

	for _, idx := range relIndices {
		rel := l.relInfo[idx]
		if vRel[rel.alias] {
			continue
		}
		vRel[rel.alias] = true

		targetLabels := []string{}
		if idx, ok := l.symbolEntTable[rel.target]; ok {
			if len(l.entityInfo[idx].labels) > 0 {
				targetLabels = l.entityInfo[idx].labels
			}
		}

		// Expand の構築（可変長か固定長かで分岐）
		var expandOp plan.PlanNode
		if rel.isVarLength {
			expandOp = &plan.VarLengthExpand{
				Input:        currentOp,
				Alias:        rel.alias,
				RelLabel:     rel.label,
				Dir:          rel.dir,
				SourceEntity: rel.source,
				TargetEntity: rel.target,
				TargetLabels: targetLabels,
				MinHops:      rel.minHops,
				MaxHops:      rel.maxHops,
			}
		} else {
			expandOp = &plan.Expand{
				Input:        currentOp,
				Alias:        rel.alias,
				RelLabel:     rel.label,
				Dir:          rel.dir,
				SourceEntity: rel.source,
				TargetEntity: rel.target,
				TargetLabels: targetLabels,
			}
		}
		currentOp = expandOp

		// --- Relationship へのフィルタ（データストアごとに直列化） ---
		if relFilters, ok := l.symbolCondMapping[rel.alias]; ok {
			currentOp = l.appendStoreSpecificFilters(currentOp, rel.alias, relFilters, plan.Relationship)
		}

		// ターゲットノード側の処理
		if !vEnt[rel.target] {
			vEnt[rel.target] = true

			// --- TargetEntity へのフィルタ（データストアごとに直列化） ---
			if nodeFilters, ok := l.symbolCondMapping[rel.target]; ok {
				currentOp = l.appendStoreSpecificFilters(currentOp, rel.target, nodeFilters, plan.Entity)
			}

			currentOp = l.buildOperatorTree(rel.target, currentOp, vEnt, vRel)
		}
	}

	return currentOp
}

func (l *QueryPlannerListener) appendStoreSpecificFilters(input plan.PlanNode, alias string, filters []*plan.ConditionNode, objType plan.ObjectType) plan.PlanNode {
	if len(filters) == 0 {
		return input
	}

	labels := filters[0].Labels

	// 1. DataStore ごとに条件を分類
	stores := make(map[string][]*plan.ConditionNode)
	for _, cond := range filters {
		store := cond.DataStore
		if store == "" {
			store = "unknown"
		}
		stores[store] = append(stores[store], cond)
	}

	currentOp := input

	// 2. データストアごとに Filter オペレータを生成し、直列に連結
	// 実行順序を安定させるため、特定の順序（例：アルファベット順）で処理することも検討してください
	for storeName, conds := range stores {
		currentOp = &plan.Filter{
			Input:       currentOp,
			Filter:      conds,
			Labels:      labels,
			Alias:       alias,
			ObjType:     objType,
			DataStore:   storeName,
			OutputAlias: nil, // RefinePlanで設定
		}
	}

	return currentOp
}

func (l *QueryPlannerListener) RefinePlan() {

	op := l.planRoot

	requiredNext := make(collections.Set)
	var nextInputSlot plan.SlotTable

	buildSlotTable := func(aliases []string) plan.SlotTable {
		st := plan.SlotTable{
			VarToSlot: make(map[string]int),
			SlotToVar: make([]string, len(aliases)), // 必要な数だけ確保
		}

		for i, a := range aliases {
			st.VarToSlot[a] = i // 0 から順に割り当て
			st.SlotToVar[i] = a // 逆引き用
		}
		return st
	}

	for op != nil {
		switch currentOp := op.(type) {
		case *plan.Projection:

			for _, item := range currentOp.Items {
				requiredNext.Insert(item.Alias)
			}
			currentOp.InputAlias = requiredNext.ConvertSlice()

			currentOp.InputSlot = buildSlotTable(currentOp.InputAlias)
			nextInputSlot = currentOp.InputSlot

		case *plan.Filter:
			currentOp.OutputAlias = requiredNext.ConvertSlice()
			currentOp.OutputSlot = nextInputSlot

			for _, cond := range currentOp.Filter {
				if cond.Alias != "" {
					requiredNext.Insert(cond.Alias)
				}
			}
			currentOp.InputAlias = requiredNext.ConvertSlice()

			currentOp.InputSlot = buildSlotTable(currentOp.InputAlias)
			nextInputSlot = currentOp.InputSlot

		case *plan.Expand:
			currentOp.OutputAlias = requiredNext.ConvertSlice()
			currentOp.OutputSlot = nextInputSlot

			requiredNext.Remove(currentOp.Alias)
			requiredNext.Remove(currentOp.TargetEntity)

			requiredNext.Insert(currentOp.SourceEntity)

			currentOp.InputAlias = requiredNext.ConvertSlice()

			currentOp.InputSlot = buildSlotTable(currentOp.InputAlias)
			nextInputSlot = currentOp.InputSlot

		case *plan.VarLengthExpand:
			currentOp.OutputAlias = requiredNext.ConvertSlice()
			currentOp.OutputSlot = nextInputSlot

			requiredNext.Remove(currentOp.Alias)
			requiredNext.Remove(currentOp.TargetEntity)
			requiredNext.Insert(currentOp.SourceEntity)

			currentOp.InputAlias = requiredNext.ConvertSlice()

			currentOp.InputSlot = buildSlotTable(currentOp.InputAlias)
			nextInputSlot = currentOp.InputSlot

		case *plan.EntityScan:
			currentOp.OutputAlias = requiredNext.ConvertSlice()
			currentOp.OutputSlot = nextInputSlot

			currentOp.OutputSlot = buildSlotTable(currentOp.OutputAlias)
		}

		children := op.Children()
		if len(children) > 0 {
			op = children[0]
		} else {
			op = nil
		}
	}
}

func printTree(op plan.PlanNode, indent string, isLast bool) {
	marker := "├── "
	if isLast {
		marker = "└── "
	}
	fmt.Printf("%s%s%s\n", indent, marker, op.String())

	children := op.Children()
	newIndent := indent
	if isLast {
		newIndent += "    "
	} else {
		newIndent += "│   "
	}
	for i, child := range children {
		printTree(child, newIndent, i == len(children)-1)
	}
}

func ParseQuery(query string, mappingPath string, params map[string]string) (plan.PlanNode, error) {

	paramedQuery, err := replaceParameters(query, params)
	if err != nil {
		return nil, err
	}

	inputStream := antlr.NewInputStream(paramedQuery)
	lexer := parser.NewCypherLexer(inputStream)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewCypherParser(stream)

	tree := p.Cypher()

	planner, err := NewQueryPlannerListener(mappingPath)
	if err != nil {
		fmt.Println("failed to make QueryPlannerListener:", err)
		return nil, fmt.Errorf("failed to make QueryPlannerListener:%w", err)
	}

	antlr.ParseTreeWalkerDefault.Walk(planner, tree)
	err = planner.BuildPlan()
	if err != nil {
		return nil, fmt.Errorf("failed to build plan:%w", err)
	}

	planner.RefinePlan()

	return planner.planRoot, nil
}

func replaceParameters(query string, params map[string]string) (string, error) {
	var missingParams []string

	re := regexp.MustCompile(`\$(\w+)`)

	result := re.ReplaceAllStringFunc(query, func(m string) string {
		key := m[1:]
		val, ok := params[key]
		if !ok {
			missingParams = append(missingParams, key)
			return m // 置換せずそのまま返す（または "?" などに変える）
		}
		if _, err := strconv.Atoi(val); err != nil {
			return "'" + val + "'"
		}
		return val
	})

	if len(missingParams) > 0 {
		return "", fmt.Errorf("missing query parameters: %s", strings.Join(missingParams, ", "))
	}

	return result, nil
}

func CheckLogicalPlan() {
	params := map[string]string{
		"personId":     "35184372093832",
		"countryName":  "United_Kingdom",
		"workFromYear": "2006",
		"maxDate":      "2012-05-27T00:00:00.000Z",
	}

	mappingPath := "./schema/mapping.json"
	query := "MATCH (p:Person {id: $personId})-[:KNOWS*1..2]-(other:Person)\n" +
		"      <-[:HAS_CREATOR]-(m:Message)\n" +
		"WHERE m.creationDate < $maxDate\n" +
		"RETURN other.id, other.firstName, other.lastName,\n" +
		"       m.id, coalesce(m.content, m.imageFile), m.creationDate\n" +
		"ORDER BY m.creationDate DESC, m.id ASC\n" +
		"LIMIT 20"

	fmt.Println("--- Original Query ---")
	fmt.Println(strings.TrimSpace(query))

	op, err := ParseQuery(query, mappingPath, params)
	if err != nil {
		fmt.Println("Parse Error", err)
		return
	}

	fmt.Println("\n--- Logical Query Plan ---")
	if op != nil {
		fmt.Println(op.String())
		children := op.Children()
		for i, child := range children {
			printTree(child, "", i == len(children)-1)
		}
	} else {
		fmt.Println("Failed to build plan.")
	}
}
