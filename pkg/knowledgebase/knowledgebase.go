package knowledgebase

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

	log.Debug("Listing knowledge bases...")

	kbList := &platformv1alpha1.KnowledgeBaseList{}
	if err := aphexClient.List(ctx, kbList, &client.ListOptions{}); err != nil {
		return k8s.FormatError(err)
	}

	if len(kbList.Items) == 0 {
		if !opts.Quiet {
			fmt.Println("No knowledge bases found")
		}
		return nil
	}

	switch opts.OutputFormat {
	case output.FormatJSON:
		return formatJSON(kbList.Items)
	case output.FormatYAML:
		return formatYAML(kbList.Items)
	case output.FormatTable:
		fallthrough
	default:
		return formatTable(kbList.Items, opts.Quiet)
	}
}


type CreateOptions struct {
	Name          string
	Namespace     string
	InputJSONFile string
	InputYAMLFile string
	RepoURL       string
	Branch        string
	DocsPath      string
	MCPImage      string
	MCPPort       int32
	MCPReplicas   int32
}

func Create(ctx context.Context, k8sClient *k8s.Client, opts CreateOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	var kb *platformv1alpha1.KnowledgeBase

	if opts.InputJSONFile != "" {
		kb, err = loadKnowledgeBaseFromJSON(opts.InputJSONFile)
		if err != nil {
			return fmt.Errorf("failed to load knowledge base from JSON: %w", err)
		}
	} else if opts.InputYAMLFile != "" {
		kb, err = loadKnowledgeBaseFromYAML(opts.InputYAMLFile)
		if err != nil {
			return fmt.Errorf("failed to load knowledge base from YAML: %w", err)
		}
	} else {
		if opts.RepoURL == "" {
			return fmt.Errorf("--repo-url is required when not using --cli-input-json/--cli-input-yaml")
		}

		kb = &platformv1alpha1.KnowledgeBase{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.Name,
				Namespace: opts.Namespace,
			},
			Spec: platformv1alpha1.KnowledgeBaseSpec{
				DisplayName: opts.Name,
				Repositories: []platformv1alpha1.Repository{
					{
						URL:    opts.RepoURL,
						Branch: opts.Branch,
						Paths:  []string{opts.DocsPath},
					},
				},
			},
		}

		if opts.MCPImage != "" && opts.MCPPort != 0 {
			kb.Spec.MCP = &platformv1alpha1.MCPConfig{
				Image: opts.MCPImage,
				Port:  opts.MCPPort,
			}
			if opts.MCPReplicas > 0 {
				kb.Spec.MCP.Replicas = opts.MCPReplicas
			}
		}
	}

	log.Debugf("Creating KnowledgeBase %q in namespace %q", kb.Name, kb.Namespace)

	if err := aphexClient.Create(ctx, kb); err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Knowledge base %q created successfully in namespace %q\n", kb.Name, kb.Namespace)
	if len(kb.Spec.Repositories) > 0 {
		fmt.Printf("Repositories: %d\n", len(kb.Spec.Repositories))
		for i, repo := range kb.Spec.Repositories {
			fmt.Printf("  [%d] %s (branch: %s)\n", i+1, repo.URL, repo.Branch)
		}
	}
	
	if kb.Spec.MCP != nil {
		fmt.Println("MCP server: enabled")
		fmt.Printf("  Image: %s\n", kb.Spec.MCP.Image)
		fmt.Printf("  Port: %d\n", kb.Spec.MCP.Port)
		if kb.Spec.MCP.Replicas > 0 {
			fmt.Printf("  Replicas: %d\n", kb.Spec.MCP.Replicas)
		}
	}

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

	kb := &platformv1alpha1.KnowledgeBase{}
	objectKey := client.ObjectKey{
		Name:      opts.Name,
		Namespace: opts.Namespace,
	}

	if err := aphexClient.Get(ctx, objectKey, kb); err != nil {
		return k8s.FormatError(err)
	}

	if !opts.Force {
		fmt.Printf("This will delete knowledge base %q in namespace %q\n", opts.Name, opts.Namespace)
		fmt.Printf("\nTracked repositories:\n")
		for _, repo := range kb.Spec.Repositories {
			fmt.Printf("  - %s (branch: %s)\n", repo.URL, repo.Branch)
		}
		fmt.Printf("\nThis action cannot be undone. Continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	log.Debugf("Deleting KnowledgeBase %q from namespace %q", opts.Name, opts.Namespace)

	if err := aphexClient.Delete(ctx, kb); err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Knowledge base %q deleted successfully\n", opts.Name)
	return nil
}

func GenerateSpec(ctx context.Context, format output.Format) error {
	template := platformv1alpha1.KnowledgeBase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "aphex.io/v1alpha1",
			Kind:       "KnowledgeBase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-kb",
			Namespace: "default",
		},
		Spec: platformv1alpha1.KnowledgeBaseSpec{
			DisplayName: "Example Knowledge Base",
			Repositories: []platformv1alpha1.Repository{
				{
					URL:    "https://github.com/org/repo",
					Branch: "main",
					Paths:  []string{".kiro/docs"},
				},
			},
			MCP: &platformv1alpha1.MCPConfig{
				Image:    "ghcr.io/bdchatham/archon-mcp-server:latest",
				Port:     8090,
				Replicas: 1,
			},
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

func loadKnowledgeBaseFromJSON(path string) (*platformv1alpha1.KnowledgeBase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var kb platformv1alpha1.KnowledgeBase
	if err := json.Unmarshal(data, &kb); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &kb, nil
}

func loadKnowledgeBaseFromYAML(path string) (*platformv1alpha1.KnowledgeBase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var kb platformv1alpha1.KnowledgeBase
	if err := yaml.Unmarshal(data, &kb); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &kb, nil
}

