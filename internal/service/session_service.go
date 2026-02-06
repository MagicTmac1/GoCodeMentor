package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/pkg/siliconflow"
	"GoCodeMentor/internal/repository"
	"context"

	"github.com/google/uuid"
)

type SessionService struct {
	client *siliconflow.Client
}

func NewSessionService(client *siliconflow.Client) *SessionService {
	return &SessionService{client: client}
}

// Chat 对话并保存历史
func (s *SessionService) Chat(ctx context.Context, sessionID, userID, userQuestion string) (string, string, error) {
	// 1. 如果没有 sessionID，创建新的
	if sessionID == "" {
		sessionID = uuid.New().String()
		repository.DB.Create(&model.ChatSession{
			ID:     sessionID,
			UserID: userID,
			Title:  userQuestion,
		})
		repository.DB.Create(&model.ChatMessage{
			SessionID: sessionID,
			Role:      "system",
			Content:   "你是一位专业的 Go 语言助教，擅长用通俗的例子解释概念。请用中文回答，提供代码示例。",
		})
	}

	// 2. 保存用户问题
	repository.DB.Create(&model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   userQuestion,
	})

	// 3. 查询历史消息
	var messages []model.ChatMessage
	repository.DB.Where("session_id = ?", sessionID).
		Order("created_at asc").
		Limit(41).
		Find(&messages)

	// 4. 调用 AI
	var history []siliconflow.Message
	for _, m := range messages {
		history = append(history, siliconflow.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	answer, err := s.client.ChatWithHistory(ctx, history)
	if err != nil {
		return "", sessionID, err
	}

	// 5. 保存 AI 回答
	repository.DB.Create(&model.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   answer,
	})

	// 6. 更新会话标题
	if len(messages) <= 2 {
		repository.DB.Model(&model.ChatSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
			"title": userQuestion,
		})
	}

	return answer, sessionID, nil
}

// GetHistory 获取历史记录
func (s *SessionService) GetHistory(sessionID, userID string) ([]model.ChatMessage, error) {
	var session model.ChatSession
	if result := repository.DB.First(&session, "id = ?", sessionID); result.Error != nil {
		return nil, result.Error
	}

	// 检查权限
	var user model.User
	repository.DB.First(&user, "id = ?", userID)

	if user.Role == "student" && session.UserID != userID {
		return nil, nil
	}

	if user.Role == "teacher" && session.UserID != userID {
		var student model.User
		repository.DB.First(&student, "id = ? AND role = ?", session.UserID, "student")
		if student.ClassID == nil {
			return nil, nil
		}
		var class model.Class
		repository.DB.First(&class, "id = ?", *student.ClassID)
		if class.TeacherID != userID {
			return nil, nil
		}
	}

	var messages []model.ChatMessage
	result := repository.DB.Where("session_id = ?", sessionID).
		Order("created_at asc").
		Find(&messages)
	return messages, result.Error
}

// GetUserSessions 获取用户的所有会话
func (s *SessionService) GetUserSessions(userID string) ([]model.ChatSession, error) {
	var sessions []model.ChatSession
	result := repository.DB.Where("user_id = ?", userID).Order("updated_at desc").Find(&sessions)
	return sessions, result.Error
}

// GetStudentSessions 教师获取班级学生的会话
func (s *SessionService) GetStudentSessions(teacherID, studentID string) ([]model.ChatSession, error) {
	var student model.User
	if result := repository.DB.First(&student, "id = ? AND role = ?", studentID, "student"); result.Error != nil {
		return nil, result.Error
	}

	if student.ClassID == nil {
		return nil, nil
	}

	var class model.Class
	if result := repository.DB.First(&class, "id = ?", *student.ClassID); result.Error != nil {
		return nil, result.Error
	}

	if class.TeacherID != teacherID {
		return nil, nil
	}

	var sessions []model.ChatSession
	result := repository.DB.Where("user_id = ?", studentID).Order("updated_at desc").Find(&sessions)
	return sessions, result.Error
}
