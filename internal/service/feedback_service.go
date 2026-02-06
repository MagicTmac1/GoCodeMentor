package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/repository"
)

type FeedbackService struct {
	feedbackRepo repository.FeedbackRepository
}

func NewFeedbackService(feedbackRepo repository.FeedbackRepository) IFeedbackService {
	return &FeedbackService{feedbackRepo: feedbackRepo}
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
	if err := s.feedbackRepo.Create(&feedback); err != nil {
		return nil, err
	}
	return &feedback, nil
}

// GetAll 获取反馈列表
func (s *FeedbackService) GetAll() ([]model.Feedback, error) {
	return s.feedbackRepo.GetAll()
}

// Like 点赞反馈
func (s *FeedbackService) Like(id uint) error {
	feedback, err := s.feedbackRepo.GetByID(id)
	if err != nil {
		return err
	}
	feedback.LikeCount++
	return s.feedbackRepo.Update(feedback)
}
