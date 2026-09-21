package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 案内の画面（/session）。1つの枠、1つの課題だけを出して、ボタンで進める。

// lastStuckNotes は、前回の振り返りの「詰まったこと」を読む。
// 先生が「前回ここで止まったね」と言えるようにするため。
func lastStuckNotes(sess *Session) []string {
	if len(sess.Past) == 0 {
		return nil
	}
	raw, err := os.ReadFile(sess.Past[len(sess.Past)-1])
	if err != nil {
		return nil
	}
	var out []string
	inside := false
	for _, line := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(line, "## 詰まったこと"):
			inside = true
		case strings.HasPrefix(line, "#"):
			inside = false
		case inside && strings.HasPrefix(line, "- ") && strings.TrimSpace(line) != "-":
			out = append(out, strings.TrimPrefix(line, "- "))
		}
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return out
}

// taskAdvice は、時間の経ち方に応じた口出し。答えは言わない。
func taskAdvice(mins int) string {
	switch {
	case mins >= 45:
		return "この課題に45分。ここで打ち切って、「詰まった」に書いてから次へ進む。残した課題は、この枠の最後に戻ってくる。"
	case mins >= 15:
		return "この課題に15分。手を止めて、順番に切り替える。エラー文を最後まで読む → `man` や `go doc` で単語を調べる → 問題が起きる最小のコードに切り出す → それでも進まなければ公式ドキュメントを目次から引く。"
	default:
		return ""
	}
}

// stepSheet は、その枠で扱う課題シートを返す（枠が課題を持たないときは nil）。
func (s *server) stepSheet(p *Progress, step Step) *Sheet {
	var unit string
	switch step.ID {
	case StepMain:
		unit = p.Unit
	case StepDB:
		unit = p.DBUnit
	default:
		return nil
	}
	sh, err := LoadSheet(s.root, unit)
	if err != nil {
		return nil
	}
	return sh
}

// guideText は、課題を持たない枠の説明。
func (s *server) guideText(p *Progress, sess *Session, step Step) string {
	switch step.ID {
	case StepWarmup:
		text := sess.Warmup() + "\n\n＋ 前回のレビューで指摘された箇所の直しを1件。"
		if lg := sess.ReviewLog(); lg != "" {
			text += fmt.Sprintf("\n\n見返すログ：[%s](%s)", rel(s.root, lg), rel(s.root, lg))
		}
		return text
	case StepDB:
		return "PROGRESS.md の「次のPostgreSQLユニット」を、いま進めるDBユニット（例：D1）に書き換えると、ここでも課題を1問ずつ案内する。"
	case StepAfternoon:
		if sess.Afternoon() == "コードリーディング" {
			return "読解課題を1本。今のステージに合うものを[デバッグドリルとコードリーディング](docs/05-debug-and-reading.md)から選ぶ。処理の流れを図か箇条書きにまとめ、`drills/reading/` に置く。"
		}
		return "仕込みバグを3つ直す。問題は `drills/debug/` の中。答え（`answers/`）は、3つ直し終わるまで開かない。直したら、そのバグを見つけるテストを1本足す。"
	case StepRetro:
		return "今日やったこと、詰まったこと、分かったことを書く。空いている欄が入力欄として並ぶので、上から埋める。"
	case StepEnd:
		return "振り返りの記入漏れ、残すもの、コミットを確かめる。"
	}
	return ""
}

func (s *server) sessionState(title, nav string) (*serveState, *Progress, *Session, *SessionState, []Step, error) {
	st, p, sess, err := s.base(title, nav)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	state := LoadState(s.root, sess)
	steps := sessionSteps(p, sess)
	st.State = state
	st.StepCount = len(steps)
	for i, step := range steps {
		st.StepViews = append(st.StepViews, stepView{
			Index:   i + 1,
			Label:   step.Label,
			Minutes: step.Minutes,
			Done:    state.IsDone(step.ID),
		})
	}
	return st, p, sess, state, steps, nil
}

