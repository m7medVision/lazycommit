package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// ProviderConfig holds the configuration for a single LLM provider.
type ProviderConfig struct {
	APIKey string `mapstructure:"api_key"`
	Model  string `mapstructure:"model"`
}

// Config is the main configuration structure for the application.
type Config struct {
	Providers      map[string]ProviderConfig `mapstructure:"providers"`
	ActiveProvider string                    `mapstructure:"active_provider"`
}

var cfg *Config

// InitConfig initializes the configuration from environment variables and config files.
func InitConfig() {
	viper.SetConfigName(".lazycommit")
	viper.SetConfigType("yaml")
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	viper.AddConfigPath(home)
	viper.AddConfigPath(".") // Also look in the current directory

	// Set defaults based on available credentials
	if token, err := LoadGitHubToken(); err == nil && token != "" {
		viper.SetDefault("active_provider", "copilot")
		viper.SetDefault("providers.copilot.api_key", token)
		viper.SetDefault("providers.copilot.model", "openai/gpt-4o") // Use GitHub Models format
	} else {
		viper.SetDefault("active_provider", "openai")
		viper.SetDefault("providers.openai.model", "gpt-3.5-turbo")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("Error reading config file:", err)
		}
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("Error unmarshalling config:", err)
		os.Exit(1)
	}
}

// GetProvider returns the active provider's name.
func GetProvider() string {
	if cfg == nil {
		InitConfig()
	}
	return cfg.ActiveProvider
}

// GetActiveProviderConfig returns the configuration for the currently active provider.
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

// GetAPIKey returns the API key for the active provider.
func GetAPIKey() (string, error) {
	providerConfig, err := GetActiveProviderConfig()
	if err != nil {
		// If config for the provider doesn't exist, that's an issue.
		// However, for copilot, we can try to load the token directly.
		if cfg.ActiveProvider == "copilot" {
			return LoadGitHubToken()
		}
		return "", err
	}

	// Special handling for copilot to always prefer a freshly loaded token.
	if cfg.ActiveProvider == "copilot" {
		return LoadGitHubToken()
	}

	if providerConfig.APIKey == "" {
		return "", fmt.Errorf("API key for provider '%s' is not set", cfg.ActiveProvider)
	}

	return providerConfig.APIKey, nil
}

// GetModel returns the model for the active provider.
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

// SetProvider sets the active provider and saves the config.
func SetProvider(provider string) error {
	if cfg == nil {
		InitConfig()
	}
	cfg.ActiveProvider = provider
	viper.Set("active_provider", provider)
	return viper.WriteConfig()
}

// SetModel sets the model for the active provider and saves the config.
func SetModel(model string) error {
	if cfg == nil {
		InitConfig()
	}
	provider := cfg.ActiveProvider
	viper.Set(fmt.Sprintf("providers.%s.model", provider), model)
	return viper.WriteConfig()
}

// SetAPIKey sets the API key for a specific provider and saves the config.
func SetAPIKey(provider, apiKey string) error {
	if cfg == nil {
		InitConfig()
	}
	viper.Set(fmt.Sprintf("providers.%s.api_key", provider), apiKey)
	return viper.WriteConfig()
}

// LoadGitHubToken tries to load a GitHub token with models scope from standard locations.
func LoadGitHubToken() (string, error) {
	// First check environment variable (recommended approach)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	// Also check for a GitHub Models specific token
	if token := os.Getenv("GITHUB_MODELS_TOKEN"); token != "" {
		return token, nil
	}

	// Fallback: try to find tokens from GitHub Copilot IDE installations
	// Note: These tokens may not have the required 'models' scope
	configDir := getConfigDir()

	// Try both hosts.json and apps.json files
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

	return "", fmt.Errorf("GitHub token not found. Please set GITHUB_TOKEN or GITHUB_MODELS_TOKEN environment variable with a Personal Access Token that has 'models' scope")
}

func getConfigDir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig
	} else if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return localAppData
		} else {
			return filepath.Join(os.Getenv("HOME"), "AppData", "Local")
		}
	} else {
		return filepath.Join(os.Getenv("HOME"), ".config")
	}
}
