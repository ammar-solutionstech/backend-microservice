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

	"backend/services/geography/internal/config"
	"backend/services/geography/internal/middleware"
	"backend/services/geography/internal/routes"
	"backend/services/geography/internal/utils"
)

func main() {
	cfg := config.Load()

	// Start gRPC server
	go startGRPCServer(cfg)

	// Start REST server
	startRESTServer(cfg)
}

func startGRPCServer(cfg *config.Config) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	var opts []grpc.ServerOption

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

	log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func startRESTServer(cfg *config.Config) {
	mux := routes.RegisterRoutes(cfg, cfg.DB)
	mux.Use(middleware.CORSMiddleware)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("REST server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped cleanly")
}

