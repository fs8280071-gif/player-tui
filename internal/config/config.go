// Package config loads player-tui settings and resume state.
package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Audio  bool
	Render string
	Theme  string
	Volume int
	Resume bool
	Dither bool
}

func Default() Config {
	return Config{
		Render: "auto",
		Theme:  "catppuccin",
		Volume: 100,
		Resume: true,
	}
}

func DefaultPath() string {
	if p := strings.TrimSpace(os.Getenv("PLAYER_TUI_CONFIG")); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(dir, "player-tui", "config.toml")
}

func Load(path string) (Config, error) {
	cfg := Default()
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		key, value, ok := splitKeyValue(line)
		if !ok {
			return cfg, fmt.Errorf("%s:%d: expected key = value", path, lineNo)
		}
		if err := cfg.set(key, value); err != nil {
			return cfg, fmt.Errorf("%s:%d: %w", path, lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func splitKeyValue(line string) (string, string, bool) {
	key, rest, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), strings.TrimSpace(stripComment(rest)), true
}

func stripComment(s string) string {
	quoted := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quoted != 0:
			if c == '\\' && quoted == '"' {
				i++
			} else if c == quoted {
				quoted = 0
			}
		case c == '"' || c == '\'':
			quoted = c
		case c == '#':
			return s[:i]
		}
	}
	return s
}

func (c *Config) set(key, raw string) error {
	value := strings.TrimSpace(raw)
	switch strings.ToLower(key) {
	case "audio":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("audio must be true or false")
		}
		c.Audio = v
	case "dither":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("dither must be true or false")
		}
		c.Dither = v
	case "resume":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("resume must be true or false")
		}
		c.Resume = v
	case "render":
		v, err := parseString(value)
		if err != nil {
			return err
		}
		if v == "" {
			return errors.New("render cannot be empty")
		}
		c.Render = v
	case "theme":
		v, err := parseString(value)
		if err != nil {
			return err
		}
		if v == "" {
			return errors.New("theme cannot be empty")
		}
		c.Theme = v
	case "volume":
		v, err := strconv.Atoi(value)
		if err != nil || v < 0 || v > 200 {
			return errors.New("volume must be an integer from 0 to 200")
		}
		c.Volume = v
	default:
		return nil
	}
	return nil
}

func parseString(value string) (string, error) {
	if len(value) < 2 {
		return "", errors.New("string value must be quoted")
	}
	q := value[0]
	if (q != '"' && q != '\'') || value[len(value)-1] != q {
		return "", errors.New("string value must be quoted")
	}
	body := value[1 : len(value)-1]
	if q == '\'' {
		return body, nil
	}
	out, err := strconv.Unquote(value)
	if err != nil {
		return "", errors.New("invalid quoted string")
	}
	return out, nil
}
