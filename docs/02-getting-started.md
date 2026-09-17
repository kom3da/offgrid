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

> この手順は、作成時に実機で検証できていない。詰まった箇所は、詰まりメモに残して直す。

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

`site/` と `.github/` は、テンプレートのドキュメントサイト用。自分のリポジトリでは使わないので、消してよい（残しても害はない）。

## 5. エディタを用意する

AIなしデーでは、AI機能を切ったエディタを使う。

1. 手元のOS（Ubuntuの外）に、Visual Studio Codeを入れる
2. VS Codeから、Ubuntuの中のファイルを開けるようにする
   - Windows：拡張機能「WSL」を入れる。Ubuntuの画面を一度閉じて開き直し、`cd ~/offgrid-log` のあと `code .` を実行する
   - macOS：拡張機能「Remote - SSH」を入れる。Macのターミナル（Ubuntuの外）で次を実行し、開いたファイルの**先頭**に `Include ~/.lima/*/ssh.config` と1行書いて保存する。VS Codeの左下の「><」→「ホストに接続する」に `lima-offgrid` が出るので、接続して `offgrid-log` フォルダを開く

     ```sh
     mkdir -p ~/.ssh && touch ~/.ssh/config && open -e ~/.ssh/config
     ```

   - Linux：そのまま開く
3. 左下の歯車 →「プロファイル」→「プロファイルの作成」で、`no-ai` という名前のプロファイルを作る。「コピー元」に既定のプロファイルを選ぶ（空のプロファイルにすると、手順2の拡張機能まで消える）
4. `no-ai` プロファイルで、AIの拡張機能（Copilot、Claude Codeなど）を無効にする。さらに、設定で `chat.disableAIFeatures` を検索してオンにする（VS Code本体に組み込まれたAI機能を切る）
5. AIなしデーの朝に `no-ai` プロファイルへ切り替える

### 平日に使うAIの置き場所

リポジトリはUbuntuの中にあるので、平日にレビューや採点を頼むAIも、Ubuntuの中のファイルを読める必要がある。どちらかを選ぶ。

- VS Codeの既定のプロファイル（AIあり）でUbuntuに接続し、その中でAIの拡張機能を使う
- ターミナルで使うAI（Claude Codeなど）を、Ubuntuの中にインストールする（入れ方は、その道具の公式の手順に従う）

F6からは、ターミナルの中のエディタ（`vim` か `nano`）も使う。

## 6. 教材と日付

- 『コンピュータはなぜ動くのか』と『新しいLinuxの教科書』を入手する（[教材・リソース](20-resources.md)）
- Stage 0のオフライン教材を用意する（[ステージ開始時のオフライン教材セットアップ](20-resources.md)）
- [学習ロードマップ](03-roadmap.md)と[AIなしデーの運用](04-ai-free-day.md)を読む
- 最初のAIなしデーの日付を決め、カレンダーに入れる。最初の3回は午前だけ（3時間）。週に何回やるかは自由（[ペースを決める](04-ai-free-day.md)）
- テンプレートの改訂を知りたければ、GitHubの <https://github.com/kom3da/offgrid> で「Watch」→「Custom」→「Releases」を選ぶ

## AIなしデーの始め方と終わり方

始めるとき：

```sh
cd ~/offgrid-log
./scripts/offgrid
```

今日が何回目で、何をやるかを表示し、振り返りのファイルを用意する。AIもネットワークも使わない。表示された課題シート（`TASKS.md`）を開いて始める。

セッション中に使えるほかのコマンド（`status`、`find`、`check`、`stuck`、`timer`、`done`、`end`）は、[学習用コマンド](04-ai-free-day.md#学習用コマンド)にある。まず覚えるのは、始めの `offgrid` と、終わりの `offgrid end` の2つだけでよい。

コマンドを試す場所は `~/sandbox/`、メモや `explain.md` を書く場所は、課題シートと同じディレクトリ。`TASKS.md` は書き換えない。

終わるときは、次を実行する。振り返りの書き忘れ、残すものの不足、コミットの漏れを教えてくれる。

```sh
./scripts/offgrid end
```

そのあと、表示されたコマンドをそのまま打つ。F6でGitを学ぶまでは、意味が分からなくてよい。毎回同じものを手で打つこと自体が練習になるので、コピーせずに打つ（`F1: ...` の部分は、その日にやったことに変える）。

```sh
git add -A
git commit -m "[no-ai] F1: xxdでUTF-8のバイト列を確認"
git push origin main
```

ユニットの完了条件を満たしたら、`./scripts/offgrid done F1` で `PROGRESS.md` に記録する。

初回のウォームアップは、再現する内容がまだないので、「Ubuntuのシェルを開き、`cd ~/offgrid-log` で移動し、`ls` で中身を見る」を何も見ずにやる（`./scripts/offgrid` も、そう表示する）。

このカリキュラムは[ブラウザでも読める](https://kom3da.github.io/offgrid/)が、セッション中に開くのは、自分のリポジトリの中のMarkdownでよい。ネットワークを切るStage 4〜5では、それが唯一の読み方になる。

## カリキュラムが改訂されたとき

テンプレートの内容（学習テキスト）は、誤りの修正や教材の更新のために、ときどき改訂される（[改訂の運用](30-revision.md)）。自分の学習の進み具合とは関係がないので、取り込むかどうか、いつ取り込むかは自分で決めてよい。取り込むには、AIを使ってよい日に、版を指定して実行する。版の一覧と変更点は、[CHANGELOG](../CHANGELOG.md)にある。

```sh
./scripts/update-curriculum.sh v2026.10.2
git diff --stat          # 何が変わったかを確認する
git add -A && git commit -m "[docs] Update curriculum to v2026.10.2"
```

途中で失敗したら、`git restore . && git clean -fd docs prompts templates scripts` で、取り込む前の状態に戻せる。

置き換わるのは、`docs/`、`prompts/`、`templates/`、`scripts/`、`CLAUDE.md`、`README.md`、`CHANGELOG.md` と、`drills/` の中の `TASKS.md`（課題シート）だけ。これらは自分では書き換えない（書き換えても、次の取り込みで消える）。

自分のものとして自由に書いてよいのは、`PROGRESS.md`、`logs/`、`answers/`、`capstone/` と、`drills/` の中の `TASKS.md` 以外のファイル。取り込みでは触れられない。
