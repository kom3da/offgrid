package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 学習者のファイルを書き換えるコマンドの検査。
// 記録を壊すと取り返せないので、表示だけのコマンドより先にここを押さえる。

// withInput は、聞かれたことに順番に答える。行末の改行は足す。
func withInput(t *testing.T, answers ...string) {
	t.Helper()
	prev := stdin
	t.Cleanup(func() { stdin = prev })
	stdin = bufio.NewScanner(strings.NewReader(strings.Join(answers, "\n") + "\n"))
}

// capture は、コマンドが画面に出したものを受け取る（学習者が読む文を確かめるため）。
func capture(t *testing.T, run func()) string {
	t.Helper()
	prev := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string)
	go func() {
		var b strings.Builder
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			b.Write(buf[:n])
			if err != nil {
				break
			}
		}
		done <- b.String()
	}()
	run()
	w.Close()
	os.Stdout = prev
	return <-done
}

func read(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestCmdTickRecordsAndRefuses(t *testing.T) {
	root := newRepo(t)
	p, err := LoadProgress(root)
	if err != nil {
		t.Fatal(err)
	}
	out := capture(t, func() {
		if err := cmdTick(root, p, []string{"2"}); err != nil {
			t.Errorf("cmdTick: %v", err)
		}
	})
	if !strings.Contains(out, "1/5") {
		t.Errorf("できた数が出ていない:\n%s", out)
	}
	sh, _ := LoadSheet(root, "F3")
	if !strings.Contains(read(t, sh.WorkPath()), "- [x] 2.") {
		t.Error("work.md に記録されていない")
	}
	// 数字でないときは、そのまま通さない
	if err := cmdTick(root, p, []string{"x"}); err == nil {
		t.Error("数字でない番号を受け付けている")
	}
	// シートに無い番号は、黙って成功させない
	if err := cmdTick(root, p, []string{"99"}); err == nil {
		t.Error("シートに無い番号を、できたことにしている")
	}
}

func TestCmdDoneNeedsConfirmation(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)

	// n と答えたら、記録しない
	withInput(t, "n")
	out := capture(t, func() {
		if err := cmdDone(root, p, []string{"F3"}); err != nil {
			t.Errorf("cmdDone: %v", err)
		}
	})
	if !strings.Contains(out, "記録しませんでした") {
		t.Errorf("断ったのに記録している:\n%s", out)
	}
	if again, _ := LoadProgress(root); again.IsDone("F3") {
		t.Fatal("n と答えたのに完了になっている")
	}

	// y なら記録し、次のユニットも設定する
	withInput(t, "y", "F4")
	capture(t, func() {
		if err := cmdDone(root, p, []string{"F3"}); err != nil {
			t.Errorf("cmdDone: %v", err)
		}
	})
	again, _ := LoadProgress(root)
	if !again.IsDone("F3") {
		t.Error("完了が記録されていない")
	}
	if again.Unit != "F4" {
		t.Errorf("次のユニット = %q, want F4", again.Unit)
	}
	if !strings.Contains(read(t, again.Path), "— ") {
		t.Error("完了の日付が入っていない")
	}
}

func TestCmdRetroFillsFieldsByLabel(t *testing.T) {
	root := newRepo(t)
	s, err := LoadSession(root)
	if err != nil {
		t.Fatal(err)
	}
	before := len(RetroFields(s.Log))
	if before != 0 {
		t.Fatalf("読む前から振り返りがある（%d欄）", before)
	}
	// 1つ答えて、次は飛ばし、そのあとやめる
	withInput(t, "F3をやった", "", "q")
	capture(t, func() {
		if err := cmdRetro(root, s, nil); err != nil {
			t.Errorf("cmdRetro: %v", err)
		}
	})
	body := read(t, s.Log)
	if !strings.Contains(body, "メインドリル（ユニット）：F3をやった") {
		t.Errorf("狙った欄に入っていない:\n%s", body)
	}
	if strings.Count(body, "F3をやった") != 1 {
		t.Errorf("同じ文が複数入っている:\n%s", body)
	}
	// 見出しが壊れていないこと
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "#") && strings.Contains(line, "F3をやった") {
			t.Errorf("見出しに連結された: %q", line)
		}
	}
}

