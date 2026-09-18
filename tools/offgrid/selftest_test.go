package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// selftest は「課題シートの書式とツールの約束」を守る仕掛け。
// これが壊れたシートを通してしまうと、仕掛け自体が無意味になる。

func TestSelftestPassesOnTheFixture(t *testing.T) {
	root := newRepo(t)
	out := capture(t, func() {
		if err := cmdSelftest(root, []string{"-q"}); err != nil {
			t.Errorf("正しいシートで落ちている: %v", err)
		}
	})
	_ = out
}

func TestSelftestCatchesBrokenSheets(t *testing.T) {
	sheet := filepath.Join("drills", "foundation", "F03-pipeline", "TASKS.md")
	cases := []struct {
		name    string
		break_  func(string) string
		wantMsg string
	}{
		{"title が無い", func(s string) string {
			return strings.Replace(s, "---\ntitle: \"F3 テキスト処理\"\n---\n", "", 1)
		}, "title"},
		{"title のIDが違う", func(s string) string {
			return strings.Replace(s, `title: "F3 テキスト処理"`, `title: "F9 テキスト処理"`, 1)
		}, "で始まっていません"},
		{"トラックの行が無い", func(s string) string {
			return strings.Replace(s, "> トラック：", "トラック：", 1)
		}, "トラック"},
		{"必須の節が無い", func(s string) string {
			return strings.Replace(s, "## このユニットで身につけること", "## べつの見出し", 1)
		}, "の節がありません"},
		{"本題が無い", func(s string) string {
			return strings.Replace(s, "**本題**：", "", 1)
		}, "本題"},
		{"番号なしの本題が2つ", func(s string) string {
			return strings.Replace(s, "4. `sed` で置き換える", "4. **本題** `sed` で置き換える", 1)
		}, "番号なし"},
		{"残すものからファイル名が読めない", func(s string) string {
			// 「残すもの」の節から、バッククォートのファイル名を全部消す
			i := strings.Index(s, "## 残すもの")
			j := strings.Index(s[i:], "## 次につながるユニット") + i
			return s[:i] + "## 残すもの\n\n- 書いたものを残す\n\n" + s[j:]
		}, "残すもの"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := newRepo(t)
			path := filepath.Join(root, sheet)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			broken := tc.break_(string(raw))
			if broken == string(raw) {
				t.Fatalf("シートを壊せていない（置換が当たっていない）")
			}
			if err := os.WriteFile(path, []byte(broken), 0o644); err != nil {
				t.Fatal(err)
			}
			var got error
			out := capture(t, func() { got = cmdSelftest(root, []string{"-q"}) })
			if got == nil {
				t.Fatalf("壊れたシートを通してしまった:\n%s", out)
			}
			// 何が壊れているかは、画面（標準エラー）に出す設計
			if !strings.Contains(out, tc.wantMsg) {
				t.Errorf("何が壊れているかを伝えていない（want %q）:\n%s", tc.wantMsg, out)
			}
		})
	}
}

// カリキュラムを取り込んでユニットが増えたのに、PROGRESS.md に行が無い状態。
// PROGRESS.md は学習者のものなので取り込みで置き換わらず、実際に起こる。
func TestSelftestCatchesUnitMissingFromProgress(t *testing.T) {
	root := newRepo(t)
	p := filepath.Join(root, "PROGRESS.md")
	raw, _ := os.ReadFile(p)
	// D1 のシートはあるのに、欄だけ消す
	cut := strings.Replace(string(raw), "- [ ] D1 データベースとは・psql\n", "", 1)
	if cut == string(raw) {
		t.Fatal("欄を消せていない")
	}
	if err := os.WriteFile(p, []byte(cut), 0o644); err != nil {
		t.Fatal(err)
	}
	var got error
	out := capture(t, func() { got = cmdSelftest(root, []string{"-q"}) })
	if got == nil {
		t.Fatalf("欄が無いユニットを通してしまった:\n%s", out)
	}
	if !strings.Contains(out, "D1") || !strings.Contains(out, "欄が無い") {
		t.Errorf("どのユニットの欄が無いかを伝えていない:\n%s", out)
	}
}

// 版を取り込んでディレクトリ名が変わると、古い TASKS.md が残って同じユニットが2つになる。
// 黙って片方を抜くと「70本」で通ってしまうので、両方の名前を出して止める。
func TestSelftestRefusesTwoDirectoriesForOneUnit(t *testing.T) {
	root := newRepo(t)
	old := filepath.Join(root, "drills", "foundation", "F03-old-name")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, "TASKS.md"), []byte(sampleSheet), 0o644); err != nil {
		t.Fatal(err)
	}
	var err error
	capture(t, func() { err = cmdSelftest(root, []string{"-q"}) })
	if err == nil {
		t.Fatal("同じユニットのディレクトリが2つあるのに selftest が通った")
	}
	for _, want := range []string{"F3", "F03-old-name", "F03-pipeline", "2個"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("エラーに %q が無い: %v", want, err)
		}
	}
	// シートの無いディレクトリ（学習者のメモ置き場）は、重複に数えない
	if err := os.Remove(filepath.Join(old, "TASKS.md")); err != nil {
		t.Fatal(err)
	}
	capture(t, func() { err = cmdSelftest(root, []string{"-q"}) })
	if err == nil {
		t.Fatal("TASKS.md を消したあとも、同じ番号のディレクトリが2つあるので止まるべき")
	}
}
