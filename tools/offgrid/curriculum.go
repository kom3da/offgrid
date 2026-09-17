package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// トラックの頭文字と、課題シートの置き場所。
var trackDir = map[string]string{
	"F": "foundation", "G": "go", "T": "ts", "D": "sql", "O": "infra", "C": "capstone",
}

// ---------------------------------------------------------------- リポジトリ

// findRepo は、今いるディレクトリから親をたどって学習用リポジトリの根を探す。
func findRepo(start string) (string, error) {
	if env := os.Getenv("OFFGRID_REPO"); env != "" {
		start = env
	}
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		_, e1 := os.Stat(filepath.Join(dir, "PROGRESS.md"))
		s2, e2 := os.Stat(filepath.Join(dir, "docs"))
		if e1 == nil && e2 == nil && s2.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("学習用リポジトリが見つかりません（PROGRESS.md のあるディレクトリで実行してください）")
		}
		dir = parent
	}
}

// ---------------------------------------------------------------- 課題シート

type Sheet struct {
	Path     string
	Unit     string
	Title    string
	Header   string
	Sections map[string]string
	Tasks    map[int]string
	nums     []int
	root     string
}

func (s *Sheet) Dir() string      { return filepath.Dir(s.Path) }
func (s *Sheet) WorkPath() string { return filepath.Join(s.Dir(), "work.md") }
func (s *Sheet) TaskNums() []int  { return s.nums }

func (s *Sheet) Section(name string) string { return strings.TrimSpace(s.Sections[name]) }

// 「残すもの」の書き方は `notes.md`、`lab/main.go`、`capstone/cmd/agent/` の3通り。
// ディレクトリは末尾の / で見分ける。
var keepFileRe = regexp.MustCompile("`([A-Za-z0-9._/-]+(?:\\.(?:md|sh|sql|go|ts|tsx|js|json|ya?ml|txt|csv|log)|/))`")

// KeepFiles は「残すもの」に書かれたファイル名を返す。
// 「コミットしない」と書かれている行（作り直せるもの）は数えない。
func (s *Sheet) KeepFiles() []string {
	skip := map[string]bool{"TASKS.md": true, "TASKS.local.md": true, "work.md": true}
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(s.Section("残すもの"), "\n") {
		if strings.Contains(line, "コミットしない") {
			continue
		}
		for _, m := range keepFileRe.FindAllStringSubmatch(line, -1) {
			if !skip[m[1]] && !seen[m[1]] {
				seen[m[1]] = true
				out = append(out, m[1])
			}
		}
	}
	sort.Strings(out)
	return out
}

// keepPath は「残すもの」の名前を、実際のパスに直す。
// 課題ディレクトリの中を先に見て、無ければリポジトリの根から見る（`capstone/` などのため）。
func (s *Sheet) keepPath(name string) string {
	inDir := filepath.Join(s.Dir(), name)
	if _, err := os.Stat(inDir); err == nil {
		return inDir
	}
	if atRoot := filepath.Join(s.root, name); !strings.HasPrefix(name, "/") {
		if _, err := os.Stat(atRoot); err == nil {
			return atRoot
		}
	}
	return inDir
}

// HaveKeep は、その「残すもの」が揃っているか。ディレクトリは、中身があるかを見る。
func (s *Sheet) HaveKeep(name string) bool {
	st, err := os.Stat(s.keepPath(name))
	if err != nil {
		return false
	}
	if st.IsDir() {
		entries, err := os.ReadDir(s.keepPath(name))
		return err == nil && len(entries) > 0
	}
	return st.Size() > 0
}

// MissingFiles は、まだ無い（または空の）「残すもの」を返す。
func (s *Sheet) MissingFiles() []string {
	var out []string
	for _, f := range s.KeepFiles() {
		if !s.HaveKeep(f) {
			out = append(out, f)
		}
	}
	return out
}

