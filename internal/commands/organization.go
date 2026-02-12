package commands

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/auth"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/organization"
	"github.com/urfave/cli/v3"
)

func OrganizationCommand() *cli.Command {
	return &cli.Command{
		Name:   "organization",
		Usage:  "Organization management",
		Before: organizationBeforeAction,
		Commands: []*cli.Command{
			{
				Name:      "bootstrap",
				Usage:     "Bootstrap a new organization",
				ArgsUsage: "[organization-name]",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "admin-email",
						Usage:    "Admin email address(es) for the organization (repeatable)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "display-name",
						Usage: "Human-readable organization name (defaults to organization name)",
					},
					&cli.StringFlag{
						Name:  "webhook-secret",
						Usage: "GitHub webhook secret (auto-generated if not provided)",
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (table, json, yaml)",
						Value:   "table",
					},
				},
				Action: organizationBootstrapAction,
			},
			{
				Name:      "get",
				Usage:     "Get details of a specific organization",
				ArgsUsage: "[organization-name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (table, json, yaml)",
						Value:   "table",
					},
				},
				Action: organizationGetAction,
			},
			{
				Name:  "list",
				Usage: "List organizations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (table, json, yaml)",
						Value:   "table",
					},
					&cli.BoolFlag{
						Name:  "quiet",
						Usage: "Suppress non-essential output",
					},
				},
				Action: organizationListAction,
			},
			{
				Name:      "delete",
				Usage:     "Delete an organization and all its resources",
				ArgsUsage: "[organization-name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Skip confirmation prompt",
					},
				},
				Action: organizationDeleteAction,
			},
		},
	}
}

func organizationBeforeAction(ctx context.Context, cmd *cli.Command) error {
	subcommand := cmd.Args().First()
	if subcommand == "list" || subcommand == "get" {
		return nil
	}

	if subcommand != "bootstrap" {
		return nil
	}

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return auth.CheckOrganizationBootstrap(ctx, client)
}

func organizationBootstrapAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("organization name is required as argument")
	}

	opts := organization.BootstrapOptions{
		Name:          args.First(),
		DisplayName:   cmd.String("display-name"),
		AdminEmails:   cmd.StringSlice("admin-email"),
		WebhookSecret: cmd.String("webhook-secret"),
		OutputFormat:  cmd.String("output"),
	}

	log.Debugf("Bootstrapping organization with options: %+v", opts)
	return organization.Bootstrap(ctx, client, opts)
}

func organizationGetAction(ctx context.Context, cmd *cli.Command) error {
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("organization name is required as argument")
	}

	opts := organization.GetOptions{
		Name:         args.First(),
		OutputFormat: cmd.String("output"),
	}

	return organization.Get(ctx, client, opts)
}

func organizationListAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	opts := organization.ListOptions{
		Quiet:        cmd.Bool("quiet"),
		OutputFormat: cmd.String("output"),
	}

	log.Debug("Listing organizations")
	return organization.List(ctx, client, opts)
}

func organizationDeleteAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("organization name is required as argument")
	}

	opts := organization.DeleteOptions{
		Name:  args.First(),
		Force: cmd.Bool("force"),
	}

	log.Debugf("Deleting organization: %s", opts.Name)
	return organization.Delete(ctx, client, opts)
}
