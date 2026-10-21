package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/term"
)

//go:embed snippets
var bundled embed.FS

type item struct {
	title string
	read  func() string
}

func titleOf(name string) string {
	t := strings.TrimSuffix(name, filepath.Ext(name))
	return strings.NewReplacer("-", " ", "_", " ").Replace(t)
}

func load() []item {
	var items []item
	if es, err := bundled.ReadDir("snippets"); err == nil {
		for _, e := range es {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			items = append(items, item{titleOf(name), func() string {
				b, _ := bundled.ReadFile("snippets/" + name)
				return string(b)
			}})
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		dir := filepath.Join(home, ".codes")
		if es, err := os.ReadDir(dir); err == nil {
			for _, e := range es {
				if e.IsDir() {
					continue
				}
				p := filepath.Join(dir, e.Name())
				items = append(items, item{titleOf(e.Name()), func() string {
					b, _ := os.ReadFile(p)
					return string(b)
				}})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].title) < strings.ToLower(items[j].title) })
	return items
}

func wrap(text string, width int) []string {
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		r := []rune(strings.ReplaceAll(line, "\t", "    "))
		if len(r) == 0 {
			out = append(out, "")
		}
		for len(r) > 0 {
			n := width
			if len(r) < n {
				n = len(r)
			}
			out = append(out, string(r[:n]))
			r = r[n:]
		}
	}
	return out
}

func main() {
	items := load()
	if len(items) == 0 {
		fmt.Println("No codes yet.")
		return
	}
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		fmt.Println("Run this in a terminal.")
		os.Exit(1)
	}
	restoreOut := enableVT()
	old, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	out := os.Stdout
	out.WriteString("\x1b[?1049h\x1b[?25l")
	quit := func() {
		out.WriteString("\x1b[?25h\x1b[?1049l")
		term.Restore(fd, old)
		restoreOut()
		os.Exit(0)
	}

	filter, msg := "", ""
	sel, scroll := 0, 0
	var view *item

	shown := func() []item {
		var l []item
		for _, it := range items {
			if strings.Contains(strings.ToLower(it.title), strings.ToLower(filter)) {
				l = append(l, it)
			}
		}
		return l
	}

	render := func() {
		w, h, e := term.GetSize(int(os.Stdout.Fd()))
		if e != nil || w < 10 {
			w, h = 80, 24
		}
		var b strings.Builder
		b.WriteString("\x1b[2J\x1b[H")
		if view != nil {
			lines := wrap(view.read(), w)
			room := h - 4
			if scroll > len(lines)-room {
				scroll = len(lines) - room
			}
			if scroll < 0 {
				scroll = 0
			}
			end := scroll + room
			if end > len(lines) {
				end = len(lines)
			}
			b.WriteString("\x1b[1m" + view.title + "\x1b[0m\r\n")
			b.WriteString("\x1b[2m" + strings.Repeat("-", min(w, 60)) + "\x1b[0m\r\n")
			b.WriteString(strings.Join(lines[scroll:end], "\r\n") + "\r\n")
			b.WriteString("\x1b[2mSpace copy | Up/Down scroll | Esc back\x1b[0m  \x1b[32m" + msg + "\x1b[0m")
		} else {
			list := shown()
			if sel >= len(list) {
				sel = max(0, len(list)-1)
			}
			room := h - 4
			start := max(0, min(sel-room/2, len(list)-room))
			b.WriteString("\x1b[1mcodes\x1b[0m \x1b[2mUp/Down move | Enter view | Space copy | type to filter | Esc quit\x1b[0m\r\n")
			b.WriteString("\x1b[36m> " + filter + "\x1b[0m\r\n")
			for i := start; i < len(list) && i < start+room; i++ {
				if i == sel {
					b.WriteString("\x1b[7m " + list[i].title + " \x1b[0m\r\n")
				} else {
					b.WriteString("  " + list[i].title + "\r\n")
				}
			}
			if len(list) == 0 {
				b.WriteString("  (no match)\r\n")
			}
			b.WriteString("\x1b[32m" + msg + "\x1b[0m")
		}
		out.WriteString(b.String())
	}

	doCopy := func(it item) {
		if err := copyToClipboard(it.read()); err != nil {
			msg = "Copy failed: " + err.Error()
		} else {
			msg = "Copied \"" + it.title + "\""
		}
	}

	render()
	buf := make([]byte, 32)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			quit()
		}
		k := string(buf[:n])
		list := shown()
		switch {
		case k == "\x03":
			quit()
		case view != nil:
			switch k {
			case " ":
				doCopy(*view)
			case "\x1b", "\r", "q":
				view, msg, scroll = nil, "", 0
			case "\x1b[A":
				scroll--
			case "\x1b[B":
				scroll++
			case "\x1b[5~":
				scroll -= 10
			case "\x1b[6~":
				scroll += 10
			}
		case k == "\x1b":
			quit()
		case k == "\x1b[A":
			sel = max(0, sel-1)
		case k == "\x1b[B":
			sel = min(len(list)-1, sel+1)
		case k == "\r":
			if sel < len(list) {
				it := list[sel]
				view, msg, scroll = &it, "", 0
			}
		case k == " ":
			if sel < len(list) {
				doCopy(list[sel])
			}
		case k == "\x7f" || k == "\b":
			if len(filter) > 0 {
				filter = filter[:len(filter)-1]
			}
			sel = 0
		case len(k) == 1 && k[0] > ' ' && k[0] < 0x7f:
			filter += k
			sel, msg = 0, ""
		}
		render()
	}
}
