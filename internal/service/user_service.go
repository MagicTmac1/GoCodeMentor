package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/repository"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) IUserService {
	return &UserService{userRepo: userRepo}
}

// Register 用户注册
func (s *UserService) Register(username, password, name, role string) (*model.User, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.GetByUsername(username); err == nil {
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

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*model.User, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	return user, nil
}

// GetByID 根据ID获取用户
func (s *UserService) GetByID(id string) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// GetByUsername 根据用户名获取用户
func (s *UserService) GetByUsername(username string) (*model.User, error) {
	return s.userRepo.GetByUsername(username)
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.userRepo.GetAll()
}

// GetStudentsByClassID 获取班级下的所有学生
func (s *UserService) GetStudentsByClassID(classID string) ([]model.User, error) {
	return s.userRepo.GetByClassID(classID)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.userRepo.Update(user)
}

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword(userID, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}
