---
title: "はじめかた（Day 0）"
---

まっさらなマシンから、最初のAIなしデーを迎えられる状態にするまでの手順。

**この準備の日（Day 0）は、AIも検索もコピー＆ペーストも使ってよい。** ここに出てくるコマンドの意味は、今は分からなくてよい。Stage 0のユニット（F2、F6、F7、F10）で、1つずつ自分の手で学び直す。

所要時間は2〜3時間。

## 全体の流れ

1. Ubuntu（Linuxの一種）のシェルを開けるようにする ← 手元のOSによって手順が違うのは、ここだけ
2. 道具を入れる
3. GitHubのアカウントを用意し、Ubuntuからつなぐ
4. このテンプレートから、自分の学習用リポジトリを作る
5. エディタを用意し、AI機能を切ったプロファイルを作る
6. 教材を用意し、最初のAIなしデーの日付を決める

学習はすべてUbuntuの中で行う。閉域の現場で触るのはLinuxで、書籍もLinuxを前提にしている。macOSやWindowsのターミナルは、同じ名前のコマンドでも動きが違うことがあるので、最初から環境をそろえる。

## 1. Ubuntuのシェルを開く

### Windows 10 / 11 の場合（WSL2）

WSL2は、Windowsの中でLinuxを動かす公式の仕組み。

1. スタートメニューで「PowerShell」を右クリックし、「管理者として実行」を選ぶ
2. 次を実行し、終わったらWindowsを再起動する

   ```powershell
   wsl --install
   ```

3. 再起動後に開く「Ubuntu」の画面で、Ubuntu用のユーザー名とパスワードを決める（Windowsのものと別でよい）。パスワードは、打っても画面に何も出ないが、入力されている
4. 次からは、スタートメニューの「Ubuntu」を開けばシェルが出る

うまくいかないときは、BIOS（UEFI）の設定で「仮想化支援機能（Intel VT-x / AMD-V）」が有効かを確認する。

ファイルは、Ubuntu側のホーム（`~`）に置く。Windows側（`/mnt/c/...`）に置くと、動作が遅くなり、パーミッションの学習もできない。

WSL2の中で動かしたサーバー（Stage 2以降）は、Windowsのブラウザから `http://localhost:ポート番号` で開ける。

> Windows 11（WSL 2.5）とUbuntu 26.04 LTSで検証済み（2026-09）。

### macOS の場合（Lima）

Limaは、macOSの中にLinuxの仮想マシン（VM。マシンの中で動く、もう1台のマシン）を作る道具。Intel・Apple Siliconの両方で動く。

1. 「ターミナル」アプリを開く（Spotlightで「ターミナル」と検索）
2. Homebrew（macOS用のパッケージ管理ツール）が入っていなければ、<https://brew.sh> の1行をコピーして実行する。途中で聞かれるパスワードは、打っても画面に何も出ないが、入力されている。終わったときに「Next steps」としてコマンドが2〜3行表示されたら、それもコピーして実行する（`brew` コマンドを使えるようにする設定）
3. Limaを入れ、UbuntuのVMを作る（初回は数分かかる）

   ```sh
   brew install lima
   limactl start --name=offgrid --mount-none --containerd=none template://ubuntu-lts
   ```

   途中で確認を求められたら、「Proceed with the current configuration」を選ぶ。`--mount-none` は、Macのファイルを VM から見えなくする指定。学習用のファイルは、すべてVMの中に置く。

4. Ubuntuのシェルに入る

   ```sh
   limactl shell offgrid
   cd ~
   ```

   入った直後に `cd: ... No such file or directory` という警告が出ることがあるが、問題ない。`cd ~` でUbuntu側のホームに移動する。

5. 次からは、ターミナルで `limactl shell offgrid` を実行すればよい。Macを再起動した後は、先に `limactl start offgrid` でVMを起動する

VMの中で動かしたサーバー（Stage 2以降）は、Macのブラウザから `http://localhost:ポート番号` で開ける。

