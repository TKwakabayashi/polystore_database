package bench

// ============================================================================
// 【一時的な検証ハーネス】migrator の動作確認用。
//
// 目的:
//   1. データ消失しないこと（Delete が走らず、移行後もソースが無傷）
//   2. 移行先へ確実にコピーされること（件数一致）
//   3. マッピング切替・ラウンドトリップの健全性
//
// 方式: 各エンティティを src → dest → src の非破壊ラウンドトリップで移行し、
//       各段の件数を独立に数えて検証する（DeleteSource=false のため元データは消えない）。
//
// 使い方:
//   go run . -mode setup                                 # 先にデータ投入（ソース=graph 想定）
//   go run . -mode verify-migrate -workload IS4 -migmode graph_to_rdb
//   ↑ -migmode を graph_to_kvs / graph_to_doc / graph_to_col に変えて各ストアを検証。
//
// 動作確認が済んだらこのファイルと main.go の "verify-migrate" ケースは削除してよい。
// ============================================================================

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"polystore_database/src/go/migrator"
	"polystore_database/src/go/storage"
	"polystore_database/src/go/workload"
)

// RunMigrationVerify は workload の各 migration 設定について、
// fwd（例: graph_to_rdb）方向の移行と逆方向の復元を行い、件数ベースで検証する。
func RunMigrationVerify(ctx context.Context, cfg storage.Config, name string, fwd migrator.MigrationMode) {
	def, ok := workload.Registry[name]
	if !ok {
		log.Fatalf("未知のワークロード %q (利用可能: %s)", name, workload.AvailableWorkloads())
	}
	_, _, migs := def(fwd, true)
	if len(migs) == 0 {
		log.Fatalf("ワークロード %s に migration 設定がありません", name)
	}
	back, err := reverseMode(fwd)
	if err != nil {
		log.Fatalf("%v", err)
	}
	srcKind, destKind, err := migrator.ModeStores(fwd)
	if err != nil {
		log.Fatalf("不正な migmode %q: %v", fwd, err)
	}

	fmt.Printf("========== Migration 検証: workload=%s, %s (%s → %s) ==========\n",
		name, fwd, srcKind, destKind)

	allPass := true
	for _, mig := range migs {
		mig.MappingPath = cfg.MappingPath
		if cfg.Mongo != nil {
			mig.MongoDbName = cfg.Mongo.DBName
		}
		mig.DeleteSource = false // 明示: ソースは消さない
		if !verifyOne(ctx, cfg, mig, srcKind, destKind, fwd, back) {
			allPass = false
		}
		fmt.Println()
	}

	if allPass {
		fmt.Println("🎉 ALL PASS: データ消失なし・コピー整合・ラウンドトリップ一致")
	} else {
		fmt.Println("❌ SOME CHECKS FAILED（上のログを確認してください）")
		os.Exit(1)
	}
}

func verifyOne(ctx context.Context, cfg storage.Config, mig migrator.MigrationConfig,
	srcKind, destKind storage.StoreKind, fwd, back migrator.MigrationMode) bool {

	fmt.Printf("── %s (%s → %s) ──\n", mig.Entity, srcKind, destKind)

	srcBefore, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		fmt.Printf("  ❌ 事前カウント失敗: %v\n", err)
		return false
	}
	fmt.Printf("  src(%s) 移行前 = %d 件\n", srcKind, srcBefore)
	if srcBefore == 0 {
		fmt.Printf("  ⚠️ ソースが空です。setup 済みか、-migmode のソースが正しいか確認してください（この項目はスキップ）。\n")
		return true
	}

	// --- 順方向: src → dest ---
	fmig := mig
	fmig.Mode = fwd
	t0 := time.Now()
	if _, err := migrator.MigrateData(fmig, cfg); err != nil {
		fmt.Printf("  ❌ 順方向 migrate 失敗: %v\n", err)
		return false
	}
	fwdDur := time.Since(t0)

	srcAfter, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		fmt.Printf("  ❌ 移行後 src カウント失敗: %v\n", err)
		return false
	}
	destAfter, err := countVia(ctx, cfg, mig, destKind)
	if err != nil {
		fmt.Printf("  ❌ 移行後 dest カウント失敗: %v\n", err)
		return false
	}

	pass := true

	// 検証1: 移行先へコピーされた（dest ≥ src移行前）
	if destAfter >= srcBefore {
		fmt.Printf("  ✅ コピー成立: dest(%s) = %d ≥ src移行前 = %d\n", destKind, destAfter, srcBefore)
	} else {
		fmt.Printf("  ❌ コピー不足: dest(%s) = %d < src移行前 = %d\n", destKind, destAfter, srcBefore)
		pass = false
	}

	// 検証2: Delete が走らず、ソースが無傷（= データ消失なし）
	if srcAfter == srcBefore {
		fmt.Printf("  ✅ ソース保全: src(%s) 移行後 = %d（Delete 未実行）\n", srcKind, srcAfter)
	} else {
		fmt.Printf("  ❌ ソースが変化: 移行前 = %d → 移行後 = %d（想定外の削除）\n", srcBefore, srcAfter)
		pass = false
	}
	fmt.Printf("  ⏱ 順方向所要時間: %v\n", fwdDur)

	// --- 逆方向: dest → src（マッピングと状態を原状復帰） ---
	bmig := mig
	bmig.Mode = back
	if _, err := migrator.MigrateData(bmig, cfg); err != nil {
		fmt.Printf("  ⚠️ 逆方向 migrate 失敗（mapping が %s のまま。setup で復旧可）: %v\n", destKind, err)
		return false
	}
	srcRestored, err := countVia(ctx, cfg, mig, srcKind)
	if err != nil {
		fmt.Printf("  ❌ 復元後 src カウント失敗: %v\n", err)
		return false
	}
	if srcRestored == srcBefore {
		fmt.Printf("  ✅ ラウンドトリップ一致: src(%s) 復元 = %d\n", srcKind, srcRestored)
	} else {
		fmt.Printf("  ❌ ラウンドトリップ不一致: %d → %d\n", srcBefore, srcRestored)
		pass = false
	}

	return pass
}

// countVia は kind ストアだけを開いて cfg.Entity を数え、すぐ閉じる。
// LevelDB はファイルロックのため、MigrateData 実行中にハンドルを保持しないよう毎回開閉する。
func countVia(ctx context.Context, cfg storage.Config, mig migrator.MigrationConfig, kind storage.StoreKind) (int64, error) {
	reg, err := storage.NewRegistryFor(ctx, cfg, kind)
	if err != nil {
		return 0, err
	}
	defer reg.Close(ctx)
	return migrator.CountEntity(ctx, mig, kind, reg)
}

// reverseMode は "a_to_b" を "b_to_a" に反転する。
func reverseMode(m migrator.MigrationMode) (migrator.MigrationMode, error) {
	parts := strings.SplitN(string(m), "_to_", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("不正な mode 形式 %q（a_to_b を想定）", m)
	}
	return migrator.MigrationMode(parts[1] + "_to_" + parts[0]), nil
}
