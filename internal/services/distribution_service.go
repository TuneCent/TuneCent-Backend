package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

type DistributionService struct {
	db *database.DB
}

func NewDistributionService(db *database.DB) *DistributionService {
	return &DistributionService{db: db}
}

type SubmitDistributionRequest struct {
	TokenID     uint64   `json:"token_id" binding:"required"`
	UserAddress string   `json:"user_address" binding:"required"`
	Platforms   []string `json:"platforms" binding:"required"`
}

type DistributionStatusResponse struct {
	TokenID      uint64                      `json:"token_id"`
	Status       string                      `json:"status"`
	SubmittedAt  time.Time                   `json:"submitted_at"`
	Platforms    []PlatformStatus            `json:"platforms"`
}

type PlatformStatus struct {
	Platform      string     `json:"platform"`
	Status        string     `json:"status"`
	ExternalID    string     `json:"external_id,omitempty"`
	ExternalURL   string     `json:"external_url,omitempty"`
	DistributedAt *time.Time `json:"distributed_at,omitempty"`
}

func (s *DistributionService) SubmitDistribution(ctx context.Context, req *SubmitDistributionRequest) (*models.DistributionSubmission, error) {
	// Check if music exists
	var music models.MusicMetadata
	if err := s.db.Where("token_id = ?", req.TokenID).First(&music).Error; err != nil {
		return nil, fmt.Errorf("music not found: %w", err)
	}

	// Check if already submitted
	var existing models.DistributionSubmission
	if err := s.db.Where("token_id = ? AND status NOT IN ('failed', 'cancelled')", req.TokenID).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("distribution already submitted for this track")
	}

	// Convert platforms to JSON
	platformsJSON, err := json.Marshal(req.Platforms)
	if err != nil {
		return nil, fmt.Errorf("failed to encode platforms: %w", err)
	}

	// Create submission
	submission := &models.DistributionSubmission{
		TokenID:     req.TokenID,
		UserAddress: req.UserAddress,
		Platforms:   string(platformsJSON),
		Status:      "processing",
		SubmittedAt: time.Now(),
	}

	if err := s.db.Create(submission).Error; err != nil {
		return nil, fmt.Errorf("failed to create distribution submission: %w", err)
	}

	// Create platform distribution records
	for _, platform := range req.Platforms {
		platformDist := &models.PlatformDistribution{
			TokenID:  req.TokenID,
			Platform: platform,
			Status:   "pending",
		}
		s.db.Create(platformDist)
	}

	return submission, nil
}

func (s *DistributionService) GetDistributionStatus(ctx context.Context, tokenID uint64) (*DistributionStatusResponse, error) {
	// Get submission
	var submission models.DistributionSubmission
	if err := s.db.Where("token_id = ?", tokenID).Order("created_at DESC").First(&submission).Error; err != nil {
		return nil, fmt.Errorf("distribution not found: %w", err)
	}

	// Get platform distributions
	var platformDists []models.PlatformDistribution
	s.db.Where("token_id = ?", tokenID).Find(&platformDists)

	// Build response
	platforms := make([]PlatformStatus, len(platformDists))
	for i, pd := range platformDists {
		platforms[i] = PlatformStatus{
			Platform:      pd.Platform,
			Status:        pd.Status,
			ExternalID:    pd.ExternalID,
			ExternalURL:   pd.ExternalURL,
			DistributedAt: pd.DistributedAt,
		}
	}

	return &DistributionStatusResponse{
		TokenID:     submission.TokenID,
		Status:      submission.Status,
		SubmittedAt: submission.SubmittedAt,
		Platforms:   platforms,
	}, nil
}

func (s *DistributionService) GetPlatformStatus(ctx context.Context, tokenID uint64, platform string) (*models.PlatformDistribution, error) {
	var platformDist models.PlatformDistribution
	if err := s.db.Where("token_id = ? AND platform = ?", tokenID, platform).First(&platformDist).Error; err != nil {
		return nil, fmt.Errorf("platform distribution not found: %w", err)
	}
	return &platformDist, nil
}

func (s *DistributionService) UpdatePlatformStatus(ctx context.Context, tokenID uint64, platform string, status string, externalID string, externalURL string) error {
	var platformDist models.PlatformDistribution
	if err := s.db.Where("token_id = ? AND platform = ?", tokenID, platform).First(&platformDist).Error; err != nil {
		return fmt.Errorf("platform distribution not found: %w", err)
	}

	platformDist.Status = status
	platformDist.ExternalID = externalID
	platformDist.ExternalURL = externalURL

	if status == "live" {
		now := time.Now()
		platformDist.DistributedAt = &now
	}

	return s.db.Save(&platformDist).Error
}

func (s *DistributionService) ListDistributions(ctx context.Context, userAddress string, limit, offset int) ([]*models.DistributionSubmission, int64, error) {
	var submissions []*models.DistributionSubmission
	var total int64

	query := s.db.Model(&models.DistributionSubmission{})
	if userAddress != "" {
		query = query.Where("user_address = ?", userAddress)
	}

	query.Count(&total)
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&submissions)

	return submissions, total, nil
}
