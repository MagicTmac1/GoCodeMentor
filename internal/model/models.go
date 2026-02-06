package model

import (
	"time"

	"gorm.io/gorm"
)

// ========== 用户系统 ==========

type User struct {
	ID        string `gorm:"primaryKey;type:uuid"`
	Username  string `gorm:"size:100;uniqueIndex"`
	Password  string `gorm:"size:100"`
	Name      string `gorm:"size:100"`
	Role      string `gorm:"size:20"` // teacher, student
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
	ID          string `gorm:"primaryKey;type:uuid"`
	Title       string `gorm:"size:200"`
	Description string `gorm:"type:text"`
	TeacherID   string `gorm:"size:100"`
	Type        string `gorm:"size:20;default:'code'"` // code, choice, fill, mixed
	Status      string `gorm:"size:20;default:'draft'"` // draft, published, closed
	ClassID     *string `gorm:"index;type:uuid"` // 发布到的班级
	Rubric      string `gorm:"type:jsonb"` // 评分标准
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
	ID            string `gorm:"primaryKey;type:uuid"`
	AssignmentID  string `gorm:"index;type:uuid"`
	StudentID     string `gorm:"size:100"`
	StudentName   string `gorm:"size:100"`
	Answers       string `gorm:"type:jsonb"`
	CodeContent   string `gorm:"type:text"`
	TotalScore    *int
	AIFeedback    string `gorm:"type:text"`
	DetailedScore string `gorm:"type:jsonb"`
	Status        string `gorm:"size:20;default:'submitted'"` // submitted, graded
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ========== 反馈系统 ==========

type Feedback struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:200"`
	Content     string `gorm:"type:text"`
	AnonymousID string `gorm:"size:100"`
	Type        string `gorm:"size:20"` // bug, feature, praise, other
	Status      string `gorm:"size:20;default:'open'"`
	LikeCount   int
	CreatedAt   time.Time
}
