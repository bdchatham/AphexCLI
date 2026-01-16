package pipeline

import (
	"context"
	"fmt"
	"os"

	platformv1alpha1 "github.com/bdchatham/ArbiterPipelineInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
}

// Create creates a Tekton Pipeline from a YAML file
func Create(ctx context.Context, client *k8s.Client, opts CreateOptions) error {
	log := logger.GetLogger(ctx)
	
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

	log.Debugf("Creating pipeline %q in namespace %q", pipeline.GetName(), namespace)

	// Create the pipeline with progress indicator
	err = progress.WithSpinner(fmt.Sprintf("Creating pipeline %q", pipeline.GetName()), func() error {
		_, err := dynamicClient.Resource(pipelineGVR).Namespace(namespace).Create(ctx, &pipeline, metav1.CreateOptions{})
		return err
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q created successfully in namespace %q\n", pipeline.GetName(), namespace)

	// Verify pipeline was created
	log.Debugf("Verifying pipeline creation...")
	
	createdPipeline, err := dynamicClient.Resource(pipelineGVR).Namespace(namespace).Get(ctx, pipeline.GetName(), metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("pipeline was created but verification failed: %w", err)
	}
	
	log.Debugf("Pipeline verified: %s/%s", createdPipeline.GetNamespace(), createdPipeline.GetName())

	// Create RepoBinding with pipeline YAML
	if err := createRepoBinding(ctx, client, opts, namespace, pipelineData); err != nil {
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

// createRepoBinding creates a RepoBinding using controller-runtime client with typed structs
func createRepoBinding(ctx context.Context, k8sClient *k8s.Client, opts CreateOptions, pipelineNamespace string, pipelineYAML []byte) error {
	// Create scheme and register our types
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("failed to add types to scheme: %w", err)
	}

	// Create controller-runtime client
	runtimeClient, err := client.New(k8sClient.Config, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("failed to create runtime client: %w", err)
	}

	// Create typed RepoBinding
	repoBinding := &platformv1alpha1.RepoBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-binding", opts.Name),
			Namespace: "platform-system",
		},
		Spec: platformv1alpha1.RepoBindingSpec{
			AphexOrg:     opts.AphexOrg,
			RepoOrg:      opts.RepoOrg,
			RepoName:     opts.RepoName,
			PipelineName: opts.Name,
			TemplateRef:  "run-pipeline-v1",
			PipelineSpec: string(pipelineYAML),
		},
	}

	return runtimeClient.Create(ctx, repoBinding)
}
