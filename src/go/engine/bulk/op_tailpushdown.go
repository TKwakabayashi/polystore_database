package bulk

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/store"
)

// tailInsertBatch は一時テーブルへの 1 回の INSERT でまとめる行数。
const tailInsertBatch = 2000

// bulkTailFragment は tail 委譲形の StoreFragment を実行する。
//  1. Plan に入れ子の束縛フラグメント（spec.Source）を実行して束縛 UUID の Record 群を得る。
//  2. 単一の *sql.Conn を pin し、CREATE TEMPORARY TABLE に staging エンティティの uuid を bulk INSERT。
//  3. 一時テーブルを各エンティティの永続テーブルへ uuid で JOIN し、GROUP BY/集約/ORDER BY/LIMIT を
//     ネイティブクエリで実行して最終行を得る（tail を対象ストアのエンジンで計算）。
//  4. 一時テーブル/コレクションを破棄。
//
// 対応ストアは relational(MySQL) / document(Mongo)。それ以外は呼び出し側が Plan を通常実行する。
func bulkTailFragment(qp *Processor, o *plan.StoreFragment, spec plan.TailSpec, counter *int) ([]Row, error) {
	// 1. 束縛フラグメントを実行（record パイプラインとして別途 step 計測される）。
	recs, err := ExecuteOperatorBulk(qp, spec.Source, counter)
	if err != nil {
		return nil, err
	}

	// 2. tail を対象ストアのネイティブエンジンで実行（capability を満たすストアのみ。
	//    fusion 側で store を絞っているが、実行側でも未対応ストアは明示エラーにする）。
	var (
		rows              []Row
		loadDur, queryDur time.Duration
	)
	switch o.Store {
	case store.Relational:
		rows, loadDur, queryDur, err = runRelationalTail(qp, spec, recs)
	case store.Document:
		rows, loadDur, queryDur, err = runDocumentTail(qp, spec, recs)
	default:
		return nil, fmt.Errorf("TailFragment: unsupported store %q", o.Store)
	}
	if err != nil {
		return nil, err
	}
	// load（一時テーブル作成＋INSERT）と query（JOIN+GROUP BY 実行）を別 step として計測し、
	// 「tail 集約/ソートだけ」を load オーバーヘッドと切り分けて比較できるようにする。
	recordRowOp(qp, counter, "TailLoad["+o.Store.String()+"]", loadDur, len(recs), len(recs))
	recordRowOp(qp, counter, "TailQuery["+o.Store.String()+"]", queryDur, len(recs), len(rows))
	return rows, nil
}

// runRelationalTail は MySQL 上で tail を実行し、(結果行, load時間, query時間, err) を返す。
//   - load時間: CREATE TEMPORARY TABLE ＋ 中間 UUID の bulk INSERT（クロスストア由来のオーバーヘッド）。
//   - query時間: JOIN + GROUP BY + ORDER BY + LIMIT の SQL 実行＋結果 scan（＝tail 計算そのもの）。
func runRelationalTail(qp *Processor, o plan.TailSpec, recs []Record) ([]Row, time.Duration, time.Duration, error) {
	if qp.sqlDb == nil {
		return nil, 0, 0, fmt.Errorf("TailPushdown[relational]: sqlDb is nil")
	}
	if len(o.Entities) == 0 {
		return nil, 0, 0, fmt.Errorf("TailPushdown: staging エンティティが空")
	}

	// alias → エンティティ index（temp 列 c{i} / JOIN 別名 e{i}）。
	idxOf := make(map[string]int, len(o.Entities))
	slotIdx := make([]int, len(o.Entities)) // エンティティ i の Record スロット番号（-1=不明）
	for i, e := range o.Entities {
		idxOf[e.Alias] = i
		if s, ok := o.InputSlot.VarToSlot[e.Alias]; ok {
			slotIdx[i] = s
		} else {
			slotIdx[i] = -1
		}
	}

	// TEMPORARY TABLE は接続ローカルなので単一 conn を pin する。
	conn, err := qp.sqlDb.Conn(qp.ctx)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("TailPushdown: conn 取得失敗: %w", err)
	}
	defer conn.Close()

	tmp := fmt.Sprintf("_tp_tail_%d", time.Now().UnixNano())

	// ---- load 区間: CREATE TEMPORARY TABLE ＋ 中間 UUID の bulk INSERT ----
	loadStart := time.Now()
	colDefs := make([]string, len(o.Entities))
	for i := range o.Entities {
		colDefs[i] = fmt.Sprintf("c%d VARCHAR(255)", i)
	}
	createSQL := fmt.Sprintf("CREATE TEMPORARY TABLE %s (%s)", bt(tmp), strings.Join(colDefs, ", "))
	if _, err := conn.ExecContext(qp.ctx, createSQL); err != nil {
		return nil, 0, 0, fmt.Errorf("TailPushdown: create temp 失敗: %w", err)
	}
	defer conn.ExecContext(qp.ctx, "DROP TEMPORARY TABLE IF EXISTS "+bt(tmp))
	if err := insertTailRows(qp, conn, tmp, len(o.Entities), slotIdx, recs); err != nil {
		return nil, 0, 0, err
	}
	loadDur := time.Since(loadStart)

	// ---- query 区間: JOIN + GROUP BY + ORDER BY + LIMIT の SQL 実行＋scan（＝tail 計算そのもの） ----
	queryStart := time.Now()
	sqlStr := buildTailSQL(o, tmp, idxOf)
	rows, err := conn.QueryContext(qp.ctx, sqlStr)
	if err != nil {
		return nil, loadDur, 0, fmt.Errorf("TailPushdown[relational] error: %w\n  SQL: %s", err, sqlStr)
	}
	defer rows.Close()
	cnames, _ := rows.Columns()
	var result []Row
	for rows.Next() {
		vals := make([]interface{}, len(cnames))
		ptrs := make([]interface{}, len(cnames))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(Row, len(cnames))
		for i, c := range cnames {
			row[c] = core.CoerceScalar(vals[i])
		}
		result = append(result, row)
	}
	queryDur := time.Since(queryStart)
	return result, loadDur, queryDur, rows.Err()
}

