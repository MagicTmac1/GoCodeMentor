package service

import (
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/pkg/siliconflow"
	"GoCodeMentor/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AssignmentService 作业服务
type AssignmentService struct {
	assignRepo          repository.AssignmentRepository
	assignmentClassRepo repository.AssignmentClassRepository
	questionRepo        repository.QuestionRepository
	submissionRepo      repository.SubmissionRepository
	userRepo            repository.UserRepository
	classRepo           repository.ClassRepository
	siliconFlow         *siliconflow.Client
}

// NewAssignmentService 创建作业服务
func NewAssignmentService(
	assignRepo repository.AssignmentRepository,
	assignmentClassRepo repository.AssignmentClassRepository,
	questionRepo repository.QuestionRepository,
	submissionRepo repository.SubmissionRepository,
	userRepo repository.UserRepository,
	classRepo repository.ClassRepository,
	siliconFlow *siliconflow.Client,
) *AssignmentService {
	return &AssignmentService{
		assignRepo:          assignRepo,
		assignmentClassRepo: assignmentClassRepo,
		questionRepo:        questionRepo,
		submissionRepo:      submissionRepo,
		userRepo:            userRepo,
		classRepo:           classRepo,
		siliconFlow:         siliconFlow,
	}
}

// GenerateAssignmentByAI 通过AI生成作业
func (s *AssignmentService) GenerateAssignmentByAI(ctx context.Context, topic string, difficulty string, teacherID string) (*model.Assignment, error) {
	// 调用硅基流动API生成作业内容
	systemPrompt := `你是一个严谨的编程作业生成器。
你必须严格遵守以下格式要求，且只能输出合法的 JSON 数据。
严禁输出任何 Markdown 代码块标签（如 ` + "```" + `json）、严禁输出任何解释性文字、严禁输出 JSON 之外的任何内容。`

	userPrompt := fmt.Sprintf(`请生成一个关于 %s 的编程作业，难度级别：%s。

### 必须严格遵守的 JSON 结构：
{
  "title": "作业标题",
  "description": "作业描述",
  "questions": [
    {
      "type": "choice",
      "content": "题目内容",
      "options": ["A. 选项1", "B. 选项2", "C. 选项3", "D. 选项4"],
      "answer": "A",
      "score": 10
    },
    {
      "type": "fill",
      "content": "题目内容，使用 ___ 表示填空位置",
      "answer": "标准答案",
      "score": 10
    },
    {
      "type": "code",
      "content": "编程题要求描述",
      "answer": "这里必须是完整的、可运行的 Go 代码实现",
      "score": 40
    }
  ]
}

### 要求：
1. 包含 3-5 个题目。
2. 编程题的 answer 字段必须包含真实可用的 Go 代码。
3. 只能返回上述 JSON 对象，不要有任何前缀或后缀文字。`, topic, difficulty)

	messages := []siliconflow.Message{
		{Role: "system", Content: systemPrompt},
	}

	response, err := s.siliconFlow.ChatCompletion(ctx, userPrompt, messages)
	if err != nil {
		return nil, fmt.Errorf("调用AI接口失败: %w", err)
	}

	// 记录原始响应以便排查
	log.Printf("[AI生成作业] 原始响应: %s", response)

	// 解析AI返回的JSON
	var aiResponse struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Questions   []struct {
			Type    string      `json:"type"`
			Content string      `json:"content"`
			Options interface{} `json:"options"`
			Answer  string      `json:"answer"`
			Score   int         `json:"score"`
		} `json:"questions"`
	}

	cleanJSON := s.cleanAIJSON(response)
	log.Printf("[AI生成作业] 清理后的JSON: %s", cleanJSON)

	if cleanJSON == "" || !json.Valid([]byte(cleanJSON)) {
		return nil, fmt.Errorf("AI返回的内容不是有效的JSON。清理后的内容: %s", cleanJSON)
	}

	if err := json.Unmarshal([]byte(cleanJSON), &aiResponse); err != nil {
		return nil, fmt.Errorf("解析AI响应失败: %w。清理后的内容: %s", err, cleanJSON)
	}

	// 创建作业
	assign := &model.Assignment{
		ID:          uuid.New().String(),
		Title:       aiResponse.Title,
		Description: aiResponse.Description,
		TeacherID:   teacherID,
		Type:        "mixed",
		Status:      "draft",
		Rubric:      "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.assignRepo.Create(assign); err != nil {
		return nil, err
	}

	// 创建题目
	for i, q := range aiResponse.Questions {
		optionsJSON := "{}"
		if q.Options != nil {
			if bytes, err := json.Marshal(q.Options); err == nil {
				optionsJSON = string(bytes)
			}
		}

		question := &model.Question{
			ID:           uuid.New().String(),
			AssignmentID: assign.ID,
			Type:         q.Type,
			Content:      q.Content,
			Answer:       q.Answer,
			Options:      optionsJSON,
			Score:        q.Score,
			OrderNum:     i + 1,
		}
		if err := s.questionRepo.Create(question); err != nil {
			return nil, err
		}
	}

	return assign, nil
}

