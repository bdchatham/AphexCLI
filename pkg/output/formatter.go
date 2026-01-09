package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Format represents output format types
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// Options holds output formatting options
type Options struct {
	Format  Format
	Quiet   bool
	Verbose bool
}

// PipelineInfo represents pipeline information for output
type PipelineInfo struct {
	Namespace string `json:"namespace" yaml:"namespace"`
	Name      string `json:"name" yaml:"name"`
	Created   string `json:"created" yaml:"created"`
}

// FormatPipelines formats pipeline list output according to specified format
func FormatPipelines(pipelines []unstructured.Unstructured, opts Options) error {
	if len(pipelines) == 0 {
		if !opts.Quiet {
			fmt.Println("No pipelines found")
		}
		return nil
	}

	// Convert to PipelineInfo structs
	var pipelineInfos []PipelineInfo
	for _, pipeline := range pipelines {
		info := PipelineInfo{
			Namespace: pipeline.GetNamespace(),
			Name:      pipeline.GetName(),
			Created:   pipeline.GetCreationTimestamp().Format(time.RFC3339),
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

// formatJSON outputs pipelines as JSON
func formatJSON(pipelines []PipelineInfo) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(pipelines)
}

// formatYAML outputs pipelines as YAML
func formatYAML(pipelines []PipelineInfo) error {
	data, err := yaml.Marshal(pipelines)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

// formatTable outputs pipelines as a table
func formatTable(pipelines []PipelineInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAMESPACE\tNAME\tCREATED")

	for _, pipeline := range pipelines {
		fmt.Fprintf(w, "%s\t%s\t%s\n", pipeline.Namespace, pipeline.Name, pipeline.Created)
	}

	return w.Flush()
}

// FormatSuccess formats success messages
func FormatSuccess(message string, opts Options) {
	if !opts.Quiet {
		fmt.Println(message)
	}
}

// FormatVerbose formats verbose messages
func FormatVerbose(message string, opts Options) {
	if opts.Verbose {
		fmt.Println(message)
	}
}
