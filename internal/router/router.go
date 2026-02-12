package router

import (
	"GoCodeMentor/internal/handler"

	"github.com/gin-gonic/gin"
)

func Setup(
	r *gin.Engine,
	userHandler *handler.UserHandler,
	classHandler *handler.ClassHandler,
	assignmentHandler *handler.AssignmentHandler,
	feedbackHandler *handler.FeedbackHandler,
	sessionHandler *handler.SessionHandler,
	pageHandler *handler.PageHandler,
	excelHandler *handler.ExcelHandler,
	authMiddleware gin.HandlerFunc,
	teacherAuthMiddleware gin.HandlerFunc,
) {
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	// Public routes
	r.GET("/login", pageHandler.LoginPage)
	r.POST("/api/register", userHandler.Register)
	r.POST("api/login", userHandler.Login)

	// Handle common browser/tool ghost requests to keep logs clean
	r.GET("/.well-known/appspecific/com.chrome.devtools.json", func(c *gin.Context) { c.Status(204) })
	r.GET("/favicon.ico", func(c *gin.Context) { c.Status(204) })

	// Authenticated routes
	authorized := r.Group("/")
	authorized.Use(authMiddleware)
	{
		authorized.GET("/", pageHandler.IndexPage)
		authorized.GET("/assignments", pageHandler.AssignmentPage)
		authorized.GET("/feedback", pageHandler.FeedbackPage)
	}

	// Teacher-only routes
	teacherOnly := r.Group("/")
	teacherOnly.Use(authMiddleware, teacherAuthMiddleware)
	{
		teacherOnly.GET("/teacher/classes", pageHandler.TeacherClassesPage)
		teacherOnly.GET("/class/:id/students", pageHandler.ClassStudentsPage)
		teacherOnly.GET("/student/:id/chats", pageHandler.StudentChatsPage)
	}

	r.GET("/student_assignments.html", authMiddleware, teacherAuthMiddleware, pageHandler.StudentAssignmentsPage)

	// API routes
	api := r.Group("/api")
	{
		api.POST("/chat", sessionHandler.Chat)
		api.GET("/history", sessionHandler.GetHistory)
		api.GET("/sessions", sessionHandler.GetUserSessions)

		// Class management
		api.POST("/classes", classHandler.CreateClass)
		api.GET("/classes", classHandler.GetClassesByTeacherID)
		api.GET("/classes/:id", classHandler.GetClassByID)
		api.GET("/classes/:id/students", classHandler.GetStudentsByClassID)
		api.POST("/classes/join", classHandler.JoinClass)
		api.POST("/classes/:id/students", classHandler.AddStudentToClass)
		api.DELETE("/classes/:id/students/:studentId", classHandler.RemoveStudentFromClass)
		api.DELETE("/classes/:id", classHandler.DeleteClass)
		api.GET("/classes/:id/stats", classHandler.GetClassStats)

		// User management
		api.GET("/users/find", userHandler.FindUser)

		// Student data for teachers
		api.GET("/students/:id/sessions", sessionHandler.GetStudentSessions)
		api.GET("/students/:id/assignments", assignmentHandler.GetStudentAssignments)

		// Student's own data
		api.GET("/my/assignments", assignmentHandler.GetMyAssignments)

		// Assignment management
		api.POST("/assignments/generate", assignmentHandler.GenerateAssignmentByAI)
		api.GET("/assignments", assignmentHandler.GetAssignments)
		api.POST("/assignments/:id/publish", assignmentHandler.PublishAssignment)
		api.GET("/assignments/:id", assignmentHandler.GetAssignmentDetail)
		api.GET("/assignments/:id/qrcode", assignmentHandler.GetAssignmentQRCode)
		api.POST("/assignments/:id/submit", assignmentHandler.SubmitAssignment)
		api.GET("/assignments/:id/student/:studentId", assignmentHandler.GetAssignmentSubmissionForStudent)
		api.GET("/assignments/:id/published", assignmentHandler.GetPublishedClasses)
		api.DELETE("/assignments/:id", assignmentHandler.DeleteAssignment)

		// Teacher submission management
		api.PUT("/submissions/:id/score", assignmentHandler.UpdateSubmissionScore)
		api.PUT("/submissions/:id/feedback", assignmentHandler.UpdateTeacherFeedback)
		api.POST("/submissions/:id/regrade", assignmentHandler.RegradeSubmission)
		api.GET("/submissions/:id/download", assignmentHandler.DownloadSubmissionCode)

		// Feedback
		api.POST("/feedback", feedbackHandler.CreateFeedback)
		api.GET("/feedback", feedbackHandler.GetFilteredFeedback)
		api.GET("/feedback/:id", feedbackHandler.GetFeedbackByID)
		api.POST("/feedback/:id/like", feedbackHandler.LikeFeedback)
		api.PUT("/feedback/:id/status", feedbackHandler.UpdateFeedbackStatus)
		api.POST("/feedback/:id/respond", feedbackHandler.RespondFeedback)
		api.GET("/feedback/stats", feedbackHandler.GetFeedbackStats)
		api.GET("/feedback/filter", feedbackHandler.GetFilteredFeedback)

		// Excel templates and imports
		api.GET("/templates/students", excelHandler.DownloadStudentTemplate)
		api.GET("/templates/classes", excelHandler.DownloadClassTemplate)
		api.POST("/classes/:id/students/import", excelHandler.ImportStudentsToClass)
		api.POST("/classes/import", excelHandler.ImportClassesAndStudents)
	}

	// Standalone page routes
	r.GET("/assignments/do", pageHandler.DoAssignmentPage)
}
