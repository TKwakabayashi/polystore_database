# datastore

polystore の各データストア（Neo4j / MongoDB / MySQL / Cassandra）を Docker で起動する。
データセット単位で切り替えられる。

## 構成

```
datastore/
├── compose.yaml            # 全ストアの定義。データセット差分は env/ に外出し
├── .env                    # --env-file 省略時の既定（= ldbc_sf1）
├── env/<dataset>.env       # データセット定義（イメージ / DB名 / メモリ）
├── scripts/
│   ├── snapshot.sh         # Neo4j の dump を取る
│   └── migrate/            # 一回性の移行スクリプト（記録用）
├── snapshots/              # dump 実体（git 管理外）
│   ├── <dataset>/neo4j/<db>.dump
│   └── _archive/           # 旧世代 dump の退避先
└── runtime/polystore_kvs   # LevelDB 実体（git 管理外）
```

データセットごとにプロジェクト名とボリューム名が分かれる（`polystore_<dataset>_neo4j_data`）ので、
切り替えても再ロードは不要。ポートは共通なので **同時に up できるのは 1 データセットだけ**。

## 使い方

```bash
# 起動 / 停止
docker compose --env-file env/ldbc_sf1.env up -d
docker compose --env-file env/ldbc_sf1.env stop

# dump から復元（neo4j 停止中に実行）
docker compose --env-file env/openalex.env --profile load run --rm neo4j-load

# Cassandra の keyspace 作成（初回のみ。アプリ側に作成コードが無い）
docker compose --env-file env/openalex.env --profile init run --rm cassandra-init

# dump を取る
scripts/snapshot.sh ldbc_sf1
```

## データセット一覧

| dataset | Neo4j | DB 名 | 規模 |
|---|---|---|---|
| ldbc_sf1 | neo4j:5 | neo4j | 3.0M nodes / 17.2M rels |
| openalex | neo4j:5 | openalex | Work 21.9M / Author 9.6M |
| amazon_reviewdata2023 | neo4j:5 | amazon | User 23.2M / Item 3.7M |
| movielens | neo4j:5 | movielens | User 162K / Movie 62K |

全て Neo4j 5 系 Community（GPLv3）で動作確認済み。

## データセットを追加する

1. dump を `snapshots/<dataset>/neo4j/<db名>.dump` に置く。
   **ファイル名＝データベース名**。`neo4j-admin database load` がこの規約で dump を探す。
2. `env/<dataset>.env` を作る。既存ファイルをコピーして `DATASET` / `NEO4J_DATABASE` /
   `NEO4J_IMAGE` / メモリを埋める。
3. `--profile load` で復元 → `up -d` で起動。

## 注意点

- **Neo4j のストアはバージョン間で互換が無く、ダウングレードもできない。**
  dump を作った Neo4j のメジャーバージョンと `NEO4J_IMAGE` は必ず一致させる。
  異なるバージョンへ移す場合は CSV 経由の論理コピーが必要
  （`scripts/migrate/ldbc_to_neo4j5.sh` が実例）。
- **block 形式の dump は Enterprise 専用。** Community では読めない。
  `snapshots/_archive/ldbc_sf1_neo4j2025.dump` がこれに該当し、読むには
  `neo4j:2025-enterprise` が要る。
- **heap + pagecache が Docker VM の物理メモリを超えると Neo4j は起動を拒否する。**
  現状 VM は 7.7G なので各 env は heap 4G / pagecache 1G に抑えてある。
  Docker Desktop の割当を増やしたら env の値も上げてよい。
- LevelDB のパスは `config/config.json` の `leveldb.path` と対応している
  （`../datastore/runtime/polystore_kvs`）。
