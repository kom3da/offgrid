---
title: "O：運用（Stage 4）"
---

Stage 0の知識と、GoとDBを結びつけ、「ネットワークがなくても開発・デプロイできる」状態を作るトラック。Stage 0から使っているUbuntuの環境（WSL2、Lima、Linux）で、そのまま進められる。

## このトラックで身につくこと

- 初めて触るサーバーの状態を、自分のスクリプトで一通り確かめられる
- アプリをsystemdで常駐させ、ログを追える
- 静的バイナリと軽量なコンテナを作り、`docker compose up` だけで全体を動かせる
- 同じ構成を、コードから何度でも作り直せる
- `ss`、`strace`、`pprof` で、障害や遅さの原因を特定できる
- リポジトリだけから、別の環境で同じ成果物を作れる
- 他人（AI）が書いたサービスを引き継いで、読み、問題を見つけて直し、動かせる

## ユニット

各ユニットのページが、課題シート（答えのない問題集）になっている。下に添えた1行は、そのユニットで何ができるようになるかの**代表的な目安**。実際に満たすべき条件は、課題シートの「完了条件の確かめ方」が正本で、もっと細かい（テストの終了コードや、数値の基準まで含む）。

- **[O1 Linuxの管理](../drills/infra/O01-linux-admin/TASKS.md)**：OS・ディスク・メモリ・ユーザー・ポート・サービス・ログの場所を集めるスクリプトを書ける
- **[O2 systemd](../drills/infra/O02-systemd/TASKS.md)**：G12〜G17で作ったAPIが、再起動後も自動で起動する
- **[O3 調査ツール](../drills/infra/O03-investigation/TASKS.md)**：AIに仕込ませた障害の原因を、3問中2問特定できる
- **[O4 Dockerfile](../drills/infra/O04-dockerfile/TASKS.md)**：イメージのサイズが30MB以下
- **[O5 docker compose](../drills/infra/O05-compose/TASKS.md)**：`docker compose up` だけで、API・PostgreSQL・リバースプロキシが動く
- **[O6 構成管理入門](../drills/infra/O06-ansible/TASKS.md)**：Ansibleで、同じ手順で何度でも作り直せる
- **[O7 再現できるビルド](../drills/infra/O07-reproducible-build/TASKS.md)**：`build.sh` をまっさらな環境で実行して、手元と同じ `checksums.txt` になる
- **[O8 引き継ぎ](../drills/infra/O08-handover/TASKS.md)**：AIが書いて問題を仕込んだサービスを、答えを見ずに読み、仕込まれた問題の6割以上を見つけて直し、機能を1つ足して動かす
- **[O9 負荷試験入門](../drills/infra/O09-load-test/TASKS.md)**：改善前後のp95（遅いほうから5%目の応答時間）を記録できる

## 進め方

- 課題は `drills/infra/<ユニット>/` に置く
- O3の障害は、自分で仕込むと答えを知ってしまう。平日に、AIに仕込ませる（[デバッグドリルとコードリーディング](05-debug-and-reading.md)）

## 再現できるビルド（O7で満たすこと）

- `go mod vendor` で依存を固め、`-mod=vendor` でビルドできる
- `package-lock.json` と npm のキャッシュで、`npm ci` が同じ結果になる
- ベースイメージをダイジェストで固定し、`docker save` / `docker load` で持ち運べる
- `build.sh` が、リポジトリだけを入力に、バイナリ・`dist/`・イメージの tar と `checksums.txt` を出す
- まっさらな環境で `build.sh` を実行して、手元と同じ `checksums.txt` になる

## 引き継ぎ（O8）

現場で渡されるのは、自分が書いたコードではない。セッションの間の日に、[prompts/handover.md](../prompts/handover.md) の文面でAIに小さなサービスを書かせ、問題を5〜8個仕込ませる（答えは `answers/` に置き、見ない）。AIなしデーに、それを読み、動かし、境界を壊して観察し、問題を見つけて直し、機能を足し、引き継ぎ文書を書く。デバッグドリルをシステムの大きさでやる回で、**このカリキュラムの動機を直接測る**。

## 使う教材

『［試して理解］Linuxのしくみ』、『新しいLinuxの教科書』、各ツールの公式ドキュメントと `man`。入手先は[教材とツール](20-resources.md)にある。
