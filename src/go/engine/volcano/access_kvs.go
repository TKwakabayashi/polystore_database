package volcano

import (
	"bytes"
	"strconv"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"

	"github.com/syndtr/goleveldb/leveldb/util"
)

// ---------- EntityScan ----------

func (p *Processor) scanKvsIDs(o *plan.EntityScan) ([]string, error) {
	if p.ldb == nil {
		return nil, nil
	}
	var ids []string
	seen := make(map[string]struct{})
	primaryEq := core.FindPrimaryEqCondition(o.Filter)

	for _, label := range o.Labels {
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

		p.countRoundTrip()
		iter := p.ldb.NewIterator(util.BytesPrefix(prefix), nil)
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
			if p.matchConditionsKVS(label, uuid, o.Filter) {
				seen[uuid] = struct{}{}
				ids = append(ids, uuid)
			}
		}
		iter.Release()
	}
	return ids, nil
}

// ---------- Filter ----------

func (p *Processor) filterKvsValid(o *plan.Filter, ids []string) (map[string]struct{}, error) {
	valid := make(map[string]struct{})
	for _, id := range ids {
		for _, label := range o.Labels {
			if p.matchConditionsKVS(label, id, o.Filter) {
				valid[id] = struct{}{}
				break
			}
		}
	}
	return valid, nil
}

// ---------- Projection fetch ----------

func (p *Processor) fetchKvsProps(ids []string, unit *plan.ProjectionUnit, fetch *plan.FetchPlan) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	if p.ldb == nil || len(ids) == 0 || len(unit.Labels) == 0 || len(fetch.Props) == 0 {
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
				p.countRoundTrip()
				valByte, err := p.ldb.Get(codec.BuildEntityKey(label, uuid, propName), nil)
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

// ---------- 条件評価（点 Get） ----------

// matchConditionsKVS は条件に現れるプロパティだけを点 Get して判定する。
func (p *Processor) matchConditionsKVS(label, uuid string, filters []*plan.ConditionNode) bool {
	for _, cond := range filters {
		if cond == nil {
			continue
		}
		p.countRoundTrip()
		valBytes, err := p.ldb.Get(codec.BuildEntityKey(label, uuid, cond.Property), nil)
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
