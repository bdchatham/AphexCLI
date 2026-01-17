package pipeline

import (
	"context"
	"fmt"

	platformv1alpha1 "github.com/bdchatham/AphexPlatformInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeleteOptions struct {
	Name  string
	Force bool
}

func Delete(ctx context.Context, k8sClient *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)
	
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	targetNamespace := opts.Name

	log.Debugf("Deleting pipeline %q from namespace %q", opts.Name, targetNamespace)

	err = progress.WithSpinner(fmt.Sprintf("Deleting pipeline %q", opts.Name), func() error {
		pipeline := &tektonv1.Pipeline{}
		pipeline.Name = opts.Name
		pipeline.Namespace = targetNamespace
		return aphexClient.Delete(ctx, pipeline)
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q deleted successfully from namespace %q\n", opts.Name, targetNamespace)

	if err := deleteRepoBinding(ctx, aphexClient, opts.Name); err != nil {
		fmt.Printf("Warning: Failed to delete RepoBinding: %v\n", err)
	} else {
		fmt.Printf("RepoBinding %q-binding deleted successfully\n", opts.Name)
	}

	if err := deleteNamespaceIfEmpty(ctx, aphexClient, targetNamespace); err != nil {
		log.Debugf("Note: Namespace %q not deleted (may contain other resources): %v", targetNamespace, err)
	} else {
		fmt.Printf("Namespace %q deleted successfully\n", targetNamespace)
	}

	return nil
}

func deleteRepoBinding(ctx context.Context, aphexClient client.Client, pipelineName string) error {
	repoBinding := &platformv1alpha1.RepoBinding{}
	repoBinding.Name = fmt.Sprintf("%s-binding", pipelineName)
	repoBinding.Namespace = "platform-system"
	
	return aphexClient.Delete(ctx, repoBinding)
}

func deleteNamespaceIfEmpty(ctx context.Context, aphexClient client.Client, namespace string) error {
	ns := &corev1.Namespace{}
	ns.Name = namespace
	return aphexClient.Delete(ctx, ns)
}
