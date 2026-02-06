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

	feedback, err := h.feedbackSvc.Create(req.Type, req.Title, req.Content, req.AnonymousID)
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

// LikeFeedback handles liking a feedback.
func (h *FeedbackHandler) LikeFeedback(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.feedbackSvc.Like(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "点赞成功"})
}
