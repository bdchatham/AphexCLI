package pipeline

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/output"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ListOptions holds options for pipeline listing
type ListOptions struct {
	OutputFormat output.Format
	Quiet        bool
	Verbose      bool
}

// List lists all Tekton Pipelines across all accessible namespaces
func List(ctx context.Context, client *k8s.Client, opts ListOptions) error {
	// List all namespaces user has access to
	namespaceList, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return k8s.FormatError(err)
	}
	
	var namespaces []string
	for _, ns := range namespaceList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	if opts.Verbose {
		fmt.Printf("Listing pipelines across all namespaces\n")
	}

	// Collect all pipelines
	var allPipelines []unstructured.Unstructured
	
	for _, namespace := range namespaces {
		// List pipelines in this namespace
		result := client.Clientset.Discovery().RESTClient().
			Get().
			AbsPath("/apis", pipelineGVR.Group, pipelineGVR.Version, "namespaces", namespace, pipelineGVR.Resource).
			Do(ctx)

		if err := result.Error(); err != nil {
			// Skip namespaces we don't have access to
			if opts.Verbose {
				fmt.Printf("Skipping namespace %q: %v\n", namespace, err)
			}
			continue
		}

		var pipelineList unstructured.Unstructured
		if err := result.Into(&pipelineList); err != nil {
			continue
		}

		items, found, err := unstructured.NestedSlice(pipelineList.Object, "items")
		if err != nil || !found {
			continue
		}

		for _, item := range items {
			if pipeline, ok := item.(map[string]interface{}); ok {
				var p unstructured.Unstructured
				p.Object = pipeline
				allPipelines = append(allPipelines, p)
			}
		}
	}

	// Display results using output formatter
	outputOpts := output.Options{
		Format:  opts.OutputFormat,
		Quiet:   opts.Quiet,
		Verbose: opts.Verbose,
	}
	
	return output.FormatPipelines(allPipelines, outputOpts)
}
