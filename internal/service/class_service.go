package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/repository"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type ClassService struct {
	classRepo repository.ClassRepository
	userRepo  repository.UserRepository
}

func NewClassService(classRepo repository.ClassRepository, userRepo repository.UserRepository) IClassService {
	return &ClassService{
		classRepo: classRepo,
		userRepo:  userRepo,
	}
}

// generateClassCode 生成6位数字邀请码
func generateClassCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// CreateClass 创建班级
func (s *ClassService) CreateClass(name, teacherID string) (*model.Class, error) {
	class := &model.Class{
		ID:        uuid.New().String(),
		Name:      name,
		TeacherID: teacherID,
		Code:      generateClassCode(),
	}

	if err := s.classRepo.Create(class); err != nil {
		return nil, err
	}

	return class, nil
}

// GetClassByID 根据ID获取班级
func (s *ClassService) GetClassByID(id string) (*model.Class, error) {
	return s.classRepo.GetByID(id)
}

// GetClassesByTeacherID 获取教师创建的所有班级
func (s *ClassService) GetClassesByTeacherID(teacherID string) ([]model.Class, error) {
	return s.classRepo.GetByTeacherID(teacherID)
}

// GetClassByCode 根据邀请码获取班级
func (s *ClassService) GetClassByCode(code string) (*model.Class, error) {
	return s.classRepo.GetByCode(code)
}

// JoinClass 学生加入班级
func (s *ClassService) JoinClass(studentID, code string) error {
	class, err := s.classRepo.GetByCode(code)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return errors.New("学生不存在")
	}

	user.ClassID = &class.ID
	return s.userRepo.Update(user)
}

// LeaveClass 学生退出班级
func (s *ClassService) LeaveClass(studentID string) error {
	user, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return errors.New("学生不存在")
	}

	user.ClassID = nil
	return s.userRepo.Update(user)
}

// AddStudentToClass 教师添加学生到班级
func (s *ClassService) AddStudentToClass(studentID, classID string) error {
	// 验证班级存在
	if _, err := s.classRepo.GetByID(classID); err != nil {
		return errors.New("班级不存在")
	}

	// 获取学生
	user, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return errors.New("学生不存在")
	}

	if user.Role != "student" {
		return errors.New("该用户不是学生")
	}

	// 更新学生的班级ID
	user.ClassID = &classID
	return s.userRepo.Update(user)
}

// RemoveStudentFromClass 教师从班级移除学生
func (s *ClassService) RemoveStudentFromClass(studentID, classID string) error {
	// 验证学生确实在这个班级
	user, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return errors.New("学生不存在")
	}

	if user.ClassID == nil || *user.ClassID != classID {
		return errors.New("学生不在该班级")
	}

	// 清空学生的班级ID
	user.ClassID = nil
	return s.userRepo.Update(user)
}