var (
	frontRe   = regexp.MustCompile(`(?s)\A---\ntitle: "(.*?)"\n---\n`)
	headingRe = regexp.MustCompile(`^(#{2,3}) (.+)$`)
	taskRe    = regexp.MustCompile(`^(\d+)\. (.*)$`)
	unitIDRe  = regexp.MustCompile(`^([FGTDOCfgtdoc])0*(\d+)$`)
)

func parseSheet(path, unit, root string) (*Sheet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(raw)
	sh := &Sheet{Path: path, Unit: unit, Sections: map[string]string{}, Tasks: map[int]string{}, root: root}
	if m := frontRe.FindStringSubmatch(text); m != nil {
		sh.Title = m[1]
		text = text[len(m[0]):]
	}
	body := strings.TrimLeft(text, "\n")
	if strings.HasPrefix(body, ">") {
		sh.Header = strings.TrimSpace(strings.TrimLeft(strings.SplitN(body, "\n", 2)[0], "> "))
	}
	name, inTasks := "", false
	var buf []string
	flush := func() {
		if name != "" {
			sh.Sections[name] = strings.TrimSpace(strings.Join(buf, "\n"))
		}
	}
	for _, line := range strings.Split(body, "\n") {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			flush()
			name, buf = strings.TrimSpace(m[2]), nil
			inTasks = name == "課題"
			continue
		}
		buf = append(buf, line)
		if !inTasks {
			continue
		}
		if m := taskRe.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[1])
			if _, dup := sh.Tasks[n]; !dup {
				sh.nums = append(sh.nums, n)
			}
			sh.Tasks[n] = m[2]
		} else if len(sh.nums) > 0 && strings.HasPrefix(line, "   ") {
			last := sh.nums[len(sh.nums)-1]
			sh.Tasks[last] += "\n" + strings.TrimRight(line, " ")
		}
	}
	flush()
	sort.Ints(sh.nums)
	return sh, nil
}

// LoadSheet は F3 や d01 のような指定から、課題シートを読む。
func LoadSheet(root, unit string) (*Sheet, error) {
	m := unitIDRe.FindStringSubmatch(strings.TrimSpace(unit))
	if m == nil {
		return nil, fmt.Errorf("ユニットIDの形が違います: %q（例: F3, G12, D1）", unit)
	}
	letter := strings.ToUpper(m[1])
	num, _ := strconv.Atoi(m[2])
	base := filepath.Join(root, "drills", trackDir[letter])
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("%s のトラックがありません", letter)
	}
	prefix := fmt.Sprintf("%s%02d-", letter, num)
	var found []string // 同じ番号のディレクトリ（改名されると、古いものが残る）
	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(strings.ToUpper(e.Name()), prefix) {
			continue
		}
		for _, name := range []string{"TASKS.md", "TASKS.local.md"} {
			if _, err := os.Stat(filepath.Join(base, e.Name(), name)); err == nil {
				found = append(found, filepath.Join(base, e.Name(), name))
				break
			}
		}
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("%s%d の課題シートが見つかりません", letter, num)
	}
	// 黙ってどちらかを選ぶと、work.md やメモが片方に取り残される
	if len(found) > 1 {
		var dirs []string
		for _, f := range found {
			dirs = append(dirs, rel(root, filepath.Dir(f)))
		}
		sort.Strings(dirs)
		return nil, fmt.Errorf("%s%d のディレクトリが%d個あります（%s）。中身を1つにまとめて、古いほうを消してください",
			letter, num, len(found), strings.Join(dirs, "、"))
	}
	return parseSheet(found[0], fmt.Sprintf("%s%d", letter, num), root)
}

