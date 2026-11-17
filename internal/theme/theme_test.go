package theme

import (
	"strings"
	"testing"
)

func TestParseThemes(t *testing.T) {
	for _, name := range Names() {
		th, err := Parse(name)
		if err != nil || th.Name != name {
			t.Errorf("Parse(%q) = %+v, %v", name, th, err)
		}
	}
	if _, err := Parse("solarized"); err == nil {
		t.Fatal("unknown theme accepted")
	}
}

func TestNoColorDisablesPalette(t *testing.T) {
	th, _ := Parse("nord")
	if got := th.StatusStyle(true); got != "\x1b[7m" {
		t.Fatalf("no-color style = %q", got)
	}
	got := th.StatusStyle(false)
	if !strings.Contains(got, "38;2;") || !strings.Contains(got, "48;2;") {
		t.Fatalf("color style = %q", got)
	}
}
