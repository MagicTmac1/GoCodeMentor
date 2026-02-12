package handler

import (
	"GoCodeMentor/internal/service"
	"context"

	"github.com/gin-gonic/gin"
)

// SessionHandler handles session-related requests.
type SessionHandler struct {
	sessionSvc service.ISessionService
}

// NewSessionHandler creates a new SessionHandler.
func NewSessionHandler(sessionSvc service.ISessionService) *SessionHandler {
	return &SessionHandler{sessionSvc: sessionSvc}
}

// Chat handles the chat endpoint.
func (h *SessionHandler) Chat(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id"`
		Question  string `json:"question"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		userID = "anonymous" // 未登录用户使用匿名
	}

	ctx := context.Background()
	answer, sessionID, err := h.sessionSvc.Chat(ctx, req.SessionID, userID, req.Question)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"answer":     answer,
		"session_id": sessionID,
	})
}

// GetHistory handles getting chat history.
func (h *SessionHandler) GetHistory(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(200, []interface{}{})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		userID = "anonymous"
	}

	messages, err := h.sessionSvc.GetHistory(sessionID, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, messages)
}

// GetUserSessions handles getting all sessions for a user.
func (h *SessionHandler) GetUserSessions(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(401, gin.H{"error": "未登录"})
		return
	}

	sessions, err := h.sessionSvc.GetUserSessions(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, sessions)
}

// GetStudentSessions handles a teacher getting a student's sessions.
func (h *SessionHandler) GetStudentSessions(c *gin.Context) {
	studentID := c.Param("id")
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	if userID == "" || userRole != "teacher" {
		c.JSON(403, gin.H{"error": "无权查看"})
		return
	}

	sessions, err := h.sessionSvc.GetStudentSessions(userID, studentID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, sessions)
}
