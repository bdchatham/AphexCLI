package pipeline

import (
	"context"
	"fmt"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/output"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListOptions struct {
	Organization string
	OutputFormat output.Format
	Quiet        bool
}

func List(ctx context.Context, k8sClient *k8s.Client, opts ListOptions) error {
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	repoBindings := &platformv1alpha1.RepoBindingList{}

	if opts.Organization != "" {
		namespace := orgNamespace(opts.Organization)
		if err := aphexClient.List(ctx, repoBindings, client.InNamespace(namespace)); err != nil {
			return k8s.FormatError(err)
		}
	} else {
		if err := aphexClient.List(ctx, repoBindings); err != nil {
			return k8s.FormatError(err)
		}
	}

	outputOpts := output.Options{
		Format: opts.OutputFormat,
		Quiet:  opts.Quiet,
	}

	return output.FormatRepoBindings(repoBindings.Items, outputOpts)
}
