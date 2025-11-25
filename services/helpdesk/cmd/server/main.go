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

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/middleware"
	"backend/services/helpdesk/internal/routes"
	"backend/services/helpdesk/internal/services"
	helpdeskpb "backend/services/helpdesk/proto"
)

func main() {
	cfg := config.Load()

	// Initialize services
	ticketService := services.NewTicketService(cfg, cfg.DB)
	typeService := services.NewHelpDeskTypeService(cfg, cfg.DB)
	teamService := services.NewTeamService(cfg, cfg.DB)

	// Start gRPC server
	go startGRPCServer(cfg, ticketService, typeService)

	// Start REST server
	startRESTServer(cfg, ticketService, typeService, teamService)
}

func startGRPCServer(cfg *config.Config, ticketService *services.TicketService, typeService *services.HelpDeskTypeService) {
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	s := grpc.NewServer()
	helpdeskpb.RegisterHelpDeskServiceServer(s, services.NewHelpDeskGRPCServer(ticketService))
	helpdeskpb.RegisterTicketServiceServer(s, services.NewTicketGRPCServer(ticketService))
	helpdeskpb.RegisterHelpDeskTypeServiceServer(s, services.NewHelpDeskTypeGRPCServer(typeService))

	log.Printf("gRPC server listening on :%s", cfg.GRPCPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func startRESTServer(cfg *config.Config, ticketService *services.TicketService, typeService *services.HelpDeskTypeService, teamService *services.TeamService) {
	ticketController := routes.NewTicketController(cfg, ticketService)
	typeController := routes.NewHelpDeskTypeController(cfg, typeService)
	teamController := routes.NewTeamController(cfg, teamService)

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
