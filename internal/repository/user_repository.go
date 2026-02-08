package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

// userRepository implements the UserRepository interface.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *userRepository) GetByClassID(classID string) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("class_id = ?", classID).Find(&users).Error
	return users, err
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(user *model.User) error {
	return r.db.Delete(user).Error
}
