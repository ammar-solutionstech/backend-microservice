package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"backend/services/gateway/internal/config"
)

type InventoryHandler struct {
	cfg        *config.Config
	httpClient *http.Client
	serviceURL string
}

func NewInventoryHandler(cfg *config.Config) *InventoryHandler {
	serviceURL := cfg.InventoryServiceHTTP
	if serviceURL == "" {
		if strings.Contains(cfg.InventoryServiceGRPC, "inventory-service") {
			serviceURL = "http://inventory-service:8007"
		} else {
			serviceURL = "http://localhost:8007"
		}
	}

	return &InventoryHandler{
		cfg:        cfg,
		httpClient: &http.Client{},
		serviceURL: serviceURL,
	}
}

func (h *InventoryHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	// Extract the path after /api/
	path := strings.TrimPrefix(r.URL.Path, "/api")
	if path == "" {
		path = "/"
	}

	// Build the target URL
	targetURL := h.serviceURL + "/api" + path
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

	// Create new request
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

