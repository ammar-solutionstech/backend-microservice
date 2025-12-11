package middleware

import (
	"context"
	"crypto/x509"
	"net/http"
)

type contextKey string

const certContextKey contextKey = "client_certificate"

// MTLSMiddleware validates client certificates and extracts them to context
func MTLSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client certificate from TLS connection
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			http.Error(w, "client certificate required", http.StatusUnauthorized)
			return
		}

		// Get the first (leaf) certificate
		cert := r.TLS.PeerCertificates[0]

		// Verify certificate is not expired
		if cert.NotAfter.Before(cert.NotBefore) {
			http.Error(w, "invalid certificate: expiration before not before", http.StatusUnauthorized)
			return
		}

		// Store certificate in context
		ctx := context.WithValue(r.Context(), certContextKey, cert)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCertificateFromContext retrieves the client certificate from request context
func GetCertificateFromContext(r *http.Request) *x509.Certificate {
	if cert, ok := r.Context().Value(certContextKey).(*x509.Certificate); ok {
		return cert
	}
	return nil
}

