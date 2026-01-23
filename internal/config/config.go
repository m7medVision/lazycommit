package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type ProviderConfig struct {
	APIKey         string `mapstructure:"api_key"`
	Model          string `mapstructure:"model"`
	EndpointURL    string `mapstructure:"endpoint_url"`
	NumSuggestions int    `mapstructure:"num_suggestions"`
}

type Config struct {
	Providers      map[string]ProviderConfig `mapstructure:"providers"`
	ActiveProvider string                    `mapstructure:"active_provider"`
	Language       string                    `mapstructure:"language"`
}

var cfg *Config

func InitConfig() {
	viper.SetConfigName(".lazycommit")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(getConfigDir())

	viper.SetConfigFile(filepath.Join(getConfigDir(), ".lazycommit.yaml"))

	if token, err := LoadGitHubToken(); err == nil && token != "" {
		viper.SetDefault("active_provider", "copilot")
		viper.SetDefault("providers.copilot.api_key", token)
		viper.SetDefault("providers.copilot.model", "openai/gpt-5-mini")
	} else {
		viper.SetDefault("active_provider", "openai")
		viper.SetDefault("providers.openai.model", "openai/gpt-5-mini")
	}

	viper.SetDefault("providers.anthropic.model", "claude-haiku-4-5")
	viper.SetDefault("providers.anthropic.num_suggestions", 10)
	viper.SetDefault("language", "en")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			cfgDir := getConfigDir()
			_ = os.MkdirAll(cfgDir, 0o755)
			cfgPath := filepath.Join(cfgDir, ".lazycommit.yaml")
			if writeErr := viper.WriteConfigAs(cfgPath); writeErr != nil {
				fmt.Println("Error creating default config file:", writeErr)
			} else {
				fmt.Printf("Created default config at %s\n", cfgPath)
			}
			_ = viper.ReadInConfig()
		} else {
			fmt.Println("Error reading config file:", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("Error unmarshalling config:", err)
		os.Exit(1)
	}

	InitPromptConfig()
}

func GetProvider() string {
	if cfg == nil {
		InitConfig()
	}
	return cfg.ActiveProvider
}

func GetActiveProviderConfig() (*ProviderConfig, error) {
	if cfg == nil {
		InitConfig()
	}
	providerName := cfg.ActiveProvider
	providerConfig, ok := cfg.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not configured", providerName)
	}
	return &providerConfig, nil
}

func GetAPIKey() (string, error) {
	if cfg == nil {
		InitConfig()
	}

	providerConfig, err := GetActiveProviderConfig()
	if err != nil {
		return "", err
	}

	if providerConfig.APIKey == "" {
		return "", fmt.Errorf("API key for provider '%s' is not set", cfg.ActiveProvider)
	}

	apiKey := providerConfig.APIKey

	// Check if the API key is an environment variable reference
	if strings.HasPrefix(apiKey, "$") {
		envVarName := strings.TrimPrefix(apiKey, "$")
		envValue := os.Getenv(envVarName)
		if envValue == "" {
			return "", fmt.Errorf("environment variable '%s' for provider '%s' is not set or empty", envVarName, cfg.ActiveProvider)
		}
		return envValue, nil
	}

	return apiKey, nil
}

func GetModel() (string, error) {
	providerConfig, err := GetActiveProviderConfig()
	if err != nil {
		return "", err
	}
	if providerConfig.Model == "" {
		return "", fmt.Errorf("model for provider '%s' is not set", cfg.ActiveProvider)
	}
	return providerConfig.Model, nil
}

func GetEndpoint() (string, error) {
	providerConfig, err := GetActiveProviderConfig()
	if err != nil {
		return "", err
	}

	// If custom endpoint is configured, use it
	if providerConfig.EndpointURL != "" {
		return providerConfig.EndpointURL, nil
	}

	// Return default endpoints based on provider
	switch cfg.ActiveProvider {
	case "openai":
		return "https://api.openai.com/v1", nil
	case "copilot":
		return "https://api.githubcopilot.com", nil
	case "anthropic":
		return "", nil // Anthropic uses CLI, no endpoint needed
	default:
		return "", fmt.Errorf("no default endpoint available for provider '%s'", cfg.ActiveProvider)
	}
}

