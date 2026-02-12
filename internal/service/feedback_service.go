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
		Status:      "pending",
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

// GetByID 根据ID获取反馈详情
func (s *FeedbackService) GetByID(id uint) (*model.Feedback, error) {
	return s.feedbackRepo.GetByID(id)
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

// UpdateStatus 更新反馈状态
func (s *FeedbackService) UpdateStatus(id uint, status string, teacherID string) error {
	feedback, err := s.feedbackRepo.GetByID(id)
	if err != nil {
		return err
	}
	feedback.Status = status
	return s.feedbackRepo.Update(feedback)
}

// Respond 教师回复反馈
func (s *FeedbackService) Respond(id uint, response string, teacherID string) error {
	feedback, err := s.feedbackRepo.GetByID(id)
	if err != nil {
		return err
	}
	feedback.TeacherResponse = response
	feedback.Status = "processing"
	return s.feedbackRepo.Update(feedback)
}

// GetStats 获取反馈统计数据
func (s *FeedbackService) GetStats() (map[string]interface{}, error) {
	feedbacks, err := s.feedbackRepo.GetAll()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":      len(feedbacks),
		"pending":    0,
		"processing": 0,
		"resolved":   0,
		"closed":     0,
		"by_type": map[string]int{
			"bug":        0,
			"feature":    0,
			"praise":     0,
			"suggestion": 0,
			"question":   0,
			"other":      0,
		},
	}

	for _, f := range feedbacks {
		// 状态统计
		switch f.Status {
		case "pending":
			stats["pending"] = stats["pending"].(int) + 1
		case "processing":
			stats["processing"] = stats["processing"].(int) + 1
		case "resolved":
			stats["resolved"] = stats["resolved"].(int) + 1
		case "closed":
			stats["closed"] = stats["closed"].(int) + 1
		}

		// 类型统计
		if count, ok := stats["by_type"].(map[string]int)[f.Type]; ok {
			stats["by_type"].(map[string]int)[f.Type] = count + 1
		}
	}

	return stats, nil
}

// GetFiltered 根据条件过滤反馈
func (s *FeedbackService) GetFiltered(feedbackType, status, search string) ([]model.Feedback, error) {
	return s.feedbackRepo.GetFiltered(feedbackType, status, search)
}