// AllSheets は、あるトラックのすべての課題シートを読む。
func AllSheets(root string) ([]*Sheet, error) {
	var out []*Sheet
	for letter, dir := range trackDir {
		entries, err := os.ReadDir(filepath.Join(root, "drills", dir))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			m := regexp.MustCompile(`^([FGTDOC])(\d+)-`).FindStringSubmatch(strings.ToUpper(e.Name()))
			if m == nil || m[1] != letter {
				continue
			}
			num, _ := strconv.Atoi(m[2])
			sh, err := LoadSheet(root, fmt.Sprintf("%s%d", letter, num))
			if err == nil {
				out = append(out, sh)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// ---------------------------------------------------------------- 作業記録（work.md）

// workLine は、チェックリストの1行を作る。
func (s *Sheet) workLine(n int) string {
	first := strings.SplitN(s.Tasks[n], "\n", 2)[0]
	return fmt.Sprintf("- [ ] %d. %s", n, first)
}

// EnsureWork は、課題シートからチェックリストを作る。
// すでにあるときは、シートに増えた課題の行だけを足す。カリキュラムを取り込んで課題が増えても、
// 番号がそろっていないと「できた」を記録できなくなるため。
func (s *Sheet) EnsureWork() (created bool, added []int, err error) {
	raw, readErr := os.ReadFile(s.WorkPath())
	if readErr != nil {
		var b strings.Builder
		fmt.Fprintf(&b, "# %s 作業記録\n\n", s.Title)
		b.WriteString("課題の進み具合。`offgrid run` が、ここを見て次の課題を出す。\n")
		b.WriteString("自分で書き換えてもよい。メモは `notes.md` に書く。\n\n")
		for _, n := range s.nums {
			b.WriteString(s.workLine(n) + "\n")
		}
		return true, nil, writeFile(s.WorkPath(), []byte(b.String()))
	}
	lines := strings.Split(string(raw), "\n")
	have := map[int]bool{}
	last := -1
	for i, line := range lines {
		if m := workLineRe.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[2])
			have[n] = true
			last = i
		}
	}
	for _, n := range s.nums {
		if !have[n] {
			added = append(added, n)
		}
	}
	if len(added) == 0 {
		return false, nil, nil
	}
	var fresh []string
	for _, n := range added {
		fresh = append(fresh, s.workLine(n))
	}
	if last < 0 {
		lines = append(lines, fresh...)
	} else {
		lines = append(lines[:last+1], append(fresh, lines[last+1:]...)...)
	}
	return false, added, writeFile(s.WorkPath(), []byte(strings.Join(lines, "\n")))
}

var workLineRe = regexp.MustCompile(`^- \[([ x])\] (\d+)\.`)

// WorkState は、課題ごとのチェックの状態を返す。
func (s *Sheet) WorkState() map[int]bool {
	out := map[int]bool{}
	raw, err := os.ReadFile(s.WorkPath())
	if err != nil {
		return out
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if m := workLineRe.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[2])
			out[n] = m[1] == "x"
		}
	}
	return out
}

func (s *Sheet) Tick(n int, done bool) error {
	if _, _, err := s.EnsureWork(); err != nil {
		return err
	}
	raw, err := os.ReadFile(s.WorkPath())
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	mark := "- [x]"
	if !done {
		mark = "- [ ]"
	}
	found := false
	for i, line := range lines {
		if m := workLineRe.FindStringSubmatch(line); m != nil {
			if num, _ := strconv.Atoi(m[2]); num == n {
				lines[i] = mark + line[len("- [ ]"):]
				found = true
			}
		}
	}
	// 黙って書けないままにすると、同じ課題が何度も出てくる。
	if !found {
		return fmt.Errorf("%s に「- [ ] %d.」の行がありません（消したか、書き換えた可能性があります）", rel(s.root, s.WorkPath()), n)
	}
	return writeFile(s.WorkPath(), []byte(strings.Join(lines, "\n")))
}

// Counts は「できた数 / 全体」を返す。
// 全体は、いつも課題シートの数を正とする（work.md の行が消えていても、数がずれないように）。
func (s *Sheet) Counts() (int, int) {
	st := s.WorkState()
	done := 0
	for _, n := range s.nums {
		if st[n] {
			done++
		}
	}
	return done, len(s.nums)
}

