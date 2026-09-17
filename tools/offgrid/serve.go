package main

import (
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed web
var webFS embed.FS

var tmpl = template.Must(template.ParseFS(webFS, "web/layout.html", "web/session.html"))

// serveState は、1つの画面に渡すもの。テンプレートから読む。
type serveState struct {
	Title, Nav, Query, Version, Repo string
	CSRF                             string
	Unit                             string
	Progress                         *Progress
	Session                          *Session
	Sheet                            *Sheet
	SheetPath, LogPath, DocPath      string
	HeaderHTML                       template.HTML
	Steps                            []string
	Tracks                           []trackView
	TaskDone, TaskTotal, TaskPercent int
	Gaps                             []template.HTML
	Tasks                            []taskView
	Extra                            []sectionView
	Keep                             []keepView
	Fields                           []RetroField
	Logs                             []string
	Body                             template.HTML
	DocTitle                         string
	DocNav                           []docLink
	Groups                           []hitGroup
	HitCount                         int
	Heading, Message                 string
	Links                            []docLink

	// 案内の画面（/session）で使うもの
	State         *SessionState
	Step          Step
	StepIDStr     string
	PrevStepID    StepID
	StepViews     []stepView
	StepIndex     int
	StepCount     int
	Elapsed       string
	Finished      bool
	Task          *taskView
	OnlyLater     bool
	ShowHint      bool
	StuckOpen     bool
	Hints         []sectionView
	Guide         template.HTML
	CanFinishUnit bool
	Dirty         string
	Ahead         int
	Advice        template.HTML
	TaskMins      int
	LastStuck     []string
	Leftover      []int
	CheckOut      string
	CheckCmd      string
	FirstTime     bool
	NotesPath     string
	NotesBody     template.HTML
	Reading       template.HTML
	TermHint      string
}

type stepView struct {
	Index         int
	Label         string
	Minutes       string
	Done, Current bool
}

type trackView struct {
	Name        string
	Done, Total int
	Percent     int
}

type taskView struct {
	N    int
	Done bool
	HTML template.HTML
}

type sectionView struct {
	Name string
	HTML template.HTML
}

type keepView struct {
	Name string
	Have bool
}

type docLink struct {
	Slug, Label, URL string
	Current          bool
}

type hitGroup struct {
	Label string
	Hits  []hitView
}

type hitView struct {
	Path, Link string
	Line       int
	HTML       template.HTML
}

type server struct {
	root  string
	token string // 画面が出す合い言葉。書き込みのときに確かめる（CSRF対策）
}

func percent(done, total int) int {
	if total <= 0 {
		return 0
	}
	return done * 100 / total
}

// base は、どの画面でも使う値を用意する。
func (s *server) base(title, nav string) (*serveState, *Progress, *Session, error) {
	p, err := LoadProgress(s.root)
	if err != nil {
		return nil, nil, nil, err
	}
	sess, err := LoadSession(s.root)
	if err != nil {
		return nil, nil, nil, err
	}
	st := &serveState{
		Title: title, Nav: nav, Version: version, Repo: s.root,
		Unit: p.Unit, Progress: p, Session: sess, CSRF: s.token,
	}
	return st, p, sess, nil
}

func (s *server) render(w http.ResponseWriter, body string, st *serveState) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := tmpl.Clone()
	if err == nil {
		_, err = t.New("body").Parse(fmt.Sprintf(`{{template "%s" .}}`, body))
	}
	if err == nil {
		err = t.ExecuteTemplate(w, "layout", st)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) fail(w http.ResponseWriter, st *serveState, heading, message string) {
	if st == nil {
		http.Error(w, heading+": "+message, http.StatusBadRequest)
		return
	}
	st.Title, st.Heading, st.Message = heading, heading, message
	st.Links = []docLink{{Label: "いまに戻る", URL: "/"}}
	s.render(w, "message", st)
}

// linkRewriter は、Markdown の中の相対リンクを、この画面のURLに直す。
func (s *server) linkRewriter(from string) func(string) string {
	return func(href string) string {
		if strings.HasPrefix(href, "http") || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "/") {
			return href
		}
		target := filepath.Clean(filepath.Join(filepath.Dir(from), strings.SplitN(href, "#", 2)[0]))
		frag := ""
		if i := strings.Index(href, "#"); i >= 0 {
			frag = href[i:]
		}
		rel, err := filepath.Rel(s.root, target)
		if err != nil || strings.HasPrefix(rel, "..") {
			return href
		}
		switch {
		case strings.HasPrefix(rel, "docs/") && strings.HasSuffix(rel, ".md"):
			return "/doc/" + strings.TrimSuffix(filepath.Base(rel), ".md") + frag
		case strings.HasSuffix(rel, "TASKS.md"), strings.HasSuffix(rel, "TASKS.local.md"):
			dir := filepath.Base(filepath.Dir(rel))
			return "/unit/" + strings.SplitN(dir, "-", 2)[0] + frag
		case strings.HasPrefix(rel, "logs/"):
			return "/retro?file=" + rel
		default:
			return "/file/" + rel + frag
		}
	}
}

