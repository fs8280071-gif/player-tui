package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fs8280071-gif/player-tui/internal/app"
	"github.com/fs8280071-gif/player-tui/internal/cli"
	xterm "golang.org/x/term"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	opts, err := cli.Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "player-tui: %v\n", err)
		var targetErr *cli.TargetError
		if errors.As(err, &targetErr) {
			return 3
		}
		fmt.Fprintf(os.Stderr, "\n%s", cli.Usage)
		return 2
	}
	if opts.Help {
		fmt.Print(cli.Usage)
		return 0
	}

	if opts.Command == cli.CommandDoctor {
		w, h, err := xterm.GetSize(int(os.Stdout.Fd()))
		fmt.Print(app.Doctor(os.Getenv, app.DoctorOptions{
			Width: w, Height: h, TTY: err == nil, Render: opts.Render,
		}))
		return 0
	}

	if opts.Command == cli.CommandBench {
		stats, err := app.Bench(context.Background(), app.BenchOptions{
			Spec:   opts.Spec,
			Render: opts.Render,
			Dither: opts.Dither,
			Frames: opts.Frames,
			Cols:   opts.Cols,
			Rows:   opts.Rows,
		})
		if err != nil {
			return reportError(err)
		}
		fmt.Printf("render=%s source=%dx%d cells=%dx%d frames=%d wall=%.2fs display_fps=%.1f render_avg=%.2fms render_p50=%.2fms render_p95=%.2fms render_fps=%.0f bytes_per_frame=%.1fKB\n",
			stats.Mode,
			stats.Width, stats.Height,
			stats.Cols, stats.Rows,
			stats.Frames,
			stats.Wall.Seconds(),
			stats.DisplayFPS,
			stats.RenderAverage.Seconds()*1000,
			stats.RenderP50.Seconds()*1000,
			stats.RenderP95.Seconds()*1000,
			stats.RenderFPS,
			float64(stats.Bytes)/float64(stats.Frames)/1024,
		)
		return 0
	}

	if err := app.Run(context.Background(), app.Options{
		Spec:   opts.Spec,
		Specs:  opts.Specs,
		Render: opts.Render,
		Dither: opts.Dither,
		Audio:  opts.Audio,
	}); err != nil {
		return reportError(err)
	}
	return 0
}

func reportError(err error) int {
	fmt.Fprintf(os.Stderr, "player-tui: %v\n", err)
	var exitErr *app.Error
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}
	return 1
}
