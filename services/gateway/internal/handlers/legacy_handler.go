package handlers

import (
	"backend/services/gateway/internal/config"
)

type LegacyHandler struct {
	cfg *config.Config
}

func NewLegacyHandler(cfg *config.Config) *LegacyHandler {
	return &LegacyHandler{
		cfg: cfg,
	}
}

// LegacyHandler is kept for backward compatibility but routes are now handled by service handlers
