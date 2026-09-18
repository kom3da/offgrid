---
title: "O4 Dockerfile"
---

> トラック：[O：運用](../../../docs/14-track-o-ops.md) ／ Stage 4 ／ 目安：メイン枠で2回 ／ 前提：F10、G17、G18

## ねらい

自分のAPIを、どのマシンでも同じように動く、小さなコンテナイメージにする。閉域の現場へは、イメージをファイルにして持ち込む（O7）。小さいイメージは、持ち込みが楽で、中に余計なものがないぶん、安全でもある。「何を入れ、何を入れないか」を、自分で決められるようになる。

## このユニットで身につけること

- Dockerfileを書き、`docker build` でイメージを作る
- イメージが「層（レイヤー）」の積み重ねでできていることを、`docker history` で確かめる
- ビルド用の段と、実行用の段を分ける（マルチステージビルド）
- `CGO_ENABLED=0` で、ほかのライブラリに頼らない静的バイナリを作り、そのことを確かめる
- シェルすら入っていない distroless イメージで、APIを動かす
- ビルドのキャッシュが効く順番に、Dockerfileの命令を並べる
- ネットワークなしでビルドできる形（`-mod=vendor`）にする

## キーワード

- **Dockerfile**：イメージの作り方を、上から順に書いたテキストファイル
- **ベースイメージ**：`FROM` で指定する、土台のイメージ
- **レイヤー**：Dockerfileの命令ごとに積み重なる、ファイルの差分。イメージの大きさは、この合計
- **ビルドコンテキスト**：`docker build` のときに、Dockerに渡されるディレクトリの中身。`.dockerignore` で除外できる
- **マルチステージビルド**：1つのDockerfileの中に `FROM` を複数書き、前の段で作った成果物だけを、後の段にコピーする方法
- **cgo**：GoからC言語のライブラリを呼ぶ仕組み。有効だと、バイナリが実行環境のCライブラリに依存することがある
- **静的バイナリ**：必要なものをすべて自分の中に持ち、ほかの共有ライブラリなしで動く実行ファイル
- **distroless**：アプリの実行に必要な最小限だけを入れたイメージ。シェルもパッケージ管理もない

## 平日に読むもの

- 『［試して理解］Linuxのしくみ』第11章（コンテナ）、第12章（cgroup）
- Dockerの公式ドキュメントの「Dockerfile reference」「Multi-stage builds」「Building best practices」を、オフラインで読めるように保存する
- distroless（GoogleContainerTools/distroless）の `README.md` を保存する。イメージの種類と、タグの種類（root でないユーザーで動くもの、デバッグ用）の説明を読む
- `go help build`、`go help environment`、`go doc cmd/link`（リンカのフラグ）
- **平日のうちに、使うイメージを `docker pull` しておく**。AIなしデーには取りに行かない前提で進める
  - `golang` の公式イメージ（自分の `go.mod` のバージョンに合うタグ）
  - distroless の、静的バイナリ用のイメージ（名前とタグは `README.md` で確かめる）
  - 比較用に、`debian` か `ubuntu` の公式イメージ

## 準備

```sh
mkdir -p ~/sandbox/o4
cp -r <G17のAPIのディレクトリ> ~/sandbox/o4/api
cd ~/sandbox/o4/api
go mod vendor
go build -mod=vendor ./... && go test -mod=vendor ./...
docker images
```

最後の `docker images` で、平日に取っておいたイメージが並んでいることを確かめる。Dockerfileとメモは、リポジトリの `drills/infra/O04-dockerfile/` にも写しを置く。

## 課題

