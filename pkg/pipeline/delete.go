package pipeline

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/progress"
)

// DeleteOptions holds options for pipeline deletion
type DeleteOptions struct {
	Name      string
	Namespace string
	Force     bool
	Verbose   bool
}

// Delete deletes a Tekton Pipeline
func Delete(ctx context.Context, client *k8s.Client, opts DeleteOptions) error {
	namespace := opts.Namespace
	if namespace == "" {
		namespace = client.Namespace
	}

	if opts.Verbose {
		fmt.Printf("Deleting pipeline %q from namespace %q\n", opts.Name, namespace)
	}

	// Confirmation prompt unless --force is used
	if !opts.Force {
		confirmed := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Delete pipeline %q in namespace %q?", opts.Name, namespace),
			Default: false,
		}
		if err := survey.AskOne(prompt, &confirmed); err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete the pipeline with progress indicator
	err := progress.WithSpinner(fmt.Sprintf("Deleting pipeline %q", opts.Name), func() error {
		dynamicClient := client.Clientset.Discovery().RESTClient()
		result := dynamicClient.
			Delete().
			AbsPath("/apis", pipelineGVR.Group, pipelineGVR.Version, "namespaces", namespace, pipelineGVR.Resource, opts.Name).
			Do(ctx)

		return result.Error()
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q deleted successfully from namespace %q\n", opts.Name, namespace)
	return nil
}
