package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
)

// GenerateKeyAndCSR generates a private key and CSR for a container
func GenerateKeyAndCSR(containerID, organization, country string) (privateKeyPEM []byte, csrPEM []byte, err error) {
	// Generate private key (RSA 2048-bit)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create CSR template
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         fmt.Sprintf("client-container-%s", containerID),
			Organization:       []string{organization},
			OrganizationalUnit: []string{"Client Container"},
			Country:            []string{country},
		},
		DNSNames: []string{
			fmt.Sprintf("client-container-%s", containerID),
			//containerID,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	log.Printf("CSR Template: %+v", csrTemplate)

	// Create CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CSR: %v", err)
	}

	// Verify CSR contains all fields by parsing it back
	parsedCSR, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify CSR: %v", err)
	}

	// Log CSR subject for debugging
	log.Printf("Generated CSR Subject - CN: %s, O: %v, OU: %v, C: %v",
		parsedCSR.Subject.CommonName,
		parsedCSR.Subject.Organization,
		parsedCSR.Subject.OrganizationalUnit,
		parsedCSR.Subject.Country,
	)

	// Encode private key to PEM
	privateKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode CSR to PEM
	csrPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	return privateKeyPEM, csrPEM, nil
}
