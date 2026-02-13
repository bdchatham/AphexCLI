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
				Name:      "get",
				Usage:     "Get a knowledge base",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "organization",
						Usage:    "Organization name",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (table, json, yaml)",
						Value:   "table",
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
				},
				Action: knowledgeBaseGetAction,
			},
			{
				Name:  "list",
				Usage: "List knowledge bases",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "organization",
						Usage: "Organization name (lists all if omitted)",
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
				Action: knowledgeBaseListAction,
			},
			{
				Name:      "create",
				Usage:     "Create a knowledge base",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "organization",
						Usage:    "Organization name (must reference an existing Organization resource)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "cli-input-json",
						Usage: "Path to JSON file with complete knowledge base specification",
					},
					&cli.StringFlag{
						Name:  "cli-input-yaml",
						Usage: "Path to YAML file with complete knowledge base specification",
					},
					&cli.StringFlag{
						Name:  "repo-url",
						Usage: "Repository URL (any Git provider)",
					},
					&cli.StringFlag{
						Name:  "branch",
						Usage: "Git branch to track",
						Value: "mainline",
					},
					&cli.StringFlag{
						Name:  "source-type",
						Usage: "Source type (docs, code)",
						Value: "code",
					},
					&cli.StringFlag{
						Name:  "path",
						Usage: "Path within repository to process",
					},
					&cli.StringFlag{
						Name:  "mcp-image",
						Usage: "MCP server container image (required for MCP)",
					},
					&cli.IntFlag{
						Name:  "mcp-port",
						Usage: "MCP server port (required for MCP)",
					},
					&cli.IntFlag{
						Name:  "mcp-replicas",
						Usage: "MCP server replicas",
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
						Name:     "organization",
						Usage:    "Organization name",
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
				},
				Action: knowledgeBaseDeleteAction,
			},
			{
				Name:  "generate-spec",
				Usage: "Generate knowledge base specification template",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (json, yaml)",
						Value:   "yaml",
					},
				},
				Action: knowledgeBaseGenerateSpecAction,
			},
		},
	}
}

func knowledgeBaseGetAction(ctx context.Context, cmd *cli.Command) error {
	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("knowledge base name is required as argument")
	}

	return knowledgebase.Get(ctx, client, knowledgebase.GetOptions{
		Name:         args.First(),
		Organization: cmd.String("organization"),
		OutputFormat: output.Format(cmd.String("output")),
	})
}

func knowledgeBaseListAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	opts := knowledgebase.ListOptions{
		Organization: cmd.String("organization"),
		OutputFormat: output.Format(cmd.String("output")),
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
	if args.Len() == 0 && cmd.String("cli-input-json") == "" && cmd.String("cli-input-yaml") == "" {
		return fmt.Errorf("knowledge base name is required as argument (or use --cli-input-json/--cli-input-yaml)")
	}

	opts := knowledgebase.CreateOptions{
		Name:          args.First(),
		Organization:  cmd.String("organization"),
		InputJSONFile: cmd.String("cli-input-json"),
		InputYAMLFile: cmd.String("cli-input-yaml"),
		RepoURL:       cmd.String("repo-url"),
		Branch:        cmd.String("branch"),
		SourceType:    cmd.String("source-type"),
		DocsPath:      cmd.String("path"),
		MCPImage:      cmd.String("mcp-image"),
		MCPPort:       int32(cmd.Int("mcp-port")),
		MCPReplicas:   int32(cmd.Int("mcp-replicas")),
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
		Name:         args.First(),
		Organization: cmd.String("organization"),
		Force:        cmd.Bool("force"),
	}

	log.Debugf("Deleting knowledge base: %s", opts.Name)
	return knowledgebase.Delete(ctx, client, opts)
}

func knowledgeBaseGenerateSpecAction(ctx context.Context, cmd *cli.Command) error {
	return knowledgebase.GenerateSpec(ctx, output.Format(cmd.String("output")))
}
