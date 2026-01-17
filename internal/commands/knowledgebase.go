package commands

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/knowledgebase"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/output"
	"github.com/urfave/cli/v3"
)

func KnowledgeBaseCommand() *cli.Command {
	return &cli.Command{
		Name:  "knowledgebase",
		Usage: "Knowledge base management",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List knowledge bases",
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
				Action: knowledgeBaseListAction,
			},
			{
				Name:      "create",
				Usage:     "Create a knowledge base",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "namespace",
						Usage: "Namespace for the knowledge base",
						Value: "default",
					},
					&cli.StringFlag{
						Name:  "repo-url",
						Usage: "Repository URL (https://github.com/org/repo)",
					},
					&cli.StringFlag{
						Name:  "branch",
						Usage: "Git branch to track",
						Value: "main",
					},
					&cli.StringFlag{
						Name:  "docs-path",
						Usage: "Documentation path within repository",
						Value: ".kiro/docs",
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
				},
				Action: knowledgeBaseCreateAction,
			},
			{
				Name:      "delete",
				Usage:     "Delete a knowledge base",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "namespace",
						Usage: "Namespace of the knowledge base",
						Value: "default",
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Skip confirmation prompt",
					},
				},
				Action: knowledgeBaseDeleteAction,
			},
		},
	}
}

func knowledgeBaseListAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	outputFormat := output.Format(cmd.String("output"))

	opts := knowledgebase.ListOptions{
		OutputFormat: outputFormat,
		Quiet:        cmd.Bool("quiet"),
	}

	log.Debug("Listing knowledge bases")
	return knowledgebase.List(ctx, client, opts)
}


func knowledgeBaseCreateAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("knowledge base name is required as argument")
	}

	opts := knowledgebase.CreateOptions{
		Name:      args.First(),
		Namespace: cmd.String("namespace"),
		RepoURL:   cmd.String("repo-url"),
		Branch:    cmd.String("branch"),
		DocsPath:  cmd.String("docs-path"),
	}

	log.Debugf("Creating knowledge base: %s", opts.Name)
	return knowledgebase.Create(ctx, client, opts)
}


func knowledgeBaseDeleteAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("knowledge base name is required as argument")
	}

	opts := knowledgebase.DeleteOptions{
		Name:      args.First(),
		Namespace: cmd.String("namespace"),
		Force:     cmd.Bool("force"),
	}

	log.Debugf("Deleting knowledge base: %s", opts.Name)
	return knowledgebase.Delete(ctx, client, opts)
}
