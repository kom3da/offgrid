---
title: "T9 Reactの基本"
---

> トラック：[T：Webフロントエンド](../../../docs/12-track-t-web.md) ／ Stage 2 ／ 目安：メイン枠で3回 ／ 前提：T7、T3（T5でTypeScriptにしたToDoを使う）

## ねらい

T3で手で書いたDOM操作を、Reactに肩代わりさせる。Reactは「データが変わったら、画面を描き直す」というT3の課題7の考え方を、仕組みとして持っている。ここでは、コンポーネント、props、state、リスト、フォーム、`useEffect` の6つを、1つずつ実験してから、T3のToDoを作り直す。閉域では、Reactのバージョンを気軽に上げたり、便利なライブラリを足したりできない。Reactそのものだけで画面が作れることと、公式ドキュメントをオフラインで引けることを、ここで確かめる。

## このユニットで身につけること

- テンプレート生成を使わず、T7の環境にReactを足して、コンポーネントを1つ表示する
- propsに型を付けたコンポーネントを書き、親から子へ値を渡す
- `useState` で状態を持ち、状態が変わると画面が描き直されることを確かめる
- 配列を `key` 付きで一覧表示し、`key` の付け方による違いを観察する
- 制御されたフォーム（入力欄の値をstateで持つ形）を書く
- `useEffect` がいつ実行されるかを、依存配列とStrictModeの有無で観察する
- `useEffect` のクリーンアップで、タイマーやイベントリスナーを後片付けする
- 状態を変える処理を純粋な関数に切り出し、Vitestでテストする

## キーワード

- **コンポーネント**：画面の一部を返す関数。小さな部品を組み合わせて画面を作る
- **JSX**：JavaScriptの中にHTMLに似た書き方で要素を書く構文。`.tsx` ファイルに書き、ビルドで関数呼び出しに変換される
- **props**：親のコンポーネントから子へ渡す値。関数の引数に当たる
- **state**：コンポーネントが自分で持ち、変わると描き直しが起きる値。`useState` で作る
- **描き直し（再レンダー）**：stateやpropsが変わったとき、Reactがコンポーネントの関数をもう一度呼び、画面との差分を反映すること
- **`key`**：一覧の要素1つずつに付ける目印。Reactが「どれがどれか」を追うために使う
- **制御されたコンポーネント**：入力欄の値をstateで持ち、`onChange` でstateを更新する書き方
- **`useEffect`**：描画のあとに実行したい処理（外部との同期）を登録する仕組み
- **クリーンアップ**：`useEffect` に登録した処理の後片付け。コンポーネントが消えるときと、次の実行の前に呼ばれる
- **StrictMode**：開発中だけ、問題を見つけやすくするために、Reactが一部の処理をわざと2回実行するモード

## 平日に読むもの

- React公式ドキュメント（DevDocsのオフライン版）の「Learn」から：Your First Component、Passing Props to a Component、Rendering Lists、State: A Component's Memory、State as a Snapshot、Updating Arrays in State、Sharing State Between Components、Synchronizing with Effects、You Might Not Need an Effect
- 同じく「Learn」のUsing TypeScript（型定義パッケージと、propsやイベントの型の書き方）
- 同じく「Reference」の `useState`、`useEffect`、`createRoot`、`StrictMode`
- Viteのドキュメントの「Features」にあるJSXの項と、TSConfigのリファレンスの `jsx` オプション
- パッケージの取得にはネットワークが要る。回線が使えない場合は、準備の `npm install` までを平日に済ませておく

## 準備

T7で自分が書いた設定ファイルをコピーし、Reactを足す。テンプレート生成は使わない。

```sh
cd ~/offgrid-log/drills/ts/T09-react-basics
cp ../T07-build-tooling/{package.json,package-lock.json,tsconfig.json,vite.config.ts,index.html} .
cp ../T07-build-tooling/<ESLintの設定ファイル> .
npm ci
npm install react react-dom
mkdir -p src && touch notes.md
```

