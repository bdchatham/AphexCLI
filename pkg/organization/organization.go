package organization

import (
	"context"
	"fmt"
	"time"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
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
	AdminEmails   []string
	WebhookSecret string
	OutputFormat  string
}

type GetOptions struct {
	Name         string
	OutputFormat string
}

type ListOptions struct {
	Quiet        bool
	OutputFormat string
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

	org := &platformv1alpha1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: platformSystemNamespace,
		},
		Spec: platformv1alpha1.OrganizationSpec{
			DisplayName: displayName,
			AdminUsers:  opts.AdminEmails,
		},
	}

	if opts.WebhookSecret != "" {
		org.Spec.WebhookSecret = opts.WebhookSecret
	}

	log.Debugf("Creating organization %q with admins %v", opts.Name, opts.AdminEmails)

	err = progress.WithSpinner(fmt.Sprintf("Creating organization %q", opts.Name), func() error {
		return aphexClient.Create(ctx, org)
	})
	if err != nil {
		return k8s.FormatError(err)
	}

	readyOrg, err := waitForReady(ctx, aphexClient, opts.Name)
	if err != nil {
		fmt.Printf("Organization %q created but status polling failed: %v\n", opts.Name, err)
		fmt.Printf("Check status with: aphex organization get %s\n", opts.Name)
		return nil
	}

	return FormatOrganization(*readyOrg, opts.OutputFormat)
}

func Get(ctx context.Context, k8sClient *k8s.Client, opts GetOptions) error {
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	org := &platformv1alpha1.Organization{}
	key := client.ObjectKey{Name: opts.Name, Namespace: platformSystemNamespace}
	if err := aphexClient.Get(ctx, key, org); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("organization %q not found", opts.Name)
		}
		return k8s.FormatError(err)
	}

	return FormatOrganization(*org, opts.OutputFormat)
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

	return FormatOrganizations(orgList.Items, opts.OutputFormat, opts.Quiet)
}

func Delete(ctx context.Context, k8sClient *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	org := &platformv1alpha1.Organization{}
	if err := aphexClient.Get(ctx, client.ObjectKey{Name: opts.Name, Namespace: platformSystemNamespace}, org); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("organization %q not found", opts.Name)
		}
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if !opts.Force {
		fmt.Printf("This will delete organization %q and ALL associated resources including:\n", opts.Name)
		fmt.Printf("  - Organization namespace (%s)\n", org.Status.Namespace)
		fmt.Printf("  - All pipeline namespaces for this organization\n")
		fmt.Printf("  - All secrets, webhooks, and configurations\n")
		fmt.Printf("  - All RepoBindings for this organization\n")
		fmt.Printf("\nThis action cannot be undone. Continue? (y/N): ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil || (response != "y" && response != "Y" && response != "yes" && response != "Yes") {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	log.Debugf("Deleting organization %q...", opts.Name)

	err = progress.WithSpinner(fmt.Sprintf("Deleting organization %q", opts.Name), func() error {
		return aphexClient.Delete(ctx, org)
	})
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	fmt.Printf("Organization %q deleted successfully\n", opts.Name)
	return nil
}

func waitForReady(ctx context.Context, aphexClient client.Client, name string) (*platformv1alpha1.Organization, error) {
	key := client.ObjectKey{Name: name, Namespace: platformSystemNamespace}
	timeout := 2 * time.Minute
	interval := 2 * time.Second

	var org platformv1alpha1.Organization

	err := progress.WithSpinner(fmt.Sprintf("Waiting for organization %q to become ready", name), func() error {
		deadline := time.After(timeout)
		for {
			select {
			case <-deadline:
				return fmt.Errorf("timed out after %s", timeout)
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := aphexClient.Get(ctx, key, &org); err != nil {
					return err
				}
				switch org.Status.Phase {
				case "Ready":
					return nil
				case "Failed":
					return fmt.Errorf("provisioning failed: %s", org.Status.Message)
				}
				time.Sleep(interval)
			}
		}
	})

	if err != nil {
		return nil, err
	}
	return &org, nil
}
