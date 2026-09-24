package main

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
)

// 課題シートとカリキュラムで実際に使っている記法だけを、HTMLに直す。
// 外部のライブラリを使わないのは、閉域でもビルドできるようにするため。

var (
	mdHeading  = regexp.MustCompile(`^(#{1,4}) +(.+)$`)
	mdFence    = regexp.MustCompile("^```([a-zA-Z0-9]*)\\s*$")
	mdBullet   = regexp.MustCompile(`^(\s*)[-*] +(.*)$`)
	mdOrdered  = regexp.MustCompile(`^(\s*)(\d+)\. +(.*)$`)
	mdCheckbox = regexp.MustCompile(`^\[([ xX])\] +(.*)$`)
	mdTableRow = regexp.MustCompile(`^\|(.+)\|\s*$`)
	mdTableSep = regexp.MustCompile(`^\|[\s:|-]+\|\s*$`)
	mdCode     = regexp.MustCompile("`([^`]+)`")
	mdStrong   = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdLink     = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	// html.EscapeString を通したあとに探すので、山かっこは &lt; &gt; になっている
	mdBareURL = regexp.MustCompile(`&lt;(https?://[^&\s]+)&gt;`)
)

// safeHref は、開いてよい行き先だけを通す。
// 学習者自身の notes.md も表示するので、javascript: のような仕組みは弾く。
func safeHref(h string) string {
	l := strings.ToLower(strings.TrimSpace(h))
	switch {
	case strings.HasPrefix(l, "http://"), strings.HasPrefix(l, "https://"),
		strings.HasPrefix(l, "/"), strings.HasPrefix(l, "#"):
		return h
	}
	// 「:」より前が仕組みの名前。知らないものは開かせない
	if i := strings.IndexAny(l, ":/"); i >= 0 && l[i] == ':' {
		return ""
	}
	return h
}

// inlineHTML は、行の中の記法（コード、強調、リンク）をHTMLにする。
func inlineHTML(s string, link func(string) string) string {
	out := html.EscapeString(s)
	// リンクを先に処理する（中のテキストは、あとで装飾される）
	out = mdBareURL.ReplaceAllStringFunc(out, func(m string) string {
		url := mdBareURL.FindStringSubmatch(m)[1]
		return fmt.Sprintf(`<a href="%s" rel="noreferrer" target="_blank">%s</a>`, url, url)
	})
	out = mdLink.ReplaceAllStringFunc(out, func(m string) string {
		g := mdLink.FindStringSubmatch(m)
		// out は先にエスケープ済みなので、いったん元の文字に戻してから行き先を組み立てる。
		// 書き換えはファイルのディレクトリ名を継ぎ足すので、最後にもう一度エスケープして属性に入れる
		// （ディレクトリ名に `">` を含めると、属性を抜けて要素を差し込めた）
		href := html.UnescapeString(g[2])
		if link != nil {
			href = link(href)
		}
		href = safeHref(href)
		if href == "" {
			return g[1] // 行き先が怪しいとき、開けないときは、文字だけ出す
		}
		rel := ""
		if l := strings.ToLower(href); strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "https://") {
			rel = ` rel="noreferrer" target="_blank"`
		}
		return fmt.Sprintf(`<a href="%s"%s>%s</a>`, html.EscapeString(href), rel, g[1])
	})
	out = mdStrong.ReplaceAllString(out, "<strong>$1</strong>")
	out = mdCode.ReplaceAllString(out, "<code>$1</code>")
	return out
}

type listState struct {
	tag    string // ul か ol
	indent int
}

