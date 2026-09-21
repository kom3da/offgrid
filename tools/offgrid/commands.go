package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// welcome は、はじめて使う人に、3つだけ伝える。
func welcome(root string, p *Progress, s *Session) {
	if !IsFirstTime(root, s) {
		return
	}
	banner([]string{
		bold("はじめまして。使い方は3つだけ"),
		"1. この道具が「今やること」を1つずつ出す。上から順に手を動かす",
		"2. コマンドは、このターミナルで打つ。エディタと本は、必要なときに開く",
		"3. できたら Enter。分からなければ k（手がかり）、止まったら s（詰まった）",
	})
	fmt.Println()
	say("画面を広く使いたいなら " + bold("offgrid serve") + " で、ブラウザでも同じ案内ができる")
	say(dim("答えは出さない。自分で気づくのが目的だから"))
	fmt.Println()
}

// ---------------------------------------------------------------- 画面

func dashboard(root string, p *Progress, s *Session) {
	title := ""
	if sh, err := LoadSheet(root, p.Unit); err == nil {
		title = sh.Title
	}
	banner([]string{
		bold(fmt.Sprintf("offgrid   第%d回   Stage %s", s.Number, p.Stage)),
		fmt.Sprintf("次のユニット: %s  %s", bold(p.Unit), title),
		"今日の終わり: " + s.Finish(),
	})
	fmt.Println()
	rule("進み具合")
	td, ta := 0, 0
	for _, t := range p.Tracks {
		if len(t.Units) == 0 {
			continue
		}
		done := 0
		for _, u := range t.Units {
			if u.Done {
				done++
			}
		}
		td, ta = td+done, ta+len(t.Units)
		fmt.Printf("  %s %s %2d/%d\n", pad(t.Name, 30), progressBar(done, len(t.Units)), done, len(t.Units))
	}
	if ta > 0 {
		fmt.Printf("  %s %s %2d/%d\n", pad("合計", 30), progressBar(td, ta), td, ta)
	}
	if sh, err := LoadSheet(root, p.Unit); err == nil {
		done, total := sh.Counts()
		fmt.Printf("  %s %s %2d/%d\n", pad(p.Unit+" の課題", 30), progressBar(done, total), done, total)
	}
	if gaps := FindGaps(root, p, s); len(gaps) > 0 {
		fmt.Println()
		rule("気になるところ")
		for _, g := range gaps {
			say(yellow("! ") + g)
		}
	}
}

func showTask(sh *Sheet, n int) {
	done, total := sh.Counts()
	rule(fmt.Sprintf("%s 課題 %d／%d（できた %d）", sh.Unit, n, total, done))
	fmt.Println()
	say(sh.Tasks[n])
	fmt.Println()
}

func showSections(sh *Sheet, names ...string) {
	for _, name := range names {
		body := sh.Section(name)
		if body == "" {
			continue
		}
		rule(name)
		for _, line := range strings.Split(body, "\n") {
			if strings.TrimSpace(line) == "" {
				fmt.Println()
			} else {
				fmt.Println("  " + line)
			}
		}
		fmt.Println()
	}
}

// ---------------------------------------------------------------- 案内

