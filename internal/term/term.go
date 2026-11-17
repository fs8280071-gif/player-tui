package term

import (
	"bufio"
	"context"
	"errors"
	"os"
	"sync"

	xterm "golang.org/x/term"
)

type Key int

const (
	KeyUnknown Key = iota
	KeySpace
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyMute
	KeyNext
	KeyPrev
	KeyQuit
	KeyEsc
	KeyVolumeUp
	KeyVolumeDown
	KeySpeedUp
	KeySpeedDown
	KeySubtitle
	KeyAudio
	KeyHelp
	KeyEnter
	KeyBackspace
	KeyBack
)

type Terminal struct {
	in   *os.File
	out  *os.File
	old  *xterm.State
	once sync.Once
}

func Open() (*Terminal, error) {
	in, out := os.Stdin, os.Stdout
	if !xterm.IsTerminal(int(in.Fd())) || !xterm.IsTerminal(int(out.Fd())) {
		return nil, errors.New("player-tui needs an interactive terminal (stdin/stdout are not a TTY)")
	}
	old, err := xterm.MakeRaw(int(in.Fd()))
	if err != nil {
		return nil, err
	}
	t := &Terminal{in: in, out: out, old: old}
	t.WriteString("\x1b[?1049h\x1b[?25l\x1b[2J\x1b[H")
	return t, nil
}

func (t *Terminal) Close() {
	t.once.Do(func() {
		t.WriteString("\x1b[0m\x1b[?25h\x1b[?1049l")
		if t.old != nil {
			_ = xterm.Restore(int(t.in.Fd()), t.old)
		}
	})
}

func (t *Terminal) Size() (int, int) {
	w, h, err := xterm.GetSize(int(t.out.Fd()))
	if err != nil || w <= 0 || h <= 0 {
		return 80, 24
	}
	return w, h
}

func (t *Terminal) WriteString(s string) {
	_, _ = t.out.WriteString(s)
}

func (t *Terminal) ReadKeys(ctx context.Context, out chan<- Key) {
	r := bufio.NewReader(t.in)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return
		}
		var k Key
		switch b {
		case 'q', 'Q', 0x03:
			k = KeyQuit
		case ' ':
			k = KeySpace
		case 'm', 'M':
			k = KeyMute
		case 'n', 'N':
			k = KeyNext
		case 'p', 'P':
			k = KeyPrev
		case '+', '=':
			k = KeyVolumeUp
		case '-', '_':
			k = KeyVolumeDown
		case ']':
			k = KeySpeedUp
		case '[':
			k = KeySpeedDown
		case 's', 'S':
			k = KeySubtitle
		case 'a', 'A':
			k = KeyAudio
		case '?':
			k = KeyHelp
		case '\r', '\n':
			k = KeyEnter
		case 0x7f, 0x08:
			k = KeyBackspace
		case 'h', 'H':
			k = KeyBack
		case 'j', 'J':
			k = KeyDown
		case 'k', 'K':
			k = KeyUp
		case 0x1b:
			k = readEscape(r)
		default:
			k = KeyUnknown
		}
		select {
		case out <- k:
		case <-ctx.Done():
			return
		}
	}
}

func readEscape(r *bufio.Reader) Key {
	b, err := r.ReadByte()
	if err != nil {
		return KeyEsc
	}
	if b != '[' && b != 'O' {
		_ = r.UnreadByte()
		return KeyEsc
	}
	params := make([]byte, 0, 8)
	for len(params) < 8 {
		c, err := r.ReadByte()
		if err != nil {
			return KeyUnknown
		}
		if c >= 0x40 && c <= 0x7e {
			switch c {
			case 'C':
				return KeyRight
			case 'D':
				return KeyLeft
			case 'A':
				return KeyUp
			case 'B':
				return KeyDown
			}
			return KeyUnknown
		}
		params = append(params, c)
	}
	return KeyUnknown
}
