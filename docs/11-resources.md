# 教材・リソース

紙の書籍かPDFを中心に揃え、オンライン教材は事前にオフラインで使えるようにしておく。書名と入手方法は購入前に最新版を確認すること（記憶に基づく一覧）。

書籍は先にまとめて買わない。最初に買うのは2冊だけにし、残りはそのステージの開始時に、その時点の最新版を確認して買う。各ユニットで「読む章」は、ステージ開始時の小改訂（[改訂の運用](12-revision.md)）で、目次を検索で確認して追記する。読む時間は、平日の火曜と木曜（[平日の補助メニュー](02-ai-free-day.md)）。

## 書籍

| 分野 | 書名 | 使うユニット | 入手の時期 |
| --- | --- | --- | --- |
| コンピュータ | 『コンピュータはなぜ動くのか』（矢沢久雄） | F1 | 最初に買う |
| コンピュータ | 『プログラムはなぜ動くのか』（矢沢久雄） | F1、G1 | 使うユニットに入る前 |
| Linux | 『新しいLinuxの教科書』（三宅英明、大角祐介） | F2〜F5 | 最初に買う |
| Linux | 『［試して理解］Linuxのしくみ』（武内覚） | F4、O1〜O3 | 使うユニットに入る前 |
| Git | 『Pro Git』（無料、日本語版あり） | F6、F7 | 最初に入手（無料） |
| ネットワーク | 『マスタリングTCP/IP 入門編』 | F8 | 使うユニットに入る前 |
| HTTP | 『Webを支える技術』（山本陽平） | F9、G12 | 使うユニットに入る前 |
| Go | 『初めてのGo言語』（Jon Bodner、『Learning Go』の邦訳） | G1〜G11 | Stage 1の開始時 |
| Go | 『プログラミング言語Go』（Donovan、Kernighan）。2016年刊でモジュールやジェネリクス以前の内容なので、環境構築は公式ドキュメントで補う | G3〜G10の練習問題 | Stage 1の開始時 |
| Go | 『実用 Go言語』（渋川よしき ほか） | G12〜G18 | Stage 2の開始時 |
| Go | 『Go言語による並行処理』（Katherine Cox-Buday） | G15、G16 | Stage 3の開始時 |
| アルゴリズム | 『なっとく！アルゴリズム』（Aditya Bhargava） | G9、G10 | Stage 1の開始時 |
| JavaScript | 『JavaScript Primer』（Web版は無料） | T2〜T4 | Stage 2の開始時 |
| TypeScript | 『サバイバルTypeScript』（Web版は無料） | T5〜T7 | Stage 2の開始時 |
| TypeScript | 『プロを目指す人のためのTypeScript入門』（鈴木僚太） | T5〜T8 | Stage 2の開始時 |
| SQL | 『SQL ゼロからはじめるデータベース操作』（ミック） | D1〜D5 | Stage 1の開始時 |
| DB設計 | 『達人に学ぶDB設計徹底指南書』（ミック） | D6、D7 | Stage 2の開始時 |
| SQL | 『達人に学ぶSQL徹底指南書』（ミック） | D10〜D14 | Stage 2の開始時 |
| PostgreSQL | 『内部構造から学ぶPostgreSQL 設計・運用計画の鉄則』 | D14 | Stage 3の開始時 |
| 監視 | 『入門 監視』（Mike Julian） | C1、C7 | Stage 5の開始時 |
| SRE | 『Site Reliability Engineering』（Google、Web版は無料・英語） | C8 | Stage 5の開始時 |

## オンライン教材（オフライン化の方法）

| 教材 | オフライン化 |
| --- | --- |
| Linux標準教科書（LPI-Japan） | 公式サイトから無料PDFをダウンロード |
| Pro Git | 公式サイトからPDFまたはEPUBをダウンロード |
| MDN Web Docs（HTML・CSS・JavaScript） | DevDocsのオフライン機能 |
| JavaScript Primer | GitHubのリポジトリをcloneして手元で閲覧 |
| A Tour of Go | `go install golang.org/x/website/tour@latest`で手元に起動 |
| Go標準ライブラリのドキュメント | `go install golang.org/x/pkgsite/cmd/pkgsite@latest`でローカル表示 |
| Effective Go、Go言語仕様 | ブラウザでPDF保存、またはDevDocsのオフライン機能 |
| TypeScript Handbook、React公式ドキュメント | DevDocsのオフライン機能 |
| PostgreSQL公式マニュアル | 公式サイトからPDFをダウンロード |

## ステージ開始時のオフライン教材セットアップ

AIなしデーでは、初回からオフラインのドキュメントを使う。各ステージに入る前の平日（AIを使ってよい日）に、次を済ませる。

| ステージ | 用意するもの |
| --- | --- |
| Stage 0 | Linux標準教科書とPro GitのPDF、DevDocsでBash・Git・HTTPをオフライン化、`man`が引けることの確認 |
| Stage 1 | A Tour of Goと`pkgsite`のローカル起動、PostgreSQLマニュアルのPDF、サンプルDBのpagila |
| Stage 2 | DevDocsでMDN（HTML・CSS・JavaScript）・TypeScript・Reactをオフライン化、JavaScript Primerのclone |
| Stage 3 | `pgx`のソースとドキュメントのclone、コードリーディング対象のOSSのclone |
| Stage 4 | [閉域環境の準備チェックリスト](07-ops-offline-track.md)（O7）で、ここまでの手順をコード化する |

DevDocsのオフライン保存はブラウザの中に置かれるので、ブラウザのデータを消すと失われる。消えたら取り直す。

## ツール

| 用途 | ツール |
| --- | --- |
| シェルスクリプトの検査 | `shellcheck` |
| Go静的解析 | `go vet`、`staticcheck`、`golangci-lint` |
| デバッガ | `dlv`（Delve） |
| プロファイル | `pprof`、`go tool trace` |
| 負荷試験 | `k6` または `vegeta` |
| ローカルレジストリ | Athens（Go）、Verdaccio（npm） |
| ローカルLLM | Ollama、llama.cpp |
| グラフ | uPlotやChart.jsなど（ローカルに同梱） |

## 最初にやる準備

- [ ] このリポジトリをcloneし、READMEを読む
- [ ] 『コンピュータはなぜ動くのか』と『新しいLinuxの教科書』を入手する
- [ ] Stage 0のオフライン教材（上の表）を用意する
- [ ] `shellcheck` をインストールする
- [ ] 壊してもよい練習用の環境（`~/sandbox`、またはWSL2や仮想マシン）を用意する
- [ ] エディタのAI機能をワンクリックで切り替えられるプロファイルを作る
- [ ] 最初のAIなしデーの日付を決め、カレンダーに入れる
