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
	c.File("web/templates/login.html")
}

// IndexPage renders the index page, redirecting based on role.
func (h *PageHandler) IndexPage(c *gin.Context) {
	userRole, _ := c.Cookie("user_role")
	if userRole == "" {
		userRole = c.GetHeader("X-User-Role")
	}

	if userRole == "teacher" {
		c.File("web/templates/teacher_dashboard.html")
	} else {
		c.HTML(http.StatusOK, "chat.html", gin.H{
			"Title": "课堂答疑",
			"Nav":   "chat",
		})
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
	c.HTML(http.StatusOK, "feedback.html", gin.H{
		"Title": "意见反馈",
		"Nav":   "feedback",
		"ID":    feedbackID,
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

// StudentChatsPage renders the student chats page.
func (h *PageHandler) StudentChatsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "student_chats.html", gin.H{
		"Title": "学生答疑记录",
		"Nav":   "",
	})
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
