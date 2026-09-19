package core

import (
	"fmt"

	"polystore_database/src/go/codec"
	"polystore_database/src/go/plan"
)

// graphpred.go は ConditionNode / VarLengthFilter を Cypher の WHERE 断片へ変換する共通ヘルパ。
// 全エンジンの graph access（scan/filter/var-length）と record 融合ビルダ（graphrecord.go）が
// 同じ規約で述語を生成できるよう 1 本化する（比較演算子は core.SQLOp と共通）。

// BuildCypherPredicate は単一の葉条件（比較 / IN / IS NULL / IS NOT NULL）を
// "targetVar.prop OP $p" 形式の Cypher 断片へ変換し、params にパラメータを登録する。
// seq はパラメータ名の一意採番（呼び出し側で共有）。表現不能な型は error。
//
//	targetVar: 対象の Cypher 変数名（融合クエリでは cond.Alias、per-op では "n"/"r"）。
func BuildCypherPredicate(targetVar string, cond *plan.ConditionNode, params map[string]interface{}, seq *int) (string, error) {
	prop := cond.Property
	switch cond.Type {
	case plan.CondEq, plan.CondNeq, plan.CondGreater, plan.CondLess, plan.CondGreaterEq, plan.CondLessEq:
		p := nextParam(seq)
		params[p], _ = codec.ConvertToNativeType(cond.Value, cond.DataType)
		return fmt.Sprintf("%s.%s %s $%s", targetVar, prop, SQLOp(cond.Type), p), nil
	case plan.CondIn:
		p := nextParam(seq)
		vals := make([]interface{}, 0, len(cond.Values))
		for _, v := range cond.Values {
			cv, _ := codec.ConvertToNativeType(v, cond.DataType)
			vals = append(vals, cv)
		}
		params[p] = vals
		return fmt.Sprintf("%s.%s IN $%s", targetVar, prop, p), nil
	case plan.CondIsNull:
		return fmt.Sprintf("%s.%s IS NULL", targetVar, prop), nil
	case plan.CondIsNotNull:
		return fmt.Sprintf("%s.%s IS NOT NULL", targetVar, prop), nil
	default:
		return "", fmt.Errorf("BuildCypherPredicate: unsupported condition type %v", cond.Type)
	}
}

// BuildPathFilterClauses は可変長パス全体への述語（VarLengthFilter 群）を Cypher の WHERE 断片群へ
// 変換する。pathVar は生成パターンで束縛したパス変数名（例 "p"）。
//   - 要素述語: all|any|none|single(x IN nodes(p)|relationships(p) WHERE x.prop OP $v)
//   - 大域述語: length(p) OP $v
func BuildPathFilterClauses(pathVar string, filters []plan.VarLengthFilter, params map[string]interface{}, seq *int) ([]string, error) {
	var clauses []string
	for _, f := range filters {
		switch f.Target {
		case plan.TargetGlobalPath:
			if f.Func == "length" {
				p := nextParam(seq)
				params[p], _ = codec.ConvertToNativeType(f.Value, "int")
				clauses = append(clauses, fmt.Sprintf("length(%s) %s $%s", pathVar, SQLOp(f.CmpOp), p))
				continue
			}
			return nil, fmt.Errorf("BuildPathFilterClauses: unsupported path function %q", f.Func)
		case plan.TargetNode, plan.TargetRelationship:
			quant := f.Quant.CypherQuant()
			if quant == "" || f.Cond == nil {
				return nil, fmt.Errorf("BuildPathFilterClauses: invalid element predicate")
			}
			list := "nodes"
			if f.Target == plan.TargetRelationship {
				list = "relationships"
			}
			inner, err := BuildCypherPredicate(f.IterVar, f.Cond, params, seq)
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, fmt.Sprintf("%s(%s IN %s(%s) WHERE %s)", quant, f.IterVar, list, pathVar, inner))
		}
	}
	return clauses, nil
}

// nextParam は seq を進めつつ一意なパラメータ名 val%d を返す。
// graph access / 融合ビルダの既存規約と揃える（ids / in_* とは衝突しない）。
func nextParam(seq *int) string {
	p := fmt.Sprintf("val%d", *seq)
	*seq++
	return p
}
