package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

type ReinvestmentService struct {
	db *database.DB
}

func NewReinvestmentService(db *database.DB) *ReinvestmentService {
	return &ReinvestmentService{db: db}
}

type SuggestionResponse struct {
	UserAddress     string                   `json:"user_address"`
	AvailableFunds  string                   `json:"available_funds"`
	SuggestedPools  []SuggestedPool          `json:"suggested_pools"`
	TotalExpectedROI float64                  `json:"total_expected_roi"`
}

type SuggestedPool struct {
	CampaignID        uint64  `json:"campaign_id"`
	TokenID           uint64  `json:"token_id"`
	MusicTitle        string  `json:"music_title"`
	MusicArtist       string  `json:"music_artist"`
	RoyaltyPercentage uint16  `json:"royalty_percentage"`
	EstimatedROI      float64 `json:"estimated_roi"`
	RiskScore         uint8   `json:"risk_score"`
	Reasoning         string  `json:"reasoning"`
}

type QuickReinvestRequest struct {
	UserAddress  string `json:"user_address" binding:"required"`
	CampaignID   uint64 `json:"campaign_id" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	FromSource   string `json:"from_source" binding:"required"`
}

func (s *ReinvestmentService) GetSuggestions(ctx context.Context, userAddress string) (*SuggestionResponse, error) {
	// Calculate available funds
	var totalEarnings struct {
		Total string
	}
	s.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ?", userAddress).
		Scan(&totalEarnings)

	var totalInvested struct {
		Total string
	}
	s.db.Model(&models.Contribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Where("contributor_address = ?", userAddress).
		Scan(&totalInvested)

	availableFunds := totalEarnings.Total // Simplified for PoC

	// Get active campaigns with good metrics
	type CampaignData struct {
		CampaignID        uint64
		TokenID           uint64
		MusicTitle        string
		MusicArtist       string
		RoyaltyPercentage uint16
		EstimatedROI      float64
		RiskScore         uint8
		RaisedAmount      string
		GoalAmount        string
	}

	var campaigns []CampaignData
	s.db.Table("campaigns").
		Select(`campaigns.campaign_id, campaigns.token_id, campaigns.royalty_percentage,
			campaigns.estimated_roi, campaigns.risk_score, campaigns.raised_amount, campaigns.goal_amount,
			music_metadata.title as music_title, music_metadata.artist as music_artist`).
		Joins("JOIN music_metadata ON campaigns.token_id = music_metadata.token_id").
		Where("campaigns.status = ? AND campaigns.risk_score < ?", "active", 70).
		Order("campaigns.estimated_roi DESC, campaigns.risk_score ASC").
		Limit(5).
		Scan(&campaigns)

	// Build suggestions
	suggestions := make([]SuggestedPool, len(campaigns))
	totalROI := 0.0

	for i, camp := range campaigns {
		reasoning := fmt.Sprintf("High ROI potential (%.1f%%) with low risk score (%d/100). Currently %.0f%% funded.",
			camp.EstimatedROI, camp.RiskScore, 75.0) // Mock funded percentage

		suggestions[i] = SuggestedPool{
			CampaignID:        camp.CampaignID,
			TokenID:           camp.TokenID,
			MusicTitle:        camp.MusicTitle,
			MusicArtist:       camp.MusicArtist,
			RoyaltyPercentage: camp.RoyaltyPercentage,
			EstimatedROI:      camp.EstimatedROI,
			RiskScore:         camp.RiskScore,
			Reasoning:         reasoning,
		}
		totalROI += camp.EstimatedROI
	}

	avgROI := 0.0
	if len(suggestions) > 0 {
		avgROI = totalROI / float64(len(suggestions))
	}

	// Save suggestion
	poolIDs := make([]uint64, len(campaigns))
	for i, camp := range campaigns {
		poolIDs[i] = camp.CampaignID
	}
	poolIDsJSON, _ := json.Marshal(poolIDs)

	suggestion := &models.ReinvestmentSuggestion{
		UserAddress:    userAddress,
		AvailableFunds: availableFunds,
		SuggestedPools: string(poolIDsJSON),
		ExpectedROI:    avgROI,
		Reasoning:      fmt.Sprintf("Top %d performing pools based on ROI and risk", len(suggestions)),
	}
	s.db.Create(suggestion)

	return &SuggestionResponse{
		UserAddress:      userAddress,
		AvailableFunds:   availableFunds,
		SuggestedPools:   suggestions,
		TotalExpectedROI: avgROI,
	}, nil
}

func (s *ReinvestmentService) QuickReinvest(ctx context.Context, req *QuickReinvestRequest) (*models.ReinvestmentHistory, error) {
	// Verify campaign exists and is active
	var campaign models.Campaign
	if err := s.db.Where("campaign_id = ? AND status = ?", req.CampaignID, "active").First(&campaign).Error; err != nil {
		return nil, fmt.Errorf("campaign not found or not active: %w", err)
	}

	// Create reinvestment history record
	history := &models.ReinvestmentHistory{
		UserAddress:  req.UserAddress,
		FromSource:   req.FromSource,
		ToCampaignID: req.CampaignID,
		Amount:       req.Amount,
		TxHash:       fmt.Sprintf("0x%064x", time.Now().UnixNano()), // Mock tx hash
	}

	if err := s.db.Create(history).Error; err != nil {
		return nil, fmt.Errorf("failed to create reinvestment history: %w", err)
	}

	// Create contribution record
	contribution := &models.Contribution{
		CampaignID:         req.CampaignID,
		ContributorAddress: req.UserAddress,
		Amount:             req.Amount,
		SharePercentage:    0, // Calculate based on total
		TxHash:             history.TxHash,
		ContributedAt:      time.Now(),
	}
	s.db.Create(contribution)

	return history, nil
}

func (s *ReinvestmentService) GetReinvestmentHistory(ctx context.Context, userAddress string, limit, offset int) ([]*models.ReinvestmentHistory, int64, error) {
	var history []*models.ReinvestmentHistory
	var total int64

	query := s.db.Model(&models.ReinvestmentHistory{}).Where("user_address = ?", userAddress)
	query.Count(&total)
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&history)

	return history, total, nil
}

func (s *ReinvestmentService) GetReinvestmentStats(ctx context.Context, userAddress string) (map[string]interface{}, error) {
	// Get total reinvested
	var totalReinvested struct {
		Total string
		Count int64
	}
	s.db.Model(&models.ReinvestmentHistory{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total, COUNT(*) as count").
		Where("user_address = ?", userAddress).
		Scan(&totalReinvested)

	// Get average ROI from reinvested pools
	var avgROI struct {
		Avg float64
	}
	s.db.Table("reinvestment_histories rh").
		Select("COALESCE(AVG(c.estimated_roi), 0) as avg").
		Joins("JOIN campaigns c ON rh.to_campaign_id = c.campaign_id").
		Where("rh.user_address = ?", userAddress).
		Scan(&avgROI)

	return map[string]interface{}{
		"total_reinvested":      totalReinvested.Total,
		"reinvestment_count":    totalReinvested.Count,
		"average_expected_roi":  avgROI.Avg,
	}, nil
}
