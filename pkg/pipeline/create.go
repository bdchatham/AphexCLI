package pipeline

import (
	"context"
	"fmt"
	"os"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
)

var (
	// Tekton Pipeline GroupVersionResource
	pipelineGVR = schema.GroupVersionResource{
		Group:    "tekton.dev",
		Version:  "v1",
		Resource: "pipelines",
	}
)

// CreateOptions holds options for pipeline creation
type CreateOptions struct {
	Name        string
	FilePath    string
	AphexOrg    string
	RepoOrg     string
	RepoName    string
	TenantName  string
	IngressHost string
	Verbose     bool
}

// Create creates a Tekton Pipeline from a YAML file
func Create(ctx context.Context, client *k8s.Client, opts CreateOptions) error {
	// Validate file exists
	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("pipeline file not found: %s", opts.FilePath)
	}

	// Read pipeline definition from file
	pipelineData, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline file %q: %w", opts.FilePath, err)
	}

	// Parse YAML into unstructured object
	var pipeline unstructured.Unstructured
	if err := yaml.Unmarshal(pipelineData, &pipeline); err != nil {
		return fmt.Errorf("invalid YAML in pipeline file: %w", err)
	}

	// Validate it's a Tekton Pipeline
	if pipeline.GetKind() != "Pipeline" || (pipeline.GetAPIVersion() != "tekton.dev/v1beta1" && pipeline.GetAPIVersion() != "tekton.dev/v1") {
		return fmt.Errorf("file must contain a Tekton Pipeline resource (kind: Pipeline, apiVersion: tekton.dev/v1beta1 or tekton.dev/v1)")
	}

	// Set name if provided via command line
	if opts.Name != "" {
		pipeline.SetName(opts.Name)
	}

	// Ensure pipeline has a name
	if pipeline.GetName() == "" {
		return fmt.Errorf("pipeline name is required (specify via command line argument or metadata.name in YAML)")
	}

	// Create dynamic client for Tekton resources
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Set namespace to pipeline name (since names are globally unique)
	namespace := pipeline.GetName()
	pipeline.SetNamespace(namespace)

	// Create namespace if it doesn't exist
	if err := ensureNamespaceExists(ctx, dynamicClient, namespace); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("Creating pipeline %q in namespace %q\n", pipeline.GetName(), namespace)
	}

	// Create the pipeline with progress indicator
	err = progress.WithSpinner(fmt.Sprintf("Creating pipeline %q", pipeline.GetName()), func() error {
		_, err := dynamicClient.Resource(pipelineGVR).Namespace(namespace).Create(ctx, &pipeline, metav1.CreateOptions{})
		return err
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q created successfully in namespace %q\n", pipeline.GetName(), namespace)

	// Create RepoBinding
	if err := createRepoBinding(ctx, dynamicClient, opts, namespace); err != nil {
		return fmt.Errorf("failed to create RepoBinding: %w", err)
	}

	fmt.Printf("RepoBinding created successfully for pipeline %q\n", pipeline.GetName())
	return nil
}

// ensureNamespaceExists creates a namespace if it doesn't exist
func ensureNamespaceExists(ctx context.Context, dynamicClient dynamic.Interface, namespace string) error {
	nsGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	// Check if namespace exists
	_, err := dynamicClient.Resource(nsGVR).Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		// Namespace exists
		return nil
	}

	// Create namespace
	ns := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": namespace,
			},
		},
	}

	_, err = dynamicClient.Resource(nsGVR).Create(ctx, ns, metav1.CreateOptions{})
	return err
}

// createRepoBinding creates a RepoBinding for the pipeline
func createRepoBinding(ctx context.Context, dynamicClient dynamic.Interface, opts CreateOptions, pipelineNamespace string) error {
	repoBindingGVR := schema.GroupVersionResource{
		Group:    "arbiter.io",
		Version:  "v1alpha1",
		Resource: "repobindings",
	}

	repoBinding := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "arbiter.io/v1alpha1",
			"kind":       "RepoBinding",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-binding", opts.Name),
				"namespace": "platform-system",
			},
			"spec": map[string]interface{}{
				"aphexOrg":     opts.AphexOrg,
				"repoOrg":      opts.RepoOrg,
				"repoName":     opts.RepoName,
				"tenantName":   opts.TenantName,
				"pipelineName": opts.Name,
				"ingressHost":  opts.IngressHost,
			},
		},
	}

	_, err := dynamicClient.Resource(repoBindingGVR).Namespace("platform-system").Create(ctx, repoBinding, metav1.CreateOptions{})
	return err
}
