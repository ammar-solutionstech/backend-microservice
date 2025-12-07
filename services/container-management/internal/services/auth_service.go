package services

import (
	"crypto/x509"

	"backend/services/container-management/internal/utils"
)

// AuthService handles certificate-based authorization
type AuthService struct {
}

// NewAuthService creates a new auth service
func NewAuthService() *AuthService {
	return &AuthService{}
}

// ExtractRoles extracts roles from certificate extensions
func (s *AuthService) ExtractRoles(cert *x509.Certificate) ([]string, error) {
	return utils.ExtractRoles(cert)
}

// ExtractPermissions extracts permissions from certificate extensions
func (s *AuthService) ExtractPermissions(cert *x509.Certificate) ([]string, error) {
	return utils.ExtractPermissions(cert)
}

// Authorize checks if certificate has required role
func (s *AuthService) Authorize(cert *x509.Certificate, requiredRole string) (bool, error) {
	return utils.HasRole(cert, requiredRole)
}

// AuthorizePermission checks if certificate has required permission
func (s *AuthService) AuthorizePermission(cert *x509.Certificate, permission string) (bool, error) {
	return utils.HasPermission(cert, permission)
}

