package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fs8280071-gif/player-tui/internal/browser"
	"github.com/fs8280071-gif/player-tui/internal/sanitize"
	"github.com/fs8280071-gif/player-tui/internal/term"
	"github.com/fs8280071-gif/player-tui/internal/theme"
)

type BrowseOptions struct {
	StartDir string
	Theme    theme.Theme
	NoColor  bool
}

func Browse(ctx context.Context, opts BrowseOptions) (string, error) {
	m, err := browser.New(opts.StartDir)
	if err != nil {
		return "", &Error{Code: 3, Err: fmt.Errorf("cannot browse %q: %w", opts.StartDir, err)}
	}
	t, err := term.Open()
	if err != nil {
		return "", &Error{Code: 4, Err: err}
	}
	defer t.Close()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	keys := make(chan term.Key, 16)
	go t.ReadKeys(runCtx, keys)

	help := false
	dirty := true
	lastW, lastH := 0, 0
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		w, h := t.Size()
		if w != lastW || h != lastH {
			lastW, lastH = w, h
			dirty = true
		}
		if dirty {
			dirty = false
			t.WriteString(string(drawBrowser(m, w, h, opts.NoColor, opts.Theme, help)))
		}
		select {
		case <-ctx.Done():
			return "", nil
		case k := <-keys:
			switch k {
			case term.KeyQuit, term.KeyEsc:
				return "", nil
			case term.KeyUp:
				m.Move(-1)
				dirty = true
			case term.KeyDown:
				m.Move(1)
				dirty = true
			case term.KeyEnter:
				if path, ok := m.Enter(); ok {
					return path, nil
				}
				dirty = true
			case term.KeyBackspace, term.KeyBack, term.KeyLeft:
				m.Parent()
				dirty = true
			case term.KeyHelp:
				help = !help
				dirty = true
			}
		case <-ticker.C:
		}
	}
}

func drawBrowser(m *browser.Model, width, height int, noColor bool, th theme.Theme, help bool) []byte {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	var dst []byte
	dst = append(dst, "\x1b[2J"...)
	dst = append(dst, "\x1b[H"...)
	if th.Name == "" {
		th, _ = theme.Parse("catppuccin")
	}
	dst = append(dst, th.StatusStyle(noColor)...)
	header := " browse  " + m.Dir
	dst = append(dst, truncate(header, width)...)
	dst = append(dst, strings.Repeat(" ", max(0, width-utf8.RuneCountInString(truncate(header, width))))...)
	dst = append(dst, "\x1b[0m"...)

	listRows := height - 2
	if listRows < 1 {
		listRows = 1
	}
	offset := 0
	if m.Cursor >= listRows {
		offset = m.Cursor - listRows + 1
	}
	for i := 0; i < listRows && offset+i < len(m.Entries); i++ {
		entry := m.Entries[offset+i]
		name := sanitize.Text(entry.Name)
		if name == "" {
			name = "?"
		}
		if entry.Dir {
			name += "/"
		}
		line := "  " + name
		if offset+i == m.Cursor {
			line = "> " + name
		}
		line = truncate(line, width)
		dst = append(dst, "\x1b[0m"...)
		if offset+i == m.Cursor {
			dst = append(dst, th.HelpStyle(noColor)...)
		} else if entry.Dir {
			dst = append(dst, th.AccentStyle(noColor)...)
		}
		dst = append(dst, "\x1b["...)
		dst = strconv.AppendInt(dst, int64(2+i), 10)
		dst = append(dst, ";1H"...)
		dst = append(dst, line...)
		dst = append(dst, strings.Repeat(" ", max(0, width-utf8.RuneCountInString(line)))...)
		dst = append(dst, "\x1b[0m"...)
	}
	if len(m.Entries) == 0 {
		dst = append(dst, "\x1b[2;1H"...)
		dst = append(dst, truncate("  (empty directory)", width)...)
	}
	status := " j/k move  enter open  h parent  q quit  ? help "
	if m.Err != nil {
		status = " error: " + sanitize.Text(m.Err.Error()) + " "
	}
	dst = append(dst, "\x1b[0m"...)
	dst = append(dst, th.StatusStyle(noColor)...)
	dst = append(dst, "\x1b["...)
	dst = strconv.AppendInt(dst, int64(height), 10)
	dst = append(dst, ";1H"...)
	status = truncate(status, width)
	dst = append(dst, status...)
	dst = append(dst, strings.Repeat(" ", max(0, width-utf8.RuneCountInString(status)))...)
	dst = append(dst, "\x1b[0m"...)
	if help {
		dst = append(dst, buildBrowserHelp(width, height, noColor, th)...)
	}
	return dst
}

func buildBrowserHelp(width, height int, noColor bool, th theme.Theme) []byte {
	lines := []string{
		" player-tui browser ",
		"",
		" j / down       next entry",
		" k / up         previous entry",
		" enter          open file or directory",
		" h / backspace  parent directory",
		" ?              close this help",
		" q              quit",
		"",
	}
	boxW := 0
	for _, line := range lines {
		if n := utf8.RuneCountInString(line) + 2; n > boxW {
			boxW = n
		}
	}
	if boxW > width {
		boxW = width
	}
	if boxW < 4 {
		return nil
	}
	startCol := (width - boxW) / 2
	if startCol < 0 {
		startCol = 0
	}
	startRow := (height - len(lines)) / 2
	if startRow < 1 {
		startRow = 1
	}
	var dst []byte
	for i, line := range lines {
		if startRow+i > height {
			break
		}
		dst = append(dst, "\x1b["...)
		dst = strconv.AppendInt(dst, int64(startRow+i), 10)
		dst = append(dst, ';')
		dst = strconv.AppendInt(dst, int64(startCol+1), 10)
		dst = append(dst, 'H')
		dst = append(dst, "\x1b[0m"...)
		dst = append(dst, th.HelpStyle(noColor)...)
		line = truncate(line, boxW)
		dst = append(dst, line...)
		dst = append(dst, strings.Repeat(" ", max(0, boxW-utf8.RuneCountInString(line)))...)
		dst = append(dst, "\x1b[0m"...)
	}
	return dst
}

func SelectedTitle(path string) string {
	title := sanitize.Text(filepath.Base(path))
	if title == "" {
		return "player-tui"
	}
	return title
}
