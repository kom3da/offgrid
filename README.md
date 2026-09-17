# offgrid

ネットワークもAIも切った状態で「書ける・読める・直せる」力を、ゼロから付け直すための学習カリキュラム。コンピュータの仕組み・Linux・Git・ネットワークから、Go・PostgreSQL・TypeScript、閉域での運用までを、週1回の「AIなしデー」で進める。最後に、閉域でも動く監視ダッシュボード「watchdeck」を自力で完成させる。

**ブラウザで読む：<https://kom3da.github.io/offgrid/>**（このリポジトリの `docs/`、課題シート、`prompts/` と同じ内容）

## 使い方

このリポジトリはテンプレート。ここに直接書き込まず、コピーして作った自分のリポジトリの `main` で学習を進める。

```sh
gh repo create offgrid-log --template kom3da/offgrid --private --clone
```

まっさらなマシンからの手順（Windows・macOS・Linux共通。学習はすべてUbuntuの中で行う）は、[はじめかた](docs/02-getting-started.md)にある。カリキュラムが改訂されたら、`./scripts/update-curriculum.sh <版>` で取り込む。

## 読む順番

1. [offgrid とは](docs/01-overview.md)：目的、ゴール像、前提
2. [はじめかた（Day 0）](docs/02-getting-started.md)：環境の準備、毎回の始め方と終わり方
3. [学習ロードマップ](docs/03-roadmap.md)：6つのステージと目安
4. [AIなしデーの進め方](docs/04-ai-free-day.md)：ルール、1日の時間割、平日のメニュー
5. [デバッグドリルとコードリーディング](docs/05-debug-and-reading.md)
6. [評価とステージ通過](docs/06-evaluation.md)：「説明できる」の確かめ方、確認項目、実技

## トラックとユニット

ユニットのIDの頭文字が、トラックを表す。各ユニットに、答えのない課題シート（`drills/<トラック>/<ユニット>/TASKS.md`）がある。

- [F：基礎](docs/10-track-f-foundation.md)（Stage 0、F1〜F10）
- [G：Go](docs/11-track-g-go.md)（Stage 1〜3、G1〜G20）
- [T：Webフロントエンド](docs/12-track-t-web.md)（Stage 2〜3、T1〜T11）
- [D：PostgreSQL](docs/13-track-d-postgresql.md)（Stage 1〜3、D1〜D15）
- [O：運用と閉域](docs/14-track-o-ops.md)（Stage 4、O1〜O9）
- [C：卒業制作 watchdeck](docs/15-track-c-capstone.md)（Stage 5、C1〜C8）

資料は[教材とツール](docs/20-resources.md)、AIに頼むときの文面は [prompts/](prompts/) にある。セッションを案内する学習用コマンド `offgrid` は、[はじめかた](docs/02-getting-started.md)で入れる。

## ディレクトリ構成

```text
.
├── docs/              # カリキュラム本体（サイトの各ページ）
├── drills/
│   ├── foundation/    # F：基礎          （F01-binary/TASKS.md など）
│   ├── go/            # G：Go            （G02-control-flow/ など）
│   ├── ts/            # T：Webフロントエンド（T03-dom-events/ など）
│   ├── sql/           # D：PostgreSQL    （D03-aggregate/ など）
│   ├── infra/         # O：運用と閉域      （O04-dockerfile/ など）
│   ├── capstone/      # C：卒業制作の各ステップの課題シート
│   ├── debug/         # デバッグドリルの問題（YYYY-MM-DD-<テーマ>/）
│   └── reading/       # コードリーディングのメモ
├── answers/           # デバッグドリルの答え（答え合わせまで開かない）
├── capstone/          # 卒業制作 watchdeck の実装
├── logs/              # AIなしデーの振り返り（YYYY/MM-DD.md）
├── PROGRESS.md        # 自分の進捗
├── templates/         # 振り返りテンプレート
├── prompts/           # AIに頼むときの文面
├── tools/offgrid/     # 学習用コマンド offgrid のソース（Go）
├── scripts/           # 補助スクリプト（ダミーのログ生成、カリキュラムの取り込み）
├── site/              # ドキュメントサイト（Astro Starlight）。テンプレート側でだけ使う
├── CLAUDE.md          # Claude Code 向けのコンテキストとレビュー方針
└── CHANGELOG.md       # 版ごとの変更履歴
```

**自分のもの**として自由に書いてよいのは、`PROGRESS.md`、`logs/`、`answers/`、`capstone/` と、`drills/` の中の `TASKS.md` 以外のファイル。それ以外はカリキュラムの一部で、取り込みのたびに置き換わる。

## 1回の回し方

1. エディタのAI機能をオフにする
2. 学習用コマンドを起動する（AIもネットワークも使わない）
   ```sh
   offgrid
   ```
3. メニューの「1. セッションを始める（案内つき）」を選ぶ。ウォームアップから終わりの手続きまで、課題を1問ずつ案内する
4. 中断しても続きから再開できる。進捗と作業漏れは `offgrid status` で分かる。ブラウザで使いたければ `offgrid serve`

ペースは週1回でも、週2〜3回でもよい（[AIなしデーの進め方](docs/04-ai-free-day.md)）。

## コミット規約

AIの関与度をあとから区別できるよう、コミットメッセージの先頭にタグを付ける。

- `[no-ai]`：AIなしで書いたもの（AIなしデー）
- `[ai]`：AIが生成した、または大きく補助したもの
- `[mixed]`：自分で書き、AIにレビューや部分的な修正をさせたもの
- `[docs]`：ドキュメントと振り返りのみ

例：`[no-ai] G11: csvstat の平均計算を実装`

## 版と改訂

版は、学習テキストとしての内容に付ける（カレンダー形式。`v2026.10`、同じ月の2回目以降は `v2026.10.1`）。学習の進み具合には、タグを付けない。変更点は [CHANGELOG.md](CHANGELOG.md)、運用は[改訂の運用](docs/30-revision.md)にある。誤りや、詰まりやすい箇所を見つけたら、Issueで知らせてほしい。

## 注意

- テンプレートは公開されている。自分の学習用リポジトリも、公開・非公開にかかわらず「いつか公開されるかもしれない」前提で書く。一度pushした内容は、消しても残る
- 題材は、公開情報とダミーデータのみ。業務上の機密、実在システムの非公開情報、実在のホスト名・IPアドレス・認証情報は置かない。`logs/` の振り返りにも、仕事の内容は書かない
- 書籍の問題文や本文は転載しない（自分の解答と、出典のページ番号だけを残す）
