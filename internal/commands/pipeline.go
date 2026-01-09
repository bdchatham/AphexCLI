package commands

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/auth"
	"github.com/bdchatham/AphexCLI/pkg/interactive"
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
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Pipeline definition file path",
					},
					&cli.StringFlag{
						Name:    "namespace",
						Aliases: []string{"n"},
						Usage:   "Kubernetes namespace",
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
						Name:    "namespace",
						Aliases: []string{"n"},
						Usage:   "Kubernetes namespace",
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
						Name:    "namespace",
						Aliases: []string{"n"},
						Usage:   "Kubernetes namespace",
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.BoolFlag{
						Name:  "all-namespaces",
						Usage: "List pipelines across all namespaces",
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
	// Create Kubernetes client first to get default namespace
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	var pipelineName, namespace, filePath string

	// Check if we need interactive mode
	args := cmd.Args()
	if args.Len() == 0 || cmd.String("file") == "" {
		// Enter interactive mode
		pipelineName, namespace, filePath, err = interactive.PromptPipelineCreation(client.Namespace)
		if err != nil {
			return err
		}
	} else {
		// Use command line arguments
		pipelineName = args.First()
		namespace = cmd.String("namespace")
		filePath = cmd.String("file")
		
		if filePath == "" {
			return fmt.Errorf("--file flag is required")
		}
	}

	// Create pipeline
	opts := pipeline.CreateOptions{
		Name:      pipelineName,
		Namespace: namespace,
		FilePath:  filePath,
		Verbose:   cmd.Bool("verbose"),
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
		Name:      pipelineName,
		Namespace: cmd.String("namespace"),
		Force:     cmd.Bool("force"),
		Verbose:   cmd.Bool("verbose"),
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
		Namespace:     cmd.String("namespace"),
		AllNamespaces: cmd.Bool("all-namespaces"),
		OutputFormat:  outputFormat,
		Quiet:         cmd.Bool("quiet"),
		Verbose:       cmd.Bool("verbose"),
	}

	return pipeline.List(ctx, client, opts)
}
