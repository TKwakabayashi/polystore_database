// Package engine は実行エンジンの統一インターフェースと名前引きレジストリを提供する。
//
// 各エンジン（engine/stream, engine/bulk, engine/volcano）は自身の init() で Register し、
// engine/all を blank import することで登録が集約される。新エンジン追加は
// 「settings.EngineKind 追加 + engine/<name>/ 作成 + Register + all.go に import」の4箇所。
package engine

import (
	"context"
	"fmt"

	"polystore_database/src/go/engine/core"
	"polystore_database/src/go/plan"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/storage"
)

// Engine は実行エンジンのファクトリ。Open で接続を確立した Instance を返す。
type Engine interface {
	Name() string
	Open(ctx context.Context, cfg storage.Config) (Instance, error)
}

// Instance は1接続分の実行コンテキスト。Warmup+Trials の間再利用される。
type Instance interface {
	// Run はパース済みプランを1回実行する（冒頭で内部状態を Reset）。
	// 返す core.Result の ExecTime に自身の実行時間を格納する（PlanTime は呼び出し側が付与）。
	Run(op plan.PlanNode) (core.Result, error)
	Close() error
}

var registry = map[settings.EngineKind]func() Engine{}

// Register は各エンジンの init() から呼ばれ、EngineKind とファクトリを登録する。
func Register(k settings.EngineKind, f func() Engine) { registry[k] = f }

// New は登録済みエンジンを生成する。未登録なら engine/all の import 漏れを疑う。
func New(k settings.EngineKind) (Engine, error) {
	f, ok := registry[k]
	if !ok {
		return nil, fmt.Errorf("未登録のエンジン %q（engine/all を import しているか確認）", k)
	}
	return f(), nil
}

// Registered は登録済み EngineKind の一覧（順序は不定）。
func Registered() []settings.EngineKind {
	out := make([]settings.EngineKind, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}
