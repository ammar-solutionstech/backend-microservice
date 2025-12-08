package services

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"backend/services/client-container/internal/config"
)

// CertificateClient is an HTTP client for communicating with container-management service
type CertificateClient struct {
	baseURL    string
	httpClient *http.Client
}

// CertificateResponse represents a certificate response
type CertificateResponse struct {
	ID             int     `json:"id"`
	SerialNumber   string  `json:"serial_number"`
	CSRID          *int    `json:"csr_id,omitempty"`
	CertificatePEM string  `json:"certificate_pem"`
	IssuedAt       string  `json:"issued_at"`
	ExpiresAt      string  `json:"expires_at"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// NewCertificateClient creates a new certificate client with mTLS
func NewCertificateClient(cfg *config.Config) (*CertificateClient, error) {
	var httpClient *http.Client

	if cfg.ContainerMgmtServiceTLSConfig != nil {
		// Create HTTP client with mTLS configuration
		transport := &http.Transport{
			TLSClientConfig: cfg.ContainerMgmtServiceTLSConfig,
		}
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
	} else {
		// Fallback to regular HTTP client (not recommended for production)
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &CertificateClient{
		baseURL:    cfg.ContainerMgmtServiceURL,
		httpClient: httpClient,
	}, nil
}

// RequestContainerCertificate requests a certificate for the container itself
func (c *CertificateClient) RequestContainerCertificate(orgID int, csrPEM string) (*CertificateResponse, error) {
	url := fmt.Sprintf("%s/api/v1/organizations/%d/containers", c.baseURL, orgID)

	reqBody := map[string]interface{}{
		"csr_pem": csrPEM,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call container-management service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("container-management service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var containerResp struct {
		CertificateSerial *string `json:"certificate_serial"`
	}
	if err := json.Unmarshal(bodyBytes, &containerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Get certificate details
	if containerResp.CertificateSerial != nil {
		return c.GetCertificate(*containerResp.CertificateSerial)
	}

	return nil, fmt.Errorf("certificate serial not found in response")
}

// RequestAgentCertificate requests an agent certificate (called after verification)
func (c *CertificateClient) RequestAgentCertificate(containerID, deviceSerial, csrPEM string, containerInfo map[string]string) (*CertificateResponse, error) {
	// This goes through container-management service which forwards to certificate service
	// The container-management service handles the certificate request
	url := fmt.Sprintf("%s/api/v1/containers/%s/certificates/request", c.baseURL, containerID)

	reqBody := map[string]interface{}{
		"csr_pem":       csrPEM,
		"device_serial": deviceSerial,
		"container_id":  containerInfo["container_id"],
		"org_id":        containerInfo["org_id"],
		"org_domain":    containerInfo["org_domain"],
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call container-management service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("container-management service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var certResp CertificateResponse
	if err := json.Unmarshal(bodyBytes, &certResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v. Response was: %s", err, string(bodyBytes))
	}

	return &certResp, nil
}

// GetCertificate retrieves a certificate by serial number
func (c *CertificateClient) GetCertificate(serial string) (*CertificateResponse, error) {
	url := fmt.Sprintf("%s/api/certificates/%s", c.baseURL, serial)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call container-management service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("container-management service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var certResp CertificateResponse
	if err := json.Unmarshal(bodyBytes, &certResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &certResp, nil
}

// RevokeCertificate revokes a certificate
func (c *CertificateClient) RevokeCertificate(serial string, reason int) error {
	url := fmt.Sprintf("%s/api/certificates/%s/revoke", c.baseURL, serial)

	reqBody := map[string]interface{}{
		"reason": reason,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call container-management service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("container-management service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

