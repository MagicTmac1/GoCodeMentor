package handler

import (
	"GoCodeMentor/internal/dto"
	"GoCodeMentor/internal/service"

	"github.com/gin-gonic/gin"
)

// ResourceHandler handles resource-related requests.
type ResourceHandler struct {
	resourceSvc service.IResourceService
}

// NewResourceHandler creates a new ResourceHandler.
func NewResourceHandler(resourceSvc service.IResourceService) *ResourceHandler {
	return &ResourceHandler{resourceSvc: resourceSvc}
}

// GetAllResources handles the request to get all resources.
func (h *ResourceHandler) GetAllResources(c *gin.Context) {
	resources, err := h.resourceSvc.GetAllResources()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取资源列表失败: " + err.Error()})
		return
	}
	c.JSON(200, resources)
}

// CreateResource handles the request to create a new resource.
func (h *ResourceHandler) CreateResource(c *gin.Context) {
	var req dto.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	createdResource, err := h.resourceSvc.CreateResource(&req)
	if err != nil {
		c.JSON(500, gin.H{"error": "创建资源失败: " + err.Error()})
		return
	}

	c.JSON(201, createdResource)
}

// DeleteResource handles the request to delete a resource.
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	resourceID := c.Param("resourceId")
	if resourceID == "" {
		c.JSON(400, gin.H{"error": "无效的资源ID"})
		return
	}

	if err := h.resourceSvc.DeleteResource(resourceID); err != nil {
		c.JSON(500, gin.H{"error": "删除资源失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "资源删除成功"})
}

// GetResourceStats handles the request to get all stats for the resource page.
func (h *ResourceHandler) GetResourceStats(c *gin.Context) {
	userID := c.GetString("userID")
	stats, err := h.resourceSvc.GetResourceStats(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取资源数据失败: " + err.Error()})
		return
	}
	c.JSON(200, stats)
}

// ToggleResourceLike handles the request to like/unlike a resource.
func (h *ResourceHandler) ToggleResourceLike(c *gin.Context) {
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
