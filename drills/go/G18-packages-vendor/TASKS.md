---
title: "G18 パッケージ構成と依存管理"
---

> トラック：[G：Go](../../../docs/11-track-g-go.md) ／ Stage 3 ／ 目安：メイン枠で2回 ／ 前提：G11、G17

## ねらい

これまで1つのディレクトリに詰め込んできたコードを、複数のパッケージに分け、外から使わせたくないものを `internal/` に隠す。そして、外部の依存（G17で初めて入った `pgx`）を `vendor/` に固め、**ネットワークのない場所でもビルドできる**ことを、自分の手で確かめる。閉域の現場では、`go get` は動かない。持ち込んだソースだけでビルドが完結する形にしておくことが、Goの配布の出発点になる（O7、C6でそのまま使う）。

## このユニットで身につけること

- モジュールの中に複数のパッケージを作り、モジュールパス＋ディレクトリ名でimportする
- `cmd/` に `main` を、`internal/` に共通の処理を置く構成に、G11とG17のコードを移す
- `internal/` の中のパッケージが、外のモジュールからimportできないことを確かめる
- `go mod tidy` で、`go.mod` と `go.sum` を実際の依存に合わせる
- `go mod vendor` で依存のソースを `vendor/` に写し、その中身を読む
- ネットワークとモジュールキャッシュを使えない状態で、`-mod=vendor` でビルドとテストが通ることを確かめる
- 「閉域でビルドするまでの手順」を、スクリプトと文章にする

## キーワード

- **モジュールパス**：`go.mod` の1行目に書かれた名前（例：`offgrid/g18-packages-vendor`）。このモジュールの中のパッケージは、「モジュールパス/ディレクトリ」の形でimportする
- **`cmd/`**：実行ファイルになる `main` パッケージを、1つずつディレクトリに分けて置く場所。慣習であり、`go` コマンドの決まりではない
- **`internal/`**：この名前のディレクトリの中のパッケージは、その親ディレクトリより下からしかimportできない。これは `go` コマンドの決まり
- **エクスポート**：名前の先頭を大文字にして、パッケージの外から使えるようにすること
- **`go.sum`**：依存の各バージョンのハッシュ（中身から計算した短い値）の一覧。中身がすり替わっていないかの確認に使う
- **モジュールキャッシュ**：`go get` で取ってきた依存が置かれる場所。`go env GOMODCACHE` で分かる
- **ベンダリング**：依存のソースを、自分のリポジトリの `vendor/` に丸ごと写して、一緒に管理すること
- **`GOPROXY`／`GOFLAGS`**：`go` コマンドの動きを変える環境変数。`go help environment` に一覧がある

## 平日に読むもの

- 『初めてのGo言語』10章（モジュールとパッケージ）、11章（各種ツール）
- 『実用 Go言語』7章（パッケージ、モジュール）、8章（Goプログラミングの環境を整備する）
- `go help modules`、`go help mod tidy`、`go help mod vendor`、`go help build`（`-mod` の説明）、`go help environment`
- Go Modules Reference（`go help modules` に案内がある。オフライン用に保存しておく）の、`go.mod` ファイル、`internal` ディレクトリ、ベンダリングを扱う節
- `docs/15-track-c-capstone.md` の「ディレクトリ案」（今回の構成が、卒業制作の形になる）

## 準備

**`pgx` はG17でモジュールキャッシュに入っている前提**。もし `go mod tidy` がネットワークを要求したら、平日に済ませる。

```sh
go mod init offgrid/g18-packages-vendor
mkdir -p cmd/csvstat cmd/logagg cmd/api internal
cp -r ../G11-cli-tools/testdata .
cp ../G17-database/schema.sql .
```

G11の `csvstat/`、`logagg/` と、G17のAPI（`experiments/` は除く）を、課題の中でこの構成に移していく。元のディレクトリは変えない。

## 課題

