package handler

import (
	"GoCodeMentor/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// AssignmentHandler handles assignment-related requests.
type AssignmentHandler struct {
	assignSvc service.IAssignmentService
	userSvc   service.IUserService
}

// NewAssignmentHandler creates a new AssignmentHandler.
func NewAssignmentHandler(assignSvc service.IAssignmentService, userSvc service.IUserService) *AssignmentHandler {
	return &AssignmentHandler{assignSvc: assignSvc, userSvc: userSvc}
}

// GenerateAssignmentByAI handles the generation of an assignment by AI.
func (h *AssignmentHandler) GenerateAssignmentByAI(c *gin.Context) {
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以生成作业"})
		return
	}

	var req struct {
		Topic      string `json:"topic"`
		Difficulty string `json:"difficulty"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	ctx := context.Background()
	assign, err := h.assignSvc.GenerateAssignmentByAI(ctx, req.Topic, req.Difficulty, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, assign)
}

// GetAssignments handles getting a list of assignments.
func (h *AssignmentHandler) GetAssignments(c *gin.Context) {
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userRole == "teacher" {
		assignments, err := h.assignSvc.GetAssignmentList(userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, assignments)
	} else {
		user, err := h.userSvc.GetByID(userID)
		if err != nil {
			c.JSON(200, []interface{}{})
			return
		}
		if user.ClassID == nil {
			c.JSON(200, []interface{}{})
			return
		}
		assignments, err := h.assignSvc.GetAssignmentsByClass(*user.ClassID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, assignments)
	}
}

// PublishAssignment handles publishing an assignment to a class.
func (h *AssignmentHandler) PublishAssignment(c *gin.Context) {
	assignID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以发布作业"})
		return
	}

	var req struct {
		ClassID  string `json:"class_id"`
		Deadline string `json:"deadline"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if req.ClassID == "" {
		c.JSON(400, gin.H{"error": "班级ID不能为空"})
		return
	}

	students, err := h.userSvc.GetStudentsByClassID(req.ClassID)
	if err != nil {
		c.JSON(500, gin.H{"error": "检查班级学生失败"})
		return
	}

	if len(students) == 0 {
		c.JSON(400, gin.H{"error": "该班级暂无学生，无法发布作业"})
		return
	}

	// 处理截止时间（必填）
	if req.Deadline == "" {
		c.JSON(400, gin.H{"error": "截止时间不能为空"})
		return
	}

	deadlineTime, err := time.Parse("2006-01-02", req.Deadline)
	if err != nil {
		c.JSON(400, gin.H{"error": "截止时间格式错误，请使用YYYY-MM-DD格式"})
		return
	}

	// 检查截止时间是否早于当前时间
	now := time.Now()
	if deadlineTime.Before(now) {
		c.JSON(400, gin.H{"error": "截止时间不能早于当前时间"})
		return
	}

	// 将时间设置为当天的23:59:59
	endOfDay := time.Date(deadlineTime.Year(), deadlineTime.Month(), deadlineTime.Day(), 23, 59, 59, 0, deadlineTime.Location())
	deadline := &endOfDay

	err = h.assignSvc.PublishAssignment(assignID, req.ClassID, deadline)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "发布成功"})
}

// GetAssignmentDetail handles getting the details of an assignment.
func (h *AssignmentHandler) GetAssignmentDetail(c *gin.Context) {
	id := c.Param("id")
	assign, questions, err := h.assignSvc.GetAssignmentDetail(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "作业不存在"})
		return
	}
	c.JSON(200, gin.H{
		"assignment": assign,
		"questions":  questions,
	})
}

// GetAssignmentQRCode handles generating a QR code for an assignment.
func (h *AssignmentHandler) GetAssignmentQRCode(c *gin.Context) {
	id := c.Param("id")
	url := fmt.Sprintf("http://localhost:8081/assignments/do?id=%s", id)

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		c.JSON(500, gin.H{"error": "生成二维码失败"})
		return
	}

	c.Data(200, "image/png", png)
}

