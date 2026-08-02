// Package profile はスコープ指定できる pprof 採取ユーティリティ。
//
// settings.ProfileScopes で有効化されたスコープだけ採取し、無効なら no-op。
// 出力は {settings.ResultsDir}/profile/{label}_{scope}_{kind}.prof。
// 旧 bench.RunCustom 内にハードコードされていた pprof ブロック（cpu/heap/allocs/
// block/mutex、block/mutex rate=1）をそのまま一般化したもの。
package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"polystore_database/src/go/settings"
)

// Session は採取中の1回分。Stop で CPU 採取を止め、heap/allocs/block/mutex を dump する。
type Session struct {
	active bool
	label  string
	scope  settings.ProfileScope
	cpu    *os.File
}

// Start は settings.ProfileScopes[scope] が有効なときだけ採取を開始する。
// 無効なら active=false の Session を返す（Stop も no-op）。
// CPU プロファイルはプロセスで同時に1つしか採れないため、多重 Start は上位で避ける。
func Start(scope settings.ProfileScope, label string) *Session {
	if !settings.ProfileScopes[scope] {
		return &Session{active: false}
	}
	if err := os.MkdirAll(filepath.Join(settings.ResultsDir, "profile"), 0o755); err != nil {
		fmt.Printf("⚠️ profile ディレクトリ作成に失敗: %v\n", err)
		return &Session{active: false}
	}
	s := &Session{active: true, label: label, scope: scope}
	s.cpu, _ = os.Create(s.path("cpu"))
	pprof.StartCPUProfile(s.cpu)
	runtime.SetBlockProfileRate(1) // 1ns = 全ブロックイベント記録
	runtime.SetMutexProfileFraction(1)
	return s
}

func (s *Session) path(kind string) string {
	name := fmt.Sprintf("%s_%s_%s.prof", s.label, s.scope, kind)
	return filepath.Join(settings.ResultsDir, "profile", name)
}

// Stop は CPU 採取を止め、heap/allocs/block/mutex を dump する。no-op Session でも安全。
func (s *Session) Stop() {
	if s == nil || !s.active {
		return
	}
	pprof.StopCPUProfile()
	s.cpu.Close()

	dump := func(kind string) {
		f, _ := os.Create(s.path(kind))
		defer f.Close()
		pprof.Lookup(kind).WriteTo(f, 0)
	}
	runtime.GC() // heap のライブ集合を正確に
	dump("heap")
	dump("allocs")
	dump("block")
	dump("mutex")
}
