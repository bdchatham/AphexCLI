package pipeline

import (
	"context"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/output"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListOptions struct {
	OutputFormat output.Format
	Quiet        bool
}

func List(ctx context.Context, k8sClient *k8s.Client, opts ListOptions) error {
	log := logger.GetLogger(ctx)
	
	aphexClient, err := k8s.NewTypedClient(k8sClient.Config)
	if err != nil {
		return err
	}

	namespaceList := &corev1.NamespaceList{}
	if err := aphexClient.List(ctx, namespaceList); err != nil {
		return k8s.FormatError(err)
	}

	log.Debugf("Listing pipelines across all namespaces")

	var allPipelines []tektonv1.Pipeline
	
	for _, ns := range namespaceList.Items {
		pipelineList := &tektonv1.PipelineList{}
		if err := aphexClient.List(ctx, pipelineList, client.InNamespace(ns.Name)); err != nil {
			log.Debugf("Skipping namespace %q: %v", ns.Name, err)
			continue
		}

		allPipelines = append(allPipelines, pipelineList.Items...)
	}

	outputOpts := output.Options{
		Format: opts.OutputFormat,
		Quiet:  opts.Quiet,
	}
	
	return output.FormatPipelines(allPipelines, outputOpts)
}
