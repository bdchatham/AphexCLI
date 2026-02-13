package commands

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/auth"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/output"
	"github.com/bdchatham/AphexCLI/pkg/pipeline"
	"github.com/urfave/cli/v3"
)

func PipelineCommand() *cli.Command {
	return &cli.Command{
		Name:   "pipeline",
		Usage:  "Pipeline management",
		Before: pipelineBeforeAction,
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create pipeline instance",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "Pipeline definition file path",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "organization",
						Usage:    "Aphex organization name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "repo-org",
						Usage:    "GitHub organization name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "repo-name",
						Usage:    "GitHub repository name",
						Required: true,
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
				Action: pipelineCreateAction,
			},
			{
				Name:      "delete",
				Usage:     "Delete pipeline instance",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "organization",
						Usage:    "Aphex organization name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Skip confirmation prompt",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (table, json, yaml)",
						Value:   "table",
					},
				},
				Action: pipelineDeleteAction,
			},
			{
				Name:  "list",
				Usage: "List pipeline instances",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "organization",
						Usage: "Aphex organization name (lists all if omitted)",
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
					&cli.BoolFlag{
						Name:  "quiet",
						Usage: "Suppress non-essential output",
					},
				},
				Action: pipelineListAction,
			},
		},
	}
}

func pipelineBeforeAction(ctx context.Context, cmd *cli.Command) error {
	subcommand := cmd.Args().First()
	if subcommand == "list" {
		return nil
	}

	if subcommand != "create" && subcommand != "delete" {
		return nil
	}

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	org := cmd.String("organization")
	if org == "" {
		return nil
	}
	namespace := "org-" + org

	switch subcommand {
	case "create":
		return auth.CheckPipelineCreate(ctx, client, namespace)
	case "delete":
		return auth.CheckPipelineDelete(ctx, client, namespace)
	}

	return nil
}

func pipelineCreateAction(ctx context.Context, cmd *cli.Command) error {
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("pipeline name is required as argument")
	}

	return pipeline.Create(ctx, client, pipeline.CreateOptions{
		Name:         args.First(),
		FilePath:     cmd.String("file"),
		Organization: cmd.String("organization"),
		RepoOrg:      cmd.String("repo-org"),
		RepoName:     cmd.String("repo-name"),
	})
}

func pipelineDeleteAction(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("pipeline name is required")
	}

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return pipeline.Delete(ctx, client, pipeline.DeleteOptions{
		Name:         args.First(),
		Organization: cmd.String("organization"),
		Force:        cmd.Bool("force"),
	})
}

func pipelineListAction(ctx context.Context, cmd *cli.Command) error {
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return pipeline.List(ctx, client, pipeline.ListOptions{
		Organization: cmd.String("organization"),
		OutputFormat: output.Format(cmd.String("output")),
		Quiet:        cmd.Bool("quiet"),
	})
}
