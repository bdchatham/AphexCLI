package knowledgebase

import (
	"context"
	"fmt"

	platformv1alpha1 "github.com/bdchatham/ArbiterPipelineInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/output"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
