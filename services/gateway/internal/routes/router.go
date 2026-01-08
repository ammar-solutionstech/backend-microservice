package routes

import (
	"github.com/go-chi/chi"

	"backend/services/gateway/internal/clients"
	"backend/services/gateway/internal/config"
	"backend/services/gateway/internal/handlers"
	"backend/services/gateway/internal/middleware"
)

type Router struct {
	cfg                *config.Config
	authHandler        *handlers.AuthHandler
	helpDeskHandler    *handlers.HelpDeskHandler
	inventoryHandler   *handlers.InventoryHandler
	geographyHandler   *handlers.GeographyHandler
	navigationHandler  *handlers.NavigationHandler
	legacyHandler      *handlers.LegacyHandler
}

func NewRouter(cfg *config.Config, authHandler *handlers.AuthHandler, helpDeskHandler *handlers.HelpDeskHandler, inventoryHandler *handlers.InventoryHandler, geographyHandler *handlers.GeographyHandler, navigationHandler *handlers.NavigationHandler, legacyHandler *handlers.LegacyHandler) *chi.Mux {
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

	// Inventory routes (protected)
	router.Route("/api/brands", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/models", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/equipment-types", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/operating-systems", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/software-categories", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/software", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/equipment", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/documents", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})
	router.Route("/api/maintenance", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", inventoryHandler.Proxy)
		r.Post("/*", inventoryHandler.Proxy)
		r.Put("/*", inventoryHandler.Proxy)
		r.Delete("/*", inventoryHandler.Proxy)
	})

	// Geography routes (protected)
	router.Route("/api/countries", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", geographyHandler.Proxy)
		r.Post("/*", geographyHandler.Proxy)
		r.Put("/*", geographyHandler.Proxy)
		r.Delete("/*", geographyHandler.Proxy)
	})
	router.Route("/api/cities", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", geographyHandler.Proxy)
		r.Post("/*", geographyHandler.Proxy)
		r.Put("/*", geographyHandler.Proxy)
		r.Delete("/*", geographyHandler.Proxy)
	})
	router.Route("/api/locations", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", geographyHandler.Proxy)
		r.Post("/*", geographyHandler.Proxy)
		r.Put("/*", geographyHandler.Proxy)
		r.Delete("/*", geographyHandler.Proxy)
	})
	router.Route("/api/contacts", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", geographyHandler.Proxy)
		r.Post("/*", geographyHandler.Proxy)
		r.Put("/*", geographyHandler.Proxy)
		r.Delete("/*", geographyHandler.Proxy)
	})

	// Navigation routes (protected)
	router.Route("/api/menus", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", navigationHandler.Proxy)
		r.Post("/*", navigationHandler.Proxy)
		r.Put("/*", navigationHandler.Proxy)
		r.Delete("/*", navigationHandler.Proxy)
	})

	// Roles and Permissions routes (protected - now handled by Auth Service)
	router.Route("/api/roles", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", authHandler.Proxy)
		r.Post("/*", authHandler.Proxy)
		r.Put("/*", authHandler.Proxy)
		r.Delete("/*", authHandler.Proxy)
	})

	router.Route("/api/permissions", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
		r.Get("/*", authHandler.Proxy)
		r.Post("/*", authHandler.Proxy)
		r.Put("/*", authHandler.Proxy)
		r.Delete("/*", authHandler.Proxy)
	})

	return router
}
