package utils

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
)

// Custom OIDs for certificate extensions
var (
	OIDRoles       = []int{1, 3, 6, 1, 4, 1, 12345, 1, 1} // 1.3.6.1.4.1.12345.1.1
	OIDPermissions = []int{1, 3, 6, 1, 4, 1, 12345, 1, 2} // 1.3.6.1.4.1.12345.1.2
)

// ExtractRoles extracts roles from custom certificate extension
func ExtractRoles(cert *x509.Certificate) ([]string, error) {
	return extractExtensionValue(cert, OIDRoles)
}

// ExtractPermissions extracts permissions from custom certificate extension
func ExtractPermissions(cert *x509.Certificate) ([]string, error) {
	return extractExtensionValue(cert, OIDPermissions)
}

// extractExtensionValue extracts a JSON array of strings from a custom extension
func extractExtensionValue(cert *x509.Certificate, oid []int) ([]string, error) {
	oidString := oidToString(oid)

	for _, ext := range cert.ExtraExtensions {
		if ext.Id.String() == oidString {
			var values []string
			if err := json.Unmarshal(ext.Value, &values); err != nil {
				return nil, fmt.Errorf("failed to parse extension value: %v", err)
			}
			return values, nil
		}
	}

	return []string{}, nil
}

// oidToString converts an OID array to string format
func oidToString(oid []int) string {
	if len(oid) == 0 {
		return ""
	}

	result := fmt.Sprintf("%d", oid[0])
	for i := 1; i < len(oid); i++ {
		result += fmt.Sprintf(".%d", oid[i])
	}
	return result
}

// HasRole checks if certificate has a specific role
func HasRole(cert *x509.Certificate, role string) (bool, error) {
	roles, err := ExtractRoles(cert)
	if err != nil {
		return false, err
	}

	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}

	return false, nil
}

// HasPermission checks if certificate has a specific permission
func HasPermission(cert *x509.Certificate, permission string) (bool, error) {
	permissions, err := ExtractPermissions(cert)
	if err != nil {
		return false, err
	}

	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

