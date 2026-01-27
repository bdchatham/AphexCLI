package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	platformv1alpha1 "github.com/bdchatham/AphexPlatformInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/output"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type ListOptions struct {
	OutputFormat output.Format
	Quiet        bool
}

func List(ctx context.Context, k8sClient *k8s.Client, opts ListOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	log.Debug("Listing agents...")

	agentList := &platformv1alpha1.AgentList{}
	if err := aphexClient.List(ctx, agentList, &client.ListOptions{}); err != nil {
		return k8s.FormatError(err)
	}

	if len(agentList.Items) == 0 {
		if !opts.Quiet {
			fmt.Println("No agents found")
		}
		return nil
	}

	switch opts.OutputFormat {
	case output.FormatJSON:
		return formatJSON(agentList.Items)
	case output.FormatYAML:
		return formatYAML(agentList.Items)
	case output.FormatTable:
		fallthrough
	default:
		return formatTable(agentList.Items, opts.Quiet)
	}
}

type CreateOptions struct {
	Name          string
	Namespace     string
	InputJSONFile string
	InputYAMLFile string
	Model         string
	Provider      string
	GPUCount      int32
	Quantization  string
	KBName        string
	KBNamespace   string
	Orchestration bool
}

func Create(ctx context.Context, k8sClient *k8s.Client, opts CreateOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	var agent *platformv1alpha1.Agent

	if opts.InputJSONFile != "" {
		agent, err = loadAgentFromJSON(opts.InputJSONFile)
		if err != nil {
			return fmt.Errorf("failed to load agent from JSON: %w", err)
		}
	} else if opts.InputYAMLFile != "" {
		agent, err = loadAgentFromYAML(opts.InputYAMLFile)
		if err != nil {
			return fmt.Errorf("failed to load agent from YAML: %w", err)
		}
	} else {
		if opts.Model == "" {
			return fmt.Errorf("--model is required when not using --cli-input-json/--cli-input-yaml")
		}

		agent = &platformv1alpha1.Agent{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.Name,
				Namespace: opts.Namespace,
			},
			Spec: platformv1alpha1.AgentSpec{
				DisplayName: opts.Name,
				Model: platformv1alpha1.ModelSpec{
					Provider:     opts.Provider,
					Name:         opts.Model,
					Quantization: opts.Quantization,
					GPUCount:     opts.GPUCount,
				},
			},
		}

		if opts.KBName != "" {
			kbNamespace := opts.KBNamespace
			if kbNamespace == "" {
				kbNamespace = opts.Namespace
			}
			agent.Spec.KnowledgeBase = &platformv1alpha1.KnowledgeBaseConfig{
				Name:      opts.KBName,
				Namespace: kbNamespace,
			}
		}

		if opts.Orchestration {
			agent.Spec.Orchestration = &platformv1alpha1.OrchestrationConfig{}
		}
	}

	log.Debugf("Creating agent %s in namespace %s", agent.Name, agent.Namespace)

	if err := aphexClient.Create(ctx, agent); err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Agent %s created successfully\n", agent.Name)
	return nil
}

type DeleteOptions struct {
	Name      string
	Namespace string
	Force     bool
}

func Delete(ctx context.Context, k8sClient *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	if !opts.Force {
		fmt.Printf("Are you sure you want to delete agent %s? (y/N): ", opts.Name)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	agent := &platformv1alpha1.Agent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
	}

	log.Debugf("Deleting agent %s from namespace %s", opts.Name, opts.Namespace)

	if err := aphexClient.Delete(ctx, agent); err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Agent %s deleted successfully\n", opts.Name)
	return nil
}

func GenerateSpec(ctx context.Context, format output.Format) error {
	template := platformv1alpha1.Agent{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "aphex.io/v1alpha1",
			Kind:       "Agent",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-agent",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AgentSpec{
			DisplayName: "Example Agent",
			Model: platformv1alpha1.ModelSpec{
				Provider:     "vllm",
				Name:         "meta-llama/Llama-3.1-70B-Instruct",
				Quantization: "awq",
				GPUCount:     4,
			},
			KnowledgeBase: &platformv1alpha1.KnowledgeBaseConfig{
				Name:      "platform-docs",
				Namespace: "archon-knowledge-base",
			},
			Orchestration: &platformv1alpha1.OrchestrationConfig{},
		},
	}

	switch format {
	case output.FormatJSON:
		data, err := json.MarshalIndent(template, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal template: %w", err)
		}
		fmt.Println(string(data))
	case output.FormatYAML:
		fallthrough
	default:
		data, err := yaml.Marshal(template)
		if err != nil {
			return fmt.Errorf("failed to marshal template: %w", err)
		}
		fmt.Println(string(data))
	}

	return nil
}

func loadAgentFromJSON(path string) (*platformv1alpha1.Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var agent platformv1alpha1.Agent
	if err := json.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &agent, nil
}

func loadAgentFromYAML(path string) (*platformv1alpha1.Agent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var agent platformv1alpha1.Agent
	if err := yaml.Unmarshal(data, &agent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &agent, nil
}
