package clients

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	helpdeskpb "backend/services/helpdesk/proto"
)

type HelpDeskClient struct {
	ticketClient helpdeskpb.HelpDeskServiceClient
	typeClient   helpdeskpb.HelpDeskTypeServiceClient
}

func NewHelpDeskClient(helpDeskServiceAddr string) *HelpDeskClient {
	conn, err := grpc.Dial(helpDeskServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
