package engine_test

// engine/all を blank import して全エンジンの init() 登録を集約した状態で、
// レジストリと store.Kind の語彙が壊れていないことを DB 無しで検証する。
// （外部テストパッケージにするのは engine ← engine/all の import 循環を避けるため。）

import (
	"sort"
	"testing"

	"polystore_database/src/go/engine"
	_ "polystore_database/src/go/engine/all"
	"polystore_database/src/go/settings"
	"polystore_database/src/go/store"
)

// 4 実行モデルすべてが engine.New で解決できる（engine/all の import 漏れ検知）。
func TestAllEngineKindsResolve(t *testing.T) {
	kinds := []settings.EngineKind{
		settings.EngineStream,
		settings.EngineBulk,
		settings.EngineVolcano,
		settings.EngineVectorized,
	}
	for _, k := range kinds {
		e, err := engine.New(k)
		if err != nil {
			t.Errorf("engine.New(%q) 失敗: %v", k, err)
			continue
		}
		if e.Name() == "" {
			t.Errorf("engine.New(%q).Name() が空", k)
		}
	}
	if got := len(engine.Registered()); got < len(kinds) {
		t.Errorf("登録数 = %d, want >= %d", got, len(kinds))
	}
}

func TestUnknownEngineKindErrors(t *testing.T) {
	if _, err := engine.New(settings.EngineKind("no-such-engine")); err == nil {
		t.Errorf("未登録エンジンでエラーになるはず")
	}
}

// store.Kind の 5 種すべてが String → ParseKind で往復する。
func TestStoreKindRoundTrip(t *testing.T) {
	kinds := []store.Kind{store.Graph, store.Columnar, store.Relational, store.Document, store.Kvs}
	seen := map[string]bool{}
	for _, k := range kinds {
		name := k.String()
		if seen[name] {
			t.Errorf("重複した文字列表現: %q", name)
		}
		seen[name] = true
		got, err := store.ParseKind(name)
		if err != nil {
			t.Errorf("ParseKind(%q) 失敗: %v", name, err)
			continue
		}
		if got != k {
			t.Errorf("ParseKind(%q) = %v, want %v", name, got, k)
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names)
	if len(names) != 5 {
		t.Errorf("distinct store 名 = %v, want 5 種", names)
	}
}

func TestParseKindUnknownErrors(t *testing.T) {
	if _, err := store.ParseKind("no-such-store"); err == nil {
		t.Errorf("未知ストアでエラーになるはず")
	}
}