TypeScriptでReactを書くには、型定義のパッケージも要る。何が要るかは、React公式ドキュメントのUsing TypeScriptで調べて入れる。JSXを変換する設定（`tsconfig.json` の `jsx` と、Viteの側で必要なもの）は、ドキュメントを読んで自分で決め、決めた内容と理由を `notes.md` に書く。Vite公式のReact用プラグインは、Reactを動かすための部品として入れてよい。`react` と `react-dom` を `dependencies` に入れるか `devDependencies` に入れるかも、T5の課題1に戻って決める。

## 課題

1. `index.html` に、Reactが描画先にする空の要素を1つ置き、`src/main.tsx` から `createRoot` で「こんにちは」と表示する（`createRoot` のリファレンスを読む）。`npm run dev` で起動し、ホスト側のブラウザで開く。Elementsパネルで、空だった要素の中に何が入ったかを見る。Networkタブで `main.tsx` の中身を開き、JSXがどんなJavaScriptに変換されているかを読む
2. propsを受け取るコンポーネントを1つ書く（名前と回数を受け取り、挨拶を回数ぶん表示する、など）。propsの型は `type` で定義する。必須のpropsを渡し忘れると、どこで、どんなエラーが出るか。コンポーネントの名前を小文字で始めると、何が起きるか。JSXの中で `class` と書いたとき、ブラウザとESLintは何と言うか
3. 文字列の配列を `ul` として表示する。`key` を付けない場合、Consoleには何が出るか。次に、各要素の中に入力欄を1つずつ置き、`key` に配列の添字を使った場合と、要素ごとに固定の番号を使った場合で、先頭の要素を削除したあとに、入力欄に打った文字がどこへ行くかを比べる。なぜ違うか。「Rendering Lists」の `key` の項と照らして `notes.md` に書く
4. `useState` でカウンターを作る。ボタンを押す処理の中で、stateを更新した直後に、その変数を `console.log` する。何が表示されるか。同じ処理の中で、更新を2回続けて書くと、いくつ増えるか。「State as a Snapshot」と「Queueing a Series of State Updates」を読んで、観察した結果を説明する
5. 入力欄を制御されたコンポーネントとして書く（`value` と `onChange`）。`onChange` を書かずに `value` だけを指定すると、入力欄はどうなるか。Consoleには何が出るか。次に、`form` の `onSubmit` で入力内容を受け取る。T3の課題5と同じ問題（ページが再読み込みされる）は、Reactでも起きるか。イベントオブジェクトの型は、何と書くか（Using TypeScriptで探す）
6. 配列のstateを持つコンポーネントで、`push` で要素を足してから同じ配列を `set` する場合と、新しい配列を作って `set` する場合を比べる。画面は、それぞれ更新されるか。なぜか。「Updating Arrays in State」の表を見て、追加・削除・置き換えのそれぞれで使うべきメソッドを `notes.md` に書く
7. **本題**：T3のToDoを、Reactで作り直す。T3の課題6の5項目（追加・空入力の拒否・完了の切り替え・削除・未完了の件数）をすべて満たす。ToDo1件の型は、T5で定義したものを使う。コンポーネントは少なくとも、全体、入力フォーム、一覧、1件、の4つに分ける。ToDoの配列のstateは、どのコンポーネントに置くべきか。「Sharing State Between Components」を読んで決め、理由を `notes.md` に書く。T3の課題8と同じ文字列（`<b>太字</b>`、`<img src=x onerror="alert(1)">`）を入力すると、Reactではどう表示されるか
8. `useEffect` で、未完了の件数を `document.title` に反映する。effectの中と、コンポーネントの関数の本体の両方に `console.log` を置き、ページを開いたときと、ToDoを1件足したときに、それぞれ何回表示されるかを数える。依存配列を「書かない」「空にする」「件数を入れる」の3通りで比べる
9. **本題**：StrictModeの観察。`main.tsx` でコンポーネントを `StrictMode` で包んだ状態と外した状態で、課題8の回数を比べる。`npm run build` して `npm run preview` で開いた場合は何回か。3つの結果を `notes.md` に書く。開発中に2回実行されるのは何のためか、「Synchronizing with Effects」の該当する項を読んで書く。実験が終わったら、`StrictMode` は付けた状態に戻す
10. **本題**：クリーンアップ。ページを開いてからの経過秒数を表示するコンポーネントを、`useEffect` の中の `setInterval` で作る。まずクリーンアップを書かずに、StrictModeの下で動かす。秒数はどう増えるか。次に、このコンポーネントを表示・非表示で切り替えるボタンを親に置き、何度か切り替えたあとにConsoleとタイマーの動きを見る。そのうえで、クリーンアップを書き、両方の問題が消えることを確かめる。`window` の `resize` イベントのリスナーでも、同じことを繰り返す
11. 課題8の「未完了の件数」を、stateとeffectで持つ書き方と、描画のたびに配列から計算する書き方の2通りで書き、コードの量と、更新のタイミングを比べる。「You Might Not Need an Effect」を読み、どちらが勧められているかと、その理由を `notes.md` に書く。effectが本当に必要なのは、どんな処理か
12. ToDoの配列を変える処理（追加、完了の切り替え、削除）を、コンポーネントの外の純粋な関数（配列と操作の内容を受け取り、新しい配列を返す）に切り出す。元の配列を変えていないことも含めて、Vitestでテーブル駆動テストを書く。Reactを使わずにテストできるのは、どんな処理か

