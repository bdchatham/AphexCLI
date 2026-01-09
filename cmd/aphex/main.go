package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bdchatham/AphexCLI/internal/commands"
	"github.com/urfave/cli/v3"
)

var (
	// Version is set at build time via -ldflags
	Version = "dev"
)

func main() {
	app := &cli.Command{
		Name:    "aphex",
		Usage:   "Manage Tekton pipelines on the Arbiter platform",
		Version: Version,
		Commands: []*cli.Command{
			commands.AuthCommand(),
			commands.PipelineCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Default action when no subcommands are provided
			return cli.ShowAppHelp(cmd)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