// Todo は、まだできていない課題の番号を小さい順に返す。
func (s *Sheet) Todo() []int {
	st := s.WorkState()
	var out []int
	for _, n := range s.nums {
		if ok, seen := st[n]; !seen || !ok {
			out = append(out, n)
		}
	}
	return out
}

// ---------------------------------------------------------------- PROGRESS.md

type TrackProgress struct {
	Name  string
	Units []UnitProgress
}

type UnitProgress struct {
	ID   string
	Done bool
}

type Progress struct {
	Path    string
	Stage   string
	Unit    string
	DBUnit  string
	Tracks  []TrackProgress
	rawText string
}

var (
	stageRe  = regexp.MustCompile(`^- ステージ：\s*(.+)$`)
	nextRe   = regexp.MustCompile(`^- 次のユニット：\s*(\S+)`)
	nextDBRe = regexp.MustCompile(`^- 次のPostgreSQLユニット：\s*(\S+)`)
	trackRe  = regexp.MustCompile(`^### (.+)$`)
	boxRe    = regexp.MustCompile(`^- \[([ x])\] ([FGTDOC]\d+)\b`)
	digitRe  = regexp.MustCompile(`\d`)
)

func LoadProgress(root string) (*Progress, error) {
	path := filepath.Join(root, "PROGRESS.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("PROGRESS.md が読めません: %w", err)
	}
	p := &Progress{Path: path, Stage: "0", rawText: string(raw)}
	cur := -1
	for _, line := range strings.Split(p.rawText, "\n") {
		switch {
		case stageRe.MatchString(line):
			if d := digitRe.FindString(stageRe.FindStringSubmatch(line)[1]); d != "" {
				p.Stage = d
			}
		case nextRe.MatchString(line):
			p.Unit = nextRe.FindStringSubmatch(line)[1]
		case nextDBRe.MatchString(line):
			v := nextDBRe.FindStringSubmatch(line)[1]
			if unitIDRe.MatchString(v) {
				p.DBUnit = v
			}
		case trackRe.MatchString(line):
			p.Tracks = append(p.Tracks, TrackProgress{Name: strings.TrimSpace(trackRe.FindStringSubmatch(line)[1])})
			cur = len(p.Tracks) - 1
		case boxRe.MatchString(line):
			if cur < 0 {
				p.Tracks = append(p.Tracks, TrackProgress{Name: "ユニット"})
				cur = 0
			}
			m := boxRe.FindStringSubmatch(line)
			p.Tracks[cur].Units = append(p.Tracks[cur].Units, UnitProgress{ID: m[2], Done: m[1] == "x"})
		}
	}
	if p.Unit == "" {
		return nil, fmt.Errorf("PROGRESS.md の「現在」に「- 次のユニット：F1」の行を書いてください")
	}
	return p, nil
}

func (p *Progress) IsDone(unit string) bool {
	for _, t := range p.Tracks {
		for _, u := range t.Units {
			if strings.EqualFold(u.ID, unit) && u.Done {
				return true
			}
		}
	}
	return false
}

func (p *Progress) MarkDone(unit string) (bool, error) {
	raw, err := os.ReadFile(p.Path)
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(raw), "\n")
	prefix := "- [ ] " + strings.ToUpper(unit) + " "
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = "- [x]" + line[len("- [ ]"):] + fmt.Sprintf(" — %s", time.Now().Format("2006-01-02"))
			return true, writeFile(p.Path, []byte(strings.Join(lines, "\n")))
		}
	}
	return false, nil
}

