package service

import (
	"GoCodeMentor/internal/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAssignmentRepository 模拟作业仓储
type MockAssignmentRepository struct {
	mock.Mock
}

func (m *MockAssignmentRepository) Create(assign *model.Assignment) error {
	args := m.Called(assign)
	return args.Error(0)
}

func (m *MockAssignmentRepository) GetByID(id string) (*model.Assignment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Assignment), args.Error(1)
}

func (m *MockAssignmentRepository) GetByTeacherID(teacherID string) ([]model.Assignment, error) {
	args := m.Called(teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Assignment), args.Error(1)
}

func (m *MockAssignmentRepository) GetByClassID(classID string) ([]model.Assignment, error) {
	args := m.Called(classID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Assignment), args.Error(1)
}

func (m *MockAssignmentRepository) Update(assign *model.Assignment) error {
	args := m.Called(assign)
	return args.Error(0)
}

func (m *MockAssignmentRepository) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockAssignmentClassRepository 模拟作业-班级关联仓储
type MockAssignmentClassRepository struct {
	mock.Mock
}

func (m *MockAssignmentClassRepository) Create(ac *model.AssignmentClass) error {
	args := m.Called(ac)
	return args.Error(0)
}

func (m *MockAssignmentClassRepository) GetByAssignmentAndClass(assignmentID, classID string) (*model.AssignmentClass, error) {
	args := m.Called(assignmentID, classID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AssignmentClass), args.Error(1)
}

func (m *MockAssignmentClassRepository) GetByAssignmentID(assignmentID string) ([]model.AssignmentClass, error) {
	args := m.Called(assignmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.AssignmentClass), args.Error(1)
}

func (m *MockAssignmentClassRepository) Update(ac *model.AssignmentClass) error {
	args := m.Called(ac)
	return args.Error(0)
}

func (m *MockAssignmentClassRepository) DeleteByAssignmentID(assignmentID string) error {
	args := m.Called(assignmentID)
	return args.Error(0)
}

func (m *MockAssignmentClassRepository) DeleteByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockQuestionRepository 模拟题目仓储
type MockQuestionRepository struct {
	mock.Mock
}

func (m *MockQuestionRepository) Create(question *model.Question) error {
	args := m.Called(question)
	return args.Error(0)
}

func (m *MockQuestionRepository) GetByAssignmentID(assignmentID string) ([]model.Question, error) {
	args := m.Called(assignmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Question), args.Error(1)
}

func (m *MockQuestionRepository) DeleteByAssignmentID(assignmentID string) error {
	args := m.Called(assignmentID)
	return args.Error(0)
}

// MockSubmissionRepository 模拟提交仓储
type MockSubmissionRepository struct {
	mock.Mock
}

func (m *MockSubmissionRepository) Create(submission *model.Submission) error {
	args := m.Called(submission)
	return args.Error(0)
}

func (m *MockSubmissionRepository) GetByID(id string) (*model.Submission, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) GetByAssignmentAndStudent(assignmentID, studentID string) (*model.Submission, error) {
	args := m.Called(assignmentID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Submission), args.Error(1)
}

func (m *MockSubmissionRepository) Update(submission *model.Submission) error {
	args := m.Called(submission)
	return args.Error(0)
}

func (m *MockSubmissionRepository) DeleteByAssignmentID(assignmentID string) error {
	args := m.Called(assignmentID)
	return args.Error(0)
}

func (m *MockSubmissionRepository) CountByAssignmentID(assignmentID, status string) (int64, error) {
	args := m.Called(assignmentID, status)
	return args.Get(0).(int64), args.Error(1)
}

// MockUserRepository 模拟用户仓储
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(id string) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// MockClassRepository 模拟班级仓储
type MockClassRepository struct {
	mock.Mock
}

func (m *MockClassRepository) GetByID(id string) (*model.Class, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Class), args.Error(1)
}

