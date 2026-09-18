---
title: "O7 再現できるビルド"
---

> トラック：[O：運用](../../../docs/14-track-o-ops.md) ／ Stage 4 ／ 目安：メイン枠で2〜3回 ／ 前提：O1、O4、O5、O6、G18、T10

## ねらい

「自分の手元では動く」を「リポジトリを渡せば、誰の手元でも同じものが作れる」に変える。ビルドが何をどこから取ってきているかを全部言えるようにし、それを固定し、まっさらな環境で同じ成果物ができることを確かめる。O8 で他人のコードを引き継ぐとき、C7 で runbook から作り直すとき、ここで身につけた見方をそのまま使う。

## このユニットで身につけること

- ビルドが何を取りに行っているかを、観察して一覧にできる
- Go の依存を `vendor/` に固め、`-mod=vendor` でビルドできる
- npm の依存を lockfile とキャッシュで固め、`npm ci` で同じ結果を出せる
- Docker のベースイメージをダイジェストで固定し、`docker save` / `docker load` で持ち運べる
- 成果物のハッシュを取り、別の環境で同じものができたことを確かめられる
- ビルドの手順をスクリプトにして、手順書の代わりにできる

## キーワード

- **再現できるビルド**：同じ入力から、いつ・どこでやっても同じ成果物ができること
- **lockfile**：依存の版を固定するファイル（`go.sum`、`package-lock.json`）
- **`vendor/`**：Go の依存のソースを、リポジトリの中に置いたもの
- **ダイジェスト**：イメージの中身から計算した識別子。タグと違い、中身が変わると変わる
- **`docker save` / `docker load`**：イメージをファイルにする／ファイルから戻す
- **ハッシュ（`sha256sum`）**：ファイルの中身の指紋。同じ中身なら同じ値

## 平日に読むもの

- Go の公式ドキュメント「Go Modules Reference」の Vendoring の節と、`go help mod vendor`
- npm の公式ドキュメントの `npm ci` と `package-lock.json` のページ
- Docker の公式ドキュメントの `docker save`、`docker load`、イメージのダイジェストの説明
- 『［試して理解］Linuxのしくみ』のファイルシステムの章（章番号は、手元の本の目次で確認する）
- 課題9で使う、まっさらな仮想マシンを O6 の `target-setup.md` の手順で作り直しておく（コンテナの中では `docker build` ができないので、VM にする）

## 準備

O4 でコピーした G18 の API と、T10 の web を使う。このユニットで「リポジトリ」と言うのは `~/sandbox/o7`（`git init` した1つのリポジトリ）のこと。課題9では、これだけを別の環境に持ち込む。

```sh
mkdir -p ~/sandbox/o7
cp -r ~/sandbox/o4/api ~/sandbox/o7/api
cp -r <T10 のディレクトリ> ~/sandbox/o7/web
rm -rf ~/sandbox/o7/api/vendor      # O4 で作った vendor/ を一度消す。課題1で「取りに行く」のを見るため
cd ~/sandbox/o7 && git init && git add -A && git commit -m "start"
```

## 課題

