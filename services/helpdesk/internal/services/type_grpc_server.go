package services

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"backend/services/helpdesk/internal/models"
	helpdeskpb "backend/services/helpdesk/proto"
)

type HelpDeskTypeGRPCServer struct {
	helpdeskpb.UnimplementedHelpDeskTypeServiceServer
	typeService *HelpDeskTypeService
}

func NewHelpDeskTypeGRPCServer(typeService *HelpDeskTypeService) *HelpDeskTypeGRPCServer {
	return &HelpDeskTypeGRPCServer{typeService: typeService}
}

func (s *HelpDeskTypeGRPCServer) CreateType(ctx context.Context, req *helpdeskpb.CreateTypeRequest) (*helpdeskpb.TypeResponse, error) {
	var teamID *int
	if req.TeamId != 0 {
		teamID = intPtr(int(req.TeamId))
	}
	isActive := req.IsActive

	helpDeskType, err := s.typeService.CreateType(req.Name, req.Description, teamID, &isActive)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toTypeResponse(helpDeskType), nil
}

func (s *HelpDeskTypeGRPCServer) GetType(ctx context.Context, req *helpdeskpb.GetTypeRequest) (*helpdeskpb.TypeResponse, error) {
	helpDeskType, err := s.typeService.GetType(int(req.TypeId))
	if err != nil {
		return nil, status.Error(codes.NotFound, "type not found")
	}

	return toTypeResponse(helpDeskType), nil
}

func (s *HelpDeskTypeGRPCServer) ListTypes(ctx context.Context, req *helpdeskpb.ListTypesRequest) (*helpdeskpb.ListTypesResponse, error) {
	types, err := s.typeService.ListTypes(req.ActiveOnly)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list types")
	}

	var typeResponses []*helpdeskpb.TypeResponse
	for _, t := range types {
		typeResponses = append(typeResponses, toTypeResponse(&t))
	}

	return &helpdeskpb.ListTypesResponse{
		Types: typeResponses,
	}, nil
}

func (s *HelpDeskTypeGRPCServer) UpdateType(ctx context.Context, req *helpdeskpb.UpdateTypeRequest) (*helpdeskpb.TypeResponse, error) {
	var name, description *string
	var isActive *bool

	if req.Name != "" {
		name = &req.Name
	}
	if req.Description != "" {
		description = &req.Description
	}
	if req.IsActive {
		isActive = &req.IsActive
	}

	helpDeskType, err := s.typeService.UpdateType(int(req.TypeId), name, description, isActive)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toTypeResponse(helpDeskType), nil
}

func (s *HelpDeskTypeGRPCServer) DeleteType(ctx context.Context, req *helpdeskpb.DeleteTypeRequest) (*helpdeskpb.DeleteTypeResponse, error) {
	if err := s.typeService.DeleteType(int(req.TypeId)); err != nil {
		return nil, status.Error(codes.Internal, "failed to delete type")
	}

	return &helpdeskpb.DeleteTypeResponse{}, nil
}

func toTypeResponse(t *models.HelpDeskType) *helpdeskpb.TypeResponse {
	var teamID int32
	if t.TeamID != nil {
		teamID = int32(*t.TeamID)
	}
	isActive := false
	if t.IsActive != nil {
		isActive = *t.IsActive
	}

	return &helpdeskpb.TypeResponse{
		Id:          int32(t.ID),
		Name:        t.Name,
		Description: t.Description,
		TeamId:      teamID,
		IsActive:    isActive,
	}
}
