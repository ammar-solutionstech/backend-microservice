package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"

	"backend/services/gateway/internal/config"
	"backend/services/gateway/internal/handlers"
	"backend/services/gateway/internal/middleware"
	"backend/services/gateway/internal/routes"
)

func main() {
	cfg := config.Load()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg)
	helpDeskHandler := handlers.NewHelpDeskHandler(cfg)
	legacyHandler := handlers.NewLegacyHandler(cfg)

	// Initialize routes
	router := routes.NewRouter(cfg, authHandler, helpDeskHandler, legacyHandler)

	// Setup middleware
	mux := chi.NewRouter()
	mux.Use(middleware.CORSMiddleware)
	mux.Use(middleware.ErrorHandler)
	mux.Mount("/", router)

	// Health check
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Gateway server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt to gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("gateway stopped cleanly")
}

