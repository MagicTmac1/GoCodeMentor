package dto

// ResourceStatsResponse defines the data structure for resource page statistics.
type ResourceStatsResponse struct {
	LikeCounts  map[string]int64  `json:"likeCounts"`
	UserLikes   map[string]bool   `json:"userLikes"`
	Leaderboard []LeaderboardItem `json:"leaderboard"`
}

// LeaderboardItem defines the structure for a single item in the leaderboard.
type LeaderboardItem struct {
	ResourceID string `json:"resourceId"`
	LikeCount  int64  `json:"likeCount"`
}

// CreateResourceRequest defines the request body for creating a new resource.
type CreateResourceRequest struct {
	ResourceID  string `json:"resourceId"`
	Title       string `json:"title" binding:"required"`
	URL         string `json:"url" binding:"required,url"`
	Description string `json:"description"`
	Category    string `json:"category" binding:"required"`
	IconURL     string `json:"iconURL"`
}
