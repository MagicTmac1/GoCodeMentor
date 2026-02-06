package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// questionRepository implements the QuestionRepository interface.
type questionRepository struct {
	db *gorm.DB
}

// NewQuestionRepository creates a new QuestionRepository.
func NewQuestionRepository(db *gorm.DB) QuestionRepository {
	return &questionRepository{db: db}
}

func (r *questionRepository) Create(question *model.Question) error {
	return r.db.Create(question).Error
}

func (r *questionRepository) GetByAssignmentID(assignmentID string) ([]model.Question, error) {
	var questions []model.Question
	err := r.db.Where("assignment_id = ?", assignmentID).Order("order_num asc").Find(&questions).Error
	return questions, err
}
