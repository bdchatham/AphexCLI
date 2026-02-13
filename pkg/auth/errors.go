package auth

import (
	"fmt"
	"strings"
)

// Capability represents a CLI operation and its required permissions
type Capability struct {
	Command       string
	Resource      string
	Verb          string
	RequiredGroup string
}

// ResourceRef identifies a Kubernetes resource
type ResourceRef struct {
	Group     string
	Version   string
	Resource  string
	Namespace string
}

// capabilityMatrix maps CLI commands to required permissions and groups
var capabilityMatrix = []Capability{
	{
		Command:       "pipeline create",
		Resource:      "pipelines",
		Verb:          "create",
		RequiredGroup: "platform-engineering",
	},
	{
		Command:       "pipeline delete",
		Resource:      "pipelines",
		Verb:          "delete",
		RequiredGroup: "platform-engineering",
	},
	{
		Command:       "pipeline list",
		Resource:      "pipelines",
		Verb:          "list",
		RequiredGroup: "platform-users",
	},
}

// FriendlyMessage returns a user-friendly error message for permission errors
func (e *PermissionError) FriendlyMessage() string {
	requiredGroup := inferRequiredGroup(e.Resource, e.Verb)
	currentGroups := getCurrentGroups()

	var msg strings.Builder
	
	msg.WriteString(fmt.Sprintf("Permission denied: You don't have permission to %s %s in namespace %s.\n\n", e.Verb, e.Resource, e.Namespace))
	
	if requiredGroup != "" {
		msg.WriteString(fmt.Sprintf("Required group: %s\n", requiredGroup))
	}
	
	if len(currentGroups) > 0 {
		msg.WriteString(fmt.Sprintf("Your current groups: %s\n\n", strings.Join(currentGroups, ", ")))
	} else {
		msg.WriteString("Your current groups: none\n\n")
	}
	
	msg.WriteString("To request access:\n")
	msg.WriteString("1. Contact your platform administrator\n")
	msg.WriteString("2. Request membership in the required group\n")
	msg.WriteString("3. See platform access control documentation: https://docs.aphex.example.com/access-control\n")
	
	return msg.String()
}

// inferRequiredGroup determines the required group for a resource/verb combination
func inferRequiredGroup(resource, verb string) string {
	for _, capability := range capabilityMatrix {
		if capability.Resource == resource && capability.Verb == verb {
			return capability.RequiredGroup
		}
	}
	return ""
}

// getCurrentGroups extracts groups from the current user's OIDC token
func getCurrentGroups() []string {
	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Load the current kubeconfig
	// 2. Extract the OIDC token from exec plugin cache
	// 3. Parse the JWT token
	// 4. Extract the groups claim
	
	// For now, return empty slice as placeholder
	// The actual implementation would require JWT parsing
	return []string{}
}
