package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeDoctor は、doctor が見る外の世界を作る。既定では Stage 5 まで全部そろった Ubuntu。
// テストごとに、1つずつ欠けさせて確かめる。
type fakeDoctor struct {
	osRelease string
	missing   map[string]bool   // PATH に無い道具
	fail      map[string]string // 失敗させるコマンド（キーは「名前 引数…」の先頭一致）→ 出力
	out       map[string]string // 成功したときの出力
}

func newFakeDoctor() *fakeDoctor {
	return &fakeDoctor{
		// Ubuntu 26.04 の /etc/os-release と同じ形（判定は ID で行う）
		osRelease: "PRETTY_NAME=\"Ubuntu 26.04 LTS\"\nNAME=\"Ubuntu\"\nVERSION_ID=\"26.04\"\nID=ubuntu\nID_LIKE=debian\n",
		missing:   map[string]bool{},
		fail:      map[string]string{},
		out: map[string]string{
			"bash --version":                "GNU bash, version 5.3.9(1)-release (aarch64-unknown-linux-gnu)",
			"git config --global user.name": "learner",
			"git -C":                        "git@github.com:learner/offgrid-log.git",
			"docker inspect":                "true postgres:18",
			"docker exec pg-offgrid psql":   "1000",
			"go version":                    "go version go1.26.0 linux/arm64",
			"node --version":                "v22.22.1",
			"npm --version":                 "10.9.7",
		},
	}
}

func (f *fakeDoctor) env() doctorEnv {
	return doctorEnv{
		run: func(name string, args ...string) (string, error) {
			cmd := strings.Join(append([]string{name}, args...), " ")
			for k, v := range f.fail {
				if strings.HasPrefix(cmd, k) {
					return v, errors.New("exit status 1")
				}
			}
			for k, v := range f.out {
				if strings.HasPrefix(cmd, k) {
					return v, nil
				}
			}
			return "", nil
		},
		look: func(name string) (string, error) {
			if f.missing[name] {
				return "", errors.New("not found")
			}
			return "/usr/bin/" + name, nil
		},
		read: func(path string) ([]byte, error) {
			if path == "/etc/os-release" && f.osRelease != "" {
				return []byte(f.osRelease), nil
			}
			return nil, os.ErrNotExist
		},
		goos: "linux",
	}
}

// repoAtStage は、指定したステージの学習用リポジトリを作る。d1Done で D1 を完了にする。
func repoAtStage(t *testing.T, stage string, d1Done bool) string {
	t.Helper()
	root := newRepo(t)
	path := filepath.Join(root, "PROGRESS.md")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := strings.Replace(string(raw), "- ステージ：Stage 1", "- ステージ：Stage "+stage, 1)
	if d1Done {
		s = strings.Replace(s, "- [ ] D1 ", "- [x] D1 ", 1)
	}
	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func runDoctor(t *testing.T, f *fakeDoctor, wd string) (string, error) {
	t.Helper()
	var err error
	out := capture(t, func() { err = cmdDoctor(f.env(), wd, nil) })
	return out, err
}

func TestDoctorAllGood(t *testing.T) {
	root := repoAtStage(t, "5", true)
	out, err := runDoctor(t, newFakeDoctor(), root)
	if err != nil {
		t.Fatalf("そろっているのに失敗した: %v\n%s", err, out)
	}
	for _, want := range []string{"Ubuntu 26.04", "bash 5.3", "gh（ログイン済み）", "pg-offgrid（postgres:18）", "pagila（film 1000行）", "go1.26.0", "node v22.22.1", "整っています"} {
		if !strings.Contains(out, want) {
			t.Errorf("出力に %q が無い:\n%s", want, out)
		}
	}
}

// Mac や Windows の側で叩いたら、そのことだけを言って止まる（ほかの ✗ で原因を埋めない）
func TestDoctorOutsideUbuntu(t *testing.T) {
	f := newFakeDoctor()
	f.osRelease = ""
	f.missing["shellcheck"] = true
	env := f.env()
	env.goos = "darwin"
	var err error
	out := capture(t, func() { err = cmdDoctor(env, t.TempDir(), nil) })
	if err == nil {
		t.Fatal("Ubuntu の外なのに通った")
	}
	if !strings.Contains(out, "macOS") {
		t.Errorf("どこにいるかを言っていない:\n%s", out)
	}
	if strings.Contains(out, "shellcheck") {
		t.Errorf("Ubuntu の外の道具まで並べて、原因を埋めている:\n%s", out)
	}
}

func TestDoctorReportsMissingBasics(t *testing.T) {
	root := repoAtStage(t, "0", false)
	cases := []struct {
		name   string
		break_ func(f *fakeDoctor)
		want   string
	}{
		{"shellcheck が無い", func(f *fakeDoctor) { f.missing["shellcheck"] = true }, "sudo apt install -y shellcheck"},
		{"gh が未ログイン", func(f *fakeDoctor) { f.fail["gh auth status"] = "You are not logged into any GitHub hosts." }, "gh auth login"},
		{"git の名前が無い", func(f *fakeDoctor) { f.out["git config --global user.name"] = "" }, "git config --global user.name"},
		{"origin が無い", func(f *fakeDoctor) { f.fail["git -C"] = "error: No such remote 'origin'" }, "origin が無い"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeDoctor()
			c.break_(f)
			out, err := runDoctor(t, f, root)
			if err == nil {
				t.Errorf("欠けているのに通った:\n%s", out)
			}
			if !strings.Contains(out, c.want) {
				t.Errorf("直し方 %q が出ていない:\n%s", c.want, out)
			}
		})
	}
}