func (p *Progress) SetNext(unit string, db bool) error {
	label := "次のユニット"
	if db {
		label = "次のPostgreSQLユニット"
	}
	raw, err := os.ReadFile(p.Path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "- "+label+"：") {
			lines[i] = "- " + label + "：" + unit
			return writeFile(p.Path, []byte(strings.Join(lines, "\n")))
		}
	}
	// 行が無ければ「次のユニット」の下に足す
	for i, line := range lines {
		if strings.HasPrefix(line, "- 次のユニット：") {
			lines = append(lines[:i+1], append([]string{"- " + label + "：" + unit}, lines[i+1:]...)...)
			return writeFile(p.Path, []byte(strings.Join(lines, "\n")))
		}
	}
	return fmt.Errorf("PROGRESS.md に「現在」の欄がありません")
}

// ---------------------------------------------------------------- セッション

type Session struct {
	Number int
	Log    string
	Past   []string
	root   string
}

func LoadSession(root string) (*Session, error) {
	now := time.Now()
	logDir := filepath.Join(root, "logs", now.Format("2006"))
	log := filepath.Join(logDir, now.Format("01-02")+".md")
	var past []string
	base := filepath.Join(root, "logs")
	_ = filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(p, ".md") {
			return nil
		}
		if strings.HasPrefix(filepath.Base(p), ".") || p == log {
			return nil
		}
		past = append(past, p)
		return nil
	})
	sort.Strings(past)
	return &Session{Number: len(past) + 1, Log: log, Past: past, root: root}, nil
}

// EnsureLog は、その日の振り返りファイルを用意する。
// 読むだけのコマンド（find、status など）からは呼ばない。読んだだけでファイルができると、
// 以後それが「済んだセッション」として数えられて、回数や時間割がずれてしまうため。
func (s *Session) EnsureLog() error {
	if _, err := os.Stat(s.Log); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.Log), 0o755); err != nil {
		return err
	}
	tmpl, err := os.ReadFile(filepath.Join(s.root, "templates", "retrospective.md"))
	if err != nil {
		return fmt.Errorf("templates/retrospective.md がありません: %w", err)
	}
	return writeFile(s.Log, tmpl)
}

func (s *Session) Afternoon() string {
	if s.Number%2 == 1 {
		return "コードリーディング"
	}
	return "デバッグドリル"
}

func (s *Session) Warmup() string {
	if len(s.Past) == 0 {
		return "初回なので、再現するものがない。代わりに「Ubuntuのシェルを開き、`cd ~/offgrid-log` で移動し、`ls` で中身を見る」を、何も見ずにやる"
	}
	if s.Number%2 == 1 {
		return "前回のセッションでやったことを、何も見ずに再現する"
	}
	return "4セッション以上前にやったことを、何も見ずに再現する"
}

func (s *Session) ReviewLog() string {
	if len(s.Past) == 0 {
		return ""
	}
	if s.Number%2 == 1 {
		return s.Past[len(s.Past)-1]
	}
	i := len(s.Past) - 4
	if i < 0 {
		i = 0
	}
	return s.Past[i]
}

func (s *Session) Finish() string {
	switch {
	case s.Number <= 3:
		return "12:00まで（3時間。最後の30分で振り返り）"
	case s.Number <= 6:
		return "14:00まで（4時間。最後の30分で振り返り）"
	default:
		return "16:00まで（6時間）"
	}
}

type Plan struct{ Main, DB string }

// StagePlan は、そのステージの1セッションの時間の配分（docs/03-roadmap.md）。
// PostgreSQL は Stage 1〜3 のあいだ、毎回1時間の枠で進める。
func StagePlan(stage string) Plan {
	switch stage {
	case "1", "2", "3":
		return Plan{Main: "2.5時間", DB: "1時間"}
	default:
		return Plan{Main: "3.5時間"}
	}
}

// ---------------------------------------------------------------- 振り返り

type RetroField struct {
	Line  int    // いま何行目か（画面の id に使うだけ。書き込みには使わない）
	Key   string // 見出しと項目名で決まる、行がずれても変わらない名前
	Label string
}

