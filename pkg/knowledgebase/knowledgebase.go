package knowledgebase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/output"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const platformSystemNamespace = "platform-system"

type GetOptions struct {
	Name         string
	OutputFormat output.Format
}

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
	default:
		return formatTable(kbList.Items, opts.Quiet)
	}
}

func Get(ctx context.Context, k8sClient *k8s.Client, opts GetOptions) error {
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	kb := &platformv1alpha1.KnowledgeBase{}
	key := client.ObjectKey{Name: opts.Name, Namespace: platformSystemNamespace}
	if err := aphexClient.Get(ctx, key, kb); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("knowledge base %q not found", opts.Name)
		}
		return k8s.FormatError(err)
	}

	switch opts.OutputFormat {
	case output.FormatJSON:
		return formatJSON([]platformv1alpha1.KnowledgeBase{*kb})
	case output.FormatYAML:
		return formatYAML([]platformv1alpha1.KnowledgeBase{*kb})
	default:
		return formatTable([]platformv1alpha1.KnowledgeBase{*kb}, false)
	}
}

type CreateOptions struct {
	Name          string
	Namespace     string
	Organization  string
	InputJSONFile string
	InputYAMLFile string
	RepoURL       string
	Branch        string
	SourceType    string
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
		kb, err = loadFromJSON(opts.InputJSONFile)
		if err != nil {
			return fmt.Errorf("failed to load knowledge base from JSON: %w", err)
		}
	} else if opts.InputYAMLFile != "" {
		kb, err = loadFromYAML(opts.InputYAMLFile)
		if err != nil {
			return fmt.Errorf("failed to load knowledge base from YAML: %w", err)
		}
	} else {
		if opts.RepoURL == "" {
			return fmt.Errorf("--repo-url is required when not using --cli-input-json/--cli-input-yaml")
		}
		if opts.Organization == "" {
			return fmt.Errorf("--organization is required")
		}

		sourceType := opts.SourceType
		if sourceType == "" {
			sourceType = "code"
		}

		namespace := opts.Namespace
		if namespace == "" {
			namespace = platformSystemNamespace
		}

		kb = &platformv1alpha1.KnowledgeBase{
			ObjectMeta: metav1.ObjectMeta{
				Name:      opts.Name,
				Namespace: namespace,
			},
			Spec: platformv1alpha1.KnowledgeBaseSpec{
				Name:         opts.Name,
				Organization: opts.Organization,
				Sources: []platformv1alpha1.Source{
					{
						URL:        opts.RepoURL,
						Branch:     opts.Branch,
						SourceType: sourceType,
						Paths:      []string{opts.DocsPath},
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

	err = progress.WithSpinner(fmt.Sprintf("Creating knowledge base %q", kb.Name), func() error {
		return aphexClient.Create(ctx, kb)
	})
	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Knowledge base %q created successfully\n", kb.Name)
	fmt.Printf("Organization: %s\n", kb.Spec.Organization)
	fmt.Printf("Sources: %d\n", len(kb.Spec.Sources))
	for i, source := range kb.Spec.Sources {
		fmt.Printf("  [%d] %s (branch: %s, type: %s)\n", i+1, source.URL, source.Branch, source.SourceType)
	}

	readyKB, err := waitForReady(ctx, aphexClient, kb.Name, kb.Namespace)
	if err != nil {
		fmt.Printf("\nStatus polling failed: %v\n", err)
		fmt.Printf("Check status with: aphex knowledgebase get %s\n", kb.Name)
		return nil
	}

	fmt.Printf("\nPhase: %s\n", readyKB.Status.Phase)
	fmt.Printf("Message: %s\n", readyKB.Status.Message)

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
	key := client.ObjectKey{Name: opts.Name, Namespace: opts.Namespace}
	if err := aphexClient.Get(ctx, key, kb); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("knowledge base %q not found in namespace %q", opts.Name, opts.Namespace)
		}
		return k8s.FormatError(err)
	}

	if !opts.Force {
		fmt.Printf("This will delete knowledge base %q in namespace %q\n", opts.Name, opts.Namespace)
		fmt.Printf("\nTracked sources:\n")
		for _, source := range kb.Spec.Sources {
			fmt.Printf("  - %s (branch: %s, type: %s)\n", source.URL, source.Branch, source.SourceType)
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

	err = progress.WithSpinner(fmt.Sprintf("Deleting knowledge base %q", opts.Name), func() error {
		return aphexClient.Delete(ctx, kb)
	})
	if err != nil {
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
			Namespace: "org-example",
		},
		Spec: platformv1alpha1.KnowledgeBaseSpec{
			Name:         "Example Knowledge Base",
			Organization: "example",
			Sources: []platformv1alpha1.Source{
				{
					URL:        "https://github.com/org/repo",
					Branch:     "main",
					SourceType: "docs",
					Paths:      []string{".kiro/docs"},
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
	default:
		data, err := yaml.Marshal(template)
		if err != nil {
			return fmt.Errorf("failed to marshal template: %w", err)
		}
		fmt.Println(string(data))
	}

	return nil
}

func loadFromJSON(path string) (*platformv1alpha1.KnowledgeBase, error) {
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

func loadFromYAML(path string) (*platformv1alpha1.KnowledgeBase, error) {
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

func waitForReady(ctx context.Context, aphexClient client.Client, name, namespace string) (*platformv1alpha1.KnowledgeBase, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	timeout := 5 * time.Minute
	interval := 3 * time.Second

	var kb platformv1alpha1.KnowledgeBase

	err := progress.WithSpinner(fmt.Sprintf("Waiting for knowledge base %q to become ready", name), func() error {
		deadline := time.After(timeout)
		for {
			select {
			case <-deadline:
				return fmt.Errorf("timed out after %s", timeout)
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := aphexClient.Get(ctx, key, &kb); err != nil {
					return err
				}
				switch kb.Status.Phase {
				case "Ready":
					return nil
				case "Failed":
					return fmt.Errorf("provisioning failed: %s", kb.Status.Message)
				}
				time.Sleep(interval)
			}
		}
	})

	if err != nil {
		return nil, err
	}
	return &kb, nil
}
