package main

import (
	"GoCodeMentor/internal/handler"
	"GoCodeMentor/internal/pkg/siliconflow"
	"GoCodeMentor/internal/repository"
	"GoCodeMentor/internal/router"
	"GoCodeMentor/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 登录验证中间件（支持Cookie或Header）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 针对所有鉴权路由禁用缓存，防止角色切换或登出后的状态残留
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

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
				// 重定向到登录页时也确保不缓存
				c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
				c.Redirect(302, "/login")
			}
			c.Abort()
			return
		}

		// 获取角色信息
		userRole, _ := c.Cookie("user_role")
		if userRole == "" {
			userRole = c.GetHeader("X-User-Role")
		}

		// 将用户信息存入上下文，方便后续使用
		c.Set("userID", userID)
		c.Set("userRole", userRole)
		c.Next()
	}
}

// TeacherAuthMiddleware 教师权限验证中间件
func TeacherAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先尝试从上下文获取（如果 AuthMiddleware 已经运行）
		userRole := c.GetString("userRole")
		if userRole == "" {
			// 否则从Cookie/Header获取
			userRole, _ = c.Cookie("user_role")
			if userRole == "" {
				userRole = c.GetHeader("X-User-Role")
			}
		}

		if userRole != "teacher" && userRole != "admin" {
			path := c.Request.URL.Path
			isAPI := len(path) >= 4 && path[:4] == "/api"
			if c.GetHeader("Accept") == "application/json" || isAPI {
				c.JSON(403, gin.H{"error": "只有教师或管理员可以访问"})
			} else {
				c.String(403, "只有教师或管理员可以访问此页面")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminAuthMiddleware 管理员权限验证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("userRole")
		if userRole == "" {
			userRole, _ = c.Cookie("user_role")
		}

		if userRole != "admin" {
			path := c.Request.URL.Path
			isAPI := len(path) >= 4 && path[:4] == "/api"
			if c.GetHeader("Accept") == "application/json" || isAPI {
				c.JSON(403, gin.H{"error": "只有管理员可以访问"})
			} else {
				c.String(403, "只有管理员可以访问此页面")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	// 1. 初始化数据库
	db, err := repository.InitDB()
	if err != nil {
		panic("数据库初始化失败：" + err.Error())
	}

	// 2. 初始化 Repositories
	repos := repository.NewRepositories(db)

	// 3. 初始化 Services
	client := siliconflow.NewClient()
	userSvc := service.NewUserService(repos.UserRepo)
	classSvc := service.NewClassService(repos.ClassRepo, repos.UserRepo)
	assignSvc := service.NewAssignmentService(repos.AssignmentRepo, repos.AssignmentClassRepo, repos.QuestionRepo, repos.SubmissionRepo, repos.UserRepo, repos.ClassRepo, client)
	feedbackSvc := service.NewFeedbackService(repos.FeedbackRepo)
	sessionSvc := service.NewSessionService(client, repos.SessionRepo, repos.MessageRepo, repos.UserRepo, repos.ClassRepo)

	// 4. 初始化 Handlers
	userHandler := handler.NewUserHandler(userSvc)
	classHandler := handler.NewClassHandler(classSvc, userSvc, assignSvc)
	assignmentHandler := handler.NewAssignmentHandler(assignSvc, userSvc)
	feedbackHandler := handler.NewFeedbackHandler(feedbackSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	pageHandler := handler.NewPageHandler()
	excelHandler := handler.NewExcelHandler(classSvc, userSvc)

	// 5. 初始化 Gin 引擎并设置路由
	r := gin.Default()
	router.Setup(
		r,
		userHandler,
		classHandler,
		assignmentHandler,
		feedbackHandler,
		sessionHandler,
		pageHandler,
		excelHandler,
		AuthMiddleware(),
		TeacherAuthMiddleware(),
		AdminAuthMiddleware(),
	)

	// 6. 启动服务
	r.Run(":8081")
}
