package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/repository"

	"gorm.io/gorm"
)

type FeedbackService struct{}

func NewFeedbackService() *FeedbackService {
	return &FeedbackService{}
}

// CreateFeedback 创建反馈（仅返回错误）
func (s *FeedbackService) CreateFeedback(title, content, feedbackType, anonymousID string) error {
	_, err := s.Create(feedbackType, title, content, anonymousID)
	return err
}

// Create 创建反馈并返回创建的反馈
func (s *FeedbackService) Create(feedbackType, title, content, anonymousID string) (*model.Feedback, error) {
	feedback := model.Feedback{
		Title:       title,
		Content:     content,
		Type:        feedbackType,
		AnonymousID: anonymousID,
		Status:      "open",
		LikeCount:   0,
	}
	if err := repository.DB.Create(&feedback).Error; err != nil {
		return nil, err
	}
	return &feedback, nil
}

// GetFeedbackList 获取反馈列表（按创建时间降序）
func (s *FeedbackService) GetFeedbackList() ([]model.Feedback, error) {
	var feedbacks []model.Feedback
	result := repository.DB.Order("created_at desc").Find(&feedbacks)
	return feedbacks, result.Error
}

// GetAll 别名兼容，调用 GetFeedbackList
func (s *FeedbackService) GetAll() ([]model.Feedback, error) {
	return s.GetFeedbackList()
}

// LikeFeedback 点赞反馈
func (s *FeedbackService) LikeFeedback(id uint) error {
	return repository.DB.Model(&model.Feedback{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
}

// Like 别名兼容，调用 LikeFeedback
func (s *FeedbackService) Like(id uint) error {
	return s.LikeFeedback(id)
}
