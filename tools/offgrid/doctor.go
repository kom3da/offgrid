package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// doctor は、学習の環境が整っているかを検査し、足りないものと、直すコマンドを1つ出す。
// 道具を勝手に入れない（入れるかどうかは学習者が決める）。ネットワークも使わない。
// 学習用リポジトリがまだ無い Day 0 の途中でも動くように、リポジトリの外でも実行できる。

// doctorEnv は、外の世界への窓口。テストでは差し替えて、道具が欠けた状態を作る。
type doctorEnv struct {
	run  func(name string, args ...string) (string, error)
	look func(name string) (string, error)
	read func(path string) ([]byte, error)
	goos string // runtime.GOOS。Ubuntu の外にいるとき、どこにいるかを言うため
}

var hostEnv = doctorEnv{
	run: func(name string, args ...string) (string, error) {
		// 固まった docker などで、検査そのものが止まらないようにする
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
		return strings.TrimSpace(string(out)), err
	},
	look: exec.LookPath,
	read: os.ReadFile,
	goos: runtime.GOOS,
}

// finding は、1つの検査の結果。
type finding struct {
	ok     bool
	warn   bool   // 動くが、カリキュラムが確かめた形と違う（問題には数えない）
	label  string // 何を見たか
	fix    string // 直し方（コマンドを1つ。長い説明は docs を指す）
	needed bool   // いまの進み具合で、もう要るか
	later  string // まだ要らないときの、いつ要るか
}

func need(stage, from int) (bool, string) {
	return stage >= from, fmt.Sprintf("Stage %d から", from)
}

func cmdDoctor(env doctorEnv, wd string, args []string) error {
	// Ubuntu の外（Mac や Windows の側）で叩くと、残りの結果はその手元の話で意味が無い。
	// bash が古い、shellcheck が無い、と並ぶと本当の原因が埋もれるので、ここで止める。
	name, ver := osRelease(env)
	if name != "Ubuntu" {
		where := name
		if where == "" {
			where = map[string]string{"darwin": "macOS", "windows": "Windows"}[env.goos]
		}
		if where == "" {
			where = "OS を読めない"
		}
		fmt.Println()
		fmt.Println(red("  ✗ ") + "Ubuntu の中で実行していない（" + where + "）")
		fmt.Println(dim("    → Ubuntu のシェルを開いてから、もう一度実行する（docs/02 の「1. Ubuntuのシェルを開く」）"))
		fmt.Println()
		return fmt.Errorf("Ubuntu の中で実行してください")
	}

	stage := -1 // リポジトリが見つからないときは、進み具合が分からない
	root, rerr := findRepo(wd)
	var p *Progress
	if rerr == nil {
		if pp, err := LoadProgress(root); err == nil {
			p = pp
			if n, err := strconv.Atoi(strings.TrimSpace(p.Stage)); err == nil {
				stage = n
			}
		}
	}
	shown := stage
	if shown < 0 {
		shown = 0 // 分からなければ、Day 0 に要るものだけを求める
	}
	groups := []struct {
		title string
		items []finding
	}{
		{"Ubuntu と基本の道具", doctorBase(env, ver)},
		{"学習用リポジトリ", doctorRepo(env, root, rerr, p)},
		{"Docker と PostgreSQL", doctorDocker(env, shown, p)},
		{"Go", doctorGo(env, shown)},
		{"Node.js", doctorNode(env, shown)},
	}

	problems := 0
	for _, g := range groups {
		fmt.Println()
		rule(g.title)
		for _, f := range g.items {
			switch {
			case f.ok && f.warn:
				fmt.Println(yellow("  ! ") + f.label)
				if f.fix != "" {
					fmt.Println(dim("    → " + f.fix))
				}
			case f.ok:
				fmt.Println(green("  ✓ ") + f.label)
			case !f.needed:
				fmt.Println(dim(fmt.Sprintf("  ・ %s（%s。まだ要らない）", f.label, f.later)))
			default:
				problems++
				fmt.Println(red("  ✗ ") + f.label)
				if f.fix != "" {
					fmt.Println(dim("    → " + f.fix))
				}
			}
		}
	}
	fmt.Println()
	if problems > 0 {
		return fmt.Errorf("%d 件、直すところがあります（上の → を1つずつ）", problems)
	}
	if stage < 0 {
		fmt.Println(green("  Day 0 に要るものは整っています"))
	} else {
		fmt.Println(green(fmt.Sprintf("  整っています（Stage %d までに要るものを見た）", stage)))
	}
	return nil
}

