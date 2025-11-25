package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"backend/services/helpdesk/internal/models"
	helpdeskpb "backend/services/helpdesk/proto"
)

type HelpDeskGRPCServer struct {
	helpdeskpb.UnimplementedHelpDeskServiceServer
	ticketService *TicketService
}

func NewHelpDeskGRPCServer(ticketService *TicketService) *HelpDeskGRPCServer {
	return &HelpDeskGRPCServer{ticketService: ticketService}
}

func (s *HelpDeskGRPCServer) CreateTicket(ctx context.Context, req *helpdeskpb.CreateTicketRequest) (*helpdeskpb.TicketResponse, error) {
	var parentID *int
	if req.ParentId != 0 {
		parentID = intPtr(int(req.ParentId))
	}
	var projectID *int
	if req.ProjectId != 0 {
		projectID = intPtr(int(req.ProjectId))
	}

	ticket, err := s.ticketService.CreateTicket(
		req.Name,
		req.Description,
		int(req.HelpDeskTypeId),
		int(req.PortalUserId),
		req.State,
		parentID,
		projectID,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toTicketResponse(ticket), nil
}

func (s *HelpDeskGRPCServer) GetTicket(ctx context.Context, req *helpdeskpb.GetTicketRequest) (*helpdeskpb.TicketResponse, error) {
	ticket, err := s.ticketService.GetTicket(int(req.TicketId))
	if err != nil {
		return nil, status.Error(codes.NotFound, "ticket not found")
	}

	return toTicketResponse(ticket), nil
}

func (s *HelpDeskGRPCServer) UpdateTicket(ctx context.Context, req *helpdeskpb.UpdateTicketRequest) (*helpdeskpb.TicketResponse, error) {
	var name, description, state *string
	var typeID *int

	if req.Name != "" {
		name = &req.Name
	}
	if req.Description != "" {
		description = &req.Description
	}
	if req.State != "" {
		state = &req.State
	}
	if req.HelpDeskTypeId != 0 {
		typeID = intPtr(int(req.HelpDeskTypeId))
	}

	ticket, err := s.ticketService.UpdateTicket(int(req.TicketId), name, description, state, typeID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toTicketResponse(ticket), nil
}

func (s *HelpDeskGRPCServer) ListTickets(ctx context.Context, req *helpdeskpb.ListTicketsRequest) (*helpdeskpb.ListTicketsResponse, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}

	var userID *int
	if req.UserId != 0 {
		userID = intPtr(int(req.UserId))
	}
	var state *string
	if req.State != "" {
		state = &req.State
	}

	tickets, total, err := s.ticketService.ListTickets(page, pageSize, userID, state)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list tickets")
	}

	var ticketResponses []*helpdeskpb.TicketResponse
	for _, ticket := range tickets {
		ticketResponses = append(ticketResponses, toTicketResponse(&ticket))
	}

	return &helpdeskpb.ListTicketsResponse{
		Tickets:  ticketResponses,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

func (s *HelpDeskGRPCServer) DeleteTicket(ctx context.Context, req *helpdeskpb.DeleteTicketRequest) (*helpdeskpb.DeleteTicketResponse, error) {
	if err := s.ticketService.DeleteTicket(int(req.TicketId)); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete ticket")
	}

	return &helpdeskpb.DeleteTicketResponse{}, nil
}

func toTicketResponse(ticket *models.HelpDesk) *helpdeskpb.TicketResponse {
	var parentID int32
	if ticket.ParentID != nil {
		parentID = int32(*ticket.ParentID)
	}
	var projectID int32
	if ticket.ProjectID != nil {
		projectID = int32(*ticket.ProjectID)
	}

	return &helpdeskpb.TicketResponse{
		Id:             int32(ticket.ID),
		Name:           ticket.Name,
		CreateDate:     ticket.CreateDate.Format("2006-01-02"),
		HelpDeskTypeId: int32(ticket.HelpDeskTypeID),
		PortalUserId:   int32(ticket.PortalUserID),
		Description:    ticket.Description,
		State:          ticket.State,
		ResolveDate:    ticket.ResolveDate.Format("2006-01-02"),
		ParentId:       parentID,
		ProjectId:      projectID,
	}
}

func intPtr(i int) *int {
	return &i
}
