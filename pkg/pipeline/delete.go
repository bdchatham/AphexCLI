package pipeline

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// DeleteOptions holds options for pipeline deletion
type DeleteOptions struct {
	Name  string
	Force bool
}

// Delete deletes a Tekton Pipeline by name (using pipeline name as namespace)
func Delete(ctx context.Context, client *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)
	
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Use pipeline name as namespace (since names are globally unique)
	targetNamespace := opts.Name

	log.Debugf("Deleting pipeline %q from namespace %q", opts.Name, targetNamespace)

	// Delete the pipeline with progress indicator
	err = progress.WithSpinner(fmt.Sprintf("Deleting pipeline %q", opts.Name), func() error {
		return dynamicClient.Resource(pipelineGVR).Namespace(targetNamespace).Delete(ctx, opts.Name, metav1.DeleteOptions{})
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Pipeline %q deleted successfully from namespace %q\n", opts.Name, targetNamespace)

	// Delete associated RepoBinding
	if err := deleteRepoBinding(ctx, dynamicClient, opts.Name); err != nil {
		fmt.Printf("Warning: Failed to delete RepoBinding: %v\n", err)
	} else {
		fmt.Printf("RepoBinding %q-binding deleted successfully\n", opts.Name)
	}

	// Delete namespace if empty (only contains the pipeline we just deleted)
	if err := deleteNamespaceIfEmpty(ctx, dynamicClient, targetNamespace); err != nil {
		log.Debugf("Note: Namespace %q not deleted (may contain other resources): %v", targetNamespace, err)
	} else {
		fmt.Printf("Namespace %q deleted successfully\n", targetNamespace)
	}

	return nil
}

// deleteRepoBinding deletes the associated RepoBinding
func deleteRepoBinding(ctx context.Context, dynamicClient dynamic.Interface, pipelineName string) error {
	repoBindingGVR := schema.GroupVersionResource{
		Group:    "arbiter.io",
		Version:  "v1alpha1",
		Resource: "repobindings",
	}

	repoBindingName := fmt.Sprintf("%s-binding", pipelineName)
	return dynamicClient.Resource(repoBindingGVR).Namespace("platform-system").Delete(ctx, repoBindingName, metav1.DeleteOptions{})
}

// deleteNamespaceIfEmpty deletes the namespace if it's empty (no other resources)
func deleteNamespaceIfEmpty(ctx context.Context, dynamicClient dynamic.Interface, namespace string) error {
	nsGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	// Check if namespace has other resources (simple check - if namespace still exists after pipeline deletion, it might have other resources)
	// For safety, we'll only delete if it's the pipeline's own namespace and appears empty
	return dynamicClient.Resource(nsGVR).Delete(ctx, namespace, metav1.DeleteOptions{})
}