// guideUnit は、課題を1問ずつ出す。「あとで」の印は state に残すので、
// ブラウザ（offgrid serve）に移っても、同じところから続けられる。
func guideUnit(sh *Sheet, s *Session, state *SessionState) error {
	created, added, err := sh.EnsureWork()
	if err != nil {
		return err
	}
	if created {
		fmt.Println(dim("  作業記録を作りました: " + rel(sh.root, sh.WorkPath())))
	} else if len(added) > 0 {
		fmt.Println(dim(fmt.Sprintf("  課題が増えていたので、作業記録に%d行足しました", len(added))))
	}
	for {
		state.Reload() // ブラウザ側で「あとで」にした課題を拾う
		next, onlyLater := state.NextTask(sh)
		done, total := sh.Counts()
		if next == 0 {
			if total > 0 && done == total {
				fmt.Println(green(fmt.Sprintf("  ✓ %s の課題は全部できました（%d/%d）", sh.Unit, done, total)))
				say(fmt.Sprintf("完了条件を確かめる → 『offgrid check %s』", sh.Unit))
			} else {
				fmt.Println(dim(fmt.Sprintf("  この枠はここまで（%d/%d）", done, total)))
			}
			return nil
		}
		fmt.Println()
		if onlyLater {
			say(dim("あとで回した課題に戻ってきました"))
		}
		showTask(sh, next)
		if err := state.OpenTask(sh.Unit, next); err != nil {
			say(yellow("開いた時刻を記録できません: " + err.Error()))
		}
		if advice := taskAdvice(state.TaskMinutes(sh.Unit, next)); advice != "" {
			say(yellow(advice))
		}
		got, err := ask("  [Enter]できた  [s]詰まった  [l]あとで  [k]手がかり  [d]枠を終える  [q]中断: ",
			"", "y", "s", "l", "k", "d", "q")
		if err != nil {
			return err
		}
		switch got {
		case "", "y":
			if err := sh.Tick(next, true); err != nil {
				return err
			}
			done, total = sh.Counts()
			fmt.Println(green(fmt.Sprintf("  ✓ 課題%d（%d/%d）", next, done, total)))
		case "s":
			memo, err := ask("  何が起きた？（1行。Enterで飛ばす）: ")
			if err != nil {
				return err
			}
			if memo != "" {
				if err := AppendStuck(sh.root, s.Log, fmt.Sprintf("%s 課題%d: %s", sh.Unit, next, memo)); err != nil {
					return err
				}
				fmt.Println(green("  → 詰まりメモに書きました"))
			}
			say(dim("15分たったら、エラー文を全文読む→man→最小の再現→公式ドキュメント、の順に切り替える。45分で打ち切る"))
			if err := state.Skip(sh.Unit, next); err != nil {
				return err
			}
		case "l":
			if err := state.Skip(sh.Unit, next); err != nil {
				return err
			}
		case "k":
			fmt.Println()
			showSections(sh, "キーワード", "詰まりやすいところ", "平日に読むもの")
		case "d":
			return nil
		case "q":
			return errAbort{}
		}
	}
}

// cmdRun は、その回を上から順に案内する。
// どこまで進んだかは state（logs/<年>/.<月-日>.state.json）に残すので、
// ブラウザ（offgrid serve）と行き来しても、続きから進められる。
func cmdRun(root string, p *Progress, s *Session) error {
	if err := s.EnsureLog(); err != nil {
		return err
	}
	welcome(root, p, s)
	dashboard(root, p, s)
	fmt.Println()
	say(dim("上から順に案内します。q で中断できます（続きから再開できます）"))
	state := LoadState(root, s)
	steps := sessionSteps(p, s)
	for {
		step, ok := state.Current(steps)
		if !ok {
			fmt.Println()
			fmt.Println(green("  今日のぶんは、全部終わりました。お疲れさま"))
			return nil
		}
		index := 1
		for i, v := range steps {
			if v.ID == step.ID {
				index = i + 1
			}
		}
		if err := state.Start(step.ID); err != nil {
			return err
		}
		fmt.Println()
		head := fmt.Sprintf("%d/%d. %s", index, len(steps), step.Label)
		if step.Minutes != "" {
			head += "（" + step.Minutes + "）"
		}
		rule(head)
		if err := runStep(root, p, s, state, step); err != nil {
			return err
		}
		if err := state.Finish(step.ID); err != nil {
			return err
		}
	}
}

// waitOrAbort は「終わったら Enter」を待つ。q なら中断する。
// 答えを捨てていたので、q と打っても中断せず、その枠を終えたことになっていた。
func waitOrAbort() error {
	got, err := ask("\n  終わったら Enter（q=中断）: ", "", "q")
	if err != nil {
		return err
	}
	if got == "q" {
		return errAbort{}
	}
	return nil
}

