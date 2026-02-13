package organization

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"gopkg.in/yaml.v2"
)

type OrganizationInfo struct {
	Name        string   `json:"name" yaml:"name"`
	DisplayName string   `json:"displayName" yaml:"displayName"`
	Namespace   string   `json:"namespace" yaml:"namespace"`
	Phase       string   `json:"phase" yaml:"phase"`
	AdminUsers  []string `json:"adminUsers" yaml:"adminUsers"`
	WebhookURL  string   `json:"webhookURL,omitempty" yaml:"webhookURL,omitempty"`
	Message     string   `json:"message,omitempty" yaml:"message,omitempty"`
}

func toInfo(org platformv1alpha1.Organization) OrganizationInfo {
	phase := string(org.Status.Phase)
	if phase == "" {
		phase = "Pending"
	}
	namespace := org.Status.Namespace
	if namespace == "" {
		namespace = fmt.Sprintf("org-%s", org.Name)
	}
	return OrganizationInfo{
		Name:        org.Name,
		DisplayName: org.Spec.DisplayName,
		Namespace:   namespace,
		Phase:       phase,
		AdminUsers:  org.Spec.AdminUsers,
		WebhookURL:  org.Status.WebhookURL,
		Message:     org.Status.Message,
	}
}

func FormatOrganization(org platformv1alpha1.Organization, format string) error {
	return FormatOrganizations([]platformv1alpha1.Organization{org}, format, false)
}

func FormatOrganizations(orgs []platformv1alpha1.Organization, format string, quiet bool) error {
	if len(orgs) == 0 {
		if !quiet {
			fmt.Println("No organizations found")
		}
		return nil
	}

	var infos []OrganizationInfo
	for _, org := range orgs {
		infos = append(infos, toInfo(org))
	}

	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(infos)
	case "yaml":
		data, err := yaml.Marshal(infos)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	default:
		if quiet {
			for _, info := range infos {
				fmt.Println(info.Name)
			}
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tDISPLAY NAME\tNAMESPACE\tPHASE\tADMIN USERS")
		for _, info := range infos {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				info.Name, info.DisplayName, info.Namespace, info.Phase,
				strings.Join(info.AdminUsers, ", "))
		}
		return w.Flush()
	}
}