### 発展（任意）

- ページを再読み込みしてもToDoが消えないようにする。`localStorage` への保存はeffectで行い、読み込みは `useState` の初期値で行う。読み込んだ値は、T8の検証関数に通す
- 「すべて／未完了／完了」の絞り込みを付ける。絞り込みの条件はstateで持ち、表示する配列は描画のたびに計算する
- `useState` の代わりに `useReducer` を使う書き方を、リファレンスで調べて、課題12の関数と組み合わせる

## 詰まりやすいところ

- `.tsx` を読み込むと、JSXに関するエラーが出る。または、`React is not defined` と出る → `tsconfig.json` の `jsx` オプションと、Viteのドキュメントの、JSXの項を読み直す。準備で決めた設定を見直す
- `Each child in a list should have a unique "key" prop` と出る → 課題3
- 入力欄に文字が打てない → 課題5。`value` に渡しているstateが、`onChange` で更新されているかを見る
- stateを変えたのに画面が変わらない → 課題6。同じ配列やオブジェクトを、そのまま `set` していないか
- `Too many re-renders` と出る → `onClick={f()}` と `onClick={f}` の違いを思い出す（T3の詰まりやすいところにも同じものがある）。描画の途中でstateを更新していないか
- effectが2回動く → 課題9。バグではない。クリーンアップが書いてあれば問題にならないはずなので、ならない理由を探す

## 完了条件の確かめ方

1. `npm run dev` で開いた画面で、T3の課題6の5項目がすべて動く。`StrictMode` が付いている。`main.tsx` と `index.html` に、外部のURLへの参照がない（`npm run build` のあと `grep -rE 'https?://' dist/`）
2. `npx tsc --noEmit --strict`、`npx eslint .`、`npx vitest run` が、すべて終了コード0で終わる。課題12のテストが含まれている
3. 課題10のコンポーネントを表示・非表示で10回切り替えたあと、Consoleにエラーがなく、タイマーが1つだけ動いている（クリーンアップに `console.log` を置いて数える）
4. `notes.md` に、課題9の3つの回数と、その理由が書いてある

## 残すもの

- `package.json`、`package-lock.json`、`tsconfig.json`、`vite.config.ts`、ESLintの設定ファイル、`index.html`、`src/`
- `notes.md`（準備で決めた設定の理由、課題3・4・6・7・9・11の観察と考察）
- `node_modules` と `dist/` は、コミットしない

コードとメモは、リポジトリの中のこのディレクトリに、AI機能を切ったエディタで書く。この `TASKS.md` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- T10（ReactとAPIの接続）：`useEffect` の中で `fetch` を呼び、クリーンアップで途中の応答を捨てる
- T11（SSEの受信）：`useEffect` で接続を開き、クリーンアップで閉じる。課題10と同じ形
- C5（卒業制作のweb）：watchdeckの一覧画面を、ここで分けたコンポーネントの単位で作る
