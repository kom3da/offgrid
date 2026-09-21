package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// セッションのどこまで進んだかを、1日ぶん覚えておく。
// CLI（run）とブラウザ（serve）が、同じファイルを読む。

type StepID string

const (
	StepWarmup    StepID = "warmup"
	StepMain      StepID = "main"
	StepDB        StepID = "db"
	StepAfternoon StepID = "afternoon"
	StepRetro     StepID = "retro"
	StepEnd       StepID = "end"
)

type StepState struct {
	Done    bool      `json:"done"`
	Started time.Time `json:"started,omitempty"`
	Ended   time.Time `json:"ended,omitempty"`
}

type SessionState struct {
	Date    string               `json:"date"`
	Steps   map[StepID]StepState `json:"steps"`
	Skipped map[string][]int     `json:"skipped,omitempty"` // ユニットID → あとで回した課題
	TaskAt  map[string]time.Time `json:"task_at,omitempty"` // 「ユニット:番号」→ その課題を開いた時刻
	path    string
	root    string
	// 読めなかった理由（「まだ無い」は nil）。空の記録で上書きしないために覚えておく。
	readErr error
}

// Step は、その回の並びの1つ。
type Step struct {
	ID      StepID
	Label   string
	Minutes string
}

// sessionSteps は、その回の並び。ここが「今日の流れ」の唯一の出どころで、
// ターミナル・ブラウザ・今日の一覧は、すべてこれを見る。
//
// 最初の6回は短くする（docs/04 の「段階的な導入」）。3時間の日に3.5時間のメインを
// 出してしまうと、初回からいきなり時間割が破れるため、DBと午後の枠はフルの回から。
func sessionSteps(p *Progress, s *Session) []Step {
	plan := StagePlan(p.Stage)
	main, full := plan.Main, s.Number > 6
	switch {
	case s.Number <= 3:
		main = "2時間"
	case s.Number <= 6:
		main = "3時間"
	}
	steps := []Step{
		{StepWarmup, "ウォームアップ", "30分"},
		{StepMain, "メイン（" + p.Unit + "）", main},
	}
	if full && plan.DB != "" {
		label := "PostgreSQL"
		if p.DBUnit != "" {
			label += "（" + p.DBUnit + "）"
		}
		steps = append(steps, Step{StepDB, label, plan.DB})
	}
	if full {
		steps = append(steps, Step{StepAfternoon, s.Afternoon(), "1.5時間"})
	}
	steps = append(steps,
		Step{StepRetro, "振り返り", "30分"},
		Step{StepEnd, "終わりの手続き", ""},
	)
	return steps
}

// sessionFlow は「今日の流れ」を、人が読む形で返す。
// 同じ並びを3か所で組み立てていたせいで食い違ったので、1つにまとめた。
func sessionFlow(root string, p *Progress, s *Session) []string {
	var out []string
	for _, step := range sessionSteps(p, s) {
		head := step.Label
		if step.Minutes != "" {
			head += "（" + step.Minutes + "）"
		}
		switch step.ID {
		case StepWarmup:
			head += "：" + s.Warmup()
		case StepMain:
			if sh, err := LoadSheet(root, p.Unit); err == nil {
				head += "：" + rel(root, sh.Path)
			}
		case StepDB:
			if p.DBUnit == "" {
				head += "：（PROGRESS.md の「次のPostgreSQLユニット」を D1 に書き換える）"
			}
		case StepRetro:
			head += "：" + rel(root, s.Log)
		case StepEnd:
			continue // 流れの一覧には出さない（案内の中で出てくる）
		}
		out = append(out, head)
	}
	return out
}

func statePath(root string, s *Session) string {
	dir := filepath.Dir(s.Log)
	base := strings.TrimSuffix(filepath.Base(s.Log), ".md")
	return filepath.Join(dir, "."+base+".state.json")
}

func LoadState(root string, s *Session) *SessionState {
	return loadStateFile(root, statePath(root, s))
}

func loadStateFile(root, path string) *SessionState {
	st := &SessionState{
		Date:    time.Now().Format("2006-01-02"),
		Steps:   map[StepID]StepState{},
		Skipped: map[string][]int{},
		TaskAt:  map[string]time.Time{},
		path:    path,
		root:    root,
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			st.readErr = err
		}
		return st
	}
	var got SessionState
	if err := json.Unmarshal(raw, &got); err != nil {
		return st
	}
	if got.Steps == nil {
		got.Steps = map[StepID]StepState{}
	}
	if got.Skipped == nil {
		got.Skipped = map[string][]int{}
	}
	if got.TaskAt == nil {
		got.TaskAt = map[string]time.Time{}
	}
	got.path = path
	got.root = root
	return &got
}

// Reload は、ファイルの今の中身を取り込む。ターミナルの run は state を持ち続けるので、
// 課題を出すたびにこれを呼び、ブラウザ側で「あとで」にした課題を見落とさないようにする。
func (st *SessionState) Reload() {
	fresh := loadStateFile(st.root, st.path)
	if fresh.readErr != nil {
		return // 読めないときは、手元の記録をそのままにする
	}
	st.Date, st.Steps, st.Skipped, st.TaskAt = fresh.Date, fresh.Steps, fresh.Skipped, fresh.TaskAt
}

