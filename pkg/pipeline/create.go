package pipeline

import (
	"context"
	"fmt"
	"os"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateOptions struct {
	Name         string
	FilePath     string
	Organization string
	RepoOrg      string
	RepoName     string
}

func orgNamespace(org string) string {
	return "org-" + org
}

func Create(ctx context.Context, k8sClient *k8s.Client, opts CreateOptions) error {
	pipelineYAML, err := os.ReadFile(opts.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read pipeline file %q: %w", opts.FilePath, err)
	}

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	namespace := orgNamespace(opts.Organization)

	repoBinding := &platformv1alpha1.RepoBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-binding", opts.Name),
			Namespace: namespace,
		},
		Spec: platformv1alpha1.RepoBindingSpec{
			AphexOrg:     opts.Organization,
			RepoOrg:      opts.RepoOrg,
			RepoName:     opts.RepoName,
			PipelineName: opts.Name,
			TemplateRef:  "run-pipeline-v1",
			PipelineSpec: string(pipelineYAML),
		},
	}

	err = progress.WithSpinner(fmt.Sprintf("Creating pipeline %q", opts.Name), func() error {
		return aphexClient.Create(ctx, repoBinding)
	})
	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q created in organization %q\n", opts.Name, opts.Organization)
	return nil
}