1. 小さく試す：`internal/greet/` に、挨拶の文字列を返す関数を1つ持つパッケージを作り、`cmd/api/main.go` からimportして表示する。importの文字列は、何を基準に決まるか。関数名の先頭を小文字にすると、コンパイラは何と言うか。`go build ./...` と `go build ./cmd/...` の違いを、`go help packages` で調べる
2. 観察：`~/sandbox/g18` に別のモジュールを作り、そこから `offgrid/g18-packages-vendor/internal/greet` をimportしてみる（別モジュールから手元のモジュールを参照する方法は `go help mod edit` の `-replace` で調べる）。エラーメッセージを `notes.md` に記録する。`internal` を `pkg` に改名すると、どうなるか。確かめたら `greet` は消してよい
3. G11の `csvstat` を移す：集計の処理を `internal/csvstat/` に、`main` を `cmd/csvstat/` に置く。テストも一緒に移す。`main.go` に残るのは、`flag` の処理と、標準入出力とのつなぎだけにする。`go test ./...` が通り、`go build -o bin/csvstat ./cmd/csvstat` でできたものが、G11のREADMEどおりに動くこと
4. `logagg` も同じように移す。2つのツールで、同じことをしている処理（ファイルか標準入力かを選んで開く、上位N件を並べる、など）はないか。あれば、`internal/` の中の共通パッケージに1つにまとめる。まとめないと決めた場合は、理由を書く
5. G17のAPIを移す：`cmd/api/main.go` は、設定の読み込みと起動・終了処理（G16）だけにする。ストアのインターフェースと、メモリ版・PostgreSQL版の実装は `internal/store/`、ハンドラとミドルウェアは `internal/` の下の別のパッケージに置く。パッケージをまたぐと、これまで同じパッケージで見えていた名前が見えなくなる。何をエクスポートし、何を隠すかを、1つずつ決める。G17の結合テストが（`TEST_DATABASE_URL` ありで）通ること
6. `go mod tidy` を試す。まず `go.mod` を `cat` で読む。`go.mod` の `require` の行を1つ手で消して `go build ./...` すると、何と言われるか。`go mod tidy` の後、戻っているか。逆に、使っていないモジュールの行を手で足して `tidy` すると、どうなるか。`go.sum` の各行は何を表しているか。`// indirect` と書かれた行は何か（Go Modules Referenceの `go.mod` の節で探す）
7. `go list -m all` で、このモジュールが使っているモジュールを全部表示する。`pgx` を1つ入れただけなのに、何個あるか。そのうちの1つについて、`go mod why` で「誰が必要としているか」を調べる
8. `go mod vendor` を実行し、`vendor/` の中を `ls -R | head`、`du -sh`、`cat vendor/modules.txt | head` で見る。`modules.txt` には何が書かれているか。`vendor/` の中には、依存モジュールの全ファイルが入っているか、それとも使っているパッケージだけか。`vendor/` をコミットするかどうかを決め、理由を `notes.md` に書く（閉域に持ち込む方法として、コミットする以外に何があるかも考える）
9. 観察：`vendor/` がある状態で、何も付けずに `go build ./...` すると、`vendor/` は使われるか。`go help build` の `-mod` の説明を読んで、答えと、その条件（`go.mod` の `go` の行）を書く。`go.mod` の `pgx` のバージョンを手で1つ変えてから `go build ./...` すると、何と言われるか（確かめたら戻す）
10. **本題**：ネットワークとキャッシュがない状態を作って、ビルドする。空のディレクトリを作り、環境変数 `GOMODCACHE` でそこをキャッシュに指定し、`GOPROXY=off` で取得を禁止する。この状態で、(a) `-mod=mod`、(b) `-mod=vendor` のそれぞれで `go build ./...` と `go test ./...` を実行し、結果を比べる。(a) のエラーメッセージのどこに、「ネットワークを使おうとした」ことが表れているか。この手順を `build-offline.sh`（F5）にまとめ、`bin/` に3つの実行ファイルができ、テストが通ったら0で終わるようにする。`GOFLAGS=-mod=vendor` を使う書き方も試し、`-mod=vendor` を毎回付けるのとどちらがよいかを決める
11. 課題10でできた `bin/api` を `go version -m` で見る。依存モジュールとバージョンは記録されているか。`-mod=mod` でビルドしたものと、出力に違いはあるか。閉域に持ち込んだ実行ファイルが「どの依存で作られたか」を、後から知る方法として使えるか
12. `explain.md` に、何も見ずに書く：ベンダリングとは何か、閉域の現場でなぜ要るか、平日（ネットワークあり）にやっておくこと、閉域で実行するコマンド、`go.mod`／`go.sum`／`vendor/modules.txt` のそれぞれの役割

