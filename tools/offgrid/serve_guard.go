package main

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// この画面は、学習者のファイルを書き換え、git を動かし、テストを走らせる。
// ブラウザは、別のサイトを開いているだけでも 127.0.0.1 へリクエストを送れてしまう
// （画像の読み込みや、隠したフォームなど）。そのため、次の3つで守る。
//
//  1. 手元（ループバック）からの接続だけを通す
//  2. Host が localhost か 127.0.0.1 であることを確かめる（DNS rebinding 対策）
//  3. 書き込み・実行は POST だけにし、画面が出した合い言葉（トークン）と発信元を確かめる

// postOnly は、書き込みか実行をするパス。GET では通さない。
var postOnly = map[string]bool{
	"/session/step": true,
	"/session/task": true,
	"/note":         true,
	"/end/check":    true,
	"/end/done":     true,
	"/end/commit":   true,
	"/tick":         true,
	"/stuck":        true,
}

func newToken() string {
	b := make([]byte, 16)
	// crypto/rand は、読めなかったときに panic する。戻り値の err は常に nil。
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// localAddr は、接続してきた相手が同じマシンかどうか。
func localAddr(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// localHost は、ブラウザがアドレス欄に持っている名前が、手元を指しているかどうか。
func localHost(h string) bool {
	host := h
	if hh, _, err := net.SplitHostPort(h); err == nil {
		host = hh
	}
	host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// sameOrigin は、この画面自身から送られた操作かどうか。
func sameOrigin(r *http.Request) bool {
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "none", "same-origin":
		// none は、アドレス欄から直に開いたとき
	default:
		return false
	}
	if o := r.Header.Get("Origin"); o != "" {
		u, err := url.Parse(o)
		if err != nil || u.Host != r.Host {
			return false
		}
	}
	return true
}

// isWrite は、そのリクエストがファイルを書き換える（か、コマンドを走らせる）ものかどうか。
func isWrite(r *http.Request) bool {
	if postOnly[r.URL.Path] {
		return true
	}
	// /retro は、GET で読み、POST で書く
	return r.URL.Path == "/retro" && r.Method != http.MethodGet && r.Method != http.MethodHead
}

// guard は、上の3つを確かめてから、中の処理に渡す。
func (s *server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !localAddr(r.RemoteAddr) {
			http.Error(w, "このページは、同じマシンからだけ開けます", http.StatusForbidden)
			return
		}
		if !localHost(r.Host) {
			http.Error(w, "http://localhost:<番号>/ で開いてください", http.StatusForbidden)
			return
		}
		if isWrite(r) {
			if r.Method != http.MethodPost {
				http.Error(w, "この操作は、画面のボタンから行ってください", http.StatusMethodNotAllowed)
				return
			}
			if !sameOrigin(r) {
				http.Error(w, "ほかのページからは操作できません", http.StatusForbidden)
				return
			}
			if err := r.ParseForm(); err != nil {
				http.Error(w, "送られた内容を読み取れません", http.StatusBadRequest)
				return
			}
			if r.PostFormValue("csrf") != s.token {
				http.Error(w, "画面を読み込み直してから、もう一度押してください", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// markdownPath は、/file/ で開いてよいファイルだけを通す。
// answers/ はデバッグドリルの答えなので、この道具からは開かせない（学習を奪わないため）。
func (s *server) markdownPath(name string) (string, bool) {
	clean := filepath.Clean("/" + strings.TrimPrefix(name, "/"))
	path := filepath.Join(s.root, clean)
	if !strings.HasSuffix(path, ".md") {
		return "", false
	}
	r, err := filepath.Rel(s.root, path)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", false
	}
	parts := strings.Split(filepath.ToSlash(r), "/")
	for _, p := range parts {
		if p == "answers" {
			return "", false
		}
	}
	if len(parts) == 1 {
		return path, true // README.md や PROGRESS.md
	}
	switch parts[0] {
	case "docs", "drills", "logs", "prompts", "templates", "capstone":
		return path, true
	}
	return "", false
}

// logPath は、/retro?file= で読み書きしてよいファイルだけを通す。
func (s *server) logPath(name string) (string, bool) {
	clean := filepath.Clean("/" + strings.TrimPrefix(name, "/"))
	path := filepath.Join(s.root, clean)
	base := filepath.Join(s.root, "logs") + string(filepath.Separator)
	if !strings.HasPrefix(path, base) || !strings.HasSuffix(path, ".md") {
		return "", false
	}
	return path, true
}

// backPath は、書き終わったあとに戻る先。外のサイトへは飛ばさない。
func backPath(v, fallback string) string {
	if strings.HasPrefix(v, "/") && !strings.HasPrefix(v, "//") {
		return v
	}
	return fallback
}
