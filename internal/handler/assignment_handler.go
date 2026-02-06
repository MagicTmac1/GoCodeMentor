package handler

import (
	"GoCodeMentor/internal/service"
	"context"
	"fmt"

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
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

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
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

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
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以发布作业"})
		return
	}

	var req struct {
		ClassID string `json:"class_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
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

	err = h.assignSvc.PublishAssignment(assignID, req.ClassID)
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
	var req struct {
		StudentID   string                 `json:"student_id"`
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

	submissionID, err := h.assignSvc.SubmitAssignment(assignID, req.StudentID, req.StudentName, answers, req.Code)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"id": submissionID, "message": "提交成功"})
}

// GetStudentAssignments handles getting all assignments for a specific student (teacher view).
func (h *AssignmentHandler) GetStudentAssignments(c *gin.Context) {
	studentID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "无权查看"})
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
			submissionInfo = gin.H{
				"id":             submission.ID,
				"student_name":   submission.StudentName,
				"answers":        submission.Answers,
				"code_content":   submission.CodeContent,
				"total_score":    submission.TotalScore,
				"ai_feedback":    submission.AIFeedback,
				"detailed_score": submission.DetailedScore,
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
	studentID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

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
			submissionInfo = gin.H{
				"id":             submission.ID,
				"student_name":   submission.StudentName,
				"answers":        submission.Answers,
				"code_content":   submission.CodeContent,
				"total_score":    submission.TotalScore,
				"ai_feedback":    submission.AIFeedback,
				"detailed_score": submission.DetailedScore,
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
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	if userID == "" || userRole != "teacher" {
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
		submissionInfo = gin.H{
			"submitted":      true,
			"student_name":   submission.StudentName,
			"answers":        submission.Answers,
			"code":           submission.CodeContent,
			"total_score":    submission.TotalScore,
			"ai_feedback":    submission.AIFeedback,
			"detailed_score": submission.DetailedScore,
			"status":         submission.Status,
			"created_at":     submission.CreatedAt,
			"updated_at":     submission.UpdatedAt,
		}
	}

	c.JSON(200, gin.H{
		"assignment": assign,
		"questions":  questions,
		"submission": submissionInfo,
	})
}
