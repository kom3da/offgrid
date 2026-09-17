# F7：Gitの応用とGitHub（課題シート）

この `TASKS.md` はカリキュラムの一部なので、書き換えない。コマンドを試す場所は、下の「準備」で作る `~/sandbox/` の中。メモ（`notes.md`）、`explain.md`、自分で書いたスクリプトは、リポジトリの中のこのディレクトリに、AI機能を切ったVS Codeで書く（F6からは `vim` や `nano` でもよい）。

## ねらい

SSH鍵でGitHubにつなぎ、ブランチを使った開発の流れ（分岐、マージ、コンフリクトの解消、プルリクエスト）を一通り経験する。

## 平日に読むもの

『Pro Git』の第3章（Gitのブランチ機能）と、GitHubの章。

## 準備

```sh
mkdir -p ~/sandbox/f07 && cd ~/sandbox/f07 && git init
```

GitHubを使う課題（1、6）のために、練習用のリポジトリをGitHubに1つ作る。最初のコミットを `main` に積んでから、次を実行する（`--push` で、`main` をGitHubに送る）。終わったら消してよい。

```sh
echo '# f07 practice' > README.md && git add README.md && git commit -m 'first commit'
gh repo create f07-practice --private --source=. --remote=origin --push
```

## 課題

1. SSH鍵
   - `ssh-keygen -t ed25519` で鍵を作る。できた2つのファイルのうち、人に渡してよいのはどちらか。それぞれのパーミッションを `ls -l` で確かめ、なぜそうなっているかを考える（F2）
   - 公開鍵をGitHubに登録し、`ssh -T git@github.com` でつながることを確かめる
   - `git remote -v` で今の接続先を見て、`git remote set-url` でSSHのURLに変える
   - `~/.ssh/config` に、`github.com` へつなぐときに使う鍵（`IdentityFile`）を書く。このファイルは何のためにあるか。パーミッションはどうあるべきか
2. ブランチ：`git switch -c feature` でブランチを作り、コミットを2つ積む。`main` に戻り、ファイルがどう見えるかを確かめる。`git merge feature` で取り込む。`git log --graph --oneline --all` で形を見る
3. マージの2つの形：`main` 側にもコミットがある状態で、別のブランチをマージする。課題2のときと、履歴の形がどう違うか
4. **本題** コンフリクト：2つのブランチで、同じファイルの同じ行を、違う内容に変更する。マージして、コンフリクトを起こす。ファイルに入った印（`<<<<<<<`、`=======`、`>>>>>>>`）を読み、手で直してコミットする
5. `rebase`：課題3と同じ状況をもう一度作り、今度は `git rebase main` で取り込む。履歴の形は、マージのときとどう違うか。「push済みのブランチをrebaseしてはいけない」と言われる理由を調べる
6. GitHub：ブランチを `git push -u origin ブランチ名` で送る。`gh pr create` でプルリクエストを作り、ブラウザで差分を見て、マージする。手元の `main` を `git pull` で最新にする。`git fetch` と `git pull` の違いは何か
7. 取り消し：`git revert` と `git reset`（`--soft`、`--hard`）の違いを、練習用のリポジトリで試す。push済みのコミットを取り消すなら、どちらを使うべきか。なぜか

## 完了条件の確かめ方

何も見ずに、次ができる：わざとコンフリクトを作る → 解消する → GitHubにpushする。次回のウォームアップで、もう一度やる。

## 残すもの

`notes.md`（課題3と5の履歴の形を、`git log --graph` の出力で貼る）。秘密鍵は、絶対にコミットしない。
