package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/pkg/siliconflow"
	"GoCodeMentor/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
)

type SessionService struct {
	client      *siliconflow.Client
	sessionRepo repository.ChatSessionRepository
	messageRepo repository.ChatMessageRepository
	userRepo    repository.UserRepository
	classRepo   repository.ClassRepository
}

func NewSessionService(
	client *siliconflow.Client,
	sessionRepo repository.ChatSessionRepository,
	messageRepo repository.ChatMessageRepository,
	userRepo repository.UserRepository,
	classRepo repository.ClassRepository,
) ISessionService {
	return &SessionService{
		client:      client,
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		classRepo:   classRepo,
	}
}

// Chat 对话并保存历史
func (s *SessionService) Chat(ctx context.Context, sessionID, userID, userQuestion string) (string, string, error) {
	// 1. 如果没有 sessionID，创建新的
	if sessionID == "" {
		sessionID = uuid.New().String()
		s.sessionRepo.Create(&model.ChatSession{
			ID:     sessionID,
			UserID: userID,
			Title:  userQuestion,
		})
		s.messageRepo.Create(&model.ChatMessage{
			SessionID: sessionID,
			Role:      "system",
			Content:   "你是一位专业的 Go 语言助教，擅长用通俗的例子解释概念。请用中文回答，提供代码示例。",
		})
	}

	// 2. 保存用户问题
	s.messageRepo.Create(&model.ChatMessage{
		SessionID: sessionID,
		Role:      "user",
		Content:   userQuestion,
	})

	// 3. 查询历史消息
	messages, err := s.messageRepo.GetBySessionID(sessionID)
	if err != nil {
		// even if we can't get history, we can still proceed
	}

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
	s.messageRepo.Create(&model.ChatMessage{
		SessionID: sessionID,
		Role:      "assistant",
		Content:   answer,
	})

	// 6. 更新会话标题
	if len(messages) <= 2 {
		session, err := s.sessionRepo.GetByID(sessionID)
		if err == nil {
			session.Title = userQuestion
			s.sessionRepo.Update(session)
		}
	}

	return answer, sessionID, nil
}

// GetHistory 获取历史记录
func (s *SessionService) GetHistory(sessionID, userID string) ([]model.ChatMessage, error) {
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("当前用户不存在")
	}

	if user.Role == "student" && session.UserID != userID {
		return nil, errors.New("无权查看该会话")
	}

	if user.Role == "teacher" && session.UserID != userID {
		student, err := s.userRepo.GetByID(session.UserID)
		if err != nil || student.Role != "student" {
			return nil, errors.New("会话所属学生不存在")
		}
		if student.ClassID == nil {
			return nil, errors.New("学生未加入任何班级")
		}
		class, err := s.classRepo.GetByID(*student.ClassID)
		if err != nil {
			return nil, errors.New("学生所在班级不存在")
		}
		if class.TeacherID != userID {
			return nil, errors.New("无权查看该班级学生的会话")
		}
	}

	return s.messageRepo.GetBySessionID(sessionID)
}

// GetUserSessions 获取用户的所有会话
func (s *SessionService) GetUserSessions(userID string) ([]model.ChatSession, error) {
	return s.sessionRepo.GetByUserID(userID)
}

// GetStudentSessions 教师获取班级学生的会话
func (s *SessionService) GetStudentSessions(teacherID, studentID string) ([]model.ChatSession, error) {
	student, err := s.userRepo.GetByID(studentID)
	if err != nil || student.Role != "student" {
		return nil, errors.New("学生不存在")
	}

	if student.ClassID == nil {
		return []model.ChatSession{}, nil // Return empty slice instead of nil
	}

	class, err := s.classRepo.GetByID(*student.ClassID)
	if err != nil {
		return nil, errors.New("班级不存在")
	}

	if class.TeacherID != teacherID {
		return nil, errors.New("无权查看该学生的会话")
	}

	return s.sessionRepo.GetByUserID(studentID)
}
