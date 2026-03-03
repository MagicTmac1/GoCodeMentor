package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/pkg/siliconflow"
	"GoCodeMentor/internal/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type ClassService struct {
	classRepo      repository.ClassRepository
	userRepo       repository.UserRepository
	assignmentRepo repository.AssignmentRepository
	submissionRepo repository.SubmissionRepository
	sfClient       *siliconflow.Client
}

func NewClassService(classRepo repository.ClassRepository, userRepo repository.UserRepository, assignmentRepo repository.AssignmentRepository, submissionRepo repository.SubmissionRepository, sfClient *siliconflow.Client) IClassService {
	rand.Seed(time.Now().UnixNano()) // 全局初始化一次随机数种子
	return &ClassService{
		classRepo:      classRepo,
		userRepo:       userRepo,
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
		sfClient:       sfClient,
	}
}

// generateClassCode 生成6位数字邀请码
func generateClassCode() string {
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
	classes, err := s.classRepo.GetByTeacherID(teacherID)
	if err != nil {
		return nil, err
	}
	if classes == nil {
		return []model.Class{}, nil
	}
	return classes, nil
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

// DeleteClass 删除班级（只有教师可以删除自己创建的班级）
func (s *ClassService) DeleteClass(classID string) error {
	// 首先验证班级是否存在
	if _, err := s.classRepo.GetByID(classID); err != nil {
		return errors.New("班级不存在")
	}

	// 注意：这里我们依赖上层handler已经验证了教师权限
	// 删除班级前，需要先删除所有学生用户记录
	students, err := s.userRepo.GetByClassID(classID)
	if err == nil && len(students) > 0 {
		// 批量删除学生用户记录
		for _, student := range students {
			if err := s.userRepo.Delete(&student); err != nil {
				// 如果删除失败，记录错误但继续尝试
				fmt.Printf("删除学生%s失败: %v\n", student.ID, err)
			}
		}
	}

	// 删除班级记录
	return s.classRepo.Delete(classID)
}

// GenerateClassAnalysisReport handles generating an AI-powered academic analysis for a class.
func (s *ClassService) GenerateClassAnalysisReport(ctx context.Context, classID string) (string, error) {
	// 1. 获取班级、学生和作业信息
	class, err := s.classRepo.GetByID(classID)
	if err != nil {
		return "", errors.New("班级不存在")
	}

	students, err := s.userRepo.GetByClassID(classID)
	if err != nil {
		return "", fmt.Errorf("获取学生列表失败: %w", err)
	}

	assignments, err := s.assignmentRepo.GetByClassID(classID)
	if err != nil {
		return "", fmt.Errorf("获取作业列表失败: %w", err)
	}

	if len(assignments) == 0 {
		return "", errors.New("该班级还没有任何已发布的作业，无法生成学情分析报告")
	}

	// 2. 获取所有相关提交记录
	assignmentIDs := make([]string, len(assignments))
	for i, a := range assignments {
		assignmentIDs[i] = a.ID
	}
	
	submissions, err := s.submissionRepo.GetByAssignmentIDs(assignmentIDs)
	if err != nil {
		return "", fmt.Errorf("获取提交记录失败: %w", err)
	}

	// 3. 构建用于AI分析的数据结构
	type StudentSubmissionInfo struct {
		Name          string   `json:"name"`
		Submitted     []string `json:"submitted"` // 提交的作业标题
		NotSubmitted  []string `json:"not_submitted"` // 未提交的作业标题
	}

	type AnalysisData struct {
		ClassName        string                  `json:"class_name"`
		StudentCount     int                     `json:"student_count"`
		AssignmentTitles []string                `json:"assignment_titles"`
		Submissions      []StudentSubmissionInfo `json:"submissions"`
	}

	// 映射：studentID -> StudentSubmissionInfo
	submissionMap := make(map[string]*StudentSubmissionInfo)
	for _, s := range students {
		submissionMap[s.ID] = &StudentSubmissionInfo{Name: s.Name}
	}

	// 映射：assignmentID -> Title
	assignmentTitleMap := make(map[string]string)
	for _, a := range assignments {
		assignmentTitleMap[a.ID] = a.Title
	}

	// 映射：studentID -> assignmentID -> bool (submitted)
	submittedStatus := make(map[string]map[string]bool)
	for _, sub := range submissions {
		if _, ok := submittedStatus[sub.StudentID]; !ok {
			submittedStatus[sub.StudentID] = make(map[string]bool)
		}
		submittedStatus[sub.StudentID][sub.AssignmentID] = true
	}

	// 填充每个学生的提交和未提交列表
	for studentID, info := range submissionMap {
		for assignID, title := range assignmentTitleMap {
			if submittedStatus[studentID] != nil && submittedStatus[studentID][assignID] {
				info.Submitted = append(info.Submitted, title)
			} else {
				info.NotSubmitted = append(info.NotSubmitted, title)
			}
		}
	}

	analysisSubmissions := make([]StudentSubmissionInfo, 0, len(submissionMap))
	for _, info := range submissionMap {
		analysisSubmissions = append(analysisSubmissions, *info)
	}

	data := AnalysisData{
		ClassName:        class.Name,
		StudentCount:     len(students),
		AssignmentTitles: assignmentTitleMapToList(assignmentTitleMap),
		Submissions:      analysisSubmissions,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化分析数据失败: %w", err)
	}

	// 4. 读取Prompt并调用AI
	systemPrompt, err := ioutil.ReadFile("configs/prompts/class_analysis_system.txt")
	if err != nil {
		return "", fmt.Errorf("读取系统Prompt失败: %w", err)
	}

	userPrompt := fmt.Sprintf("请为以下班级生成学情分析报告：\n\n%s", string(jsonData))

	messages := []siliconflow.Message{
		{Role: "system", Content: string(systemPrompt)},
		{Role: "user", Content: userPrompt},
	}

	report, err := s.sfClient.ChatWithHistory(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("AI生成报告失败: %w", err)
	}

	return report, nil
}

func assignmentTitleMapToList(m map[string]string) []string {
	list := make([]string, 0, len(m))
	for _, title := range m {
		list = append(list, title)
	}
	return list
}
