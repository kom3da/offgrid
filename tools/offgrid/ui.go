package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 端末の見た目。色は端末でないときと NO_COLOR のときは付けない。
var (
	colorOn = false
	// interactive は、標準出力が端末かどうか。メニューや less を出すかはこちらで決める。
	// 色（colorOn）で代用していたので、NO_COLOR を設定すると端末でもメニューが消えていた
	interactive = false
	stdin       = bufio.NewScanner(os.Stdin)
	ansiRe      = regexp.MustCompile("\x1b\\[[0-9;]*m")
	termWidth   = 88
)

func initUI() {
	fi, err := os.Stdout.Stat()
	tty := err == nil && (fi.Mode()&os.ModeCharDevice) != 0
	interactive = tty
	colorOn = tty && os.Getenv("NO_COLOR") == ""
	if w := os.Getenv("COLUMNS"); w != "" {
		fmt.Sscanf(w, "%d", &termWidth)
	}
	if termWidth < 48 {
		termWidth = 48
	}
	if termWidth > 92 {
		termWidth = 92
	}
	stdin.Buffer(make([]byte, 0, 64*1024), 1024*1024)
}

func paint(code, text string) string {
	if !colorOn {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

func bold(t string) string   { return paint("1", t) }
func dim(t string) string    { return paint("2", t) }
func cyan(t string) string   { return paint("36", t) }
func green(t string) string  { return paint("32", t) }
func yellow(t string) string { return paint("33", t) }
func red(t string) string    { return paint("31", t) }

// 全角は2、半角は1として数える。Go の標準ライブラリには East Asian Width が無いので、
// 日本語で使う範囲を自分で持つ。
var wideRanges = [][2]rune{
	{0x1100, 0x115F}, {0x2E80, 0x303E}, {0x3041, 0x33FF}, {0x3400, 0x4DBF},
	{0x4E00, 0x9FFF}, {0xA000, 0xA4CF}, {0xAC00, 0xD7A3}, {0xF900, 0xFAFF},
	{0xFE10, 0xFE19}, {0xFE30, 0xFE6F}, {0xFF00, 0xFF60}, {0xFFE0, 0xFFE6},
	{0x1F300, 0x1F64F}, {0x1F680, 0x1F6FF}, {0x1F900, 0x1F9FF}, {0x1FA70, 0x1FAFF},
	{0x20000, 0x2FFFD}, {0x30000, 0x3FFFD},
}

func runeWidth(r rune) int {
	// 結合文字（Mn）と書式文字（Cf。絵文字をつなぐ U+200D など）は、幅を持たない
	if r == 0 || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Cf, r) {
		return 0
	}
	for _, rg := range wideRanges {
		if r >= rg[0] && r <= rg[1] {
			return 2
		}
	}
	return 1
}

func width(s string) int {
	n := 0
	for _, r := range ansiRe.ReplaceAllString(s, "") {
		n += runeWidth(r)
	}
	return n
}

// pad は表示幅を room にそろえる（長すぎるときは詰める）。
// ansiPrefix は、先頭にある色の指定。
var ansiPrefix = regexp.MustCompile("^\x1b\\[[0-9;]*m")

// tokens は、文字列を「色の指定」と「1文字」に分ける。色の指定は途中で切らない
// （文字単位で切ると、折り返しや省略が指定の途中に入り、画面に「32m」が見えた）。
func tokens(s string) []string {
	var out []string
	for len(s) > 0 {
		if m := ansiPrefix.FindString(s); m != "" {
			out = append(out, m)
			s = s[len(m):]
			continue
		}
		_, n := utf8.DecodeRuneInString(s)
		out = append(out, s[:n])
		s = s[n:]
	}
	return out
}

func isANSI(t string) bool { return strings.HasPrefix(t, "\x1b") }

func tokWidth(t string) int {
	if isANSI(t) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(t)
	return runeWidth(r)
}

func pad(s string, room int) string {
	if w := width(s); w <= room {
		return s + strings.Repeat(" ", room-w)
	}
	var b strings.Builder
	used, colored := 0, false
	for _, t := range tokens(s) {
		if isANSI(t) {
			b.WriteString(t)
			colored = true
			continue
		}
		if used+tokWidth(t) > room-1 {
			break
		}
		b.WriteString(t)
		used += tokWidth(t)
	}
	if colored {
		b.WriteString("\x1b[0m") // 切ったところで色が続かないように戻す
	}
	return b.String() + "…"
}

var bulletRe = regexp.MustCompile(`^(\s*(?:[-*]|\d+\.)\s+)`)

// wrap は全角を考えて折り返す。日本語なので単語の途中でも切る。
func wrap(text, indent string) []string {
	var out []string
	for _, para := range strings.Split(text, "\n") {
		if strings.TrimSpace(para) == "" {
			out = append(out, "")
			continue
		}
		hang := ""
		if m := bulletRe.FindStringSubmatch(para); m != nil {
			hang = strings.Repeat(" ", width(m[1]))
		}
		head := indent
		var line []string
		lineW := 0
		for _, t := range tokens(para) {
			tw := tokWidth(t)
			// 1行に使える幅は、その行の頭（字下げ）を引いた残り。2行目以降の字下げを
			// 引き忘れていたので、箇条書きの続きの行が端末の幅をはみ出していた
			limit := termWidth - width(head)
			if lineW > 0 && tw > 0 && lineW+tw > limit {
				cut := breakAt(line, t)
				// 行末に残った色の指定は、次の行の頭へ送る（前の行に残すと色が付かない）
				for cut > 0 && isANSI(line[cut-1]) {
					cut--
				}
				rest := append([]string{}, line[cut:]...)
				for len(rest) > 0 && (rest[0] == " " || rest[0] == "\t") {
					rest = rest[1:]
				}
				out = append(out, head+strings.TrimRight(strings.Join(line[:cut], ""), " \t"))
				head = indent + hang
				line = rest
				lineW = width(strings.Join(line, ""))
			}
			if len(line) == 0 && head != indent && (t == " " || t == "\t") {
				continue // 折り返した先頭の空白は捨てる
			}
			line = append(line, t)
			lineW += tw
		}
		out = append(out, head+strings.Join(line, ""))
	}
	return out
}

// breakAt は、行があふれたときに切る位置（ルーン単位）を返す。
// 日本語はどこでも切れるが、コマンド名や英単語の途中で切ると読めなくなるので、
// 英数字の連なりの先頭まで戻す。戻しすぎるときは、あきらめてそのまま切る。
func breakAt(line []string, next string) int {
	word := func(t string) bool {
		if isANSI(t) {
			return true // 色の指定は、続く語にくっつけて扱う
		}
		r, _ := utf8.DecodeRuneInString(t)
		return isWordRune(r)
	}
	if !word(next) {
		return len(line)
	}
	i := len(line)
	for i > 0 && word(line[i-1]) {
		i--
	}
	if i == 0 || len(line)-i > 24 {
		return len(line)
	}
	return i
}

// isWordRune は、途中で切りたくない文字（英数字と、識別子に使う記号）。
func isWordRune(r rune) bool {
	switch {
	case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		return true
	case r == '-' || r == '_' || r == '.' || r == '/' || r == ':':
		return true
	}
	return false
}

func say(text string) {
	for _, line := range wrap(text, "  ") {
		fmt.Println(line)
	}
}

func rule(title string) {
	if title == "" {
		fmt.Println(dim(strings.Repeat("─", termWidth)))
		return
	}
	head := "── " + title + " "
	fmt.Println(cyan(head + strings.Repeat("─", max(0, termWidth-width(head)))))
}

func banner(lines []string) {
	fmt.Println(cyan("┌" + strings.Repeat("─", termWidth-2) + "┐"))
	for _, l := range lines {
		fmt.Println(cyan("│ ") + l + strings.Repeat(" ", max(0, termWidth-4-width(l))) + cyan(" │"))
	}
	fmt.Println(cyan("└" + strings.Repeat("─", termWidth-2) + "┘"))
}

func progressBar(done, total int) string {
	const size = 16
	if total <= 0 {
		return dim(strings.Repeat("░", size))
	}
	filled := (done*size + total/2) / total
	fill := strings.Repeat("▇", filled)
	switch {
	case done >= total:
		fill = green(fill)
	case done > 0:
		fill = yellow(fill)
	default:
		fill = dim(fill)
	}
	return fill + dim(strings.Repeat("░", size-filled))
}

// errAbort は q や Ctrl+D による中断。
type errAbort struct{}

func (errAbort) Error() string { return "中断しました" }

// ask は1行を読む。allow が空でなければ、その中の1つになるまで聞き直す。
func ask(prompt string, allow ...string) (string, error) {
	for {
		fmt.Print(bold(prompt))
		if !stdin.Scan() {
			fmt.Println()
			return "", errAbort{}
		}
		got := strings.TrimSpace(stdin.Text())
		if len(allow) == 0 {
			return got, nil
		}
		low := strings.ToLower(got)
		for _, a := range allow {
			if low == a {
				return low, nil
			}
		}
		shown := make([]string, 0, len(allow))
		for _, a := range allow {
			if a == "" {
				shown = append(shown, "Enter")
			} else {
				shown = append(shown, a)
			}
		}
		fmt.Println(dim("  " + strings.Join(shown, "、") + " のどれかを入れてください"))
	}
}

func confirm(prompt string) bool {
	got, err := ask(prompt, "y", "n", "")
	return err == nil && got == "y"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
