package services

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/models"
)

type TicketService struct {
	cfg        *config.Config
	db         *gorm.DB
	authClient *authClient
	rabbitMQ   *rabbitMQPublisher
}

func NewTicketService(cfg *config.Config, db *gorm.DB) *TicketService {
	authClient := newAuthClient(cfg)
	rabbitMQ := newRabbitMQPublisher(cfg)
	return &TicketService{
		cfg:        cfg,
		db:         db,
		authClient: authClient,
		rabbitMQ:   rabbitMQ,
	}
}

// CreateTicket creates a new help desk ticket
func (s *TicketService) CreateTicket(name, description string, helpDeskTypeID, portalUserID int, state string, parentID, projectID *int) (*models.HelpDesk, error) {
	// Validate user exists in Auth Service
	if !s.authClient.ValidateUser(portalUserID) {
		return nil, errors.New("invalid user")
	}

	ticket := &models.HelpDesk{
		Name:           name,
		Description:    description,
		HelpDeskTypeID: helpDeskTypeID,
		PortalUserID:   portalUserID,
		State:          state,
		CreateDate:     time.Now(),
		ResolveDate:    time.Now().Add(7 * 24 * time.Hour), // Default 7 days
		ParentID:       parentID,
		ProjectID:      projectID,
	}

	if err := s.db.Create(ticket).Error; err != nil {
		return nil, err
	}

	// Publish event
	s.rabbitMQ.PublishTicketCreated(ticket.ID, portalUserID)

	return ticket, nil
}

// GetTicket retrieves a ticket by ID
func (s *TicketService) GetTicket(ticketID int) (*models.HelpDesk, error) {
	var ticket models.HelpDesk
	if err := s.db.Preload("Type").Preload("Comments").Preload("Attachments").Preload("Parent").
		Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// ListTickets lists tickets with pagination
func (s *TicketService) ListTickets(page, pageSize int, userID *int, state *string) ([]models.HelpDesk, int64, error) {
	var tickets []models.HelpDesk
	var total int64

	query := s.db.Model(&models.HelpDesk{})

	if userID != nil {
		query = query.Where("portal_user_id = ?", *userID)
	}
	if state != nil {
		query = query.Where("state = ?", *state)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Preload("Type").Offset(offset).Limit(pageSize).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// UpdateTicket updates a ticket
func (s *TicketService) UpdateTicket(ticketID int, name, description, state *string, helpDeskTypeID *int) (*models.HelpDesk, error) {
	var ticket models.HelpDesk
	if err := s.db.Where("id = ?", ticketID).First(&ticket).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if description != nil {
		updates["description"] = *description
	}
	if state != nil {
		updates["state"] = *state
	}
	if helpDeskTypeID != nil {
		updates["help_desk_type_id"] = *helpDeskTypeID
	}

	if err := s.db.Model(&ticket).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Publish event
	s.rabbitMQ.PublishTicketUpdated(ticketID, ticket.PortalUserID)

	return &ticket, nil
}

// DeleteTicket deletes a ticket
func (s *TicketService) DeleteTicket(ticketID int) error {
	return s.db.Delete(&models.HelpDesk{}, ticketID).Error
}

// AddComment adds a comment to a ticket
func (s *TicketService) AddComment(ticketID, userID int, content string) (*models.TicketComment, error) {
	// Validate user
	if !s.authClient.ValidateUser(userID) {
		return nil, errors.New("invalid user")
	}

	comment := &models.TicketComment{
		TicketID: ticketID,
		UserID:   userID,
		Content:  content,
	}

	if err := s.db.Create(comment).Error; err != nil {
		return nil, err
	}

	return comment, nil
}

// GetComments retrieves comments for a ticket
func (s *TicketService) GetComments(ticketID int) ([]models.TicketComment, error) {
	var comments []models.TicketComment
	if err := s.db.Where("ticket_id = ?", ticketID).Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// AddAttachment adds an attachment to a ticket
func (s *TicketService) AddAttachment(ticketID int, name string, content []byte, documentType string) (*models.Document, error) {
	document := &models.Document{
		Name:         name,
		DocumentSize: string(rune(len(content))),
		Picture:      content,
		DocumentType: documentType,
		HelpDeskID:   ticketID,
		EquipmentID:  0,
		SupplierID:   0,
	}

	if err := s.db.Create(document).Error; err != nil {
		return nil, err
	}

	// Create attachment link
	attachment := &models.TicketAttachment{
		TicketID:   ticketID,
		DocumentID: document.ID,
	}
	if err := s.db.Create(attachment).Error; err != nil {
		return nil, err
	}

	return document, nil
}

// GetAttachments retrieves attachments for a ticket
func (s *TicketService) GetAttachments(ticketID int) ([]models.Document, error) {
	var attachments []models.TicketAttachment
	if err := s.db.Where("ticket_id = ?", ticketID).Find(&attachments).Error; err != nil {
		return nil, err
	}

	var documentIDs []int
	for _, att := range attachments {
		documentIDs = append(documentIDs, att.DocumentID)
	}

	var documents []models.Document
	if len(documentIDs) > 0 {
		if err := s.db.Where("id IN ?", documentIDs).Find(&documents).Error; err != nil {
			return nil, err
		}
	}

	return documents, nil
}

// AssignUser assigns a user to a ticket
func (s *TicketService) AssignUser(ticketID, userID int) error {
	// Validate user
	if !s.authClient.ValidateUser(userID) {
		return errors.New("invalid user")
	}

	assignment := &models.UserHelpDesk{
		UserID:     userID,
		HelpDeskID: ticketID,
	}

	if err := s.db.Where(assignment).FirstOrCreate(assignment).Error; err != nil {
		return err
	}

	// Publish event
	s.rabbitMQ.PublishTicketAssigned(ticketID, userID)

	return nil
}

// UpdateStatus updates ticket status
func (s *TicketService) UpdateStatus(ticketID int, status string) error {
	return s.db.Model(&models.HelpDesk{}).Where("id = ?", ticketID).Update("state", status).Error
}

// ListParticipants retrieves all participants for a help desk ticket
func (s *TicketService) ListParticipants(helpDeskID int) ([]models.UserHelpDesk, error) {
	var participants []models.UserHelpDesk
	if err := s.db.Where("help_desk_id = ?", helpDeskID).Find(&participants).Error; err != nil {
		return nil, err
	}
	return participants, nil
}

// AddParticipant adds a user as a participant to a help desk ticket
func (s *TicketService) AddParticipant(helpDeskID, userID int) error {
	// Validate user
	if !s.authClient.ValidateUser(userID) {
		return errors.New("invalid user")
	}

	participant := &models.UserHelpDesk{
		UserID:     userID,
		HelpDeskID: helpDeskID,
	}

	return s.db.Where(participant).FirstOrCreate(participant).Error
}

// RemoveParticipant removes a user from a help desk ticket participants
func (s *TicketService) RemoveParticipant(helpDeskID, userID int) error {
	result := s.db.Where("help_desk_id = ? AND user_id = ?", helpDeskID, userID).
		Delete(&models.UserHelpDesk{})
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}