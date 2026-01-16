package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"os"

	platformv1alpha1 "github.com/bdchatham/ArbiterPipelineInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CreateOptions struct {
	Name        string
	FilePath    string
	AphexOrg    string
	RepoOrg     string
	RepoName    string
	TenantName  string
	IngressHost string
}

func Create(ctx context.Context, k8sClient *k8s.Client, opts CreateOptions) error {
	log := logger.GetLogger(ctx)
	
	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("pipeline file not found: %s", opts.FilePath)
	}

	pipelineData, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline file %q: %w", opts.FilePath, err)
	}

	pipeline := &tektonv1.Pipeline{}
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(pipelineData), 4096)
	if err := decoder.Decode(pipeline); err != nil {
		return fmt.Errorf("invalid YAML in pipeline file: %w", err)
	}

	if opts.Name != "" {
		pipeline.Name = opts.Name
	}

	if pipeline.Name == "" {
		return fmt.Errorf("pipeline name is required (specify via command line argument or metadata.name in YAML)")
	}

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	namespace := pipeline.Name
	pipeline.Namespace = namespace

	if err := ensureNamespaceExists(ctx, aphexClient, namespace); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	log.Debugf("Creating pipeline %q in namespace %q", pipeline.Name, namespace)

	err = progress.WithSpinner(fmt.Sprintf("Creating pipeline %q", pipeline.Name), func() error {
		return aphexClient.Create(ctx, pipeline)
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q created successfully in namespace %q\n", pipeline.Name, namespace)

	log.Debugf("Verifying pipeline creation...")
	
	createdPipeline := &tektonv1.Pipeline{}
	if err := aphexClient.Get(ctx, client.ObjectKey{Name: pipeline.Name, Namespace: namespace}, createdPipeline); err != nil {
		return fmt.Errorf("pipeline was created but verification failed: %w", err)
	}
	
	log.Debugf("Pipeline verified: %s/%s", createdPipeline.Namespace, createdPipeline.Name)

	if err := createRepoBinding(ctx, aphexClient, opts, namespace, pipelineData); err != nil {
		return fmt.Errorf("failed to create RepoBinding: %w", err)
	}

	fmt.Printf("RepoBinding created successfully for pipeline %q\n", pipeline.Name)
	return nil
}

func ensureNamespaceExists(ctx context.Context, aphexClient client.Client, namespace string) error {
	ns := &corev1.Namespace{}
	err := aphexClient.Get(ctx, client.ObjectKey{Name: namespace}, ns)
	if err == nil {
		return nil
	}

	if !errors.IsNotFound(err) {
		return err
	}

	ns = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	return aphexClient.Create(ctx, ns)
}

func createRepoBinding(ctx context.Context, aphexClient client.Client, opts CreateOptions, pipelineNamespace string, pipelineYAML []byte) error {
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

	return aphexClient.Create(ctx, repoBinding)
}
