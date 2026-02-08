package repository

import "gorm.io/gorm"

// Repositories holds all repositories.
type Repositories struct {
	UserRepo            UserRepository
	ClassRepo           ClassRepository
	AssignmentRepo      AssignmentRepository
	AssignmentClassRepo AssignmentClassRepository
	QuestionRepo        QuestionRepository
	SubmissionRepo      SubmissionRepository
	FeedbackRepo        FeedbackRepository
	SessionRepo         ChatSessionRepository
	MessageRepo         ChatMessageRepository
}

// NewRepositories creates a new Repositories struct.
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserRepo:            NewUserRepository(db),
		ClassRepo:           NewClassRepository(db),
		AssignmentRepo:      NewAssignmentRepository(db),
		AssignmentClassRepo: NewAssignmentClassRepository(db),
		QuestionRepo:        NewQuestionRepository(db),
		SubmissionRepo:      NewSubmissionRepository(db),
		FeedbackRepo:        NewFeedbackRepository(db),
		SessionRepo:         NewChatSessionRepository(db),
		MessageRepo:         NewChatMessageRepository(db),
	}
}
