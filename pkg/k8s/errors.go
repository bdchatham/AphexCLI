package k8s

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
)

// FormatError converts Kubernetes API errors into human-readable messages with remediation suggestions
func FormatError(err error) error {
	if err == nil {
		return nil
	}

	// Handle Kubernetes API errors
	if statusErr, ok := err.(*errors.StatusError); ok {
		return formatStatusError(statusErr)
	}

	// Handle other common errors
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "connection refused"):
		return fmt.Errorf("cannot connect to Kubernetes cluster - check if cluster is running and kubeconfig is correct")
	case strings.Contains(errMsg, "no such host"):
		return fmt.Errorf("cannot resolve Kubernetes API server hostname - check kubeconfig server URL")
	case strings.Contains(errMsg, "timeout"):
		return fmt.Errorf("connection to Kubernetes API server timed out - check network connectivity")
	case strings.Contains(errMsg, "certificate"):
		return fmt.Errorf("TLS certificate error - run 'aphex auth login' or check kubeconfig certificates")
	case strings.Contains(errMsg, "token"):
		return fmt.Errorf("authentication token error - run 'aphex auth login' to refresh authentication")
	default:
		return err
	}
}

// formatStatusError formats Kubernetes StatusError with specific remediation suggestions
func formatStatusError(statusErr *errors.StatusError) error {
	status := statusErr.ErrStatus
	
	switch status.Code {
	case 401:
		return fmt.Errorf("authentication failed - run 'aphex login' to authenticate with the platform")
	case 403:
		return fmt.Errorf("permission denied - you may not have access to this resource or namespace")
	case 404:
		return fmt.Errorf("resource not found - check if the resource exists and you have access to the namespace")
	case 409:
		if strings.Contains(status.Message, "already exists") {
			return fmt.Errorf("resource already exists - use a different name or delete the existing resource first")
		}
		return fmt.Errorf("conflict: %s", status.Message)
	case 422:
		return fmt.Errorf("invalid resource definition: %s", status.Message)
	case 500:
		return fmt.Errorf("Kubernetes API server error - try again later or contact platform administrators")
	case 503:
		return fmt.Errorf("Kubernetes API server unavailable - try again later")
	default:
		return fmt.Errorf("Kubernetes API error (%d): %s", status.Code, status.Message)
	}
}
