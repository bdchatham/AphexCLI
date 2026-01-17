package knowledgebase

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	platformv1alpha1 "github.com/bdchatham/AphexPlatformInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	"github.com/bdchatham/AphexCLI/pkg/output"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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


type CreateOptions struct {
	Name      string
	Namespace string
	RepoURL   string
	Branch    string
	DocsPath  string
}

func Create(ctx context.Context, k8sClient *k8s.Client, opts CreateOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	if opts.RepoURL == "" {
		return promptForCreate(ctx, k8sClient, &opts)
	}

	if err := validateCreateOptions(opts); err != nil {
		return err
	}

	kb := &platformv1alpha1.KnowledgeBase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
		},
		Spec: platformv1alpha1.KnowledgeBaseSpec{
			DisplayName: opts.Name,
			Repositories: []platformv1alpha1.Repository{
				{
					URL:    opts.RepoURL,
					Branch: opts.Branch,
					Paths:  []string{opts.DocsPath},
				},
			},
		},
	}

	log.Debugf("Creating KnowledgeBase %q in namespace %q", opts.Name, opts.Namespace)

	if err := aphexClient.Create(ctx, kb); err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Knowledge base %q created successfully in namespace %q\n", opts.Name, opts.Namespace)
	fmt.Printf("Repository: %s (branch: %s)\n", opts.RepoURL, opts.Branch)
	fmt.Printf("Documentation path: %s\n", opts.DocsPath)

	return nil
}

func validateCreateOptions(opts CreateOptions) error {
	if opts.Name == "" {
		return fmt.Errorf("name is required")
	}

	if opts.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}

	if opts.RepoURL == "" {
		return fmt.Errorf("repository URL is required")
	}

	if !strings.HasPrefix(opts.RepoURL, "https://github.com/") {
		return fmt.Errorf("repository URL must start with https://github.com/")
	}

	if !strings.HasPrefix(opts.DocsPath, ".kiro/docs") {
		return fmt.Errorf("documentation path must start with .kiro/docs")
	}

	return nil
}

func promptForCreate(ctx context.Context, k8sClient *k8s.Client, opts *CreateOptions) error {
	log := logger.GetLogger(ctx)
	log.Debug("Starting interactive knowledge base creation")

	fmt.Println("Interactive knowledge base creation")

	if opts.RepoURL == "" {
		repoPrompt := &survey.Input{
			Message: "Repository URL (https://github.com/org/repo):",
		}
		if err := survey.AskOne(repoPrompt, &opts.RepoURL, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	if opts.Branch == "" {
		branchPrompt := &survey.Input{
			Message: "Git branch:",
			Default: "main",
		}
		if err := survey.AskOne(branchPrompt, &opts.Branch); err != nil {
			return err
		}
	}

	if opts.DocsPath == "" {
		docsPrompt := &survey.Input{
			Message: "Documentation path:",
			Default: ".kiro/docs",
		}
		if err := survey.AskOne(docsPrompt, &opts.DocsPath); err != nil {
			return err
		}
	}

	if err := validateCreateOptions(*opts); err != nil {
		return err
	}

	return Create(ctx, k8sClient, *opts)
}


type DeleteOptions struct {
	Name      string
	Namespace string
	Force     bool
}

func Delete(ctx context.Context, k8sClient *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	kb := &platformv1alpha1.KnowledgeBase{}
	objectKey := client.ObjectKey{
		Name:      opts.Name,
		Namespace: opts.Namespace,
	}

	if err := aphexClient.Get(ctx, objectKey, kb); err != nil {
		return k8s.FormatError(err)
	}

	if !opts.Force {
		fmt.Printf("This will delete knowledge base %q in namespace %q\n", opts.Name, opts.Namespace)
		fmt.Printf("\nTracked repositories:\n")
		for _, repo := range kb.Spec.Repositories {
			fmt.Printf("  - %s (branch: %s)\n", repo.URL, repo.Branch)
		}
		fmt.Printf("\nThis action cannot be undone. Continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	log.Debugf("Deleting KnowledgeBase %q from namespace %q", opts.Name, opts.Namespace)

	if err := aphexClient.Delete(ctx, kb); err != nil {
		return k8s.FormatError(err)
	}

	fmt.Printf("Knowledge base %q deleted successfully\n", opts.Name)
	return nil
}
