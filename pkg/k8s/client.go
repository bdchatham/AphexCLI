package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes clientset with additional functionality
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Namespace string
}

// NewClient creates a new Kubernetes client using standard kubeconfig discovery
func NewClient(kubeconfigPath string) (*Client, error) {
	config, err := loadKubeconfig(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get current namespace from kubeconfig context
	namespace, err := getCurrentNamespace(kubeconfigPath)
	if err != nil {
		// Default to "default" namespace if we can't determine current namespace
		namespace = "default"
	}

	return &Client{
		Clientset: clientset,
		Config:    config,
		Namespace: namespace,
	}, nil
}

// loadKubeconfig loads kubeconfig using standard kubectl discovery rules
func loadKubeconfig(kubeconfigPath string) (*rest.Config, error) {
	// Use provided path, or fall back to standard discovery
	if kubeconfigPath == "" {
		kubeconfigPath = getKubeconfigPath()
	}

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// getKubeconfigPath returns the kubeconfig path using kubectl discovery rules
func getKubeconfigPath() string {
	// 1. Check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// 2. Default to ~/.kube/config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".kube", "config")
}

// getCurrentNamespace extracts the current namespace from kubeconfig context
func getCurrentNamespace(kubeconfigPath string) (string, error) {
	if kubeconfigPath == "" {
		kubeconfigPath = getKubeconfigPath()
	}

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", err
	}

	if config.CurrentContext == "" {
		return "", fmt.Errorf("no current context set in kubeconfig")
	}

	context, exists := config.Contexts[config.CurrentContext]
	if !exists {
		return "", fmt.Errorf("current context %q not found in kubeconfig", config.CurrentContext)
	}

	if context.Namespace == "" {
		return "default", nil
	}

	return context.Namespace, nil
}