// runStep は、1つの枠を案内する。
func runStep(root string, p *Progress, s *Session, state *SessionState, step Step) error {
	switch step.ID {
	case StepMain, StepDB:
		unit := p.Unit
		if step.ID == StepDB {
			unit = p.DBUnit
		}
		sh, err := LoadSheet(root, unit)
		if err != nil {
			if step.ID == StepDB {
				say("PROGRESS.md の「次のPostgreSQLユニット」を、いま進めるDBユニット（例：D1）に書き換えると、ここでも案内します")
				return waitOrAbort()
			}
			return err
		}
		say(dim("シート: " + rel(root, sh.Path)))
		return guideUnit(sh, s, state)
	case StepWarmup:
		say(s.Warmup())
		say("＋ 前回のレビューで指摘された箇所の直しを1件")
		if lg := s.ReviewLog(); lg != "" {
			say(dim("見返すログ: " + rel(root, lg)))
		}
		if notes := lastStuckNotes(s); len(notes) > 0 {
			fmt.Println()
			say(yellow("前回、ここで止まっています："))
			for _, n := range notes {
				say("  - " + n)
			}
			say(dim("同じところで止まるなら、それが今の弱点。ウォームアップの題材にしてよい"))
		}
		return waitOrAbort()
	case StepAfternoon:
		if s.Afternoon() == "コードリーディング" {
			say("読解課題は docs/05-debug-and-reading.md の一覧から、今のステージのものを選ぶ。メモは drills/reading/ に置く")
		} else {
			say("問題は drills/debug/ の中。答え（answers/）は、3つ直し終わるまで開かない。直したら、そのバグを見つけるテストを1本足す")
		}
		return waitOrAbort()
	case StepRetro:
		return cmdRetro(root, s, nil)
	case StepEnd:
		return cmdEnd(root, p, s)
	}
	return nil
}

// ---------------------------------------------------------------- 各サブコマンド

func cmdToday(root string, p *Progress, s *Session) error {
	welcome(root, p, s)
	dashboard(root, p, s)
	fmt.Println()
	rule("今日の流れ")
	for i, t := range sessionFlow(root, p, s) {
		say(fmt.Sprintf("%d. %s", i+1, t))
	}
	if s.Number <= 6 {
		fmt.Println()
		say(dim(fmt.Sprintf("最初の6回は短くしてある（今日は%s）。7回目からフルの時間割になる", s.Finish())))
	}
	fmt.Println()
	say(bold("案内つきで進める → offgrid run"))
	return nil
}

func cmdNext(root string, p *Progress, args []string) error {
	unit := p.Unit
	if len(args) > 0 {
		unit = args[0]
	}
	sh, err := LoadSheet(root, unit)
	if err != nil {
		return err
	}
	created, added, err := sh.EnsureWork()
	if err != nil {
		return err
	}
	if created {
		fmt.Println(dim("  作業記録を作りました: " + rel(root, sh.WorkPath())))
	} else if len(added) > 0 {
		fmt.Println(dim(fmt.Sprintf("  課題が増えていたので、作業記録に%d行足しました", len(added))))
	}
	todo := sh.Todo()
	if len(todo) == 0 {
		done, total := sh.Counts()
		fmt.Println(green(fmt.Sprintf("%s の課題は全部できています（%d/%d）", sh.Unit, done, total)))
		say(fmt.Sprintf("完了条件を確かめる → offgrid check %s", sh.Unit))
		return nil
	}
	showTask(sh, todo[0])
	say(dim(fmt.Sprintf("できたら → offgrid tick %d", todo[0])))
	return nil
}

func cmdTick(root string, p *Progress, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("使い方: offgrid tick <課題の番号> [ユニットID]")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("課題の番号は数字で指定してください")
	}
	unit := p.Unit
	if len(args) > 1 {
		unit = args[1]
	}
	sh, err := LoadSheet(root, unit)
	if err != nil {
		return err
	}
	if err := sh.Tick(n, true); err != nil {
		return err
	}
	done, total := sh.Counts()
	fmt.Println(green(fmt.Sprintf("%s 課題%d: できた（%d/%d）", sh.Unit, n, done, total)))
	return nil
}

func cmdStatus(root string, p *Progress, s *Session) error {
	dashboard(root, p, s)
	fmt.Println()
	rule("セッション")
	if len(s.Past) > 0 {
		say(fmt.Sprintf("済んだ回数: %d回（最後: %s）", len(s.Past), rel(root, s.Past[len(s.Past)-1])))
		if n := len(s.Past) % 4; n == 0 {
			say("4回ごとの見直し: 今回がその回")
		} else {
			say(fmt.Sprintf("4回ごとの見直しまで: あと %d回", 4-n))
		}
	} else {
		say("済んだ回数: 0回")
	}
	say(fmt.Sprintf("次の関門: Stage %s の確認項目と実技（docs/06-evaluation.md）", p.Stage))
	return nil
}

