package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// 課題シートに必ずある見出し。ここが変わったら、このツールは読めなくなる。
var requiredSections = []string{
	"ねらい", "このユニットで身につけること", "キーワード", "平日に読むもの", "準備", "課題",
	"詰まりやすいところ", "完了条件の確かめ方", "残すもの", "次につながるユニット",
}

// cmdSelftest は、今のカリキュラムがこの版のツールで読めるかを検査する。
// 課題シートの書式を変えたときに、CI がここで落ちる。
func cmdSelftest(root string, args []string) error {
	quiet := false
	for _, a := range args {
		if a == "-q" || a == "--quiet" {
			quiet = true
		}
	}
	sheets, err := AllSheets(root)
	if err != nil {
		return err
	}
	if len(sheets) == 0 {
		return fmt.Errorf("課題シートが1つも見つかりません（drills/ の中を確かめてください）")
	}
	var problems []string
	for _, sh := range sheets {
		name := rel(root, sh.Path)
		if sh.Title == "" {
			problems = append(problems, name+": frontmatter の title がありません")
		} else if !strings.HasPrefix(sh.Title, sh.Unit+" ") {
			problems = append(problems, fmt.Sprintf("%s: title %q が「%s 」で始まっていません", name, sh.Title, sh.Unit))
		}
		if sh.Header == "" || !strings.HasPrefix(sh.Header, "トラック：") {
			problems = append(problems, name+": 先頭の「> トラック：…」の行がありません")
		}
		for _, sec := range requiredSections {
			if sh.Section(sec) == "" {
				problems = append(problems, fmt.Sprintf("%s: 「%s」の節がありません", name, sec))
			}
		}
		if n := len(sh.TaskNums()); n < 5 {
			problems = append(problems, fmt.Sprintf("%s: 課題が%d問しかありません（番号付きの行を読めていない可能性）", name, n))
		}
		problems = append(problems, checkTaskLines(name, sh)...)
		if len(sh.KeepFiles()) == 0 {
			problems = append(problems, name+": 「残すもの」からファイル名を読めません")
		}
	}

	// 進捗の書式も確かめる
	if p, err := LoadProgress(root); err != nil {
		problems = append(problems, "PROGRESS.md: "+err.Error())
	} else {
		total := 0
		for _, t := range p.Tracks {
			total += len(t.Units)
		}
		if total < len(sheets) {
			problems = append(problems, fmt.Sprintf("PROGRESS.md: ユニットの欄が%d個で、課題シート%d本より少ない", total, len(sheets)))
		}
		// カリキュラムを取り込んでユニットが増えたとき、PROGRESS.md は学習者のものなので
		// 置き換わらない。行が無いまま進めると、完了を記録できずに詰まる。
		have := map[string]bool{}
		for _, t := range p.Tracks {
			for _, u := range t.Units {
				have[strings.ToUpper(u.ID)] = true
			}
		}
		var missing []string
		for _, sh := range sheets {
			if !have[strings.ToUpper(sh.Unit)] {
				missing = append(missing, sh.Unit)
			}
		}
		if len(missing) > 0 {
			problems = append(problems, fmt.Sprintf(
				"PROGRESS.md: 課題シートはあるのに欄が無いユニット: %s（「- [ ] %s ...」の行を足してください）",
				strings.Join(missing, "、"), missing[0]))
		}
		if !quiet {
			fmt.Printf("PROGRESS.md: Stage %s / 次のユニット %s / ユニット欄 %d個\n", p.Stage, p.Unit, total)
		}
	}
	// 振り返りのテンプレートも読めるか
	tmpl := root + "/templates/retrospective.md"
	if _, err := os.Stat(tmpl); err != nil {
		problems = append(problems, "templates/retrospective.md がありません")
	} else if len(RetroFields(tmpl)) == 0 {
		problems = append(problems, "templates/retrospective.md から、埋める欄を読めません")
	}

	if !quiet {
		fmt.Printf("課題シート: %d本を読みました（offgrid %s）\n", len(sheets), version)
	}
	if len(problems) > 0 {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, red(fmt.Sprintf("%d件の問題が見つかりました:", len(problems))))
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "  - "+p)
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "課題シートの書式を変えたときは、tools/offgrid も直してください（CLAUDE.md の「課題シートを書くとき」）")
		return fmt.Errorf("selftest が失敗しました")
	}
	fmt.Println(green("selftest: 問題ありません"))
	return nil
}

