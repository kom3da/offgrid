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

// newServer は、検査用のリポジトリと、その上に立てた画面を返す。
func newServer(t *testing.T) (*server, http.Handler, string) {
	t.Helper()
	root := newRepo(t)
	s := &server{root: root, token: "test-token"}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.handleHome(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/session", s.handleSession)
	mux.HandleFunc("/session/task", s.handleSessionTask)
	mux.HandleFunc("/end", s.handleEnd)
	mux.HandleFunc("/end/check", s.handleEndCheck)
	mux.HandleFunc("/end/commit", s.handleEndCommit)
	mux.HandleFunc("/unit/", s.handleUnit)
	mux.HandleFunc("/tick", s.handleTick)
	mux.HandleFunc("/retro", s.handleRetro)
	mux.HandleFunc("/stuck", s.handleStuck)
	mux.HandleFunc("/note", s.handleNote)
	mux.HandleFunc("/session/step", s.handleSessionStep)
	mux.HandleFunc("/end/done", s.handleEndDone)
	mux.HandleFunc("/help", s.handleHelp)
	mux.HandleFunc("/find", s.handleFind)
	mux.HandleFunc("/doc/", s.handleDoc)
	mux.HandleFunc("/file/", s.handleFile)
	return s, s.guard(mux), root
}

func do(t *testing.T, h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// get は、ふつうのブラウザからの読み取り。
func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.Host = "localhost:7777"
	r.RemoteAddr = "127.0.0.1:50000"
	return do(t, h, r)
}

// post は、画面のボタンから送られた操作（合い言葉つき）。
func post(t *testing.T, h http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	if form == nil {
		form = url.Values{}
	}
	if form.Get("csrf") == "" {
		form.Set("csrf", "test-token")
	}
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	r.Header.Set("Origin", "http://localhost:7777")
	r.Host = "localhost:7777"
	r.RemoteAddr = "127.0.0.1:50000"
	return do(t, h, r)
}

// 外のWebページから、この画面を操作できないこと。
// 画像タグ1個で git commit や go test が走る、という穴があった。
func TestGuardBlocksCrossSiteWrites(t *testing.T) {
	_, h, _ := newServer(t)

	// postOnly に足したパスを、このテストに足し忘れないよう、表そのものを回す
	paths := []string{"/retro"}
	for p := range postOnly {
		paths = append(paths, p)
	}
	for _, path := range paths {
		// GET では通さない（<img src> で起こされるのを防ぐ）
		if path == "/retro" {
			continue // /retro は GET で読む画面。書く側は下の POST の検査で見る
		}
		if w := get(t, h, path+"?action=commit&message=x&text=x"); w.Code != http.StatusMethodNotAllowed {
			t.Errorf("GET %s = %d, want 405", path, w.Code)
		}
	}

	// 合い言葉が無いPOSTは通さない
	r := httptest.NewRequest(http.MethodPost, "/end/commit", strings.NewReader("action=commit&message=x"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Host = "localhost:7777"
	r.RemoteAddr = "127.0.0.1:50000"
	if w := do(t, h, r); w.Code != http.StatusForbidden {
		t.Errorf("合い言葉なしのPOST = %d, want 403", w.Code)
	}

	// ほかのサイトから送られたPOSTは通さない
	r = httptest.NewRequest(http.MethodPost, "/end/commit", strings.NewReader("csrf=test-token&action=commit&message=x"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	r.Host = "localhost:7777"
	r.RemoteAddr = "127.0.0.1:50000"
	if w := do(t, h, r); w.Code != http.StatusForbidden {
		t.Errorf("別サイトからのPOST = %d, want 403", w.Code)
	}
}

// 別の名前で呼ばれたとき（DNS rebinding）は、読み取りも断ること。
func TestGuardChecksHostAndRemoteAddr(t *testing.T) {
	_, h, _ := newServer(t)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "evil.example"
	r.RemoteAddr = "127.0.0.1:50000"
	if w := do(t, h, r); w.Code != http.StatusForbidden {
		t.Errorf("Host が別名のとき = %d, want 403", w.Code)
	}

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "localhost:7777"
	r.RemoteAddr = "203.0.113.9:40000"
	if w := do(t, h, r); w.Code != http.StatusForbidden {
		t.Errorf("外からの接続 = %d, want 403", w.Code)
	}
}

// 画面のボタンからの操作は、ちゃんと通ること。
func TestTickThroughForm(t *testing.T) {
	_, h, root := newServer(t)
	if w := get(t, h, "/unit/F3"); w.Code != http.StatusOK {
		t.Fatalf("/unit/F3 = %d", w.Code)
	}
	w := post(t, h, "/tick", url.Values{"unit": {"F3"}, "n": {"2"}, "done": {"1"}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("/tick = %d, body=%s", w.Code, w.Body.String())
	}
	raw, err := os.ReadFile(filepath.Join(root, "drills", "foundation", "F03-pipeline", "work.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "- [x] 2.") {
		t.Errorf("work.md に記録されていない:\n%s", raw)
	}
}

// 答えのファイルは、この道具から開けないこと（学習を奪わないため）。
func TestFileHandlerRefusesAnswers(t *testing.T) {
	_, h, root := newServer(t)
	if err := os.MkdirAll(filepath.Join(root, "answers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "answers", "secret.md"), []byte("# こたえ"), 0o644); err != nil {
		t.Fatal(err)
	}
	w := get(t, h, "/file/answers/secret.md")
	if strings.Contains(w.Body.String(), "こたえ") {
		t.Error("answers/ の中身が表示されている")
	}
	// drills の下にある answers/ も同じ
	if err := os.MkdirAll(filepath.Join(root, "drills", "debug", "answers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "drills", "debug", "answers", "a.md"), []byte("# こたえ2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if w := get(t, h, "/file/drills/debug/answers/a.md"); strings.Contains(w.Body.String(), "こたえ2") {
		t.Error("drills/debug/answers/ の中身が表示されている")
	}
	// リポジトリの外は見せない
	if _, ok := (&server{root: root}).markdownPath("../../../etc/passwd.md"); ok {
		t.Error("リポジトリの外を開こうとしている")
	}
	// カリキュラムは読める
	if w := get(t, h, "/file/templates/retrospective.md"); w.Code != http.StatusOK {
		t.Errorf("templates/retrospective.md = %d", w.Code)
	}
}

// /retro?file= は logs/ の中だけ。logs で始まる別のディレクトリは通さない。
func TestLogPathIsLimitedToLogs(t *testing.T) {
	root := t.TempDir()
	s := &server{root: root}
	if _, ok := s.logPath("logs-notes/private.md"); ok {
		t.Error("logs-notes/ を logs/ と見なしている")
	}
	if _, ok := s.logPath("logs/2026/09-18.md"); !ok {
		t.Error("logs/ の中が通らない")
	}
	if _, ok := s.logPath("logs/../PROGRESS.md"); ok {
		t.Error("logs/ の外に出られる")
	}
}

// 振り返りの書き込みは、行がずれても壊れないこと。
func TestRetroWriteSurvivesShiftedLines(t *testing.T) {
	root := newRepo(t)
	sess, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.EnsureLog(); err != nil {
		t.Fatal(err)
	}
	fields := RetroFields(sess.Log)
	if len(fields) < 3 {
		t.Fatalf("空欄が足りない: %+v", fields)
	}
	target := fields[0] // 「メインドリル（ユニット）：」
	// 画面を出したあとに、別のところから1行入る
	if err := AppendStuck(sess.Log, "xxd が読めない"); err != nil {
		t.Fatal(err)
	}
	if err := RetroWrite(sess.Log, target.Key, "F3をやった"); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(sess.Log)
	body := string(raw)
	if !strings.Contains(body, "- メインドリル（ユニット）：F3をやった") {
		t.Errorf("狙った欄に入っていない:\n%s", body)
	}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "#") && strings.Contains(line, "F3をやった") {
			t.Errorf("見出しに連結された: %q", line)
		}
	}
	// 二重送信では、同じ文を2回入れない
	if err := RetroWrite(sess.Log, target.Key, "F3をやった"); err == nil {
		t.Error("埋まっている欄に、もう一度書けてしまう")
	}
	raw, _ = os.ReadFile(sess.Log)
	if strings.Count(string(raw), "F3をやった") != 1 {
		t.Errorf("同じ文が2回入っている:\n%s", raw)
	}
}

// work.md の行が消えていても、黙って進まないこと。
func TestTickFailsWhenLineIsGone(t *testing.T) {
	root := newRepo(t)
	sh, err := LoadSheet(root, "F3")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := sh.EnsureWork(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(sh.WorkPath())
	var kept []string
	for _, line := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(line, "- [ ] 3.") {
			kept = append(kept, line)
		}
	}
	if err := os.WriteFile(sh.WorkPath(), []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	// 数は、いつも課題シートを正とする
	if _, total := sh.Counts(); total != 5 {
		t.Errorf("行を消したら全体数が変わった: total=%d, want 5", total)
	}
	// 消した行は、次に読むときに足し直す
	if _, added, err := sh.EnsureWork(); err != nil || len(added) != 1 || added[0] != 3 {
		t.Errorf("足し直していない: added=%v err=%v", added, err)
	}
	if err := sh.Tick(3, true); err != nil {
		t.Fatalf("足し直したのに記録できない: %v", err)
	}
}

// カリキュラムのページで、<URL> の形のリンクが、ちゃんとリンクになること。
func TestMarkdownLinks(t *testing.T) {
	out := inlineHTML("入手先は <https://devdocs.io/> にある", nil)
	if !strings.Contains(out, `<a href="https://devdocs.io/"`) {
		t.Errorf("<URL> がリンクになっていない: %s", out)
	}
	if strings.Contains(out, "&lt;https") {
		t.Errorf("山かっこが残っている: %s", out)
	}
	// 自分のメモに書いた javascript: は、リンクにしない
	out = inlineHTML("[押す](javascript:alert(1))", nil)
	if strings.Contains(out, "javascript:") {
		t.Errorf("javascript: が href に入っている: %s", out)
	}
	// ふつうの相対リンクは、そのまま
	out = inlineHTML("[F3](../F03-pipeline/TASKS.md)", nil)
	if !strings.Contains(out, `href="../F03-pipeline/TASKS.md"`) {
		t.Errorf("相対リンクが壊れている: %s", out)
	}
}

// 読むだけの画面は、その日の振り返りファイルを作らないこと。
func TestReadOnlyPagesDoNotStartASession(t *testing.T) {
	_, h, root := newServer(t)
	sess, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/unit/F3", "/doc/01-overview"} {
		if w := get(t, h, path); w.Code != http.StatusOK {
			t.Fatalf("%s = %d", path, w.Code)
		}
	}
	if _, err := os.Stat(sess.Log); err == nil {
		t.Error("読むだけで、振り返りファイルができている")
	}
	// 案内を始めたら、用意される
	if w := get(t, h, "/session"); w.Code != http.StatusOK {
		t.Fatalf("/session = %d", w.Code)
	}
	if _, err := os.Stat(sess.Log); err != nil {
		t.Errorf("案内を始めても用意されない: %v", err)
	}
}

// 振り返りへの書き込みが、画面から通ること。
func TestRetroPostWritesTheField(t *testing.T) {
	s, h, root := newServer(t)
	sess, _ := LoadSession(root)
	// 画面を開くと、その日の振り返りが用意される
	w := get(t, h, "/retro")
	if w.Code != http.StatusOK {
		t.Fatalf("/retro = %d", w.Code)
	}
	fields := RetroFields(sess.Log)
	if len(fields) == 0 {
		t.Fatal("空いている欄が無い")
	}
	w = post(t, h, "/retro", url.Values{
		"file":                   {rel(root, sess.Log)},
		"field-" + fields[0].Key: {"F1をやった"},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("POST /retro = %d, body=%s", w.Code, w.Body.String())
	}
	raw, _ := os.ReadFile(sess.Log)
	if !strings.Contains(string(raw), "メインドリル（ユニット）：F1をやった") {
		t.Errorf("狙った欄に入っていない:\n%s", raw)
	}
	// 同じ欄に二度書こうとしたら、画面で知らせる（黙って連結しない）
	w = post(t, h, "/retro", url.Values{
		"file":                   {rel(root, sess.Log)},
		"field-" + fields[0].Key: {"もう一度"},
	})
	if !strings.Contains(w.Body.String(), "書けなかった欄があります") {
		t.Errorf("埋まっている欄への書き込みを、黙って受けている（code=%d）", w.Code)
	}
	raw, _ = os.ReadFile(sess.Log)
	if strings.Contains(string(raw), "もう一度") {
		t.Errorf("埋まっている欄に上書きしている:\n%s", raw)
	}
	_ = s
}

// 詰まりメモとメモが、画面から書けること。
func TestStuckAndNotePostWrite(t *testing.T) {
	_, h, root := newServer(t)
	sess, _ := LoadSession(root)

	if w := post(t, h, "/stuck", url.Values{"text": {"sort の順が合わない"}}); w.Code != http.StatusSeeOther {
		t.Fatalf("POST /stuck = %d", w.Code)
	}
	raw, err := os.ReadFile(sess.Log)
	if err != nil {
		t.Fatalf("詰まりメモで振り返りが作られていない: %v", err)
	}
	body := string(raw)
	at := strings.Index(body, "## 詰まったこと")
	next := strings.Index(body[at:], "## 今日わかったこと") + at
	if !strings.Contains(body[at:next], "sort の順が合わない") {
		t.Errorf("「詰まったこと」に入っていない:\n%s", body[at:next])
	}

	// メモは、そのユニットの notes.md に入る
	if w := post(t, h, "/note", url.Values{"unit": {"F3"}, "text": {"printf は改行を付けない"}}); w.Code != http.StatusSeeOther {
		t.Fatalf("POST /note = %d", w.Code)
	}
	sh, _ := LoadSheet(root, "F3")
	notes, err := os.ReadFile(filepath.Join(sh.Dir(), "notes.md"))
	if err != nil {
		t.Fatalf("notes.md が作られていない: %v", err)
	}
	if !strings.Contains(string(notes), "printf は改行を付けない") {
		t.Errorf("メモが入っていない:\n%s", notes)
	}

	// 戻る先に外部のURLを渡しても、そこへは飛ばさない
	w := post(t, h, "/note", url.Values{"unit": {"F3"}, "text": {"x"}, "back": {"https://evil.example/"}})
	if loc := w.Header().Get("Location"); strings.HasPrefix(loc, "http") {
		t.Errorf("外部のURLへ飛ばしている: %q", loc)
	}
}