func cmdUnit(root string, p *Progress, args []string) error {
	unit := p.Unit
	if len(args) > 0 {
		unit = args[0]
	}
	sh, err := LoadSheet(root, unit)
	if err != nil {
		return err
	}
	if colorOn {
		if pager, err := exec.LookPath("less"); err == nil {
			c := exec.Command(pager, "-R", sh.Path)
			c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
			return c.Run()
		}
	}
	raw, err := os.ReadFile(sh.Path)
	if err != nil {
		return err
	}
	fmt.Print(string(raw))
	return nil
}

func cmdFind(root string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("使い方: offgrid find <語>")
	}
	word := args[0]
	groups := []struct {
		label    string
		patterns []string
		skipTask bool
	}{
		{"カリキュラム", []string{"docs/*.md"}, false},
		{"課題シート", []string{"drills/*/*/TASKS*.md"}, false},
		{"AIへの頼み方", []string{"prompts/*.md"}, false},
		{"自分のメモと振り返り", []string{"drills/*/*/*.md", "logs/*/*.md"}, true},
	}
	total := 0
	for _, g := range groups {
		var hits []string
		for _, pat := range g.patterns {
			paths, _ := filepath.Glob(filepath.Join(root, pat))
			sort.Strings(paths)
			for _, path := range paths {
				name := filepath.Base(path)
				if strings.HasPrefix(name, ".") || (g.skipTask && strings.HasPrefix(name, "TASKS")) || underAnswers(rel(root, path)) {
					continue
				}
				raw, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				for i, line := range strings.Split(string(raw), "\n") {
					if strings.Contains(line, word) {
						text := strings.TrimSpace(line)
						if width(text) > 110 {
							text = pad(text, 110)
						}
						hits = append(hits, fmt.Sprintf("  %s:%d  %s", cyan(rel(root, path)), i+1, text))
					}
				}
			}
		}
		if len(hits) == 0 {
			continue
		}
		rule(fmt.Sprintf("%s（%d件）", g.label, len(hits)))
		for i, h := range hits {
			if i >= 12 {
				fmt.Println(dim(fmt.Sprintf("  ほか%d件", len(hits)-12)))
				break
			}
			fmt.Println(h)
		}
		fmt.Println()
		total += len(hits)
	}
	if total == 0 {
		fmt.Println(dim(fmt.Sprintf("「%s」は見つかりませんでした", word)))
	}
	return nil
}

// checkPlan は、そのユニットで走らせる確認コマンドを決める。
// 戻り値は（コマンドの並び、走らせるものが無いときの説明、実行するディレクトリ）。
func checkPlan(root string, sh *Sheet) ([][]string, string, string) {
	dir := sh.Dir()
	switch sh.Unit[:1] {
	case "F", "O":
		files, _ := filepath.Glob(filepath.Join(dir, "*.sh"))
		sort.Strings(files)
		if len(files) == 0 {
			return nil, "シェルスクリプトがまだない", dir
		}
		if _, err := exec.LookPath("shellcheck"); err != nil {
			return nil, "shellcheck が入っていない（sudo apt install shellcheck）", dir
		}
		cmd := []string{"shellcheck"}
		for _, f := range files {
			cmd = append(cmd, filepath.Base(f))
		}
		return [][]string{cmd}, "", dir
	case "G", "C":
		target := dir
		if sh.Unit[:1] == "C" {
			target = filepath.Join(root, "capstone")
		}
		if _, err := os.Stat(filepath.Join(target, "go.mod")); err != nil {
			return nil, fmt.Sprintf("%s/go.mod がまだない（go mod init から）", rel(root, target)), target
		}
		return [][]string{{"go", "vet", "./..."}, {"go", "test", "-race", "-cover", "./..."}}, "", target
	case "T":
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
			return nil, fmt.Sprintf("%s/package.json がまだない（npm init から）", rel(root, dir)), dir
		}
		return [][]string{{"npx", "tsc", "--noEmit", "--strict"}, {"npx", "vitest", "run"}}, "", dir
	case "D":
		files, _ := filepath.Glob(filepath.Join(dir, "*.sql"))
		sort.Strings(files)
		if len(files) == 0 {
			return nil, ".sql ファイルがまだない", dir
		}
		var b strings.Builder
		b.WriteString("次のコマンドで実行できる（データベース名は課題シートを見る）")
		for _, f := range files {
			fmt.Fprintf(&b, "\ndocker exec -i pg-offgrid psql -U postgres -v ON_ERROR_STOP=1 < %s", rel(root, f))
		}
		return nil, b.String(), dir
	}
	return nil, "確認コマンドは課題シートの「完了条件の確かめ方」を見る", dir
}

