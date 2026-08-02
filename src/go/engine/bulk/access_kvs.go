package bulk

import (
	"bytes"
	"strconv"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	uid "polystore_database/src/go/id"
	"polystore_database/src/go/plan"

	"github.com/syndtr/goleveldb/leveldb/util"
)

func ScanKvsBulk(qp *Processor, o *plan.EntityScan) ([]Record, error) {
	if qp.ldb == nil {
		return nil, nil
	}

	newSlotCount := len(o.OutputSlot.VarToSlot)
	aliasIdx := o.OutputSlot.VarToSlot[o.Alias]
	seen := make(map[string]struct{})
	out := make([]Record, 0)

	emit := func(uuid string) {
		newSlots := make([]uid.UUID, newSlotCount)
		newSlots[aliasIdx] = uid.UUID(uuid)
		out = append(out, Record{Slots: newSlots})
	}

	primaryEq := core.FindPrimaryEqCondition(o.Filter)

	for _, label := range o.Labels {
		// --- DB特有: プレフィックス決定（Eq ありはインデックス、無ければ全スキャン）---
		var prefix []byte
		if primaryEq != nil {
			var valBytes []byte
			switch primaryEq.DataType {
			case "int", "integer", "long":
				v, _ := strconv.ParseInt(primaryEq.Value, 10, 64)
				valBytes = codec.EncodeInt(v)
			default:
				valBytes = []byte(primaryEq.Value)
			}
			prefix = []byte("index" + codec.Sep + label + codec.Sep + primaryEq.Property + codec.Sep)
			prefix = append(prefix, valBytes...)
			prefix = append(prefix, []byte(codec.Sep)...)
		} else {
			prefix = []byte(label + codec.Sep)
		}

		iter := qp.ldb.NewIterator(util.BytesPrefix(prefix), nil)
		for iter.Next() {
			key := iter.Key()
			var uuid string
			if primaryEq != nil {
				if idx := bytes.LastIndex(key, []byte(codec.Sep)); idx != -1 {
					uuid = string(key[idx+1:])
				}
			} else {
				parts := bytes.Split(key, []byte(codec.Sep))
				if len(parts) >= 2 {
					uuid = string(parts[1])
				}
			}
			if uuid == "" {
				continue
			}
			if _, done := seen[uuid]; done {
				continue
			}
			// 残りの条件を点 Get で確認
			if matchConditionsKVS(qp, label, uuid, o.Filter) {
				seen[uuid] = struct{}{}
				emit(uuid)
			}
		}
		iter.Release()
	}
	return out, nil
}

func FilterKvsBulk(qp *Processor, o *plan.Filter, in []Record) ([]Record, error) {
	filterIdxIn := o.InputSlot.VarToSlot[o.Alias]
	newSlotCount := len(o.OutputSlot.VarToSlot)

	idMap := make(map[uid.UUID]struct{})
	for _, r := range in {
		idMap[r.Slots[filterIdxIn]] = struct{}{}
	}

	// --- DB特有: 各 uuid を点 Get で判定 ---
	validMap := make(map[uid.UUID]struct{})
	for id := range idMap {
		for _, label := range o.Labels {
			if matchConditionsKVS(qp, label, id.String(), o.Filter) {
				validMap[id] = struct{}{}
				break
			}
		}
	}

	out := make([]Record, 0, len(in))
	for _, r := range in {
		if _, ok := validMap[r.Slots[filterIdxIn]]; ok {
			newRec := Record{Slots: make([]uid.UUID, newSlotCount)}
			for alias, outIdx := range o.OutputSlot.VarToSlot {
				if inIdx, exists := o.InputSlot.VarToSlot[alias]; exists {
					newRec.Slots[outIdx] = r.Slots[inIdx]
				}
			}
			out = append(out, newRec)
		}
	}
	return out, nil
}

func fetchKvsPropsBulk(qp *Processor, ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if qp.ldb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
		return result
	}

	for _, uuid := range ids {
		if uuid == "" {
			continue
		}
		if _, exists := result[uuid]; !exists {
			result[uuid] = make(map[string]interface{})
		}
		for _, label := range unit.Labels {
			for _, propName := range fetch.Props {
				valByte, err := qp.ldb.Get(codec.BuildEntityKey(label, uid.UUID(uuid), propName), nil)
				if err != nil {
					if _, ok := result[uuid][propName]; !ok {
						result[uuid][propName] = nil
					}
					continue
				}
				finalVal, _ := codec.ConvertToNativeType(codec.DecodeValue(valByte, fetch.TypeMap[propName]), fetch.TypeMap[propName])
				if finalVal != nil {
					result[uuid][propName] = finalVal
				} else if _, ok := result[uuid][propName]; !ok {
					result[uuid][propName] = nil
				}
			}
		}
	}
	return result
}

// matchConditionsKVS は条件に現れるプロパティだけを点 Get して判定する（全プロパティ走査しない）。
func matchConditionsKVS(qp *Processor, label, uuid string, filters []*plan.ConditionNode) bool {
	for _, cond := range filters {
		if cond == nil {
			continue
		}
		valBytes, err := qp.ldb.Get(codec.BuildEntityKey(label, uid.UUID(uuid), cond.Property), nil)
		if err != nil {
			return false // プロパティ欠落 = 不一致
		}
		actual, _ := codec.ConvertToNativeType(codec.DecodeValue(valBytes, cond.DataType), cond.DataType)
		if !core.EvalConditionKVS(actual, cond) {
			return false
		}
	}
	return true
}
