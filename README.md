# polystore_database

異種データベース（Neo4j / MongoDB / MySQL / Cassandra / LevelDB）を横断する
ポリストア統合システム。Cypher クエリを受け取り、U-Schema[1] に基づくデータカタログを
参照して、データ配置に応じたサブクエリを各ストアへ発行する。データ配置（placement）と
実行モデル（execution model）を切り替えて性能を比較できる研究用システム。

---

## アーキテクチャ

```
Cypher クエリ
   │
   ▼
┌──────────┐   パーサー：Cypher → 抽象構文木（AST）
│ parser   │
└────┬─────┘
     ▼
┌──────────┐   プランナー：AST → 論理計画（演算子ツリー）。
│ planner  │   schema（データカタログ）を見て scan/filter/集約 pushdown の配置を決める
└────┬─────┘
     ▼
┌──────────┐   実行エンジン：論理計画を実行し、各ストアへサブクエリを発行。
│ engine   │   実行モデルは stream / bulk / volcano / vectorized から選択（結果は同一）
│  ├ core  │     エンジン共通型（Record / Result / 条件評価 / 集約 pushdown ビルダ）
│  └ access│     ストアごとの scan/filter/fetch。ここがドライバ境界
└────┬─────┘
     ▼
  各データストア（storage 経由で接続）
```

### 主要コンポーネント

| 層 | 役割 |
|---|---|
| **schema（データカタログ）** | U-Schema[1] に基づき異種DBを横断したデータモデルを定義。各データオブジェクトの構造とデータ配置（どのプロパティがどのストアにあるか）を保持 |
| **parser** | Cypher クエリをパースして AST を生成 |
| **planner / plan** | AST から演算子（EntityScan / Filter / Expand / Projection / StorePushdown …）で構成した論理計画を作成。集約は単一ストアへ委譲（pushdown）して往復を削減 |
| **engine** | 論理計画を実行する 4 つの実行モデル。統一 `Engine`/`Instance` インターフェース + 名前引きレジストリ。中間結果は `core.Record{Slots []id.UUID}`（列指向 volcano は `Batch [][]id.UUID`） |
| **migrator** | 異種DB間のデータ移行（graph ⇄ rdb/doc/col/kvs）。移行後に mapping を更新して配置を切り替える |
| **id** | UUID の唯一の権威。採番・境界変換・順序・プロパティ名を集約。**識別子の内部表現はこのパッケージに閉じ、他層は `id.UUID` 型で扱う**（表現変更が id/ だけで済む不変条件） |
| **storage** | 各ストアの接続レジストリ（Neo4j/Mongo/MySQL/Cassandra/LevelDB） |
| **codec** | 値エンコード・KVS 複合キー生成 |
| **store** | ストア種別 enum（`store.Kind`） |

### 実行モデル（execution model）

同じ論理計画を、イテレーションモデルだけを変えて実行できる。結果（行集合）は全モデルで一致する。

| モデル | 方式 |
|---|---|
| `stream` | channel + goroutine のストリーミング（並列度を制御） |
| `bulk` | 全件マテリアライズの逐次実行（計測がクリーン） |
| `volcano` | pull 型・列指向 Batch（tuple-at-a-time） |
| `vectorized` | volcano をベクトル幅 N で（`settings.VectorSize`） |

---

## ディレクトリ構成

```
polystore_database/
├── src/go/               本体（Go）
│   ├── cmd/polystore/    単一エントリポイント（CLI）
│   ├── parser/ planner/ plan/      クエリ→AST→論理計画
│   ├── engine/           実行エンジン（core / stream / bulk / volcano / all）
│   ├── schema/ migrator/ storage/ codec/ id/ store/
│   ├── workload/         ワークロード定義（Cypher + params + 移行設定）
│   ├── bench/            計測ハーネス（RunEngine / RunModelBenchmark …）
│   ├── settings/         内部挙動の切替（実行モデル・pushdown・スイープ軸 …）
│   └── integration/      結合テスト（//go:build integration、実DB必要）
├── datastore/            各DBの Docker 定義（compose.yaml + データセット別 env/）
├── config/               config.json（config_expample.json をコピーして作成）
├── docs/                 設計ドキュメント（EXTENDING.md 等・ローカル管理）
└── results/              ベンチ結果（ローカル管理）
```

---

## セットアップと実行

前提: Docker、Go。コマンドは断りがなければ **`src/go/` を作業ディレクトリ**として実行する。

### 1. データストアを起動（`datastore/` で実行）

```bash
cd datastore
docker compose --env-file env/ldbc_sf1.env up -d          # 全ストア起動
docker compose --env-file env/ldbc_sf1.env --profile init run --rm cassandra-init   # 初回のみ: Cassandra keyspace 作成
# dump からの Neo4j 復元（初回・neo4j 停止中）:
#   docker compose --env-file env/ldbc_sf1.env --profile load run --rm neo4j-load
```

データセットは `env/<dataset>.env` で切替（ldbc_sf1 / openalex / amazon / movielens）。
ポートは共通のため**同時に起動できるのは 1 データセットだけ**。停止は `... stop`。

### 2. 設定ファイル