// renderMarkdown は Markdown を HTML にする。
// link は、相対リンクの行き先を書き換える関数（nil なら、そのまま）。
// checkbox が nil でなければ、`- [ ]` の行をチェックボックスにする（work.md 用）。
func renderMarkdown(src string, link func(string) string, checkbox func(n int, done bool, text string) string) string {
	var b strings.Builder
	var stack []listState
	inCode, inTable, para := false, false, false

	closeLists := func(to int) {
		for len(stack) > 0 && stack[len(stack)-1].indent >= to {
			fmt.Fprintf(&b, "</%s>\n", stack[len(stack)-1].tag)
			stack = stack[:len(stack)-1]
		}
	}
	closePara := func() {
		if para {
			b.WriteString("</p>\n")
			para = false
		}
	}
	closeTable := func() {
		if inTable {
			b.WriteString("</tbody></table>\n")
			inTable = false
		}
	}

	for _, line := range strings.Split(src, "\n") {
		if m := mdFence.FindStringSubmatch(line); m != nil {
			if inCode {
				b.WriteString("</code></pre>\n")
				inCode = false
			} else {
				closePara()
				closeLists(0)
				closeTable()
				lang := m[1]
				if lang == "" {
					lang = "text"
				}
				fmt.Fprintf(&b, `<pre data-lang="%s"><code>`, html.EscapeString(lang))
				inCode = true
			}
			continue
		}
		if inCode {
			b.WriteString(html.EscapeString(line) + "\n")
			continue
		}
		if strings.TrimSpace(line) == "" {
			closePara()
			closeLists(0)
			closeTable()
			continue
		}
		if m := mdHeading.FindStringSubmatch(line); m != nil {
			closePara()
			closeLists(0)
			closeTable()
			level := len(m[1])
			if level < 2 {
				level = 2 // ページの見出しは別に出すので、本文は h2 から
			}
			id := slugify(m[2])
			fmt.Fprintf(&b, `<h%d id="%s">%s</h%d>`+"\n", level, id, inlineHTML(m[2], link), level)
			continue
		}
		if strings.HasPrefix(line, "> ") {
			closePara()
			closeLists(0)
			fmt.Fprintf(&b, "<blockquote>%s</blockquote>\n", inlineHTML(strings.TrimPrefix(line, "> "), link))
			continue
		}
		if mdTableRow.MatchString(line) {
			if mdTableSep.MatchString(line) {
				continue
			}
			cells := splitRow(line)
			if !inTable {
				closePara()
				closeLists(0)
				b.WriteString("<table><thead><tr>")
				for _, c := range cells {
					fmt.Fprintf(&b, "<th>%s</th>", inlineHTML(c, link))
				}
				b.WriteString("</tr></thead><tbody>\n")
				inTable = true
				continue
			}
			b.WriteString("<tr>")
			for _, c := range cells {
				fmt.Fprintf(&b, "<td>%s</td>", inlineHTML(c, link))
			}
			b.WriteString("</tr>\n")
			continue
		}
		closeTable()
		if m := mdBullet.FindStringSubmatch(line); m != nil {
			indent := len(m[1])
			openList(&b, &stack, "ul", indent)
			closePara()
			text := m[2]
			if cb := mdCheckbox.FindStringSubmatch(text); cb != nil && checkbox != nil {
				num := 0
				if n := regexp.MustCompile(`^(\d+)\.`).FindStringSubmatch(cb[2]); n != nil {
					num, _ = strconv.Atoi(n[1])
				}
				fmt.Fprintf(&b, "<li class=\"task\">%s</li>\n", checkbox(num, strings.ToLower(cb[1]) == "x", cb[2]))
				continue
			}
			if cb := mdCheckbox.FindStringSubmatch(text); cb != nil {
				mark := "☐"
				if strings.ToLower(cb[1]) == "x" {
					mark = "☑"
				}
				fmt.Fprintf(&b, "<li>%s %s</li>\n", mark, inlineHTML(cb[2], link))
				continue
			}
			fmt.Fprintf(&b, "<li>%s</li>\n", inlineHTML(text, link))
			continue
		}
		if m := mdOrdered.FindStringSubmatch(line); m != nil {
			indent := len(m[1])
			openList(&b, &stack, "ol", indent)
			closePara()
			fmt.Fprintf(&b, "<li value=\"%s\">%s</li>\n", m[2], inlineHTML(m[3], link))
			continue
		}
		// ふつうの段落。リストの続きの行なら、前の項目にくっつける
		if len(stack) > 0 && strings.HasPrefix(line, "   ") {
			fmt.Fprintf(&b, "<p class=\"cont\">%s</p>\n", inlineHTML(strings.TrimSpace(line), link))
			continue
		}
		closeLists(0)
		if !para {
			b.WriteString("<p>")
			para = true
		} else {
			b.WriteString(" ")
		}
		b.WriteString(inlineHTML(line, link))
	}
	if inCode {
		b.WriteString("</code></pre>\n")
	}
	closePara()
	closeLists(0)
	closeTable()
	return b.String()
}

func openList(b *strings.Builder, stack *[]listState, tag string, indent int) {
	for len(*stack) > 0 {
		top := (*stack)[len(*stack)-1]
		if top.indent > indent || (top.indent == indent && top.tag != tag) {
			fmt.Fprintf(b, "</%s>\n", top.tag)
			*stack = (*stack)[:len(*stack)-1]
			continue
		}
		break
	}
	if len(*stack) == 0 || (*stack)[len(*stack)-1].indent < indent {
		fmt.Fprintf(b, "<%s>\n", tag)
		*stack = append(*stack, listState{tag: tag, indent: indent})
	}
}

func splitRow(line string) []string {
	inner := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(line), "|"), "|")
	parts := strings.Split(inner, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

var slugDrop = regexp.MustCompile(`[^\p{L}\p{N}ー・]+`)

func slugify(s string) string {
	s = strings.TrimSpace(s)
	s = mdCode.ReplaceAllString(s, "$1")
	s = slugDrop.ReplaceAllString(s, "-")
	return strings.ToLower(strings.Trim(s, "-"))
}
