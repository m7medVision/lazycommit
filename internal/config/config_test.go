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
	testProvider := "openrouter"
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

	expectedError := "environment variable 'NONEXISTENT_API_KEY' for provider 'openrouter' is not set or empty"
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
	testProvider := "openrouter"
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