func (s *server) handleSession(w http.ResponseWriter, r *http.Request) {
	st, p, sess, state, steps, err := s.sessionState("案内", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	// 案内を始めるので、ここで振り返りファイルを用意する（読むだけの画面では作らない）
	if err := sess.EnsureLog(); err != nil {
		s.fail(w, st, "書けません", err.Error())
		return
	}
	step, ok := state.Current(steps)
	if !ok {
		st.Finished = true
		s.render(w, "session", st)
		return
	}
	if err := state.Start(step.ID); err != nil {
		s.fail(w, st, "書けません", err.Error())
		return
	}
	for i := range st.StepViews {
		if st.StepViews[i].Label == step.Label {
			st.StepViews[i].Current = true
			st.StepIndex = i + 1
			if i > 0 {
				st.PrevStepID = steps[i-1].ID
			}
		}
	}
	st.FirstTime = IsFirstTime(s.root, sess)
	st.Step, st.StepIDStr = step, string(step.ID)
	st.Elapsed = state.Elapsed(step.ID)
	st.ShowHint = r.URL.Query().Get("hint") == "1"
	st.StuckOpen = r.URL.Query().Get("stuck") == "1"

	if sh := s.stepSheet(p, step); sh != nil {
		st.Sheet = sh
		if _, _, err := sh.EnsureWork(); err != nil {
			s.fail(w, st, "書けません", err.Error())
			return
		}
		st.TaskDone, st.TaskTotal = sh.Counts()
		st.TaskPercent = percent(st.TaskDone, st.TaskTotal)
		link := s.linkRewriter(sh.Path)
		if n, onlyLater := state.NextTask(sh); n > 0 {
			if err := state.OpenTask(sh.Unit, n); err != nil {
				s.fail(w, st, "書けません", err.Error())
				return
			}
			st.Task = &taskView{N: n, HTML: template.HTML(renderMarkdown(sh.Tasks[n], link, nil))}
			st.OnlyLater = onlyLater
			st.TaskMins = state.TaskMinutes(sh.Unit, n)
			if advice := taskAdvice(st.TaskMins); advice != "" {
				st.Advice = template.HTML(renderMarkdown(advice, nil, nil))
				st.StuckOpen = st.TaskMins >= 45
			}
			st.TermHint = "コマンドは、自分のターミナルで打つ。この画面では打てない"
			st.NotesPath = rel(s.root, filepath.Join(sh.Dir(), "notes.md"))
			if raw, err := os.ReadFile(filepath.Join(sh.Dir(), "notes.md")); err == nil {
				st.NotesBody = template.HTML(renderMarkdown(string(raw), link, nil))
			}
			if body := sh.Section("平日に読むもの"); body != "" {
				st.Reading = template.HTML(renderMarkdown(body, link, nil))
			}
			if st.ShowHint {
				for _, name := range []string{"キーワード", "詰まりやすいところ", "平日に読むもの"} {
					if body := sh.Section(name); body != "" {
						st.Hints = append(st.Hints, sectionView{Name: name, HTML: template.HTML(renderMarkdown(body, link, nil))})
					}
				}
			}
			s.render(w, "session", st)
			return
		}
		// 課題が全部できた枠
		st.Guide = template.HTML(renderMarkdown(
			fmt.Sprintf("%s の課題は、全部できました（%d/%d）。\n\n完了条件は、終わりの手続きで確かめる。",
				sh.Unit, st.TaskDone, st.TaskTotal), nil, nil))
		s.render(w, "session", st)
		return
	}
	if step.ID == StepWarmup {
		st.LastStuck = lastStuckNotes(sess)
	}
	st.Guide = template.HTML(renderMarkdown(s.guideText(p, sess, step), s.linkRewriter(filepath.Join(s.root, "x")), nil))
	s.render(w, "session", st)
}

func (s *server) handleSessionStep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/session", http.StatusSeeOther)
		return
	}
	st, _, sess, err := s.base("", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	state := LoadState(s.root, sess)
	id := StepID(r.FormValue("step"))
	var serr error
	switch r.FormValue("action") {
	case "finish":
		serr = state.Finish(id)
	case "reopen":
		serr = state.Reopen(id)
	}
	if serr != nil {
		s.fail(w, st, "書けません", serr.Error())
		return
	}
	http.Redirect(w, r, "/session", http.StatusSeeOther)
}

