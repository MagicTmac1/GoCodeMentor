package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"

	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/pkg/excel"
	"GoCodeMentor/internal/pkg/siliconflow"
	"GoCodeMentor/internal/repository"
	"GoCodeMentor/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// AuthMiddleware 登录验证中间件（支持Cookie或Header）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从Cookie获取，其次从Header获取
		userID, _ := c.Cookie("user_id")
		if userID == "" {
			userID = c.GetHeader("X-User-ID")
		}

		if userID == "" {
			// 检查是否是页面请求还是API请求
			path := c.Request.URL.Path
			isAPI := len(path) >= 4 && path[:4] == "/api"
			if c.GetHeader("Accept") == "application/json" || isAPI {
				c.JSON(401, gin.H{"error": "未登录"})
			} else {
				c.Redirect(302, "/login")
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文，方便后续使用
		c.Set("userID", userID)
		c.Next()
	}
}

// TeacherAuthMiddleware 教师权限验证中间件
func TeacherAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从Cookie获取
		userRole, _ := c.Cookie("user_role")
		if userRole == "" {
			userRole = c.GetHeader("X-User-Role")
		}

		if userRole != "teacher" {
			path := c.Request.URL.Path
			isAPI := len(path) >= 4 && path[:4] == "/api"
			if c.GetHeader("Accept") == "application/json" || isAPI {
				c.JSON(403, gin.H{"error": "只有教师可以访问"})
			} else {
				c.String(403, "只有教师可以访问此页面")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	if err := repository.InitDB(); err != nil {
		panic("数据库初始化失败：" + err.Error())
	}

	client := siliconflow.NewClient()
	sessionSvc := service.NewSessionService(client)
	assignSvc := service.NewAssignmentService(client)
	feedbackSvc := service.NewFeedbackService()
	userSvc := service.NewUserService()
	classSvc := service.NewClassService()

	r := gin.Default()

	// 加载模板（解析 templates 文件夹下所有 html）
	r.LoadHTMLGlob("web/templates/*")

	// 静态文件服务（CSS/JS）
	r.Static("/static", "./web/static")

	// ========== 公开路由（无需登录） ==========

	// 登录页面（独立HTML，不继承layout）
	r.GET("/login", func(c *gin.Context) {
		c.File("web/templates/login.html")
	})

	// 用户注册和登录API
	r.POST("/api/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Name     string `json:"name"`
			Role     string `json:"role"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}

		user, err := userSvc.Register(req.Username, req.Password, req.Name, req.Role)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"role":     user.Role,
		})
	})

	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}

		user, err := userSvc.Login(req.Username, req.Password)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			return
		}

		// 设置Cookie（7天有效期）
		c.SetCookie("user_id", user.ID, 7*24*60*60, "/", "", false, false)
		c.SetCookie("user_role", user.Role, 7*24*60*60, "/", "", false, false)
		c.SetCookie("user_name", user.Name, 7*24*60*60, "/", "", false, false)

		c.JSON(200, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
			"role":     user.Role,
		})
	})

	// ========== 需要登录的路由 ==========
	authorized := r.Group("/")
	authorized.Use(AuthMiddleware())
	{
		// 首页 - 根据角色显示不同界面
		authorized.GET("/", func(c *gin.Context) {
			// 从Cookie获取用户角色
			userRole, _ := c.Cookie("user_role")
			if userRole == "" {
				userRole = c.GetHeader("X-User-Role")
			}

			if userRole == "teacher" {
				// 教师跳转到教师工作台
				c.File("web/templates/teacher_dashboard.html")
			} else {
				// 学生显示课堂答疑
				c.HTML(http.StatusOK, "chat.html", gin.H{
					"Title": "课堂答疑",
					"Nav":   "chat",
				})
			}
		})

		// 作业管理页面
		authorized.GET("/assignments", func(c *gin.Context) {
			c.HTML(http.StatusOK, "assignment.html", gin.H{
				"Title": "作业管理",
				"Nav":   "assignment",
			})
		})

		// 反馈论坛页面
		authorized.GET("/feedback", func(c *gin.Context) {
			c.HTML(http.StatusOK, "feedback.html", gin.H{
				"Title": "意见反馈",
				"Nav":   "feedback",
			})
		})
	}

	// ========== 需要教师权限的路由 ==========
	teacherOnly := r.Group("/")
	teacherOnly.Use(AuthMiddleware(), TeacherAuthMiddleware())
	{
		// 教师班级管理页面（已合并到教师工作台）
		teacherOnly.GET("/teacher/classes", func(c *gin.Context) {
			c.Redirect(302, "/")
		})

		// 班级学生列表页面
		teacherOnly.GET("/class/:id/students", func(c *gin.Context) {
			c.File("web/templates/class_students.html")
		})

		// 学生答疑记录页面（教师查看）
		teacherOnly.GET("/student/:id/chats", func(c *gin.Context) {
			c.File("web/templates/student_chats.html")
		})
	}

	// ========== 学生作业详情页面（独立路由，避免路由冲突） ==========
	r.GET("/student_assignments.html", AuthMiddleware(), TeacherAuthMiddleware(), func(c *gin.Context) {
		c.File("web/templates/student_assignments.html")
	})

	// ========== API 接口 ==========

	// 聊天 API（带用户验证）
	r.POST("/api/chat", func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
			Question  string `json:"question"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}

		// 获取用户ID
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = "anonymous" // 未登录用户使用匿名
		}

		ctx := context.Background()
		answer, sessionID, err := sessionSvc.Chat(ctx, req.SessionID, userID, req.Question)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"answer":     answer,
			"session_id": sessionID,
		})
	})

	// 获取历史记录（带权限验证）
	r.GET("/api/history", func(c *gin.Context) {
		sessionID := c.Query("session_id")
		if sessionID == "" {
			c.JSON(200, []interface{}{})
			return
		}

		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = "anonymous"
		}

		messages, err := sessionSvc.GetHistory(sessionID, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, messages)
	})

	// 获取用户的所有会话
	r.GET("/api/sessions", func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(401, gin.H{"error": "未登录"})
			return
		}

		sessions, err := sessionSvc.GetUserSessions(userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, sessions)
	})

	// ========== 班级管理 API ==========

	// 创建班级
	r.POST("/api/classes", func(c *gin.Context) {
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

		class, err := classSvc.CreateClass(req.Name, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, class)
	})

	// 获取教师的班级列表
	r.GET("/api/classes", func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if userID == "" || userRole != "teacher" {
			c.JSON(403, gin.H{"error": "无权访问"})
			return
		}

		classes, err := classSvc.GetClassesByTeacherID(userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, classes)
	})

	// 获取班级信息
	r.GET("/api/classes/:id", func(c *gin.Context) {
		classID := c.Param("id")
		class, err := classSvc.GetClassByID(classID)
		if err != nil {
			c.JSON(404, gin.H{"error": "班级不存在"})
			return
		}
		c.JSON(200, class)
	})

	// 获取班级的学生列表
	r.GET("/api/classes/:id/students", func(c *gin.Context) {
		classID := c.Param("id")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		// 验证班级存在
		class, err := classSvc.GetClassByID(classID)
		if err != nil {
			c.JSON(404, gin.H{"error": "班级不存在"})
			return
		}

		// 验证权限（只有该班教师可以查看）
		if userRole != "teacher" || class.TeacherID != userID {
			c.JSON(403, gin.H{"error": "无权查看"})
			return
		}

		students, err := userSvc.GetStudentsByClassID(classID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, students)
	})

	// 学生加入班级
	r.POST("/api/classes/join", func(c *gin.Context) {
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

		err := classSvc.JoinClass(userID, req.Code)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "加入成功"})
	})

	// 查找用户（通过用户名）
	r.GET("/api/users/find", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(400, gin.H{"error": "用户名不能为空"})
			return
		}

		user, err := userSvc.GetByUsername(username)
		if err != nil {
			c.JSON(404, gin.H{"error": "用户不存在"})
			return
		}

		c.JSON(200, user)
	})

	// 教师添加学生到班级
	r.POST("/api/classes/:id/students", func(c *gin.Context) {
		classID := c.Param("id")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		// 验证权限
		class, err := classSvc.GetClassByID(classID)
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

		// 验证学生存在
		student, err := userSvc.GetByID(req.StudentID)
		if err != nil || student.Role != "student" {
			c.JSON(400, gin.H{"error": "学生不存在"})
			return
		}

		// 更新学生的班级ID
		if err := classSvc.AddStudentToClass(req.StudentID, classID); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "添加成功"})
	})

	// 教师从班级移除学生
	r.DELETE("/api/classes/:id/students/:studentId", func(c *gin.Context) {
		classID := c.Param("id")
		studentID := c.Param("studentId")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		// 验证权限
		class, err := classSvc.GetClassByID(classID)
		if err != nil {
			c.JSON(404, gin.H{"error": "班级不存在"})
			return
		}

		if userRole != "teacher" || class.TeacherID != userID {
			c.JSON(403, gin.H{"error": "无权操作"})
			return
		}

		// 移除学生的班级ID
		if err := classSvc.RemoveStudentFromClass(studentID, classID); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "移除成功"})
	})

	// 教师查看学生的会话列表
	r.GET("/api/students/:id/sessions", func(c *gin.Context) {
		studentID := c.Param("id")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if userID == "" || userRole != "teacher" {
			c.JSON(403, gin.H{"error": "无权查看"})
			return
		}

		sessions, err := sessionSvc.GetStudentSessions(userID, studentID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, sessions)
	})

	// 教师查看学生的作业情况
	r.GET("/api/students/:id/assignments", func(c *gin.Context) {
		studentID := c.Param("id")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if userID == "" || userRole != "teacher" {
			c.JSON(403, gin.H{"error": "无权查看"})
			return
		}

		// 获取学生信息
		student, err := userSvc.GetByID(studentID)
		if err != nil || student.Role != "student" {
			c.JSON(404, gin.H{"error": "学生不存在"})
			return
		}

		// 获取学生所在班级
		if student.ClassID == nil {
			c.JSON(200, gin.H{
				"student":     student,
				"assignments": []interface{}{},
			})
			return
		}

		// 获取班级的已发布作业
		assignments, err := assignSvc.GetAssignmentsByClass(*student.ClassID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 获取学生的提交情况
		var assignmentDetails []gin.H
		for _, assign := range assignments {
			// 获取学生的提交
			var submission model.Submission
			err := repository.DB.Where("assignment_id = ? AND student_id = ?", assign.ID, studentID).First(&submission).Error

			var status string
			var submissionInfo gin.H
			if err != nil {
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
	})

	// 学生查看自己的作业情况
	r.GET("/api/my/assignments", func(c *gin.Context) {
		studentID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if studentID == "" || userRole != "student" {
			c.JSON(403, gin.H{"error": "无权查看"})
			return
		}

		// 获取学生信息
		student, err := userSvc.GetByID(studentID)
		if err != nil {
			c.JSON(404, gin.H{"error": "学生不存在"})
			return
		}

		// 获取学生所在班级
		if student.ClassID == nil {
			c.JSON(200, gin.H{
				"student":     student,
				"assignments": []interface{}{},
			})
			return
		}

		// 获取班级的已发布作业
		assignments, err := assignSvc.GetAssignmentsByClass(*student.ClassID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 获取学生的提交情况
		var assignmentDetails []gin.H
		for _, assign := range assignments {
			// 获取学生的提交
			var submission model.Submission
			err := repository.DB.Where("assignment_id = ? AND student_id = ?", assign.ID, studentID).First(&submission).Error

			var status string
			var submissionInfo gin.H
			if err != nil {
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
	})

	// 获取作业详情（包含学生提交信息）
	r.GET("/api/assignments/:id/student/:studentId", func(c *gin.Context) {
		assignID := c.Param("id")
		studentID := c.Param("studentId")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if userID == "" || userRole != "teacher" {
			c.JSON(403, gin.H{"error": "无权查看"})
			return
		}

		// 获取作业信息
		assign, questions, err := assignSvc.GetAssignmentDetail(assignID)
		if err != nil {
			c.JSON(404, gin.H{"error": "作业不存在"})
			return
		}

		// 获取学生的提交
		var submission model.Submission
		err = repository.DB.Where("assignment_id = ? AND student_id = ?", assignID, studentID).First(&submission).Error

		var submissionInfo gin.H
		if err != nil {
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
	})

	// 获取班级统计信息（学生数量、作业数量、待批改数量）
	r.GET("/api/classes/:id/stats", func(c *gin.Context) {
		classID := c.Param("id")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		// 验证权限
		class, err := classSvc.GetClassByID(classID)
		if err != nil {
			c.JSON(404, gin.H{"error": "班级不存在"})
			return
		}

		if userRole != "teacher" || class.TeacherID != userID {
			c.JSON(403, gin.H{"error": "无权查看"})
			return
		}

		// 获取学生数量
		students, err := userSvc.GetStudentsByClassID(classID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 获取作业数量
		assignments, err := assignSvc.GetAssignmentsByClass(classID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		// 获取待批改数量（所有已发布作业的未批改提交总数）
		var pendingCount int
		for _, assign := range assignments {
			var count int64
			repository.DB.Model(&model.Submission{}).
				Where("assignment_id = ? AND status = ?", assign.ID, "submitted").
				Count(&count)
			pendingCount += int(count)
		}

		c.JSON(200, gin.H{
			"student_count":    len(students),
			"assignment_count": len(assignments),
			"pending_count":    pendingCount,
		})
	})

	// ========== Excel 导入 API ==========

	// 下载学生名单模板
	r.GET("/api/templates/students", func(c *gin.Context) {
		f, err := excel.CreateStudentTemplate()
		if err != nil {
			c.JSON(500, gin.H{"error": "创建模板失败"})
			return
		}

		// 将Excel写入buffer
		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			c.JSON(500, gin.H{"error": "生成文件失败"})
			return
		}

		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=\u5b66\u751f\u540d\u5355\u6a21\u677f.xlsx")
		c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
	})

	// 下载班级名单模板
	r.GET("/api/templates/classes", func(c *gin.Context) {
		f, err := excel.CreateClassTemplate()
		if err != nil {
			c.JSON(500, gin.H{"error": "创建模板失败"})
			return
		}

		// 将Excel写入buffer
		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			c.JSON(500, gin.H{"error": "生成文件失败"})
			return
		}

		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=\u73ed\u7ea7\u540d\u5355\u6a21\u677f.xlsx")
		c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
	})

	// 批量导入学生到班级（自动创建不存在的用户）
	r.POST("/api/classes/:id/students/import", func(c *gin.Context) {
		classID := c.Param("id")
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		// 验证权限
		class, err := classSvc.GetClassByID(classID)
		if err != nil {
			c.JSON(404, gin.H{"error": "班级不存在"})
			return
		}

		if userRole != "teacher" || class.TeacherID != userID {
			c.JSON(403, gin.H{"error": "无权操作"})
			return
		}

		// 获取上传的文件
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "请选择Excel文件"})
			return
		}
		defer file.Close()

		// 解析Excel
		students, err := excel.ParseStudentsExcel(file)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 批量添加学生（自动创建不存在的用户）
		var successCount, failCount, createdCount int
		var failMessages []string

		for _, s := range students {
			var studentID string

			// 查找学生，如果不存在则自动创建
			existingStudent, err := userSvc.GetByUsername(s.Username)
			if err != nil {
				// 用户不存在，自动创建
				name := s.Name
				if name == "" {
					name = s.Username
				}
				password := s.Password
				if password == "" {
					password = "123456"
				}

				newStudent, err := userSvc.Register(s.Username, password, name, "student")
				if err != nil {
					failCount++
					failMessages = append(failMessages, s.Username+": 创建用户失败 - "+err.Error())
					continue
				}
				studentID = newStudent.ID
				createdCount++
			} else {
				studentID = existingStudent.ID
			}

			// 添加到班级
			if err := classSvc.AddStudentToClass(studentID, classID); err != nil {
				failCount++
				failMessages = append(failMessages, s.Username+": 添加到班级失败 - "+err.Error())
				continue
			}

			successCount++
		}

		c.JSON(200, gin.H{
			"success":       successCount,
			"failed":        failCount,
			"created":       createdCount,
			"total":         len(students),
			"fail_messages": failMessages,
		})
	})

	// 批量导入班级和学生（统一模板）
	r.POST("/api/classes/import", func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if userRole != "teacher" {
			c.JSON(403, gin.H{"error": "只有教师可以导入班级"})
			return
		}

		// 获取上传的文件
		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "请选择Excel文件"})
			return
		}
		defer file.Close()

		// 解析Excel（统一模板）
		students, err := excel.ParseStudentsExcel(file)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 先收集所有班级名称并去重
		classMap := make(map[string]string) // className -> classID
		for _, s := range students {
			if s.Class != "" {
				classMap[s.Class] = ""
			}
		}

		// 批量创建班级
		var classSuccessCount, classFailCount int
		for className := range classMap {
			// 检查班级是否已存在
			existingClasses, _ := classSvc.GetClassesByTeacherID(userID)
			exists := false
			for _, ec := range existingClasses {
				if ec.Name == className {
					classMap[className] = ec.ID
					exists = true
					break
				}
			}

			if !exists {
				class, err := classSvc.CreateClass(className, userID)
				if err != nil {
					classFailCount++
					continue
				}
				classMap[className] = class.ID
				classSuccessCount++
			}
		}

		// 批量创建学生并分配到班级
		var studentSuccessCount, studentFailCount, studentCreatedCount int
		var failMessages []string

		for _, s := range students {
			if s.Username == "" {
				continue
			}

			var studentID string

			// 查找或创建学生
			existingStudent, err := userSvc.GetByUsername(s.Username)
			if err != nil {
				// 用户不存在，自动创建
				name := s.Name
				if name == "" {
					name = s.Username
				}
				password := s.Password
				if password == "" {
					password = "123456"
				}

				newStudent, err := userSvc.Register(s.Username, password, name, "student")
				if err != nil {
					studentFailCount++
					failMessages = append(failMessages, s.Username+": 创建用户失败 - "+err.Error())
					continue
				}
				studentID = newStudent.ID
				studentCreatedCount++
			} else {
				studentID = existingStudent.ID
			}

			// 如果有班级信息，添加到班级
			if s.Class != "" {
				classID := classMap[s.Class]
				if classID != "" {
					if err := classSvc.AddStudentToClass(studentID, classID); err != nil {
						studentFailCount++
						failMessages = append(failMessages, s.Username+": 添加到班级失败 - "+err.Error())
						continue
					}
				}
			}

			studentSuccessCount++
		}

		c.JSON(200, gin.H{
			"class_success":   classSuccessCount,
			"class_failed":    classFailCount,
			"student_success": studentSuccessCount,
			"student_failed":  studentFailCount,
			"student_created": studentCreatedCount,
			"total":           len(students),
		})
	})

	// ========== 作业 API ==========
	r.POST("/api/assignments/generate", func(c *gin.Context) {
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
		assign, err := assignSvc.GenerateAssignmentByAI(ctx, req.Topic, req.Difficulty, userID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, assign)
	})

	// 获取作业列表（教师：所有作业，学生：已发布到所在班级的作业）
	r.GET("/api/assignments", func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		userRole := c.GetHeader("X-User-Role")

		if userRole == "teacher" {
			// 教师获取自己的所有作业（包含草稿）
			assignments, err := assignSvc.GetAssignmentList(userID)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, assignments)
		} else {
			// 学生获取已发布到所在班级的作业
			// 先获取学生所在班级
			var user model.User
			if err := repository.DB.First(&user, "id = ?", userID).Error; err != nil {
				c.JSON(200, []interface{}{})
				return
			}
			if user.ClassID == nil {
				c.JSON(200, []interface{}{})
				return
			}
			assignments, err := assignSvc.GetAssignmentsByClass(*user.ClassID)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, assignments)
		}
	})

	// 发布作业到班级
	r.POST("/api/assignments/:id/publish", func(c *gin.Context) {
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

		// 检查班级是否有学生
		students, err := userSvc.GetStudentsByClassID(req.ClassID)
		if err != nil {
			c.JSON(500, gin.H{"error": "检查班级学生失败"})
			return
		}

		if len(students) == 0 {
			c.JSON(400, gin.H{"error": "该班级暂无学生，无法发布作业"})
			return
		}

		err = assignSvc.PublishAssignment(assignID, req.ClassID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "发布成功"})
	})

	r.GET("/api/assignments/:id", func(c *gin.Context) {
		id := c.Param("id")
		assign, questions, err := assignSvc.GetAssignmentDetail(id)
		if err != nil {
			c.JSON(404, gin.H{"error": "作业不存在"})
			return
		}
		c.JSON(200, gin.H{
			"assignment": assign,
			"questions":  questions,
		})
	})

	r.GET("/api/assignments/:id/qrcode", func(c *gin.Context) {
		id := c.Param("id")
		url := fmt.Sprintf("http://localhost:8081/assignments/do?id=%s", id)

		png, err := qrcode.Encode(url, qrcode.Medium, 256)
		if err != nil {
			c.JSON(500, gin.H{"error": "生成二维码失败"})
			return
		}

		c.Data(http.StatusOK, "image/png", png)
	})

	// 学生答题页面
	r.GET("/assignments/do", func(c *gin.Context) {
		assignID := c.Query("id")
		c.HTML(http.StatusOK, "do_assignment.html", gin.H{
			"AssignmentID": assignID,
		})
	})

	// 提交作业
	r.POST("/api/assignments/:id/submit", func(c *gin.Context) {
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

		// 转换 answers map 为 map[string]string
		answers := make(map[string]string)
		for k, v := range req.Answers {
			if str, ok := v.(string); ok {
				answers[k] = str
			} else {
				answers[k] = fmt.Sprintf("%v", v)
			}
		}

		submissionID, err := assignSvc.SubmitAssignment(assignID, req.StudentID, req.StudentName, answers, req.Code)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"id": submissionID, "message": "提交成功"})
	})

	// ========== 反馈 API ==========
	r.POST("/api/feedback", func(c *gin.Context) {
		var req struct {
			Type        string `json:"type"`
			Title       string `json:"title"`
			Content     string `json:"content"`
			AnonymousID string `json:"anonymous_id"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}

		feedback, err := feedbackSvc.Create(req.Type, req.Title, req.Content, req.AnonymousID)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, feedback)
	})

	r.GET("/api/feedback", func(c *gin.Context) {
		feedbacks, err := feedbackSvc.GetAll()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, feedbacks)
	})

	r.POST("/api/feedback/:id/like", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := feedbackSvc.Like(uint(id)); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "点赞成功"})
	})

	// 启动服务
	r.Run(":8081")
}
