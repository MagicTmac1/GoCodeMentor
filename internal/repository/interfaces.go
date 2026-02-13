package repository

import (
	"GoCodeMentor/internal/model"
)

// UserRepository 定义了用户数据操作的接口。
type UserRepository interface {
	// Create 创建一个新用户
	Create(user *model.User) error
	// GetByUsername 根据用户名获取用户
	GetByUsername(username string) (*model.User, error)
	// GetByID 根据用户 ID 获取用户
	GetByID(id string) (*model.User, error)
	// GetByClassID 根据班级 ID 获取该班级下的所有用户
	GetByClassID(classID string) ([]model.User, error)
	// GetAll 获取所有用户
	GetAll() ([]model.User, error)
	// Update 更新用户信息
	Update(user *model.User) error
	// Delete 删除用户
	Delete(user *model.User) error
}

// ClassRepository 定义了班级数据操作的接口。
type ClassRepository interface {
	// Create 创建一个新班级
	Create(class *model.Class) error
	// GetByID 根据班级 ID 获取班级
	GetByID(id string) (*model.Class, error)
	// GetByCode 根据班级邀请码获取班级
	GetByCode(code string) (*model.Class, error)
	// GetByTeacherID 获取教师创建的所有班级
	GetByTeacherID(teacherID string) ([]model.Class, error)
	// Delete 根据 ID 删除班级
	Delete(id string) error
}

// AssignmentRepository 定义了作业数据操作的接口。
type AssignmentRepository interface {
	// Create 创建一个新作业
	Create(assignment *model.Assignment) error
	// GetByID 根据作业 ID 获取作业
	GetByID(id string) (*model.Assignment, error)
	// GetByTeacherID 获取教师创建的所有作业
	GetByTeacherID(teacherID string) ([]model.Assignment, error)
	// GetByClassID 获取发布到特定班级的所有作业
	GetByClassID(classID string) ([]model.Assignment, error)
	// Update 更新作业信息
	Update(assignment *model.Assignment) error
	// DeleteByID 根据 ID 删除作业
	DeleteByID(id string) error
}

// AssignmentClassRepository 定义了作业与班级关联关系操作的接口。
type AssignmentClassRepository interface {
	// Create 创建作业与班级的关联关系（发布作业到班级）
	Create(assignmentClass *model.AssignmentClass) error
	// GetByAssignmentID 根据作业 ID 获取所有关联的班级关系
	GetByAssignmentID(assignmentID string) ([]model.AssignmentClass, error)
	// GetByClassID 根据班级 ID 获取该班级所有的作业关联关系
	GetByClassID(classID string) ([]model.AssignmentClass, error)
	// GetByAssignmentAndClass 根据作业 ID 和班级 ID 获取特定的关联关系
	GetByAssignmentAndClass(assignmentID, classID string) (*model.AssignmentClass, error)
	// Update 更新关联关系信息
	Update(assignmentClass *model.AssignmentClass) error
	// DeleteByAssignmentID 根据作业 ID 删除所有相关的班级关联关系
	DeleteByAssignmentID(assignmentID string) error
	// DeleteByAssignmentAndClass 根据作业 ID 和班级 ID 删除特定的关联关系（取消发布）
	DeleteByAssignmentAndClass(assignmentID, classID string) error
}

// QuestionRepository 定义了题目数据操作的接口。
type QuestionRepository interface {
	// Create 创建一个新题目
	Create(question *model.Question) error
	// GetByAssignmentID 根据作业 ID 获取该作业下的所有题目
	GetByAssignmentID(assignmentID string) ([]model.Question, error)
	// DeleteByAssignmentID 根据作业 ID 删除该作业下的所有题目
	DeleteByAssignmentID(assignmentID string) error
}

// SubmissionRepository 定义了学生提交记录数据操作的接口。
type SubmissionRepository interface {
	// Create 创建一个新的提交记录
	Create(submission *model.Submission) error
	// GetByID 根据提交 ID 获取提交详情
	GetByID(id string) (*model.Submission, error)
	// GetByAssignmentAndStudent 根据作业 ID 和学生 ID 获取特定的提交记录
	GetByAssignmentAndStudent(assignmentID, studentID string) (*model.Submission, error)
	// Update 更新提交记录（如批改结果、分数等）
	Update(submission *model.Submission) error
	// CountByAssignmentID 根据作业 ID 和状态统计提交数量
	CountByAssignmentID(assignmentID string, status string) (int64, error)
	// DeleteByAssignmentID 根据作业 ID 删除所有相关的提交记录
	DeleteByAssignmentID(assignmentID string) error
}

// ChatSessionRepository 定义了 AI 聊天会话数据操作的接口。
type ChatSessionRepository interface {
	// Create 创建一个新的聊天会话
	Create(session *model.ChatSession) error
	// GetByID 根据会话 ID 获取会话信息
	GetByID(id string) (*model.ChatSession, error)
	// GetByUserID 获取用户的所有聊天会话
	GetByUserID(userID string) ([]model.ChatSession, error)
	// Update 更新会话信息
	Update(session *model.ChatSession) error
}

// ChatMessageRepository 定义了聊天消息数据操作的接口。
type ChatMessageRepository interface {
	// Create 创建一条新的聊天消息
	Create(message *model.ChatMessage) error
	// GetBySessionID 根据会话 ID 获取该会话下的所有历史消息
	GetBySessionID(sessionID string) ([]model.ChatMessage, error)
}

// FeedbackRepository 定义了反馈数据操作的接口。
type FeedbackRepository interface {
	// Create 创建一条新反馈
	Create(feedback *model.Feedback) error
	// GetAll 获取所有反馈
	GetAll() ([]model.Feedback, error)
	// GetFiltered 根据类型、状态和搜索关键词筛选反馈
	GetFiltered(feedbackType, status, search string) ([]model.Feedback, error)
	// GetByID 根据 ID 获取反馈详情
	GetByID(id uint) (*model.Feedback, error)
	// Update 更新反馈信息（如状态更新、回复等）
	Update(feedback *model.Feedback) error
	// Delete 根据 ID 删除反馈
	Delete(id uint) error
}