// runIn は、確認コマンドを走らせて、出力をそのまま返す（ブラウザにも見せるため）。
// 学習者の go test が止まったとき、ブラウザからは Ctrl+C できないので、上限を置く。
const runLimit = 10 * time.Minute

func runIn(dir string, cmd []string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), runLimit)
	defer cancel()
	c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	if ctx.Err() != nil {
		out = append(out, "\n（10分たったので打ち切りました。止まる原因を、ターミナルで走らせて探す）\n"...)
		return string(out), ctx.Err()
	}
	return string(out), err
}

// stageAll は git add -A のあと、ビルドした実行ファイル（バイナリ）を外す。
// 課題シートは「バイナリはコミットしない」と書いているが、add -A は区別しない。
func stageAll(root string) (excluded []string, err error) {
	if out, err := gitOut(root, "add", "-A"); err != nil {
		return nil, fmt.Errorf("git add: %s", strings.TrimSpace(out))
	}
	// -z で読む。付けないと、日本語などの名前が "\345\256\237..." と引用されて出てきて、
	// そのまま git reset に渡しても外れない（「除外した」と言いながらコミットしていた）。
	out, err := gitOut(root, "diff", "--cached", "--numstat", "-z")
	if err != nil {
		return nil, fmt.Errorf("git diff: %s", strings.TrimSpace(out))
	}
	// -z の並びは「追加\t削除\t」のあとに NUL 区切りの名前が続く
	fields := strings.Split(strings.TrimRight(out, "\x00"), "\x00")
	for i := 0; i < len(fields); i++ {
		head := strings.SplitN(fields[i], "\t", 3)
		if len(head) < 3 {
			continue
		}
		name := head[2]
		if name == "" {
			// 名前が別の欄にある。rename は「旧名 NUL 新名」の2つ続くので、
			// 取り消すのは索引に載っている新しいほう。
			if i+1 >= len(fields) {
				break
			}
			i++
			name = fields[i]
			if i+1 < len(fields) && !strings.Contains(fields[i+1], "\t") && fields[i+1] != "" {
				i++
				name = fields[i]
			}
		}
		f := []string{head[0], head[1], name}
		if f[0] != "-" || f[1] != "-" {
			continue // 追加・削除の行数が出るのはテキスト
		}
		if out, err := gitOut(root, "reset", "-q", "--", f[2]); err != nil {
			return nil, fmt.Errorf("git reset %s: %s", f[2], strings.TrimSpace(out))
		}
		excluded = append(excluded, f[2])
	}
	return excluded, nil
}

func cmdCheck(root string, p *Progress, args []string) error {
	unit := p.Unit
	if len(args) > 0 {
		unit = args[0]
	}
	sh, err := LoadSheet(root, unit)
	if err != nil {
		return err
	}
	cmds, note, dir := checkPlan(root, sh)
	rule(fmt.Sprintf("%s の確認（%s）", sh.Unit, rel(root, dir)))
	if note != "" {
		say(note)
	}
	for _, c := range cmds {
		fmt.Println(dim("  ＋ " + strings.Join(c, " ")))
		run := exec.Command(c[0], c[1:]...)
		run.Dir = dir
		run.Stdout, run.Stderr = os.Stdout, os.Stderr
		if err := run.Run(); err != nil {
			fmt.Println(red("  → 失敗（出力を読んで直す）"))
		} else {
			fmt.Println(green("  → 通りました"))
		}
	}
	fmt.Println()
	rule("残すもの")
	keep := sh.KeepFiles()
	if len(keep) == 0 {
		say("課題シートの「残すもの」を見る")
	}
	for _, f := range keep {
		if sh.HaveKeep(f) {
			fmt.Println("  " + green("[済]") + " " + f)
		} else {
			fmt.Println("  " + yellow("[未]") + " " + f)
		}
	}
	done, total := sh.Counts()
	fmt.Printf("\n  課題 %s %d/%d\n\n", progressBar(done, total), done, total)
	showSections(sh, "完了条件の確かめ方")
	return nil
}

