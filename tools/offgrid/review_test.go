package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// 以下は、別ベンダーのAI（Codex）のレビューで見つかり、こちらで再現したもの。
// どれも「書けた」と言いながら学習者の記録が消える種類の壊れ方だった。

// ターミナル（別プロセス）とブラウザが同時に書いても、記録が落ちないこと。
// 直す前は、50件のうち2件しか残らず、しかもエラーは0件だった。
func TestConcurrentWritesKeepEveryRecord(t *testing.T) {
	root := newRepo(t)
	sess, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	const n = 30
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, n)
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			st := LoadState(root, sess) // 各自が別々に読む（プロセスが別なのと同じ）
			<-start
			errs <- st.Skip("F3", i)
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("保存に失敗: %v", err)
		}
	}
	if got := len(LoadState(root, sess).Skipped["F3"]); got != n {
		t.Errorf("あとで回した課題が %d/%d しか残っていない", got, n)
	}
}

// 読めないファイルを「まだ無い」と扱って上書きしないこと。
// 直す前は、権限のエラーでも空の記録を書き、メモも進みも消えた。
func TestUnreadableRecordsAreNotOverwritten(t *testing.T) {
	root := newRepo(t)
	sh, err := LoadSheet(root, "F3")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"notes.md", "work.md"} {
		path := filepath.Join(sh.Dir(), name)
		if err := os.WriteFile(path, []byte("学習者が書いたもの\n"), 0o000); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0o644) })
	}
	if err := sh.AppendNote("新しいメモ"); err == nil {
		t.Error("読めない notes.md に追記して、エラーを返していない")
	}
	if _, _, err := sh.EnsureWork(); err == nil {
		t.Error("読めない work.md を作り直して、エラーを返していない")
	}
	for _, name := range []string{"notes.md", "work.md"} {
		path := filepath.Join(sh.Dir(), name)
		if err := os.Chmod(path, 0o644); err != nil {
			t.Fatal(err)
		}
		if raw, _ := os.ReadFile(path); !strings.Contains(string(raw), "学習者が書いたもの") {
			t.Errorf("%s の中身が消えた: %q", name, raw)
		}
	}

	// セッションの進みも同じ
	sess, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	st := LoadState(root, sess)
	if err := st.Skip("F3", 1); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(st.path, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(st.path, 0o644) })
	if err := LoadState(root, sess).Finish(StepWarmup); err == nil {
		t.Error("読めない state を上書きして、エラーを返していない")
	}
	if err := os.Chmod(st.path, 0o644); err != nil {
		t.Fatal(err)
	}
	if !LoadState(root, sess).isSkipped("F3", 1) {
		t.Error("あとで回した課題が消えた")
	}
}

