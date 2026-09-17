# F10：パッケージ管理とDocker入門（課題シート）

この `README.md` は書き換えない。自分が書いたものは、同じディレクトリに別のファイルとして置く。

## ねらい

道具を自分で入れられるようになる。コンテナを起動し、中に入り、止められるようになる。Stage 1で使うPostgreSQLを用意する。

## 平日に読むもの

『新しいLinuxの教科書』のパッケージ管理の章。Dockerの公式ドキュメントの「Get started」（オフライン用に保存しておく）。

## 課題

1. `apt`：`apt search`、`apt show`、`apt list --installed | wc -l` を試す。`sudo apt install -y tree` で入れ、使ってみて、`sudo apt remove tree` で消す。`sudo` とは何か。`apt update` と `apt upgrade` の違いは何か
2. Dockerを入れる

   ```sh
   sudo apt install -y docker.io
   sudo usermod -aG docker "$USER"
   ```

   2行目は、`sudo` なしで `docker` を使えるようにする設定。**反映するには、Ubuntuに入り直す必要がある。** WSL2ならPowerShellで `wsl --shutdown`、Limaなら `limactl stop offgrid` のあと `limactl start offgrid`、Linuxならログアウトしてログインし直す。入り直したら、`docker run --rm hello-world` で確かめる
3. イメージとコンテナの違いを調べて書く。`docker images`、`docker ps`、`docker ps -a` の違いを確かめる
4. `docker run -it --rm ubuntu bash` でコンテナの中に入る。中で `ls /`、`ps aux`、`cat /etc/os-release` を実行し、外（自分のUbuntu）での結果と比べる。中でファイルを作って `exit` し、もう一度同じコマンドで入ると、そのファイルはあるか。なぜか
5. `docker run -d --name web -p 8080:80 nginx` を実行する。`curl localhost:8080` → `docker logs web` → `docker exec -it web sh` → `docker stop web` → `docker rm web` の順に試す。`-p 8080:80` の左と右は、それぞれ何の番号か。`-d` を付けないと、どうなるか
6. **本題**：[Stage 0：基礎トラック](../../../docs/03-foundation.md)の「F10で使う起動例」で、PostgreSQLを起動する。コマンドのオプション（`-d`、`--name`、`-e`、`-p`、`-v`）を、1つずつ説明する。`psql` に入り、`\l` と `\q` を試す
7. 永続化：`psql` の中で、テーブルを1つ作る（`create table t (x int);`）。コンテナを `docker rm -f pg-offgrid` で消し、同じコマンドで起動し直す。テーブルは残っているか。`-v` を外して同じことをすると、どうなるか
8. 後片付け：`docker volume ls`、`docker system df` で、ディスクの使用量を見る。使っていないものの消し方を調べる

## 完了条件の確かめ方

何も見ずに、PostgreSQLのコンテナを起動し、`psql` で中に入り、出られる。`explain.md` に、課題6のオプションの説明を書く。

## 残すもの

`explain.md`、`notes.md`。
