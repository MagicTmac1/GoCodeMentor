package repository

import (
	"GoCodeMentor/internal/model"

	"gorm.io/gorm"
)

type resourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) IResourceRepository {
	return &resourceRepository{db: db}
}

func (r *resourceRepository) CreateResource(resource *model.Resource) error {
	return r.db.Create(resource).Error
}

func (r *resourceRepository) GetAllResources() ([]model.Resource, error) {
	var resources []model.Resource
	// Ordered by creation time DESC
	if err := r.db.Order("created_at desc").Find(&resources).Error; err != nil {
		return nil, err
	}
	return resources, nil
}

// ToggleLike handles liking/unliking a resource for a user.
func (r *resourceRepository) ToggleLike(userID, resourceID string) (bool, error) {
	like := model.ResourceLike{
		UserID:     userID,
		ResourceID: resourceID,
	}

	var count int64
	r.db.Model(&model.ResourceLike{}).Where("user_id = ? AND resource_id = ?", userID, resourceID).Count(&count)

	if count > 0 {
		// User has already liked it, so unlike it
		if err := r.db.Where("user_id = ? AND resource_id = ?", userID, resourceID).Delete(&model.ResourceLike{}).Error; err != nil {
			return false, err
		}
		return false, nil // unliked
	} else {
		// User has not liked it, so add a like
		if err := r.db.Create(&like).Error; err != nil {
			return false, err
		}
		return true, nil // liked
	}
}

// GetLikesByResourceIDs returns a map of resourceID to its like count.
func (r *resourceRepository) GetLikesByResourceIDs(resourceIDs []string) (map[string]int64, error) {
	counts := make(map[string]int64)

	// Initialize all requested resource IDs with 0 likes
	for _, id := range resourceIDs {
		counts[id] = 0
	}

	var results []struct {
		ResourceID string
		TotalLikes int64
	}

	err := r.db.Model(&model.ResourceLike{}).
		Select("resource_id, count(*) as total_likes").
		Where("resource_id IN ?", resourceIDs).
		Group("resource_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, result := range results {
		counts[result.ResourceID] = result.TotalLikes
	}

	return counts, nil
}

// GetUserLikes returns a map indicating which resources a user has liked.
func (r *resourceRepository) GetUserLikes(userID string, resourceIDs []string) (map[string]bool, error) {
	userLikes := make(map[string]bool)

	// Initialize all to false
	for _, id := range resourceIDs {
		userLikes[id] = false
	}

	var likedResources []string
	err := r.db.Model(&model.ResourceLike{}).
		Where("user_id = ? AND resource_id IN ?", userID, resourceIDs).
		Pluck("resource_id", &likedResources).Error

	if err != nil {
		return nil, err
	}

	for _, resourceID := range likedResources {
		userLikes[resourceID] = true
	}

	return userLikes, nil
}

// GetLeaderboard returns the top N most liked resources.
func (r *resourceRepository) GetLeaderboard() ([]model.LeaderboardItem, error) {
	var leaderboard []model.LeaderboardItem
	err := r.db.Model(&model.ResourceLike{}).
		Select("resource_id, count(*) as total_likes").
		Group("resource_id").
		Order("total_likes desc").
		Limit(5). // Top 5
		Scan(&leaderboard).Error

	return leaderboard, err
}

// DeleteResource deletes a resource and its associated likes within a transaction.
func (r *resourceRepository) DeleteResource(resourceID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First, delete all likes associated with the resource
		if err := tx.Where("resource_id = ?", resourceID).Delete(&model.ResourceLike{}).Error; err != nil {
			return err // Rollback
		}

		// Then, delete the resource itself
		if err := tx.Where("resource_id = ?", resourceID).Delete(&model.Resource{}).Error; err != nil {
			return err // Rollback
		}

		// Commit
		return nil
	})
}