func (s *server) handleSessionTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/session", http.StatusSeeOther)
		return
	}
	st, p, sess, err := s.base("", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	state := LoadState(s.root, sess)
	unit := r.FormValue("unit")
	n, _ := strconv.Atoi(r.FormValue("n"))
	sh, err := LoadSheet(s.root, unit)
	if err != nil || n <= 0 {
		http.Redirect(w, r, "/session", http.StatusSeeOther)
		return
	}
	// 書き込みの失敗は、そのまま見せる。黙って消えると、
	// 学習者は「記録できた」と思ったまま先へ進んでしまう。
	var serr error
	switch r.FormValue("action") {
	case "done":
		serr = sh.Tick(n, true)
	case "later":
		serr = state.Skip(unit, n)
	case "stuck":
		if memo := strings.TrimSpace(r.FormValue("memo")); memo != "" {
			if serr = sess.EnsureLog(); serr == nil {
				serr = AppendStuck(s.root, sess.Log, fmt.Sprintf("%s 課題%d: %s", unit, n, memo))
			}
		}
		if serr == nil {
			serr = state.Skip(unit, n)
		}
	case "hint":
		http.Redirect(w, r, "/session?hint=1", http.StatusSeeOther)
		return
	case "finish-step":
		id := StepMain
		if unit == p.DBUnit && unit != p.Unit {
			id = StepDB
		}
		serr = state.Finish(id)
	}
	if serr != nil {
		s.fail(w, st, "記録できません", serr.Error())
		return
	}
	http.Redirect(w, r, "/session", http.StatusSeeOther)
}

// ---------------------------------------------------------------- 終わりの手続き

