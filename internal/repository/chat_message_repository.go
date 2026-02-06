package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// chatMessageRepository implements the ChatMessageRepository interface.
type chatMessageRepository struct {
	db *gorm.DB
}

// NewChatMessageRepository creates a new ChatMessageRepository.
func NewChatMessageRepository(db *gorm.DB) ChatMessageRepository {
	return &chatMessageRepository{db: db}
}

func (r *chatMessageRepository) Create(message *model.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *chatMessageRepository) GetBySessionID(sessionID string) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := r.db.Where("session_id = ?", sessionID).Order("created_at asc").Find(&messages).Error
	return messages, err
}
