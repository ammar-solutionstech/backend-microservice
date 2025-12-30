package communication

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"backend/agent/config"
	"google.golang.org/grpc/credentials"
)

// LoadMTLSCredentials loads mTLS credentials for gRPC client
func LoadMTLSCredentials(cfg *config.Config) (credentials.TransportCredentials, error) {
	if cfg.ClientCertPath == "" || cfg.ClientKeyPath == "" {
		return nil, fmt.Errorf("client certificate and key paths are required")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %v", err)
	}

	// Load CA cert for server verification
	var caCertPool *x509.CertPool
	if cfg.BackendCACert != "" {
		caCertPEM, err := os.ReadFile(cfg.BackendCACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %v", err)
		}

		caCertPool = x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCertPEM) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
	}

	// Extract server name from URL (for SNI)
	serverName := ""
	url := cfg.BackendGRPCURL
	// Remove protocol prefix if present
	if strings.HasPrefix(url, "https://") {
		url = strings.TrimPrefix(url, "https://")
	} else if strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "http://")
	}
	// Extract hostname (before port)
	if idx := strings.Index(url, ":"); idx > 0 {
		serverName = url[:idx]
	} else {
		serverName = url
	}

	// Create TLS config with mTLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if caCertPool != nil {
		tlsConfig.RootCAs = caCertPool
	}

	if serverName != "" {
		tlsConfig.ServerName = serverName
	}

	return credentials.NewTLS(tlsConfig), nil
}


