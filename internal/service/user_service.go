package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/repository"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// Register 用户注册
func (s *UserService) Register(username, password, name, role string) (*model.User, error) {
	// 检查用户名是否已存在
	var existing model.User
	if result := repository.DB.Where("username = ?", username).First(&existing); result.Error == nil {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		ID:       uuid.New().String(),
		Username: username,
		Password: string(hashedPassword),
		Name:     name,
		Role:     role, // teacher 或 student
	}

	if result := repository.DB.Create(user); result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*model.User, error) {
	var user model.User
	if result := repository.DB.Where("username = ?", username).First(&user); result.Error != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return &user, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(id string) (*model.User, error) {
	var user model.User
	if result := repository.DB.First(&user, "id = ?", id); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户
func (s *UserService) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if result := repository.DB.Where("username = ?", username).First(&user); result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetStudentsByClassID 获取班级下的所有学生
func (s *UserService) GetStudentsByClassID(classID string) ([]model.User, error) {
	var students []model.User
	result := repository.DB.Where("role = ? AND class_id = ?", "student", classID).Find(&students)
	return students, result.Error
}
