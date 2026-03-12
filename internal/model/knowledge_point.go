package model

import "gorm.io/gorm"

// KnowledgePoint represents a single node in the wisdom graph.
type KnowledgePoint struct {
	gorm.Model
	Name        string `gorm:"unique;not null"`
	Description string `gorm:"type:text"`
	Level       int
	CategoryID  uint  `gorm:"index"`
	ParentID    *uint `gorm:"index"` // Pointer to allow for null (root nodes)
}

// KnowledgePointCategory represents a category of knowledge points.
type KnowledgePointCategory struct {
	gorm.Model
	Name string `gorm:"unique;not null"`
}
