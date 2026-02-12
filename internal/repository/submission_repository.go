package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// submissionRepository implements the SubmissionRepository interface.
type submissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository creates a new SubmissionRepository.
func NewSubmissionRepository(db *gorm.DB) SubmissionRepository {
	return &submissionRepository{db: db}
}

func (r *submissionRepository) Create(submission *model.Submission) error {
	return r.db.Create(submission).Error
}

func (r *submissionRepository) GetByID(id string) (*model.Submission, error) {
	var submission model.Submission
	err := r.db.Where("id = ?", id).First(&submission).Error
	return &submission, err
}

func (r *submissionRepository) GetByAssignmentAndStudent(assignmentID, studentID string) (*model.Submission, error) {
	var submission model.Submission
	result := r.db.Where("assignment_id = ? AND student_id = ?", assignmentID, studentID).Limit(1).Find(&submission)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &submission, nil
}

func (r *submissionRepository) Update(submission *model.Submission) error {
	return r.db.Save(submission).Error
}

func (r *submissionRepository) CountByAssignmentID(assignmentID string, status string) (int64, error) {
	var count int64
	db := r.db.Model(&model.Submission{}).Where("assignment_id = ?", assignmentID)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	err := db.Count(&count).Error
	return count, err
}

func (r *submissionRepository) DeleteByAssignmentID(assignmentID string) error {
	return r.db.Where("assignment_id = ?", assignmentID).Delete(&model.Submission{}).Error
}
