package service

import (
	"GoCodeMentor/internal/model"
	"context"
	"time"
)

type IUserService interface {
	Register(username, password, name, role string) (*model.User, error)
	Login(username, password string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id string) (*model.User, error)
	GetStudentsByClassID(classID string) ([]model.User, error)
	UpdateUser(user *model.User) error
}

type IClassService interface {
	CreateClass(name, teacherID string) (*model.Class, error)
	GetClassByID(id string) (*model.Class, error)
	GetClassesByTeacherID(teacherID string) ([]model.Class, error)
	GetClassByCode(code string) (*model.Class, error)
	JoinClass(studentID, code string) error
	AddStudentToClass(studentID, classID string) error
	RemoveStudentFromClass(studentID, classID string) error
	DeleteClass(classID string) error
}

type IAssignmentService interface {
	GenerateAssignmentByAI(ctx context.Context, topic, difficulty, teacherID string) (*model.Assignment, error)
	GetAssignmentList(teacherID string) ([]model.Assignment, error)
	GetAllAssignments() ([]model.Assignment, error)
	GetAssignmentsByClass(classID string) ([]model.Assignment, error)
	PublishAssignment(assignID, classID string, deadline *time.Time) error
	GetPublishedClasses(assignID string) ([]model.AssignmentClassWithClassName, error)
	GetAssignmentDetail(assignID string) (*model.Assignment, []model.Question, error)
	SubmitAssignment(assignID, studentID, studentName string, answers map[string]string, code string) (string, error)
	GradeSubmission(ctx context.Context, subID string) error
	GetSubmission(subID string) (*model.Submission, error)
	GetSubmissionByAssignmentAndStudent(assignmentID, studentID string) (*model.Submission, error)
	GetPendingSubmissionCountByAssignment(assignmentID string) (int64, error)
	UpdateSubmissionScore(submissionID string, score int) error
	UpdateTeacherFeedback(submissionID string, feedback string) error
	RegradeSubmission(submissionID string) error
	GetSubmissionCodeForDownload(submissionID string) (string, string, error)
	DeleteAssignment(assignID string) error
}

type IFeedbackService interface {
	Create(feedbackType, title, content, anonymousID string) (*model.Feedback, error)
	GetAll() ([]model.Feedback, error)
	GetByID(id uint) (*model.Feedback, error)
	Like(id uint) error
	UpdateStatus(id uint, status string, teacherID string) error
	Respond(id uint, response string, teacherID string) error
	GetStats() (map[string]interface{}, error)
	GetFiltered(feedbackType, status, search string) ([]model.Feedback, error)
}

type ISessionService interface {
	Chat(ctx context.Context, sessionID, userID, userQuestion string) (string, string, error)
	GetHistory(sessionID, userID string) ([]model.ChatMessage, error)
	GetUserSessions(userID string) ([]model.ChatSession, error)
	GetStudentSessions(teacherID, studentID string) ([]model.ChatSession, error)
}
