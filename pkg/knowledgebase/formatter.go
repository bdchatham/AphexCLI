package knowledgebase

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"gopkg.in/yaml.v2"
)

type KnowledgeBaseInfo struct {
	Name         string `json:"name" yaml:"name"`
	Namespace    string `json:"namespace" yaml:"namespace"`
	Organization string `json:"organization" yaml:"organization"`
	SourceCount  int    `json:"sourceCount" yaml:"sourceCount"`
	Phase        string `json:"phase" yaml:"phase"`
}

func formatTable(kbs []platformv1alpha1.KnowledgeBase, quiet bool) error {
	if quiet {
		for _, kb := range kbs {
			fmt.Println(kb.Name)
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tNAMESPACE\tORGANIZATION\tSOURCES\tPHASE")

	for _, kb := range kbs {
		phase := kb.Status.Phase
		if phase == "" {
			phase = "Pending"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
			kb.Name, kb.Namespace, kb.Spec.Organization,
			len(kb.Spec.Sources), phase)
	}

	return w.Flush()
}

func formatJSON(kbs []platformv1alpha1.KnowledgeBase) error {
	infos := toInfos(kbs)
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(infos)
}

func formatYAML(kbs []platformv1alpha1.KnowledgeBase) error {
	infos := toInfos(kbs)
	data, err := yaml.Marshal(infos)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func toInfos(kbs []platformv1alpha1.KnowledgeBase) []KnowledgeBaseInfo {
	var infos []KnowledgeBaseInfo
	for _, kb := range kbs {
		phase := kb.Status.Phase
		if phase == "" {
			phase = "Pending"
		}
		infos = append(infos, KnowledgeBaseInfo{
			Name:         kb.Name,
			Namespace:    kb.Namespace,
			Organization: kb.Spec.Organization,
			SourceCount:  len(kb.Spec.Sources),
			Phase:        phase,
		})
	}
	return infos
}
