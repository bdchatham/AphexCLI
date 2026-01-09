package interactive

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
)

// PromptPipelineCreation prompts for pipeline creation parameters using survey
func PromptPipelineCreation(defaultNamespace string) (name, namespace, filePath string, err error) {
	fmt.Println("Interactive pipeline creation")
	
	// Pipeline name prompt
	namePrompt := &survey.Input{
		Message: "Pipeline name:",
	}
	if err := survey.AskOne(namePrompt, &name, survey.WithValidator(survey.Required)); err != nil {
		return "", "", "", err
	}
	
	// Namespace prompt with default
	namespacePrompt := &survey.Input{
		Message: "Namespace:",
		Default: defaultNamespace,
	}
	if err := survey.AskOne(namespacePrompt, &namespace); err != nil {
		return "", "", "", err
	}
	
	// File path prompt with validation
	filePrompt := &survey.Input{
		Message: "Pipeline definition file path:",
	}
	fileValidator := func(val interface{}) error {
		if str, ok := val.(string); ok && str != "" {
			if _, err := os.Stat(str); os.IsNotExist(err) {
				return fmt.Errorf("file not found: %s", str)
			}
		}
		return nil
	}
	
	if err := survey.AskOne(filePrompt, &filePath, survey.WithValidator(survey.Required), survey.WithValidator(fileValidator)); err != nil {
		return "", "", "", err
	}
	
	return name, namespace, filePath, nil
}
