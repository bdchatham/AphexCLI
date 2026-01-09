package commands

import (
	"context"

	"github.com/bdchatham/AphexCLI/pkg/auth"
	"github.com/urfave/cli/v3"
)

// AuthCommand returns the auth command with subcommands
func AuthCommand() *cli.Command {
	return &cli.Command{
		Name:  "auth",
		Usage: "Authentication management",
		Commands: []*cli.Command{
			{
				Name:  "login",
				Usage: "Authenticate with the platform",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Verbose output",
					},
				},
				Action: authLoginAction,
			},
		},
	}
}

func authLoginAction(ctx context.Context, cmd *cli.Command) error {
	opts := auth.LoginOptions{
		KubeconfigPath: cmd.String("kubeconfig"),
		Verbose:        cmd.Bool("verbose"),
	}

	return auth.Login(ctx, opts)
}