// 保存で、ファイルの見える範囲を広げないこと（0600 のメモが 0644 になっていた）。
func TestWriteKeepsFileMode(t *testing.T) {
	root := newRepo(t)
	sh, err := LoadSheet(root, "F3")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(sh.Dir(), "notes.md")
	if err := os.WriteFile(path, []byte("# メモ\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := sh.AppendNote("追記"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("権限が %04o に変わった（0600 のはず）", fi.Mode().Perm())
	}
}

// 振り返りの欄の識別子が、1つ書いたあともずれないこと。
// 直す前は、空いている欄だけで番号を振っていたので、同じ送信が別の欄に入った。
func TestRetroFieldKeysDoNotShift(t *testing.T) {
	root := newRepo(t)
	sess, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.EnsureLog(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sess.Log, []byte("## 記録\n1.\n2.\n3.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fields := RetroFields(sess.Log)
	if len(fields) != 3 {
		t.Fatalf("空欄が3つのはず: %v", fields)
	}
	keys := []string{fields[0].Key, fields[1].Key, fields[2].Key}
	if err := RetroWrite(root, sess.Log, keys[0], "ひとつめ"); err != nil {
		t.Fatal(err)
	}
	// 1つ埋めても、2番目・3番目の識別子は変わらない
	if err := RetroWrite(root, sess.Log, keys[2], "みっつめ"); err != nil {
		t.Fatalf("3番目に書けない: %v", err)
	}
	raw, _ := os.ReadFile(sess.Log)
	if got := string(raw); got != "## 記録\n1. ひとつめ\n2.\n3. みっつめ\n" {
		t.Errorf("狙った欄に入っていない:\n%s", got)
	}
	// 同じ送信をもう一度しても、別の欄には入らない
	if err := RetroWrite(root, sess.Log, keys[0], "ひとつめ"); err == nil {
		t.Error("埋まっている欄への二重送信を受け入れた")
	}
	raw, _ = os.ReadFile(sess.Log)
	if strings.Count(string(raw), "ひとつめ") != 1 {
		t.Errorf("同じ文が2つの欄に入った:\n%s", raw)
	}
}

// 別のサイトから GET で開かれただけで、終わりの手続きが当日ログを作らないこと。
// /session・/unit/・/retro は塞いだが、/end を入れ忘れていた。
func TestCrossSiteGetOnEndCreatesNothing(t *testing.T) {
	_, h, root := newServer(t)
	sess, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodGet, "/end", nil)
	r.Host = "localhost:7777"
	r.RemoteAddr = "127.0.0.1:50000"
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	r.Header.Set("Origin", "https://evil.example")
	if w := do(t, h, r); w.Code != http.StatusForbidden {
		t.Errorf("別サイトからの GET /end = %d, want 403", w.Code)
	}
	if _, err := os.Stat(sess.Log); err == nil {
		t.Error("別サイトからの GET で、その日の振り返りファイルができた")
	}
	if w := get(t, h, "/end"); w.Code != http.StatusOK {
		t.Errorf("自分で開いた /end = %d", w.Code)
	}
}

// docs/ や logs/ に置いたリンクから、リポジトリの外や answers/ に届かないこと。
func TestSymlinksCannotLeaveTheRepository(t *testing.T) {
	s, h, root := newServer(t)
	outside := t.TempDir()
	outsideFile := filepath.Join(outside, "probe.md")
	if err := os.WriteFile(outsideFile, []byte("# 外のファイル\n- \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(root, "docs", "alias.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "logs", "alias")); err != nil {
		t.Fatal(err)
	}
	if w := get(t, h, "/file/docs/alias.md"); strings.Contains(w.Body.String(), "外のファイル") {
		t.Error("リンク経由で、リポジトリの外のファイルが読めた")
	}
	if _, ok := s.logPath("logs/alias/probe.md"); ok {
		t.Error("リンク経由で、リポジトリの外に書ける")
	}
	// answers/ へのリンクも通さない
	if err := os.MkdirAll(filepath.Join(root, "answers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "answers", "a.md"), []byte("こたえ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "answers", "a.md"), filepath.Join(root, "docs", "ans.md")); err != nil {
		t.Fatal(err)
	}
	if w := get(t, h, "/file/docs/ans.md"); strings.Contains(w.Body.String(), "こたえ") {
		t.Error("リンク経由で answers/ が読めた")
	}
	if _, ok := s.logPath("logs/answers/a.md"); ok {
		t.Error("logs/answers/ を通している")
	}
}

// 日本語の名前のバイナリも、ちゃんとコミットから外れること。
// git の numstat は非ASCIIの名前を引用して出すので、そのまま渡しても外れなかった。
func TestBinaryWithJapaneseNameIsLeftOut(t *testing.T) {
	_, h, root := newServer(t)
	initGit(t, root)
	sh, _ := LoadSheet(root, "F3")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "notes.md"), []byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	name := "実行ファイル"
	if err := os.WriteFile(filepath.Join(root, name), []byte("\x7fELF\x00\x01binary\x00"), 0o755); err != nil {
		t.Fatal(err)
	}
	w := post(t, h, "/end/commit", url.Values{"action": {"commit"}, "message": {"メモ"}, "unit": {"F3"}})
	if w.Code != 303 {
		t.Fatalf("commit = %d\n%s", w.Code, w.Body.String())
	}
	tracked, _ := gitOut(root, "ls-files", "-z")
	if strings.Contains(tracked, name) {
		t.Error("日本語の名前の実行ファイルがコミットされた")
	}
	if !strings.Contains(w.Header().Get("Location"), "excluded=") {
		t.Error("外したことを知らせていない")
	}
}

// つなぎのシート（TASKS.local.md）だけのユニットも、selftest と status から消えないこと。
func TestLocalSheetIsNotDroppedFromAllSheets(t *testing.T) {
	root := newRepo(t)
	before, err := AllSheets(root)
	if err != nil {
		t.Fatal(err)
	}
	sh, err := LoadSheet(root, "F3")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(sh.Path, filepath.Join(sh.Dir(), "TASKS.local.md")); err != nil {
		t.Fatal(err)
	}
	after, err := AllSheets(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Errorf("TASKS.local.md にしたら %d本 → %d本 に減った", len(before), len(after))
	}
}

// 名前を変えた実行ファイルも、コミットから外れること。
// numstat の -z は、rename のとき「旧名 NUL 新名」と2つ並べて出す。
func TestRenamedBinaryIsLeftOut(t *testing.T) {
	_, h, root := newServer(t)
	old := filepath.Join(root, "旧名")
	if err := os.WriteFile(old, []byte("\x7fELF\x00\x01binary\x00"), 0o755); err != nil {
		t.Fatal(err)
	}
	initGit(t, root) // 旧名を追跡した状態から始める
	if out, err := gitOut(root, "mv", "旧名", "新名"); err != nil {
		t.Fatalf("git mv: %s", out)
	}
	sh, _ := LoadSheet(root, "F3")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "notes.md"), []byte("メモ\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if w := post(t, h, "/end/commit", url.Values{"action": {"commit"}, "message": {"メモ"}, "unit": {"F3"}}); w.Code != 303 {
		t.Fatalf("commit = %d\n%s", w.Code, w.Body.String())
	}
	staged, _ := gitOut(root, "diff", "--cached", "--name-only")
	if strings.Contains(staged, "新名") {
		t.Errorf("名前を変えた実行ファイルが、まだ索引に載っている:\n%s", staged)
	}
}

