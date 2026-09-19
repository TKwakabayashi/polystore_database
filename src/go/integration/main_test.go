//go:build integration

package integration

import (
	"os"
	"testing"
)

// TestMain はパッケージ全体で 1 度だけ cwd を src/go へ揃える（各テストは chdir しない）。
// go test は package ディレクトリ（integration/）を cwd にするため 1 段上（src/go）へ移動し、
// アプリと同じ相対パス前提（../../config/... 等）に合わせる。プロセス共有の cwd を各テストで
// 複数回 chdir すると 2 回目以降がパスを壊すため、ここで一元化する。
func TestMain(m *testing.M) {
	if err := os.Chdir(".."); err != nil {
		panic("chdir to src/go: " + err.Error())
	}
	os.Exit(m.Run())
}
