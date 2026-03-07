package handler

import (
	"GoCodeMentor/internal/model"
	"net/http"
    "fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


type WisdomGraphHandler struct {
	DB *gorm.DB
}

func NewWisdomGraphHandler(db *gorm.DB) *WisdomGraphHandler {
	return &WisdomGraphHandler{DB: db}
}


func (h *WisdomGraphHandler) GetWisdomGraph(c *gin.Context) {
	var points []model.KnowledgePoint
	var categories []model.KnowledgePointCategory

	if err := h.DB.Find(&points).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch knowledge points"})
		return
	}

	if err := h.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	type Node struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		SymbolSize int    `json:"symbolSize"`
		Category   int    `json:"category"`
	}

	type Link struct {
		Source string `json:"source"`
		Target string `json:"target"`
	}

	type Category struct {
		Name string `json:"name"`
	}

	nodes := make([]Node, 0)
	links := make([]Link, 0)
	chartCategories := make([]Category, 0)

	for _, cat := range categories {
		chartCategories = append(chartCategories, Category{Name: cat.Name})
	}

	for _, p := range points {
		node := Node{
			ID:         fmt.Sprintf("%d", p.ID),
			Name:       p.Name,
			SymbolSize: 50, // default size
			Category:   int(p.CategoryID) - 1,
		}
		if p.ParentID == nil {
			node.SymbolSize = 80 // root node
		}
		nodes = append(nodes, node)

		if p.ParentID != nil {
			links = append(links, Link{
				Source: fmt.Sprintf("%d", *p.ParentID),
				Target: fmt.Sprintf("%d", p.ID),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes":      nodes,
		"links":      links,
		"categories": chartCategories,
	})
}