package service

import (
	"GoCodeMentor/internal/dto"
	"GoCodeMentor/internal/model"
	"GoCodeMentor/internal/repository"
	"errors"
	"fmt"
)

type resourceService struct {
	resourceRepo repository.IResourceRepository
}

// NewResourceService creates a new IResourceService.
func NewResourceService(repo repository.IResourceRepository) IResourceService {
	return &resourceService{resourceRepo: repo}
}

func (s *resourceService) GetAllResources() ([]model.Resource, error) {
	return s.resourceRepo.GetAllResources()
}

func (s *resourceService) CreateResource(req *dto.CreateResourceRequest) (*model.Resource, error) {
	// Simple validation for now
	if req.Title == "" || req.URL == "" || req.Category == "" {
		return nil, errors.New("title, url, and category are required")
	}

	newResource := &model.Resource{
		ResourceID:  req.ResourceID,
		Title:       req.Title,
		URL:         req.URL,
		Description: req.Description,
		Category:    req.Category,
		IconURL:     req.IconURL,
	}

	err := s.resourceRepo.CreateResource(newResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return newResource, nil
}

// ToggleLike handles the business logic for liking/unliking a resource.
func (s *resourceService) ToggleLike(userID, resourceID string) (bool, int64, error) {
	// First, verify the resource exists to prevent liking non-existent items.
	// Note: For performance, you might cache resource IDs. For now, a direct check is fine.
	allResources, err := s.GetAllResources()
	if err != nil {
		return false, 0, fmt.Errorf("could not verify resource existence: %w", err)
	}

	found := false
	for _, r := range allResources {
		if r.ResourceID == resourceID {
			found = true
			break
		}
	}

	if !found {
		return false, 0, fmt.Errorf("resource with ID '%s' not found", resourceID)
	}

	liked, err := s.resourceRepo.ToggleLike(userID, resourceID)
	if err != nil {
		return false, 0, err
	}

	// Get the new count for the specific resource
	counts, err := s.resourceRepo.GetLikesByResourceIDs([]string{resourceID})
	if err != nil {
		// Even if we can't get the new count, the toggle itself succeeded.
		// We can return a count of 0 and log the error.
		return liked, 0, nil
	}

	return liked, counts[resourceID], nil
}

// GetResourceStats retrieves all necessary stats for the resource page.
func (s *resourceService) GetResourceStats(userID string) (*dto.ResourceStatsResponse, error) {
	// Get all resources from the database
	allResources, err := s.GetAllResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get all resources: %w", err)
	}

	// Dynamically build the list of resource IDs from the database records
	var allResourceIDs []string
	for _, r := range allResources {
		allResourceIDs = append(allResourceIDs, r.ResourceID)
	}

	// If there are no resources, return an empty response
	if len(allResourceIDs) == 0 {
		return &dto.ResourceStatsResponse{
			LikeCounts:  make(map[string]int64),
			UserLikes:   make(map[string]bool),
			Leaderboard: []dto.LeaderboardItem{},
		}, nil
	}

	// Get like counts for all resources
	likeCounts, err := s.resourceRepo.GetLikesByResourceIDs(allResourceIDs)
	if err != nil {
		return nil, err
	}

	// Get the current user's likes for all resources
	userLikes, err := s.resourceRepo.GetUserLikes(userID, allResourceIDs)
	if err != nil {
		return nil, err
	}

	// Get the leaderboard data
	leaderboardData, err := s.resourceRepo.GetLeaderboard()
	if err != nil {
		return nil, err
	}

	// Convert leaderboard data to the response format
	leaderboardItems := make([]dto.LeaderboardItem, len(leaderboardData))
	for i, item := range leaderboardData {
		leaderboardItems[i] = dto.LeaderboardItem{
			ResourceID: item.ResourceID,
			LikeCount:  item.TotalLikes, // Assuming the repository returns this field
		}
	}

	return &dto.ResourceStatsResponse{
		LikeCounts:  likeCounts,
		UserLikes:   userLikes,
		Leaderboard: leaderboardItems,
	}, nil
}

func (s *resourceService) DeleteResource(resourceID string) error {
	return s.resourceRepo.DeleteResource(resourceID)
}
