package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// chatSessionRepository implements the ChatSessionRepository interface.
type chatSessionRepository struct {
	db *gorm.DB
}

// NewChatSessionRepository creates a new ChatSessionRepository.
func NewChatSessionRepository(db *gorm.DB) ChatSessionRepository {
	return &chatSessionRepository{db: db}
}

func (r *chatSessionRepository) Create(session *model.ChatSession) error {
	return r.db.Create(session).Error
}

func (r *chatSessionRepository) GetByID(id string) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Where("id = ?", id).First(&session).Error
	return &session, err
}

func (r *chatSessionRepository) GetByUserID(userID string) ([]model.ChatSession, error) {
	var sessions []model.ChatSession
	err := r.db.Where("user_id = ?", userID).Order("updated_at desc").Find(&sessions).Error
	return sessions, err
}

func (r *chatSessionRepository) Update(session *model.ChatSession) error {
	return r.db.Save(session).Error
}
