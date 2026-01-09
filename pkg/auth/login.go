package auth

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// LoginOptions holds options for authentication
type LoginOptions struct {
	KubeconfigPath string
	Verbose        bool
}

// Login configures OIDC authentication for the platform
func Login(ctx context.Context, opts LoginOptions) error {
	// Check if kubelogin is installed
	if err := checkKubeloginInstalled(); err != nil {
		return err
	}

	if opts.Verbose {
		fmt.Println("kubelogin exec plugin found")
	}

	// Configure kubeconfig
	if err := configureKubeconfig(opts); err != nil {
		return fmt.Errorf("failed to configure kubeconfig: %w", err)
	}

	// Trigger initial authentication
	if err := triggerAuthentication(ctx, opts); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println("Authentication successful!")
	return nil
}

// checkKubeloginInstalled verifies kubelogin exec plugin is available
func checkKubeloginInstalled() error {
	_, err := exec.LookPath("kubelogin")
	if err != nil {
		return fmt.Errorf(`kubelogin exec plugin not found. Please install it:

macOS: brew install Azure/kubelogin/kubelogin
Linux: Download from https://github.com/Azure/kubelogin/releases
Windows: Download from https://github.com/Azure/kubelogin/releases

After installation, run 'aphex auth login' again.`)
	}
	return nil
}

// configureKubeconfig sets up kubeconfig with OIDC exec plugin
func configureKubeconfig(opts LoginOptions) error {
	kubeconfigPath := opts.KubeconfigPath
	if kubeconfigPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Load or create kubeconfig
	config, err := loadOrCreateKubeconfig(kubeconfigPath)
	if err != nil {
		return err
	}

	// Configure "aphex" cluster
	config.Clusters["aphex"] = &api.Cluster{
		Server: "https://127.0.0.1:58134", // Use current kubeconfig cluster endpoint
	}

	// Configure "aphex" user with exec plugin
	config.AuthInfos["aphex"] = &api.AuthInfo{
		Exec: &api.ExecConfig{
			APIVersion: "client.authentication.k8s.io/v1beta1",
			Command:    "kubelogin",
			Args: []string{
				"get-token",
				"--login", "devicecode",
				"--server-id", "kubernetes",
				"--client-id", "kubernetes",
				"--tenant-id", "common",
				"--authority-host", "https://dex.home.local",
			},
		},
	}

	// Configure "aphex" context
	config.Contexts["aphex"] = &api.Context{
		Cluster:  "aphex",
		AuthInfo: "aphex",
	}

	// Set as current context
	config.CurrentContext = "aphex"

	if opts.Verbose {
		fmt.Printf("Configured kubeconfig at %s\n", kubeconfigPath)
	}

	// Write kubeconfig
	return clientcmd.WriteToFile(*config, kubeconfigPath)
}

// loadOrCreateKubeconfig loads existing kubeconfig or creates a new one
func loadOrCreateKubeconfig(path string) (*api.Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		// Return new config
		return api.NewConfig(), nil
	}

	// Load existing config
	return clientcmd.LoadFromFile(path)
}

// triggerAuthentication makes a test API call to trigger exec plugin authentication
func triggerAuthentication(ctx context.Context, opts LoginOptions) error {
	if opts.Verbose {
		fmt.Println("Triggering initial authentication...")
	}

	// Create client to trigger authentication
	client, err := k8s.NewClient(opts.KubeconfigPath)
	if err != nil {
		return err
	}

	// Make a test API call to trigger exec plugin
	_, err = client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return k8s.FormatError(err)
	}

	return nil
}