// mutate は、ファイルから読み直したものに変更を当てて書き、その結果をメモリにも取り込む。
// ターミナル（run）とブラウザ（serve）は同じファイルを共有していて、片方が持っている
// 古い state を丸ごと書くと、もう片方が記録した「あとで」「詰まった」「枠の済み」が消える（実際に起きた）。
func (st *SessionState) mutate(fn func(*SessionState)) error {
	return withLock(st.root, func() error {
		fresh := loadStateFile(st.root, st.path)
		// 読めない理由が「まだ無い」以外なら、書かない。空の記録で上書きすると、
		// あとで回した課題も済んだ枠も消える（権限エラーで実際に消えた）。
		if fresh.readErr != nil {
			return fmt.Errorf("進みの記録を読めません（上書きを避けて中止しました）: %w", fresh.readErr)
		}
		fn(fresh)
		if err := fresh.save(); err != nil {
			return err
		}
		st.Date, st.Steps, st.Skipped, st.TaskAt = fresh.Date, fresh.Steps, fresh.Skipped, fresh.TaskAt
		return nil
	})
}

func (st *SessionState) save() error {
	raw, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(st.path), 0o755); err != nil {
		return err
	}
	return writeFile(st.path, append(raw, '\n'))
}

func (st *SessionState) IsDone(id StepID) bool { return st.Steps[id].Done }

func (st *SessionState) Start(id StepID) error {
	return st.mutate(func(st *SessionState) {
		s := st.Steps[id]
		if s.Started.IsZero() {
			s.Started = time.Now()
		}
		st.Steps[id] = s
	})
}

func (st *SessionState) Finish(id StepID) error {
	return st.mutate(func(st *SessionState) {
		s := st.Steps[id]
		s.Done, s.Ended = true, time.Now()
		if s.Started.IsZero() {
			s.Started = s.Ended
		}
		st.Steps[id] = s
	})
}

func (st *SessionState) Reopen(id StepID) error {
	return st.mutate(func(st *SessionState) {
		s := st.Steps[id]
		s.Done = false
		s.Ended = time.Time{}
		st.Steps[id] = s
	})
}

// Current は、まだ終わっていない最初の枠を返す。全部終わっていれば false。
func (st *SessionState) Current(steps []Step) (Step, bool) {
	for _, s := range steps {
		if !st.IsDone(s.ID) {
			return s, true
		}
	}
	return Step{}, false
}

func (st *SessionState) Skip(unit string, n int) error {
	return st.mutate(func(st *SessionState) {
		if st.isSkipped(unit, n) {
			return
		}
		st.Skipped[unit] = append(st.Skipped[unit], n)
	})
}

func (st *SessionState) isSkipped(unit string, n int) bool {
	for _, v := range st.Skipped[unit] {
		if v == n {
			return true
		}
	}
	return false
}

// NextTask は、その枠で次に出す課題を返す。あとで回したものは後ろへ送る。
// 2つ目の戻り値は「あとで回した分だけが残っている」かどうか。
func (st *SessionState) NextTask(sh *Sheet) (int, bool) {
	todo := sh.Todo()
	for _, n := range todo {
		if !st.isSkipped(sh.Unit, n) {
			return n, false
		}
	}
	if len(todo) > 0 {
		return todo[0], true
	}
	return 0, false
}

// OpenTask は、その課題を開いた時刻を覚える（同じ課題なら上書きしない）。
// 書けなかったら、時間の助言が出ないだけなので、呼ぶ側は知らせるだけでよい。
func (st *SessionState) OpenTask(unit string, n int) error {
	return st.mutate(func(st *SessionState) {
		key := fmt.Sprintf("%s:%d", unit, n)
		if _, ok := st.TaskAt[key]; ok {
			return
		}
		for k := range st.TaskAt {
			if strings.HasPrefix(k, unit+":") {
				delete(st.TaskAt, k)
			}
		}
		st.TaskAt[key] = time.Now()
	})
}

// TaskMinutes は、その課題を開いてから経った分数を返す。
func (st *SessionState) TaskMinutes(unit string, n int) int {
	at, ok := st.TaskAt[fmt.Sprintf("%s:%d", unit, n)]
	if !ok {
		return 0
	}
	return int(time.Since(at).Minutes())
}

// LeftoverTasks は、あとで回したまま残っている課題を返す。
func (st *SessionState) LeftoverTasks(sh *Sheet) []int {
	var out []int
	for _, n := range sh.Todo() {
		if st.isSkipped(sh.Unit, n) {
			out = append(out, n)
		}
	}
	return out
}

// Elapsed は、その枠を始めてから経った時間を「12分」の形で返す。
func (st *SessionState) Elapsed(id StepID) string {
	s := st.Steps[id]
	if s.Started.IsZero() {
		return ""
	}
	end := time.Now()
	if !s.Ended.IsZero() {
		end = s.Ended
	}
	m := int(end.Sub(s.Started).Minutes())
	if m < 1 {
		return "1分未満"
	}
	if m < 60 {
		return fmt.Sprintf("%d分", m)
	}
	return fmt.Sprintf("%d時間%d分", m/60, m%60)
}