func TestCmdStuckWritesToTheRightSection(t *testing.T) {
	root := newRepo(t)
	s, _ := LoadSession(root)
	capture(t, func() {
		if err := cmdStuck(root, s, []string{"xxd", "が", "読めない"}); err != nil {
			t.Errorf("cmdStuck: %v", err)
		}
	})
	body := read(t, s.Log)
	at := strings.Index(body, "## 詰まったこと")
	next := strings.Index(body[at:], "## 今日わかったこと") + at
	if !strings.Contains(body[at:next], "xxd が 読めない") {
		t.Errorf("「詰まったこと」に入っていない:\n%s", body[at:next])
	}
	// 何も渡さないときは、聞いてから書く
	withInput(t, "sort の順が合わない")
	capture(t, func() {
		if err := cmdStuck(root, s, nil); err != nil {
			t.Errorf("cmdStuck: %v", err)
		}
	})
	if !strings.Contains(read(t, s.Log), "sort の順が合わない") {
		t.Error("聞いた内容が書かれていない")
	}
}

func TestGuideUnitPersistsSkipsAcrossRuns(t *testing.T) {
	root := newRepo(t)
	s, _ := LoadSession(root)
	if err := s.EnsureLog(); err != nil {
		t.Fatal(err)
	}
	sh, _ := LoadSheet(root, "F3")
	state := LoadState(root, s)

	// 課題1を「あとで」、2を「できた」、そこで枠を終える
	withInput(t, "l", "", "d")
	capture(t, func() {
		if err := guideUnit(sh, s, state); err != nil {
			t.Errorf("guideUnit: %v", err)
		}
	})
	if !strings.Contains(read(t, sh.WorkPath()), "- [x] 2.") {
		t.Error("「できた」が記録されていない")
	}
	// 「あとで」は state に残り、ブラウザからも同じ扱いになる
	again := LoadState(root, s)
	if !again.isSkipped("F3", 1) {
		t.Error("「あとで」が state に残っていない")
	}
	if n, onlyLater := again.NextTask(sh); n == 1 && !onlyLater {
		t.Error("あとで回した課題が、先頭に出てきている")
	}
}

// Skip 自身が書き込むことを確かめる。
// ほかの保存（OpenTask など）の副作用で残っているだけだと、
// その呼び出しが消えた日に、黙って共有されなくなる。
func TestSkipSavesOnItsOwn(t *testing.T) {
	root := newRepo(t)
	s, _ := LoadSession(root)
	if err := s.EnsureLog(); err != nil {
		t.Fatal(err)
	}
	state := LoadState(root, s)
	if err := state.Skip("F3", 4); err != nil {
		t.Fatal(err)
	}
	// ここまでで Skip 以外の保存は呼んでいない
	if !LoadState(root, s).isSkipped("F3", 4) {
		t.Error("Skip がファイルに書いていない（ブラウザと共有されない）")
	}
}

func TestCmdEndReportsWhatIsMissing(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	// 振り返りは空、残すものも無い状態
	withInput(t, "n", "")
	out := capture(t, func() {
		if err := cmdEnd(root, p, s); err != nil {
			t.Errorf("cmdEnd: %v", err)
		}
	})
	if !strings.Contains(out, "書いていない欄が") {
		t.Errorf("振り返りの空欄を知らせていない:\n%s", out)
	}
	if !strings.Contains(out, "残すものが足りない") {
		t.Errorf("残すものの不足を知らせていない:\n%s", out)
	}
	// git が無いリポジトリでも落ちない（学習者が git init 前でも使う）
	if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
		t.Fatal("検査用リポジトリに .git がある")
	}
}

func TestCmdEndCanFinishTheUnit(t *testing.T) {
	root := newRepo(t)
	p, _ := LoadProgress(root)
	s, _ := LoadSession(root)
	sh, _ := LoadSheet(root, "F3")
	// 課題を全部できたことにし、残すものも用意する
	for _, n := range sh.TaskNums() {
		if err := sh.Tick(n, true); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range sh.KeepFiles() {
		if err := os.WriteFile(filepath.Join(sh.Dir(), f), []byte("書いた"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// 振り返りは埋めない（n）→ 完了にする（y）→ 記録する（y）→ 次のユニット
	withInput(t, "n", "y", "y", "F4")
	out := capture(t, func() {
		if err := cmdEnd(root, p, s); err != nil {
			t.Errorf("cmdEnd: %v", err)
		}
	})
	if !strings.Contains(out, "残すもの: 揃っています") {
		t.Errorf("残すものが揃ったと言っていない:\n%s", out)
	}
	again, _ := LoadProgress(root)
	if !again.IsDone("F3") {
		t.Errorf("完了が記録されていない:\n%s", out)
	}
	if again.Unit != "F4" {
		t.Errorf("次のユニット = %q, want F4", again.Unit)
	}
}
