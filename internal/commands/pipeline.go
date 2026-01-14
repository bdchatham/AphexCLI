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

// PipelineCommand returns the pipeline command with subcommands
func PipelineCommand() *cli.Command {
	return &cli.Command{
		Name:  "pipeline",
		Usage: "Pipeline management",
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
						Name:     "aphex-org",
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
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Verbose output",
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
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Verbose output",
					},
				},
				Action: pipelineDeleteAction,
			},
			{
				Name:      "list",
				Usage:     "List pipeline instances",
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
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Verbose output",
					},
				},
				Action: pipelineListAction,
			},
		},
	}
}

func pipelineBeforeAction(ctx context.Context, cmd *cli.Command) error {
	// Skip preflight checks for read-only operations
	subcommand := cmd.Args().First()
	if subcommand == "list" {
		return nil
	}

	// Only check permissions for create/delete operations
	if subcommand != "create" && subcommand != "delete" {
		return nil
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get namespace (use default if not specified)
	namespace := cmd.String("namespace")
	if namespace == "" {
		namespace = client.Namespace
	}

	// Perform preflight authorization check
	switch subcommand {
	case "create":
		return auth.CheckPipelineCreate(ctx, client, namespace)
	case "delete":
		return auth.CheckPipelineDelete(ctx, client, namespace)
	}

	return nil
}

func pipelineCreateAction(ctx context.Context, cmd *cli.Command) error {
	// Create Kubernetes client
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get pipeline name from arguments
	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("pipeline name is required as argument")
	}

	// Create pipeline
	opts := pipeline.CreateOptions{
		Name:        args.First(),
		FilePath:    cmd.String("file"),
		AphexOrg:    cmd.String("aphex-org"),
		RepoOrg:     cmd.String("repo-org"),
		RepoName:    cmd.String("repo-name"),
		TenantName:  args.First(), // Same as pipeline name
		IngressHost: "webhooks.homelab.local", // Hardcoded
		Verbose:     cmd.Bool("verbose"),
	}

	return pipeline.Create(ctx, client, opts)
}

func pipelineDeleteAction(ctx context.Context, cmd *cli.Command) error {
	// Get pipeline name from arguments
	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("pipeline name is required")
	}
	pipelineName := args.First()

	// Create Kubernetes client
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Delete pipeline
	opts := pipeline.DeleteOptions{
		Name:    pipelineName,
		Force:   cmd.Bool("force"),
		Verbose: cmd.Bool("verbose"),
	}

	return pipeline.Delete(ctx, client, opts)
}

func pipelineListAction(ctx context.Context, cmd *cli.Command) error {
	// Create Kubernetes client
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Parse output format
	outputFormat := output.Format(cmd.String("output"))
	
	// List pipelines
	opts := pipeline.ListOptions{
		OutputFormat: outputFormat,
		Quiet:        cmd.Bool("quiet"),
		Verbose:      cmd.Bool("verbose"),
	}

	return pipeline.List(ctx, client, opts)
}
