package main

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ブラウザの案内（/session と /end）を、ボタンを押す順にたどる。
// CSRF の穴があったのはこのハンドラ群で、関門を通ったあとの中身は未検証だった。

var currentTaskRe = regexp.MustCompile(`<span class="badge">課題 (\d+)</span>`)

func currentTask(t *testing.T, body string) int {
	t.Helper()
	m := currentTaskRe.FindStringSubmatch(body)
	if m == nil {
		return 0
	}
	n := 0
	for _, c := range m[1] {
		n = n*10 + int(c-'0')
	}
	return n
}

func TestSessionButtonsAdvanceTheTask(t *testing.T) {
	_, h, root := newServer(t)
	sess, _ := LoadSession(root)

	// 案内を開くと、ウォームアップ（課題なし）から始まる
	w := get(t, h, "/session")
	if !strings.Contains(w.Body.String(), "ウォームアップ") || currentTask(t, w.Body.String()) != 0 {
		t.Fatalf("最初がウォームアップでない:\n%s", w.Body.String()[:400])
	}
	// 「終わった」でメインへ
	post(t, h, "/session/step", url.Values{"step": {"warmup"}, "action": {"finish"}})
	w = get(t, h, "/session")
	if currentTask(t, w.Body.String()) != 1 {
		t.Fatalf("メインの課題1が出ていない（%d）", currentTask(t, w.Body.String()))
	}

	// できた → 2へ。work.md にも残る
	post(t, h, "/session/task", url.Values{"unit": {"F3"}, "n": {"1"}, "action": {"done"}})
	w = get(t, h, "/session")
	if currentTask(t, w.Body.String()) != 2 {
		t.Errorf("「できた」で進んでいない（%d）", currentTask(t, w.Body.String()))
	}
	sh, _ := LoadSheet(root, "F3")
	if !strings.Contains(read(t, sh.WorkPath()), "- [x] 1.") {
		t.Error("「できた」が work.md に残っていない")
	}

	// あとで → 3へ。2 は後ろに回る
	post(t, h, "/session/task", url.Values{"unit": {"F3"}, "n": {"2"}, "action": {"later"}})
	w = get(t, h, "/session")
	if currentTask(t, w.Body.String()) != 3 {
		t.Errorf("「あとで」で次に進んでいない（%d）", currentTask(t, w.Body.String()))
	}

	// 詰まった → メモが振り返りに入り、3 も後ろに回る
	post(t, h, "/session/task", url.Values{"unit": {"F3"}, "n": {"3"}, "action": {"stuck"}, "memo": {"uniq -c の数が合わない"}})
	body := read(t, sess.Log)
	if !strings.Contains(body, "F3 課題3: uniq -c の数が合わない") {
		t.Errorf("詰まりメモが入っていない:\n%s", body)
	}
	w = get(t, h, "/session")
	if currentTask(t, w.Body.String()) != 4 {
		t.Errorf("「詰まった」で次に進んでいない（%d）", currentTask(t, w.Body.String()))
	}

	// 手がかり → 答えではなく、キーワードと詰まりやすいところ
	w = post(t, h, "/session/task", url.Values{"unit": {"F3"}, "n": {"4"}, "action": {"hint"}})
	if loc := w.Header().Get("Location"); !strings.Contains(loc, "hint=1") {
		t.Errorf("手がかりの画面へ行っていない: %q", loc)
	}
	w = get(t, h, "/session?hint=1")
	if !strings.Contains(w.Body.String(), "詰まりやすいところ") {
		t.Errorf("手がかりが出ていない")
	}

	// 4, 5 をできたにすると、あとで回した 2, 3 に戻ってくる
	post(t, h, "/session/task", url.Values{"unit": {"F3"}, "n": {"4"}, "action": {"done"}})
	post(t, h, "/session/task", url.Values{"unit": {"F3"}, "n": {"5"}, "action": {"done"}})
	w = get(t, h, "/session")
	if currentTask(t, w.Body.String()) != 2 || !strings.Contains(w.Body.String(), "あとで回した課題に戻ってきました") {
		t.Errorf("あとで回した課題に戻っていない（%d）", currentTask(t, w.Body.String()))
	}

	// この枠を終える → 残り（2, 3）は次回へ。終わりの手続きで知らせる
	post(t, h, "/session/task", url.Values{"unit": {"F3"}, "n": {"2"}, "action": {"finish-step"}})
	if !LoadState(root, sess).IsDone(StepMain) {
		t.Error("枠が済みになっていない")
	}
	w = get(t, h, "/end")
	if !strings.Contains(w.Body.String(), "あとで回した課題が残っています") {
		t.Errorf("残った課題を知らせていない")
	}
	flat := strings.Join(strings.Fields(w.Body.String()), " ")
	if !strings.Contains(flat, "課題 2、3 が、まだできていません") {
		t.Errorf("どの課題が残っているかを言っていない:\n%s", flat[:min(len(flat), 800)])
	}
}

