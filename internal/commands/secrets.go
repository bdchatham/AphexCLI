package commands

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/secrets"
	"github.com/urfave/cli/v3"
)

func SecretCommand() *cli.Command {
	return &cli.Command{
		Name:  "secret",
		Usage: "Organization secret management",
		Commands: []*cli.Command{
			{
				Name:      "set",
				Usage:     "Set organization secrets (key=value pairs)",
				ArgsUsage: "key=value [key=value ...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "org",
						Usage:    "Organization name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
				},
				Action: secretsSetAction,
			},
			{
				Name:  "list",
				Usage: "List organization secret keys",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "org",
						Usage:    "Organization name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
				},
				Action: secretsListAction,
			},
			{
				Name:      "delete",
				Usage:     "Delete organization secrets",
				ArgsUsage: "key [key ...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "org",
						Usage:    "Organization name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
				},
				Action: secretsDeleteAction,
			},
		},
	}
}

func secretsSetAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("at least one key=value pair is required")
	}

	secretMap, err := secrets.ParseSecrets(args)
	if err != nil {
		return err
	}

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	opts := secrets.SetOptions{
		OrgName: cmd.String("org"),
		Secrets: secretMap,
	}

	log.Debugf("Setting secrets for organization %q", opts.OrgName)
	return secrets.Set(ctx, client, opts)
}

func secretsListAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	opts := secrets.ListOptions{
		OrgName: cmd.String("org"),
	}

	log.Debugf("Listing secrets for organization %q", opts.OrgName)
	return secrets.List(ctx, client, opts)
}

func secretsDeleteAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	args := cmd.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("at least one key is required")
	}

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	opts := secrets.DeleteOptions{
		OrgName: cmd.String("org"),
		Keys:    args,
	}

	log.Debugf("Deleting secrets for organization %q", opts.OrgName)
	return secrets.Delete(ctx, client, opts)
}
