package services

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/services/certificate/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwe"
	"github.com/lestrrat-go/jwx/jwk"
)

// StepCAClient wraps step-ca REST API calls
type StepCAClient struct {
	baseURL     string
	audiance    string
	token       string
	tokenPath   string
	keyPassword string
	provisioner string
	httpClient  *http.Client
}

// SignResponse represents the response from step-ca sign endpoint
type SignResponse struct {
	CertPEM  string   `json:"crt"`
	CaPEM    string   `json:"ca"`
	ChainPEM []string `json:"certChain"`
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
		baseURL:  cfg.StepCAURL,
		audiance: cfg.StepCAAudience,
		//token:       cfg.StepCAToken,
		tokenPath:   cfg.StepCAKeyPath,
		keyPassword: cfg.StepCAKeyPassword,
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

	// Log CSR details before sending
	csrBlock, _ := pem.Decode([]byte(csrPEM))
	if csrBlock != nil {
		csr, parseErr := x509.ParseCertificateRequest(csrBlock.Bytes)
		if parseErr == nil {
			log.Printf("Sending CSR to step-ca - CN: %s, O: %v, OU: %v, C: %v",
				csr.Subject.CommonName,
				csr.Subject.Organization,
				csr.Subject.OrganizationalUnit,
				csr.Subject.Country,
			)
		}
	}

	url := fmt.Sprintf("%s/1.0/sign", c.baseURL)

	reqBody := map[string]interface{}{
		"csr": csrPEM,
	}

	subject, err := extractSubjectFromCSR(csrPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to extract subject from CSR: %v", err)
	}
	log.Printf("Generating signed JWT for provisioner: %s, keyPath: %s, audience: %s, subject: %s", c.provisioner, c.tokenPath, c.baseURL, subject)

	signedJWT, err := generateSignedJWT(c.provisioner, c.tokenPath, c.baseURL+c.audiance, subject, c.keyPassword)
	//signedJWT, err := generateSignedJWT(c.provisioner, c.tokenPath, c.baseURL, subject, c.keyPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to generate automated token: %v", err)
	}
	c.token = signedJWT

	// Add token to request body as "ott" (one-time token) if available
	if c.token != "" {
		reqBody["ott"] = c.token
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
	// Also keep Authorization header for compatibility
	if c.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call step-ca: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body first to check what we got
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Log the actual response for debugging
	log.Printf("step-ca sign response (status %d): %s", resp.StatusCode, string(bodyBytes))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("step-ca returned error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	// Try to decode the response
	var signResp SignResponse
	if err := json.Unmarshal(bodyBytes, &signResp); err != nil {
		// If decoding fails, log the raw response for debugging
		log.Printf("Failed to decode response as SignResponse. Raw response: %s", string(bodyBytes))
		return nil, fmt.Errorf("failed to decode response: %v. Response was: %s", err, string(bodyBytes))
	}

	// Validate that we got the certificate
	if signResp.CertPEM == "" {
		return nil, fmt.Errorf("step-ca response missing certificate. Response was: %s", string(bodyBytes))
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
	log.Println(url)
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

// Update the function to accept password and decrypt JWE
func generateSignedJWT(provisionerName, keyPath, audience, subject, keyPassword string) (string, error) {
	// 1. Read the encrypted JWK private key file
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key file: %w", err)
	}

	// 2. Check if the key is encrypted (JWE format) or plain JWK
	var decryptedKeyBytes []byte

	// Try to detect JWE format (has "protected", "encrypted_key", "ciphertext" fields)
	var jweTest map[string]interface{}
	if err := json.Unmarshal(keyBytes, &jweTest); err == nil {
		if _, isJWE := jweTest["protected"]; isJWE {
			// This is a JWE encrypted key, decrypt it
			if keyPassword == "" {
				return "", fmt.Errorf("key is password-protected but no password provided")
			}

			decryptedKeyBytes, err = jwe.Decrypt(keyBytes, jwa.KeyEncryptionAlgorithm("PBES2-HS256+A128KW"), []byte(keyPassword))
			if err != nil {
				return "", fmt.Errorf("failed to decrypt JWE key: %w", err)
			}
		} else {
			// Plain JWK, use as-is
			decryptedKeyBytes = keyBytes
		}
	} else {
		// If JSON parsing fails, assume it's plain JWK
		decryptedKeyBytes = keyBytes
	}

	// 3. Parse the decrypted JWK into a jwx Key object
	key, err := jwk.ParseKey(decryptedKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse JWK: %w", err)
	}
	log.Printf("Decrypted key (first 200 chars): %s", string(decryptedKeyBytes[:min(200, len(decryptedKeyBytes))]))

	// 4. Get key type and other info from JSON directly (more reliable)
	var keys config.Keys
	if err := json.Unmarshal(decryptedKeyBytes, &keys); err != nil {
		return "", fmt.Errorf("failed to unmarshal key JSON: %w", err)
	}
	ktyStr := keys.Kty
	kid := keys.Kid
	log.Printf("Key type (kty): %s, kid: %s", ktyStr, kid)

	// 5. Convert the jwx Key into the appropriate private key type
	var rawKey interface{}
	if err := key.Raw(&rawKey); err != nil {
		return "", fmt.Errorf("failed to get raw key: %w", err)
	}

	// 6. Create JWT claims
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": provisionerName,                                // Provisioner name
		"sub": subject,                                        // Certificate subject (CN)
		"aud": audience,                                       // CA URL
		"iat": jwt.NewNumericDate(now),                        // Issued At
		"nbf": jwt.NewNumericDate(now.Add(-30 * time.Second)), // Not Before
		"exp": jwt.NewNumericDate(now.Add(15 * time.Minute)),  // Expiration
	}

	// 7. Create token with appropriate signing method based on key type
	var token *jwt.Token
	var signingKey interface{}

	switch ktyStr {
	case "EC": // ECDSA
		ecdsaKey, ok := rawKey.(*ecdsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("key type is EC but failed to cast to ECDSA private key")
		}

		// Determine curve and use appropriate signing method
		crvStr := keys.Crv
		switch crvStr {
		case "P-256":
			token = jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		case "P-384":
			token = jwt.NewWithClaims(jwt.SigningMethodES384, claims)
		case "P-521":
			token = jwt.NewWithClaims(jwt.SigningMethodES512, claims)
		default:
			// Default to ES256 for ECDSA
			token = jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		}
		signingKey = ecdsaKey

	case "RSA": // RSA
		rsaKey, ok := rawKey.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("key type is RSA but failed to cast to RSA private key")
		}
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signingKey = rsaKey

	case "OKP": // Ed25519 (Octet Key Pair)
		ed25519Key, ok := rawKey.(ed25519.PrivateKey)
		if !ok {
			return "", fmt.Errorf("key type is OKP but failed to cast to Ed25519 private key")
		}
		token = jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		signingKey = ed25519Key

	default:
		return "", fmt.Errorf("unsupported key type: %s (supported: EC, RSA, OKP)", ktyStr)
	}

	// 8. Set kid in header if available
	if kid != "" {
		token.Header["kid"] = kid
	}

	// 9. Sign the token with the appropriate private key
	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	log.Printf("Generated JWT token (first 100 chars): %s", signedToken[:min(100, len(signedToken))])

	// Decode and log the JWT payload to verify
	parts := strings.Split(signedToken, ".")
	if len(parts) >= 2 {
		// Decode the payload (second part)
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				log.Printf("JWT Payload - iss: %v, aud: %v, sub: %v, kid: %v",
					payload["iss"], payload["aud"], payload["sub"], kid)
			}
		}
	}

	// Also log the header
	if len(parts) >= 1 {
		headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err == nil {
			var header map[string]interface{}
			if err := json.Unmarshal(headerBytes, &header); err == nil {
				log.Printf("JWT Header - alg: %v, kid: %v, typ: %v",
					header["alg"], header["kid"], header["typ"])
			}
		}
	}

	return signedToken, nil
}

// Helper function to extract CN from CSR
func extractSubjectFromCSR(csrPEM string) (string, error) {
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode CSR PEM")
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse CSR: %w", err)
	}

	if csr.Subject.CommonName != "" {
		return csr.Subject.CommonName, nil
	}

	// Fallback to first SAN if CN is empty
	if len(csr.DNSNames) > 0 {
		return csr.DNSNames[0], nil
	}

	return "", fmt.Errorf("incorrect CSR subject : %v", csr)
}
