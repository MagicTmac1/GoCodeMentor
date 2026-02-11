package handler

import (
	"GoCodeMentor/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FeedbackHandler handles feedback-related requests.
type FeedbackHandler struct {
	feedbackSvc service.IFeedbackService
}

// NewFeedbackHandler creates a new FeedbackHandler.
func NewFeedbackHandler(feedbackSvc service.IFeedbackService) *FeedbackHandler {
	return &FeedbackHandler{feedbackSvc: feedbackSvc}
}

// CreateFeedback handles the creation of a new feedback.
func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
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

	// 从请求头获取用户信息
	userID := c.GetHeader("X-User-ID")
	
	feedback, err := h.feedbackSvc.Create(req.Type, req.Title, req.Content, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, feedback)
}

// GetAllFeedback handles getting all feedback.
func (h *FeedbackHandler) GetAllFeedback(c *gin.Context) {
	feedbacks, err := h.feedbackSvc.GetAll()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, feedbacks)
}

// GetFeedbackByID handles getting a single feedback by ID.
func (h *FeedbackHandler) GetFeedbackByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的反馈ID"})
		return
	}

	feedback, err := h.feedbackSvc.GetByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "反馈不存在"})
		return
	}
	c.JSON(200, feedback)
}

// LikeFeedback handles liking a feedback.
func (h *FeedbackHandler) LikeFeedback(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的反馈ID"})
		return
	}
	
	if err := h.feedbackSvc.Like(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "点赞成功"})
}

// UpdateFeedbackStatus handles updating feedback status.
func (h *FeedbackHandler) UpdateFeedbackStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的反馈ID"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 从请求头获取教师ID (UUID字符串)
	teacherID := c.GetHeader("X-User-ID")
	
	if err := h.feedbackSvc.UpdateStatus(uint(id), req.Status, teacherID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "状态更新成功"})
}

// RespondFeedback handles teacher response to feedback.
func (h *FeedbackHandler) RespondFeedback(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的反馈ID"})
		return
	}

	var req struct {
		Response string `json:"response"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 从请求头获取教师ID (UUID字符串)
	teacherID := c.GetHeader("X-User-ID")
	
	if err := h.feedbackSvc.Respond(uint(id), req.Response, teacherID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "回复成功"})
}

// GetFeedbackStats handles getting feedback statistics.
func (h *FeedbackHandler) GetFeedbackStats(c *gin.Context) {
	stats, err := h.feedbackSvc.GetStats()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, stats)
}

// GetFilteredFeedback handles getting filtered feedback list.
func (h *FeedbackHandler) GetFilteredFeedback(c *gin.Context) {
	feedbackType := c.Query("type")
	status := c.Query("status")
	search := c.Query("search")

	feedbacks, err := h.feedbackSvc.GetFiltered(feedbackType, status, search)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, feedbacks)
}