// GetAssignmentList 获取教师创建的作业列表
func (s *AssignmentService) GetAssignmentList(teacherID string) ([]model.Assignment, error) {
	return s.assignRepo.GetByTeacherID(teacherID)
}

// GetAssignmentsByClass 获取班级的作业列表
func (s *AssignmentService) GetAssignmentsByClass(classID string) ([]model.Assignment, error) {
	return s.assignRepo.GetByClassID(classID)
}

// PublishAssignment 发布作业到班级
func (s *AssignmentService) PublishAssignment(assignID string, classID string, deadline *time.Time) error {
	// 获取作业
	assign, err := s.assignRepo.GetByID(assignID)
	if err != nil {
		return err
	}

	// 检查是否已发布到该班级
	existing, err := s.assignmentClassRepo.GetByAssignmentAndClass(assignID, classID)
	if err == nil && existing != nil {
		// 更新现有的发布记录截止时间
		existing.Deadline = deadline
		return s.assignmentClassRepo.Update(existing)
	}

	// 创建作业-班级关联记录
	now := time.Now()
	assignmentClass := &model.AssignmentClass{
		ID:           uuid.New().String(),
		AssignmentID: assignID,
		ClassID:      classID,
		Deadline:     deadline,
		ReleasedAt:   &now,
		CreatedAt:    now,
	}

	// 如果作业状态是草稿，更新为已发布
	if assign.Status == "draft" {
		assign.Status = "published"
		assign.UpdatedAt = time.Now()
		if err := s.assignRepo.Update(assign); err != nil {
			return err
		}
	}

	return s.assignmentClassRepo.Create(assignmentClass)
}

// GetAssignmentDetail 获取作业详情
func (s *AssignmentService) GetAssignmentDetail(id string) (*model.Assignment, []model.Question, error) {
	assign, err := s.assignRepo.GetByID(id)
	if err != nil {
		return nil, nil, err
	}

	questions, err := s.questionRepo.GetByAssignmentID(id)
	if err != nil {
		return assign, nil, err
	}

	return assign, questions, nil
}

