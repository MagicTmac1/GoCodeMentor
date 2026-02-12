package handler

import (
	"GoCodeMentor/internal/service"

	"github.com/gin-gonic/gin"
)

// ClassHandler handles class-related requests.
type ClassHandler struct {
	classSvc  service.IClassService
	userSvc   service.IUserService
	assignSvc service.IAssignmentService
}

// NewClassHandler creates a new ClassHandler.
func NewClassHandler(classSvc service.IClassService, userSvc service.IUserService, assignSvc service.IAssignmentService) *ClassHandler {
	return &ClassHandler{classSvc: classSvc, userSvc: userSvc, assignSvc: assignSvc}
}

// CreateClass handles the creation of a new class.
func (h *ClassHandler) CreateClass(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "只有教师可以创建班级"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	class, err := h.classSvc.CreateClass(req.Name, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, class)
}

// GetClassesByTeacherID handles getting all classes for a teacher.
func (h *ClassHandler) GetClassesByTeacherID(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "无权访问"})
		return
	}

	classes, err := h.classSvc.GetClassesByTeacherID(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, classes)
}

// GetClassByID handles getting a class by its ID.
func (h *ClassHandler) GetClassByID(c *gin.Context) {
	classID := c.Param("id")
	class, err := h.classSvc.GetClassByID(classID)
	if err != nil {
		c.JSON(404, gin.H{"error": "班级不存在"})
		return
	}
	c.JSON(200, class)
}

// GetStudentsByClassID handles getting all students in a class.
func (h *ClassHandler) GetStudentsByClassID(c *gin.Context) {
	classID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	class, err := h.classSvc.GetClassByID(classID)
	if err != nil {
		c.JSON(404, gin.H{"error": "班级不存在"})
		return
	}

	if userRole != "teacher" || class.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权查看"})
		return
	}

	students, err := h.userSvc.GetStudentsByClassID(classID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, students)
}

// JoinClass handles a student joining a class.
func (h *ClassHandler) JoinClass(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	if userID == "" || userRole != "student" {
		c.JSON(403, gin.H{"error": "只有学生可以加入班级"})
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	err := h.classSvc.JoinClass(userID, req.Code)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "加入成功"})
}

// AddStudentToClass handles a teacher adding a student to a class.
func (h *ClassHandler) AddStudentToClass(c *gin.Context) {
	classID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	class, err := h.classSvc.GetClassByID(classID)
	if err != nil {
		c.JSON(404, gin.H{"error": "班级不存在"})
		return
	}

	if userRole != "teacher" || class.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权操作"})
		return
	}

	var req struct {
		StudentID string `json:"student_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	student, err := h.userSvc.GetByID(req.StudentID)
	if err != nil || student.Role != "student" {
		c.JSON(400, gin.H{"error": "学生不存在"})
		return
	}

	if err := h.classSvc.AddStudentToClass(req.StudentID, classID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "添加成功"})
}

// RemoveStudentFromClass handles a teacher removing a student from a class.
func (h *ClassHandler) RemoveStudentFromClass(c *gin.Context) {
	classID := c.Param("id")
	studentID := c.Param("studentId")
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	class, err := h.classSvc.GetClassByID(classID)
	if err != nil {
		c.JSON(404, gin.H{"error": "班级不存在"})
		return
	}

	if userRole != "teacher" || class.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权操作"})
		return
	}

	if err := h.classSvc.RemoveStudentFromClass(studentID, classID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "移除成功"})
}

// DeleteClass handles deleting a class.
func (h *ClassHandler) DeleteClass(c *gin.Context) {
	classID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	class, err := h.classSvc.GetClassByID(classID)
	if err != nil {
		c.JSON(404, gin.H{"error": "班级不存在"})
		return
	}

	if userRole != "teacher" || class.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权操作"})
		return
	}

	if err := h.classSvc.DeleteClass(classID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "班级删除成功"})
}

// GetClassStats handles getting statistics for a class.
func (h *ClassHandler) GetClassStats(c *gin.Context) {
	classID := c.Param("id")
	userID := c.GetHeader("X-User-ID")
	userRole := c.GetHeader("X-User-Role")

	class, err := h.classSvc.GetClassByID(classID)
	if err != nil {
		c.JSON(404, gin.H{"error": "班级不存在"})
		return
	}

	if userRole != "teacher" || class.TeacherID != userID {
		c.JSON(403, gin.H{"error": "无权查看"})
		return
	}

	students, err := h.userSvc.GetStudentsByClassID(classID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	assignments, err := h.assignSvc.GetAssignmentsByClass(classID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	type AssignmentStat struct {
		ID               string   `json:"id"`
		Title            string   `json:"title"`
		SubmittedCount   int      `json:"submitted_count"`
		UnsubmittedCount int      `json:"unsubmitted_count"`
		UnsubmittedUsers []string `json:"unsubmitted_users"`
	}

	var assignmentStats []AssignmentStat
	classUnsubmittedSet := make(map[string]bool)

	for _, assign := range assignments {
		stat := AssignmentStat{
			ID:    assign.ID,
			Title: assign.Title,
		}

		for _, student := range students {
			submission, err := h.assignSvc.GetSubmissionByAssignmentAndStudent(assign.ID, student.ID)
			if err != nil || submission == nil {
				stat.UnsubmittedCount++
				stat.UnsubmittedUsers = append(stat.UnsubmittedUsers, student.Name)
				classUnsubmittedSet[student.ID] = true
			} else {
				stat.SubmittedCount++
			}
		}
		assignmentStats = append(assignmentStats, stat)
	}

	c.JSON(200, gin.H{
		"student_count":     len(students),
		"assignment_count":  len(assignments),
		"unsubmitted_count": len(classUnsubmittedSet),
		"assignment_stats":  assignmentStats,
	})
}
