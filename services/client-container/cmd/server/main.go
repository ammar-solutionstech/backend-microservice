package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"backend/services/client-container/internal/config"
	"backend/services/client-container/internal/middleware"
	"backend/services/client-container/internal/routes"
	"backend/services/client-container/internal/services"
)

func main() {
	cfg := config.Load()

	// Initialize service registry
	serviceRegistry := services.NewServiceRegistry()

	// Initialize certificate client
	certClient, err := services.NewCertificateClient(cfg)
	if err != nil {
		log.Fatalf("failed to create certificate client: %v", err)
	}

	// Initialize verification service
	verificationService := services.NewVerificationService(cfg, cfg.DB)

	// Initialize notification client
	notificationClient, err := services.NewNotificationClient(cfg.NotificationServiceGRPC)
	if err != nil {
		log.Printf("Warning: Failed to create notification client: %v", err)
		log.Println("Verification code notifications may not work until notification service is available")
	}

	// Initialize device service
	//deviceService :=
	services.NewDeviceService(cfg, cfg.DB, certClient, verificationService, notificationClient)

	// Initialize container service
	//containerService :=
	services.NewContainerService(cfg, cfg.DB, certClient)

	// Initialize controllers
	serviceController := routes.NewServiceController(serviceRegistry)
	// DeviceController will be created when device routes are implemented
	// For now, we'll add device routes directly or create a simple controller

	// Setup HTTP router
	mux := chi.NewRouter()
	mux.Use(middleware.MTLSMiddleware)

	// Health check (no mTLS required)
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Internal API routes (mTLS protected - for container-management service)
	mux.Route("/api/v1/internal", func(r chi.Router) {
		// TODO: Add mTLS middleware here
		r.Use(middleware.MTLSMiddleware)

		r.Route("/services", func(svc chi.Router) {
			svc.Post("/", serviceController.AddService)
			svc.Get("/", serviceController.ListServices)
			svc.Get("/{name}", serviceController.GetService)
			svc.Put("/{name}", serviceController.UpdateService)
			svc.Post("/{name}/start", serviceController.StartService)
			svc.Post("/{name}/stop", serviceController.StopService)
			svc.Delete("/{name}", serviceController.RemoveService)
		})
	})

	// Device registration routes (mTLS protected - for agents)
	mux.Route("/api/v1/devices", func(r chi.Router) {
		// TODO: Add mTLS middleware here
		r.Use(middleware.MTLSMiddleware)

		r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement device registration handler
			// This should call deviceService.RegisterDevice
			writeError(w, http.StatusNotImplemented, "device registration not yet implemented")
		})

		r.Post("/{id}/verify", func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement device verification handler
			// This should call deviceService.VerifyAndIssueCertificateWithCSR
			writeError(w, http.StatusNotImplemented, "device verification not yet implemented")
		})

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement list devices handler
			writeError(w, http.StatusNotImplemented, "list devices not yet implemented")
		})
	})

	// Agent download route (mTLS protected)
	mux.Route("/api/v1/agents", func(r chi.Router) {
		// TODO: Add mTLS middleware here
		r.Use(middleware.MTLSMiddleware)

		r.Get("/download", func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement agent download handler
			writeError(w, http.StatusNotImplemented, "agent download not yet implemented")
		})
	})

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Configure TLS if container certificate is available
	if cfg.ContainerCert != nil {
		// TODO: Configure TLS for mTLS server
		httpServer.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{*cfg.ContainerCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    cfg.AgentCACertPool,
		}
		log.Println("Container certificate loaded, but TLS configuration not yet implemented")
	}

	// Start HTTP server
	go func() {
		log.Printf("Client Container HTTP service listening on :%s", cfg.Port)
		var err error
		if cfg.ContainerCert != nil {
			// TODO: Start HTTPS server with mTLS
			err = httpServer.ListenAndServeTLS(cfg.ContainerCertPath, cfg.ContainerKeyPath)
			log.Println("Warning: Starting HTTP server without TLS. mTLS is not yet configured.")
			//log.Println("Warning: Starting HTTP server without TLS. mTLS is not yet configured.")
			// err = httpServer.ListenAndServe()
		} else {
			log.Println("Warning: Starting HTTP server without TLS. Container certificate not configured.")
			err = httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC server (for agent communication)
	var grpcServer *grpc.Server
	var grpcListener net.Listener
	if cfg.GRPCPort != "" {
		grpcAddr := ":" + cfg.GRPCPort
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Fatalf("failed to listen on gRPC port %s: %v", cfg.GRPCPort, err)
		}

		// TODO: Configure gRPC with mTLS
		creds, err := credentials.NewServerTLSFromFile(cfg.ContainerCertPath, cfg.ContainerKeyPath)
		if err != nil {
			log.Fatalf("failed to load gRPC TLS credentials: %v", err)
		}
		grpcServer = grpc.NewServer(grpc.Creds(creds))

		// For now, start without TLS (not recommended for production)
		// grpcServer = grpc.NewServer()
		// grpcListener = lis

		// TODO: Register gRPC services when proto definitions are available
		// agentpb.RegisterAgentServiceServer(grpcServer, &agentService{})

		go func() {
			log.Printf("Client Container gRPC service listening on :%s", cfg.GRPCPort)
			if err := grpcServer.Serve(lis); err != nil {
				log.Fatalf("gRPC server error: %v", err)
			}
		}()
	}

	// Wait for interrupt to gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down client container service...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown error: %v", err)
	}
	log.Println("HTTP server stopped cleanly")

	// Shutdown gRPC server
	if grpcServer != nil {
		log.Println("stopping gRPC server...")
		grpcServer.GracefulStop()
		if grpcListener != nil {
			grpcListener.Close()
		}
		log.Println("gRPC server stopped cleanly")
	}

	// Close notification client connection
	if notificationClient != nil {
		// TODO: Add Close method to NotificationClient if needed
		// notificationClient.Close()
	}

	log.Println("client container service stopped cleanly")
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
