# F6：エディタとGit入門（課題シート）

この `TASKS.md` はカリキュラムの一部なので、書き換えない。コマンドを試す場所は、下の「準備」で作る `~/sandbox/` の中。メモ（`notes.md`）、`explain.md`、自分で書いたスクリプトは、リポジトリの中のこのディレクトリに、AI機能を切ったVS Codeで書く（F6からは `vim` や `nano` でもよい）。

## ねらい

ここまで「決まり文句」として打ってきたGitのコマンドの意味を理解し、自分の判断でコミットを積めるようになる。ターミナルの中のエディタにも慣れる。

## 平日に読むもの

『Pro Git』の第1章（使い始める）と第2章（Gitの基本）。

## 準備

練習は、学習用リポジトリの外で行う。

```sh
mkdir -p ~/sandbox/f06 && cd ~/sandbox/f06
```

## 課題

1. エディタ：`nano` でファイルを作り、保存して終了する。次に `vimtutor` を30分だけ進める。`vim` で「入力を始める」「入力をやめる」「保存する」「保存せずに終了する」方法を、メモに書く。どちらを普段使うかを決める
2. `git init` でリポジトリを作る。`git status` の1行目で、ブランチ名が `main` になっていることを確かめる（[はじめかた](../../../docs/00a-getting-started.md)の `init.defaultBranch` の設定が効いている）。ファイルを1つ作り、`git status` → `git add` → `git status` → `git commit -m` → `git status` の順に実行し、表示がどう変わるかを追う。「作業ツリー」「ステージ」「コミット」の3つの場所の関係を、図に描く
3. ファイルを変更し、`git diff` を見る。`git add` したあとの `git diff` と `git diff --staged` の違いを確かめる。コミットしたあと、`git log`、`git log --oneline`、`git show` を試す
4. 取り消し：変更をまだ `add` していないとき、`add` したあと、それぞれの戻し方を調べて試す（`git restore`、`git restore --staged`）。直前のコミットのメッセージを直す方法（`git commit --amend`）を試す
5. `.gitignore`：`*.log` と `secret.env` を無視する設定を書き、`git status` で確かめる。すでにコミットしたファイルを `.gitignore` に足すと、どうなるか
6. コミットの単位：3つのファイルを変更し、関係のある変更だけを選んで、2回に分けてコミットする
7. `explain.md`：毎回の終わりに打ってきた4行（`git add -A`、`git commit -m`、`git tag`、`git push origin main --tags`）が、それぞれ何をしているかを、何も見ずに書く
8. 学習用リポジトリ（`~/offgrid-log`）で、`git log --oneline` と `git tag` を実行し、自分のこれまでの記録を読む。`git show` で、最初のAIなしデーのコミットの中身を見る

## 完了条件の確かめ方

- 学習用リポジトリに、コマンドの意味を分かったうえで、自分の手でコミットを積める
- 課題7の `explain.md` を、平日にAIに採点してもらう

## 残すもの

`explain.md`、課題2の図（写真か、文字で描いたもの）、`notes.md`。