// SubmitAssignment 学生提交作业
func (s *AssignmentService) SubmitAssignment(assignID string, studentID string, studentName string, answers map[string]string, code string) (string, error) {
	// 获取学生信息
	student, err := s.userRepo.GetByID(studentID)
	if err != nil {
		return "", fmt.Errorf("获取学生信息失败: %w", err)
	}

	// 检查学生是否在班级中
	if student.ClassID == nil {
		return "", fmt.Errorf("学生未加入任何班级，无法提交作业")
	}

	// 获取作业信息
	assign, err := s.assignRepo.GetByID(assignID)
	if err != nil {
		return "", fmt.Errorf("获取作业失败: %w", err)
	}

	// 检查作业是否已发布到学生班级
	assignmentClass, err := s.assignmentClassRepo.GetByAssignmentAndClass(assignID, *student.ClassID)
	if err != nil || assignmentClass == nil {
		return "", fmt.Errorf("作业未发布到该班级，无法提交")
	}

	// 检查作业状态（兼容性检查）
	if assign.Status != "published" {
		return "", fmt.Errorf("作业未发布，无法提交")
	}

	// 检查截止时间（使用班级特定的截止时间）
	if assignmentClass.Deadline != nil {
		now := time.Now()
		if now.After(*assignmentClass.Deadline) {
			return "", fmt.Errorf("作业提交已截止，截止时间为: %s", assignmentClass.Deadline.Format("2006-01-02 15:04:05"))
		}
	}

	// 检查是否已经提交过
	existing, err := s.submissionRepo.GetByAssignmentAndStudent(assignID, studentID)
	if err == nil && existing != nil {
		// 更新现有提交
		existing.Answers = answersToString(answers)
		existing.CodeContent = code
		existing.UpdatedAt = time.Now()
		if err := s.submissionRepo.Update(existing); err != nil {
			return "", err
		}
		return existing.ID, nil
	}

	// 创建新的提交
	submission := &model.Submission{
		ID:               uuid.New().String(),
		AssignmentID:     assignID,
		StudentID:        studentID,
		StudentName:      studentName,
		Answers:          answersToString(answers),
		CodeContent:      code,
		QuestionFeedback: "{}",
		QuestionScores:   "{}",
		DetailedScore:    "{}",
		Status:           "submitted",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.submissionRepo.Create(submission); err != nil {
		return "", err
	}

	// 异步进行AI批改
	go func() {
		// 给数据库一点时间完成事务（如果存在）
		time.Sleep(200 * time.Millisecond)
		if err := s.GradeSubmission(context.Background(), submission.ID); err != nil {
			log.Printf("AI批改失败: %v", err)
		}
	}()

	return submission.ID, nil
}

// GetSubmissionByAssignmentAndStudent 获取学生的作业提交
func (s *AssignmentService) GetSubmissionByAssignmentAndStudent(assignID string, studentID string) (*model.Submission, error) {
	return s.submissionRepo.GetByAssignmentAndStudent(assignID, studentID)
}

// GetAllAssignments 获取所有作业
func (s *AssignmentService) GetAllAssignments() ([]model.Assignment, error) {
	// 这里实现获取所有作业的逻辑
	// 由于AssignmentRepository没有GetAll方法，我们可以通过其他方式获取
	// 暂时返回空数组
	return []model.Assignment{}, nil
}

// GetSubmission 获取提交记录
func (s *AssignmentService) GetSubmission(subID string) (*model.Submission, error) {
	return s.submissionRepo.GetByID(subID)
}

// GetPendingSubmissionCountByAssignment 获取待批改的提交数量
func (s *AssignmentService) GetPendingSubmissionCountByAssignment(assignmentID string) (int64, error) {
	return s.submissionRepo.CountByAssignmentID(assignmentID, "submitted")
}

// GradeSubmission AI批改作业
func (s *AssignmentService) GradeSubmission(ctx context.Context, submissionID string) error {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return err
	}

	assign, questions, err := s.GetAssignmentDetail(submission.AssignmentID)
	if err != nil {
		return err
	}

	// 构建批改提示
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`你是一位冷酷无情且极其严谨的编程考官。你的任务是核对学生的作业答案并给出分数。

### 规则库（必须死板地执行）：
1. **完全匹配原则**：对于选择题和填空题，学生答案必须与标准答案在字符级别完全一致。任何微小的差别（如多一个空格、大小写不一、或者填了无关内容如"1"）都必须判为 0 分。
2. **拒绝同情分**：如果学生答案错误，或者编程题代码无法运行、逻辑不通、或是填写的与题目无关（如 "1"、"不知道"、"..."），该题得分必须为 0。严禁给任何形式的辛苦分。
3. **负面示例参考**：
   - 题目：Go 的并发原语是什么？ 标准答案：goroutine
   - 学生答案：1  => 判定：错误，得分：0
   - 学生答案：不知道 => 判定：错误，得分：0
   - 学生答案：Goroutine => 判定：错误（大小写不一致），得分：0

### 评分数据源：
`)

	promptBuilder.WriteString(fmt.Sprintf("\n[作业上下文]\n标题：%s\n描述：%s\n\n", assign.Title, assign.Description))

	promptBuilder.WriteString("[标准答案库]\n")
	for i, q := range questions {
		promptBuilder.WriteString(fmt.Sprintf("Q%d (ID: %s) | 类型: %s | 满分: %d | 标准答案: %s\n",
			i+1, q.ID, q.Type, q.Score, q.Answer))
	}

	// 添加学生答案
	var answers map[string]string
	if err := json.Unmarshal([]byte(submission.Answers), &answers); err == nil {
		promptBuilder.WriteString("\n[学生提交内容]\n")
		for i, q := range questions {
			studentAns := answers[q.ID]
			if studentAns == "" {
				studentAns = "[未回答]"
			}
			promptBuilder.WriteString(fmt.Sprintf("Q%d (ID: %s) 学生答案: %s\n", i+1, q.ID, studentAns))
		}
	}

	if submission.CodeContent != "" {
		promptBuilder.WriteString(fmt.Sprintf("\n[学生编程代码]\n%s\n", submission.CodeContent))
	}

	promptBuilder.WriteString(`
### 执行指令：
1. 逐一比对 [学生提交内容] 与 [标准答案库]。
2. 计算总分 (total_score)，确保它等于所有单题得分的数学总和。
3. 生成 ai_feedback，必须包含一个 Markdown 表格展示每题的得分情况，随后进行毒舌但客观的评价。

直接返回 JSON 格式，不要包含 Markdown 代码块标记：
{
  "total_score": 整数,
  "ai_feedback": "Markdown 报告",
  "question_scores": {"题目ID": 分数, ...},
  "question_feedback": {"题目ID": "为什么给这个分", ...}
}
`)

	// 记录生成的 Prompt 以便调试
	log.Printf("--- AI Grading Prompt ---\n%s\n-----------------------", promptBuilder.String())

	// 调用AI批改
	response, err := s.siliconFlow.ChatCompletion(ctx, promptBuilder.String(), nil)
	if err != nil {
		return err
	}

	// 解析 AI 返回的 JSON
	var gradeResult struct {
		TotalScore       int               `json:"total_score"`
		AIFeedback       string            `json:"ai_feedback"`
		QuestionScores   map[string]int    `json:"question_scores"`
		QuestionFeedback map[string]string `json:"question_feedback"`
	}

	cleanJSON := s.cleanAIJSON(response)

	if err := json.Unmarshal([]byte(cleanJSON), &gradeResult); err != nil {
		log.Printf("解析AI批改JSON失败，降级为纯文本处理: %v\n原始响应: %s", err, response)
		// 降级处理：如果解析 JSON 失败，将整个响应作为评语，给个默认分
		submission.AIFeedback = response
		defaultScore := 0
		submission.TotalScore = &defaultScore
	} else {
		submission.TotalScore = &gradeResult.TotalScore
		submission.AIFeedback = gradeResult.AIFeedback

		// 序列化题目分数和批注
		if scoresJSON, err := json.Marshal(gradeResult.QuestionScores); err == nil {
			submission.QuestionScores = string(scoresJSON)
		}
		if feedbackJSON, err := json.Marshal(gradeResult.QuestionFeedback); err == nil {
			submission.QuestionFeedback = string(feedbackJSON)
		}
	}

	submission.Status = "graded"
	submission.UpdatedAt = time.Now()

	return s.submissionRepo.Update(submission)
}

