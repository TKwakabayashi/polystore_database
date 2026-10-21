package setup

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"polystore_database/src/go/id"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
)

// uuidBatch は 1 回の書き込みトランザクションで書き換える件数。
const uuidBatch = 10000

// ConvertUUIDFormat は Neo4j の全ノード・全関係の uuid を、保存形式 settings.UUIDStorage（文字列 / 17 バイト）へ
// 書き換える。値は id の境界変換で 1 対 1 に写すので、uuid の一意制約は書き換えの途中も保たれる
// （文字列とバイト列は等しくならない）。既に目的の形式の値は飛ばすので、途中で止めても再実行で続きから進む。
//
// 非 graph ストアの uuid 列（MySQL / Cassandra の列型、Mongo の値、LevelDB のキー）はここでは書き換えない。
// 形式を切り替えたら、それらのストアを setup で空にしてから移行し直す（列型は移行が保存形式に合わせて作る）。
func ConvertUUIDFormat(ctx context.Context, cfg storage.Config) error {
	if cfg.Neo4j == nil {
		return fmt.Errorf("neo4j 設定がありません")
	}
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4j.URI, neo4j.BasicAuth(cfg.Neo4j.User, cfg.Neo4j.Password, ""))
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}
	defer driver.Close(ctx)

	target := "文字列"
	if id.Binary() {
		target = "17 バイト"
	}
	log.Printf("uuid の保存形式を %s へ書き換えます", target)
	for _, k := range []struct {
		name, read, write string
	}{
		{"ノード", "MATCH (n) RETURN elementId(n) AS e, n.uuid AS u",
			"UNWIND $rows AS row MATCH (n) WHERE elementId(n) = row.e SET n.uuid = row.u"},
		{"関係", "MATCH ()-[r]->() RETURN elementId(r) AS e, r.uuid AS u",
			"UNWIND $rows AS row MATCH ()-[r]->() WHERE elementId(r) = row.e SET r.uuid = row.u"},
	} {
		n, skipped, err := convertAll(ctx, driver, k.read, k.write)
		if err != nil {
			return fmt.Errorf("%s の uuid の書き換えに失敗（%d 件まで完了。再実行で続きから）: %w", k.name, n, err)
		}
		log.Printf("%s: %d 件を書き換え、%d 件は既に目的の形式", k.name, n, skipped)
	}
	return nil
}

// convertAll は read の結果（e: elementId, u: uuid）を流しながら、目的の形式でない値を write でまとめて書き換える。
func convertAll(ctx context.Context, driver neo4j.DriverWithContext, read, write string) (converted, skipped int, err error) {
	rs := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer rs.Close(ctx)
	ws := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer ws.Close(ctx)

	res, err := rs.Run(ctx, read, nil)
	if err != nil {
		return 0, 0, err
	}
	var rows []map[string]any
	flush := func() error {
		if len(rows) == 0 {
			return nil
		}
		if _, err := ws.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			r, err := tx.Run(ctx, write, map[string]any{"rows": rows})
			if err != nil {
				return nil, err
			}
			return r.Consume(ctx)
		}); err != nil {
			return err
		}
		converted += len(rows)
		rows = rows[:0]
		return nil
	}
	start := time.Now()
	for res.Next(ctx) {
		rec := res.Record()
		e, _ := rec.Get("e")
		raw, _ := rec.Get("u")
		u := id.FromAny(raw)
		if u.Empty() {
			return converted, skipped, fmt.Errorf("uuid を解釈できない要素 %v（値 %v）", e, raw)
		}
		if inTargetFormat(raw) {
			skipped++
			continue
		}
		rows = append(rows, map[string]any{"e": e, "u": id.Stored(u)})
		if len(rows) >= uuidBatch {
			if err := flush(); err != nil {
				return converted, skipped, err
			}
			if converted%(uuidBatch*50) == 0 {
				log.Printf("  … %d 件（%v）", converted, time.Since(start).Round(time.Second))
			}
		}
	}
	if err := res.Err(); err != nil {
		return converted, skipped, err
	}
	return converted, skipped, flush()
}

// inTargetFormat は Neo4j から読んだ uuid の値が既に settings.UUIDStorage の形式か。
func inTargetFormat(v any) bool {
	_, isBytes := v.([]byte)
	return isBytes == (settings.UUIDStorage == settings.UUIDBinary)
}
