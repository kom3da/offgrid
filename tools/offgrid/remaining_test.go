package main

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 「試さない」と書いていたものを、試す。対話と表示だけの関数でも、
// 学習者が読む文の退行は起きる（cmdStatus の off-by-one を今日直したが、テストが無かった）。

// ---------------------------------------------------------------- writeFile

// writeFile は、学習の記録を守る要。一時ファイルに書いてから置き換える設計で、
// 途中で止まっても中身の切れたファイルを残さない。
//
// 「途中で止まる」は起こせないので、その設計でしか起きない挙動で確かめる：
// 置き場所のディレクトリに書けないと、一時ファイルを作れずに失敗する（素の上書きなら成功する）。
// これが失敗しなくなったら、一時ファイル＋置き換えの設計が崩れている。
func TestWriteFileKeepsTheOldContentWhenItCannotStage(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root は権限を無視するので、この検査は意味を持たない")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "work.md")
	if err := os.WriteFile(path, []byte("古い\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// ディレクトリを書けなくする（ファイル自体は書ける）
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	err := writeFile(path, []byte("新しい\n"))
	if err == nil {
		t.Fatal("一時ファイルを置けない場所で成功している（素の上書きになっていないか）")
	}
	if got := read(t, path); got != "古い\n" {
		t.Errorf("失敗したのに中身が変わっている: %q", got)
	}
}

func TestWriteFileReplacesAndCleansUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "work.md")
	if err := os.WriteFile(path, []byte("古い\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(path, []byte("新しい\n")); err != nil {
		t.Fatal(err)
	}
	if got := read(t, path); got != "新しい\n" {
		t.Errorf("置き換わっていない: %q", got)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".work.md.tmp") {
			t.Errorf("一時ファイルが残っている: %s", e.Name())
		}
	}
	missing := filepath.Join(dir, "no-such-dir", "x.md")
	if err := writeFile(missing, []byte("x")); err == nil {
		t.Error("存在しないディレクトリへの書き込みが成功している")
	}
}

// ---------------------------------------------------------------- serve の失敗の経路

// ユニットが無いメモは、黙って捨てずに画面で知らせること。
func TestNoteWithUnknownUnitIsNotDroppedSilently(t *testing.T) {
	_, h, _ := newServer(t)
	w := post(t, h, "/note", url.Values{"unit": {"Z9"}, "text": {"消えてほしくないメモ"}})
	if w.Code == 303 {
		t.Fatal("ユニットが無いのに、書けたふりをして戻している")
	}
	if !strings.Contains(w.Body.String(), "メモを書けません") {
		t.Errorf("何が起きたかを伝えていない（code=%d）:\n%s", w.Code, w.Body.String())
	}
	// 空のメモは、何も書かずに戻るだけ
	w = post(t, h, "/note", url.Values{"unit": {"F3"}, "text": {"   "}})
	if w.Code != 303 {
		t.Errorf("空のメモで %d", w.Code)
	}
}

// 振り返りのテンプレートが無いときは、詰まりメモを書けないことを画面で知らせること。
func TestStuckFailsVisiblyWithoutTemplate(t *testing.T) {
	_, h, root := newServer(t)
	if err := os.Remove(filepath.Join(root, "templates", "retrospective.md")); err != nil {
		t.Fatal(err)
	}
	w := post(t, h, "/stuck", url.Values{"text": {"xxd が読めない"}})
	if w.Code == 303 {
		t.Fatal("テンプレートが無いのに、書けたふりをして戻している")
	}
	if !strings.Contains(w.Body.String(), "retrospective.md") {
		t.Errorf("原因を伝えていない:\n%s", w.Body.String())
	}
}

// ---------------------------------------------------------------- 表示だけのコマンド

func TestCmdStatusCountsSessions(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)

	// 0回
	s, _ := LoadSession(root)
	out := capture(t, func() { _ = cmdStatus(root, p, s) })
	if !strings.Contains(out, "済んだ回数: 0回") {
		t.Errorf("0回のとき:\n%s", out)
	}

	// 3回 → あと1回
	pastLogs(t, root, 3)
	s, _ = LoadSession(root)
	out = capture(t, func() { _ = cmdStatus(root, p, s) })
	flat := strings.Join(strings.Fields(out), " ")
	if !strings.Contains(flat, "済んだ回数: 3回") || !strings.Contains(flat, "あと 1回") {
		t.Errorf("3回のとき:\n%s", out)
	}

	// 4回 → 今回がその回（4-0=4 と出していた off-by-one を直した）
	pastLogs(t, root, 4)
	s, _ = LoadSession(root)
	out = capture(t, func() { _ = cmdStatus(root, p, s) })
	if !strings.Contains(out, "今回がその回") {
		t.Errorf("4回のとき、見直しの回だと言っていない:\n%s", out)
	}
	if strings.Contains(out, "あと 4回") {
		t.Errorf("4回目なのに「あと 4回」と出ている（off-by-one の再発）:\n%s", out)
	}
}

func TestCmdTodayShowsTheFlow(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	out := capture(t, func() {
		if err := cmdToday(root, p, s); err != nil {
			t.Fatal(err)
		}
	})
	flat := strings.Join(strings.Fields(out), " ")
	for _, want := range []string{"1. ウォームアップ（30分）", "2. メイン（F3）", "振り返り（30分）", "最初の6回は短くしてある", "offgrid run"} {
		if !strings.Contains(flat, want) {
			t.Errorf("「%s」が出ていない:\n%s", want, out)
		}
	}
	// 読むだけなので、振り返りファイルは作らない
	if _, err := os.Stat(s.Log); err == nil {
		t.Error("today が振り返りファイルを作っている")
	}
}

