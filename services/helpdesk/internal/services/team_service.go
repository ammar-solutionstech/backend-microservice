package services

import (
	"gorm.io/gorm"

	"backend/services/helpdesk/internal/config"
	"backend/services/helpdesk/internal/models"
)

type TeamService struct {
	cfg        *config.Config
	db         *gorm.DB
	authClient *authClient
}

func NewTeamService(cfg *config.Config, db *gorm.DB) *TeamService {
	return &TeamService{
		cfg:        cfg,
		db:         db,
		authClient: newAuthClient(cfg.AuthServiceGRPC),
	}
}

func (s *TeamService) CreateTeam(name string, description *string) (*models.Team, error) {
	team := &models.Team{
		Name:        name,
		Description: description,
	}
	if err := s.db.Create(team).Error; err != nil {
		return nil, err
	}
	return team, nil
}

func (s *TeamService) GetTeam(teamID int) (*models.Team, error) {
	var team models.Team
	if err := s.db.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}
	return &team, nil
}

func (s *TeamService) ListTeams() ([]models.Team, error) {
	var teams []models.Team
	if err := s.db.Find(&teams).Error; err != nil {
		return nil, err
	}
	return teams, nil
}

func (s *TeamService) UpdateTeam(teamID int, name *string, description *string) (*models.Team, error) {
	var team models.Team
	if err := s.db.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if description != nil {
		updates["description"] = *description
	}

	if err := s.db.Model(&team).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &team, nil
}

func (s *TeamService) DeleteTeam(teamID int) error {
	return s.db.Delete(&models.Team{}, teamID).Error
}

func (s *TeamService) GetTeamMembers(teamID int) ([]int, error) {
	var members []struct {
		UserID int `gorm:"column:user_id"`
	}
	if err := s.db.Table("team_members").Where("help_desk_team_id = ?", teamID).Find(&members).Error; err != nil {
		return nil, err
	}

	var userIDs []int
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}
	return userIDs, nil
}

func (s *TeamService) AddTeamMember(teamID, userID int) error {
	// Validate user
	if !s.authClient.ValidateUser(userID) {
		return gorm.ErrRecordNotFound
	}

	member := struct {
		UserID         int `gorm:"column:user_id;primaryKey"`
		HelpDeskTeamID int `gorm:"column:help_desk_team_id;primaryKey"`
	}{
		UserID:         userID,
		HelpDeskTeamID: teamID,
	}

	return s.db.Table("team_members").Where(member).FirstOrCreate(&member).Error
}

func (s *TeamService) RemoveTeamMember(teamID, userID int) error {
	return s.db.Table("team_members").
		Where("help_desk_team_id = ? AND user_id = ?", teamID, userID).
		Delete(nil).Error
}
