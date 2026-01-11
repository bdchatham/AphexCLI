package organization

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/progress"
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
)

// BootstrapOptions holds options for organization bootstrapping
type BootstrapOptions struct {
	Name          string
	DisplayName   string
	AdminEmail    string
	WebhookSecret string
	Verbose       bool
}

// ListOptions holds options for organization listing
type ListOptions struct {
	Quiet   bool
	Verbose bool
}

// Bootstrap creates a new Organization resource
func Bootstrap(ctx context.Context, client *k8s.Client, opts BootstrapOptions) error {
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
				"namespace": "platform-system",
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

	if opts.Verbose {
		fmt.Printf("Creating organization %q with admin %q\n", opts.Name, opts.AdminEmail)
	}

	// Create the organization with progress indicator
	err = progress.WithSpinner(fmt.Sprintf("Bootstrapping organization %q", opts.Name), func() error {
		_, err := dynamicClient.Resource(organizationGVR).Namespace("platform-system").Create(ctx, organization, metav1.CreateOptions{})
		return err
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Organization %q bootstrapped successfully\n", opts.Name)
	fmt.Printf("Namespace: org-%s\n", opts.Name)
	fmt.Printf("Webhook URL: https://webhooks-%s.homelab.local\n", opts.Name)
	
	// Update Cloudflared credentials if provided
	if err := updateCloudflaredCredentials(ctx, dynamicClient, opts.Name); err != nil {
		fmt.Printf("Warning: Failed to update Cloudflared credentials: %v\n", err)
		fmt.Printf("To enable webhooks, manually update secret: cloudflared-credentials-%s\n", opts.Name)
	}
	
	return nil
}

// updateCloudflaredCredentials updates the Cloudflared credentials secret if env var is set
func updateCloudflaredCredentials(ctx context.Context, dynamicClient dynamic.Interface, orgName string) error {
	apiToken := os.Getenv("CLOUDFLARE_TUNNEL_CREDENTIALS")
	if apiToken == "" {
		return fmt.Errorf("CLOUDFLARE_TUNNEL_CREDENTIALS environment variable not set")
	}

	credentialsJSON := fmt.Sprintf(`{"api_token": "%s"}`, apiToken)

	secretGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}

	secretName := fmt.Sprintf("cloudflared-credentials-%s", orgName)
	namespace := fmt.Sprintf("org-%s", orgName)

	secret := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      secretName,
				"namespace": namespace,
			},
			"data": map[string]interface{}{
				"credentials.json": base64.StdEncoding.EncodeToString([]byte(credentialsJSON)),
			},
		},
	}

	_, err := dynamicClient.Resource(secretGVR).Namespace(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

// List lists all Organization resources
func List(ctx context.Context, client *k8s.Client, opts ListOptions) error {
	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	if opts.Verbose {
		fmt.Println("Listing organizations...")
	}

	// List organizations
	orgList, err := dynamicClient.Resource(organizationGVR).Namespace("platform-system").List(ctx, metav1.ListOptions{})
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
