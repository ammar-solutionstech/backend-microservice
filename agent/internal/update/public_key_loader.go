package update

import (
	"crypto/ed25519"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"backend/agent/config"
	"backend/agent/internal/security"
	"backend/agent/internal/utils"
)

// PublicKeyLoader handles loading the update public key from various sources
type PublicKeyLoader struct {
	config     *config.Config
	logger     *utils.Logger
	httpClient *http.Client
}

// NewPublicKeyLoader creates a new public key loader
func NewPublicKeyLoader(cfg *config.Config, logger *utils.Logger) *PublicKeyLoader {
	return &PublicKeyLoader{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LoadPublicKey loads the Ed25519 public key from multiple sources
func (pkl *PublicKeyLoader) LoadPublicKey() (ed25519.PublicKey, error) {
	// Try 1: Load from credential store (stored during registration)
	credentialStore := security.NewCredentialStore(pkl.config.DataDir)
	if keyBytes, err := credentialStore.Retrieve("update_public_key"); err == nil {
		if len(keyBytes) == ed25519.PublicKeySize {
			return ed25519.PublicKey(keyBytes), nil
		}
		// Try to decode as PEM
		if key, err := pkl.decodePublicKeyPEM(keyBytes); err == nil {
			return key, nil
		}
	}

	// Try 2: Load from cached file in config directory
	cachedKeyPath := filepath.Join(pkl.config.ConfigDir, "update_public_key.pem")
	if keyBytes, err := os.ReadFile(cachedKeyPath); err == nil {
		if key, err := pkl.decodePublicKeyPEM(keyBytes); err == nil {
			return key, nil
		}
	}

	// Try 3: Fetch from backend endpoint
	key, err := pkl.fetchFromBackend()
	if err == nil {
		// Cache the key for future use
		if cacheErr := pkl.cachePublicKey(key); cacheErr != nil {
			pkl.logger.Warn("Failed to cache public key", map[string]interface{}{
				"error": cacheErr.Error(),
			})
		}
		return key, nil
	}

	return nil, fmt.Errorf("failed to load public key from any source: %v", err)
}

// fetchFromBackend fetches the public key from the backend API
func (pkl *PublicKeyLoader) fetchFromBackend() (ed25519.PublicKey, error) {
	url := fmt.Sprintf("%s/api/v1/agents/public-key", pkl.config.BackendURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// TODO: Add mTLS client certificate if available

	resp, err := pkl.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch public key: status %d", resp.StatusCode)
	}

	var response struct {
		PublicKey string `json:"public_key"`
		Algorithm string `json:"algorithm"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Algorithm != "ed25519" {
		return nil, fmt.Errorf("unsupported algorithm: %s", response.Algorithm)
	}

	// Decode PEM-encoded public key
	return pkl.decodePublicKeyPEM([]byte(response.PublicKey))
}

// decodePublicKeyPEM decodes a PEM-encoded Ed25519 public key
func (pkl *PublicKeyLoader) decodePublicKeyPEM(pemData []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}

	if block.Type != "PUBLIC KEY" && block.Type != "ED25519 PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM type: %s", block.Type)
	}

	if len(block.Bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(block.Bytes))
	}

	return ed25519.PublicKey(block.Bytes), nil
}

// cachePublicKey caches the public key to file and credential store
func (pkl *PublicKeyLoader) cachePublicKey(key ed25519.PublicKey) error {
	// Cache to file
	cachedKeyPath := filepath.Join(pkl.config.ConfigDir, "update_public_key.pem")
	
	// Encode as PEM
	block := &pem.Block{
		Type:  "ED25519 PUBLIC KEY",
		Bytes: key,
	}
	pemData := pem.EncodeToMemory(block)

	/* if err := os.MkdirAll(filepath.Dir(cachedKeyPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	} */

	if err := os.WriteFile(cachedKeyPath, pemData, 0600); err != nil {
		return fmt.Errorf("failed to write cached public key: %v", err)
	}

	// Also store in credential store
	credentialStore := security.NewCredentialStore(pkl.config.DataDir)
	if err := credentialStore.Store("update_public_key", key); err != nil {
		// Log but don't fail
		pkl.logger.Warn("Failed to store public key in credential store", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return nil
}