// 「前の枠に戻る」で、済みにした枠をやり直せること。
func TestReopenAStep(t *testing.T) {
	_, h, root := newServer(t)
	sess, _ := LoadSession(root)
	get(t, h, "/session")
	post(t, h, "/session/step", url.Values{"step": {"warmup"}, "action": {"finish"}})
	if !LoadState(root, sess).IsDone(StepWarmup) {
		t.Fatal("済みになっていない")
	}
	post(t, h, "/session/step", url.Values{"step": {"warmup"}, "action": {"reopen"}})
	if LoadState(root, sess).IsDone(StepWarmup) {
		t.Error("戻したのに済みのまま")
	}
	w := get(t, h, "/session")
	if !strings.Contains(w.Body.String(), "ウォームアップ") || currentTask(t, w.Body.String()) != 0 {
		t.Error("戻した枠が出ていない")
	}
}

// 全部の枠を終えると、終わったことを出すこと。
func TestSessionFinished(t *testing.T) {
	_, h, root := newServer(t)
	p, _ := LoadProgress(root)
	sess, _ := LoadSession(root)
	get(t, h, "/session")
	for _, st := range sessionSteps(p, sess) {
		post(t, h, "/session/step", url.Values{"step": {string(st.ID)}, "action": {"finish"}})
	}
	w := get(t, h, "/session")
	if !strings.Contains(w.Body.String(), "今日のぶんは、全部終わりました") {
		t.Error("終わったことを出していない")
	}
}