func cmdStuck(root string, s *Session, args []string) error {
	text := strings.Join(args, " ")
	if text == "" {
		got, err := ask("  何が起きた？: ")
		if err != nil {
			return err
		}
		text = got
	}
	if text == "" {
		return nil
	}
	if err := s.EnsureLog(); err != nil {
		return err
	}
	if err := AppendStuck(root, s.Log, text); err != nil {
		return err
	}
	fmt.Println(green("  " + rel(root, s.Log) + " の「詰まったこと」に書きました"))
	say(dim("15分たったら手順に切り替え、45分で打ち切る"))
	return nil
}

func cmdTimer(args []string) error {
	minutes := 15
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			minutes = n
		}
	}
	end := time.Now().Add(time.Duration(minutes) * time.Minute)
	fmt.Printf("%d分を計ります（Ctrl+C で止める）。終わり: %s\n", minutes, end.Format("15:04"))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		left := time.Until(end)
		if left <= 0 {
			break
		}
		fmt.Printf("\r  残り %02d:%02d  ", int(left.Minutes()), int(left.Seconds())%60)
	}
	fmt.Print("\a\n")
	fmt.Println(yellow("  時間です。"))
	say("詰まっているなら、エラー文を全文読む→man→最小の再現→公式ドキュメント、の順に切り替える")
	return nil
}

func cmdRetro(root string, s *Session, args []string) error {
	log := s.Log
	if len(args) > 0 {
		log = filepath.Join(root, args[0])
	} else if err := s.EnsureLog(); err != nil {
		return err
	}
	if _, err := os.Stat(log); err != nil {
		return fmt.Errorf("%s がありません", rel(root, log))
	}
	rule(fmt.Sprintf("振り返り（%s）", rel(root, log)))
	if len(RetroFields(log)) == 0 {
		fmt.Println(green("  全部埋まっています"))
		return nil
	}
	say(dim("1つずつ聞きます。Enterだけで飛ばせます。q でやめます"))
	fmt.Println()
	seen := map[string]bool{}
	for {
		var target *RetroField
		for _, f := range RetroFields(log) {
			if !seen[f.Key] {
				ff := f
				target = &ff
				break
			}
		}
		if target == nil {
			break
		}
		seen[target.Key] = true
		got, err := ask(fmt.Sprintf("  %s\n  > ", target.Label))
		if err != nil {
			return err
		}
		if strings.EqualFold(got, "q") {
			break
		}
		if got != "" {
			if err := RetroWrite(root, log, target.Key, got); err != nil {
				return err
			}
		}
	}
	left := len(RetroFields(log))
	fmt.Println()
	if left == 0 {
		fmt.Println(green("  全部埋まりました"))
	} else {
		fmt.Println(yellow(fmt.Sprintf("  まだ%d個 空いています（『offgrid retro』でいつでも続けられます）", left)))
	}
	return nil
}

func cmdDone(root string, p *Progress, args []string) error {
	unit := p.Unit
	if len(args) > 0 {
		unit = args[0]
	}
	unit = strings.ToUpper(unit)
	if p.IsDone(unit) {
		fmt.Printf("%s は、すでに完了になっています\n", unit)
		return nil
	}
	if sh, err := LoadSheet(root, unit); err == nil {
		showSections(sh, "完了条件の確かめ方")
		done, total := sh.Counts()
		if total > 0 && done < total {
			fmt.Println(yellow(fmt.Sprintf("  課題が %d/%d です", done, total)))
		}
		if miss := sh.MissingFiles(); len(miss) > 0 {
			fmt.Println(yellow("  残すものが、まだ無い（または空）: " + strings.Join(miss, "、")))
		}
	}
	if !confirm("\n  上を満たしましたか？ 記録する場合は y [y/N]: ") {
		fmt.Println("  記録しませんでした")
		return nil
	}
	ok, err := p.MarkDone(unit)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("PROGRESS.md に「- [ ] %s ...」の行がありません", unit)
	}
	fmt.Println(green(fmt.Sprintf("  PROGRESS.md に記録しました（%s）", time.Now().Format("2006-01-02"))))
	next, err := ask("  次のユニットのID（Enterで変えない）: ")
	if err != nil || next == "" {
		return nil
	}
	next = strings.ToUpper(next)
	if err := p.SetNext(next, strings.HasPrefix(next, "D")); err != nil {
		return err
	}
	fmt.Println(green(fmt.Sprintf("  「次のユニット」を %s にしました", next)))
	return nil
}

