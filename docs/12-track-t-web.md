---
title: "T：Webフロントエンド（Stage 2〜3）"
---

HTMLとJavaScriptの基礎から、TypeScript、Reactまでを順に進むトラック。Stage 2で管理画面を作れるようになり、Stage 3でGoと組み合わせたリアルタイム表示を扱う。フレームワークはReactとViteを前提にする。

## このトラックで身につくこと

- HTMLとCSSを手で書き、ブラウザの開発者ツールで確かめられる
- 素のJavaScriptで、画面の操作と非同期の通信を書ける
- TypeScriptの型で設計し、漏れをコンパイラに見つけさせられる
- ビルド環境（Vite、ESLint、Vitest）を、テンプレートに頼らず設定ファイルから組める
- ReactでAPIにつながる画面を作り、APIが止まっても壊れないようにできる
- サーバーから流れてくるデータ（SSE）を受けて、画面に表示できる

## ユニット

各ユニットのページが、課題シート（答えのない問題集）になっている。

### Stage 2：WebフロントエンドとTypeScript

- **[T1 HTMLとCSS](../drills/ts/T01-html-css/TASKS.md)**：開発者ツールで、要素とスタイルを確認できる
- **[T2 JavaScriptの基本](../drills/ts/T02-javascript-basics/TASKS.md)**：G2・G3と同じ問題がNode.jsで動き、Goとの違いを3点以上書ける
- **[T3 DOMとイベント](../drills/ts/T03-dom-events/TASKS.md)**：ライブラリなしのToDoリストが動き、わざと入れたエラーの場所をConsoleから特定できる
- **[T4 非同期処理](../drills/ts/T04-async-fetch/TASKS.md)**：エラー時に画面へメッセージを出せる。CORSエラーが起きる条件を説明できる
- **[T5 npmとTypeScript導入](../drills/ts/T05-typescript-setup/TASKS.md)**：T3をTypeScriptに書き換え、`--strict` で型エラーが出ない
- **[T6 型で設計する](../drills/ts/T06-type-design/TASKS.md)**：イベントの種類を型で表し、漏れをコンパイラに検出させられる
- **[T7 ビルド環境を手で組む](../drills/ts/T07-build-tooling/TASKS.md)**：テンプレート生成を使わずに起動し、テストが動く
- **[T8 型ガードと実行時検証](../drills/ts/T08-type-guards/TASKS.md)**：不正なデータを拒否することを、テストで確認している
- **[T9 Reactの基本](../drills/ts/T09-react-basics/TASKS.md)**：T3のToDoを、Reactで作り直せる
- **[T10 ReactとAPIの接続](../drills/ts/T10-react-api/TASKS.md)**：APIを止めても、画面が壊れない

### Stage 3：リアルタイム表示

- **[T11 SSEの受信](../drills/ts/T11-sse-client/TASKS.md)**：サーバーを止めて再開すると、表示が自動で再開する

## 進め方

- 課題は `drills/ts/<ユニット>/` に置く
- T4はG12のAPI、T10はG12〜G14のAPI、T11はG20のSSEを呼ぶ。[G：Go](11-track-g-go.md)の該当ユニットを、先に済ませておく
- ブラウザは、Ubuntuの外（手元のOS）のものを使う。Ubuntuの中で動かしたサーバーは、`http://localhost:ポート番号` で開ける

## 閉域を意識した決まり

- 依存は最小限にし、`package-lock.json` を必ずコミットする
- CDN（Google Fontsなど）に依存しない。フォントやグラフ用のライブラリは、ローカルに置く
- `npm ci --offline` で再現できるかを、Stage 4で確かめる

## 手書きで確認したい型パターン

- `type Event = { kind: "move"; ... } | { kind: "alert"; ... }` と、`switch` による網羅
- `Record<K, V>`、`Partial`、`Pick`、`Omit` の使い分け
- `as const` とリテラル型
- `unknown` から安全に絞り込む、型ガード関数
- `satisfies` 演算子

## Stage 5で使う応用

- リアルタイム表示：T11の方法で、メトリクスを受信して表示する
- グラフ表示：SVGかCanvasで、時系列グラフを描く。ライブラリを使う場合は、CDNではなくローカルに同梱する
- テスト：VitestとTesting Libraryで、主要なコンポーネントをテストする

## 使う教材

『JavaScript Primer』、『サバイバルTypeScript』、『プロを目指す人のためのTypeScript入門』。MDN、TypeScript Handbook、Reactの公式ドキュメントは、オフラインで読めるようにしておく。入手先と手順は[教材とツール](20-resources.md)にある。
