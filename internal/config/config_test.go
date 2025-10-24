package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestGetAPIKey_EnvironmentVariable(t *testing.T) {
	// Reset configuration for clean test
	cfg = nil
	viper.Reset()

	// Set up test environment variable
	testEnvVar := "TEST_API_KEY"
	testAPIKey := "test-api-key-value-123"
	os.Setenv(testEnvVar, testAPIKey)
	defer os.Unsetenv(testEnvVar)

	// Initialize config
	InitConfig()

	// Set up test provider with environment variable reference
	testProvider := "openrouter"
	cfg.ActiveProvider = testProvider
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	cfg.Providers[testProvider] = ProviderConfig{
		APIKey: "$" + testEnvVar,
		Model:  "test-model",
	}

	// Test that environment variable is resolved
	resolvedKey, err := GetAPIKey()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resolvedKey != testAPIKey {
		t.Errorf("Expected resolved API key to be %s, got %s", testAPIKey, resolvedKey)
	}
}

func TestGetAPIKey_EnvironmentVariableNotSet(t *testing.T) {
	// Reset configuration for clean test
	cfg = nil
	viper.Reset()

	// Initialize config
	InitConfig()

	// Set up test provider with environment variable reference that doesn't exist
	testProvider := "openai"
	cfg.ActiveProvider = testProvider
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	cfg.Providers[testProvider] = ProviderConfig{
		APIKey: "$NONEXISTENT_API_KEY",
		Model:  "test-model",
	}

	// Test that missing environment variable returns error
	_, err := GetAPIKey()
	if err == nil {
		t.Fatal("Expected error for missing environment variable, got nil")
	}

	expectedError := "environment variable 'NONEXISTENT_API_KEY' for provider 'openai' is not set or empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestGetAPIKey_RegularAPIKey(t *testing.T) {
	// Reset configuration for clean test
	cfg = nil
	viper.Reset()

	// Initialize config
	InitConfig()

	// Set up test provider with regular API key (not environment variable)
	testProvider := "openai"
	testAPIKey := "regular-api-key-123"
	cfg.ActiveProvider = testProvider
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	cfg.Providers[testProvider] = ProviderConfig{
		APIKey: testAPIKey,
		Model:  "test-model",
	}

	// Test that regular API key is returned as-is
	resolvedKey, err := GetAPIKey()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resolvedKey != testAPIKey {
		t.Errorf("Expected API key to be %s, got %s", testAPIKey, resolvedKey)
	}
}

func TestGetEndpoint_DefaultEndpoints(t *testing.T) {
	// Reset configuration for clean test
	cfg = nil
	viper.Reset()

	// Test default endpoints for different providers
	testCases := []struct {
		provider string
		expected string
	}{
		{"openai", "https://api.openai.com/v1"},
		{"copilot", "https://api.githubcopilot.com"},
	}

	for _, tc := range testCases {
		// Initialize config
		InitConfig()

		// Set up test provider without custom endpoint
		cfg.ActiveProvider = tc.provider
		if cfg.Providers == nil {
			cfg.Providers = make(map[string]ProviderConfig)
		}
		cfg.Providers[tc.provider] = ProviderConfig{
			APIKey: "test-key",
			Model:  "test-model",
			// No EndpointURL set - should use default
		}

		// Test that default endpoint is returned
		endpoint, err := GetEndpoint()
		if err != nil {
			t.Fatalf("Expected no error for provider %s, got: %v", tc.provider, err)
		}

		if endpoint != tc.expected {
			t.Errorf("Expected endpoint %s for provider %s, got %s", tc.expected, tc.provider, endpoint)
		}
	}
}

func TestGetEndpoint_CustomEndpoint(t *testing.T) {
	// Reset configuration for clean test
	cfg = nil
	viper.Reset()

	// Initialize config
	InitConfig()

	// Set up test provider with custom endpoint
	testProvider := "openai"
	customEndpoint := "https://custom.api.com/v1"
	cfg.ActiveProvider = testProvider
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	cfg.Providers[testProvider] = ProviderConfig{
		APIKey:     "test-key",
		Model:      "test-model",
		EndpointURL: customEndpoint,
	}

	// Test that custom endpoint is returned
	endpoint, err := GetEndpoint()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if endpoint != customEndpoint {
		t.Errorf("Expected custom endpoint %s, got %s", customEndpoint, endpoint)
	}
}

func TestGetEndpoint_UnknownProvider(t *testing.T) {
	// Reset configuration for clean test
	cfg = nil
	viper.Reset()

	// Initialize config
	InitConfig()

	// Set up unknown provider without custom endpoint
	testProvider := "unknown-provider"
	cfg.ActiveProvider = testProvider
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	cfg.Providers[testProvider] = ProviderConfig{
		APIKey: "test-key",
		Model:  "test-model",
	}

	// Test that unknown provider without custom endpoint returns error
	_, err := GetEndpoint()
	if err == nil {
		t.Fatal("Expected error for unknown provider, got nil")
	}

	expectedError := "no default endpoint available for provider 'unknown-provider'"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSetEndpoint_Validation(t *testing.T) {
	// Reset configuration for clean test
	cfg = nil
	viper.Reset()

	// Initialize config
	InitConfig()

	testCases := []struct {
		endpoint string
		valid    bool
	}{
		{"", true},                           // Empty should be valid (default)
		{"https://api.openai.com/v1", true},  // Valid HTTPS URL
		{"http://localhost:11434", true},     // Valid HTTP URL
		{"ftp://invalid.com", false},         // Invalid protocol
		{"not-a-url", false},                 // Invalid format
		{"https://", false},                  // Missing host
	}

	for _, tc := range testCases {
		err := SetEndpoint("test", tc.endpoint)
		if tc.valid && err != nil {
			t.Errorf("Expected valid endpoint %s to pass, but got error: %v", tc.endpoint, err)
		} else if !tc.valid && err == nil {
			t.Errorf("Expected invalid endpoint %s to fail, but it passed", tc.endpoint)
		}
	}
}