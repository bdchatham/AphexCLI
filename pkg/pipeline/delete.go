package pipeline

import (
	"context"
	"fmt"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DeleteOptions struct {
	Name         string
	Organization string
	Force        bool
}

func Delete(ctx context.Context, k8sClient *k8s.Client, opts DeleteOptions) error {
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	namespace := orgNamespace(opts.Organization)
	bindingName := fmt.Sprintf("%s-binding", opts.Name)

	repoBinding := &platformv1alpha1.RepoBinding{}
	if err := aphexClient.Get(ctx, client.ObjectKey{Name: bindingName, Namespace: namespace}, repoBinding); err != nil {
		return k8s.FormatError(err)
	}

	err = progress.WithSpinner(fmt.Sprintf("Deleting pipeline %q", opts.Name), func() error {
		return aphexClient.Delete(ctx, repoBinding)
	})
	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q deleted from organization %q\n", opts.Name, opts.Organization)
	return nil
}