// 辅助函数：将答案map转换为JSON字符串
func answersToString(answers map[string]string) string {
	if len(answers) == 0 {
		return "{}"
	}
	bytes, err := json.Marshal(answers)
	if err != nil {
		log.Printf("Error marshaling answers: %v", err)
		return "{}"
	}
	return string(bytes)
}

// 辅助函数：将JSON字符串转换为答案map
func stringToAnswers(s string) (map[string]string, error) {
	var answers map[string]string
	err := json.Unmarshal([]byte(s), &answers)
	return answers, err
}

// UpdateSubmissionScore 手动更新提交分数
func (s *AssignmentService) UpdateSubmissionScore(submissionID string, score int) error {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return fmt.Errorf("获取提交记录失败: %w", err)
	}

	submission.TotalScore = &score
	submission.UpdatedAt = time.Now()

	return s.submissionRepo.Update(submission)
}

// UpdateTeacherFeedback 更新教师批注
func (s *AssignmentService) UpdateTeacherFeedback(submissionID string, feedback string) error {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return fmt.Errorf("获取提交记录失败: %w", err)
	}

	submission.TeacherFeedback = feedback
	submission.UpdatedAt = time.Now()

	return s.submissionRepo.Update(submission)
}

// UpdateQuestionScore 更新单个题目的分数
func (s *AssignmentService) UpdateQuestionScore(submissionID string, questionID string, score int, maxScore int) error {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return fmt.Errorf("获取提交记录失败: %w", err)
	}

	// 验证分数范围
	if score < 0 || score > maxScore {
		return fmt.Errorf("分数必须在0-%d之间", maxScore)
	}

	// 解析现有的题目分数
	var questionScores map[string]int
	if submission.QuestionScores != "" {
		if err := json.Unmarshal([]byte(submission.QuestionScores), &questionScores); err != nil {
			questionScores = make(map[string]int)
		}
	} else {
		questionScores = make(map[string]int)
	}

	// 更新分数
	questionScores[questionID] = score

	// 转换为JSON
	scoresJSON, err := json.Marshal(questionScores)
	if err != nil {
		return fmt.Errorf("序列化分数失败: %w", err)
	}

	submission.QuestionScores = string(scoresJSON)
	submission.UpdatedAt = time.Now()

	// 重新计算总分
	totalScore := 0
	for _, s := range questionScores {
		totalScore += s
	}
	submission.TotalScore = &totalScore

	return s.submissionRepo.Update(submission)
}

