package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleSheet = `---
title: "F3 テキスト処理"
---

> トラック：[F：基礎](../../../docs/10-track-f-foundation.md) ／ Stage 0 ／ 目安：メイン枠で2回 ／ 前提：F2

## ねらい

小さなコマンドをパイプでつなぐ。

## このユニットで身につけること

- 条件に合う行だけを取り出す

## キーワード

- **パイプ（` + "`|`" + `）**：左の出力を右の入力につなぐ

## 平日に読むもの

- 『新しいLinuxの教科書』CHAPTER11（標準入出力とパイプライン）

## 準備

` + "```sh" + `
mkdir -p ~/sandbox/f03
` + "```" + `

## 課題

1. ` + "`wc -l`" + ` でログを眺める
2. **本題**：上位10件を1行で出す
3. 次の3つを出す
   - ステータスコードごとの件数
   - 404になったパスの上位5件
4. ` + "`sed`" + ` で置き換える
5. ` + "`awk`" + ` で合計を出す

### 発展（任意）

- 別の順で表示する

## 詰まりやすいところ

- 結果が多い → ` + "`man uniq`" + ` を読む

## 完了条件の確かめ方

課題2を、何も見ずに1行で書ける。

## 残すもの

- 各課題のコマンドと結果（` + "`notes.md`" + `）
- 課題8の表（` + "`explain.md`" + `）
- ` + "`access.log`" + ` はコミットしない

この ` + "`TASKS.md`" + ` はカリキュラムの一部なので、書き換えない。

## 次につながるユニット

- F5（シェルスクリプト）
`

const sampleProgress = `# 進捗

## 現在

- ステージ：Stage 1
- 次のユニット：F3
- 次のPostgreSQLユニット：D1

## ユニット

### F：基礎（Stage 0）

- [x] F1 コンピュータの仕組み — 2026-09-01
- [ ] F2 シェル入門
- [ ] F3 テキスト処理

### D：PostgreSQL（Stage 1〜3）

- [ ] D1 データベースとは・psql
`

const sampleRetro = `# AIなしデー 振り返り YYYY-MM-DD

## やったこと
- メインドリル（ユニット）：
- PostgreSQL（ユニット）：

## 詰まったこと（手順のどこまで試したか、45分で打ち切ったか）
-

## 今日わかったこと（自分の言葉で）
-
`

// newRepo は、検査用の小さな学習リポジトリを作る。
func newRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mk := func(rel, body string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("PROGRESS.md", sampleProgress)
	mk("docs/01-overview.md", "---\ntitle: \"offgrid とは\"\n---\n\nはじめに。\n")
	mk("templates/retrospective.md", sampleRetro)
	mk("drills/foundation/F03-pipeline/TASKS.md", sampleSheet)
	mk("drills/sql/D01-psql/TASKS.md", strings.Replace(sampleSheet, `title: "F3 テキスト処理"`, `title: "D1 データベースとは・psql"`, 1))
	return root
}

func TestFindRepo(t *testing.T) {
	root := newRepo(t)
	deep := filepath.Join(root, "drills", "foundation", "F03-pipeline")
	got, err := findRepo(deep)
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("findRepo = %q, want %q", got, root)
	}
	if _, err := findRepo(t.TempDir()); err == nil {
		t.Fatal("リポジトリでない場所で、エラーになっていない")
	}
}

