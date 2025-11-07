package cmd

import (
	"fmt"
	"net/url"
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
		endpoint, err := config.GetEndpoint()
		if err != nil {
			fmt.Println("Error getting endpoint:", err)
			os.Exit(1)
		}
		language := config.GetLanguage()
		fmt.Printf("Active Provider: %s\n", provider)
		fmt.Printf("Model: %s\n", model)
		fmt.Printf("Endpoint: %s\n", endpoint)
		fmt.Printf("Language: %s\n", language)
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		runInteractiveConfig()
	},
}

func validateEndpointURL(val any) error {
	endpoint, ok := val.(string)
	if !ok {
		return fmt.Errorf("endpoint must be a string")
	}

	// Empty string is valid (uses default)
	if endpoint == "" {
		return nil
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("endpoint must use http or https protocol")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("endpoint must have a valid host")
	}

	return nil
}

func runInteractiveConfig() {
	currentProvider := config.GetProvider()
	currentModel, _ := config.GetModel()

	providerPrompt := &survey.Select{
		Message: "Choose a provider:",
		Options: []string{"openai", "copilot", "anthropic"},
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

	// Language configuration
	currentLanguage := config.GetLanguage()
	languagePrompt := &survey.Select{
		Message: "Choose a language for commit messages:",
		Options: []string{"English (en)", "Español (es)"},
		Default: func() string {
			if currentLanguage == "es" {
				return "Español (es)"
			}
			return "English (en)"
		}(),
	}
	var selectedLanguage string
	err = survey.AskOne(languagePrompt, &selectedLanguage)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Extract language code from selection
	langCode := "en"
	if selectedLanguage == "Español (es)" {
		langCode = "es"
	}

	if langCode != currentLanguage {
		err := config.SetLanguage(langCode)
		if err != nil {
			fmt.Printf("Error setting language: %v\n", err)
			return
		}
		fmt.Printf("Language set to: %s\n", langCode)
	}

	// API key configuration - skip for copilot and anthropic
	if selectedProvider != "copilot" && selectedProvider != "anthropic" {
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
	} else if selectedProvider == "anthropic" {
		fmt.Println("Anthropic provider uses Claude Code CLI - no API key needed.")
	}

	// Dynamically generate available models
	availableModels := map[string][]string{
		"openai":    {},
		"copilot":   {"openai/gpt-5-mini"},
		"anthropic": {},
	}

	modelDisplayToID := map[string]string{}

	if selectedProvider == "openai" {
		for id, m := range models.OpenAIModels {
			display := fmt.Sprintf("%s (%s)", m.Name, string(id))
			availableModels["openai"] = append(availableModels["openai"], display)
			modelDisplayToID[display] = string(id)
		}
	} else if selectedProvider == "anthropic" {
		for _, m := range models.AnthropicModels {
			display := fmt.Sprintf("%s (%s)", m.Name, m.APIModel)
			availableModels["anthropic"] = append(availableModels["anthropic"], display)
			modelDisplayToID[display] = m.APIModel
		}
	}

	modelPrompt := &survey.Select{
		Message: "Choose a model:",
		Options: availableModels[selectedProvider],
	}

	// Try to set the default to the current model if possible
	isValidDefault := false
	currentDisplay := ""
	if selectedProvider == "openai" || selectedProvider == "anthropic" {
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
	if selectedProvider == "openai" || selectedProvider == "anthropic" {
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

	// Number of suggestions configuration for anthropic
	if selectedProvider == "anthropic" {
		numSuggestionsPrompt := &survey.Input{
			Message: "Number of commit message suggestions (default: 10):",
			Default: "10",
		}
		var numSuggestions string
		err := survey.AskOne(numSuggestionsPrompt, &numSuggestions)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if numSuggestions != "" {
			err := config.SetNumSuggestions(selectedProvider, numSuggestions)
			if err != nil {
				fmt.Printf("Error setting num_suggestions: %v\n", err)
				return
			}
			fmt.Printf("Number of suggestions set to: %s\n", numSuggestions)
		}
	}

	// Get current endpoint
	currentEndpoint, _ := config.GetEndpoint()

	// Endpoint configuration prompt - skip for anthropic since it uses CLI
	if selectedProvider != "anthropic" {
		endpointPrompt := &survey.Input{
			Message: "Enter custom endpoint URL (leave empty for default):",
			Default: currentEndpoint,
		}
		var endpoint string
		err = survey.AskOne(endpointPrompt, &endpoint, survey.WithValidator(validateEndpointURL))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		// Only set endpoint if it's different from current
		if endpoint != currentEndpoint && endpoint != "" {
			err := config.SetEndpoint(selectedProvider, endpoint)
			if err != nil {
				fmt.Printf("Error setting endpoint: %v\n", err)
				return
			}
			fmt.Printf("Endpoint set to: %s\n", endpoint)
		} else if endpoint == "" {
			fmt.Println("Using default endpoint for provider")
		}
	}
}

func init() {
	configCmd.AddCommand(getCmd)
	configCmd.AddCommand(setCmd)
}