// insertTailRows は中間 Record の各エンティティ uuid を一時テーブルへ tailInsertBatch 行単位で INSERT する。
func insertTailRows(qp *Processor, conn *sql.Conn, tmp string, nCols int, slotIdx []int, recs []Record) error {
	if len(recs) == 0 {
		return nil
	}
	colNames := make([]string, nCols)
	for i := range colNames {
		colNames[i] = fmt.Sprintf("c%d", i)
	}
	placeholderGroup := "(" + strings.TrimSuffix(strings.Repeat("?,", nCols), ",") + ")"
	prefix := fmt.Sprintf("INSERT INTO %s (%s) VALUES ", bt(tmp), strings.Join(colNames, ", "))

	flush := func(groups []string, args []interface{}) error {
		if len(groups) == 0 {
			return nil
		}
		q := prefix + strings.Join(groups, ",")
		_, err := conn.ExecContext(qp.ctx, q, args...)
		if err != nil {
			return fmt.Errorf("TailPushdown: insert 失敗: %w", err)
		}
		return nil
	}

	groups := make([]string, 0, tailInsertBatch)
	args := make([]interface{}, 0, tailInsertBatch*nCols)
	for _, r := range recs {
		groups = append(groups, placeholderGroup)
		for _, si := range slotIdx {
			if si >= 0 && si < len(r.Slots) {
				args = append(args, r.Slots[si].String())
			} else {
				args = append(args, nil)
			}
		}
		if len(groups) >= tailInsertBatch {
			if err := flush(groups, args); err != nil {
				return err
			}
			groups = groups[:0]
			args = args[:0]
		}
	}
	return flush(groups, args)
}

// buildTailSQL は tail の SELECT / JOIN / GROUP BY / ORDER BY / LIMIT を組み立てる。
// 出力列名は ReturnItem.Name（＝コーディネータ経路の行キーと一致）。
func buildTailSQL(o plan.TailSpec, tmp string, idxOf map[string]int) string {
	// SELECT
	var selects []string
	for _, it := range o.Return {
		if it.IsAggregate && it.Agg != nil {
			selects = append(selects, tailAggExpr(*it.Agg, idxOf)+" AS "+bt(it.Name))
			continue
		}
		i, ok := idxOf[it.Alias]
		if !ok {
			continue
		}
		prop := ""
		if len(it.Props) > 0 {
			prop = it.Props[0]
		}
		if prop == "" {
			selects = append(selects, fmt.Sprintf("t.c%d AS %s", i, bt(it.Name)))
			continue
		}
		selects = append(selects, fmt.Sprintf("e%d.%s AS %s", i, bt(prop), bt(it.Name)))
	}

	// FROM tmp t JOIN table_i e{i} ON t.c{i} = e{i}.uuid
	joins := make([]string, len(o.Entities))
	for i, e := range o.Entities {
		joins[i] = fmt.Sprintf("JOIN %s e%d ON t.c%d = e%d.%s", bt(e.Table), i, i, i, bt("uuid"))
	}
	q := "SELECT " + strings.Join(selects, ", ") + " FROM " + bt(tmp) + " t " + strings.Join(joins, " ")

	// GROUP BY e{i}.prop
	var groups []string
	for _, gk := range o.GroupKeys {
		if i, ok := idxOf[gk.Alias]; ok && gk.Prop != "" {
			groups = append(groups, fmt.Sprintf("e%d.%s", i, bt(gk.Prop)))
		}
	}
	if len(groups) > 0 {
		q += " GROUP BY " + strings.Join(groups, ", ")
	}

	// ORDER BY <出力別名> dir（SELECT 別名で並べる）。
	var ords []string
	for _, oi := range o.OrderItems {
		key := oi.Key
		if key == "" {
			key = oi.Alias + "." + oi.Prop
		}
		dir := "ASC"
		if oi.Direction == plan.OrderDesc {
			dir = "DESC"
		}
		ords = append(ords, bt(key)+" "+dir)
	}
	if len(ords) > 0 {
		q += " ORDER BY " + strings.Join(ords, ", ")
	}

	if o.Limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", o.Limit)
	}
	return q
}

// tailAggExpr は集約式を SQL へ翻訳する。
//   - count(*) / count(alias)(prop 無し・非 distinct) → COUNT(*)（グループ行数）。
//   - count(DISTINCT alias) → COUNT(DISTINCT t.c{i})。
//   - FN(alias.prop) / count(alias.prop) → FN(e{i}.prop)（DISTINCT 対応）。
func tailAggExpr(a plan.AggregateItem, idxOf map[string]int) string {
	fn := strings.ToUpper(a.Func.String())
	if a.Func == plan.AggCount && a.Prop == "" && !a.Distinct {
		return "COUNT(*)"
	}
	i, ok := idxOf[a.Alias]
	if !ok {
		return "COUNT(*)"
	}
	if a.Prop == "" {
		if a.Distinct {
			return fmt.Sprintf("COUNT(DISTINCT t.c%d)", i)
		}
		return "COUNT(*)"
	}
	arg := fmt.Sprintf("e%d.%s", i, bt(a.Prop))
	if a.Distinct {
		return fn + "(DISTINCT " + arg + ")"
	}
	return fn + "(" + arg + ")"
}

// bt は MySQL 識別子をバッククォートで囲む。
func bt(s string) string { return "`" + strings.ReplaceAll(s, "`", "``") + "`" }
