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

// mkSheet は、検査用に別トラックの課題シートを足す（checkPlan のトラック別の分岐のため）。
func mkSheet(t *testing.T, root, track, dirName, unit, theme string) string {
	t.Helper()
	dir := filepath.Join(root, "drills", track, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := strings.Replace(sampleSheet, `title: "F3 テキスト処理"`,
		`title: "`+unit+" "+theme+`"`, 1)
	if err := os.WriteFile(filepath.Join(dir, "TASKS.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// pastLogs は、済んだセッションのログを作る（第N回にするため）。
func pastLogs(t *testing.T, root string, n int) {
	t.Helper()
	dir := filepath.Join(root, "logs", "2026")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		name := filepath.Join(dir, "01-"+string(rune('0'+(i+1)/10))+string(rune('0'+(i+1)%10))+".md")
		if err := os.WriteFile(name, []byte("# 過去の回\n\n## やったこと\n- メインドリル（ユニット）：やった\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// フルの回（7回目以降）で、PostgreSQL と午後の枠が案内されること。
func TestCmdRunFullDayIncludesDBAndAfternoon(t *testing.T) {
	root := newRepo(t)
	pastLogs(t, root, 6) // これで第7回になる
	p, err := LoadProgress(root)
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.Number != 7 {
		t.Fatalf("第%d回（7回目のはず）", s.Number)
	}
	if p.DBUnit != "D1" {
		t.Fatalf("DBユニット = %q", p.DBUnit)
	}

	// ウォームアップ、メイン=d、PostgreSQL=d、午後=Enter、振り返り=q、終わり=n
	withInput(t, "", "d", "d", "", "q", "n")
	out := capture(t, func() {
		if err := cmdRun(root, p, s); err != nil {
			t.Errorf("cmdRun: %v", err)
		}
	})
	var heads []string
	for _, line := range strings.Split(out, "\n") {
		if m := stepHeadRe.FindStringSubmatch(line); m != nil {
			heads = append(heads, m[1])
		}
	}
	if len(heads) != 6 {
		t.Fatalf("枠が %d 個（フルの回は6個）: %v", len(heads), heads)
	}
	joined := strings.Join(heads, " / ")
	for _, want := range []string{"ウォームアップ", "メイン（F3）", "PostgreSQL（D1）", "コードリーディング", "振り返り", "終わりの手続き"} {
		if !strings.Contains(joined, want) {
			t.Errorf("「%s」の枠が無い: %s", want, joined)
		}
	}
	// DBの枠でも課題が案内され、state に残ること
	if !LoadState(root, s).IsDone(StepDB) {
		t.Error("PostgreSQL の枠が済みになっていない")
	}
}

// DBユニットが未設定のときは、書き換え方を案内して先へ進むこと。
func TestCmdRunTellsHowToSetTheDBUnit(t *testing.T) {
	root := newRepo(t)
	pastLogs(t, root, 6)
	// 「次のPostgreSQLユニット」を、テンプレートの初期値（ユニットIDでない）に戻す
	raw := read(t, filepath.Join(root, "PROGRESS.md"))
	raw = strings.Replace(raw, "- 次のPostgreSQLユニット：D1",
		"- 次のPostgreSQLユニット：（Stage 1に入ったら D1 と書く）", 1)
	if err := os.WriteFile(filepath.Join(root, "PROGRESS.md"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	if p.DBUnit != "" {
		t.Fatalf("初期値を読んでしまっている: %q", p.DBUnit)
	}

	withInput(t, "", "d", "", "", "q", "n") // DBの枠は Enter で通過
	out := capture(t, func() {
		if err := cmdRun(root, p, s); err != nil {
			t.Errorf("cmdRun: %v", err)
		}
	})
	flat := strings.Join(strings.Fields(out), " ")
	if !strings.Contains(flat, "次のPostgreSQLユニット") || !strings.Contains(flat, "書き換える") {
		t.Errorf("書き換え方を案内していない:\n%s", out)
	}
}

// push の経路。remote を手元に作って、実際に届くことを確かめる。
func TestCmdEndPushesToOrigin(t *testing.T) {
	root := newRepo(t)
	initGit(t, root)
	bare := filepath.Join(t.TempDir(), "origin.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", bare).CombinedOutput(); err != nil {
		t.Fatalf("bare の作成: %v\n%s", err, out)
	}
	for _, args := range [][]string{
		{"remote", "add", "origin", bare},
		{"push", "-q", "-u", "origin", "HEAD"},
	} {
		if out, err := gitOut(root, args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	if err := os.WriteFile(filepath.Join(root, "drills", "foundation", "F03-pipeline", "notes.md"),
		[]byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 振り返りは埋めない(n) → メッセージ → 確認(y) → push する(y)
	withInput(t, "n", "パイプラインを試した", "y", "y")
	out := capture(t, func() {
		if err := cmdEnd(root, p, s); err != nil {
			t.Errorf("cmdEnd: %v", err)
		}
	})
	if !strings.Contains(out, "push しました") {
		t.Errorf("push したと言っていない:\n%s", out)
	}
	// remote 側に届いていること
	got, err := exec.Command("git", "--git-dir", bare, "log", "--format=%s", "-1").CombinedOutput()
	if err != nil {
		t.Fatalf("bare の log: %v\n%s", err, got)
	}
	if !strings.Contains(string(got), "[no-ai] F3: パイプラインを試した") {
		t.Errorf("remote に届いていない: %s", got)
	}
	// push 済みなので、未 push の警告は出ない
	if strings.Contains(out, "push していないコミットが") {
		t.Errorf("push したのに、未pushの警告が出ている:\n%s", out)
	}
}

// コミットしたが push しなかったときは、残っていることを知らせること。
func TestCmdEndWarnsAboutUnpushedCommits(t *testing.T) {
	root := newRepo(t)
	initGit(t, root)
	bare := filepath.Join(t.TempDir(), "origin.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", bare).CombinedOutput(); err != nil {
		t.Fatalf("bare の作成: %v\n%s", err, out)
	}
	for _, args := range [][]string{{"remote", "add", "origin", bare}, {"push", "-q", "-u", "origin", "HEAD"}} {
		if out, err := gitOut(root, args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	if err := os.WriteFile(filepath.Join(root, "drills", "foundation", "F03-pipeline", "notes.md"),
		[]byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	withInput(t, "n", "メモを書いた", "y", "n") // push は断る
	out := capture(t, func() {
		if err := cmdEnd(root, p, s); err != nil {
			t.Errorf("cmdEnd: %v", err)
		}
	})
	if !strings.Contains(out, "push していないコミットが1件あります") {
		t.Errorf("未pushを知らせていない:\n%s", out)
	}
}

// G・C・T トラックの確認コマンド。
func TestCheckPlanGoAndTypeScript(t *testing.T) {
	root := newRepo(t)

	// G：課題ディレクトリに go.mod があるときだけ走らせる
	gdir := mkSheet(t, root, "go", "G01-basics", "G1", "Goの文法")
	gsh, err := LoadSheet(root, "G1")
	if err != nil {
		t.Fatal(err)
	}
	if cmds, note, _ := checkPlan(root, gsh); len(cmds) != 0 || !strings.Contains(note, "go.mod") {
		t.Errorf("G: go.mod が無いのに走らせようとしている（cmds=%v note=%q）", cmds, note)
	}
	if err := os.WriteFile(filepath.Join(gdir, "go.mod"), []byte("module x\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmds, _, dir := checkPlan(root, gsh)
	if len(cmds) != 2 || cmds[0][1] != "vet" || cmds[1][1] != "test" {
		t.Errorf("G: go vet と go test を走らせていない: %v", cmds)
	}
	if dir != gdir {
		t.Errorf("G: 走らせる場所が違う: %s", dir)
	}

	// C：卒業制作は capstone/ を見る（課題ディレクトリではない）
	mkSheet(t, root, "capstone", "C01-requirements", "C1", "要件定義")
	csh, err := LoadSheet(root, "C1")
	if err != nil {
		t.Fatal(err)
	}
	capstone := filepath.Join(root, "capstone")
	if err := os.MkdirAll(capstone, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(capstone, "go.mod"), []byte("module w\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, dir := checkPlan(root, csh); dir != capstone {
		t.Errorf("C: capstone/ ではなく %s を見ている", dir)
	}

	// T：package.json があるときだけ走らせる
	tdir := mkSheet(t, root, "ts", "T01-html-css", "T1", "HTMLとCSS")
	tsh, err := LoadSheet(root, "T1")
	if err != nil {
		t.Fatal(err)
	}
	if cmds, note, _ := checkPlan(root, tsh); len(cmds) != 0 || !strings.Contains(note, "package.json") {
		t.Errorf("T: package.json が無いのに走らせようとしている（cmds=%v note=%q）", cmds, note)
	}
	if err := os.WriteFile(filepath.Join(tdir, "package.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmds, _, _ = checkPlan(root, tsh)
	if len(cmds) != 2 || cmds[0][1] != "tsc" || cmds[1][1] != "vitest" {
		t.Errorf("T: tsc と vitest を走らせていない: %v", cmds)
	}
}

// runIn が、出力と失敗をそのまま返すこと（画面にそのまま見せるため）。
func TestRunInReturnsOutputAndFailure(t *testing.T) {
	dir := t.TempDir()
	out, err := runIn(dir, []string{"sh", "-c", "echo ひとつめ; exit 0"})
	if err != nil || !strings.Contains(out, "ひとつめ") {
		t.Errorf("成功した出力を返していない（out=%q err=%v）", out, err)
	}
	out, err = runIn(dir, []string{"sh", "-c", "echo こわれた >&2; exit 3"})
	if err == nil {
		t.Error("失敗を返していない")
	}
	if !strings.Contains(out, "こわれた") {
		t.Errorf("標準エラーの出力を捨てている: %q", out)
	}
}
