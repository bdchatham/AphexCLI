package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	"gopkg.in/yaml.v2"
)

type AgentInfo struct {
	Name          string `json:"name" yaml:"name"`
	Namespace     string `json:"namespace" yaml:"namespace"`
	Model         string `json:"model" yaml:"model"`
	Phase         string `json:"phase" yaml:"phase"`
	Orchestrator  bool   `json:"orchestrator" yaml:"orchestrator"`
}

func formatTable(agents []platformv1alpha1.Agent, quiet bool) error {
	if quiet {
		for _, agent := range agents {
			fmt.Println(agent.Name)
		}
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tNAMESPACE\tMODEL\tPHASE\tORCHESTRATOR")

	for _, agent := range agents {
		phase := agent.Status.Phase
		if phase == "" {
			phase = "Pending"
		}

		orchestrator := "No"
		if agent.Status.Orchestrator != nil && agent.Status.Orchestrator.Deployed {
			orchestrator = "Yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			agent.Name,
			agent.Namespace,
			agent.Spec.Model.Name,
			phase,
			orchestrator,
		)
	}

	return w.Flush()
}

func formatJSON(agents []platformv1alpha1.Agent) error {
	var infos []AgentInfo
	for _, agent := range agents {
		phase := agent.Status.Phase
		if phase == "" {
			phase = "Pending"
		}

		orchestrator := false
		if agent.Status.Orchestrator != nil && agent.Status.Orchestrator.Deployed {
			orchestrator = true
		}

		infos = append(infos, AgentInfo{
			Name:         agent.Name,
			Namespace:    agent.Namespace,
			Model:        agent.Spec.Model.Name,
			Phase:        phase,
			Orchestrator: orchestrator,
		})
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(infos)
}

func formatYAML(agents []platformv1alpha1.Agent) error {
	var infos []AgentInfo
	for _, agent := range agents {
		phase := agent.Status.Phase
		if phase == "" {
			phase = "Pending"
		}

		orchestrator := false
		if agent.Status.Orchestrator != nil && agent.Status.Orchestrator.Deployed {
			orchestrator = true
		}

		infos = append(infos, AgentInfo{
			Name:         agent.Name,
			Namespace:    agent.Namespace,
			Model:        agent.Spec.Model.Name,
			Phase:        phase,
			Orchestrator: orchestrator,
		})
	}

	data, err := yaml.Marshal(infos)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
