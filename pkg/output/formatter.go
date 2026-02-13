package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	platformv1alpha1 "github.com/bdchatham/AphexControllerRuntime/api/v1alpha1"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"gopkg.in/yaml.v2"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

type Options struct {
	Format  Format
	Quiet   bool
	Verbose bool
}

type PipelineInfo struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Created   string `json:"created" yaml:"created"`
}

func FormatPipelines(pipelines []tektonv1.Pipeline, opts Options) error {
	if len(pipelines) == 0 {
		if !opts.Quiet {
			fmt.Println("No pipelines found")
		}
		return nil
	}

	var pipelineInfos []PipelineInfo
	for _, pipeline := range pipelines {
		info := PipelineInfo{
			Namespace: pipeline.Namespace,
			Name:      pipeline.Name,
			Created:   pipeline.CreationTimestamp.Format(time.RFC3339),
		}
		pipelineInfos = append(pipelineInfos, info)
	}

	switch opts.Format {
	case FormatJSON:
		return formatJSON(pipelineInfos)
	case FormatYAML:
		return formatYAML(pipelineInfos)
	case FormatTable:
		fallthrough
	default:
		return formatTable(pipelineInfos)
	}
}

func formatJSON(pipelines []PipelineInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(pipelines)
}

func formatYAML(pipelines []PipelineInfo) error {
	data, err := yaml.Marshal(pipelines)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func formatTable(pipelines []PipelineInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAMESPACE\tNAME\tCREATED")

	for _, pipeline := range pipelines {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", pipeline.Namespace, pipeline.Name, pipeline.Created)
	}

	return w.Flush()
}

type RepoBindingInfo struct {
	Name         string `json:"name" yaml:"name"`
	Organization string `json:"organization" yaml:"organization"`
	Repo         string `json:"repo" yaml:"repo"`
	Pipeline     string `json:"pipeline" yaml:"pipeline"`
	Phase        string `json:"phase" yaml:"phase"`
}

func FormatRepoBindings(bindings []platformv1alpha1.RepoBinding, opts Options) error {
	if len(bindings) == 0 {
		if !opts.Quiet {
			fmt.Println("No pipelines found")
		}
		return nil
	}

	var infos []RepoBindingInfo
	for _, rb := range bindings {
		infos = append(infos, RepoBindingInfo{
			Name:         rb.Spec.PipelineName,
			Organization: rb.Spec.AphexOrg,
			Repo:         fmt.Sprintf("%s/%s", rb.Spec.RepoOrg, rb.Spec.RepoName),
			Pipeline:     rb.Spec.PipelineName,
			Phase:        rb.Status.Phase,
		})
	}

	switch opts.Format {
	case FormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(infos)
	case FormatYAML:
		data, err := yaml.Marshal(infos)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
		return nil
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tORGANIZATION\tREPO\tPHASE")
		for _, info := range infos {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", info.Name, info.Organization, info.Repo, info.Phase)
		}
		return w.Flush()
	}
}

func FormatSuccess(message string, opts Options) {
	if !opts.Quiet {
		fmt.Println(message)
	}
}

func FormatVerbose(message string, opts Options) {
	if opts.Verbose {
		fmt.Println(message)
	}
}
