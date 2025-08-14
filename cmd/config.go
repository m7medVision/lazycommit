package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/m7medvision/lazycommit/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration for lazycommit",
	Long:  `Configure the provider, model, and other settings for lazycommit.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInteractiveConfig()
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		provider := config.GetProvider()
		model, err := config.GetModel()
		if err != nil {
			fmt.Println("Error getting model:", err)
			os.Exit(1)
		}
		fmt.Printf("Active Provider: %s\n", provider)
		fmt.Printf("Model: %s\n", model)
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		runInteractiveConfig()
	},
}

func runInteractiveConfig() {
	currentProvider := config.GetProvider()
	currentModel, _ := config.GetModel()

	providerPrompt := &survey.Select{
		Message: "Choose a provider:",
		Options: []string{"openai", "copilot"},
		Default: currentProvider,
	}
	var selectedProvider string
	err := survey.AskOne(providerPrompt, &selectedProvider)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if selectedProvider != currentProvider {
		err := config.SetProvider(selectedProvider)
		if err != nil {
			fmt.Printf("Error setting provider: %v\n", err)
			return
		}
		fmt.Printf("Provider set to: %s\n", selectedProvider)
		currentModel = ""
	}

	if selectedProvider != "copilot" {
		apiKeyPrompt := &survey.Input{
			Message: fmt.Sprintf("Enter API Key for %s:", selectedProvider),
		}
		var apiKey string
		err := survey.AskOne(apiKeyPrompt, &apiKey)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if apiKey != "" {
			err := config.SetAPIKey(selectedProvider, apiKey)
			if err != nil {
				fmt.Printf("Error setting API key: %v\n", err)
				return
			}
			fmt.Printf("API key for %s set.\n", selectedProvider)
		}
	}

	availableModels := map[string][]string{
		"openai":  {"gpt-4", "gpt-3.5-turbo", "gpt-4o"},
		"copilot": {"gpt-4o"},
	}

	modelPrompt := &survey.Select{
		Message: "Choose a model:",
		Options: availableModels[selectedProvider],
	}

	isValidDefault := false
	for _, model := range availableModels[selectedProvider] {
		if model == currentModel {
			isValidDefault = true
			break
		}
	}
	if isValidDefault {
		modelPrompt.Default = currentModel
	}

	var selectedModel string
	err = survey.AskOne(modelPrompt, &selectedModel)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if selectedModel != currentModel {
		err := config.SetModel(selectedModel)
		if err != nil {
			fmt.Printf("Error setting model: %v\n", err)
			return
		}
		fmt.Printf("Model set to: %s\n", selectedModel)
	}
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
}
