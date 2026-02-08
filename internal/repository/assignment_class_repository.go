package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// assignmentClassRepository implements the AssignmentClassRepository interface.
type assignmentClassRepository struct {
	db *gorm.DB
}

// NewAssignmentClassRepository creates a new AssignmentClassRepository.
func NewAssignmentClassRepository(db *gorm.DB) AssignmentClassRepository {
	return &assignmentClassRepository{db: db}
}

func (r *assignmentClassRepository) Create(assignmentClass *model.AssignmentClass) error {
	return r.db.Create(assignmentClass).Error
}

func (r *assignmentClassRepository) GetByAssignmentID(assignmentID string) ([]model.AssignmentClass, error) {
	var assignmentClasses []model.AssignmentClass
	err := r.db.Where("assignment_id = ?", assignmentID).Find(&assignmentClasses).Error
	return assignmentClasses, err
}

func (r *assignmentClassRepository) GetByClassID(classID string) ([]model.AssignmentClass, error) {
	var assignmentClasses []model.AssignmentClass
	err := r.db.Where("class_id = ?", classID).Find(&assignmentClasses).Error
	return assignmentClasses, err
}

func (r *assignmentClassRepository) GetByAssignmentAndClass(assignmentID, classID string) (*model.AssignmentClass, error) {
	var assignmentClass model.AssignmentClass
	err := r.db.Where("assignment_id = ? AND class_id = ?", assignmentID, classID).First(&assignmentClass).Error
	if err != nil {
		return nil, err
	}
	return &assignmentClass, nil
}

func (r *assignmentClassRepository) Update(assignmentClass *model.AssignmentClass) error {
	return r.db.Save(assignmentClass).Error
}

func (r *assignmentClassRepository) DeleteByAssignmentID(assignmentID string) error {
	return r.db.Where("assignment_id = ?", assignmentID).Delete(&model.AssignmentClass{}).Error
}

func (r *assignmentClassRepository) DeleteByAssignmentAndClass(assignmentID, classID string) error {
	return r.db.Where("assignment_id = ? AND class_id = ?", assignmentID, classID).Delete(&model.AssignmentClass{}).Error
}
