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
}

// Step は、その回の並び。DBの枠は Stage 1〜2 だけ。
type Step struct {
	ID      StepID
	Label   string
	Minutes string
}

func sessionSteps(p *Progress, s *Session) []Step {
	plan := StagePlan(p.Stage)
	steps := []Step{
		{StepWarmup, "ウォームアップ", "30分"},
		{StepMain, "メイン（" + p.Unit + "）", plan.Main},
	}
	if plan.DB != "" {
		label := "PostgreSQL"
		if p.DBUnit != "" {
			label += "（" + p.DBUnit + "）"
		}
		steps = append(steps, Step{StepDB, label, plan.DB})
	}
	steps = append(steps,
		Step{StepAfternoon, s.Afternoon(), "1.5時間"},
		Step{StepRetro, "振り返り", "30分"},
		Step{StepEnd, "終わりの手続き", ""},
	)
	return steps
}

func statePath(root string, s *Session) string {
	dir := filepath.Dir(s.Log)
	base := strings.TrimSuffix(filepath.Base(s.Log), ".md")
	return filepath.Join(dir, "."+base+".state.json")
}

func LoadState(root string, s *Session) *SessionState {
	path := statePath(root, s)
	st := &SessionState{
		Date:    time.Now().Format("2006-01-02"),
		Steps:   map[StepID]StepState{},
		Skipped: map[string][]int{},
		TaskAt:  map[string]time.Time{},
		path:    path,
	}
	raw, err := os.ReadFile(path)
	if err != nil {
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
	return &got
}

func (st *SessionState) save() error {
	raw, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(st.path, append(raw, '\n'), 0o644)
}

func (st *SessionState) IsDone(id StepID) bool { return st.Steps[id].Done }

func (st *SessionState) Start(id StepID) error {
	s := st.Steps[id]
	if s.Started.IsZero() {
		s.Started = time.Now()
	}
	st.Steps[id] = s
	return st.save()
}

func (st *SessionState) Finish(id StepID) error {
	s := st.Steps[id]
	s.Done, s.Ended = true, time.Now()
	if s.Started.IsZero() {
		s.Started = s.Ended
	}
	st.Steps[id] = s
	return st.save()
}

func (st *SessionState) Reopen(id StepID) error {
	s := st.Steps[id]
	s.Done = false
	s.Ended = time.Time{}
	st.Steps[id] = s
	return st.save()
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
	for _, v := range st.Skipped[unit] {
		if v == n {
			return nil
		}
	}
	st.Skipped[unit] = append(st.Skipped[unit], n)
	return st.save()
}

func (st *SessionState) ClearSkips(unit string) error {
	delete(st.Skipped, unit)
	return st.save()
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
func (st *SessionState) OpenTask(unit string, n int) {
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
	_ = st.save()
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