func doctorBase(env doctorEnv, ver string) []finding {
	var out []finding
	always := func(f finding) finding { f.needed = true; return f }

	switch {
	case ver != "26.04":
		out = append(out, finding{ok: true, warn: true, needed: true,
			label: "Ubuntu " + ver + "（カリキュラムは 26.04 で確かめている。多くはそのまま動く）"})
	default:
		out = append(out, finding{ok: true, needed: true, label: "Ubuntu " + ver})
	}

	if v, err := env.run("bash", "--version"); err == nil {
		m := regexp.MustCompile(`version (\d+)\.(\d+)`).FindStringSubmatch(v)
		if major, _ := strconv.Atoi(safeIndex(m, 1)); major >= 5 {
			out = append(out, finding{ok: true, needed: true, label: "bash " + m[1] + "." + m[2]})
		} else {
			out = append(out, always(finding{label: "bash が 5 より古い", fix: "Ubuntu 26.04 を使う（docs/02）"}))
		}
	}

	if _, err := env.look("git"); err != nil {
		out = append(out, always(finding{label: "git が無い", fix: "sudo apt install -y git（docs/02 の「2. 道具を入れる」）"}))
	} else if n, _ := env.run("git", "config", "--global", "user.name"); n == "" {
		out = append(out, always(finding{label: "git に名前が設定されていない（コミットに要る）",
			fix: `git config --global user.name "自分の名前またはID"（docs/02 の「3. GitHubにつなぐ」）`}))
	} else {
		out = append(out, finding{ok: true, needed: true, label: "git（名前：" + n + "）"})
	}

	if _, err := env.look("gh"); err != nil {
		out = append(out, always(finding{label: "gh が無い", fix: "sudo apt install -y gh（docs/02 の「2. 道具を入れる」）"}))
	} else if _, err := env.run("gh", "auth", "status"); err != nil {
		out = append(out, always(finding{label: "gh が GitHub にログインしていない", fix: "gh auth login（docs/02 の「3. GitHubにつなぐ」）"}))
	} else {
		out = append(out, finding{ok: true, needed: true, label: "gh（ログイン済み）"})
	}

	if _, err := env.look("shellcheck"); err != nil {
		out = append(out, always(finding{label: "shellcheck が無い", fix: "sudo apt install -y shellcheck（docs/02 の「2. 道具を入れる」）"}))
	} else {
		out = append(out, finding{ok: true, needed: true, label: "shellcheck"})
	}
	return out
}

// osRelease は /etc/os-release から、名前と版を読む。
func osRelease(env doctorEnv) (name, version string) {
	raw, err := env.read("/etc/os-release")
	if err != nil {
		return "", ""
	}
	for _, line := range strings.Split(string(raw), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.Trim(v, `"`)
		switch k {
		case "NAME":
			name = v
		case "VERSION_ID":
			version = v
		}
	}
	return name, version
}

func doctorRepo(env doctorEnv, root string, rerr error, p *Progress) []finding {
	if rerr != nil {
		return []finding{{needed: true, label: "学習用リポジトリの中で実行していない",
			fix: "cd ~/offgrid-log してから実行する（まだ作っていなければ docs/02 の「4. 自分の学習用リポジトリを作る」）"}}
	}
	out := []finding{{ok: true, needed: true, label: "学習用リポジトリ：" + root}}
	url, err := env.run("git", "-C", root, "remote", "get-url", "origin")
	switch {
	case err != nil:
		out = append(out, finding{needed: true, label: "origin が無い（push できない）",
			fix: "docs/02 の「4. 自分の学習用リポジトリを作る」の gh repo create で作ったか確かめる"})
	case isTemplate(url):
		// 学習の記録がテンプレートに入ってしまう
		out = append(out, finding{needed: true, label: "ここはテンプレート（kom3da/offgrid）そのもの。学習の記録はここに書かない",
			fix: "自分の学習用リポジトリに移る（docs/02 の「4. 自分の学習用リポジトリを作る」）"})
	default:
		out = append(out, finding{ok: true, needed: true, label: "origin：" + url})
	}
	if p == nil {
		out = append(out, finding{needed: true, label: "PROGRESS.md を読めない", fix: "リポジトリの根に PROGRESS.md があるか確かめる"})
	} else {
		out = append(out, finding{ok: true, needed: true, label: fmt.Sprintf("PROGRESS.md（Stage %s、次のユニット %s）", p.Stage, p.Unit)})
	}
	return out
}

// isTemplate は、origin がテンプレート（kom3da/offgrid）そのものかどうか。
// ホスト名では見ない。SSH の Host に別名（github-foo など）を付けている人がいる。
func isTemplate(url string) bool {
	u := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(url), "/"), ".git")
	return strings.HasSuffix(u, "/kom3da/offgrid") || strings.HasSuffix(u, ":kom3da/offgrid")
}

