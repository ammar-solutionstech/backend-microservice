package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"backend/services/gateway/internal/clients"
	"backend/services/gateway/internal/config"
)

type AuthHandler struct {
	cfg        *config.Config
	authClient *clients.AuthClient
	httpClient *http.Client
	authURL    string
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	// Use Auth Service HTTP URL from config
	authURL := cfg.AuthServiceHTTP
	if authURL == "" {
		// Fallback logic
		if cfg.AuthServiceGRPC == "localhost:9001" {
			authURL = "http://localhost:8001"
		} else {
			authURL = "http://auth-service:8001"
		}
	}

	return &AuthHandler{
		cfg:        cfg,
		authClient: clients.NewAuthClient(cfg.AuthServiceGRPC),
		httpClient: &http.Client{},
		authURL:    authURL,
	}
}

func (h *AuthHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	// Extract the path after /api/auth
	path := strings.TrimPrefix(r.URL.Path, "/api/auth")
	if path == "" {
		path = "/"
	}

	// Build the target URL
	targetURL := h.authURL + "/api/auth" + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Read request body
	var body io.Reader
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		body = bytes.NewBuffer(bodyBytes)
	}

	// Create new request to Auth Service
	req, err := http.NewRequest(r.Method, targetURL, body)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers (except Host)
	for key, values := range r.Header {
		if strings.ToLower(key) != "host" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	// Forward the request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		http.Error(w, "failed to forward request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	if resp.Body != nil {
		io.Copy(w, resp.Body)
	}
}