var (
	rawTaskRe    = regexp.MustCompile(`^(\d+)\. `)
	mainTaskRe   = regexp.MustCompile(`\*\*本題\*\*`)
	mainNumberRe = regexp.MustCompile(`\*\*本題（その(\d+)）\*\*`)
)

// checkTaskLines は、「## 課題」の節を生の行で読み、解析の結果と突き合わせる。
// 解析器は、同じ番号の課題を後のほうで上書きし、課題の途中の小見出し（###）で読むのを止める。
// 解析したあとの数だけを見ていると、課題が画面から消えていても通ってしまった。
func checkTaskLines(name string, sh *Sheet) []string {
	raw, err := os.ReadFile(sh.Path)
	if err != nil {
		return []string{name + ": 読めません: " + err.Error()}
	}
	var (
		problems  []string
		nums      []int
		taskLines []string
		inTasks   bool
		inSub     bool // 「### 発展（任意）」など、課題の節の中の小見出しより後ろ
	)
	for _, line := range strings.Split(string(raw), "\n") {
		switch {
		case strings.HasPrefix(line, "## "):
			inTasks = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "課題"
			inSub = false
			continue
		case !inTasks:
			continue
		case strings.HasPrefix(line, "### "):
			inSub = true
			continue
		}
		m := rawTaskRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if inSub {
			problems = append(problems, fmt.Sprintf(
				"%s: 課題の途中に小見出し（###）があり、その後ろの課題 %s. を読めていません（小見出しは課題の最後に置く）", name, m[1]))
			continue
		}
		n, _ := strconv.Atoi(m[1])
		nums = append(nums, n)
		taskLines = append(taskLines, line)
	}

	// 1 から順に、飛ばさず、重ねずに並んでいるか
	for i, n := range nums {
		if n != i+1 {
			problems = append(problems, fmt.Sprintf(
				"%s: 課題の番号が 1 から順に並んでいません（%d 問目が %d.）。同じ番号があると、後の課題で前の課題が上書きされて画面から消えます", name, i+1, n))
			break
		}
	}
	if got := len(sh.TaskNums()); got != len(nums) && len(problems) == 0 {
		problems = append(problems, fmt.Sprintf("%s: 番号付きの行は %d 行あるのに、%d 問しか読めていません", name, len(nums), got))
	}

	// 本題は、番号付きの課題の行に付ける。前置きの文に「本題」と書いてあるだけでは、どれが核か分からない
	plain, numbered := 0, []int{}
	for _, l := range taskLines {
		plain += len(mainTaskRe.FindAllString(l, -1))
		for _, m := range mainNumberRe.FindAllStringSubmatch(l, -1) {
			k, _ := strconv.Atoi(m[1])
			numbered = append(numbered, k)
		}
	}
	switch {
	case plain == 0 && len(numbered) == 0:
		problems = append(problems, name+": 本題（**本題** か **本題（その1）**）が付いた課題がありません")
	case plain > 1:
		problems = append(problems, fmt.Sprintf("%s: 番号なしの「**本題**」が%d個あります（複数あるなら「**本題（その1）**」と番号を振る）", name, plain))
	case plain > 0 && len(numbered) > 0:
		problems = append(problems, name+": 「**本題**」と「**本題（そのN）**」が混ざっています（どちらかにそろえる）")
	}
	for i, k := range numbered {
		if k != i+1 {
			problems = append(problems, fmt.Sprintf("%s: 本題の番号が「その1」から順に並んでいません（%d つ目が「その%d」）", name, i+1, k))
			break
		}
	}
	return problems
}
