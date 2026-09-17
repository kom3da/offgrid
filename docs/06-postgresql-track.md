# PostgreSQLトラック

Stage 1〜2のAIなしデーで、毎回1時間ずつ進める。F10で起動したPostgreSQLのコンテナに`psql`で入り、SQLだけで操作する。練習用データは公式チュートリアルの例か、サンプルDBのpagila（DVDレンタル店の架空データ）を事前にダウンロードしておく。課題は `drills/sql/<ユニット>-<名前>/` に `.sql` ファイルで残す。

## Stage 1〜2：SQLの基礎（D1〜D12）

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| D1 | データベースとは・psql | 表・行・列の考え方、`\l`、`\dt`、`\d テーブル名`、`\x`、`\?` | DBとテーブルを作って削除できる |
| D2 | SELECTの基本 | `WHERE`、`ORDER BY`、`LIMIT`、`LIKE`で検索問題を10問 | 何も見ずに10問中8問 |
| D3 | 集約 | `COUNT`、`SUM`、`AVG`、`GROUP BY`、`HAVING` | pagilaで「月別・店舗別の売上合計」を何も見ずに書ける |
| D4 | JOIN | `INNER JOIN`と`LEFT JOIN`の違い、サブクエリ | 結果の行数を実行前に予想できる |
| D5 | 更新とトランザクション | `INSERT`、`UPDATE`、`DELETE`、`BEGIN`/`COMMIT`/`ROLLBACK` | 誤更新をROLLBACKで戻せる |
| D6 | テーブル設計 | 型（`text`、`integer`、`numeric`、`timestamptz`、`jsonb`）、主キー、外部キー、`CHECK`制約 | G11の`csvstat`の結果を入れる表を設計できる |
| D7 | 正規化 | 第1〜第3正規形、重複データの問題 | 1枚の表を正規化して分割できる |
| D8 | NULLと時刻 | NULLの比較、`COALESCE`、`timestamptz`とタイムゾーン設定 | NULL絡みの結果を正しく予想できる |
| D9 | インデックス入門 | `CREATE INDEX`、`EXPLAIN`で Seq Scan と Index Scan を見分ける | インデックスの有無で実行計画が変わるのを確認できる |
| D10 | CTEとビュー | `WITH`句、`CREATE VIEW` | 複雑な集計を段階的に書ける |
| D11 | ユーザー・権限・バックアップ | ロール、`GRANT`、`pg_dump`/`pg_restore` | 読み取り専用ユーザーを作り、バックアップから復元できる |
| D12 | ウィンドウ関数入門 | `ROW_NUMBER`、`RANK`、累計 | 「種別ごとの上位3件」を書ける |

## Stage 3：実践（D13〜D15）

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| D13 | マイグレーション | スキーマ変更を番号付きのSQLファイルで管理し、G17のAPIに適用する | 空のDBから再構築できる |
| D14 | 実行計画 | 100万行のダミーデータで`EXPLAIN (ANALYZE, BUFFERS)`を読み、インデックスを設計する | 改善前後の実行時間を記録できる |
| D15 | 時系列データ | 時刻で区切るパーティション、BRINインデックス、`date_trunc`での集約、古いデータの削除 | 1000万行の計測データから、直近1時間の「サーバーごとの1分平均」を1秒以内（`EXPLAIN ANALYZE`の実行時間）で出せる（Stage 5の準備） |

`psql`のメタコマンド（バックスラッシュで始まるもの）は`\?`で一覧が出るので、AIなしでも調べられる。

## 課題の実行

```sh
docker exec -i pg-offgrid psql -U postgres -v ON_ERROR_STOP=1 < drills/sql/D03-aggregate/answer.sql
```
