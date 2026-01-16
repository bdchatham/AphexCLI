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
