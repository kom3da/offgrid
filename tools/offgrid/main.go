// offgrid は、学習カリキュラム offgrid を進めるためのコマンド。
// AIもネットワークも使わない。答えも出さない。
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// version はリリース時に -ldflags で入れる。カリキュラムの版と同じ番号にする。
var version = "dev"

type menuItem struct {
	key, label, cmd string
}

var menu = []menuItem{
	{"1", "セッションを始める（案内つき）", "run"},
	{"2", "次の課題だけ見る", "next"},
	{"3", "課題シートを読む", "unit"},
	{"4", "調べる（オフライン検索）", "find"},
	{"5", "確認コマンドを実行する", "check"},
	{"6", "詰まりメモを書く", "stuck"},
	{"7", "振り返りを書く", "retro"},
	{"8", "終わりの手続き", "end"},
	{"9", "ブラウザで開く（serve）", "serve"},
	{"q", "やめる", ""},
}

func usage() {
	fmt.Println("offgrid — 学習カリキュラム offgrid を進めるためのコマンド")
	fmt.Println()
	fmt.Println("使い方: offgrid [サブコマンド] [引数]")
	fmt.Println("引数なしで実行すると、ダッシュボードとメニューを出す。")
	fmt.Println()
	rows := [][2]string{
		{"run", "セッションを順に案内する（課題を1問ずつ）"},
		{"today", "今日やることの一覧"},
		{"next [ID]", "次の課題だけ表示する"},
		{"tick <番号> [ID]", "課題を「できた」にする"},
		{"status", "進捗と、作業漏れの疑い"},
		{"unit [ID]", "課題シートを読む"},
		{"find <語>", "カリキュラム・課題シート・自分のメモを検索する"},
		{"check [ID]", "確認コマンドを実行し、残すものと完了条件を見る"},
		{"stuck [文]", "詰まりメモを、振り返りの「詰まったこと」に書く"},
		{"timer [分]", "残り時間を計る（既定15分）"},
		{"retro [ファイル]", "振り返りの空欄を、1つずつ聞いて埋める"},
		{"done [ID]", "完了条件を確かめて、PROGRESS.md に記録する"},
		{"end", "終わりの手続き（記入漏れ、残すもの、コミット）"},
		{"serve [--port N]", "ブラウザで開く画面を、手元に立てる（ネットワークには出さない）"},
		{"selftest", "カリキュラムの課題シートが、この版のツールで読めるかを検査する"},
		{"version", "版を表示する"},
	}
	for _, r := range rows {
		fmt.Printf("  %s%s\n", pad(r[0], 26), r[1])
	}
	fmt.Println()
	say(dim("このコマンドはAIを使わない。答えも出さない。AIに頼むときの文面は prompts/ にある"))
}

func showMenu(root string, p *Progress, s *Session) error {
	for {
		fmt.Println()
		rule("何をする？")
		for _, m := range menu {
			fmt.Printf("  %s. %s\n", bold(m.key), m.label)
		}
		keys := make([]string, 0, len(menu))
		for _, m := range menu {
			keys = append(keys, m.key)
		}
		got, err := ask("\n  番号を入れて Enter: ", append(keys, "")...)
		if err != nil {
			return err
		}
		if got == "q" || got == "" {
			return nil
		}
		var cmd string
		for _, m := range menu {
			if m.key == got {
				cmd = m.cmd
			}
		}
		fmt.Println()
		var runErr error
		switch cmd {
		case "run":
			return cmdRun(root, p, s)
		case "next":
			runErr = cmdNext(root, p, nil)
		case "unit":
			id, err := ask(fmt.Sprintf("  ユニットID（Enterで %s）: ", p.Unit))
			if err != nil {
				return err
			}
			if id == "" {
				id = p.Unit
			}
			runErr = cmdUnit(root, p, []string{id})
		case "find":
			word, err := ask("  探す言葉: ")
			if err != nil {
				return err
			}
			if word != "" {
				runErr = cmdFind(root, []string{word})
			}
		case "check":
			runErr = cmdCheck(root, p, nil)
		case "stuck":
			runErr = cmdStuck(root, s, nil)
		case "retro":
			runErr = cmdRetro(root, s, nil)
		case "end":
			runErr = cmdEnd(root, p, s)
		case "serve":
			runErr = cmdServe(root, nil)
		}
		if runErr != nil {
			if errors.As(runErr, &errAbort{}) {
				fmt.Println(dim("  中断しました"))
				continue
			}
			fmt.Println(red("  " + runErr.Error()))
		}
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help", "help":
			usage()
			return nil
		case "version", "--version", "-v":
			fmt.Println("offgrid " + version)
			return nil
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := findRepo(wd)
	if err != nil {
		return err
	}
	if len(args) > 0 && args[0] == "selftest" {
		return cmdSelftest(root, args[1:])
	}
	p, err := LoadProgress(root)
	if err != nil {
		return err
	}
	s, err := LoadSession(root)
	if err != nil {
		return err
	}
	sub := "menu"
	if !colorOn {
		sub = "today" // 端末でないとき（パイプなど）は、対話しない
	}
	if len(args) > 0 {
		sub = args[0]
		args = args[1:]
	} else {
		args = nil
	}
	switch sub {
	case "menu":
		dashboard(root, p, s)
		return showMenu(root, p, s)
	case "run":
		return cmdRun(root, p, s)
	case "today":
		return cmdToday(root, p, s)
	case "next":
		return cmdNext(root, p, args)
	case "tick":
		return cmdTick(root, p, args)
	case "status":
		return cmdStatus(root, p, s)
	case "unit":
		return cmdUnit(root, p, args)
	case "find":
		return cmdFind(root, args)
	case "check":
		return cmdCheck(root, p, args)
	case "stuck":
		return cmdStuck(root, s, args)
	case "timer":
		return cmdTimer(args)
	case "retro":
		return cmdRetro(root, s, args)
	case "done":
		return cmdDone(root, p, args)
	case "end":
		return cmdEnd(root, p, s)
	case "serve":
		return cmdServe(root, args)
	default:
		fmt.Fprintln(os.Stderr, red("error: 知らないサブコマンド: "+sub))
		fmt.Fprintln(os.Stderr)
		usage()
		return fmt.Errorf("")
	}
}

func main() {
	initUI()
	if err := run(); err != nil {
		if errors.As(err, &errAbort{}) {
			fmt.Println(dim("中断しました。続きは offgrid run"))
			return
		}
		if msg := strings.TrimSpace(err.Error()); msg != "" {
			fmt.Fprintln(os.Stderr, red("error: "+msg))
		}
		os.Exit(1)
	}
}