// テンプレートそのものの中で叩いたら止める。SSH のホスト別名でも見分ける
func TestDoctorRefusesTheTemplate(t *testing.T) {
	root := repoAtStage(t, "0", false)
	for _, url := range []string{
		"https://github.com/kom3da/offgrid.git",
		"git@github.com:kom3da/offgrid.git",
		"git@github-kom3da:kom3da/offgrid.git", // ~/.ssh/config の Host 別名
	} {
		f := newFakeDoctor()
		f.out["git -C"] = url
		out, err := runDoctor(t, f, root)
		if err == nil || !strings.Contains(out, "テンプレート") {
			t.Errorf("%s をテンプレートと見分けていない:\n%s", url, out)
		}
	}
	// 学習者が自分のリポジトリを offgrid と名付けていても、テンプレート扱いしない
	f := newFakeDoctor()
	f.out["git -C"] = "git@github.com:someone/offgrid.git"
	if out, err := runDoctor(t, f, root); err != nil {
		t.Errorf("他人の offgrid をテンプレートと誤認した: %v\n%s", err, out)
	}
}

// まだ要らないものは、無くても失敗にしない（Stage 0 の人に Go や Docker を求めない）
func TestDoctorDoesNotAskTooEarly(t *testing.T) {
	root := repoAtStage(t, "0", false)
	f := newFakeDoctor()
	f.missing["docker"] = true
	f.missing["go"] = true
	f.missing["gcc"] = true
	f.missing["node"] = true
	f.missing["npm"] = true
	out, err := runDoctor(t, f, root)
	if err != nil {
		t.Fatalf("Stage 0 なのに Go や Docker を求めた: %v\n%s", err, out)
	}
	if !strings.Contains(out, "まだ要らない") {
		t.Errorf("いつ要るかを言っていない:\n%s", out)
	}

	// Stage 1 になったら、Docker と Go は要る
	root1 := repoAtStage(t, "1", false)
	out, err = runDoctor(t, f, root1)
	if err == nil {
		t.Errorf("Stage 1 で Docker と Go が無いのに通った:\n%s", out)
	}
	if !strings.Contains(out, "node が無い（Stage 2 から") {
		t.Errorf("Stage 1 なのに Node.js まで求めている（まだ要らない、と出るはず）:\n%s", out)
	}
}

func TestDoctorDockerStates(t *testing.T) {
	root := repoAtStage(t, "1", false)
	cases := []struct {
		name   string
		break_ func(f *fakeDoctor)
		want   string
	}{
		{"権限が無い", func(f *fakeDoctor) {
			f.fail["docker info"] = "permission denied while trying to connect to the Docker daemon socket"
		}, "usermod -aG docker"},
		{"動いていない", func(f *fakeDoctor) { f.fail["docker info"] = "Cannot connect to the Docker daemon" }, "systemctl start docker"},
		{"コンテナが無い", func(f *fakeDoctor) { f.fail["docker inspect"] = "No such object: pg-offgrid" }, "F10で使う起動例"},
		{"止まっている", func(f *fakeDoctor) { f.out["docker inspect"] = "false postgres:18" }, "docker start pg-offgrid"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeDoctor()
			c.break_(f)
			out, err := runDoctor(t, f, root)
			if err == nil || !strings.Contains(out, c.want) {
				t.Errorf("err=%v、%q が出ていない:\n%s", err, c.want, out)
			}
		})
	}
}

// pagila は D1 の課題11で入れる。D1 を終えるまでは、無くても失敗にしない
func TestDoctorPagilaAfterD1(t *testing.T) {
	f := newFakeDoctor()
	f.fail["docker exec pg-offgrid psql"] = `FATAL:  database "pagila" does not exist`

	out, err := runDoctor(t, f, repoAtStage(t, "1", false))
	if err != nil {
		t.Errorf("D1 の前なのに pagila を求めた: %v\n%s", err, out)
	}
	out, err = runDoctor(t, f, repoAtStage(t, "1", true))
	if err == nil || !strings.Contains(out, "pagila-v3.1.0") {
		t.Errorf("D1 を終えたのに pagila が無いことを言わない: %v\n%s", err, out)
	}
}

func TestDoctorOldGo(t *testing.T) {
	f := newFakeDoctor()
	f.out["go version"] = "go version go1.21.13 linux/amd64"
	out, err := runDoctor(t, f, repoAtStage(t, "1", false))
	if err == nil || !strings.Contains(out, "1.22 より古い") {
		t.Errorf("古い Go を通した: %v\n%s", err, out)
	}
}

