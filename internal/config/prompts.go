package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/m7medvision/lazycommit/internal/git"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type PromptConfig struct {
	SystemMessage         string `yaml:"system_message"`
	CommitMessageTemplate string `yaml:"commit_message_template"`
	PRTitleTemplate       string `yaml:"pr_title_template"`
	Language              string `yaml:"language,omitempty"`
}

var promptsCfg *PromptConfig

func InitPromptConfig() {
	if promptsCfg != nil {
		return
	}

	globalPromptsFile := filepath.Join(getConfigDir(), ".lazycommit.prompts.yaml")
	globalPrompts := loadPromptConfigFromFile(globalPromptsFile)

	if globalPrompts.Language == "" {
		legacyLanguage := viper.GetString("language")
		if legacyLanguage != "" {
			globalPrompts.Language = legacyLanguage
			_ = savePromptConfig(globalPromptsFile, globalPrompts)
		}
	}

	effectivePrompts := globalPrompts

	if repoRoot, err := git.GetRepoRoot(); err == nil {
		localPromptsFile := filepath.Join(repoRoot, ".lazycommit.prompts.yaml")
		if _, err := os.Stat(localPromptsFile); err == nil {
			localPrompts := loadPromptConfigFromFile(localPromptsFile)
			effectivePrompts = mergePromptConfigs(globalPrompts, localPrompts)
		}
	}

	promptsCfg = effectivePrompts
}

func loadPromptConfigFromFile(path string) *PromptConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return getDefaultPromptConfig()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return getDefaultPromptConfig()
	}

	var config PromptConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return getDefaultPromptConfig()
	}

	return &config
}

func mergePromptConfigs(global, local *PromptConfig) *PromptConfig {
	merged := *global
	if local.SystemMessage != "" {
		merged.SystemMessage = local.SystemMessage
	}
	if local.CommitMessageTemplate != "" {
		merged.CommitMessageTemplate = local.CommitMessageTemplate
	}
	if local.PRTitleTemplate != "" {
		merged.PRTitleTemplate = local.PRTitleTemplate
	}
	if local.Language != "" {
		merged.Language = local.Language
	}
	return &merged
}

func getDefaultPromptConfig() *PromptConfig {
	return &PromptConfig{
		SystemMessage:         "You are a helpful assistant that generates git commit messages, and pull request titles.",
		CommitMessageTemplate: "Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s",
		PRTitleTemplate:       "Based on the following git diff, generate 10 pull request title suggestions. Each title should be on a new line, without any numbering or bullet points:\n\n%s",
		Language:              "English",
	}
}

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

func GetPromptConfig() *PromptConfig {
	if promptsCfg == nil {
		InitPromptConfig()
	}
	return promptsCfg
}

func GetSystemMessageFromConfig() string {
	config := GetPromptConfig()
	if config.SystemMessage != "" {
		return config.SystemMessage
	}
	return "You are a helpful assistant that generates git commit messages."
}

func getLanguageInstruction(language string) string {
	if language == "" {
		return ""
	}

	return fmt.Sprintf("\n\nIMPORTANT: Generate all content in %s.", language)
}

func GetCommitMessagePromptFromConfig(diff string) string {
	config := GetPromptConfig()
	var basePrompt string
	if config.CommitMessageTemplate != "" {
		basePrompt = fmt.Sprintf(config.CommitMessageTemplate, diff)
	} else {
		basePrompt = fmt.Sprintf("Based on the following git diff, generate 10 conventional commit messages. Each message should be on a new line, without any numbering or bullet points:\n\n%s", diff)
	}

	basePrompt += getLanguageInstruction(config.Language)

	return basePrompt
}

func GetPRTitlePromptFromConfig(diff string) string {
	config := GetPromptConfig()
	var basePrompt string
	if config.PRTitleTemplate != "" {
		basePrompt = fmt.Sprintf(config.PRTitleTemplate, diff)
	} else {
		basePrompt = fmt.Sprintf("Based on the following git diff, generate 10 pull request title suggestions. Each title should be on a new line, without any numbering or bullet points:\n\n%s", diff)
	}

	basePrompt += getLanguageInstruction(config.Language)

	return basePrompt
}

func SetLanguage(language string) error {
	globalPromptsFile := filepath.Join(getConfigDir(), ".lazycommit.prompts.yaml")
	globalPrompts := loadPromptConfigFromFile(globalPromptsFile)
	globalPrompts.Language = language
	err := savePromptConfig(globalPromptsFile, globalPrompts)
	if err != nil {
		return err
	}

	if promptsCfg != nil {
		promptsCfg.Language = language
	}

	return nil
}

func GetLanguage() string {
	return GetPromptConfig().Language
}