// UpdateQuestionFeedback 更新单个题目的批注
func (s *AssignmentService) UpdateQuestionFeedback(submissionID string, questionID string, feedback string) error {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return fmt.Errorf("获取提交记录失败: %w", err)
	}

	// 解析现有的题目批注
	var questionFeedback map[string]string
	if submission.QuestionFeedback != "" {
		if err := json.Unmarshal([]byte(submission.QuestionFeedback), &questionFeedback); err != nil {
			questionFeedback = make(map[string]string)
		}
	} else {
		questionFeedback = make(map[string]string)
	}

	// 更新批注
	questionFeedback[questionID] = feedback

	// 转换为JSON
	feedbackJSON, err := json.Marshal(questionFeedback)
	if err != nil {
		return fmt.Errorf("序列化批注失败: %w", err)
	}

	submission.QuestionFeedback = string(feedbackJSON)
	submission.UpdatedAt = time.Now()

	return s.submissionRepo.Update(submission)
}

// GetQuestionScore 获取单个题目的分数
func (s *AssignmentService) GetQuestionScore(submissionID string, questionID string) (int, error) {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return 0, fmt.Errorf("获取提交记录失败: %w", err)
	}

	if submission.QuestionScores == "" {
		return 0, nil
	}

	var questionScores map[string]int
	if err := json.Unmarshal([]byte(submission.QuestionScores), &questionScores); err != nil {
		return 0, fmt.Errorf("解析题目分数失败: %w", err)
	}

	score, exists := questionScores[questionID]
	if !exists {
		return 0, nil
	}

	return score, nil
}

// GetQuestionFeedback 获取单个题目的批注
func (s *AssignmentService) GetQuestionFeedback(submissionID string, questionID string) (string, error) {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return "", fmt.Errorf("获取提交记录失败: %w", err)
	}

	if submission.QuestionFeedback == "" {
		return "", nil
	}

	var questionFeedback map[string]string
	if err := json.Unmarshal([]byte(submission.QuestionFeedback), &questionFeedback); err != nil {
		return "", fmt.Errorf("解析题目批注失败: %w", err)
	}

	feedback, exists := questionFeedback[questionID]
	if !exists {
		return "", nil
	}

	return feedback, nil
}

// RegradeSubmission 重新触发AI批改
func (s *AssignmentService) RegradeSubmission(submissionID string) error {
	// 异步重新批改
	go func() {
		ctx := context.Background()
		if err := s.GradeSubmission(ctx, submissionID); err != nil {
			log.Printf("重新批改失败: %v", err)
		}
	}()

	return nil
}

// GetSubmissionCodeForDownload 获取要下载的代码内容和文件名
func (s *AssignmentService) GetSubmissionCodeForDownload(submissionID string) (string, string, error) {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return "", "", fmt.Errorf("获取提交记录失败: %w", err)
	}

	if submission.CodeContent == "" {
		return "", "", fmt.Errorf("学生未提交代码")
	}

	// 生成文件名：学生姓名-作业ID-提交时间.go
	fileName := fmt.Sprintf("%s-%s-%s.go",
		submission.StudentName,
		submission.AssignmentID[:8],
		submission.CreatedAt.Format("20060102-150405"))

	return submission.CodeContent, fileName, nil
}

