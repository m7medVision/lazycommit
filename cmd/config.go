package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/m7medvision/lazycommit/internal/config"
	"github.com/m7medvision/lazycommit/internal/provider/models"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration for lazycommit",
	Long:  `Configure the provider, model, and other settings for lazycommit.`,
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
		
		// Show base URL if set for OpenAI provider
		if provider == "openai" {
			baseURL, err := config.GetBaseURL()
			if err == nil && baseURL != "" {
				fmt.Printf("Base URL: %s\n", baseURL)
			}
		}
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
		Options: []string{"openai", "openrouter", "copilot"},
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

	// Ask for custom base URL if using OpenAI provider
	if selectedProvider == "openai" {
		baseURLPrompt := &survey.Input{
			Message: "Enter custom API base URL (leave empty for default OpenAI endpoint):",
		}
		var baseURL string
		err := survey.AskOne(baseURLPrompt, &baseURL)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if baseURL != "" {
			err := config.SetBaseURL(selectedProvider, baseURL)
			if err != nil {
				fmt.Printf("Error setting base URL: %v\n", err)
				return
			}
			fmt.Printf("Base URL for %s set to: %s\n", selectedProvider, baseURL)
		}
	}

	// Dynamically generate available models for OpenAI
	availableModels := map[string][]string{
		"openai":     {},
		"openrouter": {},
		"copilot":    {"gpt-4o"}, // TODO: update if copilot models are dynamic
	}

	modelDisplayToID := map[string]string{}
	if selectedProvider == "openai" {
		for id, m := range models.OpenAIModels {
			display := fmt.Sprintf("%s (%s)", m.Name, string(id))
			availableModels["openai"] = append(availableModels["openai"], display)
			modelDisplayToID[display] = string(id)
		}
	} else if selectedProvider == "openrouter" {
		for id, m := range models.OpenRouterModels {
			display := fmt.Sprintf("%s (%s)", m.Name, string(id))
			availableModels["openrouter"] = append(availableModels["openrouter"], display)
			modelDisplayToID[display] = string(id)
		}
	}

	modelPrompt := &survey.Select{
		Message: "Choose a model:",
		Options: availableModels[selectedProvider],
	}

	// Try to set the default to the current model if possible
	isValidDefault := false
	currentDisplay := ""
	if selectedProvider == "openai" || selectedProvider == "openrouter" {
		for display, id := range modelDisplayToID {
			if id == currentModel || display == currentModel {
				isValidDefault = true
				currentDisplay = display
				break
			}
		}
	} else {
		for _, model := range availableModels[selectedProvider] {
			if model == currentModel {
				isValidDefault = true
				currentDisplay = model
				break
			}
		}
	}
	if isValidDefault {
		modelPrompt.Default = currentDisplay
	}

	var selectedDisplay string
	err = survey.AskOne(modelPrompt, &selectedDisplay)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	selectedModel := selectedDisplay
	if selectedProvider == "openai" || selectedProvider == "openrouter" {
		selectedModel = modelDisplayToID[selectedDisplay]
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
