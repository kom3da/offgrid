# offgrid

コンピュータの仕組み・Linux・Git・ネットワークから、Go・TypeScript・PostgreSQLまでをゼロから学び直すための個人学習リポジトリ。ネットワークもAIも切った状態で「書ける・読める・直せる」力をつけ、閉域環境でも開発を回せるようになることを目指す。

期限は設けず、6つのステージをチェックリストで区切って進む。最後に、閉域でも動く監視ダッシュボード「watchdeck」（Go＋PostgreSQL＋TypeScript/React）を自力で完成させる。

## 最初に読む

| ドキュメント | 内容 |
| --- | --- |
| [docs/00-overview.md](docs/00-overview.md) | 目的、ゴール像、前提 |
| [docs/01-roadmap.md](docs/01-roadmap.md) | 6ステージの構成と配分 |
| [docs/02-ai-free-day.md](docs/02-ai-free-day.md) | 週1回の「AIなしデー」のルールと1日のテンプレート |
| [docs/03-foundation.md](docs/03-foundation.md) | Stage 0：コンピュータ・Linux・Git・ネットワーク・HTTP（F1〜F10） |
| [docs/04-go-track.md](docs/04-go-track.md) | Go（G1〜G18） |
| [docs/05-typescript-track.md](docs/05-typescript-track.md) | HTML・JavaScript・TypeScript・React（T1〜T10） |
| [docs/06-postgresql-track.md](docs/06-postgresql-track.md) | PostgreSQL（D1〜D15） |
| [docs/07-ops-offline-track.md](docs/07-ops-offline-track.md) | Stage 4：運用と閉域（O1〜O8） |
| [docs/08-capstone-watchdeck.md](docs/08-capstone-watchdeck.md) | Stage 5：卒業制作 watchdeck（C1〜C8） |
| [docs/09-debug-and-reading.md](docs/09-debug-and-reading.md) | デバッグドリルとコードリーディング |
| [docs/10-evaluation.md](docs/10-evaluation.md) | 到達レベルとステージ別チェックリスト |
| [docs/11-resources.md](docs/11-resources.md) | 書籍・教材・ツール、オフライン化の方法 |
| [docs/12-revision.md](docs/12-revision.md) | 改訂の運用（版の付け方、チェックリスト） |
| [PROGRESS.md](PROGRESS.md) | 現在のステージと通過記録 |
| [CHANGELOG.md](CHANGELOG.md) | 版ごとの変更履歴 |

## ステージとユニット

| ステージ | ユニット | 目安（AIなしデー） |
| --- | --- | --- |
| Stage 0：基礎 | F1〜F10 | 8〜12回 |
| Stage 1：Go入門 | G1〜G11、D1〜D5 | 10〜15回 |
| Stage 2：Webの基礎 | T1〜T10、G12〜G14、D6〜D12 | 12〜18回 |
| Stage 3：並行処理とDB連携 | G15〜G18、D13〜D15 | 8〜12回 |
| Stage 4：運用と閉域 | O1〜O8 | 8〜12回 |
| Stage 5：卒業制作 | C1〜C8 | 12〜20回 |

## ディレクトリ構成

```text
.
├── CLAUDE.md          # Claude Code 向けのコンテキストとレビュー方針
├── PROGRESS.md        # 進捗
├── CHANGELOG.md       # 改訂履歴
├── docs/              # カリキュラム本体
├── drills/
│   ├── foundation/    # Stage 0（F01-binary/ など）
│   ├── go/            # Go（G02-fizzbuzz/ など）
│   ├── ts/            # HTML・JS・TS（T03-todo/ など）
│   ├── sql/           # PostgreSQL（D03-aggregate/ など）
│   ├── infra/         # 運用・閉域（O04-dockerfile/ など）
│   ├── debug/         # デバッグドリルの問題（YYYY-MM-DD-<テーマ>/）
│   └── reading/       # コードリーディングのメモ
├── answers/           # デバッグドリルの答え（答え合わせまで開かない）
├── capstone/          # 卒業制作 watchdeck
├── logs/              # AIなしデーの振り返り（YYYY/MM-DD.md）
├── templates/         # 振り返りテンプレート
├── prompts/           # デバッグドリル生成・改訂用プロンプト
└── scripts/           # 補助スクリプト（GitHub への初回 push）
```

ユニットのディレクトリ名は `<ユニットID 2桁>-<名前>`（例：`F03-pipeline`、`G11-csvstat`）にそろえる。

## 1回の回し方

1. `PROGRESS.md` で次のユニットを確認する
2. 振り返りテンプレートをコピーする
   ```sh
   mkdir -p logs/$(date +%Y) && cp templates/retrospective.md logs/$(date +%Y)/$(date +%m-%d).md
   ```
3. エディタのAI機能をオフにして、テンプレートの時間割どおりに進める
4. 夕方に振り返りを書き、`PROGRESS.md` を更新してコミットする

## コミット規約

AIの関与度をあとから区別できるよう、コミットメッセージの先頭にタグを付ける。

| タグ | 意味 |
| --- | --- |
| `[no-ai]` | AIなしで書いたもの（AIなしデー） |
| `[ai]` | AIが生成・大きく補助したもの |
| `[mixed]` | 自分で書き、AIにレビューや部分修正をさせた |
| `[docs]` | ドキュメント・振り返りのみ |

例：`[no-ai] G11: csvstat の平均計算を実装`

AIなしデーの終わりに `git tag noai-YYYY-MM-DD` を付ける。

## 版と改訂

版はカレンダー形式（`v2026.09`）。改訂の手順は [docs/12-revision.md](docs/12-revision.md) を参照。

## 注意

- 題材は公開情報とダミーデータのみ。業務上の機密や実在システムの非公開情報は置かない
- 非公開（private）リポジトリとして運用する
