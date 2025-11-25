package services

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	helpdeskpb "backend/services/helpdesk/proto"
)

type TicketGRPCServer struct {
	helpdeskpb.UnimplementedTicketServiceServer
	ticketService *TicketService
}

func NewTicketGRPCServer(ticketService *TicketService) *TicketGRPCServer {
	return &TicketGRPCServer{ticketService: ticketService}
}

func (s *TicketGRPCServer) AddComment(ctx context.Context, req *helpdeskpb.AddCommentRequest) (*helpdeskpb.CommentResponse, error) {
	comment, err := s.ticketService.AddComment(int(req.TicketId), int(req.UserId), req.Content)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &helpdeskpb.CommentResponse{
		Id:        int32(comment.ID),
		TicketId:  int32(comment.TicketID),
		UserId:    int32(comment.UserID),
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *TicketGRPCServer) GetComments(ctx context.Context, req *helpdeskpb.GetCommentsRequest) (*helpdeskpb.GetCommentsResponse, error) {
	comments, err := s.ticketService.GetComments(int(req.TicketId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get comments")
	}

	var commentResponses []*helpdeskpb.CommentResponse
	for _, comment := range comments {
		commentResponses = append(commentResponses, &helpdeskpb.CommentResponse{
			Id:        int32(comment.ID),
			TicketId:  int32(comment.TicketID),
			UserId:    int32(comment.UserID),
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		})
	}

	return &helpdeskpb.GetCommentsResponse{
		Comments: commentResponses,
	}, nil
}

func (s *TicketGRPCServer) AddAttachment(ctx context.Context, req *helpdeskpb.AddAttachmentRequest) (*helpdeskpb.AttachmentResponse, error) {
	document, err := s.ticketService.AddAttachment(int(req.TicketId), req.Name, req.Content, req.DocumentType)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &helpdeskpb.AttachmentResponse{
		Id:           int32(document.ID),
		TicketId:     int32(document.HelpDeskID),
		Name:         document.Name,
		DocumentType: document.DocumentType,
		CreatedAt:    time.Now().Format(time.RFC3339),
	}, nil
}

func (s *TicketGRPCServer) GetAttachments(ctx context.Context, req *helpdeskpb.GetAttachmentsRequest) (*helpdeskpb.GetAttachmentsResponse, error) {
	attachments, err := s.ticketService.GetAttachments(int(req.TicketId))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get attachments")
	}

	var attachmentResponses []*helpdeskpb.AttachmentResponse
	for _, att := range attachments {
		attachmentResponses = append(attachmentResponses, &helpdeskpb.AttachmentResponse{
			Id:           int32(att.ID),
			TicketId:     int32(att.HelpDeskID),
			Name:         att.Name,
			DocumentType: att.DocumentType,
			CreatedAt:    time.Now().Format(time.RFC3339),
		})
	}

	return &helpdeskpb.GetAttachmentsResponse{
		Attachments: attachmentResponses,
	}, nil
}

func (s *TicketGRPCServer) AssignUser(ctx context.Context, req *helpdeskpb.AssignUserRequest) (*helpdeskpb.AssignUserResponse, error) {
	if err := s.ticketService.AssignUser(int(req.TicketId), int(req.UserId)); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &helpdeskpb.AssignUserResponse{
		Success: true,
	}, nil
}

func (s *TicketGRPCServer) UpdateStatus(ctx context.Context, req *helpdeskpb.UpdateStatusRequest) (*helpdeskpb.UpdateStatusResponse, error) {
	if err := s.ticketService.UpdateStatus(int(req.TicketId), req.Status); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &helpdeskpb.UpdateStatusResponse{
		Success: true,
	}, nil
}