`config/config_expample.json` を `config/config.json` にコピーし、各DBの接続情報を記入する。

### 3. データセットアップ（初回）

```bash
cd src/go
go run ./cmd/polystore -mode setup      # Neo4j にUUID付与・インデックス作成 + 他4ストア初期化
```

### 4. クエリ実行

```bash
# 単発実行（settings のエンジン・出力形式に従う）
go run ./cmd/polystore -mode run -workload Q11
```

### 5. データ移行

```bash
# graph → rdb へ指定ワークロードのプロパティを移行し、時間・件数を出力
go run ./cmd/polystore -mode migrate  -workload AGG5 -migmode graph_to_rdb
# 移行 → クエリ実行（Neo4j 比較）まで通し
go run ./cmd/polystore -mode workflow -workload AGG5 -migmode graph_to_rdb
```

`-migmode` は `a_to_b`（例: `graph_to_rdb` / `graph_to_doc` / `graph_to_col` / `graph_to_kvs` と逆方向）。

### 6. ベンチマーク

```bash
# baseline(Neo4j直) と placement×pushdown を計測し CSV へ追記
go run ./cmd/polystore -mode bench        -workload all -out ../../results/bench/bench.csv
# placement × 実行モデル（stream/bulk/volcano/vectorized）を計測（long 形式 CSV）
go run ./cmd/polystore -mode bench-models -workload Q11,AGG5 -out ../../results/bench/models.csv
```

利用可能なワークロード: `Q2 Q8 Q9 Q11 IS1..IS6 AGG1..AGG6 BI3 BI9`（`-workload all` で全件、カンマ区切りで複数）。
Q2/Q8/Q9/Q11 は公式 LDBC Interactive Complex IC2/8/9/11、IS1/3/4/5/6 は Interactive Short、
BI3/BI9 は Business Intelligence の read クエリに対応する（対応表と拡張ロードマップは `docs/LDBC_WORKLOADS.md`）。

### 内部挙動の切替

実行モデル・pushdown 方針・配置/モデルのスイープ軸・ベクトル長などは CLI ではなく
**`src/go/settings/` を編集して再ビルド**で切り替える（CLI は「何を実行するか」だけを受ける）。

---

## テスト

```bash
cd src/go
go test ./...                                                         # ユニット（DB不要）
# golden プランの再生成（ワークロードやプランナを変えたとき。差分を確認してコミット）:
go test ./planner/ -run Golden -update
```

### 結合テスト（実 DB・合成 mini-SNB）

実 LDBC dump は重いため、結合テストは小さな合成 mini-SNB（`integration/seed.go`）を**隔離スタック**
（`datastore/env/citest.env`・実データと別ポート/別ボリューム）へ投入して回す。`POLYSTORE_SEED=1` で
`TestMain` が seed する（**この時 Neo4j は全削除されるため、実データスタックに対しては付けない**）。

```bash
# 1) 隔離スタックを起動（実スタックと同時起動可）
cd datastore
docker compose --env-file env/citest.env up -d
docker compose --env-file env/citest.env --profile init run --rm cassandra-init
# 2) seed して結合テスト（全ワークロード非空 / 4モデル等価 / 移行ラウンドトリップ）
cd ../src/go
POLYSTORE_SEED=1 POLYSTORE_CONFIG=../../config/config.citest.json \
  go test -tags integration ./integration/ -v
# 3) 破棄
cd ../datastore && docker compose --env-file env/citest.env down -v
```

ロード済みの実データに対して流す場合は `POLYSTORE_SEED` を付けず、`POLYSTORE_CONFIG` を実データ用の
`config/config.json` にする（`WorkloadNonempty` は実在しない id で 0 件になる params ドリフトを検知する）。

ユニット（DB 不要）の主なカバレッジ:
- **planner golden プラン**（`planner/plan_golden_test.go` + `planner/testdata/*.plan`）:
  全ワークロードの Cypher を `planner.ParseQuery` で論理プランへ変換し、木のダンプが golden と一致するか検証。
  パーサ/プランナ/pushdown 判定の回帰を DB 無しでガードする。
- **純関数**: `engine/core`（pushdown クエリビルダ・演算子変換）、`codec`（型変換・KVS キー）、
  `schema`（mapping 引き）、`engine`（レジストリ解決）、`workload`（params 完備性）。

CI は目的別に 3 分割（`.github/workflows/`）:
- `lint.yml`: gofmt / go vet / staticcheck（DB 不要・除外方針は `src/go/staticcheck.conf`）
- `unit.yml`: build + `go test ./...`（**ubuntu + windows** マトリクスで Mac/Windows 両対応を検証）
- `integration.yml`: 合成 mini-SNB を隔離スタックへ seed して結合テスト（手動 dispatch と main への push）

---

## 拡張

新しい実行エンジン / ストア / ワークロードの追加手順、UUID 表現の不変条件、pushdown の拡張点は
`docs/EXTENDING.md` を参照。

---

## 参考文献

[1] Carlos J. Fernández Candel, Diego Sevilla Ruiz, and Jesús J. García-Molina.
A unified metamodel for nosql and relational databases. *Information Systems*,
Vol. 104, p. 101898, 2022.
