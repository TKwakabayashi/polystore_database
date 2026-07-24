//go:build integration

// main_test は integration パッケージ共通の前処理（cwd 調整・設定読み込み・任意の seed）を集約する。
//
// seed は POLYSTORE_SEED=1 のときだけ実行する（安全側の既定）。
// 実データ（ロード済み LDBC dump 等）に対して流すと DETACH DELETE で全消去してしまうため、
// 合成 mini-SNB を使う隔離スタックでのみ明示的に opt-in する。
package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"polystore_database/src/go/storage"
)

func TestMain(m *testing.M) {
	// go test は package ディレクトリ(integration/)を cwd にする。アプリと同じ相対パス前提
	// （cwd=src/go）へ 1 段上げる。ここで 1 度だけ行い、各テストは chdir しない。
	if wd, err := os.Getwd(); err == nil && filepath.Base(wd) == "integration" {
		if err := os.Chdir(".."); err != nil {
			log.Fatalf("chdir to src/go: %v", err)
		}
	}

	if os.Getenv("POLYSTORE_SEED") == "1" {
		cfg, err := storage.LoadConfig(cfgPath())
		if err != nil {
			log.Fatalf("seed 用 config 読み込み失敗 (%s): %v", cfgPath(), err)
		}
		if err := SeedMiniSNB(context.Background(), cfg); err != nil {
			log.Fatalf("mini-SNB seed 失敗: %v", err)
		}
	}

	os.Exit(m.Run())
}

// cfgPath は設定ファイルパス（POLYSTORE_CONFIG、未設定なら既定）。cwd=src/go 前提。
func cfgPath() string {
	if p := os.Getenv("POLYSTORE_CONFIG"); p != "" {
		return p
	}
	return "../../config/config.json"
}

// loadCfg は共通の設定読み込み。cwd は TestMain で src/go に揃っている前提。
func loadCfg(t *testing.T) storage.Config {
	t.Helper()
	cfg, err := storage.LoadConfig(cfgPath())
	if err != nil {
		t.Fatalf("config load (%s): %v", cfgPath(), err)
	}
	return cfg
}
