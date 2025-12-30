package clients

import (
	"context"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"backend/services/gateway/internal/config"
	"backend/services/gateway/internal/utils"
	helpdeskpb "backend/services/helpdesk/proto"
)

type HelpDeskClient struct {
	ticketClient helpdeskpb.HelpDeskServiceClient
	typeClient   helpdeskpb.HelpDeskTypeServiceClient
}

func NewHelpDeskClient(cfg *config.Config) *HelpDeskClient {
	var opts []grpc.DialOption

	// Use mTLS if certificates are configured
	if cfg.HelpdeskServiceGRPCMTLSCA != "" && cfg.HelpdeskServiceGRPCMTLSClientCert != "" && cfg.HelpdeskServiceGRPCMTLSClientKey != "" {
		// Extract server name from address (e.g., "helpdesk-service:9002" -> "helpdesk-service")
		serverName := strings.Split(cfg.HelpDeskServiceGRPC, ":")[0]
		if serverName == "" || serverName == "localhost" {
			serverName = "helpdesk-service"
		}

		creds, err := utils.LoadGRPCClientCredentials(
			cfg.HelpdeskServiceGRPCMTLSClientCert,
			cfg.HelpdeskServiceGRPCMTLSClientKey,
			cfg.HelpdeskServiceGRPCMTLSCA,
			serverName,
		)
		if err != nil {
			log.Printf("Warning: failed to load Helpdesk Service gRPC TLS credentials: %v", err)
			log.Println("Falling back to insecure connection")
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(creds))
			log.Println("Helpdesk Service gRPC client configured with mTLS")
		}
	} else {
		log.Println("Warning: Helpdesk Service gRPC mTLS not configured, using insecure connection")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.HelpDeskServiceGRPC, opts...)
	if err != nil {
		log.Printf("Warning: failed to connect to Help Desk Service: %v", err)
		return &HelpDeskClient{ticketClient: nil, typeClient: nil}
	}

	return &HelpDeskClient{
		ticketClient: helpdeskpb.NewHelpDeskServiceClient(conn),
		typeClient:   helpdeskpb.NewHelpDeskTypeServiceClient(conn),
	}
}

func (c *HelpDeskClient) CreateTicket(ctx context.Context, req *helpdeskpb.CreateTicketRequest) (*helpdeskpb.TicketResponse, error) {
	if c.ticketClient == nil {
		return nil, grpc.ErrClientConnClosing
	}
	return c.ticketClient.CreateTicket(ctx, req)
}

func (c *HelpDeskClient) GetTicket(ctx context.Context, req *helpdeskpb.GetTicketRequest) (*helpdeskpb.TicketResponse, error) {
	if c.ticketClient == nil {
		return nil, grpc.ErrClientConnClosing
	}
	return c.ticketClient.GetTicket(ctx, req)
}

func (c *HelpDeskClient) ListTickets(ctx context.Context, req *helpdeskpb.ListTicketsRequest) (*helpdeskpb.ListTicketsResponse, error) {
	if c.ticketClient == nil {
		return nil, grpc.ErrClientConnClosing
	}
	return c.ticketClient.ListTickets(ctx, req)
}