// SubmitAssignment handles a student submitting an assignment.
func (h *AssignmentHandler) SubmitAssignment(c *gin.Context) {
	assignID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "student" {
		c.JSON(403, gin.H{"error": "只有学生可以提交作业"})
		return
	}

	var req struct {
		StudentName string                 `json:"student_name"`
		Answers     map[string]interface{} `json:"answers"`
		Code        string                 `json:"code"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	answers := make(map[string]string)
	for k, v := range req.Answers {
		if str, ok := v.(string); ok {
			answers[k] = str
		} else {
			answers[k] = fmt.Sprintf("%v", v)
		}
	}

	// 检查作业是否已过期
	assign, _, err := h.assignSvc.GetAssignmentDetail(assignID)
	if err != nil {
		c.JSON(404, gin.H{"error": "作业不存在"})
		return
	}

	// 检查截止时间
	if assign.Deadline != nil {
		now := time.Now()
		if now.After(*assign.Deadline) {
			c.JSON(400, gin.H{"error": "作业提交已截止，无法提交"})
			return
		}
	}

	submissionID, err := h.assignSvc.SubmitAssignment(assignID, userID, req.StudentName, answers, req.Code)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"id": submissionID, "message": "提交成功"})
}

// GetStudentAssignments handles getting all assignments for a specific student (teacher view).
func (h *AssignmentHandler) GetStudentAssignments(c *gin.Context) {
	studentID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以查看学生作业"})
		return
	}

	student, err := h.userSvc.GetByID(studentID)
	if err != nil || student.Role != "student" {
		c.JSON(404, gin.H{"error": "学生不存在"})
		return
	}

	if student.ClassID == nil {
		c.JSON(200, gin.H{
			"student":     student,
			"assignments": []interface{}{},
		})
		return
	}

	assignments, err := h.assignSvc.GetAssignmentsByClass(*student.ClassID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var assignmentDetails []gin.H
	for _, assign := range assignments {
		submission, err := h.assignSvc.GetSubmissionByAssignmentAndStudent(assign.ID, studentID)

		var status string
		var submissionInfo gin.H
		if err != nil || submission == nil {
			status = "未提交"
			submissionInfo = gin.H{}
		} else {
			if submission.Status == "graded" {
				status = "已查看"
			} else {
				status = "已提交"
			}
			var answers map[string]interface{}
			var detailedScore map[string]interface{}
			json.Unmarshal([]byte(submission.Answers), &answers)
			json.Unmarshal([]byte(submission.DetailedScore), &detailedScore)

			submissionInfo = gin.H{
				"id":             submission.ID,
				"student_name":   submission.StudentName,
				"answers":        answers,
				"code_content":   submission.CodeContent,
				"total_score":    submission.TotalScore,
				"ai_feedback":    submission.AIFeedback,
				"detailed_score": detailedScore,
				"status":         submission.Status,
				"created_at":     submission.CreatedAt,
				"updated_at":     submission.UpdatedAt,
			}
		}

		assignmentDetails = append(assignmentDetails, gin.H{
			"assignment": assign,
			"status":     status,
			"submission": submissionInfo,
		})
	}

	c.JSON(200, gin.H{
		"student":     student,
		"assignments": assignmentDetails,
	})
}

// GetMyAssignments handles a student getting their own assignments.
func (h *AssignmentHandler) GetMyAssignments(c *gin.Context) {
	studentID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if studentID == "" || userRole != "student" {
		c.JSON(403, gin.H{"error": "无权查看"})
		return
	}

	student, err := h.userSvc.GetByID(studentID)
	if err != nil {
		c.JSON(404, gin.H{"error": "学生不存在"})
		return
	}

	if student.ClassID == nil {
		c.JSON(200, gin.H{
			"student":     student,
			"assignments": []interface{}{},
		})
		return
	}

	assignments, err := h.assignSvc.GetAssignmentsByClass(*student.ClassID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var assignmentDetails []gin.H
	for _, assign := range assignments {
		submission, err := h.assignSvc.GetSubmissionByAssignmentAndStudent(assign.ID, studentID)

		var status string
		var submissionInfo gin.H
		if err != nil || submission == nil {
			status = "未提交"
			submissionInfo = gin.H{}
		} else {
			if submission.Status == "graded" {
				status = "已查看"
			} else {
				status = "已提交"
			}
			var answers map[string]interface{}
			var detailedScore map[string]interface{}
			json.Unmarshal([]byte(submission.Answers), &answers)
			json.Unmarshal([]byte(submission.DetailedScore), &detailedScore)

			submissionInfo = gin.H{
				"id":             submission.ID,
				"student_name":   submission.StudentName,
				"answers":        answers,
				"code_content":   submission.CodeContent,
				"total_score":    submission.TotalScore,
				"ai_feedback":    submission.AIFeedback,
				"detailed_score": detailedScore,
				"status":         submission.Status,
				"created_at":     submission.CreatedAt,
				"updated_at":     submission.UpdatedAt,
			}
		}

		assignmentDetails = append(assignmentDetails, gin.H{
			"assignment": assign,
			"status":     status,
			"submission": submissionInfo,
		})
	}

	c.JSON(200, gin.H{
		"student":     student,
		"assignments": assignmentDetails,
	})
}

// GetAssignmentSubmissionForStudent handles getting a student's submission for an assignment.
func (h *AssignmentHandler) GetAssignmentSubmissionForStudent(c *gin.Context) {
	assignID := c.Param("id")
	studentID := c.Param("studentId")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" {
		c.JSON(401, gin.H{"error": "请先登录"})
		return
	}

	// 权限检查：只有教师或者学生本人可以查看
	if userRole != "teacher" && userID != studentID {
		c.JSON(403, gin.H{"error": "无权查看"})
		return
	}

	assign, questions, err := h.assignSvc.GetAssignmentDetail(assignID)
	if err != nil {
		c.JSON(404, gin.H{"error": "作业不存在"})
		return
	}

	submission, err := h.assignSvc.GetSubmissionByAssignmentAndStudent(assignID, studentID)

	var submissionInfo gin.H
	if err != nil || submission == nil {
		submissionInfo = gin.H{
			"submitted": false,
			"answers":   gin.H{},
			"code":      "",
		}
	} else {
		// 解析 JSON 字段
		var answers map[string]interface{}
		var detailedScore map[string]interface{}
		var questionScores map[string]interface{}
		var questionFeedback map[string]interface{}

		json.Unmarshal([]byte(submission.Answers), &answers)
		json.Unmarshal([]byte(submission.DetailedScore), &detailedScore)
		json.Unmarshal([]byte(submission.QuestionScores), &questionScores)
		json.Unmarshal([]byte(submission.QuestionFeedback), &questionFeedback)

		submissionInfo = gin.H{
			"submitted":         true,
			"student_name":      submission.StudentName,
			"answers":           answers,
			"code":              submission.CodeContent,
			"total_score":       submission.TotalScore,
			"ai_feedback":       submission.AIFeedback,
			"teacher_feedback":  submission.TeacherFeedback,
			"detailed_score":    detailedScore,
			"question_scores":   questionScores,
			"question_feedback": questionFeedback,
			"status":            submission.Status,
			"created_at":        submission.CreatedAt,
			"updated_at":        submission.UpdatedAt,
		}
	}

	c.JSON(200, gin.H{
		"assignment": assign,
		"questions":  questions,
		"submission": submissionInfo,
	})
}

// UpdateSubmissionScore handles updating a submission score manually by teacher.
func (h *AssignmentHandler) UpdateSubmissionScore(c *gin.Context) {
	submissionID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以修改分数"})
		return
	}

	var req struct {
		Score int `json:"score"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if req.Score < 0 || req.Score > 100 {
		c.JSON(400, gin.H{"error": "分数必须在0-100之间"})
		return
	}

	err := h.assignSvc.UpdateSubmissionScore(submissionID, req.Score)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "分数更新成功"})
}

