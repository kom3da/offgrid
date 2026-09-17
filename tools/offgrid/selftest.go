package main

import (
	"fmt"
	"os"
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
		if !strings.Contains(sh.Section("課題"), "本題") {
			problems = append(problems, name+": 「本題」と書かれた課題がありません")
		}
		// 番号なしの「本題」が2つ以上あると、どれが核か分からなくなる（CLAUDE.md の書式）
		if n := strings.Count(sh.Section("課題"), "**本題**"); n > 1 {
			problems = append(problems, fmt.Sprintf("%s: 番号なしの「**本題**」が%d個あります（複数あるなら「**本題（その1）**」と番号を振る）", name, n))
		}
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