func TestParseSheet(t *testing.T) {
	root := newRepo(t)
	sh, err := LoadSheet(root, "f3")
	if err != nil {
		t.Fatal(err)
	}
	if sh.Unit != "F3" {
		t.Errorf("Unit = %q", sh.Unit)
	}
	if sh.Title != "F3 テキスト処理" {
		t.Errorf("Title = %q", sh.Title)
	}
	if !strings.HasPrefix(sh.Header, "トラック：") {
		t.Errorf("Header = %q", sh.Header)
	}
	if got := len(sh.TaskNums()); got != 5 {
		t.Fatalf("課題の数 = %d, want 5（%v）", got, sh.TaskNums())
	}
	if !strings.Contains(sh.Tasks[3], "404になったパスの上位5件") {
		t.Errorf("続きの行を取り込めていない: %q", sh.Tasks[3])
	}
	if !strings.Contains(sh.Tasks[2], "本題") {
		t.Errorf("本題の印がない: %q", sh.Tasks[2])
	}
	for _, name := range requiredSections {
		if sh.Section(name) == "" {
			t.Errorf("節がない: %s", name)
		}
	}
	if sh.Section("発展（任意）") == "" {
		t.Error("発展（任意）の節を読めていない")
	}
	// access.log は「コミットしない」と書かれているので、残すものには数えない
	want := []string{"explain.md", "notes.md"}
	got := sh.KeepFiles()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("KeepFiles = %v, want %v", got, want)
	}
	if len(sh.MissingFiles()) != 2 {
		t.Errorf("MissingFiles = %v（まだ何も書いていないので2つ）", sh.MissingFiles())
	}
	if err := os.WriteFile(filepath.Join(sh.Dir(), "notes.md"), []byte("メモ"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := sh.MissingFiles(); len(got) != 1 || got[0] != "explain.md" {
		t.Errorf("notes.md を書いたあとの MissingFiles = %v", got)
	}
}

func TestWorkRecord(t *testing.T) {
	root := newRepo(t)
	sh, err := LoadSheet(root, "F3")
	if err != nil {
		t.Fatal(err)
	}
	created, added, err := sh.EnsureWork()
	if err != nil || !created || len(added) != 0 {
		t.Fatalf("EnsureWork: created=%v added=%v err=%v", created, added, err)
	}
	if created, added, _ := sh.EnsureWork(); created || len(added) != 0 {
		t.Errorf("2回目で作り直している（created=%v added=%v）", created, added)
	}
	if done, total := sh.Counts(); done != 0 || total != 5 {
		t.Fatalf("Counts = %d/%d, want 0/5", done, total)
	}
	if err := sh.Tick(2, true); err != nil {
		t.Fatal(err)
	}
	if done, total := sh.Counts(); done != 1 || total != 5 {
		t.Errorf("Tick 後の Counts = %d/%d", done, total)
	}
	if todo := sh.Todo(); len(todo) != 4 || todo[0] != 1 {
		t.Errorf("Todo = %v", todo)
	}
	if err := sh.Tick(2, false); err != nil {
		t.Fatal(err)
	}
	if done, _ := sh.Counts(); done != 0 {
		t.Errorf("外したあとの done = %d", done)
	}
	// 学習者が手で書き換えても読めること
	raw, _ := os.ReadFile(sh.WorkPath())
	if err := os.WriteFile(sh.WorkPath(), []byte(strings.Replace(string(raw), "- [ ] 1.", "- [x] 1.", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	if done, _ := sh.Counts(); done != 1 {
		t.Errorf("手で付けたチェックを読めていない（done=%d）", done)
	}
}

func TestProgress(t *testing.T) {
	root := newRepo(t)
	p, err := LoadProgress(root)
	if err != nil {
		t.Fatal(err)
	}
	if p.Stage != "1" || p.Unit != "F3" || p.DBUnit != "D1" {
		t.Fatalf("Stage=%q Unit=%q DBUnit=%q", p.Stage, p.Unit, p.DBUnit)
	}
	if len(p.Tracks) != 2 || len(p.Tracks[0].Units) != 3 {
		t.Fatalf("Tracks = %+v", p.Tracks)
	}
	if !p.IsDone("F1") || p.IsDone("F2") {
		t.Error("IsDone の判定が違う")
	}
	ok, err := p.MarkDone("F2")
	if err != nil || !ok {
		t.Fatalf("MarkDone: ok=%v err=%v", ok, err)
	}
	if again, _ := LoadProgress(root); !again.IsDone("F2") {
		t.Error("MarkDone が書き込まれていない")
	}
	raw, _ := os.ReadFile(p.Path)
	if !strings.Contains(string(raw), "- [x] F2 シェル入門 — ") {
		t.Errorf("日付が付いていない:\n%s", raw)
	}
	if err := p.SetNext("F3", false); err != nil {
		t.Fatal(err)
	}
	if err := p.SetNext("D2", true); err != nil {
		t.Fatal(err)
	}
	again, _ := LoadProgress(root)
	if again.Unit != "F3" || again.DBUnit != "D2" {
		t.Errorf("SetNext 後: Unit=%q DBUnit=%q", again.Unit, again.DBUnit)
	}
}

func TestSessionAndRetro(t *testing.T) {
	root := newRepo(t)
	s, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	if s.Number != 1 {
		t.Errorf("Number = %d, want 1", s.Number)
	}
	// 読むだけでは、振り返りファイルを作らない（作ると、回数が1つずれる）
	if _, err := os.Stat(s.Log); err == nil {
		t.Fatal("LoadSession が振り返りファイルを作っている")
	}
	if err := s.EnsureLog(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.Log); err != nil {
		t.Fatalf("EnsureLog で用意されていない: %v", err)
	}
	// 2回目に呼んでも、同じ日なら番号は増えない
	again, err := LoadSession(root)
	if err != nil || again.Number != 1 {
		t.Fatalf("同じ日の2回目で Number = %d", again.Number)
	}
	fields := RetroFields(s.Log)
	if len(fields) != 4 {
		t.Fatalf("空欄の数 = %d（%+v）", len(fields), fields)
	}
	if err := RetroWrite(s.Log, fields[0].Key, "F3をやった"); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(s.Log)
	if !strings.Contains(string(raw), "メインドリル（ユニット）：F3をやった") {
		t.Errorf("ラベルの欄に書けていない:\n%s", raw)
	}
	if len(RetroFields(s.Log)) != 3 {
		t.Error("埋めた欄が、まだ空欄として数えられている")
	}
	if err := AppendStuck(s.Log, "sort の順が合わない"); err != nil {
		t.Fatal(err)
	}
	if err := AppendStuck(s.Log, "uniq が効かない"); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(s.Log)
	body := string(raw)
	stuckAt := strings.Index(body, "## 詰まったこと")
	nextAt := strings.Index(body[stuckAt:], "## 今日わかったこと") + stuckAt
	section := body[stuckAt:nextAt]
	if !strings.Contains(section, "sort の順が合わない") || !strings.Contains(section, "uniq が効かない") {
		t.Errorf("詰まりメモが正しい節に入っていない:\n%s", section)
	}
	if strings.Contains(section, "\n-\n") {
		t.Errorf("空の「-」が残っている:\n%s", section)
	}
}

func TestStagePlanAndSessionShape(t *testing.T) {
	if StagePlan("0").DB != "" || StagePlan("1").DB == "" || StagePlan("4").DB != "" {
		t.Error("ステージごとの時間割が違う")
	}
	// Stage 3 はDBがメイン枠に合流する（別枠を立てない）
	if p3 := StagePlan("3"); p3.DB != "" || !p3.DBInMain {
		t.Errorf("Stage 3 の扱いが違う: %+v", p3)
	}
	// 短縮の回は、時間割が1日に収まること
	short := sessionSteps(&Progress{Stage: "0", Unit: "F3"}, &Session{Number: 1})
	for _, st := range short {
		if st.ID == StepAfternoon || st.ID == StepDB {
			t.Errorf("短縮の回に %s の枠が出ている", st.ID)
		}
	}
	full := sessionSteps(&Progress{Stage: "1", Unit: "G1", DBUnit: "D1"}, &Session{Number: 7})
	var ids []StepID
	for _, st := range full {
		ids = append(ids, st.ID)
	}
	if len(ids) != 6 {
		t.Errorf("フルの回の枠 = %v", ids)
	}
	s := &Session{Number: 1}
	if s.Afternoon() != "コードリーディング" || !strings.Contains(s.Finish(), "12:00") {
		t.Error("第1回の午後と終わりが違う")
	}
	s2 := &Session{Number: 8, Past: []string{"a", "b", "c", "d", "e", "f", "g"}}
	if s2.Afternoon() != "デバッグドリル" || !strings.Contains(s2.Finish(), "16:00") {
		t.Error("第8回の午後と終わりが違う")
	}
	if s2.ReviewLog() != "d" {
		t.Errorf("4回以上前のログ = %q, want d", s2.ReviewLog())
	}
}

func TestFindGaps(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	sh, _ := LoadSheet(root, "F3")
	if _, _, err := sh.EnsureWork(); err != nil {
		t.Fatal(err)
	}
	for _, n := range sh.TaskNums() {
		if err := sh.Tick(n, true); err != nil {
			t.Fatal(err)
		}
	}
	gaps := FindGaps(root, p, s)
	found := false
	for _, g := range gaps {
		if strings.HasPrefix(g, "F3: 課題は全部できている") {
			found = true
		}
	}
	if !found {
		t.Errorf("「課題は全部できている」が出ていない: %v", gaps)
	}
}

func TestWidthAndWrap(t *testing.T) {
	if width("あい") != 4 || width("ab") != 2 || width("あa") != 3 {
		t.Error("全角の幅を数えられていない")
	}
	if width("\x1b[32mあ\x1b[0m") != 2 {
		t.Error("色の指定を除いていない")
	}
	if got := pad("あいう", 10); width(got) != 10 {
		t.Errorf("pad の幅 = %d", width(got))
	}
	if got := pad("あいうえおかきくけこ", 6); width(got) > 6 {
		t.Errorf("pad が縮めていない: %q（幅%d）", got, width(got))
	}
	termWidth = 20
	defer func() { termWidth = 88 }()
	lines := wrap("あいうえおかきくけこさしすせそ", "  ")
	if len(lines) < 2 {
		t.Fatalf("折り返せていない: %v", lines)
	}
	for _, l := range lines {
		if width(l) > 20 {
			t.Errorf("行が長すぎる（幅%d）: %q", width(l), l)
		}
	}
}

// TestRealCurriculum は、このリポジトリの本物の課題シートを読む。
// 課題シートの書式を変えてツールを直し忘れたら、ここで落ちる。
func TestRealCurriculum(t *testing.T) {
	root, err := findRepo("../..")
	if err != nil {
		t.Skip("本物のリポジトリの中ではないので飛ばす")
	}
	sheets, err := AllSheets(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) < 10 {
		t.Fatalf("課題シートが %d 本しか読めていない", len(sheets))
	}
	for _, sh := range sheets {
		name := rel(root, sh.Path)
		if !strings.HasPrefix(sh.Title, sh.Unit+" ") {
			t.Errorf("%s: title %q が %q で始まっていない", name, sh.Title, sh.Unit)
		}
		if len(sh.TaskNums()) < 5 {
			t.Errorf("%s: 課題が %d 問しか読めていない", name, len(sh.TaskNums()))
		}
		for _, sec := range requiredSections {
			if sh.Section(sec) == "" {
				t.Errorf("%s: 「%s」の節を読めていない", name, sec)
			}
		}
		if len(sh.KeepFiles()) == 0 {
			t.Errorf("%s: 「残すもの」からファイル名を読めていない", name)
		}
	}
}

func TestWrapKeepsWordsWhole(t *testing.T) {
	termWidth = 40
	got := wrap("メイン（G17）：drills/go/G17-database/TASKS.md　＋ PostgreSQL（D13）", "  ")
	for _, line := range got {
		if strings.HasSuffix(strings.TrimRight(line, " "), "PostgreSQ") {
			t.Errorf("英単語の途中で折り返している:\n%s", strings.Join(got, "\n"))
		}
	}
	joined := strings.ReplaceAll(strings.Join(got, ""), " ", "")
	if !strings.Contains(joined, "PostgreSQL") {
		t.Errorf("PostgreSQL が分断されている:\n%s", strings.Join(got, "\n"))
	}
	if !strings.Contains(joined, "drills/go/G17-database/TASKS.md") {
		t.Errorf("パスが分断されている:\n%s", strings.Join(got, "\n"))
	}
}
