package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type ProviderConfig struct {
	APIKey     string `mapstructure:"api_key"`
	Model      string `mapstructure:"model"`
	EndpointURL string `mapstructure:"endpoint_url"`
}

type Config struct {
	Providers      map[string]ProviderConfig `mapstructure:"providers"`
	ActiveProvider string                    `mapstructure:"active_provider"`
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
		viper.SetDefault("providers.openai.model", "gpt-5-mini")
	}

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
	if cfg.ActiveProvider == "copilot" {
		return LoadGitHubToken()
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
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	if token := os.Getenv("GITHUB_MODELS_TOKEN"); token != "" {
		return token, nil
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

	return "", fmt.Errorf("GitHub token not found. Please set GITHUB_TOKEN or GITHUB_MODELS_TOKEN environment variable with a Personal Access Token that has 'models' scope")
}

func getConfigDir() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig

	// WARNING: The code is not woking
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