// UpdateTeacherFeedback handles updating teacher feedback for a submission.
func (h *AssignmentHandler) UpdateTeacherFeedback(c *gin.Context) {
	submissionID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以添加批注"})
		return
	}

	var req struct {
		Feedback string `json:"feedback"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	err := h.assignSvc.UpdateTeacherFeedback(submissionID, req.Feedback)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "批注更新成功"})
}

// RegradeSubmission handles triggering AI regrading for a submission.
func (h *AssignmentHandler) RegradeSubmission(c *gin.Context) {
	submissionID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以重新批改"})
		return
	}

	err := h.assignSvc.RegradeSubmission(submissionID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "重新批改已触发，请稍后查看结果"})
}

// GetPublishedClasses handles getting the list of classes an assignment is published to.
func (h *AssignmentHandler) GetPublishedClasses(c *gin.Context) {
	assignID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以查看发布信息"})
		return
	}

	// 验证教师是否有权查看此作业的发布信息
	assign, _, err := h.assignSvc.GetAssignmentDetail(assignID)
	if err != nil {
		c.JSON(404, gin.H{"error": "作业不存在"})
		return
	}

	if assign.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权查看其他教师的作业发布信息"})
		return
	}

	publishedClasses, err := h.assignSvc.GetPublishedClasses(assignID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, publishedClasses)
}

// DownloadSubmissionCode handles downloading student's code as a file.
func (h *AssignmentHandler) DownloadSubmissionCode(c *gin.Context) {
	submissionID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以下载代码"})
		return
	}

	code, fileName, err := h.assignSvc.GetSubmissionCodeForDownload(submissionID)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.String(200, code)
}

// DeleteAssignment handles deleting an assignment and all related data.
func (h *AssignmentHandler) DeleteAssignment(c *gin.Context) {
	assignID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以删除作业"})
		return
	}

	// 验证教师是否有权删除此作业
	assign, _, err := h.assignSvc.GetAssignmentDetail(assignID)
	if err != nil {
		c.JSON(404, gin.H{"error": "作业不存在"})
		return
	}

	if assign.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权删除其他教师的作业"})
		return
	}

	err = h.assignSvc.DeleteAssignment(assignID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "作业删除成功"})
}
