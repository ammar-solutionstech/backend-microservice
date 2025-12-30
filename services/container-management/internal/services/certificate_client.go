package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"backend/services/container-management/internal/config"
)

// CertificateClient is an HTTP client for communicating with the certificate service via mTLS
type CertificateClient struct {
	baseURL    string
	httpClient *http.Client
}

// CSRResponse represents a response from the certificate service for CSR operations
type CSRResponse struct {
	ID              int     `json:"id"`
	CSRPEM          string  `json:"csr_pem"`
	Status          string  `json:"status"`
	RequesterUserID *int    `json:"requester_user_id,omitempty"`
	RequesterEmail  string  `json:"requester_email"`
	ApprovedBy      *int    `json:"approved_by,omitempty"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// CertificateResponse represents a certificate from the certificate service
type CertificateResponse struct {
	ID             int    `json:"id"`
	SerialNumber   string `json:"serial_number"`
	CSRID          *int   `json:"csr_id,omitempty"`
	CertificatePEM string `json:"certificate_pem"`
	IssuedAt       string `json:"issued_at"`
	ExpiresAt      string `json:"expires_at"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// NewCertificateClient creates a new certificate service client with mTLS
func NewCertificateClient(cfg *config.Config) (*CertificateClient, error) {
	var httpClient *http.Client

	if cfg.CertificateServiceTLSConfig != nil {
		// Create HTTP client with mTLS configuration
		transport := &http.Transport{
			TLSClientConfig: cfg.CertificateServiceTLSConfig,
		}
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
		log.Println("Warning: Starting Certificate Client with TLS. mTLS is configured.")
		log.Printf("Certificate Service URL: %s", cfg.CertificateServiceURL)
		log.Printf("Certificate Service TLS Config: %+v", cfg.CertificateServiceTLSConfig)
		log.Printf("Certificate Service CA: %s", cfg.CertificateServiceCA)
		log.Printf("Certificate Service Cert: %s", cfg.CertificateServiceCert)
		log.Printf("Certificate Service Key: %s", cfg.CertificateServiceKey)
		log.Printf("Certificate Service Key Pass: %s", cfg.CertificateServiceKeyPass)
	} else {
		// Fallback to regular HTTP client (not recommended for production)
		log.Println("Warning: Starting Certificate Client without TLS. mTLS is not configured.")
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &CertificateClient{
		baseURL:    cfg.CertificateServiceURL,
		httpClient: httpClient,
	}, nil
}

// SubmitCSR submits a CSR to the certificate service
func (c *CertificateClient) SubmitCSR(csrPEM string, requesterEmail string) (*CSRResponse, error) {
	url := fmt.Sprintf("%s/api/csr", c.baseURL)

	reqBody := map[string]interface{}{
		"csr_pem":         csrPEM,
		"requester_email": requesterEmail,
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
		return nil, fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var csrResp CSRResponse
	if err := json.Unmarshal(bodyBytes, &csrResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v. Response was: %s", err, string(bodyBytes))
	}

	return &csrResp, nil
}

// GetCSR retrieves a CSR by ID from the certificate service
func (c *CertificateClient) GetCSR(id int) (*CSRResponse, error) {
	url := fmt.Sprintf("%s/api/csr/%d", c.baseURL, id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var csrResp CSRResponse
	if err := json.Unmarshal(bodyBytes, &csrResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &csrResp, nil
}

// ApproveCSR approves a CSR in the certificate service
func (c *CertificateClient) ApproveCSR(id int, approverID int, autoIssue bool) error {
	url := fmt.Sprintf("%s/api/csr/%d/approve", c.baseURL, id)

	reqBody := map[string]interface{}{
		"approver_id": approverID,
		"auto_issue":  autoIssue,
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
		return fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// IssueCertificate issues a certificate from an approved CSR
func (c *CertificateClient) IssueCertificate(csrID int) (*CertificateResponse, error) {
	url := fmt.Sprintf("%s/api/certificates", c.baseURL)

	reqBody := map[string]interface{}{
		"csr_id": csrID,
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
		return nil, fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
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
		return nil, fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
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
		return fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// ListCertificates lists certificates with optional filters
func (c *CertificateClient) ListCertificates(status string) ([]CertificateResponse, error) {
	url := fmt.Sprintf("%s/api/certificates", c.baseURL)

	if status != "" {
		url += "?status=" + status
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var certsResp []CertificateResponse
	if err := json.Unmarshal(bodyBytes, &certsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return certsResp, nil
}

// RequestAgentCertificate requests an agent certificate from certificate service
// This is called by client container after verification is complete
// Note: Verification is done by client container, not here
func (c *CertificateClient) RequestAgentCertificate(containerID, deviceSerial, csrPEM string, containerInfo map[string]string) (*CertificateResponse, error) {
	url := fmt.Sprintf("%s/api/certificates/request", c.baseURL)

	reqBody := map[string]interface{}{
		"csr_pem":         csrPEM,
		"requester_email": fmt.Sprintf("container:%s:device:%s", containerID, deviceSerial),
		"container_id":    containerInfo["container_id"],
		"org_id":          containerInfo["org_id"],
		"org_domain":      containerInfo["org_domain"],
		"device_serial":   deviceSerial,
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
		return nil, fmt.Errorf("failed to call certificate service: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("certificate service returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var certResp CertificateResponse
	if err := json.Unmarshal(bodyBytes, &certResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v. Response was: %s", err, string(bodyBytes))
	}

	return &certResp, nil
}
