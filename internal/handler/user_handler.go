package handler

import (
	"GoCodeMentor/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related requests.
type UserHandler struct {
	userSvc service.IUserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userSvc service.IUserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// Register handles user registration.
func (h *UserHandler) Register(c *gin.Context) {
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

	user, err := h.userSvc.Register(req.Username, req.Password, req.Name, req.Role)
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
}

// Login handles user login.
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	user, err := h.userSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}

	// 设置会话Cookie（不设置MaxAge，即浏览器关闭即失效）
	c.SetCookie("user_id", user.ID, 0, "/", "", false, false)
	c.SetCookie("user_role", user.Role, 0, "/", "", false, false)
	c.SetCookie("user_name", user.Name, 0, "/", "", false, false)

	c.JSON(200, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"name":     user.Name,
		"role":     user.Role,
	})
}

// FindUser handles finding a user by username.
func (h *UserHandler) FindUser(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "用户名不能为空"})
		return
	}

	user, err := h.userSvc.GetByUsername(username)
	if err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(200, user)
}

// GetAllUsers handles getting all users (Admin only).
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	userRole := c.GetString("userRole")
	if userRole != "admin" {
		c.JSON(403, gin.H{"error": "权限不足，只有管理员可以查看所有账号"})
		return
	}

	users, err := h.userSvc.GetAllUsers()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取用户列表失败: " + err.Error()})
		return
	}

	// 过滤敏感信息（如密码）
	type UserInfo struct {
		ID        string `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}

	var userInfos []UserInfo
	for _, u := range users {
		userInfos = append(userInfos, UserInfo{
			ID:        u.ID,
			Username:  u.Username,
			Name:      u.Name,
			Role:      u.Role,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(200, userInfos)
}

// ResetPassword handles password reset (Admin only).
func (h *UserHandler) ResetPassword(c *gin.Context) {
	userRole := c.GetString("userRole")
	if userRole != "admin" {
		c.JSON(403, gin.H{"error": "权限不足，只有管理员可以重置密码"})
		return
	}

	var req struct {
		UserID      string `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if req.UserID == "" || req.NewPassword == "" {
		c.JSON(400, gin.H{"error": "用户ID和新密码不能为空"})
		return
	}

	if err := h.userSvc.ResetPassword(req.UserID, req.NewPassword); err != nil {
		c.JSON(500, gin.H{"error": "重置密码失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "密码重置成功"})
}
