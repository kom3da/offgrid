package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// v1.1.0 前のレビューで実測された穴。それぞれ、再発したら落ちる形で残す。

// ターミナルの run とブラウザの serve は同じ state ファイルを共有する。
// 片方がメモリ上の古い state を丸ごと書くと、もう片方の記録が消えていた。
func TestStateWritesMergeBetweenTerminalAndBrowser(t *testing.T) {
	root := newRepo(t)
	sess, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	terminal := LoadState(root, sess) // run が持ち続けるもの
	browser := LoadState(root, sess)  // serve はリクエストごとに読む

	if err := browser.Skip("F3", 2); err != nil {
		t.Fatal(err)
	}
	if err := terminal.Skip("F3", 1); err != nil {
		t.Fatal(err)
	}
	if err := browser.Finish(StepWarmup); err != nil {
		t.Fatal(err)
	}
	if err := terminal.Start(StepMain); err != nil {
		t.Fatal(err)
	}

	got := LoadState(root, sess)
	if !got.isSkipped("F3", 1) || !got.isSkipped("F3", 2) {
		t.Errorf("「あとで」が片方消えた: %v", got.Skipped)
	}
	if !got.IsDone(StepWarmup) {
		t.Error("ブラウザで済ませた枠が、ターミナルの保存で戻された")
	}
	if got.Steps[StepMain].Started.IsZero() {
		t.Error("ターミナルで始めた枠が記録されていない")
	}
	// メモリ上の state も、書いたあとは合流した中身になっている
	if !terminal.isSkipped("F3", 2) || !terminal.IsDone(StepWarmup) {
		t.Error("run の持つ state が、ブラウザの記録を取り込んでいない")
	}
	terminal.Reload()
	if !terminal.isSkipped("F3", 2) {
		t.Error("Reload が効いていない")
	}
}

// 別のサイトの <img src> で開かれた GET が、振り返りや作業記録のファイルを作っていた。
func TestCrossSiteGetDoesNotCreateFiles(t *testing.T) {
	_, h, root := newServer(t)
	sh, _ := LoadSheet(root, "F3")
	logsBefore, _ := filepath.Glob(filepath.Join(root, "logs", "*", "*.md"))

	for _, path := range []string{"/session", "/unit/F3", "/retro"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Host = "localhost:7777"
		r.RemoteAddr = "127.0.0.1:50000"
		r.Header.Set("Sec-Fetch-Site", "cross-site")
		r.Header.Set("Origin", "https://evil.example")
		if w := do(t, h, r); w.Code != http.StatusForbidden {
			t.Errorf("別サイトからの GET %s = %d, want 403", path, w.Code)
		}
	}
	if _, err := os.Stat(sh.WorkPath()); err == nil {
		t.Error("別サイトからの GET で work.md ができた")
	}
	if logsAfter, _ := filepath.Glob(filepath.Join(root, "logs", "*", "*.md")); len(logsAfter) != len(logsBefore) {
		t.Errorf("別サイトからの GET で振り返りファイルができた: %v", logsAfter)
	}

	// アドレス欄から開いたときは通る
	if w := get(t, h, "/unit/F3"); w.Code != http.StatusOK {
		t.Errorf("自分で開いた /unit/F3 = %d", w.Code)
	}
}

// 検索が drills/*/answers/ の中身（デバッグドリルの答え）をそのまま出していた。
func TestFindSkipsAnswers(t *testing.T) {
	_, h, root := newServer(t)
	dir := filepath.Join(root, "drills", "debug", "answers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("こたえ ANSWERTOKEN の本文\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if w := get(t, h, "/find?q=ANSWERTOKEN"); strings.Contains(w.Body.String(), "こたえ") {
		t.Error("ブラウザの検索が answers/ の行を出している")
	}
	out := capture(t, func() {
		if err := cmdFind(root, []string{"ANSWERTOKEN"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(out, "answers") || strings.Contains(out, "こたえ") {
		t.Errorf("ターミナルの検索が answers/ を出している:\n%s", out)
	}
}

// 検索語の印は、エスケープした行ではなく生の行に付ける（&amp; の途中に <mark> が入っていた）。
func TestFindMarksTheRawLine(t *testing.T) {
	_, h, root := newServer(t)
	sh, _ := LoadSheet(root, "F3")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "notes.md"), []byte("amp & x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := get(t, h, "/find?q=amp").Body.String()
	if !strings.Contains(body, "<mark>amp</mark> &amp; x") {
		t.Errorf("印の付け方がおかしい:\n%s", body)
	}
}

// 画面が「完了にする」を出していなくても、直接の POST で未完了のまま記録できていた。
func TestEndDoneRefusesUnfinishedUnit(t *testing.T) {
	_, h, root := newServer(t)
	w := post(t, h, "/end/done", url.Values{"unit": {"F3"}, "next": {"F4"}})
	if !strings.Contains(w.Body.String(), "まだ完了にできません") {
		t.Errorf("未完了の F3 を記録した（%d）\n%s", w.Code, w.Body.String())
	}
	p, _ := LoadProgress(root)
	if p.IsDone("F3") || p.Unit != "F3" {
		t.Errorf("PROGRESS.md が書き換わった: done=%v next=%s", p.IsDone("F3"), p.Unit)
	}
}

// git add -A が、ビルドした実行ファイルまで公開リポジトリに入れていた。
func TestCommitLeavesBinariesOut(t *testing.T) {
	_, h, root := newServer(t)
	initGit(t, root)
	sh, _ := LoadSheet(root, "F3")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "notes.md"), []byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := []byte("\x7fELF\x00\x01\x02\x00\x00binary\x00")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "mywc"), bin, 0o755); err != nil {
		t.Fatal(err)
	}
	w := post(t, h, "/end/commit", url.Values{"action": {"commit"}, "message": {"メモ"}, "unit": {"F3"}})
	if w.Code != 303 || !strings.Contains(w.Header().Get("Location"), "excluded=") {
		t.Fatalf("commit = %d Location=%s\n%s", w.Code, w.Header().Get("Location"), w.Body.String())
	}
	tracked, _ := gitOut(root, "ls-files")
	if strings.Contains(tracked, "mywc") {
		t.Error("実行ファイルがコミットされた")
	}
	if !strings.Contains(tracked, "F03-pipeline/notes.md") {
		t.Error("メモがコミットされていない")
	}
	if status, _ := gitOut(root, "status", "--porcelain"); !strings.Contains(status, "?? ") {
		t.Errorf("外したファイルが未追跡として残っていない:\n%s", status)
	}
	// 外したことが、戻った画面に出る
	body := get(t, h, w.Header().Get("Location")).Body.String()
	if !strings.Contains(body, "コミットから外した") || !strings.Contains(body, "mywc") {
		t.Error("外したファイルが画面に出ていない")
	}
}
