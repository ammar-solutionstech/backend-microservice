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

	"backend/services/notification/internal/config"
	"backend/services/notification/internal/middleware"
	"backend/services/notification/internal/providers"
	"backend/services/notification/internal/services"
	"backend/services/notification/internal/utils"
	"backend/services/notification/internal/workers"
	notificationpb "backend/services/notification/proto"
)

func main() {
	cfg := config.Load()

	// Initialize providers
	smtpProvider := providers.NewSMTPProvider(cfg)
	smsProvider := providers.NewMockSMSProvider()

	// Initialize services
	templateService := services.NewTemplateService(cfg.DB)
	notificationService := services.NewNotificationService(cfg, cfg.DB, smtpProvider, smsProvider, templateService)

	// Start RabbitMQ worker
	go workers.StartNotificationWorker(cfg, notificationService)

	// Start gRPC server
	go startGRPCServer(cfg, notificationService, templateService)

	// Start REST server
	startRESTServer(cfg, notificationService, templateService)
}

func startGRPCServer(cfg *config.Config, notificationService *services.NotificationService, templateService *services.TemplateService) {
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
	notificationpb.RegisterNotificationServiceServer(s, services.NewNotificationGRPCServer(notificationService, templateService))

	log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func startRESTServer(cfg *config.Config, notificationService *services.NotificationService, templateService *services.TemplateService) {
	controller := services.NewNotificationController(notificationService, templateService)

	mux := chi.NewRouter()
	mux.Use(middleware.CORSMiddleware)

	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Notification routes (internal use, protected by Gateway)
	mux.Post("/api/notifications/email", controller.SendEmail)
	mux.Post("/api/notifications/sms", controller.SendSMS)
	mux.Get("/api/templates/{name}", controller.GetTemplate)
	mux.Post("/api/templates", controller.CreateTemplate)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("REST server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt to gracefully shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped cleanly")
}