// 学習用リポジトリがまだ無い Day 0 の途中でも、動いて、次にやることを言う
func TestDoctorOutsideRepository(t *testing.T) {
	out, err := runDoctor(t, newFakeDoctor(), t.TempDir())
	if err == nil {
		t.Fatal("リポジトリの外なのに、問題なしと言った")
	}
	if !strings.Contains(out, "4. 自分の学習用リポジトリを作る") {
		t.Errorf("次にやることを言っていない:\n%s", out)
	}
	if strings.Contains(out, "go が無い") && !strings.Contains(out, "まだ要らない") {
		t.Errorf("進み具合が分からないのに、Go まで求めている:\n%s", out)
	}
}

// 以下は #27 で見つかったもの。

// os-release は、機械向けの ID で判定する。単引用符でもよく、/etc に無ければ /usr/lib を見る。
// 直す前は、表示名（NAME）の完全一致と二重引用符しか扱わず、Ubuntu を「Ubuntu の外」と言った。
func TestDoctorReadsOSReleaseByTheRules(t *testing.T) {
	root := repoAtStage(t, "0", false)
	cases := []struct {
		name, content, path string
		wantErr             bool
		want                string
	}{
		{"単引用符", "NAME='Ubuntu'\nID='ubuntu'\nVERSION_ID='26.04'\n", "/etc/os-release", false, "Ubuntu 26.04"},
		{"表示名が違っても ID が ubuntu", "NAME=\"Ubuntu Server\"\nID=ubuntu\nVERSION_ID=\"26.04\"\n", "/etc/os-release", false, "Ubuntu 26.04"},
		{"/usr/lib にだけある", "NAME=\"Ubuntu\"\nID=ubuntu\nVERSION_ID=\"26.04\"\n", "/usr/lib/os-release", false, "Ubuntu 26.04"},
		{"Ubuntu の派生は警告で通す", "NAME=\"Linux Mint\"\nID=linuxmint\nID_LIKE=\"ubuntu debian\"\nVERSION_ID=\"22\"\n", "/etc/os-release", false, "Ubuntu の派生"},
		{"Ubuntu と関係ない", "NAME=\"Fedora Linux\"\nID=fedora\nVERSION_ID=41\n", "/etc/os-release", true, "Ubuntu の中で実行していない"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeDoctor()
			env := f.env()
			env.read = func(p string) ([]byte, error) {
				if p == c.path {
					return []byte(c.content), nil
				}
				return nil, os.ErrNotExist
			}
			var err error
			out := capture(t, func() { err = cmdDoctor(env, root, nil) })
			if (err != nil) != c.wantErr {
				t.Errorf("err=%v（期待：失敗=%v）\n%s", err, c.wantErr, out)
			}
			if !strings.Contains(out, c.want) {
				t.Errorf("%q が出ていない:\n%s", c.want, out)
			}
		})
	}
}

// bash・Go・Node が「入っているが動かない」ときに、黙って通さない。
// 直す前は、bash は項目ごと消え、Node は失敗の文面を版として ✓ に出していた。
func TestDoctorReportsToolsThatDoNotRun(t *testing.T) {
	cases := []struct {
		name, stage, cmd, want string
	}{
		{"bash", "0", "bash --version", "bash を実行できない"},
		{"go", "1", "go version", "go を実行できない"},
		{"node", "2", "node --version", "node を実行できない"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeDoctor()
			f.fail[c.cmd] = "error while loading shared libraries: libfoo.so"
			out, err := runDoctor(t, f, repoAtStage(t, c.stage, false))
			if err == nil || !strings.Contains(out, c.want) {
				t.Errorf("動かない %s を通した: %v\n%s", c.name, err, out)
			}
			if strings.Contains(out, "✓ "+c.name+" error") {
				t.Errorf("失敗の文面を版として ✓ に出した:\n%s", out)
			}
		})
	}
}

// Docker が動いていないときの直し方は、接続先で変える。Docker Desktop に systemctl を案内しても直らない
func TestDoctorDockerStartHint(t *testing.T) {
	cases := []struct {
		name, out   string
		noSystemctl bool
		want        string
	}{
		{"Ubuntu の中の docker", "Cannot connect to the Docker daemon at unix:///var/run/docker.sock.", false, "sudo systemctl start docker"},
		{"Docker Desktop", "Cannot connect to the Docker daemon at unix:///home/l/.docker/desktop/docker.sock.", false, "Docker Desktop を起動する"},
		{"systemd の無い WSL", "Cannot connect to the Docker daemon at unix:///var/run/docker.sock.", true, "sudo service docker start"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newFakeDoctor()
			f.fail["docker info"] = c.out
			if c.noSystemctl {
				f.missing["systemctl"] = true
			}
			out, _ := runDoctor(t, f, repoAtStage(t, "1", false))
			if !strings.Contains(out, c.want) {
				t.Errorf("%q を案内していない:\n%s", c.want, out)
			}
		})
	}
}
