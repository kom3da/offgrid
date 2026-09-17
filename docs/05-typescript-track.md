# TypeScriptトラック

Stage 2で、HTMLとJavaScriptの基礎からTypeScript、Reactまで順に進む。Stage 3のT11で、Goと組み合わせたリアルタイム表示を扱う。フレームワークはReact＋Viteを前提にし、課題は `drills/ts/<ユニット>-<名前>/` に置く。T4はG12のAPI、T10はG12〜G14のAPIを呼ぶので、[Goトラック](04-go-track.md)の該当ユニットを先に済ませておく。

## Stage 2：WebフロントエンドとTypeScript（T1〜T10）

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| T1 | HTMLとCSS | 自己紹介ページを手書きし、見出し・リスト・表・フォームを使う。Flexboxで2カラムにする | ブラウザの開発者ツールで要素とスタイルを確認できる |
| T2 | JavaScriptの基本 | Node.jsで変数、関数、配列、オブジェクト、ループを練習（G2・G3と同じ問題を解く） | 同じ問題がNode.jsで動き、GoとJavaScriptの違いを3点以上 `explain.md` に書ける |
| T3 | DOMとイベント | 素のJavaScriptでToDoリストを作る（追加・完了・削除）。開発者ツールのConsole（エラーの表示）とブレークポイントで動きを追う | ライブラリなしで動く。わざと入れたエラーの場所をConsoleから特定できる |
| T4 | 非同期処理 | コールバック、Promise、`async`/`await`、`fetch`。G12のAPIを呼ぶ。HTMLはG12のGoサーバーから配信する（同一オリジン＝同じアドレスとポート）。別のポートから開くと、ブラウザが通信を止める（CORSエラー）ことを、開発者ツールのNetworkタブで確認する | エラー時に画面へメッセージを出せる。CORSエラーが起きる条件を説明できる |
| T5 | npmとTypeScript導入 | `npm init`、`tsconfig.json`、基本の型、`tsc --strict` | T3をTypeScriptに書き換え、型エラーが出ない |
| T6 | 型で設計する | ユニオン型、判別可能ユニオン、`never`による網羅チェック、ジェネリクス | イベントの種類を型で表し、漏れをコンパイラに検出させられる |
| T7 | ビルド環境を手で組む | Vite、ESLint、Vitestを設定ファイルから組む。開発中のAPI呼び出しは、Viteのproxy設定かG14のCORSミドルウェアで通す | テンプレート生成を使わずに起動し、テストが動く |
| T8 | 型ガードと実行時検証 | `unknown`から安全に絞り込む関数を手書きし、APIレスポンスを検証する | 不正なデータを拒否することを、Vitestのテストで確認している |
| T9 | Reactの基本 | コンポーネント、props、state、リスト表示、フォーム、`useEffect`とクリーンアップ（後片付け） | T3のToDoをReactで作り直せる |
| T10 | ReactとAPIの接続 | G12〜G14のAPIに接続し、一覧・詳細・登録画面を作る。ローディング、エラー、再試行も実装 | APIを止めても画面が壊れない |

## Stage 3：リアルタイム表示（T11）

| ユニット | テーマ | 課題 | 完了条件 |
| --- | --- | --- | --- |
| T11 | SSEの受信 | `EventSource`でG20のSSEを受け、届いた値をReactで表示する。切断時の再接続と、`useEffect`でのクリーンアップ | サーバーを止めて再開すると、表示が自動で再開する |

## Stage 5で使う応用

| テーマ | 内容 |
| --- | --- |
| リアルタイム表示 | T11の方法で、メトリクスを受信して表示 |
| グラフ表示 | SVGかCanvasで時系列グラフを描く。ライブラリを使う場合は、CDNではなくローカルに同梱する |
| テスト | Vitest＋Testing Libraryで主要コンポーネントをテスト |

## 手書きで確認したい型パターン

- `type Event = { kind: "move"; ... } | { kind: "alert"; ... }`と`switch`による網羅
- `Record<K, V>`、`Partial`、`Pick`、`Omit`の使い分け
- `as const`とリテラル型
- `unknown`から安全に絞り込む型ガード関数
- `satisfies`演算子

## 閉域を意識した注意点

- 依存は最小限にし、`package-lock.json`を必ずコミットする
- CDN（Google Fontsなど）に依存しない。フォントやグラフ用ライブラリはローカルに置く
- `npm ci --offline`で再現できるかを、Stage 4で確認する
