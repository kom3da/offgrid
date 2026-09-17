---
title: "O2 systemd"
---

> トラック：[O：運用と閉域](../../../docs/14-track-o-ops.md) ／ Stage 4 ／ 目安：メイン枠で2回 ／ 前提：O1、G12〜G17

## ねらい

自分で書いたAPIを、ターミナルを閉じても、マシンを再起動しても動き続ける「サービス」にする。閉域の現場では、コンテナが使えず、バイナリ1つと systemd だけで動かす場面がある。卒業制作の agent も、この形で配る。

## このユニットで身につけること

- `systemctl` で、サービスの状態を見て、起動し、止め、自動起動を切り替える
- ユニットファイルを自分で書き、systemd に読み込ませる
- サービスを、root ではない専用のユーザーで動かす
- 異常終了したら自動で再起動するように設定し、実際に落として確かめる
- 接続先などの設定を、ユニットファイルの外から渡す
- `journalctl` で、サービスのログを絞り込んで読む
- マシンの再起動後に、サービスが自動で起動していることを確かめる

## キーワード

- **systemd**：Linuxの起動と、サービスの管理を受け持つプログラム。PID 1（最初に動くプロセス）として動く
- **ユニット**：systemd が管理する対象の単位。サービス（`.service`）、タイマー（`.timer`）などの種類がある
- **ユニットファイル**：ユニットの設定を書いたテキストファイル。`[Unit]`、`[Service]`、`[Install]` のセクションに分かれる
- **ディレクティブ**：ユニットファイルの中の「名前=値」の1行
- **enable と start**：enable は「次の起動時から自動で起動する」、start は「今すぐ起動する」。別々の操作
- **ターゲット**：複数のユニットをまとめた、起動の段階を表すユニット
- **journal**：systemd が集めているログ。`journalctl` で読む
- **`daemon-reload`**：ユニットファイルを書き換えたあと、systemd に読み直させる操作

## 平日に読むもの

- 『［試して理解］Linuxのしくみ』の、プロセス管理を扱う章（F4の復習。シグナルと終了コード）
- 平日のうちに、次の `man` が引けることを確かめる：`man systemctl`、`man journalctl`、`man systemd.unit`、`man systemd.service`、`man systemd.exec`。Stage 4のAIなしデーはWeb検索が0回なので、これが唯一の資料になる
- `man systemd.service` の、サービスの種類、起動コマンド、再起動の条件を説明している部分を、先に眺めておく
- Dockerの公式ドキュメントの「Start containers automatically」のページを保存しておく（課題10で使う）

## 準備

systemd が動いていることを確かめる。

```sh
systemctl is-system-running
ps -p 1 -o comm=
```

1行目が `running` か `degraded`、2行目が `systemd` なら進める。`offline` と出る、またはエラーになるWSL2の環境では、systemd が有効になっていない。平日に、MicrosoftのWSLの公式ドキュメントで「systemd」を扱うページを読み、有効にしておく（設定ファイル `/etc/wsl.conf` を編集し、WSLを再起動する手順が載っている）。

```sh
mkdir -p ~/sandbox/o2
cd ~/sandbox/o2
```

G17のAPIを、`go build` でバイナリにしておく。DBは、これまでどおり `pg-offgrid` のコンテナを使う。ユニットファイルの写しとメモは、リポジトリの `drills/infra/O02-systemd/` に置く。

## 課題

