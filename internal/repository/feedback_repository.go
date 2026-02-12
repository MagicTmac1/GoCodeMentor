package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// feedbackRepository implements the FeedbackRepository interface.
type feedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository creates a new FeedbackRepository.
func NewFeedbackRepository(db *gorm.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

func (r *feedbackRepository) Create(feedback *model.Feedback) error {
	return r.db.Create(feedback).Error
}

func (r *feedbackRepository) GetAll() ([]model.Feedback, error) {
	var feedbacks []model.Feedback
	// 使用数据库中的实际字段名 like_count，而不是 likes
	err := r.db.Order("like_count desc, created_at desc").Find(&feedbacks).Error
	return feedbacks, err
}

func (r *feedbackRepository) GetFiltered(feedbackType, status, search string) ([]model.Feedback, error) {
	var feedbacks []model.Feedback
	query := r.db.Model(&model.Feedback{})

	if feedbackType != "" {
		query = query.Where("type = ?", feedbackType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		searchText := "%" + search + "%"
		query = query.Where("title LIKE ? OR content LIKE ? OR anonymous_id LIKE ?", searchText, searchText, searchText)
	}

	err := query.Order("like_count desc, created_at desc").Find(&feedbacks).Error
	return feedbacks, err
}

func (r *feedbackRepository) GetByID(id uint) (*model.Feedback, error) {
	var feedback model.Feedback
	err := r.db.First(&feedback, id).Error
	return &feedback, err
}

func (r *feedbackRepository) Update(feedback *model.Feedback) error {
	return r.db.Save(feedback).Error
}
