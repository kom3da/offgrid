---
title: "O：運用と閉域（Stage 4）"
---

Stage 0の知識と、GoとDBを結びつけ、「ネットワークがなくても開発・デプロイできる」状態を作るトラック。Stage 0から使っているUbuntuの環境（WSL2、Lima、Linux）で、そのまま進められる。

## このトラックで身につくこと

- 初めて触るサーバーの状態を、自分のスクリプトで一通り確かめられる
- アプリをsystemdで常駐させ、ログを追える
- 静的バイナリと軽量なコンテナを作り、`docker compose up` だけで全体を動かせる
- 同じ構成を、コードから何度でも作り直せる
- `ss`、`strace`、`pprof` で、障害や遅さの原因を特定できる
- ネットワークを切った状態で、ビルドから起動までをやり切れる

## ユニット

各ユニットのページが、課題シート（答えのない問題集）になっている。

- **[O1 Linuxの管理](../drills/infra/O01-linux-admin/TASKS.md)**：OS・ディスク・メモリ・ユーザー・ポート・サービス・ログの場所を集めるスクリプトを書ける
- **[O2 systemd](../drills/infra/O02-systemd/TASKS.md)**：G12〜G17で作ったAPIが、再起動後も自動で起動する
- **[O3 調査ツール](../drills/infra/O03-investigation/TASKS.md)**：AIに仕込ませた障害の原因を、3問中2問特定できる
- **[O4 Dockerfile](../drills/infra/O04-dockerfile/TASKS.md)**：イメージのサイズが30MB以下
- **[O5 docker compose](../drills/infra/O05-compose/TASKS.md)**：`docker compose up` だけで、API・PostgreSQL・リバースプロキシが動く
- **[O6 構成管理入門](../drills/infra/O06-ansible/TASKS.md)**：Ansibleで、同じ手順で何度でも作り直せる
- **[O7 閉域の準備](../drills/infra/O07-offline-prep/TASKS.md)**：下の「閉域環境の準備」を、すべて満たす
- **[O8 閉域シミュレーション](../drills/infra/O08-offline-day/TASKS.md)**：ネットワークを切ったまま、AIなしデー1回分の作業を完了できる
- **[O9 負荷試験入門](../drills/infra/O09-load-test/TASKS.md)**：改善前後のp95（遅いほうから5%目の応答時間）を記録できる

## 進め方

- 課題は `drills/infra/<ユニット>/` に置く
- 1日ぶんの作業をネットワークなしで通すのは、O8（閉域シミュレーション）とC7の回だけ。ふだんのセッションは、つないだままでよい（[AIなしデーの進め方](04-ai-free-day.md)のルール）。O7などで、確かめるために短時間切ることは、課題の中にある
- オフラインのドキュメント一式を揃えるのは、O7の課題そのもの。先回りして用意しておく必要はない
- O3の障害は、自分で仕込むと答えを知ってしまう。平日に、AIに仕込ませる（[デバッグドリルとコードリーディング](05-debug-and-reading.md)）

## 閉域環境の準備（O7で満たすこと）

- `go mod vendor` で依存を固め、`-mod=vendor` でビルドできる
- npmのオフラインキャッシュを用意し、`npm ci --offline` で再現できる
- Dockerイメージを `docker save` で保存し、`docker load` で復元できる
- Goの公式ドキュメントを、`pkgsite` でローカルに表示できる
- PostgreSQL・MDN・TypeScriptのドキュメントの、オフラインコピーがある
- ローカルLLM（Ollamaなど）とモデルを事前に取得し、ネットワークを切った状態で推論できる
- ここまでの手順そのものを、シェルスクリプトかAnsibleでコード化している

発展（任意）：Goモジュールのローカルプロキシ（`GOPROXY=file://...` またはAthens）と、npmのローカルレジストリ（Verdaccio）を立てる。複数人や複数台で依存を共有するときに、必要になる。

## 閉域AIデー（任意・O8以降）

AIなしデーとは別に、任意で「閉域AIデー」を設ける（O8以降、10回に1回ほどが目安）。ネットワークを切り、ローカルLLMだけでコード補完とレビューを回し、クラウドのAIとの差（精度、速度、扱える文脈の長さ）を記録する。外部にコードを送れない現場でもAIを活用して開発できる、という実績になる。

## 使う教材

『［試して理解］Linuxのしくみ』、『新しいLinuxの教科書』、各ツールの公式ドキュメント（オフラインに保存したもの）と `man`。入手先は[教材とツール](20-resources.md)にある。