1. `systemctl status ssh`（または `cron`）を実行し、出力の各行（`Loaded`、`Active`、`Main PID`、末尾のログ）が何を表すかを調べる。`systemctl cat ssh` で、そのユニットファイルを読む。`[Unit]`、`[Service]`、`[Install]` のそれぞれに、何が書かれているか
2. `systemctl list-units --type=service` と `systemctl list-unit-files --type=service` の違いは何か。`enabled`、`disabled`、`static` は、それぞれどういう状態か（`man systemctl`）
3. APIを動かす専用のユーザーを作る。`man useradd` で、ログインしない「システムユーザー」を作るオプションを探す。バイナリをどこに置くかを決め、持ち主とパーミッションを決める。その場所と権限にした理由を `notes.md` に書く
4. 自分で作るユニットファイルは、どのディレクトリに置くのか。`man systemd.unit` の、ユニットファイルを探す場所の一覧で確かめる。パッケージが置く場所と、管理者が置く場所は、なぜ分かれているか
5. 最小のユニットファイルを書き、APIを起動する。まずは「起動するコマンド」の指定だけでよい。`systemctl start` のあと、`systemctl status` と `curl` で確かめる。ユニットファイルを書き換えたのに動きが変わらないときは、何を忘れているか
6. この時点で、APIはどのユーザーで動いているか。`ps -o user=,pid=,cmd= -p <Main PID>` で確かめる。`man systemd.exec` で、動かすユーザーを指定するディレクティブを探し、課題3のユーザーで動くようにする
7. DBの接続先やトークンを、ユニットファイルに直接書かずに渡す。`man systemd.exec` で、環境変数を「ファイルから読み込む」ディレクティブを探す。そのファイルは、誰が読める権限にするべきか。このファイルはコミットしない（写しを残すときは、値をダミーにする）
8. `journalctl -u <ユニット名>` でログを読む。次のそれぞれのオプションを `man journalctl` で探して試す：流れてくるログを見続ける、直近の50行だけ、今日の分だけ、今回の起動以降だけ、JSON形式で出す。G14で `log/slog` が出した構造化ログは、journal の中でどう見えるか
9. `systemctl stop` でAPIを止める。このとき、APIにはどのシグナルが届くか。G16で書いた終了処理のログは、`journalctl` に出ているか。`curl` で時間のかかるリクエストを投げている最中に止めた場合は、どうなるか
10. **本題（その1）**：`systemctl enable` で自動起動を有効にし、マシンを再起動する（WSL2はPowerShellで `wsl --shutdown`、Limaは `limactl stop` と `limactl start`、Linuxは `sudo reboot`）。入り直したあと、何も操作せずに `systemctl is-active`、`systemctl is-enabled`、`curl` で確かめる。`pg-offgrid` のコンテナは動いているか。動いていない場合、APIはどうなっているか。コンテナも自動で起動するようにする方法を、保存しておいたDockerのドキュメントで調べる
11. **本題（その2）**：APIのプロセスを `sudo kill -9 <Main PID>` で落とす。何も設定していないと、どうなるか。`man systemd.service` で、異常終了したら再起動するディレクティブと、再起動までの待ち時間のディレクティブを探して設定する。もう一度落とし、`Main PID` が変わること、`journalctl` に再起動の記録が残ることを確かめる。`systemctl stop` で止めたときにも、再起動されるか。なぜか
12. `[Unit]` セクションで、「ネットワークが使えるようになってから起動する」という順序を指定する方法を `man systemd.unit` で探す。順序の指定（`After=`）と、依存の指定（`Wants=`、`Requires=`）は、何が違うか

### 発展（任意）

- `systemd-analyze verify <ユニットファイル>` で、書き間違いを検査する。わざと綴りを間違えたディレクティブを入れると、何が出るか
- `systemd-analyze security <ユニット名>` の出力を読む。`man systemd.exec` から、サービスが触れる範囲を狭めるディレクティブを2つ選んで設定し、点数がどう変わるかを見る
- 起動に失敗し続けるバイナリ（わざと存在しない設定を渡す）で、再起動が何回で打ち切られるかを観察する。打ち切りの条件を決めているディレクティブを `man systemd.unit` で探す

## 詰まりやすいところ

- `systemctl start` が失敗し、`status` に `code=exited, status=203/EXEC` のような表示が出る → `journalctl -u` の直近のログを読む。バイナリのパス、実行権限、動かすユーザーからそのパスが読めるかを、1つずつ確かめる
- 手で実行すると動くのに、サービスにするとDBにつながらない → 手元のシェルとサービスとで、環境変数と作業ディレクトリが同じかを疑う
- ユニットファイルを直したのに、古い設定で動く → `systemctl status` の出力に、警告の1行が出ていないか
- 再起動後にAPIが `failed` になっている → `journalctl -b -u <ユニット名>` で、起動時に何が起きたかを時刻順に読む。DBのコンテナが立ち上がる時刻と比べる
- `journalctl` に何も出ない、または他人のログが読めない → 自分がどのグループに入っているかと、`man journalctl` の権限の説明を確かめる
- 環境変数のファイルに書いた値が、引用符ごと渡ってしまう → そのファイルの書式を `man systemd.exec` で確かめる。シェルの書式と同じとは限らない

## 完了条件の確かめ方

マシンを再起動し、入り直したあと、次の4つを順に実行して、結果を `notes.md` に貼る。

- `systemctl is-enabled <ユニット名>` が `enabled`、`systemctl is-active <ユニット名>` が `active`
- `curl` でAPIの一覧取得が成功する（DBの中身が返る）
- `ps -o user= -p <Main PID>` が `root` ではない
- `sudo kill -9 <Main PID>` のあと、数秒待つと `is-active` が `active` に戻り、`Main PID` が変わっている

## 残すもの

- ユニットファイルの写し（環境変数のファイルは、値をダミーにした `env.example` だけ）
- 完了条件の4つの確認結果と、各課題で観察したこと（`notes.md`）
- 設置から起動までの手順を、自分の言葉で箇条書きにしたもの（`install-steps.md`）。O6で、手順をコードにするときの材料になる

コマンドを試す場所は `~/sandbox/o2`。メモは、リポジトリの中のこのディレクトリに、AI機能を切ったエディタで書く。この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- O3（調査ツール）：ここでサービスにしたAPIを、調査の対象にする
- O6（構成管理入門）：`install-steps.md` のような手作業の手順を、Ansibleのコードに置き換える考え方を学ぶ
- C6（配布の形にする）：卒業制作の agent を、同じ方法で systemd のサービスにする
