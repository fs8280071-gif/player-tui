package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fs8280071-gif/player-tui/internal/app"
	"github.com/fs8280071-gif/player-tui/internal/cli"
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
	if err := app.Run(context.Background(), opts.Spec); err != nil {
		fmt.Fprintf(os.Stderr, "player-tui: %v\n", err)
		var exitErr *app.Error
		if errors.As(err, &exitErr) {
			return exitErr.Code
		}
		return 1
	}
	return 0
}
