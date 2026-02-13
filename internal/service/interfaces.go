package service

import (
	"GoCodeMentor/internal/model"
	"context"
	"time"
)

// IUserService 定义了用户相关的业务逻辑接口。
type IUserService interface {
	// Register 处理用户注册逻辑
	Register(username, password, name, role string) (*model.User, error)
	// Login 处理用户登录逻辑
	Login(username, password string) (*model.User, error)
	// GetByUsername 根据用户名查找用户
	GetByUsername(username string) (*model.User, error)
	// GetByID 根据用户 ID 查找用户
	GetByID(id string) (*model.User, error)
	// GetAllUsers 获取系统中的所有用户
	GetAllUsers() ([]model.User, error)
	// GetStudentsByClassID 获取特定班级的所有学生
	GetStudentsByClassID(classID string) ([]model.User, error)
	// UpdateUser 更新用户信息
	UpdateUser(user *model.User) error
	// ResetPassword 重置用户密码
	ResetPassword(userID, newPassword string) error
}

// IClassService 定义了班级管理相关的业务逻辑接口。
type IClassService interface {
	// CreateClass 创建一个新班级
	CreateClass(name, teacherID string) (*model.Class, error)
	// GetClassByID 获取班级详情
	GetClassByID(id string) (*model.Class, error)
	// GetClassesByTeacherID 获取某位教师创建的所有班级
	GetClassesByTeacherID(teacherID string) ([]model.Class, error)
	// GetClassByCode 根据邀请码查找班级
	GetClassByCode(code string) (*model.Class, error)
	// JoinClass 学生通过邀请码加入班级
	JoinClass(studentID, code string) error
	// AddStudentToClass 将学生手动添加到班级
	AddStudentToClass(studentID, classID string) error
	// RemoveStudentFromClass 将学生从班级中移除
	RemoveStudentFromClass(studentID, classID string) error
	// DeleteClass 删除班级及其相关关联
	DeleteClass(classID string) error
}

// IAssignmentService 定义了作业生成、发布与批改相关的业务逻辑接口。
type IAssignmentService interface {
	// GenerateAssignmentByAI 使用 AI 技术根据主题和难度生成作业题目
	GenerateAssignmentByAI(ctx context.Context, topic, difficulty, teacherID string) (*model.Assignment, error)
	// GetAssignmentList 获取教师创建的作业列表
	GetAssignmentList(teacherID string) ([]model.Assignment, error)
	// GetAllAssignments 获取系统内所有作业
	GetAllAssignments() ([]model.Assignment, error)
	// GetAssignmentsByClass 获取发布到特定班级的作业
	GetAssignmentsByClass(classID string) ([]model.Assignment, error)
	// PublishAssignment 将作业发布到指定班级并设置截止日期
	PublishAssignment(assignID, classID string, deadline *time.Time) error
	// GetPublishedClasses 获取作业已发布到的班级列表（包含班级名称）
	GetPublishedClasses(assignID string) ([]model.AssignmentClassWithClassName, error)
	// GetAssignmentDetail 获取作业详情及包含的所有题目
	GetAssignmentDetail(assignID string) (*model.Assignment, []model.Question, error)
	// SubmitAssignment 学生提交作业答案
	SubmitAssignment(assignID, studentID, studentName string, answers map[string]string, code string) (string, error)
	// GradeSubmission 使用 AI 对学生提交的作业进行自动评分和反馈
	GradeSubmission(ctx context.Context, subID string) error
	// GetSubmission 获取特定的提交记录
	GetSubmission(subID string) (*model.Submission, error)
	// GetSubmissionByAssignmentAndStudent 获取学生针对某个作业的提交记录
	GetSubmissionByAssignmentAndStudent(assignmentID, studentID string) (*model.Submission, error)
	// GetPendingSubmissionCountByAssignment 统计作业待批改的提交数
	GetPendingSubmissionCountByAssignment(assignmentID string) (int64, error)
	// UpdateSubmissionScore 手动更新学生作业得分
	UpdateSubmissionScore(submissionID string, score int) error
	// UpdateTeacherFeedback 更新教师对作业的评语
	UpdateTeacherFeedback(submissionID string, feedback string) error
	// RegradeSubmission 重新触发 AI 对作业的批改过程
	RegradeSubmission(submissionID string) error
	// GetSubmissionCodeForDownload 获取提交的代码内容用于下载
	GetSubmissionCodeForDownload(submissionID string) (string, string, error)
	// DeleteAssignment 删除作业及其关联题目和提交记录
	DeleteAssignment(assignID string) error
}

// IFeedbackService 定义了系统反馈与意见管理相关的业务逻辑接口。
type IFeedbackService interface {
	// Create 创建一条新的反馈记录
	Create(feedbackType, title, content, anonymousID string) (*model.Feedback, error)
	// GetAll 获取所有反馈列表
	GetAll() ([]model.Feedback, error)
	// GetByID 根据 ID 获取反馈详情
	GetByID(id uint) (*model.Feedback, error)
	// Like 为反馈点赞
	Like(id uint) error
	// UpdateStatus 更新反馈处理状态
	UpdateStatus(id uint, status string, teacherID string) error
	// Respond 教师对反馈进行回复
	Respond(id uint, response string, teacherID string) error
	// GetStats 获取反馈统计信息（如总数、待处理数等）
	GetStats() (map[string]interface{}, error)
	// GetFiltered 根据类型、状态和关键词过滤反馈
	GetFiltered(feedbackType, status, search string) ([]model.Feedback, error)
	// Delete 删除一条反馈
	Delete(id uint) error
}

// ISessionService 定义了 AI 助教对话会话相关的业务逻辑接口。
type ISessionService interface {
	// Chat 处理用户与 AI 助教的对话，支持上下文会话
	Chat(ctx context.Context, sessionID, userID, userQuestion string) (string, string, error)
	// GetHistory 获取会话的历史聊天记录
	GetHistory(sessionID, userID string) ([]model.ChatMessage, error)
	// GetUserSessions 获取用户创建的所有对话会话
	GetUserSessions(userID string) ([]model.ChatSession, error)
	// GetStudentSessions 教师获取特定学生的对话会话记录
	GetStudentSessions(teacherID, studentID string) ([]model.ChatSession, error)
}
