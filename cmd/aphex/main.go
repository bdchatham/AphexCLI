package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bdchatham/AphexCLI/internal/commands"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/urfave/cli/v3"
)

var (
	// Version is set at build time via -ldflags
	Version = "dev"
)

func main() {
	// Parse flags manually to determine log level before creating context
	logLevel := "info" // default
	verbose := false
	
	// Simple flag parsing to extract log-level and verbose before cli.Run
	for i, arg := range os.Args {
		if arg == "--log-level" || arg == "-l" {
			if i+1 < len(os.Args) {
				logLevel = os.Args[i+1]
			}
		} else if strings.HasPrefix(arg, "--log-level=") {
			logLevel = strings.TrimPrefix(arg, "--log-level=")
		} else if arg == "--verbose" {
			verbose = true
		}
	}
	
	// Verbose flag takes precedence
	if verbose {
		logLevel = "debug"
	}
	
	// Create logger and inject into context
	log := logger.NewLogger(logLevel)
	ctx := logger.WithLogger(context.Background(), log)

	app := &cli.Command{
		Name:    "aphex",
		Usage:   "Manage Tekton pipelines on the Aphex platform",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Usage:   "Set log level (debug, info, warn, error)",
				Value:   "info",
				Aliases: []string{"l"},
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose output (equivalent to --log-level=debug)",
			},
		},
		Commands: []*cli.Command{
			commands.AuthCommand(),
			commands.OrganizationCommand(),
			commands.PipelineCommand(),
			commands.SecretCommand(),
			commands.KnowledgeBaseCommand(),
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// Default action when no subcommands are provided
			return cli.ShowAppHelp(cmd)
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
