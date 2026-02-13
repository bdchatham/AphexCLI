package auth

import (
	"context"
	"fmt"
	"time"

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
func CheckPermission(ctx context.Context, client *k8s.Client, resource, verb, namespace, group string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      verb,
				Group:     group,
				Version:   "v1alpha1",
				Resource:  resource,
			},
		},
	}

	// Perform the access review with timeout
	result, err := client.Clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(timeoutCtx, sar, metav1.CreateOptions{})
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

// CheckPipelineCreate checks permissions for pipeline creation in the org namespace
func CheckPipelineCreate(ctx context.Context, client *k8s.Client, namespace string) error {
	return CheckPermission(ctx, client, "repobindings", "create", namespace, "aphex.io")
}

// CheckPipelineDelete checks permissions for pipeline deletion in the org namespace
func CheckPipelineDelete(ctx context.Context, client *k8s.Client, namespace string) error {
	return CheckPermission(ctx, client, "repobindings", "delete", namespace, "aphex.io")
}

// CheckOrganizationBootstrap checks permissions for organization bootstrapping
func CheckOrganizationBootstrap(ctx context.Context, client *k8s.Client) error {
	// Create context with timeout to prevent hanging
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create SelfSubjectAccessReview for organizations
	sar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: "platform-system",
				Verb:      "create",
				Group:     "aphex.io",
				Version:   "v1alpha1",
				Resource:  "organizations",
			},
		},
	}

	// Perform the access review with timeout
	result, err := client.Clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(timeoutCtx, sar, metav1.CreateOptions{})
	if err != nil {
		// Log warning but don't fail - let the actual operation handle the error
		fmt.Printf("Warning: failed to perform preflight authorization check: %v\n", err)
		return nil
	}

	// Check if access is allowed
	if !result.Status.Allowed {
		return &PermissionError{
			Resource:  "organizations",
			Verb:      "create",
			Namespace: "platform-system",
			Reason:    result.Status.Reason,
		}
	}

	return nil
}
