package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// セッションの案内（run）と、終わりの手続きの git の部分の検査。
// どちらも学習者のファイルと履歴に触るのに、テストが1つも当たっていなかった。

// 枠の見出し（「── 2/4. メイン（F3）（2時間） ───」）から、枠の名前を取り出す
var stepHeadRe = regexp.MustCompile(`── \d+/\d+\. (.+?) ─`)

// initGit は、検査用リポジトリを git の管理下に置く（cmdEnd の git の経路のため）。
func initGit(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "t@example.test"},
		{"config", "user.name", "t"},
		{"add", "-A"},
		{"commit", "-qm", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func gitSubjects(t *testing.T, root string) string {
	t.Helper()
	out, err := gitOut(root, "log", "--format=%s")
	if err != nil {
		t.Fatalf("git log: %v\n%s", err, out)
	}
	return out
}

// cmdRun が、枠を上から順に案内して、進み具合を残すこと。
func TestCmdRunWalksEveryStepAndRemembers(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)

	// ウォームアップ=Enter、メイン=d（枠を終える）、振り返り=q（やめる）、
	// 終わり=n（振り返りは埋めない。git が無いので、そこで止まる）
	withInput(t, "", "d", "q", "n")
	out := capture(t, func() {
		if err := cmdRun(root, p, s); err != nil {
			t.Errorf("cmdRun: %v", err)
		}
	})
	if !strings.Contains(out, "今日のぶんは、全部終わりました") {
		t.Errorf("最後まで進んでいない:\n%s", out)
	}
	// Stage 1 の第1回は短縮なので、DBと午後の枠は出ない。
	// 進み具合の表示にも「PostgreSQL」が出るので、枠の見出しだけを見る
	var heads []string
	for _, line := range strings.Split(out, "\n") {
		if m := stepHeadRe.FindStringSubmatch(line); m != nil {
			heads = append(heads, m[1])
		}
	}
	if len(heads) != 4 {
		t.Errorf("枠が %d 個（短縮の回は4個）: %v", len(heads), heads)
	}
	for _, h := range heads {
		for _, ng := range []string{"PostgreSQL", "コードリーディング", "デバッグドリル"} {
			if strings.Contains(h, ng) {
				t.Errorf("短縮の回に「%s」の枠が出ている: %v", ng, heads)
			}
		}
	}
	// 進み具合が残り、もう一度呼んでも枠をやり直さない
	state := LoadState(root, s)
	for _, id := range []StepID{StepWarmup, StepMain, StepRetro, StepEnd} {
		if !state.IsDone(id) {
			t.Errorf("%s が済みになっていない", id)
		}
	}
	withInput(t) // 何も聞かれないはず
	again := capture(t, func() {
		if err := cmdRun(root, p, s); err != nil {
			t.Errorf("2回目の cmdRun: %v", err)
		}
	})
	if !strings.Contains(again, "全部終わりました") {
		t.Errorf("2回目に、枠をやり直している:\n%s", again)
	}
}

// 「q=中断」と表示しているなら、q で中断すること。
// 答えを捨てていたため、q を押すとその枠が「済み」になり、二度と出てこなかった。
func TestQuitActuallyAborts(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)

	withInput(t, "q")
	var err error
	capture(t, func() { err = cmdRun(root, p, s) })
	if !errors.As(err, &errAbort{}) {
		t.Fatalf("q で中断していない（err=%v）", err)
	}
	// 中断した枠は、済みになっていない
	if LoadState(root, s).IsDone(StepWarmup) {
		t.Error("中断したのに、ウォームアップが済みになっている")
	}
	// 続きから再開できる
	withInput(t, "", "d", "q", "n")
	out := capture(t, func() {
		if err := cmdRun(root, p, s); err != nil {
			t.Errorf("再開できない: %v", err)
		}
	})
	if !strings.Contains(out, "ウォームアップ") {
		t.Errorf("中断した枠から再開していない:\n%s", out)
	}
}

// 課題を1問ずつ出す枠の、それぞれの答えの扱い。
func TestGuideUnitEachAnswer(t *testing.T) {
	root := newRepo(t)
	s, _ := LoadSession(root)
	if err := s.EnsureLog(); err != nil {
		t.Fatal(err)
	}
	sh, _ := LoadSheet(root, "F3")
	state := LoadState(root, s)

	// k=手がかり（進まない）→ s=詰まった（メモを書いて飛ばす）→ d=枠を終える
	withInput(t, "k", "s", "uniq の使い方が分からない", "d")
	out := capture(t, func() {
		if err := guideUnit(sh, s, state); err != nil {
			t.Errorf("guideUnit: %v", err)
		}
	})
	// 手がかりは、答えではなくキーワードと詰まりやすいところ
	if !strings.Contains(out, "キーワード") || !strings.Contains(out, "詰まりやすいところ") {
		t.Errorf("手がかりが出ていない:\n%s", out)
	}
	// 詰まりメモが、振り返りの正しい節に入る
	body := read(t, s.Log)
	at := strings.Index(body, "## 詰まったこと")
	next := strings.Index(body[at:], "## 今日わかったこと") + at
	if !strings.Contains(body[at:next], "F3 課題1: uniq の使い方が分からない") {
		t.Errorf("詰まりメモが入っていない:\n%s", body[at:next])
	}
	// 詰まった課題は、できたことにしない
	if done, _ := sh.Counts(); done != 0 {
		t.Errorf("詰まったのに、できた数が %d", done)
	}
	if !LoadState(root, s).isSkipped("F3", 1) {
		t.Error("詰まった課題が、あとで回されていない")
	}
}

