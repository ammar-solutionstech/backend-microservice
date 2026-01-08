package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"backend/services/gateway/internal/config"
)

type NavigationHandler struct {
	cfg        *config.Config
	httpClient *http.Client
	serviceURL string
}

func NewNavigationHandler(cfg *config.Config) *NavigationHandler {
	serviceURL := cfg.NavigationServiceHTTP
	if serviceURL == "" {
		if strings.Contains(cfg.NavigationServiceGRPC, "navigation-service") {
			serviceURL = "http://navigation-service:8009"
		} else {
			serviceURL = "http://localhost:8009"
		}
	}

	return &NavigationHandler{
		cfg:        cfg,
		httpClient: &http.Client{},
		serviceURL: serviceURL,
	}
}

func (h *NavigationHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	if path == "" {
		path = "/"
	}

	targetURL := h.serviceURL + "/api" + path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	var body io.Reader
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}
		body = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(r.Method, targetURL, body)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		if strings.ToLower(key) != "host" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		http.Error(w, "failed to forward request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	if resp.Body != nil {
		io.Copy(w, resp.Body)
	}
}