// 以下は、#27（まだ誰もレビューしていなかった範囲を別ベンダーのAIに見せた）で見つかり、再現したもの。

// リンクの書き換えが継ぎ足すディレクトリ名から、HTML を差し込めないこと。
// Markdown の本文は先にエスケープしていたが、継ぎ足した部分はそのまま属性に入っていた。
func TestRewrittenLinkIsEscaped(t *testing.T) {
	_, h, root := newServer(t)
	dir := filepath.Join(root, "drills", "foundation", `F02-"><img src=x onerror=alert(1)>`)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte("[次へ](next.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := get(t, h, "/file/drills/foundation/"+url.PathEscape(`F02-"><img src=x onerror=alert(1)>`)+"/notes.md").Body.String()
	// 書き換えが本当に起きていること（起きていなければ、この検査は素通りする）
	if !strings.Contains(body, `href="/file/drills/foundation/F02-`) {
		t.Fatalf("相対リンクが /file/ に書き換わっていない:\n%s", body)
	}
	if strings.Contains(body, "<img src=x") {
		t.Errorf("ディレクトリ名から要素を差し込めた:\n%s", body)
	}
}

// /find が、リンクをたどってリポジトリの外や answers/ の中身を返さないこと。
// /file/ と /retro は #25 で塞いだが、/find は見かけのパスしか見ていなかった。
func TestFindDoesNotFollowLinksOut(t *testing.T) {
	_, h, root := newServer(t)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("OUTSIDE_SENTINEL\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sh, _ := LoadSheet(root, "F3")
	if err := os.Symlink(outside, filepath.Join(sh.Dir(), "outside.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "answers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "answers", "a.md"), []byte("ANSWER_SENTINEL\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "answers", "a.md"), filepath.Join(sh.Dir(), "ans.md")); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"OUTSIDE_SENTINEL", "ANSWER_SENTINEL"} {
		if body := get(t, h, "/find?q="+q).Body.String(); strings.Contains(body, "<mark>"+q) {
			t.Errorf("ブラウザの検索が、リンク越しに %s を返した", q)
		}
		out := capture(t, func() { _ = cmdFind(root, []string{q}) })
		if strings.Contains(out, q) && strings.Contains(out, ".md:") {
			t.Errorf("ターミナルの検索が、リンク越しに %s を返した:\n%s", q, out)
		}
	}
}

// answers/ を指す相対リンクは、押せるリンクにしない（開けない行き先を見せない）。
// 大文字のスキーム（HTTPS://）は、外のリンクのまま扱う。
func TestLinkTargetsThatShouldNotBecomeFileLinks(t *testing.T) {
	_, h, root := newServer(t)
	sh, _ := LoadSheet(root, "F3")
	md := "[メモ](notes.md)\n\n[こたえ](../../../answers/probe.md)\n\n[外](HTTPS://example.com/)\n"
	if err := os.WriteFile(filepath.Join(sh.Dir(), "links.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	body := get(t, h, "/file/drills/foundation/F03-pipeline/links.md").Body.String()
	if !strings.Contains(body, `href="/file/drills/foundation/F03-pipeline/notes.md"`) {
		t.Fatalf("ふつうの相対リンクが書き換わっていない:\n%s", body)
	}
	if strings.Contains(body, ">こたえ</a>") {
		t.Errorf("answers/ を指すリンクを、押せるリンクとして出した:\n%s", body)
	}
	if strings.Contains(body, "HTTPS:/example.com") || !strings.Contains(body, `href="HTTPS://example.com/"`) {
		t.Errorf("大文字の HTTPS をファイルへのリンクに書き換えた:\n%s", body)
	}
}

// 学習用リポジトリの置き場所の途中にシンボリックリンクがあっても、相対リンクが書き換わること。
// #25 で /file/ がリンクをたどった先のパスを返すようにしたとき、linkRewriter は
// たどる前のルートと比べていたので、書き換えが丸ごと効かなくなっていた。
// Linux の /tmp はリンクではないので、環境まかせにせず、リンクを挟んだルートを明示的に作る。
func TestLinksAreRewrittenUnderASymlinkedRoot(t *testing.T) {
	actual := newRepo(t)
	linked := filepath.Join(t.TempDir(), "offgrid-log")
	if err := os.Symlink(actual, linked); err != nil {
		t.Fatal(err)
	}
	s := &server{root: linked, token: "test-token"}
	mux := http.NewServeMux()
	mux.HandleFunc("/file/", s.handleFile)
	h := s.guard(mux)
	sh, _ := LoadSheet(linked, "F3")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "links.md"), []byte("[メモ](notes.md)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := get(t, h, "/file/drills/foundation/F03-pipeline/links.md").Body.String()
	if !strings.Contains(body, `href="/file/drills/foundation/F03-pipeline/notes.md"`) {
		t.Errorf("リンクを挟んだルートで、相対リンクが書き換わらない:\n%s", body)
	}
}

// NO_COLOR は色だけを消す。端末なら、引数なしでメニューが出る（直す前は、色の有無で端末かどうかを決めていた）
func TestNoColorKeepsTheMenu(t *testing.T) {
	root := newRepo(t)
	t.Setenv("OFFGRID_REPO", root)
	prevArgs, prevC, prevI := os.Args, colorOn, interactive
	t.Cleanup(func() { os.Args, colorOn, interactive = prevArgs, prevC, prevI })
	os.Args = []string{"offgrid"}
	colorOn, interactive = false, true // 端末だが NO_COLOR が付いている
	withInput(t, "q")
	out := capture(t, func() {
		if err := run(); err != nil {
			t.Errorf("run: %v", err)
		}
	})
	if !strings.Contains(out, "何をする？") {
		t.Errorf("NO_COLOR でメニューが消えた:\n%s", out)
	}
}

// 引数を取らないコマンドに余分な引数が来たら、黙って通さない
func TestStrayArgumentsAreRejected(t *testing.T) {
	root := newRepo(t)
	t.Setenv("OFFGRID_REPO", root)
	prev := os.Args
	t.Cleanup(func() { os.Args = prev })
	for _, args := range [][]string{{"version", "x"}, {"today", "x"}, {"status", "x"}, {"selftest", "--bogus"}, {"doctor", "x"}} {
		os.Args = append([]string{"offgrid"}, args...)
		var err error
		capture(t, func() { err = run() })
		if err == nil {
			t.Errorf("offgrid %s を通した", strings.Join(args, " "))
		}
	}
	// 正しい使い方は、これまでどおり通る
	os.Args = []string{"offgrid", "selftest", "-q"}
	var err error
	capture(t, func() { err = run() })
	if err != nil {
		t.Errorf("selftest -q が通らない: %v", err)
	}
}

// 折り返しは、2行目以降の字下げを引いた幅に収め、色の指定を途中で切らない
func TestWrapKeepsWidthAndColour(t *testing.T) {
	prevW, prevC := termWidth, colorOn
	t.Cleanup(func() { termWidth, colorOn = prevW, prevC })
	termWidth, colorOn = 48, true

	for _, l := range wrap("- "+strings.Repeat("あ", 60), "  ") {
		if width(l) > 48 {
			t.Errorf("端末の幅（48）をはみ出した（%d）: %q", width(l), l)
		}
	}
	for _, l := range wrap(strings.Repeat("あ", 22)+green("XYZ")+strings.Repeat("い", 10), "  ") {
		if strings.HasSuffix(l, "\x1b[") || strings.HasPrefix(strings.TrimLeft(l, " "), "32m") {
			t.Errorf("色の指定を途中で切った: %q", l)
		}
		if width(l) > 48 {
			t.Errorf("端末の幅をはみ出した: %q", l)
		}
	}
	p := pad(green("abcdefgh"), 5)
	if strings.Contains(p, "\x1b[32…") || !strings.HasSuffix(p, "…") || width(p) > 5 {
		t.Errorf("省略が色の指定を切った、または幅を超えた: %q（幅 %d）", p, width(p))
	}
}

// 抜けていた絵文字の範囲と、幅を持たない文字
func TestEmojiWidths(t *testing.T) {
	for _, c := range []struct {
		s    string
		want int
	}{{"🚀", 2}, {"日本語", 6}, {"é", 1}, {"‍", 0}, {"🪐", 2}} {
		if got := width(c.s); got != c.want {
			t.Errorf("width(%q) = %d, want %d", c.s, got, c.want)
		}
	}
}

// Markdown の装飾は、コードの中とリンク先の中に当てない。「**[リンク](…)**」の太字は保つ
func TestInlineFormattingStaysOutOfCodeAndHrefs(t *testing.T) {
	for _, c := range []struct {
		in, want, notWant string
	}{
		{"`[x](https://example.com/)`", "<code>[x](https://example.com/)</code>", "<a "},
		{"[a](https://example.com/**draft**)", `href="https://example.com/**draft**"`, "<strong>draft"},
		{"**[O8](https://example.com/)**", "<strong><a href=\"https://example.com/\"", ""},
		{"`**本題**` と書く", "<code>**本題**</code>", "<code><strong>"},
		{"**`COUNT(*)` と `COUNT(列)`**：", "<strong><code>COUNT(*)</code> と <code>COUNT(列)</code></strong>", ""},
	} {
		got := inlineHTML(c.in, nil)
		if !strings.Contains(got, c.want) {
			t.Errorf("inlineHTML(%q)\n got: %s\nwant: %s", c.in, got, c.want)
		}
		if c.notWant != "" && strings.Contains(got, c.notWant) {
			t.Errorf("inlineHTML(%q) に %q が出た: %s", c.in, c.notWant, got)
		}
	}
}