func SetProvider(provider string) error {
	if cfg == nil {
		InitConfig()
	}
	cfg.ActiveProvider = provider
	viper.Set("active_provider", provider)
	return viper.WriteConfig()
}

func SetModel(model string) error {
	if cfg == nil {
		InitConfig()
	}
	provider := cfg.ActiveProvider
	viper.Set(fmt.Sprintf("providers.%s.model", provider), model)
	return viper.WriteConfig()
}

func SetAPIKey(provider, apiKey string) error {
	if cfg == nil {
		InitConfig()
	}
	viper.Set(fmt.Sprintf("providers.%s.api_key", provider), apiKey)
	return viper.WriteConfig()
}

func validateEndpointURL(endpoint string) error {
	if endpoint == "" {
		return nil // Empty endpoint is valid (will use default)
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

func SetEndpoint(provider, endpoint string) error {
	if cfg == nil {
		InitConfig()
	}

	// Validate endpoint URL
	if err := validateEndpointURL(endpoint); err != nil {
		return err
	}

	viper.Set(fmt.Sprintf("providers.%s.endpoint_url", provider), endpoint)
	return viper.WriteConfig()
}

func LoadGitHubToken() (string, error) {
	tok, err := tryGetTokenFromGHCLI()
	if err == nil && tok != "" {
		return tok, nil
	}

	configDir := getConfigDir()

	filePaths := []string{
		filepath.Join(configDir, "github-copilot", "hosts.json"),
		filepath.Join(configDir, "github-copilot", "apps.json"),
	}

	for _, filePath := range filePaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var configData map[string]map[string]interface{}
		if err := json.Unmarshal(data, &configData); err != nil {
			continue
		}

		for key, value := range configData {
			if strings.Contains(key, "github.com") {
				if oauthToken, ok := value["oauth_token"].(string); ok && oauthToken != "" {
					return oauthToken, nil
				}
			}
		}
	}

	return "", fmt.Errorf("GitHub token not found via 'gh auth token'; run 'gh auth login' to authenticate the GitHub CLI")
}
func tryGetTokenFromGHCLI() (string, error) {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return "", err
	}
	tok := strings.TrimSpace(string(out))
	if tok == "" {
		return "", fmt.Errorf("gh returned empty token")
	}
	return tok, nil
}

func GetNumSuggestions() int {
	if cfg == nil {
		InitConfig()
	}
	providerConfig, err := GetActiveProviderConfig()
	if err != nil {
		return 10 // Default to 10 if error
	}
	if providerConfig.NumSuggestions <= 0 {
		return 10 // Default to 10 if not set or invalid
	}
	return providerConfig.NumSuggestions
}

func SetNumSuggestions(provider, numSuggestions string) error {
	if cfg == nil {
		InitConfig()
	}
	viper.Set(fmt.Sprintf("providers.%s.num_suggestions", provider), numSuggestions)
	return viper.WriteConfig()
}

func GetLanguage() string {
	if cfg == nil {
		InitConfig()
	}
	if cfg.Language == "" {
		return "en" // Default to English
	}
	return cfg.Language
}

func SetLanguage(language string) error {
	if cfg == nil {
		InitConfig()
	}
	cfg.Language = language
	viper.Set("language", language)
	return viper.WriteConfig()
}

func getConfigDir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig
	} else if runtime.GOOS == "windows" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting user home directory:", err)
			os.Exit(1)
		}
		return filepath.Join(homeDir, "AppData", "Local")
	} else {
		return filepath.Join(os.Getenv("HOME"), ".config")
	}
}
