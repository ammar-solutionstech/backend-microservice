package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"backend/services/gateway/internal/clients"
)

type contextKey string

const (
	ContextUserID contextKey = "userID"
	ContextRoles  contextKey = "roles"
)

func AuthMiddleware(authClient *clients.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			resp, err := authClient.ValidateToken(tokenStr)
			if err != nil || !resp.Valid {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, int(resp.UserId))
			ctx = context.WithValue(ctx, ContextRoles, resp.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) int {
	if userID, ok := ctx.Value(ContextUserID).(int); ok {
		return userID
	}
	return 0
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