1. まず、コンテナを使わずに観察する。`go build` でふつうに作ったバイナリと、`CGO_ENABLED=0 go build` で作ったバイナリを、`ls -l`、`file`、`ldd` で比べる。`ldd` の結果はどう違うか。それは、バイナリを「別のマシンに持っていく」ときに、何を意味するか
2. 1段だけのDockerfileを書く。`golang` のイメージを土台に、ソースをコピーし、ビルドし、起動する。`docker build -t api:step1 .` で作り、`docker images` で大きさを記録する。以後、各段階の大きさを `sizes.md` に記録していく
3. `docker history api:step1` を読む。いちばん大きいレイヤーは、どの命令が作ったものか。自分のバイナリは、全体のうち何MBか
4. ソースを1文字だけ変えて、もう一度ビルドする。どの命令からやり直しになったか。`go.mod` と `vendor/` を先にコピーする場合と、全部を一度にコピーする場合で、やり直しの範囲はどう変わるか。ビルドの出力の、キャッシュが使われたことを示す表示を見つける
5. `.dockerignore` を書く。何を除外するべきか（`.git`、`.env`、テストの出力など）を、自分で決める。除外の前後で、ビルドの最初に表示される「Dockerに送った量」は、どう変わるか。`.env` がイメージに入ってしまうと、何が困るか
6. マルチステージにする。1段目でビルドし、2段目（`debian` か `ubuntu`）にはバイナリだけをコピーする。「Multi-stage builds」のページで、段に名前を付ける書き方と、前の段からコピーする書き方を調べる。`api:step2` として作り、大きさを記録する
7. `api:step2` を `docker run` で起動し、`curl` で確かめる。DBの接続先は、環境変数で渡す（`-e` か `--env-file`。値をDockerfileに書かない）。コンテナの中から見た `localhost` は、どこを指すか。`pg-offgrid` のコンテナにつなぐには、どうすればよいか（Dockerのドキュメントの、ユーザー定義のブリッジネットワークの説明を読む）
8. **本題**：2段目を distroless の静的バイナリ用のイメージに変え、`api:step3` として作る。1段目のビルドには `CGO_ENABLED=0` と `-mod=vendor` を指定する。`CGO_ENABLED=0` を付けずに作ったバイナリを distroless で動かすと、どうなるか（環境によっては動いてしまう。その場合は、課題1の `ldd` の結果から、理由を考える）。大きさを記録し、30MBを超えていたら、何が場所を取っているかを `docker history` で調べる
9. `docker exec -it <コンテナ> sh` で、`api:step3` のコンテナの中に入ろうとすると、どうなるか。なぜか。シェルがないことの、良い点と困る点を書く。困ったときのための道具が、distroless の `README.md` に載っているので、探す
10. `api:step3` のコンテナの中で、APIはどのユーザーで動いているか。`docker inspect` か `docker top` で確かめる。root でないユーザーで動かす方法を、Dockerfile reference と distroless の `README.md` の両方で調べて、設定する
11. `go help build` と `go doc cmd/link` で、バイナリを小さくするリンカのフラグ（デバッグ情報を省くもの）を探し、付けた場合と付けない場合の大きさを比べる。省いたことで、何ができなくなるか（O3で使った道具を思い出す）
12. `docker stop` でコンテナを止める。G16の終了処理のログは、`docker logs` に出ているか。止まるまでに10秒かかる場合は、Dockerfileの起動コマンドの書き方（シェル形式と exec 形式）の違いを、Dockerfile reference で調べる

### 発展（任意）

- 2段目を `scratch`（空のイメージ）にする。distroless と比べて、何が足りなくなるか。APIから外へ HTTPS でつなぐ処理や、タイムゾーンを扱う処理があったら、どうなるか
- ビルドのときにバージョン文字列（Gitのコミットのハッシュなど）をバイナリに埋め込み、起動時のログに出す。`go doc cmd/link` の `-X` の説明を読む
- 1段目の中で `go test` も実行し、テストが落ちたらイメージができないようにする

## 詰まりやすいところ

- `docker build` の途中で、Goが依存をダウンロードしようとする（または、しようとして失敗する）→ `vendor/` がコピーされているか、`.dockerignore` で除外していないか、ビルドのコマンドに何を指定したか
- コンテナを起動すると、`no such file or directory` や `exec format error` ですぐ終わる → バイナリのパス、実行権限、課題1の `file` と `ldd` の結果を確かめる
- コンテナのAPIが、DBにつながらない → コンテナの中の `localhost` が、どこを指すかを考える。`docker network ls` と `docker network inspect` で、2つのコンテナが同じネットワークにいるかを確かめる
- `curl` がつながらない → `-p` の指定と、APIが待ち受けているアドレス（`127.0.0.1` か、すべてのアドレスか）を、コンテナの中の視点で考える
- イメージの大きさが、思ったほど減らない → `docker images` で見ているタグが、作り直したものか。`docker history` で、どのレイヤーが大きいかを見る
- 古いイメージとコンテナで、ディスクが埋まる → `docker system df` で使用量を見る（F10）

## 完了条件の確かめ方

- `docker images api:step3` の `SIZE` が30MB以下
- `api:step3` のコンテナに対して、`curl` で全エンドポイントの確認ができる（DBは `pg-offgrid`）
- `sizes.md` に、`step1`、`step2`、`step3` の大きさと、減った理由が1行ずつ書いてある
- ネットワークにつながずに、`docker build` が最後まで通る（ビルドの出力に、依存のダウンロードが出ていない）

## 残すもの

- `Dockerfile`（最終版）と `.dockerignore`
- `sizes.md`（各段階の大きさと、その理由）
- 各課題で観察したこと（`notes.md`）
- `.env` や、値の入った環境変数のファイルはコミットしない

コマンドを試す場所は `~/sandbox/o4`。メモとDockerfileの写しは、リポジトリの中のこのディレクトリに、AI機能を切ったエディタで書く。この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- O5（docker compose）：ここで作ったイメージを、PostgreSQL、リバースプロキシと一緒に起動する
- O7（再現できるビルド）：ベースイメージをダイジェストで固定し、`docker save` で持ち運ぶ
- C6（配布の形にする）：卒業制作の server を、同じ方法でイメージにする
