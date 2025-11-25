package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"backend/services/auth/internal/config"
	"backend/services/auth/internal/services"
)

type contextKey string

const (
	ContextUserID contextKey = "userID"
)

func JWTAuth(cfg *config.Config, tokenService *services.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "missing or invalid authorization header",
				})
				return
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			claims, err := tokenService.ValidateAccessToken(tokenStr)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "invalid token",
				})
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
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

func GetUserIDFromRequest(r *http.Request) int {
	return GetUserIDFromContext(r.Context())
}

func GetIDParam(r *http.Request, param string) (int, error) {
	idStr := chi.URLParam(r, param)
	return strconv.Atoi(idStr)
}

