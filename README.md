# polystore_database
異種データベース統合システムの外部公開用リポジトリ

アーキテクチャ

データセットアップ<br>
Neo4jの全ノード、エッジに対してUUIDを付与、インデックスを作成<br>
UUIDは現状string型で保持している

データカタログ<br>
U-Schema[1]に基づいて異種データベースを横断したデータモデルを定義する<br>
各データオブジェクトのデータ構造とデータ配置を保持している

パーサー<br>
ユーザーから与えられたCypherクエリをパースして抽象構文木を作成する

ロジカルプラン作成<br>
抽象構文木を元に演算子で構成された論理計画を作成

実行エンジン<br>
論理計画をもとにサブクエリを生成し、各データストアに発行する

データマイグレーター<br>
異種データベース間のデータ移行機能を実装している部分<br>
現状は静的な移行しか導入してない

# 参考文献
[1] Carlos J. Fernández Candel, Diego Sevilla Ruiz, and Jesús
J. García-Molina. A unified metamodel for nosql and relational
databases. Information Systems, Vol.104, p.101898,
2022.