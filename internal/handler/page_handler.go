package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PageHandler handles requests for rendering HTML pages.
type PageHandler struct{}

// NewPageHandler creates a new PageHandler.
func NewPageHandler() *PageHandler {
	return &PageHandler{}
}

// LoginPage renders the login page.
func (h *PageHandler) LoginPage(c *gin.Context) {
	// 登录页也禁用缓存，确保状态是最新的
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.File("web/templates/login.html")
}

// IndexPage renders the index page, redirecting based on role.
func (h *PageHandler) IndexPage(c *gin.Context) {
	// 禁用缓存，防止角色切换时显示旧的 Dashboard
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	userRole := c.GetString("userRole")

	switch userRole {
	case "teacher", "admin":
		c.File("web/templates/teacher_dashboard.html")
	case "student":
		c.File("web/templates/student_dashboard.html")
	default:
		c.Redirect(http.StatusFound, "/login")
	}
}

// AssignmentPage renders the assignment management page.
func (h *PageHandler) AssignmentPage(c *gin.Context) {
	c.HTML(http.StatusOK, "assignment.html", gin.H{
		"Title": "作业管理",
		"Nav":   "assignment",
	})
}

// FeedbackPage renders the feedback page.
func (h *PageHandler) FeedbackPage(c *gin.Context) {
	feedbackID := c.Query("id")
	userRole := c.GetString("userRole")
	userID := c.GetString("userID")
	userName := c.GetString("userName")

	c.HTML(http.StatusOK, "feedback.html", gin.H{
		"Title": "意见反馈",
		"Nav":   "feedback",
		"ID":    feedbackID,
		"User": gin.H{
			"ID":   userID,
			"Role": userRole,
			"Name": userName,
		},
	})
}

// TeacherClassesPage redirects to the teacher dashboard.
func (h *PageHandler) TeacherClassesPage(c *gin.Context) {
	c.Redirect(302, "/")
}

// ClassStudentsPage renders the class students page.
func (h *PageHandler) ClassStudentsPage(c *gin.Context) {
	c.File("web/templates/class_students.html")
}

// AccountManagementPage renders the account management page (Admin only).
func (h *PageHandler) AccountManagementPage(c *gin.Context) {
	userRole := c.GetString("userRole")
	if userRole != "admin" {
		c.String(http.StatusForbidden, "只有管理员可以访问此页面")
		return
	}
	c.File("web/templates/account_management.html")
}

// StudentChatsPage renders the student chats page.
func (h *PageHandler) StudentChatsPage(c *gin.Context) {
	requestedID := c.Param("id")
	userRole := c.GetString("userRole")
	userID := c.GetString("userID")
	studentName := c.Query("name")

	switch userRole {
	case "student":
		// 如果是学生，只能看自己的聊天记录
		if requestedID != userID {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "student_chat.html", gin.H{
			"StudentID":   userID,
			"StudentName": studentName,
			"IsTeacher":   false,
		})
	case "teacher", "admin":
		c.HTML(http.StatusOK, "teacher_chat.html", gin.H{
			"StudentID":   requestedID,
			"StudentName": studentName,
			"IsTeacher":   true,
		})
	default:
		c.Redirect(http.StatusFound, "/login")
	}
}

// StudentAssignmentsPage renders the student assignments page.
func (h *PageHandler) StudentAssignmentsPage(c *gin.Context) {
	c.File("web/templates/student_assignments.html")
}

// DoAssignmentPage renders the page for a student to do an assignment.
func (h *PageHandler) DoAssignmentPage(c *gin.Context) {
	assignID := c.Query("id")
	c.HTML(http.StatusOK, "do_assignment.html", gin.H{
		"AssignmentID": assignID,
	})
}
