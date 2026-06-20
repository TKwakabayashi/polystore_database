package plan

type SlotTable struct {
	VarToSlot map[string]int
	SlotToVar []string
}

func NewSlotTable() *SlotTable {
	return &SlotTable{
		VarToSlot: make(map[string]int),
	}
}
