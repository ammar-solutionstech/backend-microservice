package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"backend/services/certificate/internal/config"
	"backend/services/certificate/internal/middleware"
	"backend/services/certificate/internal/routes"
	"backend/services/certificate/internal/services"
	"backend/services/certificate/internal/utils"
)

func main() {
	cfg := config.Load()

	// Initialize step-ca client
	stepCAClient, err := services.NewStepCAClient(cfg)
	if err != nil {
		log.Fatalf("failed to create step-ca client: %v", err)
	}

	// Check step-ca connectivity
	if err := stepCAClient.CheckConnectivity(); err != nil {
		log.Printf("Warning: step-ca connectivity check failed: %v", err)
		log.Println("Service will start but certificate operations may fail until step-ca is available")
	}

	// Initialize services
	caService := services.NewCAService(cfg, cfg.DB, stepCAClient)
	raService := services.NewRAService(cfg, cfg.DB, caService)
	vaService := services.NewVAService(cfg, cfg.DB, stepCAClient)

	// Start gRPC server
	go startGRPCServer(cfg, caService, raService, vaService)

	// Initialize controllers
	raController := routes.NewRAController(raService)
	caController := routes.NewCAController(caService)
	vaController := routes.NewVAController(vaService)

	// Setup router
	mux := chi.NewRouter()
	mux.Use(middleware.CORSMiddleware)

	// Health check
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// RA routes
	mux.Route("/api/csr", func(r chi.Router) {
		r.Post("/", raController.SubmitCSR)
		r.Get("/", raController.ListCSRs)
		r.Get("/{id}", raController.GetCSR)
		r.Post("/{id}/approve", raController.ApproveCSR)
		r.Post("/{id}/reject", raController.RejectCSR)
	})

	// CA routes
	mux.Route("/api/certificates", func(r chi.Router) {
		r.Post("/", caController.IssueCertificate)
		r.Get("/", caController.ListCertificates)
		r.Get("/{serial}", caController.GetCertificate)
		r.Post("/{serial}/revoke", caController.RevokeCertificate)
		r.Post("/renew", caController.RenewCertificate)
	})

	// VA routes
	mux.Route("/api", func(r chi.Router) {
		r.Get("/ocsp/{serial}", vaController.GetOCSPStatus)
		r.Get("/crl", vaController.GetCRL)
		r.Get("/validate/{serial}", vaController.ValidateCertificate)
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Certificate service listening on :%s", cfg.Port)
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
	log.Println("shutting down certificate service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("certificate service stopped cleanly")
}

func startGRPCServer(cfg *config.Config, caService *services.CAService, raService *services.RAService, vaService *services.VAService) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	var opts []grpc.ServerOption

	// Use mTLS if certificates are configured
	if cfg.GRPCMTLSCACert != "" && cfg.GRPCMTLSServerCert != "" && cfg.GRPCMTLSServerKey != "" {
		creds, err := utils.LoadGRPCServerCredentials(
			cfg.GRPCMTLSServerCert,
			cfg.GRPCMTLSServerKey,
			cfg.GRPCMTLSCACert,
		)
		if err != nil {
			log.Printf("Warning: failed to load gRPC server TLS credentials: %v", err)
			log.Println("Falling back to insecure connection")
			opts = append(opts, grpc.Creds(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.Creds(creds))
			log.Println("gRPC server configured with mTLS")
		}
	} else {
		log.Println("Warning: gRPC mTLS not configured, using insecure connection")
		opts = append(opts, grpc.Creds(insecure.NewCredentials()))
	}

	s := grpc.NewServer(opts...)
	// TODO: Register gRPC services when proto files are available
	// certificatepb.RegisterCertificateServiceServer(s, services.NewCertificateGRPCServer(caService, raService, vaService))

	log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