func (s *server) handleHome(w http.ResponseWriter, r *http.Request) {
	st, p, sess, err := s.base("いま", "home")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	plan := StagePlan(p.Stage)
	sheetPath := p.Unit
	if sh, err := LoadSheet(s.root, p.Unit); err == nil {
		st.Sheet = sh
		st.TaskDone, st.TaskTotal = sh.Counts()
		st.TaskPercent = percent(st.TaskDone, st.TaskTotal)
		sheetPath = rel(s.root, sh.Path)
	}
	steps := []string{
		"ウォームアップ（30分）：" + sess.Warmup(),
		fmt.Sprintf("メイン（%s）：%s", plan.Main, sheetPath),
	}
	if plan.DB != "" {
		db := p.DBUnit
		if db == "" {
			db = "（PROGRESS.md に「次のPostgreSQLユニット」を書く）"
		}
		steps = append(steps, fmt.Sprintf("PostgreSQL（%s）：%s", plan.DB, db))
	}
	steps = append(steps, fmt.Sprintf("%s（1.5時間）", sess.Afternoon()), "振り返り（30分）："+rel(s.root, sess.Log))
	st.Steps = steps
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
		st.Tracks = append(st.Tracks, trackView{Name: t.Name, Done: done, Total: len(t.Units), Percent: percent(done, len(t.Units))})
	}
	for _, g := range FindGaps(s.root, p, sess) {
		st.Gaps = append(st.Gaps, template.HTML(linkifyGap(g)))
	}
	s.render(w, "home", st)
}

// linkifyGap は、作業漏れの文の中のユニットIDとパスを、クリックできるようにする。
func linkifyGap(g string) string {
	esc := template.HTMLEscapeString(g)
	if i := strings.Index(esc, ":"); i > 0 {
		head := esc[:i]
		if unitIDRe.MatchString(head) {
			return fmt.Sprintf(`<a href="/unit/%s">%s</a>%s`, head, head, esc[i:])
		}
		if strings.HasPrefix(head, "logs/") {
			return fmt.Sprintf(`<a href="/retro?file=%s">%s</a>%s`, head, head, esc[i:])
		}
	}
	return esc
}

func (s *server) handleUnit(w http.ResponseWriter, r *http.Request) {
	st, _, _, err := s.base("課題", "unit")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/unit/")
	sh, err := LoadSheet(s.root, id)
	if err != nil {
		s.fail(w, st, "見つかりません", err.Error())
		return
	}
	if _, _, err := sh.EnsureWork(); err != nil {
		s.fail(w, st, "書けません", err.Error())
		return
	}
	st.Sheet = sh
	st.Title = sh.Title
	st.SheetPath = rel(s.root, sh.Path)
	st.HeaderHTML = template.HTML(inlineHTML(sh.Header, s.linkRewriter(sh.Path)))
	st.TaskDone, st.TaskTotal = sh.Counts()
	st.TaskPercent = percent(st.TaskDone, st.TaskTotal)
	link := s.linkRewriter(sh.Path)
	state := sh.WorkState()
	for _, n := range sh.TaskNums() {
		st.Tasks = append(st.Tasks, taskView{
			N:    n,
			Done: state[n],
			HTML: template.HTML(renderMarkdown(sh.Tasks[n], link, nil)),
		})
	}
	for _, name := range []string{"ねらい", "このユニットで身につけること", "キーワード", "平日に読むもの", "準備", "発展（任意）", "詰まりやすいところ", "完了条件の確かめ方", "次につながるユニット"} {
		if body := sh.Section(name); body != "" {
			st.Extra = append(st.Extra, sectionView{Name: name, HTML: template.HTML(renderMarkdown(body, link, nil))})
		}
	}
	for _, f := range sh.KeepFiles() {
		st.Keep = append(st.Keep, keepView{Name: f, Have: sh.HaveKeep(f)})
	}
	s.render(w, "unit", st)
}

