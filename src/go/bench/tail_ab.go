package bench

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/plan"
	planner "polystore_database/src/go/planner"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// RunTailAB は tail pushdown の A/B 計測を行う（bulk エンジン固定）。
//
// 対象プロパティを rdb(MySQL) へ非破壊コピーした上で、PushdownAuto のもとで:
//   - A: TailPushdown=OFF → tail を bulk エンジンで計算（Case2 record-mode 融合）
//   - B: TailPushdown=ON  → 中間 UUID を MySQL 一時テーブルへ載せ tail をネイティブ実行
//
// を計測し、行数・結果一致（多重集合）・レイテンシ・speedup を出力する。基準に全 graph の
// Neo4j 直接も測る。計測後は mapping を graph へ復元する。
func RunTailAB(ctx context.Context, cfg storage.Config, name string) error {
	def, ok := workload.Registry[name]
	if !ok {
		return fmt.Errorf("未知のワークロード %q (利用可能: %s)", name, workload.AvailableWorkloads())
	}
	cypher, params, migs := def(migrator.ModeGraphToRdb, true)
	if len(migs) == 0 {
		return fmt.Errorf("%s に移行対象プロパティが無い（tail pushdown 検証には非 graph 配置が必要）", name)
	}

	// mapping スナップショット → 終了時に必ず graph へ復元。
	snapshot, rerr := os.ReadFile(cfg.MappingPath)
	if rerr != nil {
		return fmt.Errorf("mapping 読込失敗: %w", rerr)
	}
	restore := func() { _ = os.WriteFile(cfg.MappingPath, snapshot, 0644) }
	defer restore()

	// settings の一時退避。
	prevPushdown, prevTail := settings.Pushdown, settings.TailPushdown
	prevProfile := settings.ProfileScopes[settings.ScopeEngine]
	settings.ProfileScopes[settings.ScopeEngine] = false
	defer func() {
		settings.Pushdown = prevPushdown
		settings.TailPushdown = prevTail
		settings.ProfileScopes[settings.ScopeEngine] = prevProfile
	}()

	fmt.Printf("=== Tail pushdown A/B: %s ===\n", name)

	// 基準: 全 graph 配置での Neo4j 直接。
	var neoRows int
	var neoMs float64
	var neoCanon []string
	if cfg.Neo4j != nil {
		if r, err := RunNeo4j(ctx, *cfg.Neo4j, cypher, toValuedParams(params)); err != nil {
			fmt.Printf("[neo4j-direct] エラー: %v\n", err)
		} else {
			neoRows, neoMs, neoCanon = r.RowCount(), toMs(r), canonMultiset(r.Rows)
			fmt.Printf("  neo4j-direct(graph)      rows=%-5d %10.3f ms\n", neoRows, neoMs)
		}
	}

	// rdb へ非破壊コピー。
	pmigs := make([]migrator.MigrationConfig, len(migs))
	copy(pmigs, migs)
	for i := range pmigs {
		pmigs[i].Mode = migrator.ModeGraphToRdb
		pmigs[i].DeleteSource = false
	}
	if _, err := RunMigration(ctx, cfg, pmigs); err != nil {
		return fmt.Errorf("migration(copy→rdb) 失敗: %w", err)
	}

	settings.Pushdown = settings.PushdownAuto

	// tail pushdown が実際に発火するか（プラン形状）を確認。
	settings.TailPushdown = true
	fired := planHasTailPushdown(cypher, cfg.MappingPath, params)
	fmt.Printf("  tail pushdown 発火(plan): %v\n", fired)

	// A: OFF（tail を engine で計算）。
	settings.TailPushdown = false
	rA, errA := RunEngine(ctx, cfg, settings.EngineBulk, cypher, params)
	if errA != nil {
		return fmt.Errorf("[A OFF] 実行失敗: %w", errA)
	}
	// B: ON（tail を MySQL でネイティブ実行）。
	settings.TailPushdown = true
	rB, errB := RunEngine(ctx, cfg, settings.EngineBulk, cypher, params)
	if errB != nil {
		return fmt.Errorf("[B ON] 実行失敗: %w", errB)
	}

	restore() // これ以降 mapping は graph（保険で defer でも復元）。

	aMs, bMs := toMs(rA), toMs(rB)
	fmt.Printf("  A OFF(engine tail, rdb)  rows=%-5d %10.3f ms\n", rA.RowCount(), aMs)
	fmt.Printf("  B ON (MySQL tail, rdb)   rows=%-5d %10.3f ms\n", rB.RowCount(), bMs)

	// 結果一致（多重集合）。
	cA, cB := canonMultiset(rA.Rows), canonMultiset(rB.Rows)
	fmt.Printf("  結果一致 A==B: %v", equalCanon(cA, cB))
	if neoCanon != nil {
		fmt.Printf(" / B==neo4j: %v", equalCanon(cB, neoCanon))
	}
	fmt.Println()

	if aMs > 0 && bMs > 0 {
		fmt.Printf("  速度比 A/B(=engine tail / MySQL tail 端到端): %.2fx", aMs/bMs)
		if neoMs > 0 {
			fmt.Printf("  | B speedup vs neo4j: %.2fx", neoMs/bMs)
		}
		fmt.Println()
	}

	// ---- tail 集約/ソートだけの切り出し比較（load オーバーヘッドと traversal を除外）----
	// A(engine): Aggregate + Sort + Limit の step 時間合計。
	// B(MySQL):  TailQuery step（JOIN+GROUP BY+ORDER BY+LIMIT の SQL 実行）。TailLoad は別途 overhead。
	var engTailMs, mysqlTailMs, loadMs float64
	for _, s := range rA.Steps {
		if s.Op == "Aggregate" || s.Op == "Sort" || s.Op == "Limit" {
			engTailMs += durToMs(s.Duration)
		}
	}
	for _, s := range rB.Steps {
		switch {
		case strings.HasPrefix(s.Op, "TailQuery["):
			mysqlTailMs += durToMs(s.Duration)
		case strings.HasPrefix(s.Op, "TailLoad["):
			loadMs += durToMs(s.Duration)
		}
	}
	fmt.Println("  --- tail のみ切り出し（traversal / load を除外）---")
	fmt.Printf("  A engine tail (Aggregate+Sort+Limit): %10.3f ms\n", engTailMs)
	fmt.Printf("  B MySQL  tail (JOIN+GROUP BY SQL のみ): %10.3f ms   (別途 load overhead=%.3f ms)\n", mysqlTailMs, loadMs)
	if engTailMs > 0 && mysqlTailMs > 0 {
		fmt.Printf("  tail のみ速度比 engine/MySQL: %.2fx\n", engTailMs/mysqlTailMs)
	}
	return nil
}

// planHasTailPushdown は現行 settings 下で cypher をプランして TailPushdown ノードが現れるかを返す。
func planHasTailPushdown(cypher, mappingPath string, params map[string]string) bool {
	root, err := planner.ParseQuery(cypher, mappingPath, params)
	if err != nil {
		return false
	}
	found := false
	var walk func(n plan.PlanNode)
	walk = func(n plan.PlanNode) {
		if n == nil || found {
			return
		}
		// tail 委譲形は「Plan に束縛フラグメントが入れ子の StoreFragment」で判別する。
		if f, ok := n.(*plan.StoreFragment); ok {
			if _, isTail := plan.LowerTail(f.Plan); isTail {
				found = true
				return
			}
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(root)
	return found
}

// canonMultiset は行スライスを順序非依存で比較できる正規形へ変換する
// （各行を "k=v;" ソート連結し、行同士もソート）。値は fmt.Sprint で文字列化して型差を吸収。
func canonMultiset(rows []map[string]interface{}) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b []byte
		for _, k := range keys {
			b = append(b, k...)
			b = append(b, '=')
			b = append(b, fmt.Sprint(row[k])...)
			b = append(b, ';')
		}
		out[i] = string(b)
	}
	sort.Strings(out)
	return out
}

func equalCanon(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
