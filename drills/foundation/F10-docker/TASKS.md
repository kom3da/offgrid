---
title: "F10 パッケージ管理とDocker入門"
---

> トラック：[F：基礎](../../../docs/10-track-f-foundation.md) ／ Stage 0 ／ 目安：メイン枠で1〜2回 ／ 前提：F2、F4、F8、F9

## ねらい

道具を自分で入れられるようになる。コンテナを起動し、中に入り、止め、消せるようになる。閉域の現場では、必要なものを持ち込んで自分で入れるしかなく、「入れ方はAIに聞く」が通用しない。ここで入れたDockerと、起動したPostgreSQLは、Stage 1（D1〜）でそのまま使う。

## このユニットで身につけること

- `apt` で、パッケージを探す・調べる・入れる・消す
- `sudo` が何をしているかを言い、`sudo` なしで失敗する場面を見分ける
- Dockerを入れ、自分のユーザーで `docker` を使えるようにする
- イメージとコンテナの違いを、`docker images` と `docker ps -a` の出力で示す
- コンテナの中に入り、外（自分のUbuntu）と何が同じで何が違うかを言う
- `docker run` の `-d`、`--name`、`-e`、`-p`、`-v`、`--rm` を、1つずつ説明する
- `docker ps`、`logs`、`exec`、`stop`、`rm` で、コンテナの一生を扱う
- コンテナを消してもデータが残る条件を、実験で確かめる

## キーワード

分からない語は、`man apt` と、平日に読むDockerのドキュメントで調べる。

- **パッケージ**：プログラムと、その設定や依存関係の情報をまとめた、配布の単位
- **リポジトリ（apt）**：パッケージの配布元。Gitのリポジトリ（F6）とは別物
- **依存関係**：あるパッケージが動くために必要な、別のパッケージ
- **`sudo`**：管理者（root）の権限で、コマンドを1つ実行する
- **コンテナ**：1台のLinuxの中で、プロセスとファイルを隔離して動かす仕組み。カーネル（OSの中心部）は外と共有する
- **イメージ**：コンテナの元になる、ファイル一式の雛形。読み取り専用
- **レジストリ**：イメージの配布元。何も指定しないと Docker Hub から取ってくる
- **デーモン**：裏で動き続けるプログラム。Dockerの本体もその1つで、`docker` コマンドはそれに指示を出す窓口
- **ポートの公開（`-p`）**：コンテナのポートを、自分のマシンのポートにつなぐ（F8）
- **ボリューム（`-v`）**：コンテナを消しても残る、データの置き場所

## 平日に読むもの

- 『新しいLinuxの教科書』CHAPTER18（アーカイブと圧縮）、CHAPTER20（ソフトウェアパッケージ）
- 『［試して理解］Linuxのしくみ』第10章（仮想化機能）、第11章（コンテナ）
- Docker公式ドキュメントの Get started（オフラインに保存したもの）の、コンテナとイメージの説明と、`docker run` の使い方
- `man apt`。`docker run --help`（`man docker-run` が入っていれば、そちらも）

## 準備

```sh
mkdir -p ~/sandbox/f10 && cd ~/sandbox/f10
```

シェルを2つ開いておく。課題3でDockerを入れるまで、`docker` コマンドは使えない。

## 課題

1. `apt` で探す・調べる：`apt search tree`、`apt show tree`、`apt list --installed | wc -l` を試す。`apt show` の出力の各行（バージョン、依存関係、説明）を読む。`apt show curl` のように、既に入っているものを見ると、何が違うか
2. 入れて消す：`sudo apt install -y tree` で入れ、`tree ~/sandbox` で使ってみて、`sudo apt remove tree` で消す。入れる前と後で `which tree` はどう変わるか。`sudo` を付けずに `apt install` すると、どうなるか。`sudo` とは何か。`apt update` と `apt upgrade` の違いは何か。`apt remove` と `apt purge` の違いを `man apt` で探す
3. Dockerを入れる

   ```sh
   sudo apt install -y docker.io
   sudo usermod -aG docker "$USER"
   ```

   2行目は、`sudo` なしで `docker` を使えるようにする設定。入り直す前に `docker ps` を実行すると、どうなるか。**反映するには、Ubuntuに入り直す必要がある。** WSL2ならPowerShellで `wsl --shutdown`、Limaなら `limactl stop offgrid` のあと `limactl start offgrid`、Linuxならログアウトしてログインし直す。入り直したら、`groups` に `docker` が増えていることを見て、`docker run --rm hello-world` で確かめる。表示された文章には、Dockerが今やった手順が書いてある。何段階か