func (s *server) handleTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	st, _, _, err := s.base("課題", "unit")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	unit := r.FormValue("unit")
	n, _ := strconv.Atoi(r.FormValue("n"))
	done := r.FormValue("done") == "1"
	sh, err := LoadSheet(s.root, unit)
	if err != nil || n <= 0 {
		s.fail(w, st, "記録できません", "ユニットか課題の番号が違います")
		return
	}
	if err := sh.Tick(n, done); err != nil {
		s.fail(w, st, "記録できません", err.Error())
		return
	}
	http.Redirect(w, r, "/unit/"+unit+"#task", http.StatusSeeOther)
}

func (s *server) handleRetro(w http.ResponseWriter, r *http.Request) {
	st, _, sess, err := s.base("振り返り", "retro")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	log := sess.Log
	if v := r.FormValue("file"); v != "" {
		if cand, ok := s.logPath(v); ok {
			log = cand
		}
	}
	if log == sess.Log {
		// 振り返りは書くために開く画面なので、ここでファイルを用意する
		if err := sess.EnsureLog(); err != nil {
			s.fail(w, st, "書けません", err.Error())
			return
		}
	}
	if r.Method == http.MethodPost {
		// 見出しと項目名で探して書く。行番号を覚えておくと、
		// 画面を出したあとにファイルへ1行入るだけで、書き先がずれてしまう。
		var failed []string
		for key, vals := range r.Form {
			if !strings.HasPrefix(key, "field-") || len(vals) == 0 || strings.TrimSpace(vals[0]) == "" {
				continue
			}
			if err := RetroWrite(log, strings.TrimPrefix(key, "field-"), strings.TrimSpace(vals[0])); err != nil {
				failed = append(failed, err.Error())
			}
		}
		if len(failed) > 0 {
			sort.Strings(failed)
			st.LogPath = rel(s.root, log)
			s.fail(w, st, "書けなかった欄があります", strings.Join(failed, " / "))
			return
		}
		http.Redirect(w, r, "/retro?file="+url.QueryEscape(rel(s.root, log)), http.StatusSeeOther)
		return
	}
	st.LogPath = rel(s.root, log)
	st.Fields = RetroFields(log)
	if raw, err := os.ReadFile(log); err == nil {
		st.Body = template.HTML(renderMarkdown(string(raw), s.linkRewriter(log), nil))
	}
	for i := len(sess.Past) - 1; i >= 0 && len(st.Logs) < 8; i-- {
		if p := rel(s.root, sess.Past[i]); p != st.LogPath {
			st.Logs = append(st.Logs, p)
		}
	}
	s.render(w, "retro", st)
}

func (s *server) handleStuck(w http.ResponseWriter, r *http.Request) {
	st, _, sess, err := s.base("", "home")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	if text := strings.TrimSpace(r.FormValue("text")); text != "" {
		if err := sess.EnsureLog(); err != nil {
			s.fail(w, st, "書けません", err.Error())
			return
		}
		if err := AppendStuck(sess.Log, text); err != nil {
			s.fail(w, st, "書けません", err.Error())
			return
		}
	}
	http.Redirect(w, r, "/retro", http.StatusSeeOther)
}

func (s *server) docList(current string) []docLink {
	paths, _ := filepath.Glob(filepath.Join(s.root, "docs", "*.md"))
	sort.Strings(paths)
	var out []docLink
	for _, p := range paths {
		slug := strings.TrimSuffix(filepath.Base(p), ".md")
		label := slug
		if sh, err := os.ReadFile(p); err == nil {
			if m := frontRe.FindStringSubmatch(string(sh)); m != nil {
				label = m[1]
			}
		}
		out = append(out, docLink{Slug: slug, Label: label, Current: slug == current})
	}
	return out
}

func (s *server) handleDoc(w http.ResponseWriter, r *http.Request) {
	st, _, _, err := s.base("カリキュラム", "doc")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	slug := strings.TrimPrefix(r.URL.Path, "/doc/")
	if slug == "" {
		slug = "01-overview"
	}
	path := filepath.Join(s.root, "docs", filepath.Base(slug)+".md")
	raw, err := os.ReadFile(path)
	if err != nil {
		s.fail(w, st, "見つかりません", "docs/"+slug+".md がありません")
		return
	}
	body := string(raw)
	title := slug
	if m := frontRe.FindStringSubmatch(body); m != nil {
		title = m[1]
		body = body[len(m[0]):]
	}
	st.Title, st.DocTitle, st.DocPath = title, title, rel(s.root, path)
	st.DocNav = s.docList(slug)
	st.Body = template.HTML(renderMarkdown(body, s.linkRewriter(path), nil))
	s.render(w, "doc", st)
}

