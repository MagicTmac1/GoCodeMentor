package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// classRepository implements the ClassRepository interface.
type classRepository struct {
	db *gorm.DB
}

// NewClassRepository creates a new ClassRepository.
func NewClassRepository(db *gorm.DB) ClassRepository {
	return &classRepository{db: db}
}

func (r *classRepository) Create(class *model.Class) error {
	return r.db.Create(class).Error
}

func (r *classRepository) GetByID(id string) (*model.Class, error) {
	var class model.Class
	err := r.db.Where("id = ?", id).First(&class).Error
	return &class, err
}

func (r *classRepository) GetByCode(code string) (*model.Class, error) {
	var class model.Class
	err := r.db.Where("code = ?", code).First(&class).Error
	return &class, err
}

func (r *classRepository) GetByTeacherID(teacherID string) ([]model.Class, error) {
	var classes []model.Class
	err := r.db.Where("teacher_id = ?", teacherID).Find(&classes).Error
	return classes, err
}