func TestCmdNextShowsTheNextTask(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	out := capture(t, func() {
		if err := cmdNext(root, p, nil); err != nil {
			t.Fatal(err)
		}
	})
	flat := strings.Join(strings.Fields(out), " ")
	if !strings.Contains(flat, "作業記録を作りました") {
		t.Errorf("初回に work.md を作ったと言っていない:\n%s", out)
	}
	if !strings.Contains(flat, "課題 1／5") || !strings.Contains(flat, "offgrid tick 1") {
		t.Errorf("次の課題と、できたときの打ち方を出していない:\n%s", out)
	}
	// 全部できたら、確認へ誘導する
	sh, _ := LoadSheet(root, "F3")
	for _, n := range sh.TaskNums() {
		if err := sh.Tick(n, true); err != nil {
			t.Fatal(err)
		}
	}
	out = capture(t, func() { _ = cmdNext(root, p, nil) })
	flat = strings.Join(strings.Fields(out), " ")
	if !strings.Contains(flat, "全部できています") || !strings.Contains(flat, "offgrid check F3") {
		t.Errorf("全部できたあとの案内:\n%s", out)
	}
}

func TestCmdFindSearchesTheCurriculum(t *testing.T) {
	root := newRepo(t)
	// シートにある語
	out := capture(t, func() {
		if err := cmdFind(root, []string{"パイプ"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "課題シート") || !strings.Contains(out, "F03-pipeline/TASKS.md") {
		t.Errorf("シートの中の語を見つけていない:\n%s", out)
	}
	// 無い語
	out = capture(t, func() { _ = cmdFind(root, []string{"存在しない語zzz"}) })
	if !strings.Contains(out, "見つかりませんでした") {
		t.Errorf("無い語のとき:\n%s", out)
	}
	// 語が無いときは使い方を返す
	if err := cmdFind(root, nil); err == nil {
		t.Error("引数なしを通している")
	}
}

func TestCmdCheckShowsKeepFiles(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	out := capture(t, func() {
		if err := cmdCheck(root, p, nil); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "[未] notes.md") || !strings.Contains(out, "[未] explain.md") {
		t.Errorf("無い残すものを [未] にしていない:\n%s", out)
	}
	if !strings.Contains(out, "完了条件の確かめ方") {
		t.Errorf("完了条件を見せていない:\n%s", out)
	}
	sh, _ := LoadSheet(root, "F3")
	if err := os.WriteFile(filepath.Join(sh.Dir(), "notes.md"), []byte("メモ"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = capture(t, func() { _ = cmdCheck(root, p, nil) })
	if !strings.Contains(out, "[済] notes.md") {
		t.Errorf("書いた残すものを [済] にしていない:\n%s", out)
	}
}

func TestCmdUnitPrintsTheSheet(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	// 端末でないときはページャを使わず、そのまま出す
	out := capture(t, func() {
		if err := cmdUnit(root, p, []string{"D1"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "D1 データベースとは・psql") {
		t.Errorf("シートの中身を出していない:\n%s", out)
	}
	if err := cmdUnit(root, p, []string{"Z9"}); err == nil {
		t.Error("無いユニットを通している")
	}
}

// ---------------------------------------------------------------- 入口

func TestMenuQuitsCleanly(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	withInput(t, "q")
	capture(t, func() {
		if err := showMenu(root, p, s); err != nil {
			t.Errorf("q で終われない: %v", err)
		}
	})
	// 番号を選ぶと、そのコマンドが動いてメニューに戻る（3=次の課題）
	withInput(t, "3", "q")
	out := capture(t, func() {
		if err := showMenu(root, p, s); err != nil {
			t.Errorf("番号を選んだあと: %v", err)
		}
	})
	if !strings.Contains(out, "課題 1／5") {
		t.Errorf("3 で次の課題が出ていない:\n%s", out)
	}
}

func TestRunDispatchesSubcommands(t *testing.T) {
	root := newRepo(t)
	t.Setenv("OFFGRID_REPO", root)
	prev := os.Args
	t.Cleanup(func() { os.Args = prev })

	os.Args = []string{"offgrid", "version"}
	out := capture(t, func() {
		if err := run(); err != nil {
			t.Errorf("version: %v", err)
		}
	})
	if !strings.Contains(out, "offgrid "+version) {
		t.Errorf("版を出していない: %q", out)
	}

	os.Args = []string{"offgrid", "today"}
	out = capture(t, func() {
		if err := run(); err != nil {
			t.Errorf("today: %v", err)
		}
	})
	if !strings.Contains(out, "今日の流れ") {
		t.Errorf("today が動いていない:\n%s", out)
	}

	os.Args = []string{"offgrid", "sonzai-shinai"}
	var err error
	out = capture(t, func() { err = run() })
	if err == nil {
		t.Error("知らないサブコマンドを通している")
	}
	if !strings.Contains(out, "知らないサブコマンド") || !strings.Contains(out, "使い方") {
		t.Errorf("何が悪いかと使い方を出していない:\n%s", out)
	}
}

// 端末の幅は、狭すぎても広すぎても読みにくいので、48〜92 に収める。
func TestInitUIClampsTheWidth(t *testing.T) {
	prevW, prevC := termWidth, colorOn
	t.Cleanup(func() { termWidth, colorOn = prevW, prevC })
	for _, tc := range []struct {
		env  string
		want int
	}{{"30", 48}, {"70", 70}, {"200", 92}, {"", 88}} {
		termWidth = 88
		t.Setenv("COLUMNS", tc.env)
		initUI()
		if termWidth != tc.want {
			t.Errorf("COLUMNS=%q → %d, want %d", tc.env, termWidth, tc.want)
		}
	}
	// テストの標準出力は端末ではないので、色は付かない
	if colorOn {
		t.Error("端末でないのに色を付けている")
	}
}