func (s *server) handleFile(w http.ResponseWriter, r *http.Request) {
	st, _, _, err := s.base("ファイル", "")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	path, ok := s.markdownPath(strings.TrimPrefix(r.URL.Path, "/file/"))
	if !ok {
		s.fail(w, st, "開けません", "この道具から開けるのは、カリキュラムと自分の書いたMarkdownだけです（答えのファイルは開きません）")
		return
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		s.fail(w, st, "開けません", "そのファイルはありません")
		return
	}
	st.Title, st.DocTitle, st.DocPath = filepath.Base(path), filepath.Base(path), rel(s.root, path)
	st.Body = template.HTML(renderMarkdown(string(raw), s.linkRewriter(path), nil))
	s.render(w, "doc", st)
}

func (s *server) handleFind(w http.ResponseWriter, r *http.Request) {
	st, _, _, err := s.base("検索", "")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	q := strings.TrimSpace(r.FormValue("q"))
	st.Query = q
	if q == "" {
		s.render(w, "find", st)
		return
	}
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
	for _, g := range groups {
		var hits []hitView
		for _, pat := range g.patterns {
			paths, _ := filepath.Glob(filepath.Join(s.root, pat))
			sort.Strings(paths)
			for _, path := range paths {
				name := filepath.Base(path)
				if strings.HasPrefix(name, ".") || (g.skipTask && strings.HasPrefix(name, "TASKS")) {
					continue
				}
				raw, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				for i, line := range strings.Split(string(raw), "\n") {
					if !strings.Contains(line, q) {
						continue
					}
					text := template.HTMLEscapeString(strings.TrimSpace(line))
					text = strings.ReplaceAll(text, template.HTMLEscapeString(q), "<mark>"+template.HTMLEscapeString(q)+"</mark>")
					hits = append(hits, hitView{
						Path: rel(s.root, path),
						Link: s.linkRewriter(filepath.Join(s.root, "x"))(rel(s.root, path)),
						Line: i + 1,
						HTML: template.HTML(text),
					})
					if len(hits) >= 40 {
						break
					}
				}
			}
		}
		if len(hits) > 0 {
			st.Groups = append(st.Groups, hitGroup{Label: g.label, Hits: hits})
			st.HitCount += len(hits)
		}
	}
	s.render(w, "find", st)
}

func cmdServe(root string, args []string) error {
	port := "7777"
	open := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			}
		case "--open":
			open = true
		default:
			if _, err := strconv.Atoi(args[i]); err == nil {
				port = args[i]
			}
		}
	}
	s := &server{root: root, token: newToken()}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			s.handleHome(w, r)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/session", s.handleSession)
	mux.HandleFunc("/session/step", s.handleSessionStep)
	mux.HandleFunc("/session/task", s.handleSessionTask)
	mux.HandleFunc("/note", s.handleNote)
	mux.HandleFunc("/help", s.handleHelp)
	mux.HandleFunc("/end", s.handleEnd)
	mux.HandleFunc("/end/check", s.handleEndCheck)
	mux.HandleFunc("/end/done", s.handleEndDone)
	mux.HandleFunc("/end/commit", s.handleEndCommit)
	mux.HandleFunc("/unit/", s.handleUnit)
	mux.HandleFunc("/tick", s.handleTick)
	mux.HandleFunc("/retro", s.handleRetro)
	mux.HandleFunc("/stuck", s.handleStuck)
	mux.HandleFunc("/doc/", s.handleDoc)
	mux.HandleFunc("/file/", s.handleFile)
	mux.HandleFunc("/find", s.handleFind)
	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		raw, err := webFS.ReadFile("web/style.css")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(raw)
	})

	addr := "127.0.0.1:" + port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("%s を開けません: %w（別の番号にするなら offgrid serve --port 7788）", addr, err)
	}
	url := "http://localhost:" + port + "/"
	fmt.Println(green("  " + url + " を開いてください"))
	say("ネットワークにもAIにもつながない。手元のファイルだけを読み書きする")
	say(dim("止めるときは Ctrl+C。Lima や WSL2 の中で動かしていても、手元のブラウザから開ける"))
	if open {
		go openBrowser(url)
	}
	srv := &http.Server{Handler: s.guard(mux), ReadHeaderTimeout: 5 * time.Second}
	return srv.Serve(ln)
}

func openBrowser(url string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "explorer"
	default:
		cmd = "xdg-open"
	}
	if path, err := exec.LookPath(cmd); err == nil {
		_ = exec.Command(path, url).Start()
	}
}