4. イメージとコンテナ：`docker run --rm hello-world` をもう一度実行する。1回目と出力はどう違うか（1回目にだけあった行は何か）。`docker images`、`docker ps`、`docker ps -a` を実行し、それぞれ何が出るか。`--rm` を付けずに `hello-world` を実行してから `docker ps -a` を見ると、何が増えるか。イメージとコンテナの違いを、この出力を証拠にして書く
5. `docker run -it --rm ubuntu bash` でコンテナの中に入る。中で `ls /`、`ps aux`、`cat /etc/os-release`、`uname -r`、`hostname` を実行し、外（自分のUbuntu）での結果と比べる。同じものと違うものはどれか。それは「コンテナはカーネルを共有する」と、どう関係するか（中では `sudo` が要らない。なぜか）。中でファイルを作って `exit` し、もう一度同じコマンドで入ると、そのファイルはあるか。なぜか。`exit` した後、`docker ps -a` にこのコンテナは残っているか
6. `docker run -d --name web -p 8080:80 nginx` を実行する。`curl localhost:8080` → `docker logs web` → `docker exec -it web sh` の順に試す。`-p 8080:80` の左と右は、それぞれ何の番号か。`ss -ltn` で、8080番の待ち受けを見つける（F8）。`docker logs web` の1行を、F9のサーバーのログと比べる。`docker exec` で入ったシェルを `exit` しても、コンテナは動いたままか。課題5と何が違うか
7. コンテナの一生：`docker stop web` → `docker rm web` の順に試す。`stop` の後、`curl localhost:8080` と `docker ps`、`docker ps -a`、`docker logs web` はそれぞれどうなるか。`rm` せずにもう一度課題6の `docker run` を実行すると、どうなるか。`-d` を付けないと、どうなるか（止めるのは Ctrl+C）
8. **本題**：[Stage 0：基礎トラック](../../../docs/10-track-f-foundation.md)の「F10で使う起動例」で、PostgreSQLを起動する。`docker logs pg-offgrid` で、起動が終わったことを示す行を探す。`docker ps` の PORTS の列を読む。コマンドのオプション（`-d`、`--name`、`-e`、`-p`、`-v`）を、1つずつ `explain.md` に説明する。`-p` の前に付いている `127.0.0.1:` は、何を制限しているか（F8）。`psql` に入り、`\l` と `\q` を試す
9. コンテナの中を見る：`docker exec -it pg-offgrid bash` で入り、`ls /var/lib/postgresql` を見る。`-v pgdata:/var/lib/postgresql` の左と右は、それぞれ何か。`exit` して、`docker volume ls` に `pgdata` があることを確かめる
10. 永続化：`psql` の中で、テーブルを1つ作る（`create table t (x int);`）。コンテナを `docker rm -f pg-offgrid` で消し、同じコマンドで起動し直す。テーブルは残っているか。`-v` を外して同じことをすると、どうなるか。終わったら、必ず起動例のとおり（`-v` 付き）で起動し直しておく（Stage 1で使う）
11. 後片付け：`docker volume ls`、`docker system df`、`docker images` で、ディスクの使用量と、各イメージの大きさを見る。使っていないものの消し方を調べる。**`pg-offgrid` と `pgdata` は消さない。** 課題10で `-v` を外したときにできたボリュームは、どれか

### 発展（任意）

- `docker run --rm ubuntu cat /etc/os-release` のように、コマンドを1つだけ実行して終わる形を試す。課題5の `-it ... bash` と、何が違うか
- コンテナが動いている状態で、外（自分のUbuntu）から `ps aux` を実行し、コンテナの中のプロセスが見えるかを確かめる。仮想マシン（VM）との違いを、自分の言葉で書く
- `docker inspect pg-offgrid` の出力から、コンテナのIPアドレスを探す。F8で調べたプライベートIPアドレスの範囲に入っているか
- Docker公式の Get started に書かれた入れ方と、今回の `docker.io` パッケージの違いを、`apt show docker.io` の情報から考える

## 詰まりやすいところ

- `permission denied while trying to connect to the Docker daemon socket` → 課題3の入り直しが済んでいるか。`groups` に `docker` があるか
- `Cannot connect to the Docker daemon` → Dockerの本体（デーモン）が動いているかを、`systemctl status docker` で見る
- `Conflict. The container name ... is already in use` → `docker ps -a` で、同じ名前の止まったコンテナを探す（課題7）
- `port is already allocated` や `address already in use` → `ss -ltnp` で誰が使っているか（F8）と、`docker ps` の PORTS の列を見る
- `docker exec -it pg-offgrid psql` がエラーになる → 起動の直後は、まだ準備中のことがある。`docker logs pg-offgrid` で準備完了の行を待つ（課題8）
- `Unable to find image` の後にエラーが出て止まる → イメージを取ってくるのに外のネットワークが要る。F8のやり方で切り分ける。閉域でどう持ち込むかは、Stage 4（O7）で扱う

## 完了条件の確かめ方

何も見ずに、PostgreSQLのコンテナを起動し、`psql` で中に入り、`\l` を実行して、出られる。次回のウォームアップでもう一度やる。課題8の `explain.md`（オプションの説明）は、平日にAIに採点を頼む（[`prompts/grade-explain.md`](../../../prompts/grade-explain.md)）。

## 残すもの

- `explain.md`
- 課題4と5の比較のメモと、試したコマンドと気づきのメモ（`notes.md`）
- `pg-offgrid` が `-v` 付きで動いていることを、`docker ps` で確かめる

コマンドを試す場所は `~/sandbox/f10`。メモは、リポジトリの中のこのディレクトリに、AI機能を切ったエディタ（VS Codeか、F6で選んだ `vim`・`nano`）で書く。この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- D1（データベースとは・psql）：ここで起動した `pg-offgrid` に、毎回 `docker exec -it pg-offgrid psql -U postgres` で入る
- O1（Linuxの管理）：`sudo`、ユーザーとグループ、ディスクの使用量を、管理者の視点でもう一度扱う
- O7（閉域の準備）：外のネットワークなしで `apt` と `docker` を使えるように、ここで入れたものを持ち込む手順にする
