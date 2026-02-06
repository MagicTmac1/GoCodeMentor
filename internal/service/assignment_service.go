package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/pkg/siliconflow"
	"GoCodeMentor/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type AssignmentService struct {
	client *siliconflow.Client
}

func NewAssignmentService(client *siliconflow.Client) *AssignmentService {
	return &AssignmentService{client: client}
}

// GenerateAssignmentByAI AI生成作业
func (s *AssignmentService) GenerateAssignmentByAI(ctx context.Context, topic, difficulty, teacherID string) (*model.Assignment, error) {
	// 构造Prompt让AI生成作业
	prompt := fmt.Sprintf(`请作为一位Go语言教师，根据以下要求生成一份作业：

知识点：%s
难度：%s（初级/中级/高级）

请生成以下内容（JSON格式）：
{
  "title": "作业标题",
  "description": "作业描述和要求",
  "type": "mixed",
  "rubric": {"code_style": 30, "functionality": 50, "documentation": 20},
  "questions": [
    {
      "type": "choice",
      "content": "选择题题目",
      "options": ["A. xxx", "B. xxx", "C. xxx", "D. xxx"],
      "answer": "A",
      "score": 10
    },
    {
      "type": "fill",
      "content": "填空题，用____表示填空位置",
      "answer": "正确答案",
      "score": 10
    },
    {
      "type": "code",
      "content": "编程题要求",
      "answer": "参考实现思路",
      "score": 80
    }
  ]
}

要求：
1. 包含2-3道选择题（考察基础概念）
2. 包含1-2道填空题（考察语法细节）
3. 包含1道编程题（考察综合运用）
4. 总分100分
5. 只返回JSON，不要有其他文字`, topic, difficulty)

	systemPrompt := "你是一位专业的Go语言教师，擅长设计编程练习题。"
	response, err := s.client.Chat(ctx, systemPrompt, prompt)
	if err != nil {
		return nil, err
	}

	// 清理AI返回的内容（去除可能的Markdown代码块标记）
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// 解析AI返回的JSON
	var result struct {
		Title       string         `json:"title"`
		Description string         `json:"description"`
		Type        string         `json:"type"`
		Rubric      map[string]int `json:"rubric"`
		Questions   []struct {
			Type    string   `json:"type"`
			Content string   `json:"content"`
			Options []string `json:"options"`
			Answer  string   `json:"answer"`
			Score   int      `json:"score"`
		} `json:"questions"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("解析AI生成内容失败: %v, 原始内容: %s", err, response[:min(len(response), 200)])
	}

	// 创建作业
	assignID := uuid.New().String()
	rubricJSON, _ := json.Marshal(result.Rubric)

	assign := model.Assignment{
		ID:          assignID,
		Title:       result.Title,
		Description: result.Description,
		TeacherID:   teacherID,
		Rubric:      string(rubricJSON),
		Type:        result.Type,
		Status:      "draft",
	}

	if err := repository.DB.Create(&assign).Error; err != nil {
		return nil, err
	}

	// 创建题目
	for i, q := range result.Questions {
		optionsJSON, _ := json.Marshal(q.Options)
		question := model.Question{
			ID:           uuid.New().String(),
			AssignmentID: assignID,
			Type:         q.Type,
			Content:      q.Content,
			Options:      string(optionsJSON),
			Answer:       q.Answer,
			Score:        q.Score,
			OrderNum:     i + 1,
		}
		repository.DB.Create(&question)
	}

	return &assign, nil
}

// GetAssignmentList 获取教师的作业列表（包含草稿和已发布）
func (s *AssignmentService) GetAssignmentList(teacherID string) ([]model.Assignment, error) {
	var assignments []model.Assignment
	// 查询当前教师的作业，以及teacher_id为空的作业（兼容旧数据）
	result := repository.DB.Where("teacher_id = ? OR teacher_id = ? OR teacher_id IS NULL", teacherID, "").Order("created_at desc").Find(&assignments)
	return assignments, result.Error
}

// GetAllAssignments 别名兼容
func (s *AssignmentService) GetAllAssignments() ([]model.Assignment, error) {
	return s.GetAssignmentList("")
}

// GetAssignmentsByClass 获取发布到指定班级的作业
func (s *AssignmentService) GetAssignmentsByClass(classID string) ([]model.Assignment, error) {
	var assignments []model.Assignment
	result := repository.DB.Where("class_id = ? AND status = ?", classID, "published").Order("created_at desc").Find(&assignments)
	return assignments, result.Error
}

// PublishAssignment 发布作业到班级
func (s *AssignmentService) PublishAssignment(assignID, classID string) error {
	// 更新作业状态为已发布
	result := repository.DB.Model(&model.Assignment{}).Where("id = ?", assignID).Updates(map[string]interface{}{
		"status":   "published",
		"class_id": classID,
	})
	return result.Error
}

// GetAssignmentDetail 获取作业详情（包含题目）
func (s *AssignmentService) GetAssignmentDetail(assignID string) (*model.Assignment, []model.Question, error) {
	var assign model.Assignment
	if err := repository.DB.First(&assign, "id = ?", assignID).Error; err != nil {
		return nil, nil, err
	}

	var questions []model.Question
	repository.DB.Where("assignment_id = ?", assignID).Order("order_num asc").Find(&questions)

	return &assign, questions, nil
}

// SubmitAssignment 提交作业
func (s *AssignmentService) SubmitAssignment(assignID, studentID, studentName string, answers map[string]string, code string) (string, error) {
	id := uuid.New().String()
	answersJSON, _ := json.Marshal(answers)

	sub := model.Submission{
		ID:           id,
		AssignmentID: assignID,
		StudentID:    studentID,
		StudentName:  studentName,
		Answers:      string(answersJSON),
		CodeContent:  code,
		Status:       "submitted",
	}

	return id, repository.DB.Create(&sub).Error
}

// GradeSubmission AI批改作业
func (s *AssignmentService) GradeSubmission(ctx context.Context, subID string) error {
	// 查询提交
	var sub model.Submission
	if err := repository.DB.First(&sub, "id = ?", subID).Error; err != nil {
		return err
	}

	// 查询作业信息
	var assign model.Assignment
	if err := repository.DB.First(&assign, "id = ?", sub.AssignmentID).Error; err != nil {
		return err
	}

	// 查询题目
	var questions []model.Question
	repository.DB.Where("assignment_id = ?", sub.AssignmentID).Find(&questions)

	// 构建批改Prompt
	prompt := fmt.Sprintf(`请作为严格的编程教师，批改以下作业：

作业要求：%s
评分标准：%s

学生提交：
`, assign.Description, assign.Rubric)

	// 添加各题答案
	var answerMap map[string]string
	json.Unmarshal([]byte(sub.Answers), &answerMap)

	for _, q := range questions {
		prompt += fmt.Sprintf("\n题目[%s]: %s\n学生答案: %s\n参考答案: %s\n",
			q.Type, q.Content, answerMap[q.ID], q.Answer)
	}

	if sub.CodeContent != "" {
		prompt += fmt.Sprintf("\n编程题代码：\n```go\n%s\n```", sub.CodeContent)
	}

	prompt += `\n\n请按以下JSON格式返回评分结果：
{
  "total_score": 85,
  "detailed_score": {"选择题": 20, "填空题": 15, "编程题": 50},
  "feedback": "总体评语，指出主要问题和优点",
  "suggestions": ["具体改进建议1", "建议2"]
}

只返回JSON，不要有其他文字。`

	systemPrompt := "你是一位严格的Go语言教师，批改作业时既指出问题又给予鼓励。"
	response, err := s.client.Chat(ctx, systemPrompt, prompt)
	if err != nil {
		return err
	}

	// 解析评分结果
	var result struct {
		TotalScore    int            `json:"total_score"`
		DetailedScore map[string]int `json:"detailed_score"`
		Feedback      string         `json:"feedback"`
		Suggestions   []string       `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果解析失败，保存原始文本
		sub.AIFeedback = "AI返回格式错误，原始内容：\n" + response
		sub.Status = "graded"
		repository.DB.Save(&sub)
		return nil
	}

	// 保存评分结果
	detailedJSON, _ := json.Marshal(result.DetailedScore)
	sub.TotalScore = &result.TotalScore
	sub.DetailedScore = string(detailedJSON)
	sub.AIFeedback = result.Feedback + "\n\n改进建议：\n- " + joinStrings(result.Suggestions, "\n- ")
	sub.Status = "graded"

	return repository.DB.Save(&sub).Error
}

// GetSubmission 获取提交详情
func (s *AssignmentService) GetSubmission(subID string) (*model.Submission, error) {
	var sub model.Submission
	result := repository.DB.First(&sub, "id = ?", subID)
	return &sub, result.Error
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