1. ビルドが外に何を取りに行くかを観察する。`~/sandbox/o7/api` で `go clean -modcache` してから `go build ./...` を実行し（準備で `vendor/` を消してあるので、取りに行く）、そのあいだ別のターミナルで `ss -tn state established` を数秒おきに見る。どこへ接続し、何を取得したかを `notes.md` に書く。web でも `rm -rf node_modules` → `npm ci` で同じことをする
2. `go mod vendor` を実行し、`vendor/` に何が入ったかを見る。`go build -mod=vendor ./...` と `go test -mod=vendor ./...` が通ることを確かめる。`vendor/modules.txt` は何を記録しているか
3. もう一度 `go clean -modcache` してから、`-mod=vendor` でビルドする。今度は `ss` に何が出るか。課題1と比べて、なぜ違うのかを書く
4. web で `package-lock.json` を消して `npm install` すると何が変わるか、`git diff` で見る。戻してから、`npm ci` と `npm install` の違いを `npm help ci` で調べ、なぜ自動のビルドでは `ci` を使うのかを書く
5. `npm ci` が取りに行くものを npm のキャッシュに入れる。`npm config get cache` でキャッシュの場所を見つけ、`npm ci --offline` が通ることを確かめる。通らなければ、何が足りないかをエラーの文から読む
6. O4 の Dockerfile のベースイメージを、タグではなくダイジェスト（`image@sha256:…`）で指定する。ダイジェストは `docker images --digests` で調べる。タグで指定したときと、何が違うかを書く
7. **本題（その1）**：ビルドの手順を `build.sh` にする。入力は `~/sandbox/o7` のリポジトリだけ。出力は API のバイナリ、web の `dist/`、Docker イメージの tar（`docker save`）。バイナリと `dist/` の各ファイルの `sha256sum` を `checksums.txt` に書き出す（tar は書かない。理由は課題8で分かる）。最初に `go version`、`node --version`、`npm --version` を出力する。`set -euo pipefail` を付け、`shellcheck` を通す
8. `build.sh` を2回続けて実行し、2つの `checksums.txt` を比べる。同じにならないものがあれば、何が変わっているのか（時刻、パス、並び順）を突き止める。Go のバイナリは `-trimpath` を付けると変わるか。**2回目でバイナリだけが変わる**なら、`go version -m <バイナリ>` の `build` の行を2回分比べ、何が違うかを `go help build` で調べる。tar の `sha256sum` も2回分取り、なぜ揃わないのかを `docker history` の時刻から考える
9. **本題（その2）**：まっさらな環境で再現する。平日に用意した仮想マシンに、`~/sandbox/o7` のリポジトリだけを持ち込み、`build.sh` を実行する。最初に出る `go version` などが手元と同じことを見てから、`checksums.txt` が手元と一致することを確かめる。一致しない、または途中で止まるなら、リポジトリに足りないものは何かを書く
10. `docker load` で tar からイメージを戻し、O5 の compose で起動する。動いたら、イメージを消して（`docker rmi`）もう一度 `load` → 起動をやる。イメージの中のバイナリを取り出し（`docker create` → `docker cp`）、その `sha256sum` が課題7のバイナリと同じかを見る
11. `build.sh` の各段階に「何を入力にして、何を出力するか」をコメントで書く。人が読む手順書は別に書かない。スクリプトが手順書になっているかを、課題9をもう一度やって確かめる

### 発展（任意）

- `GOFLAGS=-mod=vendor` を環境変数にしたときと、コマンドの引数にしたときの違いを調べる
- `docker build` に `--no-cache` を付けたときと付けないときで、イメージから取り出したバイナリ（課題10）は変わるか。tar はどうか

## 詰まりやすいところ

- `-mod=vendor` で「inconsistent vendoring」と出る → `go.mod` と `vendor/modules.txt` の食い違い。`go mod vendor` をやり直す前に、`go mod tidy` で何が変わるかを `git diff` で見る
- `npm ci --offline` が失敗する → キャッシュに無いパッケージがある。エラーに出た名前を、`npm cache ls` の出力と突き合わせる
- `checksums.txt` が毎回変わる → どのファイルが変わっているかを `sha256sum -c` で1つずつ絞る。ビルドの時刻を埋め込んでいるものが多い
- Go のバイナリだけ、2回目で変わる → 1回目の実行で `checksums.txt` ができ、リポジトリに未追跡のファイルが増えている。`go version -m` の `vcs.modified` を見て、`go help build` で `-buildvcs` を読む
- 新しい環境で `go` が別の版を取りに行く → `go.mod` の `go` の行と `GOTOOLCHAIN` の関係を `go help toolchain` で読む。手元と同じ版を入れる
- 新しい環境でだけ失敗する → 手元にあって新しい環境に無いものを探す。`which`、`env`、`ls ~/.config` の差
- `docker save` の tar が大きすぎる → `docker history` で層を見る。O4 のマルチステージビルドになっているか

## 完了条件の確かめ方

1. `~/sandbox/o7/api` で `go clean -modcache && go build -mod=vendor ./... && go test -mod=vendor ./...` が、`ss` に外向きの接続を出さずに通る
2. `build.sh` を手元で2回実行して `checksums.txt`（バイナリと `dist/` の各ファイル）が同じになり、まっさらな仮想マシンで実行しても同じになる（3つの `checksums.txt` と、それぞれの `go version` の出力を `notes.md` に貼る）。tar の一致は求めない
3. `docker load` で戻したイメージで、O5 の compose が起動し、イメージから取り出したバイナリの `sha256sum` が課題7のものと同じになる

## 残すもの

- `build.sh` と `checksums.txt`（`~/sandbox/o7` にあるものの写しを、このディレクトリに置く。手元2回分と、まっさらな環境の分）
- `notes.md`（課題1・3・4・8・9・10 の観察）
- `~/sandbox/o7` の `vendor/` はコミットしてよい。`node_modules` と tar はコミットしない

ファイルは、リポジトリの中のこのディレクトリに、AI機能を切ったエディタで書く。この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- O8（引き継ぎ）：渡されたリポジトリを、ここで身につけた目で読む
- C6（配布の形にする）：`build.sh` の考え方を watchdeck に持ち込む
- C7（障害試験と runbook）：runbook だけから再構築する