// GetPublishedClasses 获取作业发布到的班级列表（包含班级名称）
func (s *AssignmentService) GetPublishedClasses(assignID string) ([]model.AssignmentClassWithClassName, error) {
	assignmentClasses, err := s.assignmentClassRepo.GetByAssignmentID(assignID)
	if err != nil {
		return nil, err
	}

	result := []model.AssignmentClassWithClassName{}
	for _, ac := range assignmentClasses {
		class, err := s.classRepo.GetByID(ac.ClassID)
		if err != nil {
			continue // 如果班级不存在，跳过
		}

		result = append(result, model.AssignmentClassWithClassName{
			AssignmentClass: model.AssignmentClass{
				ID:           ac.ID,
				AssignmentID: ac.AssignmentID,
				ClassID:      ac.ClassID,
				Deadline:     ac.Deadline,
				ReleasedAt:   ac.ReleasedAt,
				CreatedAt:    ac.CreatedAt,
			},
			ClassName: class.Name,
		})
	}

	return result, nil
}

// DeleteAssignment 删除作业
func (s *AssignmentService) DeleteAssignment(assignID string) error {
	// 由于仓储层没有暴露db字段，我们直接调用各个仓储的删除方法
	// 注意：这里没有使用事务，但应该足够安全，因为各个删除操作是独立的
	// 首先删除所有学生提交记录
	if err := s.submissionRepo.DeleteByAssignmentID(assignID); err != nil {
		return fmt.Errorf("删除提交记录失败: %w", err)
	}

	// 删除所有题目
	if err := s.questionRepo.DeleteByAssignmentID(assignID); err != nil {
		return fmt.Errorf("删除题目失败: %w", err)
	}

	// 删除作业与班级的关联
	if err := s.assignmentClassRepo.DeleteByAssignmentID(assignID); err != nil {
		return fmt.Errorf("删除作业-班级关联失败: %w", err)
	}

	// 删除作业本身
	return s.assignRepo.DeleteByID(assignID)
}

// GetAssignmentCorrectionDetail 获取作业批改详情（教师查看）
func (s *AssignmentService) GetAssignmentCorrectionDetail(assignID string, studentID string) (*model.Assignment, []model.Question, *model.Submission, error) {
	// 获取作业详情
	assign, questions, err := s.GetAssignmentDetail(assignID)
	if err != nil {
		return nil, nil, nil, err
	}

	// 获取学生提交
	submission, err := s.submissionRepo.GetByAssignmentAndStudent(assignID, studentID)
	if err != nil && err.Error() != "record not found" {
		return nil, nil, nil, err
	}

	return assign, questions, submission, nil
}

// cleanAIJSON 尝试清理可能的 Markdown 标记，提取纯 JSON 字符串
func (s *AssignmentService) cleanAIJSON(response string) string {
	// 1. 尝试直接清理常见的 Markdown 块标记
	clean := strings.TrimSpace(response)

	// 处理嵌套或多次包裹的情况
	for {
		if strings.HasPrefix(clean, "```") {
			// 移除 ```json 或 ```
			firstLineEnd := strings.Index(clean, "\n")
			if firstLineEnd != -1 {
				clean = strings.TrimSpace(clean[firstLineEnd:])
			} else {
				clean = strings.TrimPrefix(clean, "```json")
				clean = strings.TrimPrefix(clean, "```")
			}
			clean = strings.TrimSuffix(clean, "```")
			clean = strings.TrimSpace(clean)
			continue
		}
		break
	}

	// 2. 如果清理后还是不合法，尝试通过定位 { } 来提取
	if !json.Valid([]byte(clean)) {
		start := strings.Index(clean, "{")
		end := strings.LastIndex(clean, "}")
		if start != -1 && end != -1 && end > start {
			extracted := clean[start : end+1]
			if json.Valid([]byte(extracted)) {
				return extracted
			}
		}
	}

	return clean
}