func cmdEnd(root string, p *Progress, s *Session) error {
	if err := s.EnsureLog(); err != nil {
		return err
	}
	rule("今日の振り返り")
	holes := RetroFields(s.Log)
	if len(holes) > 0 {
		fmt.Println(yellow(fmt.Sprintf("  書いていない欄が%d個あります", len(holes))))
		for i, h := range holes {
			if i >= 8 {
				break
			}
			fmt.Println("    - " + h.Label)
		}
		if confirm("\n  いま埋めますか？ [y/N]: ") {
			if err := cmdRetro(root, s, nil); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(green("  全部埋まっています"))
	}

	if sh, err := LoadSheet(root, p.Unit); err == nil {
		fmt.Println()
		rule(p.Unit + " の状態")
		done, total := sh.Counts()
		fmt.Printf("  課題 %s %d/%d\n", progressBar(done, total), done, total)
		miss := sh.MissingFiles()
		if len(miss) > 0 {
			fmt.Println(yellow("  残すものが足りない: " + strings.Join(miss, "、")))
		} else {
			fmt.Println(green("  残すもの: 揃っています"))
		}
		if total > 0 && done == total && len(miss) == 0 && !p.IsDone(p.Unit) {
			if confirm(fmt.Sprintf("\n  %s を完了として記録しますか？ [y/N]: ", p.Unit)) {
				if err := cmdDone(root, p, []string{p.Unit}); err != nil {
					return err
				}
			}
		}
	}

	fmt.Println()
	rule("コミット")
	dirty, err := gitOut(root, "status", "--porcelain")
	if err != nil {
		say(dirty)
		return nil
	}
	if strings.TrimSpace(dirty) == "" {
		fmt.Println(green("  変更はありません（すべてコミット済み）"))
	} else {
		for i, line := range strings.Split(strings.TrimRight(dirty, "\n"), "\n") {
			if i >= 20 {
				break
			}
			fmt.Println("  " + line)
		}
		head := fmt.Sprintf("[no-ai] %s: ", p.Unit)
		extra, err := ask(fmt.Sprintf("\n  コミットメッセージ（「%s」に続ける。Enterで手で打つ）: ", head))
		if err != nil {
			return err
		}
		extra = strings.TrimSpace(extra)
		if extra != "" && !confirm(fmt.Sprintf("  「%s%s」でコミットしますか？ [y/N]: ", head, extra)) {
			extra = ""
		}
		if extra != "" {
			excluded, err := stageAll(root)
			if err != nil {
				return err
			}
			for _, f := range excluded {
				fmt.Println(yellow("  実行ファイルらしいので、コミットから外しました: " + f))
			}
			if out, err := gitOut(root, "commit", "-m", head+extra); err != nil {
				fmt.Println(red("  失敗: " + out))
			} else {
				fmt.Println(green("  コミットしました"))
				if confirm("  push しますか？ [y/N]: ") {
					if out, err := gitOut(root, "push", "origin", "HEAD"); err != nil {
						fmt.Println(red("  失敗: " + out))
					} else {
						fmt.Println(green("  push しました"))
					}
				}
			}
		} else {
			fmt.Println(dim("  git add -A"))
			fmt.Println(dim(fmt.Sprintf("  git commit -m \"%s<やったこと>\"", head)))
			fmt.Println(dim("  git push origin main"))
		}
	}
	if ahead, err := gitOut(root, "log", "--oneline", "@{u}.."); err == nil && strings.TrimSpace(ahead) != "" {
		n := len(strings.Split(strings.TrimSpace(ahead), "\n"))
		fmt.Println(yellow(fmt.Sprintf("  push していないコミットが%d件あります", n)))
	}
	return nil
}

func gitOut(root string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = root
	out, err := c.CombinedOutput()
	return string(out), err
}
