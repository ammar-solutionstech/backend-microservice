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

	"backend/services/container-management/internal/config"
	"backend/services/container-management/internal/middleware"
	"backend/services/container-management/internal/routes"
	"backend/services/container-management/internal/services"
)

func main() {
	cfg := config.Load()

	// Initialize certificate service client
	certClient, err := services.NewCertificateClient(cfg)
	if err != nil {
		log.Fatalf("failed to create certificate client: %v", err)
	}

	// Initialize CSR validator
	csrValidator := services.NewCSRValidator(cfg)

	// Initialize Docker client
	dockerClient, err := services.NewDockerClient(cfg.DockerHost)
	if err != nil {
		log.Fatalf("failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	// Initialize container orchestrator
	orchestrator := services.NewContainerOrchestrator(dockerClient, cfg)

	// Initialize service manager
	serviceManager := services.NewServiceManager(cfg, cfg.DB, dockerClient)

	// Initialize services
	containerService := services.NewContainerService(cfg, cfg.DB, certClient, csrValidator)
	bootstrapService := services.NewBootstrapService(cfg, cfg.DB)
	orgService := services.NewOrganizationService(cfg, cfg.DB)
	clientContainerService := services.NewClientContainerService(cfg, cfg.DB, certClient, orgService, orchestrator, dockerClient)
	authService := services.NewAuthService()

	// Initialize controllers
	certController := routes.NewCertificateController(certClient, csrValidator)
	containerController := routes.NewContainerController(containerService)
	bootstrapController := routes.NewBootstrapController(bootstrapService, containerService, certClient, csrValidator)
	orgController := routes.NewOrganizationController(orgService)
	clientContainerController := routes.NewClientContainerController(clientContainerService)
	serviceController := routes.NewServiceController(serviceManager)

	// Setup router
	mux := chi.NewRouter()
	mux.Use(middleware.CORSMiddleware)

	// Health check (no mTLS required)
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Bootstrap routes (no mTLS required for registration)
	mux.Route("/api/v1/bootstrap", func(r chi.Router) {
		r.Post("/register", bootstrapController.Register)
		r.Get("/status", bootstrapController.Status)
	})

	// Organization routes
	mux.Route("/api/v1/organizations", func(r chi.Router) {
		r.Use(middleware.MTLSMiddleware)
		r.Post("/", orgController.CreateOrganization)
		r.Get("/", orgController.ListOrganizations)
		r.Get("/{id}", orgController.GetOrganization)
		r.Put("/{id}", orgController.UpdateOrganization)
		r.Post("/{id}/containers", clientContainerController.RegisterClientContainer)
		r.Get("/{id}/containers", clientContainerController.GetClientContainerByOrg)
	})

	// Protected routes (require mTLS)
	mux.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.MTLSMiddleware)
		r.Use(middleware.CertAuthMiddleware(authService, "", "")) // Apply cert auth middleware (no specific role/permission required)

		// Client container routes
		r.Route("/containers", func(cont chi.Router) {
			cont.Get("/{container_id}", clientContainerController.GetClientContainer)
			cont.Post("/{container_id}/start", clientContainerController.StartContainer)
			cont.Post("/{container_id}/stop", clientContainerController.StopContainer)
			cont.Post("/{container_id}/restart", clientContainerController.RestartContainer)
			cont.Put("/{container_id}/update", clientContainerController.UpdateContainer)
			cont.Delete("/{container_id}", clientContainerController.RemoveContainer)
			cont.Get("/{container_id}/status", clientContainerController.GetContainerStatus)

			// Service management routes
			cont.Route("/{container_id}/services", func(svc chi.Router) {
				svc.Post("/", serviceController.AddService)
				svc.Get("/", serviceController.ListServices)
				svc.Get("/{service_name}", serviceController.GetService)
				svc.Put("/{service_name}", serviceController.UpdateService)
				svc.Post("/{service_name}/enable", serviceController.EnableService)
				svc.Post("/{service_name}/disable", serviceController.DisableService)
				svc.Delete("/{service_name}", serviceController.RemoveService)
			})
		})

		// Certificate routes
		r.Route("/certificates", func(cert chi.Router) {
			cert.Post("/request", certController.RequestCertificate)
			cert.Get("/", certController.ListCertificates)
			cert.Get("/{serial}", certController.GetCertificate)
			cert.Post("/{serial}/revoke", certController.RevokeCertificate)
		})

		// Container routes (legacy)
		r.Route("/containers/{container_id}", func(cont chi.Router) {
			cont.Route("/certificates", func(cert chi.Router) {
				cert.Post("/request", containerController.RequestContainerCertificate)
				cert.Get("/", containerController.GetContainerCertificates)
				cert.Post("/{serial}/revoke", containerController.RevokeContainerCertificate)
			})
			cont.Route("/applications/{app_name}/certificates", func(app chi.Router) {
				app.Post("/request", containerController.RequestApplicationCertificate)
			})
		})
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Configure TLS if mTLS is enabled
	if cfg.MTLSTLSConfig != nil {
		srv.TLSConfig = cfg.MTLSTLSConfig
	}

	// Start server
	go func() {
		log.Printf("Container Management service listening on :%s", cfg.Port)
		var err error
		if cfg.MTLSTLSConfig != nil {
			// Start HTTPS server with mTLS
			err = srv.ListenAndServeTLS(cfg.MTLSServerCert, cfg.MTLSServerKey)
		} else {
			// Fallback to HTTP (not recommended for production)
			log.Println("Warning: Starting without TLS. mTLS is not configured.")
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt to gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down container management service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("container management service stopped cleanly")
}