func (s *server) handleEnd(w http.ResponseWriter, r *http.Request) {
	st, p, sess, _, _, err := s.sessionState("終わりの手続き", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	if err := sess.EnsureLog(); err != nil {
		s.fail(w, st, "書けません", err.Error())
		return
	}
	st.Fields = RetroFields(sess.Log)
	st.LogPath = rel(s.root, sess.Log)
	if v := r.FormValue("excluded"); v != "" {
		st.Excluded = strings.Split(v, "\n")
	}
	state := LoadState(s.root, sess)
	if sh, err := LoadSheet(s.root, p.Unit); err == nil {
		st.Sheet = sh
		st.Leftover = state.LeftoverTasks(sh)
		st.TaskDone, st.TaskTotal = sh.Counts()
		st.TaskPercent = percent(st.TaskDone, st.TaskTotal)
		for _, f := range sh.KeepFiles() {
			st.Keep = append(st.Keep, keepView{Name: f, Have: sh.HaveKeep(f)})
		}
		st.CanFinishUnit = st.TaskTotal > 0 && st.TaskDone == st.TaskTotal &&
			len(sh.MissingFiles()) == 0 && !p.IsDone(p.Unit)
	}
	if dirty, err := gitOut(s.root, "status", "--porcelain"); err == nil {
		st.Dirty = strings.TrimRight(dirty, "\n")
	}
	if ahead, err := gitOut(s.root, "log", "--oneline", "@{u}.."); err == nil && strings.TrimSpace(ahead) != "" {
		st.Ahead = len(strings.Split(strings.TrimSpace(ahead), "\n"))
	}
	s.render(w, "end", st)
}

// handleEndCheck は、そのユニットの確認コマンドを走らせて、結果をそのまま見せる。
func (s *server) handleEndCheck(w http.ResponseWriter, r *http.Request) {
	st, p, _, _, _, err := s.sessionState("確認", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	sh, err := LoadSheet(s.root, p.Unit)
	if err != nil {
		s.fail(w, st, "見つかりません", err.Error())
		return
	}
	cmds, note, dir := checkPlan(s.root, sh)
	var b strings.Builder
	if note != "" {
		b.WriteString(note + "\n")
	}
	for _, c := range cmds {
		fmt.Fprintf(&b, "$ %s\n", strings.Join(c, " "))
		out, err := runIn(dir, c)
		b.WriteString(out)
		if err != nil {
			b.WriteString("→ 失敗（出力を読んで直す）\n")
		} else {
			b.WriteString("→ 通りました\n")
		}
	}
	if b.Len() == 0 {
		b.WriteString("走らせるものがありません。課題シートの「完了条件の確かめ方」を見る。")
	}
	st.CheckCmd = sh.Unit
	st.CheckOut = b.String()
	st.Fields = RetroFields(st.Session.Log)
	st.LogPath = rel(s.root, st.Session.Log)
	st.Sheet = sh
	st.TaskDone, st.TaskTotal = sh.Counts()
	st.TaskPercent = percent(st.TaskDone, st.TaskTotal)
	for _, f := range sh.KeepFiles() {
		st.Keep = append(st.Keep, keepView{Name: f, Have: sh.HaveKeep(f)})
	}
	st.CanFinishUnit = st.TaskTotal > 0 && st.TaskDone == st.TaskTotal &&
		len(sh.MissingFiles()) == 0 && !p.IsDone(p.Unit)
	if dirty, gerr := gitOut(s.root, "status", "--porcelain"); gerr == nil {
		st.Dirty = strings.TrimRight(dirty, "\n")
	}
	s.render(w, "end", st)
}

func (s *server) handleEndDone(w http.ResponseWriter, r *http.Request) {
	st, p, _, err := s.base("", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	unit := strings.ToUpper(strings.TrimSpace(r.FormValue("unit")))
	// 画面は揃ったときだけボタンを出すが、古いタブや直接の POST でも同じ条件で断る
	sh, err := LoadSheet(s.root, unit)
	if err != nil {
		s.fail(w, st, "記録できません", err.Error())
		return
	}
	done, total := sh.Counts()
	if missing := sh.MissingFiles(); total == 0 || done != total || len(missing) > 0 {
		s.fail(w, st, "まだ完了にできません",
			fmt.Sprintf("%s は課題 %d/%d、足りないファイル: %s", unit, done, total, strings.Join(missing, "、")))
		return
	}
	if p.IsDone(unit) {
		s.fail(w, st, "記録できません", unit+" はすでに完了になっています")
		return
	}
	ok, err := p.MarkDone(unit)
	if err != nil {
		// 書けなかった理由を、行が無いことにすり替えない（直す場所を間違える）
		s.fail(w, st, "記録できません", err.Error())
		return
	}
	if !ok {
		s.fail(w, st, "記録できません", fmt.Sprintf("PROGRESS.md に「- [ ] %s ...」の行がありません", unit))
		return
	}
	if next := strings.ToUpper(strings.TrimSpace(r.FormValue("next"))); next != "" {
		if err := p.SetNext(next, strings.HasPrefix(next, "D")); err != nil {
			s.fail(w, st, unit+" は完了にしましたが、次のユニットを記録できません", err.Error())
			return
		}
	}
	http.Redirect(w, r, "/end", http.StatusSeeOther)
}

func (s *server) handleEndCommit(w http.ResponseWriter, r *http.Request) {
	st, p, _, err := s.base("", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	action := r.FormValue("action")
	if action == "commit" || action == "commit-push" {
		msg := strings.TrimSpace(r.FormValue("message"))
		if msg == "" {
			s.fail(w, st, "コミットできません", "何をやったかを、ひとことで書いてください")
			return
		}
		// 完了にして「次」を書き換えたあとでも、コミット文はやっていたユニットの名前にする
		unit := strings.ToUpper(strings.TrimSpace(r.FormValue("unit")))
		if !unitIDRe.MatchString(unit) {
			unit = p.Unit
		}
		excluded, err := stageAll(s.root)
		if err != nil {
			s.fail(w, st, "git add が失敗しました", err.Error())
			return
		}
		if out, err := gitOut(s.root, "commit", "-m", fmt.Sprintf("[no-ai] %s: %s", unit, msg)); err != nil {
			s.fail(w, st, "コミットが失敗しました", out)
			return
		}
		if len(excluded) > 0 {
			http.Redirect(w, r, "/end?excluded="+url.QueryEscape(strings.Join(excluded, "\n")), http.StatusSeeOther)
			return
		}
	}
	if action == "push" || action == "commit-push" {
		if out, err := gitOut(s.root, "push", "origin", "HEAD"); err != nil {
			s.fail(w, st, "push が失敗しました", out)
			return
		}
	}
	http.Redirect(w, r, "/end", http.StatusSeeOther)
}

// handleNote は、ブラウザからメモを1行足す（エディタに移らなくても書けるように）。
func (s *server) handleNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/session", http.StatusSeeOther)
		return
	}
	st, _, _, err := s.base("", "session")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	unit := r.FormValue("unit")
	text := strings.TrimSpace(r.FormValue("text"))
	if text != "" {
		// ユニットが読めないときに黙って捨てると、書いたつもりのメモが消える
		sh, err := LoadSheet(s.root, unit)
		if err != nil {
			s.fail(w, st, "メモを書けません", err.Error())
			return
		}
		if err := sh.AppendNote(text); err != nil {
			s.fail(w, st, "書けません", err.Error())
			return
		}
	}
	http.Redirect(w, r, backPath(r.FormValue("back"), "/session"), http.StatusSeeOther)
}

// handleHelp は、この画面の使い方。
func (s *server) handleHelp(w http.ResponseWriter, r *http.Request) {
	st, _, _, err := s.base("使い方", "help")
	if err != nil {
		s.fail(w, nil, "読めません", err.Error())
		return
	}
	s.render(w, "help", st)
}
