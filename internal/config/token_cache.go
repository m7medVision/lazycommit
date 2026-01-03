package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CopilotTokenCache holds a cached Copilot API token with expiration metadata.
type CopilotTokenCache struct {
	Token           string    `json:"token"`
	ExpiresAt       time.Time `json:"expires_at"`
	GitHubTokenHash string    `json:"github_token_hash"`
}

var (
	memoryCache   *CopilotTokenCache
	memoryCacheMu sync.RWMutex
)

func getCacheFilePath() string {
	return filepath.Join(getConfigDir(), ".lazycommit.token.cache")
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GetCachedCopilotToken returns a valid cached token for the given GitHub token, or nil if unavailable.
func GetCachedCopilotToken(githubToken string) *CopilotTokenCache {
	tokenHash := hashToken(githubToken)

	memoryCacheMu.RLock()
	if memoryCache != nil &&
		memoryCache.GitHubTokenHash == tokenHash &&
		isTokenValid(memoryCache) {
		cache := *memoryCache
		memoryCacheMu.RUnlock()
		return &cache
	}
	memoryCacheMu.RUnlock()

	cache, err := loadCacheFromDisk()
	if err != nil {
		return nil
	}

	if cache.GitHubTokenHash != tokenHash || !isTokenValid(cache) {
		return nil
	}

	memoryCacheMu.Lock()
	memoryCache = cache
	memoryCacheMu.Unlock()

	return cache
}

// SaveCopilotTokenCache persists the token to memory and disk.
func SaveCopilotTokenCache(token string, expiresAtUnix int64, githubToken string) error {
	cache := &CopilotTokenCache{
		Token:           token,
		ExpiresAt:       time.Unix(expiresAtUnix, 0),
		GitHubTokenHash: hashToken(githubToken),
	}

	memoryCacheMu.Lock()
	memoryCache = cache
	memoryCacheMu.Unlock()

	return saveCacheToDisk(cache)
}

// InvalidateCopilotTokenCache clears the cached token from memory and disk.
func InvalidateCopilotTokenCache() {
	memoryCacheMu.Lock()
	memoryCache = nil
	memoryCacheMu.Unlock()

	_ = os.Remove(getCacheFilePath())
}

func isTokenValid(cache *CopilotTokenCache) bool {
	if cache == nil || cache.Token == "" {
		return false
	}
	expirationBuffer := 60 * time.Second
	return time.Now().Add(expirationBuffer).Before(cache.ExpiresAt)
}

func loadCacheFromDisk() (*CopilotTokenCache, error) {
	data, err := os.ReadFile(getCacheFilePath())
	if err != nil {
		return nil, err
	}

	var cache CopilotTokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func saveCacheToDisk(cache *CopilotTokenCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal token cache: %w", err)
	}

	cacheDir := filepath.Dir(getCacheFilePath())
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	if err := os.WriteFile(getCacheFilePath(), data, 0600); err != nil {
		return fmt.Errorf("failed to write token cache: %w", err)
	}

	return nil
}
