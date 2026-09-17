# Goトラック

Goは初めて学ぶプログラミング言語として扱う。Stage 1で「変数とは何か」から始め、データ構造とアルゴリズムの基礎もGoで学ぶ。G1〜G16の課題は標準ライブラリだけで解き（G17で初めて外部ライブラリの`pgx`を使う）、`drills/go/<ユニット>-<名前>/` に置く。

## Stage 1：Go入門（G1〜G11）

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| G1 | 環境・変数・型・関数 | Goをインストールし、Hello World、四則演算、関数を書く。A Tour of Goの「Basics」 | `go run`と`go build`の違いを説明できる |
| G2 | 条件分岐とループ | FizzBuzz、素数判定、九九の表 | 3つとも何も見ずに書ける |
| G3 | スライスとマップ | 文章中の単語を数えて多い順に並べる | `append`と`make`の使い方を説明できる |
| G4 | 構造体とメソッド | 銀行口座（入金・出金・残高）を構造体で表す | 値レシーバとポインタレシーバの違いを説明できる |
| G5 | ファイルとCLI | `os`、`bufio`、`flag`で`wc`コマンドのクローンを作る | 行数・単語数・バイト数が本物の`wc`と一致する |
| G6 | エラー処理 | `error`の返し方、`errors.Is`/`As`、`%w`でのラップ。設定ファイルローダーを作る | 存在しないファイルや不正な値で、分かりやすいエラーが出る |
| G7 | テスト | テーブル駆動テスト、`testdata/`、`go test -cover` | G2〜G6にテストを付け、カバレッジ80%以上 |
| G8 | インターフェースとio | `io.Reader`/`io.Writer`を受けるgrep風フィルタ | 標準入力とファイルの両方で動く |
| G9 | データ構造 | スタック、キュー、連結リスト、二分探索木を自作する | それぞれの追加・取り出しの計算量を説明できる |
| G10 | アルゴリズム | 線形探索と二分探索、バブルソートとマージソート、ベンチマークで速度比較 | 計算量（O記法）の違いを実測値で示せる |
| G11 | まとめ | CSVを集計する`csvstat`と、アクセスログを集計する`logagg` | オプション、テスト、READMEが揃っている |

## Stage 2：GoでWeb API（G12〜G14）

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| G12 | net/http | `ServeMux`でCRUDのREST APIを作る。Go 1.22以降のパターン構文（`GET /items/{id}`の形）を使う。データはメモリ上 | `curl`で全エンドポイントを確認できる |
| G13 | JSON・入力検証・テスト | リクエストの検証、エラーレスポンスの統一、`httptest`でのテスト | ハンドラのテストがすべて通る |
| G14 | ミドルウェアとログ | リクエストID、アクセスログ、`log/slog`での構造化ログ、簡易トークン認証 | ミドルウェアの単体テストがある |

## Stage 3：並行処理とDB連携（G15〜G18）

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| G15 | goroutineとchannel | 複数URLの死活監視を並行実行するワーカープール、`sync.WaitGroup`、`Mutex` | `go test -race`で警告が出ない |
| G16 | contextと終了処理 | タイムアウト、キャンセル、`signal.NotifyContext`での安全な停止 | Ctrl+Cで処理中のリクエストを終えてから止まる |
| G17 | DB連携 | G12のAPIの保存先をPostgreSQLに置き換える。`database/sql`と`pgx`の両方を試す | トランザクションとエラー処理を含めて動く |
| G18 | パッケージ構成と依存管理 | `internal/`を使った構成、`go mod tidy`、`go mod vendor` | `-mod=vendor`でビルドできる |

## 覚える順番の目安

1. `fmt`、`strings`、`strconv`、`os`、`bufio`
2. `errors`、`io`、`slices`（Go 1.21以降。`sort`より先にこちらを覚える）、`encoding/csv`、`encoding/json`
3. `testing`、`net/http`、`net/http/httptest`
4. `context`、`sync`、`time`、`log/slog`
5. `database/sql`、`pgx`

## Go特有のつまずきポイント

- nilインターフェースとnilポインタの違い
- スライスの共有（`append`後に元の配列を書き換えてしまう）
- ループ変数とgoroutine（Go 1.22で挙動が変わった点も含めて）
- `defer`の評価タイミング
- マップの並行書き込みによるpanic
- エラーを握りつぶさない書き方（`_ =`を使わない）

## 課題の作り方

```sh
mkdir -p drills/go/G02-fizzbuzz && cd drills/go/G02-fizzbuzz
go mod init offgrid/g02-fizzbuzz
go vet ./... && go test -race -cover ./...
```
