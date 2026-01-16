package organization

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	// Organization GroupVersionResource
	organizationGVR = schema.GroupVersionResource{
		Group:    "arbiter.io",
		Version:  "v1alpha1",
		Resource: "organizations",
	}
	
	// Platform system namespace where organizations are managed
	platformSystemNamespace = "platform-system"
)

// BootstrapOptions holds options for organization bootstrapping
type BootstrapOptions struct {
	Name          string
	DisplayName   string
	AdminEmail    string
	WebhookSecret string
}

// ListOptions holds options for organization listing
type ListOptions struct {
	Quiet bool
}

// Bootstrap creates a new Organization resource
func Bootstrap(ctx context.Context, client *k8s.Client, opts BootstrapOptions) error {
	log := logger.GetLogger(ctx)
	
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Set display name default
	displayName := opts.DisplayName
	if displayName == "" {
		displayName = opts.Name
	}

	// Build Organization resource
	organization := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "arbiter.io/v1alpha1",
			"kind":       "Organization",
			"metadata": map[string]interface{}{
				"name":      opts.Name,
				"namespace": platformSystemNamespace,
			},
			"spec": map[string]interface{}{
				"displayName": displayName,
				"adminUsers":  []string{opts.AdminEmail},
			},
		},
	}

	// Add webhook secret if provided
	if opts.WebhookSecret != "" {
		spec := organization.Object["spec"].(map[string]interface{})
		spec["webhookSecret"] = opts.WebhookSecret
	}

	log.Debugf("Creating organization %q with admin %q", opts.Name, opts.AdminEmail)

	// Create the organization with progress indicator
	err = progress.WithSpinner(fmt.Sprintf("Bootstrapping organization %q", opts.Name), func() error {
		_, err := dynamicClient.Resource(organizationGVR).Namespace(platformSystemNamespace).Create(ctx, organization, metav1.CreateOptions{})
		return err
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Organization %q bootstrapped successfully\n", opts.Name)
	fmt.Printf("Namespace: org-%s\n", opts.Name)
	fmt.Printf("Webhook URL: https://webhooks-%s.homelab.local\n", opts.Name)
	
	return nil
}

// List lists all Organization resources
func List(ctx context.Context, client *k8s.Client, opts ListOptions) error {
	log := logger.GetLogger(ctx)
	
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	log.Debug("Listing organizations...")

	// List organizations
	orgList, err := dynamicClient.Resource(organizationGVR).Namespace(platformSystemNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return k8s.FormatError(err)
	}

	if len(orgList.Items) == 0 {
		if !opts.Quiet {
			fmt.Println("No organizations found")
		}
		return nil
	}

	// Display organizations
	if !opts.Quiet {
		fmt.Printf("%-20s %-30s %-15s %-10s\n", "NAME", "DISPLAY NAME", "NAMESPACE", "PHASE")
		fmt.Printf("%-20s %-30s %-15s %-10s\n", "----", "------------", "---------", "-----")
	}

	for _, org := range orgList.Items {
		name := org.GetName()
		
		// Extract fields safely
		displayName := extractStringField(org.Object, "spec", "displayName")
		namespace := extractStringField(org.Object, "status", "namespace")
		phase := extractStringField(org.Object, "status", "phase")
		
		if displayName == "" {
			displayName = name
		}
		if namespace == "" {
			namespace = fmt.Sprintf("org-%s", name)
		}
		if phase == "" {
			phase = "Pending"
		}

		if opts.Quiet {
			fmt.Println(name)
		} else {
			fmt.Printf("%-20s %-30s %-15s %-10s\n", name, displayName, namespace, phase)
		}
	}

	return nil
}

// extractStringField safely extracts a string field from nested map structure
func extractStringField(obj map[string]interface{}, keys ...string) string {
	current := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			// Last key - extract the value
			if val, ok := current[key].(string); ok {
				return val
			}
			return ""
		}
		// Intermediate key - navigate deeper
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

// DeleteOptions holds options for organization deletion
type DeleteOptions struct {
	Name  string
	Force bool
}

// Delete deletes an organization and all its resources
func Delete(ctx context.Context, client *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)
	
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	orgGVR := schema.GroupVersionResource{
		Group:    "arbiter.io",
		Version:  "v1alpha1",
		Resource: "organizations",
	}

	// Check if organization exists
	_, err = dynamicClient.Resource(orgGVR).Namespace(platformSystemNamespace).Get(ctx, opts.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("organization %q not found", opts.Name)
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	// Confirmation prompt unless --force is used
	if !opts.Force {
		fmt.Printf("This will delete organization %q and ALL associated resources including:\n", opts.Name)
		fmt.Printf("  - Organization namespace (org-%s)\n", opts.Name)
		fmt.Printf("  - All pipeline namespaces for this organization\n")
		fmt.Printf("  - All secrets, webhooks, and configurations\n")
		fmt.Printf("  - All RepoBindings for this organization\n")
		fmt.Printf("\nThis action cannot be undone. Continue? (y/N): ")
		
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	log.Debugf("Deleting organization %q...", opts.Name)

	// Delete the organization (cascading delete via finalizers and owner references)
	err = progress.WithSpinner(fmt.Sprintf("Deleting organization %q", opts.Name), func() error {
		return dynamicClient.Resource(orgGVR).Namespace(platformSystemNamespace).Delete(ctx, opts.Name, metav1.DeleteOptions{})
	})

	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	fmt.Printf("Organization %q deleted successfully\n", opts.Name)
	return nil
}