> macOS（Apple Silicon）とLima 2.2、Ubuntu 26.04 LTSで検証済み（2026-09）。

### Linux の場合

そのまま使う。Ubuntu・Debian系以外のディストリビューションでは、以降の `apt` を、手元のパッケージ管理コマンド（`dnf`、`pacman` など）に読み替える。

## 2. 道具を入れる

ここから先は、どのOSでも同じ。Ubuntuのシェルで実行する。

```sh
sudo apt update
sudo apt install -y git gh shellcheck
```

- `git`：変更履歴を管理する道具（F6で学ぶ）
- `gh`：GitHubをコマンドで操作する道具
- `shellcheck`：シェルスクリプトの間違いを指摘してくれる道具（F5で使う）

`sudo` で聞かれるパスワードは、手順1で決めたUbuntuのもの（Limaでは聞かれない）。`man`、`xxd`、`grep`、`curl`、`vim`、`nano` などは、最初から入っている。Dockerは、F10で入れる。

**学習用コマンド `offgrid` を入れる。** セッションを案内してくれる道具で、今日いちばん大事な準備。下はそのまま貼ればよい（CPUの種類は、`uname -m` の結果から自動で選ぶ）。

```sh
mkdir -p ~/.local/bin
case "$(uname -m)" in
  aarch64|arm64) arch=arm64 ;;
  x86_64|amd64)  arch=amd64 ;;
  *) echo "分からないCPUです: $(uname -m)"; arch= ;;
esac
curl -fsSL -o ~/.local/bin/offgrid \
  "https://github.com/kom3da/offgrid/releases/latest/download/offgrid-linux-$arch"
chmod +x ~/.local/bin/offgrid
offgrid version
```

`offgrid version` が版を表示すれば成功。`Exec format error` と出たら、CPUの種類が合っていない（`uname -m` の結果を見て、`arch=` を手で書き直す）。

`offgrid: command not found` と出たら、`~/.local/bin` が `PATH` に入っていない。次を実行して、シェルを開き直す（`PATH` の意味はF4で学ぶ）。

```sh
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
```

