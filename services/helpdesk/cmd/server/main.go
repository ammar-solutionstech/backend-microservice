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

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/middleware"
	"backend/services/helpdesk/internal/models"
	"backend/services/helpdesk/internal/routes"
	"backend/services/helpdesk/internal/services"
	"backend/services/helpdesk/internal/utils"
	helpdeskpb "backend/services/helpdesk/proto"
)

func main() {
	cfg := config.Load()

	// Initialize services
	ticketService := services.NewTicketService(cfg, cfg.DB)
	typeService := services.NewHelpDeskTypeService(cfg, cfg.DB)
	teamService := services.NewTeamService(cfg, cfg.DB)
	ratingService := services.NewRatingService(cfg, cfg.DB)
	transactionService := services.NewTransactionService(cfg, cfg.DB)
	transactionTypeService := services.NewGenericService[models.TransactionType](cfg.DB)

	// Start gRPC server
	go startGRPCServer(cfg, ticketService, typeService)

	// Start REST server
	startRESTServer(cfg, ticketService, typeService, teamService, ratingService, transactionService, transactionTypeService)
}

func startGRPCServer(cfg *config.Config, ticketService *services.TicketService, typeService *services.HelpDeskTypeService) {
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
	helpdeskpb.RegisterHelpDeskServiceServer(s, services.NewHelpDeskGRPCServer(ticketService))
	helpdeskpb.RegisterTicketServiceServer(s, services.NewTicketGRPCServer(ticketService))
	helpdeskpb.RegisterHelpDeskTypeServiceServer(s, services.NewHelpDeskTypeGRPCServer(typeService))

	log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func startRESTServer(cfg *config.Config, ticketService *services.TicketService, typeService *services.HelpDeskTypeService, teamService *services.TeamService, ratingService *services.RatingService, transactionService *services.TransactionService, transactionTypeService *services.GenericService[models.TransactionType]) {
	ticketController := routes.NewTicketController(cfg, ticketService)
	typeController := routes.NewHelpDeskTypeController(cfg, typeService)
	teamController := routes.NewTeamController(cfg, teamService)
	ratingController := routes.NewRatingController(cfg, ratingService)
	transactionController := routes.NewTransactionController(cfg, transactionService, ticketService)
	participantController := routes.NewParticipantController(cfg, ticketService)
	transactionTypeController := routes.NewTransactionTypeController(cfg, transactionTypeService)

	mux := chi.NewRouter()
	mux.Use(middleware.CORSMiddleware)

	// Public health check
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// Help Desk routes (will be protected by Gateway)
	mux.Route("/api/help-desk", func(r chi.Router) {
		r.Get("/", ticketController.ListTickets)
		r.Post("/", ticketController.CreateTicket)
		r.Get("/{id}", ticketController.GetTicket)
		r.Put("/{id}", ticketController.UpdateTicket)
		r.Delete("/{id}", ticketController.DeleteTicket)
		r.Post("/{id}/comments", ticketController.AddComment)
		r.Get("/{id}/comments", ticketController.GetComments)
		r.Post("/{id}/attachments", ticketController.AddAttachment)
		r.Get("/{id}/attachments", ticketController.GetAttachments)
		r.Post("/{id}/assign", ticketController.AssignUser)
		r.Put("/{id}/status", ticketController.UpdateStatus)

		// Participants routes
		r.Get("/{helpDeskId}/participants", participantController.GetParticipants)
		r.Post("/{helpDeskId}/participants", participantController.AddParticipant)
		r.Delete("/{helpDeskId}/participants/{userId}", participantController.RemoveParticipant)

		// Ratings routes
		r.Get("/ratings", ratingController.ListRatings)
		r.Post("/ratings", ratingController.CreateRating)
		r.Get("/ratings/{id}", ratingController.GetRating)
		r.Put("/ratings/{id}", ratingController.UpdateRating)
		r.Delete("/ratings/{id}", ratingController.DeleteRating)

		// Transactions routes
		r.Get("/transactions", transactionController.ListTransactions)
		r.Post("/transactions", transactionController.CreateTransaction)
		r.Get("/transactions/{id}", transactionController.GetTransaction)
		r.Put("/transactions/{id}", transactionController.UpdateTransaction)
		r.Delete("/transactions/{id}", transactionController.DeleteTransaction)

		// Transaction users routes
		r.Get("/transactions/{transactionId}/users", transactionController.GetTransactionUsers)
		r.Post("/transactions/{transactionId}/users", transactionController.AddTransactionUser)
		r.Delete("/transactions/{transactionId}/users/{userId}", transactionController.RemoveTransactionUser)
	})

	mux.Route("/api/help-desk-types", func(r chi.Router) {
		r.Get("/", typeController.ListTypes)
		r.Post("/", typeController.CreateType)
		r.Get("/{id}", typeController.GetType)
		r.Put("/{id}", typeController.UpdateType)
		r.Delete("/{id}", typeController.DeleteType)
	})

	mux.Route("/api/help-desk/teams", func(r chi.Router) {
		r.Get("/", teamController.ListTeams)
		r.Post("/", teamController.CreateTeam)
		r.Get("/{id}", teamController.GetTeam)
		r.Put("/{id}", teamController.UpdateTeam)
		r.Delete("/{id}", teamController.DeleteTeam)
		r.Get("/{id}/members", teamController.GetTeamMembers)
		r.Post("/{id}/members", teamController.AddTeamMember)
		r.Delete("/{id}/members/{userId}", teamController.RemoveTeamMember)
	})

	mux.Route("/api/help-desk/transaction-types", func(r chi.Router) {
		r.Get("/", transactionTypeController.ListTransactionTypes)
		r.Post("/", transactionTypeController.CreateTransactionType)
		r.Get("/{id}", transactionTypeController.GetTransactionType)
		r.Put("/{id}", transactionTypeController.UpdateTransactionType)
		r.Delete("/{id}", transactionTypeController.DeleteTransactionType)
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
