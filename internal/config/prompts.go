package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type PromptConfig struct {
	SystemMessage         string `yaml:"system_message"`
	CommitMessageTemplate string `yaml:"commit_message_template"`
	PRTitleTemplate       string `yaml:"pr_title_template"`
}

var promptsCfg *PromptConfig

// InitPromptConfig initializes the prompt configuration
func InitPromptConfig() {
	if promptsCfg != nil {
		return
	}

	promptsFile := filepath.Join(getConfigDir(), ".lazycommit.prompts.yaml")

	// Check if prompts file exists
	if _, err := os.Stat(promptsFile); os.IsNotExist(err) {
		// Create default prompts file
		defaultConfig := getDefaultPromptConfig()
		if err := savePromptConfig(promptsFile, defaultConfig); err != nil {
			fmt.Printf("Error creating default prompts file: %v\n", err)
			fmt.Printf("Using default prompts\n")
		} else {
			fmt.Printf("Created default prompts config at %s\n", promptsFile)
		}
		promptsCfg = defaultConfig
		return
	}

	// Load existing prompts file
	data, err := os.ReadFile(promptsFile)
	if err != nil {
		fmt.Printf("Error reading prompts file: %v\n", err)
		fmt.Printf("Using default prompts\n")
		promptsCfg = getDefaultPromptConfig()
		return
	}

	var config PromptConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("Error parsing prompts file: %v\n", err)
		fmt.Printf("Using default prompts\n")
		promptsCfg = getDefaultPromptConfig()
		return
	}

	promptsCfg = &config
}

// getDefaultPromptConfig returns the default prompt configuration
func getDefaultPromptConfig() *PromptConfig {
	return &PromptConfig{
		SystemMessage:         "You are a helpful assistant that generates git commit messages, and pull request titles.",
		CommitMessageTemplate: "Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s",
		PRTitleTemplate:       "Based on the following git diff, generate 10 pull request title suggestions. Each title should be on a new line, without any numbering or bullet points:\n\n%s",
	}
}

// savePromptConfig saves the prompt configuration to a file
func savePromptConfig(filename string, config *PromptConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling prompt config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("error writing prompt config file: %w", err)
	}

	return nil
}

// GetPromptConfig returns the current prompt configuration
func GetPromptConfig() *PromptConfig {
	if promptsCfg == nil {
		InitPromptConfig()
	}
	return promptsCfg
}

// GetSystemMessageFromConfig returns the system message from configuration
func GetSystemMessageFromConfig() string {
	config := GetPromptConfig()
	if config.SystemMessage != "" {
		return config.SystemMessage
	}
	// Fallback to hardcoded default
	return "You are a helpful assistant that generates git commit messages."
}

// GetCommitMessagePromptFromConfig returns the commit message prompt from configuration
func GetCommitMessagePromptFromConfig(diff string) string {
	config := GetPromptConfig()
	if config.CommitMessageTemplate != "" {
		return fmt.Sprintf(config.CommitMessageTemplate, diff)
	}
	// Fallback to hardcoded default
	return fmt.Sprintf("Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s", diff)
}

// GetPRTitlePromptFromConfig returns the pull request title prompt from configuration
func GetPRTitlePromptFromConfig(diff string) string {
	config := GetPromptConfig()
	if config.PRTitleTemplate != "" {
		return fmt.Sprintf(config.PRTitleTemplate, diff)
	}
	// Fallback to hardcoded default
	return fmt.Sprintf("Based on the following git diff, generate 10 pull request title suggestions. Each title should be on a new line, without any numbering or bullet points:\n\n%s", diff)
}
