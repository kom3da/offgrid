---
title: "D：PostgreSQL（Stage 1〜3）"
---

データベースを、`psql` とSQLだけで操作できるようになるトラック。Stage 1〜2のAIなしデーで、毎回1時間ずつ進める。Stage 3で、Goのトラックと合流する。

## このトラックで身につくこと

- `psql` で、SELECT・集約・JOINを、何も見ずに書ける
- トランザクションを使って更新し、間違えたら戻せる
- 主キー・外部キー・制約を含むテーブルを設計し、正規化できる
- NULLとタイムゾーンを含むクエリの結果を、実行前に予想できる
- 実行計画を読んで、遅い理由とインデックスの案を説明できる
- 大量の時系列データを、区切って持ち、集約し、古いものを消せる

## ユニット

各ユニットのページが、課題シート（答えのない問題集）になっている。

### Stage 1：SQLの基本

- **[D1 データベースとは・psql](../drills/sql/D01-psql/TASKS.md)**：データベースとテーブルを作って、削除できる
- **[D2 SELECTの基本](../drills/sql/D02-select/TASKS.md)**：検索問題を、何も見ずに10問中8問書ける
- **[D3 集約](../drills/sql/D03-aggregate/TASKS.md)**：pagilaで「月別・店舗別の売上合計」を、何も見ずに書ける
- **[D4 JOIN](../drills/sql/D04-join/TASKS.md)**：結果の行数を、実行前に予想できる
- **[D5 更新とトランザクション](../drills/sql/D05-transactions/TASKS.md)**：誤った更新を、ROLLBACKで戻せる

### Stage 2：設計と、一歩進んだSQL

- **[D6 テーブル設計](../drills/sql/D06-table-design/TASKS.md)**：G11の `csvstat` の結果を入れる表を設計できる
- **[D7 正規化](../drills/sql/D07-normalization/TASKS.md)**：1枚の表を正規化して、分割できる
- **[D8 NULLと時刻](../drills/sql/D08-null-time/TASKS.md)**：NULLが絡む結果を、正しく予想できる
- **[D9 インデックス入門](../drills/sql/D09-index/TASKS.md)**：インデックスの有無で、実行計画が変わるのを確認できる
- **[D10 CTEとビュー](../drills/sql/D10-cte-view/TASKS.md)**：複雑な集計を、段階的に書ける
- **[D11 ユーザー・権限・バックアップ](../drills/sql/D11-roles-backup/TASKS.md)**：読み取り専用ユーザーを作り、バックアップから復元できる
- **[D12 ウィンドウ関数入門](../drills/sql/D12-window/TASKS.md)**：「種別ごとの上位3件」を書ける

### Stage 3：実践

- **[D13 マイグレーション](../drills/sql/D13-migrations/TASKS.md)**：番号付きのSQLファイルで、空のDBから再構築できる
- **[D14 実行計画](../drills/sql/D14-explain/TASKS.md)**：100万行のデータで、改善前後の実行時間を記録できる
- **[D15 時系列データ](../drills/sql/D15-timeseries/TASKS.md)**：1000万行から、直近1時間の「サーバーごとの1分平均」を1秒以内で出せる

## 進め方

- F10で起動したPostgreSQLのコンテナに、`psql` で入って操作する

  ```sh
  docker exec -it pg-offgrid psql -U postgres
  ```

- 課題は `drills/sql/<ユニット>/` に、`.sql` ファイルで残す。ファイルからの実行は、次のとおり

  ```sh
  docker exec -i pg-offgrid psql -U postgres -v ON_ERROR_STOP=1 < drills/sql/D03-aggregate/work.sql
  ```

- 練習用のデータは、サンプルデータベースのpagila（DVDレンタル店の架空のデータ）を使う。Stage 1に入る前の平日に、ダウンロードしておく
- `psql` のメタコマンド（バックスラッシュで始まるもの）は `\?` で、SQLの書式は `\h コマンド名` で調べられる。AIなしでも引ける

## つまずきやすいところ

- `NULL = NULL` は真にならない。`NOT IN` とNULLの組み合わせは、直感と違う結果になる
- 文の終わりのセミコロンを忘れると、`psql` は続きの入力を待ち続ける（プロンプトの形が変わる）
- `timestamp` と `timestamptz` は別物。タイムゾーンの設定で、表示が変わる
- インデックスを作っても、使われるとは限らない。`EXPLAIN` で確かめる

## 使う教材

『SQL ゼロからはじめるデータベース操作』、『達人に学ぶDB設計徹底指南書』、『達人に学ぶSQL徹底指南書』、『内部構造から学ぶPostgreSQL 設計・運用計画の鉄則』、PostgreSQLの公式マニュアル。入手先は[教材とツール](20-resources.md)にある。
