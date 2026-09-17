---
title: "G：Go（Stage 1〜3）"
---

Goを、初めて学ぶプログラミング言語として扱うトラック。Stage 1で「変数とは何か」から始め、データ構造とアルゴリズムの基礎もGoで学ぶ。Stage 2でWeb APIを作り、Stage 3で並行処理とデータベースにつなぐ。

## このトラックで身につくこと

- 条件分岐、ループ、スライス、マップ、構造体を使った処理を、何も見ずに書ける
- エラーを返し、包み、判定できる。テーブル駆動テストを書ける
- 標準ライブラリだけで、CLIツールとREST APIを書ける
- goroutineとchannelで並行処理を書き、`-race` で確かめられる
- `context` でタイムアウトとキャンセルを伝え、安全に停止できる
- PostgreSQLにつなぎ、プレースホルダとトランザクションを正しく使える
- ネットワークなしでビルドできるよう、依存を手元に固められる

## ユニット

各ユニットのページが、課題シート（答えのない問題集）になっている。

### Stage 1：Go入門

- **[G1 環境・変数・型・関数](../drills/go/G01-basics/TASKS.md)**：`go run` と `go build` の違いを説明できる
- **[G2 条件分岐とループ](../drills/go/G02-control-flow/TASKS.md)**：FizzBuzz、素数判定、九九の表を、何も見ずに書ける
- **[G3 スライスとマップ](../drills/go/G03-slices-maps/TASKS.md)**：`append` と `make` の使い方を説明できる
- **[G4 構造体・ポインタ・メソッド](../drills/go/G04-structs-pointers/TASKS.md)**：値レシーバとポインタレシーバの違いを説明できる
- **[G5 ファイルとCLI](../drills/go/G05-files-cli/TASKS.md)**：`wc` のクローンの結果が、本物の `wc` と一致する
- **[G6 エラー処理](../drills/go/G06-errors/TASKS.md)**：3つの異常系で、エラー文にファイル名と項目名が含まれる
- **[G7 テストとデバッグ](../drills/go/G07-testing-debugging/TASKS.md)**：G2〜G6にテストを付け、カバレッジ80%以上
- **[G8 インターフェースとio](../drills/go/G08-interfaces-io/TASKS.md)**：grep風フィルタが、標準入力とファイルの両方で動く
- **[G9 データ構造](../drills/go/G09-data-structures/TASKS.md)**：スタック、キュー、連結リストの計算量を説明できる
- **[G10 アルゴリズム](../drills/go/G10-algorithms/TASKS.md)**：計算量（O記法）の違いを、実測値で示せる
- **[G11 まとめ](../drills/go/G11-cli-tools/TASKS.md)**：`csvstat` と `logagg` に、オプション、テスト、READMEが揃っている

### Stage 2：GoでWeb API

- **[G12 net/http](../drills/go/G12-net-http/TASKS.md)**：CRUDのREST APIの全エンドポイントを、`curl` で確認できる
- **[G13 JSON・入力検証・テスト](../drills/go/G13-json-validation/TASKS.md)**：ハンドラのテストがすべて通る
- **[G14 ミドルウェアとログ](../drills/go/G14-middleware/TASKS.md)**：ミドルウェアの単体テストがある

### Stage 3：並行処理とDB連携

- **[G15 goroutineとchannel](../drills/go/G15-goroutines/TASKS.md)**：`go test -race` で警告が出ない
- **[G16 contextと終了処理](../drills/go/G16-context-shutdown/TASKS.md)**：Ctrl+Cで、処理中のリクエストを終えてから止まる
- **[G17 DB連携](../drills/go/G17-database/TASKS.md)**：トランザクションとエラー処理を含めて動き、危険な入力でも壊れないことをテストで確認している
- **[G18 パッケージ構成と依存管理](../drills/go/G18-packages-vendor/TASKS.md)**：`-mod=vendor` でビルドできる
- **[G19 HTTPクライアントとリトライ](../drills/go/G19-http-client-retry/TASKS.md)**：APIを止めて再開すると、自動で送信が再開する
- **[G20 Server-Sent Events](../drills/go/G20-sse/TASKS.md)**：`curl -N` で流れ続け、切断後にgoroutineが残らない

## 進め方

- 課題は標準ライブラリだけで解く。外部ライブラリを使うのは、G17の `pgx` だけ
- 課題は `drills/go/<ユニット>/` に置く。各ディレクトリで `go mod init` し、書くたびに次を実行する

  ```sh
  go vet ./... && go test -race -cover ./...
  ```

- Stage 2では、T（Webフロントエンド）より先にGを進める。G12はT4の前に、G13〜G14はT10の前に済ませる
- 二分探索木、マージソート、ジェネリクス、`database/sql` は、発展（任意）の扱い

## 覚える順番の目安

1. `fmt`、`strings`、`strconv`、`os`、`bufio`
2. `errors`、`io`、`slices`（`sort` より先にこちらを覚える）、`encoding/csv`、`encoding/json`
3. `testing`、`net/http`、`net/http/httptest`
4. `context`、`sync`、`time`、`log/slog`
5. `pgx`（発展で `database/sql`）

## つまずきやすいところ

- nilインターフェースとnilポインタの違い
- スライスの共有（`append` のあとに、元の配列を書き換えてしまう）
- ループ変数とgoroutine（Go 1.22で挙動が変わった点も含めて）
- `defer` が評価されるタイミング
- マップへの並行書き込みによるpanic
- エラーを握りつぶさない書き方（`_ =` で捨てない）

## 使う教材

『初めてのGo言語』、『プログラミング言語Go』、『なっとく！アルゴリズム』、『実用 Go言語』、『Go言語による並行処理』。A Tour of Goと標準ライブラリのドキュメントは、手元で動かす。入手先と手順は[教材とツール](20-resources.md)にある。