func doctorDocker(env doctorEnv, stage int, p *Progress) []finding {
	needed, later := need(stage, 1)
	if _, err := env.look("docker"); err != nil {
		return []finding{{needed: needed, later: later + "。F10 で入れる", label: "docker が無い",
			fix: "F10 の手順で入れる（docs/10 の「F10で使う起動例」）"}}
	}
	if out, err := env.run("docker", "info"); err != nil {
		if strings.Contains(out, "permission denied") {
			return []finding{{needed: needed, later: later, label: "docker を使う権限が無い",
				fix: `sudo usermod -aG docker "$USER" のあと、ログインし直す（F10）`}}
		}
		return []finding{{needed: needed, later: later, label: "Docker が動いていない", fix: "sudo systemctl start docker"}}
	}
	res := []finding{{ok: true, needed: true, label: "Docker が動いている"}}

	st, err := env.run("docker", "inspect", "-f", "{{.State.Running}} {{.Config.Image}}", "pg-offgrid")
	running := false
	switch {
	case err != nil:
		res = append(res, finding{needed: needed, later: later, label: "pg-offgrid のコンテナが無い",
			fix: "docs/10 の「F10で使う起動例」の docker run"})
	case strings.HasPrefix(st, "false"):
		res = append(res, finding{needed: needed, later: later, label: "pg-offgrid が止まっている", fix: "docker start pg-offgrid"})
	default:
		running = true
		img := strings.TrimSpace(strings.TrimPrefix(st, "true"))
		if img != "postgres:18" {
			res = append(res, finding{ok: true, warn: true, needed: true,
				label: "pg-offgrid は動いているが、イメージが " + img + "（カリキュラムは postgres:18）"})
		} else {
			res = append(res, finding{ok: true, needed: true, label: "pg-offgrid（postgres:18）が動いている"})
		}
	}

	// pagila は D1 の課題11で入れる。それより前は要らない
	pagilaNeeded := p != nil && p.IsDone("D1")
	if !running {
		return res
	}
	n, err := env.run("docker", "exec", "pg-offgrid", "psql", "-U", "postgres", "-d", "pagila", "-tAc", "select count(*) from film")
	switch {
	case err != nil:
		res = append(res, finding{needed: pagilaNeeded, later: "D1 の課題11で入れる", label: "pagila が入っていない",
			fix: "D1 の課題11（タグ pagila-v3.1.0 の schema と data を読み込む）"})
	case strings.TrimSpace(n) != "1000":
		res = append(res, finding{ok: true, warn: true, needed: true,
			label: "pagila の film が " + strings.TrimSpace(n) + " 行（配布どおりなら 1000）"})
	default:
		res = append(res, finding{ok: true, needed: true, label: "pagila（film 1000行）"})
	}
	return res
}

func doctorGo(env doctorEnv, stage int) []finding {
	needed, later := need(stage, 1)
	var out []finding
	if _, err := env.look("go"); err != nil {
		out = append(out, finding{needed: needed, later: later + "。G1 で入れる", label: "go が無い",
			fix: "G1 の準備（公式サイトの「Download and install」）"})
	} else {
		v, _ := env.run("go", "version")
		m := regexp.MustCompile(`go1\.(\d+)`).FindStringSubmatch(v)
		if minor, err := strconv.Atoi(safeIndex(m, 1)); err != nil || minor < 22 {
			out = append(out, finding{needed: needed, later: later, label: "Go が 1.22 より古い（" + v + "）",
				fix: "G1 の準備（公式サイトの「Download and install」）"})
		} else {
			out = append(out, finding{ok: true, needed: true, label: strings.TrimPrefix(v, "go version ")})
		}
	}
	if _, err := env.look("gcc"); err != nil {
		out = append(out, finding{needed: needed, later: later + "。G1 で入れる", label: "gcc が無い",
			fix: "sudo apt install -y build-essential（go test -race に要る。G1 の準備）"})
	} else {
		out = append(out, finding{ok: true, needed: true, label: "gcc（go test -race が使える）"})
	}
	return out
}

func doctorNode(env doctorEnv, stage int) []finding {
	needed, later := need(stage, 2)
	var out []finding
	for _, tool := range []string{"node", "npm"} {
		if _, err := env.look(tool); err != nil {
			out = append(out, finding{needed: needed, later: later + "。T2 で入れる", label: tool + " が無い",
				fix: "T2 の準備（Node.js の公式ドキュメントの方法で、LTS 版を入れる）"})
			continue
		}
		v, _ := env.run(tool, "--version")
		out = append(out, finding{ok: true, needed: true, label: tool + " " + v})
	}
	return out
}

func safeIndex(s []string, i int) string {
	if i < len(s) {
		return s[i]
	}
	return ""
}
