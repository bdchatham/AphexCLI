package k8s

import (
	"fmt"

	platformv1alpha1 "github.com/bdchatham/ArbiterPipelineInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewAphexClient(config *rest.Config) (client.Client, error) {
	scheme := runtime.NewScheme()

	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add core types to scheme: %w", err)
	}

	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add platform types to scheme: %w", err)
	}

	if err := tektonv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add Tekton types to scheme: %w", err)
	}

	aphexClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create typed client: %w", err)
	}

	return aphexClient, nil
}
