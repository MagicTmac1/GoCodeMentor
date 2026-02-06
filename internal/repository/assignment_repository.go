package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// assignmentRepository implements the AssignmentRepository interface.
type assignmentRepository struct {
	db *gorm.DB
}

// NewAssignmentRepository creates a new AssignmentRepository.
func NewAssignmentRepository(db *gorm.DB) AssignmentRepository {
	return &assignmentRepository{db: db}
}

func (r *assignmentRepository) Create(assignment *model.Assignment) error {
	return r.db.Create(assignment).Error
}

func (r *assignmentRepository) GetByID(id string) (*model.Assignment, error) {
	var assignment model.Assignment
	err := r.db.Where("id = ?", id).First(&assignment).Error
	return &assignment, err
}

func (r *assignmentRepository) GetByTeacherID(teacherID string) ([]model.Assignment, error) {
	var assignments []model.Assignment
	// 查询当前教师的作业，以及teacher_id为空的作业（兼容旧数据）
	err := r.db.Where("teacher_id = ? OR teacher_id = ? OR teacher_id IS NULL", teacherID, "").Order("created_at desc").Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) GetByClassID(classID string) ([]model.Assignment, error) {
	var assignments []model.Assignment
	err := r.db.Where("class_id = ? AND status = ?", classID, "published").Order("created_at desc").Find(&assignments).Error
	return assignments, err
}

func (r *assignmentRepository) Update(assignment *model.Assignment) error {
	return r.db.Save(assignment).Error
}
