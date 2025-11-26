package services

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"backend/services/certificate/internal/config"
)

// StepCAClient wraps step-ca REST API calls
type StepCAClient struct {
	baseURL     string
	token       string
	provisioner string
	httpClient  *http.Client
}

// SignResponse represents the response from step-ca sign endpoint
type SignResponse struct {
	CertPEM string `json:"cert"`
	CaPEM   string `json:"ca"`
}

// CertificateResponse represents a certificate from step-ca
type CertificateResponse struct {
	ID          string    `json:"id"`
	Serial      string    `json:"serial"`
	Certificate string    `json:"cert"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
	Status      string    `json:"status"`
}

// OCSPResponse represents OCSP status
type OCSPResponse struct {
	Status string `json:"status"` // good, revoked, unknown
}

// NewStepCAClient creates a new step-ca client
func NewStepCAClient(cfg *config.Config) (*StepCAClient, error) {
	client := &StepCAClient{
		baseURL:     cfg.StepCAURL,
		token:       cfg.StepCAToken,
		provisioner: cfg.StepCAProvisioner,
	}

	// Configure HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.StepCAUseMTLS, // Skip verify if not using mTLS
		},
	}

	// If mTLS is enabled, load client certificates
	if cfg.StepCAUseMTLS && cfg.StepCACert != "" && cfg.StepCAKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.StepCACert, cfg.StepCAKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificates: %v", err)
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	// If root CA is provided, add it to the certificate pool
	if cfg.StepCARootCA != "" {
		caCert, err := x509.SystemCertPool()
		if err != nil {
			caCert = x509.NewCertPool()
		}
		// Load custom root CA if provided
		// Note: In production, load from file
		transport.TLSClientConfig.RootCAs = caCert
	}

	client.httpClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client, nil
}

// SignCSR signs a CSR and returns the certificate
func (c *StepCAClient) SignCSR(csrPEM string) (*SignResponse, error) {
	url := fmt.Sprintf("%s/1.0/sign", c.baseURL)

	reqBody := map[string]interface{}{
		"csr": csrPEM,
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
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call step-ca: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("step-ca returned error: %d - %s", resp.StatusCode, string(body))
	}

	var signResp SignResponse
	if err := json.NewDecoder(resp.Body).Decode(&signResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &signResp, nil
}

// GetCertificate retrieves a certificate by serial number
func (c *StepCAClient) GetCertificate(serial string) (*CertificateResponse, error) {
	url := fmt.Sprintf("%s/1.0/certificates/%s", c.baseURL, serial)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call step-ca: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("certificate not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("step-ca returned error: %d - %s", resp.StatusCode, string(body))
	}

	var certResp CertificateResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &certResp, nil
}

// RevokeCertificate revokes a certificate
func (c *StepCAClient) RevokeCertificate(serial string, reason int) error {
	url := fmt.Sprintf("%s/1.0/revoke", c.baseURL)

	reqBody := map[string]interface{}{
		"serial": serial,
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
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call step-ca: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("step-ca returned error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// RenewCertificate renews a certificate (typically by signing a new CSR)
func (c *StepCAClient) RenewCertificate(csrPEM string) (*SignResponse, error) {
	// Renewal is essentially signing a new CSR
	return c.SignCSR(csrPEM)
}

// GetCRL retrieves the Certificate Revocation List
func (c *StepCAClient) GetCRL() ([]byte, error) {
	url := fmt.Sprintf("%s/1.0/crl", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call step-ca: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("step-ca returned error: %d - %s", resp.StatusCode, string(body))
	}

	crl, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read CRL: %v", err)
	}

	return crl, nil
}

// GetOCSPStatus gets OCSP status for a certificate
func (c *StepCAClient) GetOCSPStatus(serial string) (*OCSPResponse, error) {
	// Note: step-ca may not have a direct OCSP endpoint
	// This is a simplified implementation
	// In production, you might need to query OCSP responder directly
	url := fmt.Sprintf("%s/1.0/certificates/%s", c.baseURL, serial)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call step-ca: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &OCSPResponse{Status: "unknown"}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("step-ca returned error: %d - %s", resp.StatusCode, string(body))
	}

	var certResp CertificateResponse
	if err := json.NewDecoder(resp.Body).Decode(&certResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Map certificate status to OCSP status
	status := "good"
	if certResp.Status == "revoked" {
		status = "revoked"
	}

	return &OCSPResponse{Status: status}, nil
}

// ListCertificates lists certificates from step-ca
func (c *StepCAClient) ListCertificates(limit, offset int) ([]CertificateResponse, error) {
	url := fmt.Sprintf("%s/1.0/certificates?limit=%d&offset=%d", c.baseURL, limit, offset)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call step-ca: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("step-ca returned error: %d - %s", resp.StatusCode, string(body))
	}

	var certs []CertificateResponse
	if err := json.NewDecoder(resp.Body).Decode(&certs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return certs, nil
}

// CheckConnectivity checks if step-ca is reachable
func (c *StepCAClient) CheckConnectivity() error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("step-ca not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("step-ca health check failed: %d", resp.StatusCode)
	}

	log.Println("step-ca connectivity check passed")
	return nil
}
