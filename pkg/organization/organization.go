package organization

import (
	"context"
	"fmt"

	platformv1alpha1 "github.com/bdchatham/AphexPlatformInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/progress"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const platformSystemNamespace = "platform-system"

type BootstrapOptions struct {
	Name          string
	DisplayName   string
	AdminEmail    string
	WebhookSecret string
}

type ListOptions struct {
	Quiet bool
}

type DeleteOptions struct {
	Name  string
	Force bool
}

func Bootstrap(ctx context.Context, k8sClient *k8s.Client, opts BootstrapOptions) error {
	log := logger.GetLogger(ctx)
	
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	displayName := opts.DisplayName
	if displayName == "" {
		displayName = opts.Name
	}

	organization := &platformv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: platformSystemNamespace,
		},
		Spec: platformv1alpha1.OrganizationSpec{
			DisplayName: displayName,
			AdminUsers:  []string{opts.AdminEmail},
		},
	}

	if opts.WebhookSecret != "" {
		organization.Spec.WebhookSecret = opts.WebhookSecret
	}

	log.Debugf("Creating organization %q with admin %q", opts.Name, opts.AdminEmail)

	err = progress.WithSpinner(fmt.Sprintf("Bootstrapping organization %q", opts.Name), func() error {
		return aphexClient.Create(ctx, organization)
	})

	if err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Organization %q bootstrapped successfully\n", opts.Name)
	fmt.Printf("Namespace: org-%s\n", opts.Name)
	fmt.Printf("Webhook URL: https://webhooks-%s.homelab.local\n", opts.Name)
	
	return nil
}

func List(ctx context.Context, k8sClient *k8s.Client, opts ListOptions) error {
	log := logger.GetLogger(ctx)
	
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	log.Debug("Listing organizations...")

	orgList := &platformv1alpha1.OrganizationList{}
	if err := aphexClient.List(ctx, orgList, client.InNamespace(platformSystemNamespace)); err != nil {
		return k8s.FormatError(err)
	}

	if len(orgList.Items) == 0 {
		if !opts.Quiet {
			fmt.Println("No organizations found")
		}
		return nil
	}

	if !opts.Quiet {
		fmt.Printf("%-20s %-30s %-15s %-10s\n", "NAME", "DISPLAY NAME", "NAMESPACE", "PHASE")
		fmt.Printf("%-20s %-30s %-15s %-10s\n", "----", "------------", "---------", "-----")
	}

	for _, org := range orgList.Items {
		displayName := org.Spec.DisplayName
		if displayName == "" {
			displayName = org.Name
		}
		
		namespace := org.Status.Namespace
		if namespace == "" {
			namespace = fmt.Sprintf("org-%s", org.Name)
		}
		
		phase := org.Status.Phase
		if phase == "" {
			phase = "Pending"
		}

		if opts.Quiet {
			fmt.Println(org.Name)
		} else {
			fmt.Printf("%-20s %-30s %-15s %-10s\n", org.Name, displayName, namespace, phase)
		}
	}

	return nil
}

func Delete(ctx context.Context, k8sClient *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)
	
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	organization := &platformv1alpha1.Organization{}
	if err := aphexClient.Get(ctx, client.ObjectKey{Name: opts.Name, Namespace: platformSystemNamespace}, organization); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("organization %q not found", opts.Name)
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

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

	err = progress.WithSpinner(fmt.Sprintf("Deleting organization %q", opts.Name), func() error {
		return aphexClient.Delete(ctx, organization)
	})

	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	fmt.Printf("Organization %q deleted successfully\n", opts.Name)
	return nil
}
