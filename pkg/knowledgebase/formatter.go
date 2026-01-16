package knowledgebase

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	platformv1alpha1 "github.com/bdchatham/ArbiterPipelineInfrastructure/platform/platform-controller/controller/api/v1alpha1"
	"gopkg.in/yaml.v2"
)

type KnowledgeBaseInfo struct {
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	RepoCount  int    `json:"repoCount" yaml:"repoCount"`
	Phase      string `json:"phase" yaml:"phase"`
}

func formatTable(kbs []platformv1alpha1.KnowledgeBase, quiet bool) error {
	if quiet {
		for _, kb := range kbs {
			fmt.Println(kb.Name)
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tNAMESPACE\tREPOSITORIES\tPHASE")

	for _, kb := range kbs {
		phase := kb.Status.Phase
		if phase == "" {
			phase = "Pending"
		}

		repoCount := len(kb.Spec.Repositories)

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			kb.Name,
			kb.Namespace,
			repoCount,
			phase,
		)
	}

	return w.Flush()
}

func formatJSON(kbs []platformv1alpha1.KnowledgeBase) error {
	var infos []KnowledgeBaseInfo
	for _, kb := range kbs {
		phase := kb.Status.Phase
		if phase == "" {
			phase = "Pending"
		}

		infos = append(infos, KnowledgeBaseInfo{
			Name:      kb.Name,
			Namespace: kb.Namespace,
			RepoCount: len(kb.Spec.Repositories),
			Phase:     phase,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(infos)
}

func formatYAML(kbs []platformv1alpha1.KnowledgeBase) error {
	var infos []KnowledgeBaseInfo
	for _, kb := range kbs {
		phase := kb.Status.Phase
		if phase == "" {
			phase = "Pending"
		}

		infos = append(infos, KnowledgeBaseInfo{
			Name:      kb.Name,
			Namespace: kb.Namespace,
			RepoCount: len(kb.Spec.Repositories),
			Phase:     phase,
		})
	}

	data, err := yaml.Marshal(infos)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
