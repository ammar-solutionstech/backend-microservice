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

	"backend/services/auth/internal/config"
	"backend/services/auth/internal/middleware"
	"backend/services/auth/internal/routes"
	"backend/services/auth/internal/services"
	"backend/services/auth/internal/utils"
	authpb "backend/services/auth/proto"
)

func main() {
	cfg := config.Load()

	// Initialize services
	authService := services.NewAuthService(cfg, cfg.DB)
	tokenService := services.NewTokenService(cfg, cfg.DB)

	// Start gRPC server
	go startGRPCServer(cfg, authService, tokenService)

	// Start REST server
	startRESTServer(cfg, authService, tokenService)
}

func startGRPCServer(cfg *config.Config, authService *services.AuthService, tokenService *services.TokenService) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	var opts []grpc.ServerOption

	// Configure mTLS if certificates are provided
	if cfg.GRPCMTLSCACert != "" && cfg.GRPCMTLSServerCert != "" && cfg.GRPCMTLSServerKey != "" {
		creds, err := utils.LoadGRPCServerCredentials(cfg.GRPCMTLSServerCert, cfg.GRPCMTLSServerKey, cfg.GRPCMTLSCACert)
		if err != nil {
			log.Fatalf("failed to load gRPC TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
		log.Println("gRPC server configured with mTLS")
	} else {
		log.Println("Warning: gRPC server starting without TLS (not recommended for production)")
	}

	s := grpc.NewServer(opts...)
	authpb.RegisterAuthServiceServer(s, services.NewAuthGRPCServer(cfg, authService, tokenService))
	authpb.RegisterUserServiceServer(s, services.NewUserGRPCServer(authService))

	log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func startRESTServer(cfg *config.Config, authService *services.AuthService, tokenService *services.TokenService) {
	authController := routes.NewAuthController(cfg, authService, tokenService)

	mux := chi.NewRouter()
	mux.Use(middleware.CORSMiddleware)

	// Public routes
	mux.Post("/api/auth/login", authController.Login)
	mux.Post("/api/auth/register", authController.Register)
	mux.Post("/api/auth/refresh", authController.RefreshToken)
	mux.Post("/api/auth/logout", authController.Logout)
	mux.Post("/api/auth/verify-email", authController.VerifyEmail)
	mux.Post("/api/auth/verify-phone", authController.VerifyPhone)
	mux.Post("/api/auth/resend-otp", authController.ResendOTP)
	mux.Post("/api/auth/forgot-password", authController.ForgotPassword)
	mux.Post("/api/auth/reset-password", authController.ResetPassword)

	// Protected routes
	mux.Route("/api/users", func(r chi.Router) {
		r.Use(middleware.JWTAuth(cfg, tokenService))
		r.Get("/", authController.GetUsers)
		r.Get("/{id}", authController.GetUser)
		r.Post("/", authController.CreateUser)
		r.Put("/{id}", authController.UpdateUser)
		r.Delete("/{id}", authController.DeleteUser)
	})

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