func TestGuideUnitAllDone(t *testing.T) {
	root := newRepo(t)
	s, _ := LoadSession(root)
	sh, _ := LoadSheet(root, "F3")
	for _, n := range sh.TaskNums() {
		if err := sh.Tick(n, true); err != nil {
			t.Fatal(err)
		}
	}
	withInput(t) // 何も聞かれない
	out := capture(t, func() {
		if err := guideUnit(sh, s, LoadState(root, s)); err != nil {
			t.Errorf("guideUnit: %v", err)
		}
	})
	if !strings.Contains(out, "全部できました") {
		t.Errorf("全部できたことを言っていない:\n%s", out)
	}
	// 端末の幅で折り返されるので、空白をつぶして比べる
	flat := strings.Join(strings.Fields(out), " ")
	if !strings.Contains(flat, "offgrid check F3") {
		t.Errorf("次にやることを示していない:\n%s", out)
	}
}

// 終わりの手続きの git の部分。学習者の履歴に触るので、いちばん慎重に見る。
func TestCmdEndCommitsWithTheConvention(t *testing.T) {
	root := newRepo(t)
	initGit(t, root)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	// コミットしていない変更を作る
	if err := os.WriteFile(filepath.Join(root, "drills", "foundation", "F03-pipeline", "notes.md"),
		[]byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 振り返りは埋めない(n) → メッセージ → 確認(y) → push しない(n)
	withInput(t, "n", "xxdでUTF-8のバイト列を確認", "y", "n")
	out := capture(t, func() {
		if err := cmdEnd(root, p, s); err != nil {
			t.Errorf("cmdEnd: %v", err)
		}
	})
	if !strings.Contains(out, "コミットしました") {
		t.Errorf("コミットしたと言っていない:\n%s", out)
	}
	subjects := gitSubjects(t, root)
	want := "[no-ai] F3: xxdでUTF-8のバイト列を確認"
	if !strings.Contains(subjects, want) {
		t.Errorf("規約どおりのコミットが無い（want %q）:\n%s", want, subjects)
	}
	if dirty, _ := gitOut(root, "status", "--porcelain"); strings.TrimSpace(dirty) != "" {
		t.Errorf("コミットしたのに、変更が残っている:\n%s", dirty)
	}
}

// メッセージを空で返したら、コミットせず、打つべきコマンドを見せること。
// F6 で Git を学ぶまでは、自分で打つことを勧めているため。
func TestCmdEndShowsTheCommandsWhenNoMessage(t *testing.T) {
	root := newRepo(t)
	initGit(t, root)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	if err := os.WriteFile(filepath.Join(root, "drills", "foundation", "F03-pipeline", "notes.md"),
		[]byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := gitSubjects(t, root)

	withInput(t, "n", "") // 振り返りは埋めない、メッセージは空
	out := capture(t, func() {
		if err := cmdEnd(root, p, s); err != nil {
			t.Errorf("cmdEnd: %v", err)
		}
	})
	if gitSubjects(t, root) != before {
		t.Error("メッセージが空なのにコミットしている")
	}
	for _, want := range []string{"git add -A", "git commit -m", "git push origin main"} {
		if !strings.Contains(out, want) {
			t.Errorf("打つべきコマンド %q を見せていない:\n%s", want, out)
		}
	}
}

// 確認で n と答えたら、コミットしないこと。
func TestCmdEndRespectsDeclinedConfirmation(t *testing.T) {
	root := newRepo(t)
	initGit(t, root)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	if err := os.WriteFile(filepath.Join(root, "drills", "foundation", "F03-pipeline", "notes.md"),
		[]byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := gitSubjects(t, root)

	withInput(t, "n", "まちがえた", "n") // 確認で n
	capture(t, func() {
		if err := cmdEnd(root, p, s); err != nil {
			t.Errorf("cmdEnd: %v", err)
		}
	})
	if gitSubjects(t, root) != before {
		t.Error("確認で n と答えたのにコミットしている")
	}
}

// トラックごとに、走らせる確認コマンドが正しいこと。
func TestCheckPlanPerTrack(t *testing.T) {
	root := newRepo(t)
	sh, _ := LoadSheet(root, "F3") // F トラック＝シェル
	cmds, note, dir := checkPlan(root, sh)
	if len(cmds) != 0 || !strings.Contains(note, "シェルスクリプト") {
		t.Errorf("F: .sh が無いのに走らせようとしている（cmds=%v note=%q）", cmds, note)
	}
	if err := os.WriteFile(filepath.Join(dir, "x.sh"), []byte("#!/bin/sh\necho x\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	cmds, note, _ = checkPlan(root, sh)
	if _, err := exec.LookPath("shellcheck"); err == nil {
		if len(cmds) != 1 || cmds[0][0] != "shellcheck" {
			t.Errorf("F: shellcheck を走らせていない: %v", cmds)
		}
	} else if !strings.Contains(note, "shellcheck") {
		t.Errorf("F: shellcheck が無いことを伝えていない: %q", note)
	}

	// D トラック＝SQL。実行はせず、打つコマンドを見せる
	db, _ := LoadSheet(root, "D1")
	cmds, note, dir = checkPlan(root, db)
	if len(cmds) != 0 || !strings.Contains(note, ".sql") {
		t.Errorf("D: .sql が無いのに走らせようとしている（cmds=%v note=%q）", cmds, note)
	}
	if err := os.WriteFile(filepath.Join(dir, "q.sql"), []byte("select 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmds, note, _ = checkPlan(root, db)
	if len(cmds) != 0 {
		t.Errorf("D: SQL を勝手に実行しようとしている: %v", cmds)
	}
	if !strings.Contains(note, "pg-offgrid") || !strings.Contains(note, "q.sql") {
		t.Errorf("D: 実行するコマンドを見せていない: %q", note)
	}
}