このコマンドは、ネットワークもAIも使わない。入れるのは今日だけで、あとはオフラインでも動く。ターミナルが狭いと感じたら、`offgrid serve` で手元にブラウザの画面も立てられる（[学習用コマンド](04-ai-free-day.md#学習用コマンド)）。

## 3. GitHubにつなぐ

1. <https://github.com> でアカウントを作る（既にあれば不要）
2. Ubuntuのシェルでログインする。質問には「GitHub.com」「HTTPS」「Login with a web browser」を選び、表示されたコードをブラウザに入力する

   ```sh
   gh auth login
   gh auth setup-git
   ```

3. コミット（変更の記録）に残す名前とメールアドレスを設定する。メールアドレスを公開したくなければ、GitHubの Settings → Emails に表示される `...@users.noreply.github.com` を使う

   ```sh
   git config --global user.name "自分の名前またはID"
   git config --global user.email "メールアドレス"
   git config --global init.defaultBranch main
   ```

   3行目は、新しく作るリポジトリの最初のブランチ名を `main` にそろえる設定（設定しないと `master` になることがある）。

SSH鍵を使う接続は、F7で自分の手で設定する。

## 4. 自分の学習用リポジトリを作る

このリポジトリ（`kom3da/offgrid`）はテンプレート。ここに直接書き込むのではなく、コピーして自分のリポジトリを作り、その中で学習する。

```sh
cd ~
gh repo create offgrid-log --template kom3da/offgrid --private --clone
cd offgrid-log
```

- `offgrid-log` は自分のリポジトリの名前。好きな名前でよい
- `--private` は非公開。公開して続ける励みにしたいなら `--public` にする。どちらの場合も、「いつか公開されるかもしれない」前提で書く（仕事の内容、実在のホスト名、認証情報は書かない）
- 学習は、このリポジトリの `main` ブランチでそのまま進める。ブランチの使い方はF7で学ぶ

`site/` と `.github/` は、テンプレート側の道具（ドキュメントサイトと、その検査）。自分のリポジトリでは使わないので、消してよい。

```sh
rm -rf site .github
```

消さなくても、中の設定は「テンプレートのリポジトリでだけ動く」ようにしてあるので、自分のpushで無駄に動くことはない。

## 5. エディタを用意する

エディタは何でもよい。**満たすべき要件は2つだけ**。

1. Ubuntuの中のファイルを編集できる
2. AIの機能を、確実に切れる（または最初から無い）

好きなものを選んでよいが、迷ったら下の「A」がいちばん簡単。

### A. Ubuntuの中のエディタを使う（おすすめ。追加の設定が要らない）

`vim` と `nano` は、最初から入っている。AIの機能は無いので、切る手間もない。閉域の現場で使うのも、たいていこれ。

```sh
nano ~/offgrid-log/README.md   # 保存は Ctrl+O、終了は Ctrl+X
```

どちらもF6で練習する。最初は `nano` で足りる。慣れたら `vim` に移ってもよい。

### B. 手元のOSのエディタから、Ubuntuの中を開く

画面の広いエディタで書きたい場合。Ubuntuの中のファイルを開くには、リモート接続の仕組みが要る。

- **Visual Studio Code**（例）
  - Windows：拡張機能「WSL」を入れる。Ubuntuの画面を一度閉じて開き直し、`cd ~/offgrid-log` のあと `code .` を実行する
  - macOS：拡張機能「Remote - SSH」を入れる。Macのターミナル（Ubuntuの外）で次を実行し、開いたファイルの**先頭**に `Include ~/.lima/*/ssh.config` と1行書いて保存する。左下の「><」→「ホストに接続する」に `lima-offgrid` が出る

    ```sh
    mkdir -p ~/.ssh && touch ~/.ssh/config && open -e ~/.ssh/config
    ```

  - Linux：そのまま開く
  - AIを切る：歯車 →「プロファイル」→「プロファイルの作成」で `no-ai` を作る（「コピー元」に既定のプロファイルを選ぶ。空にすると上の拡張機能まで消える）。そのプロファイルで、AIの拡張機能（Copilot、Claude Codeなど）を無効にし、設定で `chat.disableAIFeatures` をオンにする。AIなしデーの朝に、このプロファイルへ切り替える
- **JetBrains の IDE**（GoLand など）：Remote Development で接続する。AI Assistant のプラグインを無効にする
- **Neovim、Helix、Emacs など**：Ubuntuの中に入れて使う。AIのプラグインを入れなければ、それで満たせる

### どれを選んでも、最初のセッションの前に確かめること

- そのエディタで、`~/offgrid-log` の中のファイルを保存できる
- **補完の候補が、AIから出てこない**（コメントを書いて、続きが勝手に生えてこないか試す）
- AIなしデーの朝に、AIを切った状態へ切り替える手順を、自分の言葉で言える

### 平日に使うAIの置き場所

リポジトリはUbuntuの中にあるので、セッションの間の日にレビューや採点を頼むAIも、Ubuntuの中のファイルを読める必要がある。どちらかを選ぶ。

- 手元のOSのエディタから、AIを切っていない状態でUbuntuに接続し、そこでAIの拡張機能を使う
- ターミナルで使うAI（Claude Codeなど）を、Ubuntuの中にインストールする（入れ方は、その道具の公式の手順に従う）

## 6. 教材と日付

- 『コンピュータはなぜ動くのか』と『新しいLinuxの教科書』を入手する（[教材・リソース](20-resources.md)）
- 本が手元に無くても始められる。オフラインのドキュメントを先に揃える必要もない（[教材の手元への置きかた](20-resources.md)）
- [学習ロードマップ](03-roadmap.md)と[AIなしデーの運用](04-ai-free-day.md)を読む
- 最初のAIなしデーの日付を決め、カレンダーに入れる。最初の3回は午前だけ（3時間）。週に何回やるかは自由（[ペースを決める](04-ai-free-day.md)）
- テンプレートの改訂を知りたければ、GitHubの <https://github.com/kom3da/offgrid> で「Watch」→「Custom」→「Releases」を選ぶ（新しい版の `offgrid` も、そのリリースに付く）

## AIなしデーの始め方と終わり方

始めるとき：

```sh
cd ~/offgrid-log
offgrid
```

今日が何回目で、何をやるかを表示する。AIもネットワークも使わない。表示された課題シート（`TASKS.md`）を開いて始める。

メニューの「1. セッションを始める」を選ぶと、ウォームアップから終わりの手続きまで、課題を1問ずつ案内する。振り返りのファイルも、そこで用意される。まず覚えるのは `offgrid` の1つだけでよい。ほかのサブコマンドは[学習用コマンド](04-ai-free-day.md#学習用コマンド)にある。

コマンドを試す場所は `~/sandbox/`、メモや `explain.md` を書く場所は、課題シートと同じディレクトリ。`TASKS.md` は書き換えない。

終わるときは、次を実行する。振り返りの書き忘れ、残すものの不足、コミットの漏れを教えてくれる。

```sh
offgrid end
```

コミットメッセージを聞かれるので、その場で書けば `offgrid` がコミットする。**ただしF6でGitを学ぶまでは、Enterで飛ばして、表示されたコマンドを自分で打つことを勧める。** 毎回同じものを手で打つこと自体が練習になるので、コピーせずに打つ（`F1: ...` の部分は、その日にやったことに変える）。

```sh
git add -A
git commit -m "[no-ai] F1: xxdでUTF-8のバイト列を確認"
git push origin main
```

ユニットの完了条件を満たしたら、`offgrid done F1` で `PROGRESS.md` に記録する。F1 は「説明できる」が完了条件に入っているので、当日は記録できない（`explain.md` の採点は平日）。`offgrid end` で次のユニットを聞かれたら F2 と答え、`done F1` は採点が通ってから打つ。

初回のウォームアップは、再現する内容がまだないので、「Ubuntuのシェルを開き、`cd ~/offgrid-log` で移動し、`ls` で中身を見る」を何も見ずにやる（`offgrid` も、そう表示する）。

このカリキュラムは[ブラウザでも読める](https://kom3da.github.io/offgrid/)が、セッション中に開くのは、自分のリポジトリの中のMarkdownでよい。

## カリキュラムが改訂されたとき

テンプレートの内容（学習テキスト）は、誤りの修正や教材の更新のために、ときどき改訂される（[改訂の運用](30-revision.md)）。自分の学習の進み具合とは関係がないので、取り込むかどうか、いつ取り込むかは自分で決めてよい。取り込むには、AIを使ってよい日に、版を指定して実行する。版の一覧と変更点は、[CHANGELOG](../CHANGELOG.md)にある。

```sh
./scripts/update-curriculum.sh <版>     # 例: ./scripts/update-curriculum.sh v1.2.0
git diff --stat                         # 何が変わったかを確認する
git add -A && git commit -m "[docs] Update curriculum to <版>"
```

版の最初の数字（`1`）が上がったときは、**学習用コマンドも入れ替える**（手順2のダウンロードをもう一度やる）。課題シートの書き方が変わり、古いツールでは読めなくなるため。

途中で失敗したら、`git restore . && git clean -fd docs prompts templates scripts tools` で、取り込む前の状態に戻せる。

置き換わるのは、`docs/`、`prompts/`、`templates/`、`scripts/`、`tools/`、`CLAUDE.md`、`README.md`、`CHANGELOG.md` と、`drills/` の中の `TASKS.md`（課題シート）だけ。これらは自分では書き換えない（書き換えても、次の取り込みで消える）。

自分のものとして自由に書いてよいのは、`PROGRESS.md`、`logs/`、`answers/`、`capstone/` と、`drills/` の中の `TASKS.md` 以外のファイル。取り込みでは触れられない。
