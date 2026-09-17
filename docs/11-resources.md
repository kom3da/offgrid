# 教材・リソース

紙の書籍かPDFを中心に揃え、オンライン教材は事前にオフラインで使えるようにしておく。書名と入手方法は購入前に最新版を確認すること（記憶に基づく一覧）。

## 書籍

| 分野 | 書名 | 使うユニット |
| --- | --- | --- |
| コンピュータ | 『コンピュータはなぜ動くのか』（矢沢久雄） | F1 |
| コンピュータ | 『プログラムはなぜ動くのか』（矢沢久雄） | F1、G1 |
| Linux | 『新しいLinuxの教科書』（三宅英明、大角祐介） | F2〜F5 |
| Linux | 『［試して理解］Linuxのしくみ』（武内覚） | F4、O1〜O3 |
| Git | 『Pro Git』（無料、日本語版あり） | F6、F7 |
| ネットワーク | 『マスタリングTCP/IP 入門編』 | F8 |
| HTTP | 『Webを支える技術』（山本陽平） | F9、G12 |
| Go | 『初めてのGo言語』（Jon Bodner、『Learning Go』の邦訳） | G1〜G11 |
| Go | 『プログラミング言語Go』（Donovan、Kernighan）。2016年刊でモジュールやジェネリクス以前の内容なので、環境構築は公式ドキュメントで補う | G3〜G10の練習問題 |
| Go | 『実用 Go言語』（渋川よしき ほか） | G12〜G18 |
| Go | 『Go言語による並行処理』（Katherine Cox-Buday） | G15、G16 |
| アルゴリズム | 『なっとく！アルゴリズム』（Aditya Bhargava） | G9、G10 |
| JavaScript | 『JavaScript Primer』（Web版は無料） | T2〜T4 |
| TypeScript | 『サバイバルTypeScript』（Web版は無料） | T5〜T7 |
| TypeScript | 『プロを目指す人のためのTypeScript入門』（鈴木僚太） | T5〜T8 |
| SQL | 『SQL ゼロからはじめるデータベース操作』（ミック） | D1〜D5 |
| DB設計 | 『達人に学ぶDB設計徹底指南書』（ミック） | D6、D7 |
| SQL | 『達人に学ぶSQL徹底指南書』（ミック） | D10〜D14 |
| PostgreSQL | 『内部構造から学ぶPostgreSQL 設計・運用計画の鉄則』 | D14 |
| 監視 | 『入門 監視』（Mike Julian） | C1、C7 |
| SRE | 『Site Reliability Engineering』（Google、Web版は無料・英語） | C8 |

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
- [ ] 壊してもよい練習用の環境（`~/sandbox`、またはWSL2や仮想マシン）を用意する
- [ ] エディタのAI機能をワンクリックで切り替えられるプロファイルを作る
- [ ] 最初のAIなしデーの日付を決め、カレンダーに入れる