// 終わりの手続きで、完了の記録とコミットができること。
func TestEndDoneAndCommit(t *testing.T) {
	_, h, root := newServer(t)
	initGit(t, root)
	sh, _ := LoadSheet(root, "F3")
	for _, n := range sh.TaskNums() {
		if err := sh.Tick(n, true); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range sh.KeepFiles() {
		if err := os.WriteFile(filepath.Join(sh.Dir(), f), []byte("書いた"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// 揃ったので「完了にする」が出る
	w := get(t, h, "/end")
	if !strings.Contains(w.Body.String(), "F3 を完了にする") {
		t.Fatal("完了にするボタンが出ていない")
	}
	w = post(t, h, "/end/done", url.Values{"unit": {"F3"}, "next": {"F4"}})
	if w.Code != 303 {
		t.Fatalf("/end/done = %d\n%s", w.Code, w.Body.String())
	}
	p, _ := LoadProgress(root)
	if !p.IsDone("F3") || p.Unit != "F4" {
		t.Errorf("記録されていない（done=%v next=%s）", p.IsDone("F3"), p.Unit)
	}
	// 無いユニットは、画面で断る
	w = post(t, h, "/end/done", url.Values{"unit": {"Z9"}})
	if !strings.Contains(w.Body.String(), "記録できません") {
		t.Errorf("無いユニットを黙って受けている（%d）", w.Code)
	}

	// コミット。メッセージが空なら断る
	w = post(t, h, "/end/commit", url.Values{"action": {"commit"}, "message": {"  "}})
	if !strings.Contains(w.Body.String(), "コミットできません") {
		t.Errorf("空のメッセージでコミットしようとしている（%d）", w.Code)
	}
	// 画面のフォームは、やっていたユニット（F3）を持っている。「次」を F4 にしたあとでも F3 で付く
	w = post(t, h, "/end/commit", url.Values{"action": {"commit"}, "message": {"パイプラインを試した"}, "unit": {"F3"}})
	if w.Code != 303 {
		t.Fatalf("/end/commit = %d\n%s", w.Code, w.Body.String())
	}
	if !strings.Contains(gitSubjects(t, root), "[no-ai] F3: パイプラインを試した") {
		t.Errorf("規約どおりに付いていない:\n%s", gitSubjects(t, root))
	}
	if dirty, _ := gitOut(root, "status", "--porcelain"); strings.TrimSpace(dirty) != "" {
		t.Errorf("コミット後に変更が残っている:\n%s", dirty)
	}
}

// 「コミットして push」が、remote に届くこと。
func TestEndCommitAndPush(t *testing.T) {
	_, h, root := newServer(t)
	initGit(t, root)
	bare := filepath.Join(t.TempDir(), "origin.git")
	if out, err := exec.Command("git", "init", "-q", "--bare", bare).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, args := range [][]string{{"remote", "add", "origin", bare}, {"push", "-q", "-u", "origin", "HEAD"}} {
		if out, err := gitOut(root, args...); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	sh, _ := LoadSheet(root, "F3")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "notes.md"), []byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := post(t, h, "/end/commit", url.Values{"action": {"commit-push"}, "message": {"メモを書いた"}})
	if w.Code != 303 {
		t.Fatalf("commit-push = %d\n%s", w.Code, w.Body.String())
	}
	got, _ := exec.Command("git", "--git-dir", bare, "log", "--format=%s", "-1").CombinedOutput()
	if !strings.Contains(string(got), "[no-ai] F3: メモを書いた") {
		t.Errorf("remote に届いていない: %s", got)
	}
	// 届いているので、終わりの画面に未 push の警告は出ない
	w = get(t, h, "/end")
	if strings.Contains(w.Body.String(), "push していないコミット") {
		t.Error("push したのに、未 push の警告が出ている")
	}
}

// 確認コマンドを走らせる画面。F で .sh が無ければ、走らせずにそう言う。
func TestEndCheckRunsThePlan(t *testing.T) {
	_, h, root := newServer(t)
	w := post(t, h, "/end/check", nil)
	if w.Code != 200 {
		t.Fatalf("/end/check = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "シェルスクリプトがまだない") {
		t.Errorf("走らせるものが無いことを言っていない")
	}
	// D は SQL を勝手に実行せず、打つコマンドを見せる
	raw := read(t, filepath.Join(root, "PROGRESS.md"))
	raw = strings.Replace(raw, "- 次のユニット：F3", "- 次のユニット：D1", 1)
	if err := os.WriteFile(filepath.Join(root, "PROGRESS.md"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	db, _ := LoadSheet(root, "D1")
	if err := os.WriteFile(filepath.Join(db.Dir(), "q.sql"), []byte("select 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	w = post(t, h, "/end/check", nil)
	if !strings.Contains(w.Body.String(), "pg-offgrid") || !strings.Contains(w.Body.String(), "q.sql") {
		t.Errorf("D の実行コマンドを見せていない")
	}
}

// オフライン検索と使い方の画面。
func TestFindAndHelpPages(t *testing.T) {
	_, h, _ := newServer(t)
	w := get(t, h, "/find")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "言葉を入れて検索") {
		t.Errorf("空の検索画面: %d", w.Code)
	}
	w = get(t, h, "/find?q="+url.QueryEscape("パイプ"))
	if !strings.Contains(w.Body.String(), "<mark>パイプ</mark>") || !strings.Contains(w.Body.String(), "/unit/F3") {
		t.Errorf("検索の結果が出ていない、または課題シートへのリンクが無い")
	}
	w = get(t, h, "/find?q="+url.QueryEscape("存在しない語zzz"))
	if !strings.Contains(w.Body.String(), "見つかりませんでした") {
		t.Errorf("無い語のとき")
	}
	w = get(t, h, "/help")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "答えは出さない") {
		t.Errorf("使い方の画面: %d", w.Code)
	}
}

// ---------------------------------------------------------------- 小さな部品

func TestSmallHelpers(t *testing.T) {
	// 合い言葉は毎回違い、十分に長い
	a, b := newToken(), newToken()
	if a == b || len(a) < 32 {
		t.Errorf("合い言葉が弱い: %q %q", a, b)
	}
	// 表の行の分割
	if got := splitRow("| a | b | c |"); len(got) != 3 || got[1] != "b" {
		t.Errorf("splitRow = %v", got)
	}
	// 作業漏れの文の中のユニットとログを、リンクにする
	if got := linkifyGap("F3: 課題は全部できている"); !strings.Contains(got, `href="/unit/F3"`) {
		t.Errorf("ユニットがリンクになっていない: %s", got)
	}
	if got := linkifyGap("logs/2026/01-05.md: 振り返りの空欄が2個"); !strings.Contains(got, `href="/retro?file=logs/2026/01-05.md"`) {
		t.Errorf("ログがリンクになっていない: %s", got)
	}
	if got := linkifyGap("<b>x</b>: y"); strings.Contains(got, "<b>") {
		t.Errorf("HTML をそのまま通している: %s", got)
	}
	if (errAbort{}).Error() == "" {
		t.Error("errAbort に文が無い")
	}
}
