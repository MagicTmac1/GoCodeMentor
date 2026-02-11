package model

import (
	"time"

	"gorm.io/gorm"
)

// ========== 用户系统 ==========

type User struct {
	ID        string  `gorm:"primaryKey;type:uuid"`
	Username  string  `gorm:"size:100;uniqueIndex"`
	Password  string  `gorm:"size:100"`
	Name      string  `gorm:"size:100"`
	Role      string  `gorm:"size:20"` // teacher, student
	ClassID   *string `gorm:"index;type:uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Class struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	Name      string `gorm:"size:100"`
	TeacherID string `gorm:"index;type:uuid"`
	Code      string `gorm:"size:20;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ========== 答疑系统 ==========

type ChatSession struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	UserID    string `gorm:"index;type:uuid"`
	Title     string `gorm:"size:200"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ChatMessage struct {
	ID        uint   `gorm:"primaryKey"`
	SessionID string `gorm:"index;type:uuid"`
	Role      string `gorm:"size:20"` // system, user, assistant
	Content   string `gorm:"type:text"`
	CreatedAt time.Time
}

// ========== 作业系统 ==========

type Assignment struct {
	ID          string     `gorm:"primaryKey;type:uuid"`
	Title       string     `gorm:"size:200"`
	Description string     `gorm:"type:text"`
	TeacherID   string     `gorm:"size:100"`
	Type        string     `gorm:"size:20;default:'code'"`  // code, choice, fill, mixed
	Status      string     `gorm:"size:20;default:'draft'"` // draft, published, closed
	ClassID     *string    `gorm:"index;type:uuid"`         // 发布到的班级
	Rubric      string     `gorm:"type:jsonb"`              // 评分标准
	Deadline    *time.Time `gorm:"type:timestamp"`          // 截止时间
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Question struct {
	ID           string `gorm:"primaryKey;type:uuid"`
	AssignmentID string `gorm:"index;type:uuid"`
	Type         string `gorm:"size:20"` // choice, fill, code
	Content      string `gorm:"type:text"`
	Options      string `gorm:"type:jsonb"` // 选择题选项
	Answer       string `gorm:"type:text"`
	Score        int
	OrderNum     int
}

type Submission struct {
	ID               string `gorm:"primaryKey;type:uuid"`
	AssignmentID     string `gorm:"index;type:uuid"`
	StudentID        string `gorm:"size:100"`
	StudentName      string `gorm:"size:100"`
	Answers          string `gorm:"type:jsonb"`
	CodeContent      string `gorm:"type:text"`
	TotalScore       *int
	AIFeedback       string `gorm:"type:text"`
	TeacherFeedback  string `gorm:"type:text"`  // 教师总体批注
	QuestionFeedback string `gorm:"type:jsonb"` // 每个题目的批注，JSON格式：{"question_id": "feedback"}
	QuestionScores   string `gorm:"type:jsonb"` // 每个题目的分数，JSON格式：{"question_id": score}
	DetailedScore    string `gorm:"type:jsonb"`
	Status           string `gorm:"size:20;default:'submitted'"` // submitted, graded
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// 作业与班级关联表（支持多班级发布）
type AssignmentClass struct {
	ID           string     `gorm:"primaryKey;type:uuid"`
	AssignmentID string     `gorm:"index;type:uuid"`
	ClassID      string     `gorm:"index;type:uuid"`
	Deadline     *time.Time `gorm:"type:timestamp"` // 该班级的截止时间
	ReleasedAt   *time.Time `gorm:"type:timestamp"` // 发放时间（发布给该班级的时间）
	CreatedAt    time.Time
}

// 包含班级名称的作业-班级关联结构
type AssignmentClassWithClassName struct {
	AssignmentClass
	ClassName string `json:"class_name"`
}

// ========== 反馈系统 ==========

type Feedback struct {
	ID               uint       `gorm:"primaryKey"`
	Title            string     `gorm:"size:200"`
	Content          string     `gorm:"type:text"`
	AnonymousID      string     `gorm:"size:100"`
	Type             string     `gorm:"size:20"` // bug, feature, praise, other
	Status           string     `gorm:"size:20;default:'open'"`
	LikeCount        int
	TeacherResponse  string     `gorm:"type:text"`          // 教师回复内容
	RespondedAt      *time.Time `gorm:"type:timestamp"`     // 回复时间
	CreatedAt        time.Time
	UpdatedAt        time.Time    `gorm:"autoUpdateTime"`   // 添加更新时间
}
