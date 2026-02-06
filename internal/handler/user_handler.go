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
