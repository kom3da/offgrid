# Stage 0：基礎トラック

プログラミングの前に、コンピュータ・Linux・Git・通信の仕組みを押さえる。ここを飛ばすと、後のステージでエラーの意味が読めなくなる。課題はすべてターミナルで行い、`drills/foundation/<ユニット>-<名前>/` に記録する。

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| F1 | コンピュータの仕組み | CPU・メモリ・ストレージの役割、2進数・16進数、文字コード（UTF-8）を学び、`xxd`でファイルの中身を見る | 「あ」がUTF-8で何バイトか、`xxd`の出力から説明できる |
| F2 | シェル入門 | `ls`、`cd`、`cp`、`mv`、`rm`、`mkdir`、`find`、絶対パスと相対パス、パーミッション | `chmod 755`の意味を説明し、何も見ずにファイル操作できる |
| F3 | テキスト処理 | `cat`、`less`、`grep`、`sort`、`uniq`、`wc`、`cut`、パイプとリダイレクト、`xargs`。正規表現（文字列のパターンを表す書き方）の初歩と、`sed`・`awk`の1行処理 | アクセスログからアクセス数上位10件のIPを1行で出せる |
| F4 | プロセスとOS | プロセス、シグナル、`ps`、`top`、`kill`、環境変数、標準入力・出力・エラー | 長時間動くコマンドをバックグラウンドで動かし、止められる |
| F5 | シェルスクリプト | 変数、`if`、`for`、関数、終了コード、`set -euo pipefail` | 日付付きでディレクトリをバックアップするスクリプトを書ける |
| F6 | エディタとGit入門 | ターミナルのエディタ（vimかnano）、`git init`、`add`、`commit`、`log`、`diff`、`.gitignore` | このリポジトリに、自分の手でコミットを積める |
| F7 | Gitの応用とGitHub | SSH鍵の作成（`ssh-keygen`）とGitHubへの登録、`~/.ssh/config`。ブランチ、マージ、コンフリクト解消、`rebase`、`push`、プルリクエスト | わざとコンフリクトを作り、解消してpushできる |
| F8 | ネットワークの基礎 | IPアドレス、ポート、TCPとUDP、DNS、`ip`、`ping`、`ss`、`dig` | 名前解決（DNS）からTCP接続までの流れを何も見ずに `explain.md` に書き、`dig`と`ss`（macOSは`netstat`）の出力で示せる |
| F9 | HTTPの基礎 | リクエストとレスポンス、メソッド、ステータスコード、ヘッダ、TLSの概要、`curl -v` | `curl -v`の出力を保存し、1行ずつの意味（DNS→TCP→TLS→HTTPのどの段階か）を何も見ずに `explain.md` に書ける |
| F10 | パッケージ管理とDocker入門 | `apt`（またはHomebrew）、コンテナとは何か、`docker run`、`ps`、`logs`、`exec` | PostgreSQLのコンテナを起動し、中に入れる（Stage 1の準備） |

## Stage 0で意識すること

- コマンドは、使う前に`man`か`--help`で引数を確認する癖をつける
- 実行したコマンドと結果を、振り返りメモに貼っておく。後で自分用の辞書になる
- 壊してもよい練習用ディレクトリ（`~/sandbox`など）で試す
- シェルスクリプトは `shellcheck` で検査する
- macOSには`ip`や`ss`がない。F8は`ifconfig`、`netstat`、`lsof -i`で代用するか、Linuxの仮想マシンで行う

## F10で使う起動例

```sh
docker run -d --name pg-offgrid -e POSTGRES_PASSWORD=drill -p 127.0.0.1:5432:5432 -v pgdata:/var/lib/postgresql postgres:18
docker exec -it pg-offgrid psql -U postgres
```

- 公式イメージは18から、データの置き場所が `/var/lib/postgresql` に変わった（17以前は `/var/lib/postgresql/data`）。イメージのバージョンを変えるときは、マウント先も確認する
- `-p 127.0.0.1:5432:5432` は、自分のマシンからだけ接続できるようにする指定
