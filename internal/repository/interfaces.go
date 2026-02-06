package repository

import (
	"GoCodeMentor/internal/model"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	Create(user *model.User) error
	GetByUsername(username string) (*model.User, error)
	GetByID(id string) (*model.User, error)
	GetByClassID(classID string) ([]model.User, error)
	Update(user *model.User) error
}

// ClassRepository defines the interface for class data operations.
type ClassRepository interface {
	Create(class *model.Class) error
	GetByID(id string) (*model.Class, error)
	GetByCode(code string) (*model.Class, error)
	GetByTeacherID(teacherID string) ([]model.Class, error)
}

// AssignmentRepository defines the interface for assignment data operations.
type AssignmentRepository interface {
	Create(assignment *model.Assignment) error
	GetByID(id string) (*model.Assignment, error)
	GetByTeacherID(teacherID string) ([]model.Assignment, error)
	GetByClassID(classID string) ([]model.Assignment, error)
	Update(assignment *model.Assignment) error
}

// QuestionRepository defines the interface for question data operations.
type QuestionRepository interface {
	Create(question *model.Question) error
	GetByAssignmentID(assignmentID string) ([]model.Question, error)
}

// SubmissionRepository defines the interface for submission data operations.
type SubmissionRepository interface {
	Create(submission *model.Submission) error
	GetByID(id string) (*model.Submission, error)
	GetByAssignmentAndStudent(assignmentID, studentID string) (*model.Submission, error)
	Update(submission *model.Submission) error
	CountByAssignmentID(assignmentID string, status string) (int64, error)
}

// ChatSessionRepository defines the interface for chat session data operations.
type ChatSessionRepository interface {
	Create(session *model.ChatSession) error
	GetByID(id string) (*model.ChatSession, error)
	GetByUserID(userID string) ([]model.ChatSession, error)
	Update(session *model.ChatSession) error
}

// ChatMessageRepository defines the interface for chat message data operations.
type ChatMessageRepository interface {
	Create(message *model.ChatMessage) error
	GetBySessionID(sessionID string) ([]model.ChatMessage, error)
}

// FeedbackRepository defines the interface for feedback data operations.
type FeedbackRepository interface {
	Create(feedback *model.Feedback) error
	GetAll() ([]model.Feedback, error)
	GetByID(id uint) (*model.Feedback, error)
	Update(feedback *model.Feedback) error
}
