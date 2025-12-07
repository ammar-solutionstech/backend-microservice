package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ParseCSR parses a PEM-encoded CSR and returns the certificate request
func ParseCSR(csrPEM string) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE REQUEST" && block.Type != "NEW CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("invalid PEM type: %s", block.Type)
	}

	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate request: %v", err)
	}

	return csr, nil
}

// ExtractCSRFields extracts key fields from a CSR
func ExtractCSRFields(csr *x509.CertificateRequest) map[string]interface{} {
	fields := make(map[string]interface{})

	// Subject fields
	if len(csr.Subject.Organization) > 0 {
		fields["organization"] = csr.Subject.Organization[0]
	}
	if len(csr.Subject.OrganizationalUnit) > 0 {
		fields["organizational_unit"] = csr.Subject.OrganizationalUnit
	}
	if csr.Subject.CommonName != "" {
		fields["common_name"] = csr.Subject.CommonName
	}
	if len(csr.Subject.Country) > 0 {
		fields["country"] = csr.Subject.Country[0]
	}
	if len(csr.Subject.Province) > 0 {
		fields["province"] = csr.Subject.Province
	}
	if len(csr.Subject.Locality) > 0 {
		fields["locality"] = csr.Subject.Locality
	}

	// SANs
	if len(csr.DNSNames) > 0 {
		fields["dns_names"] = csr.DNSNames
	}
	if len(csr.IPAddresses) > 0 {
		fields["ip_addresses"] = csr.IPAddresses
	}
	if len(csr.EmailAddresses) > 0 {
		fields["email_addresses"] = csr.EmailAddresses
	}

	// Key size
	if csr.PublicKey != nil {
		fields["public_key_algorithm"] = csr.PublicKeyAlgorithm.String()
	}

	return fields
}