### 発展（任意）

- `vendor/` を使わない別の方法：平日に `go mod download` でキャッシュを埋め、`GOMODCACHE` のディレクトリごと `tar` で持ち込む。課題10と同じ条件でビルドできるか。`vendor/` と比べて、何が楽で、何が面倒か（O7で、どちらかを選ぶ）
- `go mod verify` は、何を確かめているか。`vendor/` の中のファイルを1文字書き換えると、ビルドとテストは気づくか
- 依存を1つ増やす場合（例：`golang.org/x/sync`）の手順を、`build-offline.sh` が通る状態を保ったまま、`notes.md` に書き出す
- `cmd/` の3つの `main.go` に共通する処理（ログの設定など）を、`internal/` に1つにまとめる

## 詰まりやすいところ

- `package xxx is not in std` や `cannot find module providing package` と出る → importの文字列と、`go.mod` のモジュールパス＋ディレクトリ名を、1文字ずつ見比べる
- `use of internal package ... not allowed` と出る → importしている側のディレクトリが、`internal` の親より下にあるかを見る
- 移した後、`undefined: xxx` が大量に出る → パッケージをまたいだ名前は、先頭が大文字か。同じパッケージだったときの前提が崩れている
- `import cycle not allowed` と出る → AがBを、BがAをimportしていないか。どちらか一方だけが知っていればよい構成を考える
- `go build` が `inconsistent vendoring` で止まる → `go.mod` と `vendor/modules.txt` の食い違い。`go mod vendor` をやり直す
- 課題10 (b) でもネットワークを使おうとする → `GOFLAGS`、`GOPROXY`、`GOMODCACHE` が、そのシェルで本当に設定されているかを `go env` で見る。`go vet` や `go test` にも `-mod=vendor` が効いているか

## 完了条件の確かめ方

```sh
./build-offline.sh && echo OK
go vet ./... && go test -race -cover ./...
```

- `build-offline.sh` が、`GOPROXY=off` と空の `GOMODCACHE` の下で、`-mod=vendor` により3つの実行ファイルを作り、テストを通して0で終わる
- `bin/csvstat` と `bin/logagg` が、G11の `compare.sh`（コピーして、パスを直す）を通る
- `tree -d -L 2`（なければ `find . -maxdepth 2 -type d`）で、`cmd/`、`internal/`、`vendor/` の構成が見える。`internal/` の外に、`main` 以外のパッケージがない
- `explain.md` を、何も見ずに書いてある。平日に、AIに採点を頼む（[`prompts/grade-explain.md`](../../../prompts/grade-explain.md)）

## 残すもの

- `cmd/`、`internal/`、`go.mod`、`go.sum`、`vendor/`（コミットすると決めた場合）、`build-offline.sh`、`schema.sql`
- `notes.md`（課題2・6〜9・11の観察結果と、課題4・8の決めたこと）、`explain.md`
- `bin/`、`testdata/access.log` はコミットしない

この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- G19・G20（HTTPクライアント、SSE）：新しく書くコードを、この構成（`cmd/` と `internal/`）に置く
- O7（閉域の準備）：`build-offline.sh` の手順を、閉域環境の準備チェックリストの1項目として組み込む。npmの依存も同じ考え方で固める
- C2〜C6（卒業制作）：`capstone/` は、最初からこの構成で始める。`docker build` の中で `-mod=vendor` を使う（C6）
