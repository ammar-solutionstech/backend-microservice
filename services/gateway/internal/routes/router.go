package routes

import (
	"github.com/go-chi/chi"

	"backend/services/gateway/internal/clients"
	"backend/services/gateway/internal/config"
	"backend/services/gateway/internal/handlers"
	"backend/services/gateway/internal/middleware"
)

type Router struct {
	cfg             *config.Config
	authHandler     *handlers.AuthHandler
	helpDeskHandler *handlers.HelpDeskHandler
	legacyHandler   *handlers.LegacyHandler
}

func NewRouter(cfg *config.Config, authHandler *handlers.AuthHandler, helpDeskHandler *handlers.HelpDeskHandler, legacyHandler *handlers.LegacyHandler) *chi.Mux {
	router := chi.NewRouter()

	// Auth routes (public)
	router.Route("/api/auth", func(r chi.Router) {
		r.Post("/*", authHandler.Proxy)
		r.Get("/*", authHandler.Proxy)
		r.Put("/*", authHandler.Proxy)
		r.Delete("/*", authHandler.Proxy)
	})

	// Help Desk routes (protected)
	router.Route("/api/help-desk", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/", helpDeskHandler.ListTickets)
		r.Post("/", helpDeskHandler.CreateTicket)
		r.Get("/{id}", helpDeskHandler.GetTicket)
		// Add more routes as needed
	})

	// Legacy routes (temporary)
	router.Route("/api/inventory", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", legacyHandler.HandleInventory)
		r.Post("/*", legacyHandler.HandleInventory)
		r.Put("/*", legacyHandler.HandleInventory)
		r.Delete("/*", legacyHandler.HandleInventory)
	})

	router.Route("/api/geography", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", legacyHandler.HandleGeography)
		r.Post("/*", legacyHandler.HandleGeography)
		r.Put("/*", legacyHandler.HandleGeography)
		r.Delete("/*", legacyHandler.HandleGeography)
	})

	router.Route("/api/navigation", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", legacyHandler.HandleNavigation)
		r.Post("/*", legacyHandler.HandleNavigation)
		r.Put("/*", legacyHandler.HandleNavigation)
		r.Delete("/*", legacyHandler.HandleNavigation)
	})

	router.Route("/api/roles", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", legacyHandler.HandleRoles)
		r.Post("/*", legacyHandler.HandleRoles)
		r.Put("/*", legacyHandler.HandleRoles)
		r.Delete("/*", legacyHandler.HandleRoles)
	})

	router.Route("/api/permissions", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", legacyHandler.HandlePermissions)
		r.Post("/*", legacyHandler.HandlePermissions)
		r.Put("/*", legacyHandler.HandlePermissions)
		r.Delete("/*", legacyHandler.HandlePermissions)
	})

	return router
}
