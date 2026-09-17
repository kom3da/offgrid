---
title: "F：基礎（Stage 0）"
---

プログラミングの前に、コンピュータ・Linux・Git・通信の仕組みを押さえるトラック。ここを飛ばすと、後のステージでエラーの意味が読めなくなる。Stage 0のすべてで、ユニットは F1 から F10 まで順に進む。

## このトラックで身につくこと

- 数も文字も「バイトの並び」であることを、自分の目で確かめられる
- マウスを使わずに、ファイル、プロセス、権限を扱える
- 小さなコマンドをパイプでつなぎ、ログから欲しい情報を取り出せる
- 失敗したらきちんと止まるシェルスクリプトを書ける
- Gitで変更を記録し、ブランチを使い、GitHubとやり取りできる
- 「つながらない」ときに、名前解決・経路・ポート・HTTPのどこで止まっているかを切り分けられる
- コンテナを起動し、中に入り、止められる

## ユニット

各ユニットのページが、課題シート（答えのない問題集）になっている。

- **[F1 コンピュータの仕組み](../drills/foundation/F01-binary/TASKS.md)**：「あ」がUTF-8で何バイトか、`xxd` の出力から説明できる
- **[F2 シェル入門](../drills/foundation/F02-shell/TASKS.md)**：`chmod 755` の意味を説明し、何も見ずにファイル操作できる
- **[F3 テキスト処理](../drills/foundation/F03-pipeline/TASKS.md)**：アクセスログから、アクセス数上位10件のIPを1行で出せる
- **[F4 プロセスとOS](../drills/foundation/F04-process/TASKS.md)**：長時間動くコマンドをバックグラウンドで動かし、探して止められる
- **[F5 シェルスクリプト](../drills/foundation/F05-script/TASKS.md)**：日付付きでディレクトリをバックアップするスクリプトを書け、`shellcheck` の指摘がゼロ
- **[F6 エディタとGit入門](../drills/foundation/F06-git/TASKS.md)**：意味を分かったうえで、自分の手でコミットを積める
- **[F7 Gitの応用とGitHub](../drills/foundation/F07-github/TASKS.md)**：わざとコンフリクトを作り、解消してpushできる
- **[F8 ネットワークの基礎](../drills/foundation/F08-network/TASKS.md)**：名前解決からTCP接続までの流れを、`dig` と `ss` の出力で示せる
- **[F9 HTTPの基礎](../drills/foundation/F09-http/TASKS.md)**：保存した `curl -v` の出力を4段階に分け、HTTPの部分を1行ずつ説明できる
- **[F10 パッケージ管理とDocker入門](../drills/foundation/F10-docker/TASKS.md)**：PostgreSQLのコンテナを起動し、中に入れる（Stage 1の準備）

## 進め方

- 課題は、すべてUbuntuのターミナルで行う。macOSやWindowsのターミナルでは行わない（同じ名前のコマンドでも、動きやオプションが違う。環境の用意は[はじめかた](02-getting-started.md)）
- コマンドを試す場所は `~/sandbox/`。メモと `explain.md` は、リポジトリの `drills/foundation/<ユニット>/` に残す
- コマンドは、使う前に `man` か `--help` で引数を確認する癖をつける
- 実行したコマンドと結果を、メモに貼っておく。後で自分用の辞書になる
- F6でGitを学ぶまでは、[はじめかた](02-getting-started.md)の「決まり文句」でコミットする
- F3で使うダミーのアクセスログは、`scripts/gen-access-log.sh` で作る

## つまずきやすいところ

- 新しいUbuntuでは、`ls` などの基本コマンドがRust製の互換実装に置き換わっていて、`man ls` の見た目が書籍と少し違う。使い方は同じ
- `sudo` やログインで打つパスワードは、画面に何も表示されないが、入力されている
- F10でDockerを入れたあと、`sudo` なしで使えるようにする設定は、Ubuntuに入り直すまで効かない（入り直し方は、F10のページにOS別に書いてある）

## F10で使う起動例

```sh
docker run -d --name pg-offgrid -e POSTGRES_PASSWORD=drill -p 127.0.0.1:5432:5432 -v pgdata:/var/lib/postgresql postgres:18
docker exec -it pg-offgrid psql -U postgres
```

- 公式イメージは18から、データの置き場所が `/var/lib/postgresql` に変わった（17以前は `/var/lib/postgresql/data`）。イメージのバージョンを変えるときは、マウント先も確認する
- `-p 127.0.0.1:5432:5432` は、自分のマシンからだけ接続できるようにする指定

## 使う教材

『コンピュータはなぜ動くのか』、『新しいLinuxの教科書』、『Pro Git』、『マスタリングTCP/IP 入門編』、『Webを支える技術』。入手先は[教材とツール](20-resources.md)にある。
