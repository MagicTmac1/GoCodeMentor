package handler

import (
	"GoCodeMentor/internal/dto"
	"GoCodeMentor/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// --- Resource Routes ---

// GetAllResources handles the request to get all resources.
func (h *FeedbackHandler) GetAllResources(c *gin.Context) {
	resources, err := h.resourceSvc.GetAllResources()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取资源列表失败: " + err.Error()})
		return
	}
	c.JSON(200, resources)
}

// CreateResource handles the request to create a new resource.
func (h *FeedbackHandler) CreateResource(c *gin.Context) {
	var req dto.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// Here you could add more validation or generate a resource ID if needed
	// For example, using a slug from the title or a UUID.
	// For now, we assume the frontend provides a unique ResourceID.

	createdResource, err := h.resourceSvc.CreateResource(&req)
	if err != nil {
		c.JSON(500, gin.H{"error": "创建资源失败: " + err.Error()})
		return
	}

	c.JSON(201, createdResource)
}

// DeleteResource handles the request to delete a resource.
func (h *FeedbackHandler) DeleteResource(c *gin.Context) {
	resourceID := c.Param("resourceId")
	if resourceID == "" {
		c.JSON(400, gin.H{"error": "无效的资源ID"})
		return
	}

	// In a real-world scenario, you'd check for user permissions here.
	// For now, we allow any authenticated user to delete.

	if err := h.resourceSvc.DeleteResource(resourceID); err != nil {
		c.JSON(500, gin.H{"error": "删除资源失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "资源删除成功"})
}

// FeedbackHandler handles feedback-related requests.
type FeedbackHandler struct {
	feedbackSvc service.IFeedbackService
	resourceSvc service.IResourceService
}

// NewFeedbackHandler creates a new FeedbackHandler.
func NewFeedbackHandler(feedbackSvc service.IFeedbackService, resourceSvc service.IResourceService) *FeedbackHandler {
	return &FeedbackHandler{feedbackSvc: feedbackSvc, resourceSvc: resourceSvc}
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

	// 从上下文获取用户信息
	userID := c.GetString("userID")

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

	// 从上下文获取教师ID (UUID字符串)
	teacherID := c.GetString("userID")

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

	// 从上下文获取教师ID (UUID字符串)
	teacherID := c.GetString("userID")

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

// DeleteFeedback handles deleting a feedback.
func (h *FeedbackHandler) DeleteFeedback(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的反馈ID"})
		return
	}

	// 从上下文获取当前用户信息
	userID := c.GetString("userID")
	userRole := c.GetString("userRole")

	// 获取反馈详情以检查所有权
	feedback, err := h.feedbackSvc.GetByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "反馈不存在"})
		return
	}

	// 权限检查：只有教师、管理员或反馈发布者可以删除
	isOwner := feedback.AnonymousID == userID
	isStaff := userRole == "teacher" || userRole == "admin"

	if !isOwner && !isStaff {
		c.JSON(403, gin.H{"error": "无权删除此反馈"})
		return
	}

	if err := h.feedbackSvc.Delete(uint(id)); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "反馈删除成功"})
}

// --- Resource Routes ---

// GetResourceStats handles the request to get all stats for the resource page.
func (h *FeedbackHandler) GetResourceStats(c *gin.Context) {
	userID := c.GetString("userID")
	stats, err := h.resourceSvc.GetResourceStats(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取资源数据失败: " + err.Error()})
		return
	}
	c.JSON(200, stats)
}

// ToggleResourceLike handles the request to like/unlike a resource.
func (h *FeedbackHandler) ToggleResourceLike(c *gin.Context) {
	userID := c.GetString("userID")
	resourceID := c.Param("resourceId")

	if resourceID == "" {
		c.JSON(400, gin.H{"error": "无效的资源ID"})
		return
	}

	liked, newCount, err := h.resourceSvc.ToggleLike(userID, resourceID)
	if err != nil {
		c.JSON(500, gin.H{"error": "操作失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"resourceId": resourceID,
		"liked":      liked,
		"newCount":   newCount,
	})
}
