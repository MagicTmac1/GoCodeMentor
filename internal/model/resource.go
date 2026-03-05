package model

import "gorm.io/gorm"

// Resource represents a learning resource.
type Resource struct {
	gorm.Model
	ResourceID  string `gorm:"uniqueIndex;not null" json:"resourceId"`
	Title       string `gorm:"not null" json:"title"`
	URL         string `gorm:"not null" json:"url"`
	Description string `json:"description"`
	Category    string `gorm:"not null;index" json:"category"`
	IconURL     string `json:"iconURL"`
}

// ResourceLike tracks user likes for resources.
type ResourceLike struct {
	gorm.Model
	UserID     string `gorm:"index:idx_user_resource,unique"`
	ResourceID string `gorm:"index:idx_user_resource,unique"`
}

// LeaderboardItem represents an item in the resource leaderboard.
type LeaderboardItem struct {
	ResourceID string `json:"resourceId"`
	TotalLikes int64  `json:"totalLikes"`
}
