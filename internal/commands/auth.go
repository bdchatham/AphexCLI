package commands

import (
	"context"

	"github.com/bdchatham/AphexCLI/pkg/auth"
	"github.com/bdchatham/AphexCLI/pkg/logger"
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
				},
				Action: authLoginAction,
			},
		},
	}
}

func authLoginAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)
	
	opts := auth.LoginOptions{
		KubeconfigPath: cmd.String("kubeconfig"),
	}

	return auth.Login(ctx, log, opts)
}