// TestAssignmentService_GetPublishedClasses 测试获取已发布班级列表
func TestAssignmentService_GetPublishedClasses(t *testing.T) {
	// 初始化模拟仓储
	mockAssignRepo := new(MockAssignmentRepository)
	mockAssignmentClassRepo := new(MockAssignmentClassRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	mockSubmissionRepo := new(MockSubmissionRepository)
	mockUserRepo := new(MockUserRepository)
	mockClassRepo := new(MockClassRepository)

	// 创建服务
	service := &AssignmentService{
		assignRepo:          mockAssignRepo,
		assignmentClassRepo: mockAssignmentClassRepo,
		questionRepo:        mockQuestionRepo,
		submissionRepo:      mockSubmissionRepo,
		userRepo:            mockUserRepo,
		classRepo:           mockClassRepo,
		siliconFlow:         nil, // 不需要AI客户端
	}

	// 测试数据
	assignmentID := uuid.New().String()
	classID := uuid.New().String()
	className := "测试班级"
	deadline := time.Now().Add(24 * time.Hour)
	releasedAt := time.Now()

	// 设置模拟期望
	mockAssignRepo.On("GetByID", assignmentID).Return(&model.Assignment{
		ID:          assignmentID,
		Title:       "测试作业",
		Description: "测试描述",
		TeacherID:   "teacher-id",
		Type:        "mixed",
		Status:      "published",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	mockQuestionRepo.On("GetByAssignmentID", assignmentID).Return([]model.Question{}, nil)

	mockAssignmentClassRepo.On("GetByAssignmentID", assignmentID).Return([]model.AssignmentClass{
		{
			ID:           uuid.New().String(),
			AssignmentID: assignmentID,
			ClassID:      classID,
			Deadline:     &deadline,
			ReleasedAt:   &releasedAt,
			CreatedAt:    releasedAt,
		},
	}, nil)

	mockClassRepo.On("GetByID", classID).Return(&model.Class{
		ID:        classID,
		Name:      className,
		Code:      "TEST001",
		TeacherID: "teacher-id",
		CreatedAt: releasedAt,
	}, nil)

	// 执行测试
	result, err := service.GetPublishedClasses(assignmentID)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, className, result[0].ClassName)
	assert.Equal(t, assignmentID, result[0].AssignmentID)
	assert.Equal(t, classID, result[0].ClassID)
	assert.Equal(t, &deadline, result[0].Deadline)

	// 验证模拟调用
	mockAssignRepo.AssertExpectations(t)
	mockAssignmentClassRepo.AssertExpectations(t)
	mockClassRepo.AssertExpectations(t)
}

// TestAssignmentService_PublishAssignment 测试发布作业到班级
func TestAssignmentService_PublishAssignment(t *testing.T) {
	// 初始化模拟仓储
	mockAssignRepo := new(MockAssignmentRepository)
	mockAssignmentClassRepo := new(MockAssignmentClassRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	mockSubmissionRepo := new(MockSubmissionRepository)
	mockUserRepo := new(MockUserRepository)
	mockClassRepo := new(MockClassRepository)

	// 创建服务
	service := &AssignmentService{
		assignRepo:          mockAssignRepo,
		assignmentClassRepo: mockAssignmentClassRepo,
		questionRepo:        mockQuestionRepo,
		submissionRepo:      mockSubmissionRepo,
		userRepo:            mockUserRepo,
		classRepo:           mockClassRepo,
		siliconFlow:         nil,
	}

	// 测试数据
	assignmentID := uuid.New().String()
	classID := uuid.New().String()
	deadline := time.Now().Add(24 * time.Hour)

	// 情况1：首次发布（作业为草稿状态）
	mockAssignRepo.On("GetByID", assignmentID).Return(&model.Assignment{
		ID:          assignmentID,
		Title:       "测试作业",
		Description: "测试描述",
		TeacherID:   "teacher-id",
		Type:        "mixed",
		Status:      "draft",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	mockAssignmentClassRepo.On("GetByAssignmentAndClass", assignmentID, classID).Return((*model.AssignmentClass)(nil), nil)
	mockAssignmentClassRepo.On("Create", mock.AnythingOfType("*model.AssignmentClass")).Return(nil)
	mockAssignRepo.On("Update", mock.AnythingOfType("*model.Assignment")).Return(nil)

	// 执行发布
	err := service.PublishAssignment(assignmentID, classID, &deadline)
	assert.NoError(t, err)

	// 情况2：已发布，更新截止时间
	mockAssignRepo.On("GetByID", assignmentID).Return(&model.Assignment{
		ID:          assignmentID,
		Title:       "测试作业",
		Description: "测试描述",
		TeacherID:   "teacher-id",
		Type:        "mixed",
		Status:      "published",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	existingAssignmentClass := &model.AssignmentClass{
		ID:           uuid.New().String(),
		AssignmentID: assignmentID,
		ClassID:      classID,
		Deadline:     &deadline,
		ReleasedAt:   &deadline,
		CreatedAt:    deadline,
	}

	mockAssignmentClassRepo.On("GetByAssignmentAndClass", assignmentID, classID).Return(existingAssignmentClass, nil)
	mockAssignmentClassRepo.On("Update", mock.AnythingOfType("*model.AssignmentClass")).Return(nil)

	// 执行更新
	newDeadline := time.Now().Add(48 * time.Hour)
	err = service.PublishAssignment(assignmentID, classID, &newDeadline)
	assert.NoError(t, err)

	// 验证模拟调用
	mockAssignRepo.AssertExpectations(t)
	mockAssignmentClassRepo.AssertExpectations(t)
}

// TestAssignmentService_DeleteAssignment 测试删除作业
func TestAssignmentService_DeleteAssignment(t *testing.T) {
	// 初始化模拟仓储
	mockAssignRepo := new(MockAssignmentRepository)
	mockAssignmentClassRepo := new(MockAssignmentClassRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	mockSubmissionRepo := new(MockSubmissionRepository)
	mockUserRepo := new(MockUserRepository)
	mockClassRepo := new(MockClassRepository)

	// 创建服务
	service := &AssignmentService{
		assignRepo:          mockAssignRepo,
		assignmentClassRepo: mockAssignmentClassRepo,
		questionRepo:        mockQuestionRepo,
		submissionRepo:      mockSubmissionRepo,
		userRepo:            mockUserRepo,
		classRepo:           mockClassRepo,
		siliconFlow:         nil,
	}

	// 测试数据
	assignmentID := uuid.New().String()

	// 设置模拟期望
	mockSubmissionRepo.On("DeleteByAssignmentID", assignmentID).Return(nil)
	mockQuestionRepo.On("DeleteByAssignmentID", assignmentID).Return(nil)
	mockAssignmentClassRepo.On("DeleteByAssignmentID", assignmentID).Return(nil)
	mockAssignRepo.On("DeleteByID", assignmentID).Return(nil)

	// 执行删除
	err := service.DeleteAssignment(assignmentID)

	// 验证结果
	assert.NoError(t, err)

	// 验证模拟调用
	mockSubmissionRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
	mockAssignmentClassRepo.AssertExpectations(t)
	mockAssignRepo.AssertExpectations(t)
}

// TestAssignmentService_GetAssignmentDetail 测试获取作业详情
func TestAssignmentService_GetAssignmentDetail(t *testing.T) {
	// 初始化模拟仓储
	mockAssignRepo := new(MockAssignmentRepository)
	mockAssignmentClassRepo := new(MockAssignmentClassRepository)
	mockQuestionRepo := new(MockQuestionRepository)
	mockSubmissionRepo := new(MockSubmissionRepository)
	mockUserRepo := new(MockUserRepository)
	mockClassRepo := new(MockClassRepository)

	// 创建服务
	service := &AssignmentService{
		assignRepo:          mockAssignRepo,
		assignmentClassRepo: mockAssignmentClassRepo,
		questionRepo:        mockQuestionRepo,
		submissionRepo:      mockSubmissionRepo,
		userRepo:            mockUserRepo,
		classRepo:           mockClassRepo,
		siliconFlow:         nil,
	}

	// 测试数据
	assignmentID := uuid.New().String()
	teacherID := uuid.New().String()

	// 设置模拟期望
	mockAssignRepo.On("GetByID", assignmentID).Return(&model.Assignment{
		ID:          assignmentID,
		Title:       "测试作业",
		Description: "测试描述",
		TeacherID:   teacherID,
		Type:        "mixed",
		Status:      "published",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil)

	mockQuestionRepo.On("GetByAssignmentID", assignmentID).Return([]model.Question{
		{
			ID:           uuid.New().String(),
			AssignmentID: assignmentID,
			Type:         "choice",
			Content:      "测试题目",
			Answer:       "A",
			Score:        10,
			OrderNum:     1,
		},
	}, nil)

	// 执行获取
	assignment, questions, err := service.GetAssignmentDetail(assignmentID)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, assignment)
	assert.Equal(t, assignmentID, assignment.ID)
	assert.Len(t, questions, 1)
	assert.Equal(t, "测试题目", questions[0].Content)

	// 验证模拟调用
	mockAssignRepo.AssertExpectations(t)
	mockQuestionRepo.AssertExpectations(t)
}