var (
	emptyBulletRe = regexp.MustCompile(`^- *$`)
	emptyNumRe    = regexp.MustCompile(`^(\d+)\. *$`)
	labelRe       = regexp.MustCompile(`：\s*$`)
)

// RetroFields は、テンプレートのまま空いている欄を返す。
func RetroFields(log string) []RetroField {
	raw, err := os.ReadFile(log)
	if err != nil {
		return nil
	}
	var out []RetroField
	heading := ""
	seen := map[string]int{}
	add := func(i int, label string) {
		k := heading + "|" + label
		seen[k]++
		out = append(out, RetroField{Line: i, Key: fmt.Sprintf("%s|%d", k, seen[k]), Label: label})
	}
	for i, line := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(line, "#"):
			heading = strings.TrimSpace(strings.TrimLeft(line, "# "))
		case emptyBulletRe.MatchString(line), emptyNumRe.MatchString(line):
			label := heading
			if label == "" {
				label = "（見出しなし）"
			}
			add(i, label)
		case labelRe.MatchString(line):
			add(i, strings.TrimLeft(strings.TrimSpace(line), "- "))
		}
	}
	return out
}

// RetroWrite は、まだ空いている欄に書き込む。
// 行番号ではなく見出しと項目名で探すので、書いている間にファイルへ1行入っても、別の行を壊さない。
// すでに埋まっている欄には書かない（二重送信で同じ文が2回入るのを防ぐ）。
func RetroWrite(log, key, text string) error {
	index := -1
	for _, f := range RetroFields(log) {
		if f.Key == key {
			index = f.Line
			break
		}
	}
	if index < 0 {
		return fmt.Errorf("その欄は、もう埋まっています（画面を読み込み直してください）")
	}
	raw, err := os.ReadFile(log)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	if index >= len(lines) {
		return fmt.Errorf("行が見つかりません")
	}
	line := lines[index]
	switch {
	case emptyBulletRe.MatchString(line):
		lines[index] = "- " + text
	case emptyNumRe.MatchString(line):
		lines[index] = emptyNumRe.FindStringSubmatch(line)[1] + ". " + text
	case labelRe.MatchString(line):
		lines[index] = strings.TrimRight(line, " ") + text
	default:
		return fmt.Errorf("その欄は、もう埋まっています")
	}
	return writeFile(log, []byte(strings.Join(lines, "\n")))
}

// AppendStuck は、詰まりメモを「詰まったこと」の欄に時刻付きで書く。
func AppendStuck(log, text string) error {
	raw, err := os.ReadFile(log)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	entry := fmt.Sprintf("- %s %s", time.Now().Format("15:04"), text)
	for i, line := range lines {
		if !strings.HasPrefix(line, "## 詰まったこと") {
			continue
		}
		j := i + 1
		for j < len(lines) && !strings.HasPrefix(lines[j], "#") {
			j++
		}
		k := j - 1
		for k > i && strings.TrimSpace(lines[k]) == "" {
			k--
		}
		if emptyBulletRe.MatchString(lines[k]) {
			lines[k] = entry
		} else {
			lines = append(lines[:k+1], append([]string{entry}, lines[k+1:]...)...)
		}
		return writeFile(log, []byte(strings.Join(lines, "\n")))
	}
	lines = append(lines, entry)
	return writeFile(log, []byte(strings.Join(lines, "\n")))
}

