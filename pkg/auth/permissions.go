package auth

import (
	"context"
	"fmt"

	"github.com/bdchatham/AphexCLI/pkg/k8s"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PermissionError represents an authorization error with context
type PermissionError struct {
	Resource  string
	Verb      string
	Namespace string
	Reason    string
}

func (e *PermissionError) Error() string {
	// Use friendly message for better user experience
	return e.FriendlyMessage()
}

// CheckPermission performs a SelfSubjectAccessReview to verify user permissions
func CheckPermission(ctx context.Context, client *k8s.Client, resource, verb, namespace string) error {
	// Create SelfSubjectAccessReview
	sar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     "tekton.dev",
				Version:   "v1",
				Resource:  resource,
			},
		},
	}

	// Perform the access review
	result, err := client.Clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
	if err != nil {
		// Log warning but don't fail - let the actual operation handle the error
		fmt.Printf("Warning: failed to perform preflight authorization check: %v\n", err)
		return nil
	}

	// Check if access is allowed
	if !result.Status.Allowed {
		return &PermissionError{
			Resource:  resource,
			Verb:      verb,
			Namespace: namespace,
			Reason:    result.Status.Reason,
		}
	}

	return nil
}

// CheckPipelineCreate checks permissions for pipeline creation
func CheckPipelineCreate(ctx context.Context, client *k8s.Client, namespace string) error {
	return CheckPermission(ctx, client, "pipelines", "create", namespace)
}

// CheckPipelineDelete checks permissions for pipeline deletion
func CheckPipelineDelete(ctx context.Context, client *k8s.Client, namespace string) error {
	return CheckPermission(ctx, client, "pipelines", "delete", namespace)
}
