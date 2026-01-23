package secrets

import (
	"context"
	"fmt"
	"strings"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	"github.com/bdchatham/AphexCLI/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const orgSecretsName = "org-secrets"

type SetOptions struct {
	OrgName string
	Secrets map[string]string
}

type ListOptions struct {
	OrgName string
}

type DeleteOptions struct {
	OrgName string
	Keys    []string
}

func Set(ctx context.Context, k8sClient *k8s.Client, opts SetOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	namespace := fmt.Sprintf("org-%s", opts.OrgName)
	log.Debugf("Setting %d secrets in %s", len(opts.Secrets), namespace)

	existing := &corev1.Secret{}
	err = aphexClient.Get(ctx, client.ObjectKey{Name: orgSecretsName, Namespace: namespace}, existing)

	if errors.IsNotFound(err) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      orgSecretsName,
				Namespace: namespace,
				Labels: map[string]string{
					"platform.aphex/organization": opts.OrgName,
					"platform.aphex/managed-by":   "aphex-cli",
				},
			},
			StringData: opts.Secrets,
		}
		if err := aphexClient.Create(ctx, secret); err != nil {
			return k8s.FormatError(err)
		}
	} else if err != nil {
		return k8s.FormatError(err)
	} else {
		if existing.Data == nil {
			existing.Data = make(map[string][]byte)
		}
		for k, v := range opts.Secrets {
			existing.Data[k] = []byte(v)
		}
		if err := aphexClient.Update(ctx, existing); err != nil {
			return k8s.FormatError(err)
		}
	}

	for k := range opts.Secrets {
		fmt.Printf("✓ Set secret: %s\n", k)
	}
	return nil
}

func List(ctx context.Context, k8sClient *k8s.Client, opts ListOptions) error {
	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	namespace := fmt.Sprintf("org-%s", opts.OrgName)

	secret := &corev1.Secret{}
	err = aphexClient.Get(ctx, client.ObjectKey{Name: orgSecretsName, Namespace: namespace}, secret)

	if errors.IsNotFound(err) {
		fmt.Println("No secrets found")
		return nil
	} else if err != nil {
		return k8s.FormatError(err)
	}

	if len(secret.Data) == 0 {
		fmt.Println("No secrets found")
		return nil
	}

	fmt.Printf("Secrets in organization %q:\n", opts.OrgName)
	for k := range secret.Data {
		fmt.Printf("  - %s\n", k)
	}
	return nil
}

func Delete(ctx context.Context, k8sClient *k8s.Client, opts DeleteOptions) error {
	log := logger.GetLogger(ctx)

	aphexClient, err := k8s.NewAphexClient(k8sClient.Config)
	if err != nil {
		return fmt.Errorf("failed to create typed client: %w", err)
	}

	namespace := fmt.Sprintf("org-%s", opts.OrgName)
	log.Debugf("Deleting %d secrets from %s", len(opts.Keys), namespace)

	secret := &corev1.Secret{}
	err = aphexClient.Get(ctx, client.ObjectKey{Name: orgSecretsName, Namespace: namespace}, secret)

	if errors.IsNotFound(err) {
		return fmt.Errorf("no secrets found for organization %q", opts.OrgName)
	} else if err != nil {
		return k8s.FormatError(err)
	}

	for _, key := range opts.Keys {
		delete(secret.Data, key)
		fmt.Printf("✓ Deleted secret: %s\n", key)
	}

	if err := aphexClient.Update(ctx, secret); err != nil {
		return k8s.FormatError(err)
	}

	return nil
}

func ParseSecrets(args []string) (map[string]string, error) {
	secrets := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid secret format %q, expected key=value", arg)
		}
		secrets[parts[0]] = parts[1]
	}
	return secrets, nil
}
