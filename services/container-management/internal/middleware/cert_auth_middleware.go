package middleware

import (
	"net/http"

	"backend/services/container-management/internal/services"
)

// CertAuthMiddleware creates middleware that checks for required role or permission
func CertAuthMiddleware(authService *services.AuthService, requiredRole string, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cert := GetCertificateFromContext(r)
			if cert == nil {
				http.Error(w, "certificate not found in context", http.StatusUnauthorized)
				return
			}

			// Check role if required
			if requiredRole != "" {
				authorized, err := authService.Authorize(cert, requiredRole)
				if err != nil {
					http.Error(w, "authorization check failed", http.StatusInternalServerError)
					return
				}
				if !authorized {
					http.Error(w, "insufficient permissions: role required", http.StatusForbidden)
					return
				}
			}

			// Check permission if required
			if requiredPermission != "" {
				authorized, err := authService.AuthorizePermission(cert, requiredPermission)
				if err != nil {
					http.Error(w, "authorization check failed", http.StatusInternalServerError)
					return
				}
				if !authorized {
					http.Error(w, "insufficient permissions: permission required", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

