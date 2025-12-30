package communication

import (
	"crypto/ed25519"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"backend/agent/config"
	"backend/agent/internal/utils"
)

// PublicKeyClient fetches the update public key from the backend
type PublicKeyClient struct {
	config *config.Config
	logger *utils.Logger
	client *http.Client
}

// NewPublicKeyClient creates a new public key client
func NewPublicKeyClient(cfg *config.Config, logger *utils.Logger) *PublicKeyClient {
	return &PublicKeyClient{
		config: cfg,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchPublicKey fetches the Ed25519 public key from the backend
func (c *PublicKeyClient) FetchPublicKey() (ed25519.PublicKey, error) {
	url := fmt.Sprintf("%s/api/v1/agents/public-key", c.config.BackendURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// TODO: Add mTLS client certificate if available

	resp, err := c.client.Do(req)
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
	block, _ := pem.Decode([]byte(response.PublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM public key")
	}

	if block.Type != "PUBLIC KEY" && block.Type != "ED25519 PUBLIC KEY" {
		return nil, fmt.Errorf("unexpected PEM type: %s", block.Type)
	}

	// Parse Ed25519 public key (32 bytes)
	if len(block.Bytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: expected %d, got %d", ed25519.PublicKeySize, len(block.Bytes))
	}

	return ed25519.PublicKey(block.Bytes), nil
}

