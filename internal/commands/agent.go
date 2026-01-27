package commands

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/agent"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/output"
	"github.com/urfave/cli/v3"
)

func AgentCommand() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "Agent management",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List agents",
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
				Action: agentListAction,
			},
			{
				Name:      "create",
				Usage:     "Create an agent",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "namespace",
						Usage: "Namespace for the agent",
						Value: "default",
					},
					&cli.StringFlag{
						Name:  "cli-input-json",
						Usage: "Path to JSON file with complete agent specification",
					},
					&cli.StringFlag{
						Name:  "cli-input-yaml",
						Usage: "Path to YAML file with complete agent specification",
					},
					&cli.StringFlag{
						Name:  "model",
						Usage: "Model name (e.g., meta-llama/Llama-3.1-70B-Instruct)",
					},
					&cli.StringFlag{
						Name:  "provider",
						Usage: "Model provider (vllm, openai, anthropic, bedrock)",
						Value: "vllm",
					},
					&cli.IntFlag{
						Name:  "gpu-count",
						Usage: "Number of GPUs to allocate",
					},
					&cli.StringFlag{
						Name:  "quantization",
						Usage: "Quantization method (awq, gptq)",
					},
					&cli.StringFlag{
						Name:  "kb-name",
						Usage: "KnowledgeBase name to reference",
					},
					&cli.StringFlag{
						Name:  "kb-namespace",
						Usage: "KnowledgeBase namespace (defaults to agent namespace)",
					},
					&cli.BoolFlag{
						Name:  "orchestration",
						Usage: "Enable orchestrator for unified RAG endpoint",
					},
					&cli.StringFlag{
						Name:  "kubeconfig",
						Usage: "Path to kubeconfig file",
					},
				},
				Action: agentCreateAction,
			},
			{
				Name:      "delete",
				Usage:     "Delete an agent",
				ArgsUsage: "[name]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "namespace",
						Usage: "Namespace of the agent",
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
				Action: agentDeleteAction,
			},
			{
				Name:  "generate-spec",
				Usage: "Generate agent specification template",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Output format (json, yaml)",
						Value:   "yaml",
					},
				},
				Action: agentGenerateSpecAction,
			},
		},
	}
}

func agentListAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	outputFormat := output.Format(cmd.String("output"))

	opts := agent.ListOptions{
		OutputFormat: outputFormat,
		Quiet:        cmd.Bool("quiet"),
	}

	log.Debug("Listing agents")
	return agent.List(ctx, client, opts)
}

func agentCreateAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 && cmd.String("cli-input-json") == "" && cmd.String("cli-input-yaml") == "" {
		return fmt.Errorf("agent name is required as argument (or use --cli-input-json/--cli-input-yaml)")
	}

	opts := agent.CreateOptions{
		Name:          args.First(),
		Namespace:     cmd.String("namespace"),
		InputJSONFile: cmd.String("cli-input-json"),
		InputYAMLFile: cmd.String("cli-input-yaml"),
		Model:         cmd.String("model"),
		Provider:      cmd.String("provider"),
		GPUCount:      int32(cmd.Int("gpu-count")),
		Quantization:  cmd.String("quantization"),
		KBName:        cmd.String("kb-name"),
		KBNamespace:   cmd.String("kb-namespace"),
		Orchestration: cmd.Bool("orchestration"),
	}

	log.Debugf("Creating agent: %s", opts.Name)
	return agent.Create(ctx, client, opts)
}

func agentDeleteAction(ctx context.Context, cmd *cli.Command) error {
	log := logger.GetLogger(ctx)

	client, err := k8s.NewClient(cmd.String("kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	args := cmd.Args()
	if args.Len() == 0 {
		return fmt.Errorf("agent name is required as argument")
	}

	opts := agent.DeleteOptions{
		Name:      args.First(),
		Namespace: cmd.String("namespace"),
		Force:     cmd.Bool("force"),
	}

	log.Debugf("Deleting agent: %s", opts.Name)
	return agent.Delete(ctx, client, opts)
}

func agentGenerateSpecAction(ctx context.Context, cmd *cli.Command) error {
	outputFormat := output.Format(cmd.String("output"))
	return agent.GenerateSpec(ctx, outputFormat)
}
