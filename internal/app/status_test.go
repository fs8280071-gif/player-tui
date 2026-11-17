package app

import (
	"strings"
	"testing"
	"time"

	"github.com/fs8280071-gif/player-tui/internal/engine"
	"github.com/fs8280071-gif/player-tui/internal/theme"
)

func TestBuildStatusFields(t *testing.T) {
	th, _ := theme.Parse("nord")
	snap := engine.Snapshot{
		State: engine.StatePlaying, Time: 75 * time.Second,
		Duration: 10 * time.Minute, Volume: 55, Rate: 1.5,
	}
	out := string(buildStatus(nil, "clip.mp4", snap, 120, 24, false, "halfblock", snap.Rate, "", th, true))
	for _, want := range []string{"clip.mp4", "01:15", "10:00", "playing", "vol 55%", "x1.50", "halfblock"} {
		if !strings.Contains(out, want) {
			t.Errorf("status missing %q: %q", want, out)
		}
	}
	out = string(buildStatus(nil, "clip.mp4", snap, 120, 24, true, "halfblock", 1, "", th, false))
	if !strings.Contains(out, "vol off") {
		t.Errorf("audio-off status = %q", out)
	}
}
