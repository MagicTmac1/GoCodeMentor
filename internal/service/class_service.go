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

type ClassService struct{}

func NewClassService() *ClassService {
	return &ClassService{}
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

	if result := repository.DB.Create(class); result.Error != nil {
		return nil, result.Error
	}

	return class, nil
}

// GetClassByID 根据ID获取班级
func (s *ClassService) GetClassByID(id string) (*model.Class, error) {
	var class model.Class
	if result := repository.DB.First(&class, "id = ?", id); result.Error != nil {
		return nil, result.Error
	}
	return &class, nil
}

// GetClassesByTeacherID 获取教师创建的所有班级
func (s *ClassService) GetClassesByTeacherID(teacherID string) ([]model.Class, error) {
	var classes []model.Class
	result := repository.DB.Where("teacher_id = ?", teacherID).Order("created_at desc").Find(&classes)
	return classes, result.Error
}

// GetClassByCode 根据邀请码获取班级
func (s *ClassService) GetClassByCode(code string) (*model.Class, error) {
	var class model.Class
	if result := repository.DB.Where("code = ?", code).First(&class); result.Error != nil {
		return nil, errors.New("班级不存在")
	}
	return &class, nil
}

// JoinClass 学生加入班级
func (s *ClassService) JoinClass(studentID, code string) error {
	class, err := s.GetClassByCode(code)
	if err != nil {
		return err
	}

	// 更新学生的班级ID
	result := repository.DB.Model(&model.User{}).Where("id = ?", studentID).Updates(map[string]interface{}{
		"class_id": class.ID,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// LeaveClass 学生退出班级
func (s *ClassService) LeaveClass(studentID string) error {
	result := repository.DB.Model(&model.User{}).Where("id = ?", studentID).Updates(map[string]interface{}{
		"class_id": nil,
	})
	return result.Error
}

// AddStudentToClass 教师添加学生到班级
func (s *ClassService) AddStudentToClass(studentID, classID string) error {
	// 验证班级存在
	var class model.Class
	if result := repository.DB.First(&class, "id = ?", classID); result.Error != nil {
		return errors.New("班级不存在")
	}

	// 更新学生的班级ID
	result := repository.DB.Model(&model.User{}).Where("id = ? AND role = ?", studentID, "student").Updates(map[string]interface{}{
		"class_id": class.ID,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("学生不存在或不是学生角色")
	}

	return nil
}

// RemoveStudentFromClass 教师从班级移除学生
func (s *ClassService) RemoveStudentFromClass(studentID, classID string) error {
	// 验证学生确实在这个班级
	var user model.User
	if result := repository.DB.First(&user, "id = ? AND class_id = ?", studentID, classID); result.Error != nil {
		return errors.New("学生不在该班级")
	}

	// 清空学生的班级ID
	result := repository.DB.Model(&model.User{}).Where("id = ?", studentID).Updates(map[string]interface{}{
		"class_id": nil,
	})
	return result.Error
}