// AppendNote は、そのユニットの notes.md に、時刻付きで1行足す。
// ブラウザからでも書けるようにして、エディタに移る手間を減らす。
func (s *Sheet) AppendNote(text string) error {
	path := filepath.Join(s.Dir(), "notes.md")
	body := ""
	if raw, err := os.ReadFile(path); err == nil {
		body = string(raw)
	} else {
		body = fmt.Sprintf("# %s メモ\n\n課題をやりながら気づいたこと、試したコマンドと結果を書く。\n", s.Title)
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += fmt.Sprintf("\n- %s %s\n", time.Now().Format("15:04"), strings.TrimSpace(text))
	return writeFile(path, []byte(body))
}

// IsFirstTime は「まだ1回もセッションをやっていない」かどうか。
// はじめての人に、何をすればよいかを出すために使う。
func IsFirstTime(root string, s *Session) bool {
	if len(s.Past) > 0 {
		return false
	}
	matches, _ := filepath.Glob(filepath.Join(root, "drills", "*", "*", "work.md"))
	return len(matches) == 0
}

// ---------------------------------------------------------------- 作業漏れ

// FindGaps は「どこまでやったか分からない」を防ぐための検査。
func FindGaps(root string, p *Progress, s *Session) []string {
	var out []string
	for _, t := range p.Tracks {
		for _, u := range t.Units {
			sh, err := LoadSheet(root, u.ID)
			if err != nil {
				continue
			}
			done, total := sh.Counts()
			miss := sh.MissingFiles()
			_, workMissing := os.Stat(sh.WorkPath())
			others := false
			if entries, err := os.ReadDir(sh.Dir()); err == nil {
				for _, e := range entries {
					if !e.IsDir() && !strings.HasPrefix(e.Name(), "TASKS") {
						others = true
					}
				}
			}
			switch {
			case u.Done && len(miss) > 0:
				out = append(out, fmt.Sprintf("%s: 完了だが、残すものが無い → %s", u.ID, strings.Join(miss, "、")))
			case !u.Done && total > 0 && done == total:
				out = append(out, fmt.Sprintf("%s: 課題は全部できている → 完了条件を確かめて『offgrid done %s』", u.ID, u.ID))
			case !u.Done && done > 0 && done < total:
				out = append(out, fmt.Sprintf("%s: 途中（課題 %d/%d）", u.ID, done, total))
			case !u.Done && others && workMissing != nil:
				out = append(out, fmt.Sprintf("%s: 書いたものがあるのに、記録が無い → 『offgrid check %s』", u.ID, u.ID))
			}
		}
	}
	type hole struct {
		path string
		n    int
	}
	var holes []hole
	for _, lg := range s.Past {
		if n := len(RetroFields(lg)); n > 0 {
			holes = append(holes, hole{lg, n})
		}
	}
	sort.Slice(holes, func(i, j int) bool { return holes[i].path > holes[j].path })
	for i, h := range holes {
		if i >= 3 {
			out = append(out, fmt.Sprintf("ほか、空欄のある振り返りが%d件", len(holes)-3))
			break
		}
		out = append(out, fmt.Sprintf("%s: 振り返りの空欄が%d個 → 『offgrid retro %s』", rel(root, h.path), h.n, rel(root, h.path)))
	}
	if dirty, err := gitOut(root, "status", "--porcelain"); err == nil && strings.TrimSpace(dirty) != "" {
		out = append(out, fmt.Sprintf("コミットしていない変更が%d件 → 『offgrid end』", len(strings.Split(strings.TrimSpace(dirty), "\n"))))
	}
	// 上流（origin/main）が無いときは err になるので、そのときは何も言わない
	if ahead, err := gitOut(root, "log", "--oneline", "@{u}.."); err == nil && strings.TrimSpace(ahead) != "" {
		out = append(out, fmt.Sprintf("push していないコミットが%d件 → git push origin main", len(strings.Split(strings.TrimSpace(ahead), "\n"))))
	}
	return out
}

func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

// writeFile は、同じ場所に一時ファイルを作って書き、最後に置き換える。
// ターミナルとブラウザを行き来しながら使うので、途中で止まっても
// 書きかけの（中身が切れた）ファイルが残らないようにする。
func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name) // 置き換えに成功していれば、もう無い
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, 0o644); err != nil {
		return err
	}
	return os.Rename(name, path)
}
