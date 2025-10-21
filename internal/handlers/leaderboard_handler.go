package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

// LeaderboardHandler handles leaderboard-related endpoints
type LeaderboardHandler struct {
	db *database.DB
}

func NewLeaderboardHandler(db *database.DB) *LeaderboardHandler {
	return &LeaderboardHandler{db: db}
}

// GetTopArtists returns top artists leaderboard
// GET /api/v1/leaderboard/top-artists?limit=10
func (h *LeaderboardHandler) GetTopArtists(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	type LeaderboardEntry struct {
		Rank            int     `json:"rank"`
		WalletAddress   string  `json:"wallet_address"`
		DisplayName     string  `json:"display_name"`
		Tier            string  `json:"tier"`
		IsVerified      bool    `json:"is_verified"`
		TotalWorks      uint64  `json:"total_works"`
		TotalEarnings   string  `json:"total_earnings"`
		TotalCampaigns  uint64  `json:"total_campaigns"`
		Score           float64 `json:"score"`
	}

	var leaderboard []LeaderboardEntry

	// Calculate leaderboard on the fly
	h.db.Table("users u").
		Select(`
			u.wallet_address,
			u.username as display_name,
			'starter' as tier,
			u.is_verified,
			COUNT(DISTINCT m.token_id) as total_works,
			COALESCE(SUM(CAST(rd.amount AS DECIMAL(30,0))), 0) as total_earnings,
			COUNT(DISTINCT c.campaign_id) as total_campaigns,
			(COUNT(DISTINCT m.token_id) * 100 +
			 COALESCE(SUM(CAST(rd.amount AS DECIMAL(30,0))) / 1e18, 0) * 10 +
			 COUNT(DISTINCT c.campaign_id) * 50) as score
		`).
		Joins("LEFT JOIN music_metadata m ON u.wallet_address = m.creator_address").
		Joins("LEFT JOIN royalty_distributions rd ON m.token_id = rd.token_id AND rd.beneficiary = u.wallet_address").
		Joins("LEFT JOIN campaigns c ON u.wallet_address = c.creator_address").
		Where("u.role IN (?)", []string{"creator", "both"}).
		Group("u.wallet_address").
		Order("score DESC").
		Limit(limit).
		Scan(&leaderboard)

	// Add rank numbers after fetching
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	// If no results, return empty array
	if leaderboard == nil {
		leaderboard = []LeaderboardEntry{}
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
		"total":       len(leaderboard),
	})
}

// GetUserRank returns a user's position in the leaderboard
// GET /api/v1/leaderboard/:address/rank
func (h *LeaderboardHandler) GetUserRank(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// Get user's stats
	var user models.User
	if err := h.db.Where("wallet_address = ?", address).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Calculate user's score
	var userStats struct {
		TotalWorks     uint64
		TotalEarnings  string
		TotalCampaigns uint64
		Score          float64
	}

	h.db.Table("users u").
		Select(`
			COUNT(DISTINCT m.token_id) as total_works,
			COALESCE(SUM(CAST(rd.amount AS DECIMAL(30,0))), 0) as total_earnings,
			COUNT(DISTINCT c.campaign_id) as total_campaigns,
			(COUNT(DISTINCT m.token_id) * 100 +
			 COALESCE(SUM(CAST(rd.amount AS DECIMAL(30,0))) / 1e18, 0) * 10 +
			 COUNT(DISTINCT c.campaign_id) * 50) as score
		`).
		Joins("LEFT JOIN music_metadata m ON u.wallet_address = m.creator_address").
		Joins("LEFT JOIN royalty_distributions rd ON m.token_id = rd.token_id").
		Joins("LEFT JOIN campaigns c ON u.wallet_address = c.creator_address").
		Where("u.wallet_address = ?", address).
		Group("u.wallet_address").
		Scan(&userStats)

	// Calculate rank (count how many users have higher scores)
	var rank int64
	h.db.Table("users u").
		Select("COUNT(DISTINCT u.wallet_address)").
		Joins("LEFT JOIN music_metadata m ON u.wallet_address = m.creator_address").
		Joins("LEFT JOIN royalty_distributions rd ON m.token_id = rd.token_id").
		Joins("LEFT JOIN campaigns c ON u.wallet_address = c.creator_address").
		Where("u.role IN (?)", []string{"creator", "both"}).
		Group("u.wallet_address").
		Having(`(COUNT(DISTINCT m.token_id) * 100 +
			    COALESCE(SUM(CAST(rd.amount AS DECIMAL(30,0))) / 1e18, 0) * 10 +
			    COUNT(DISTINCT c.campaign_id) * 50) > ?`, userStats.Score).
		Count(&rank)

	// Rank is count + 1 (number of people ahead + 1)
	userRank := rank + 1

	c.JSON(http.StatusOK, gin.H{
		"address":        address,
		"rank":           userRank,
		"display_name":   user.DisplayName,
		"tier":           user.Tier,
		"is_verified":    user.IsVerified,
		"total_works":    userStats.TotalWorks,
		"total_earnings": userStats.TotalEarnings,
		"total_campaigns": userStats.TotalCampaigns,
		"score":          userStats.Score,
	})
}

// GetLeaderboardStats returns overall leaderboard statistics
// GET /api/v1/leaderboard/stats
func (h *LeaderboardHandler) GetLeaderboardStats(c *gin.Context) {
	var stats struct {
		TotalCreators   int64
		TotalWorks      int64
		TotalEarnings   string
		VerifiedCreators int64
	}

	// Total creators
	h.db.Model(&models.User{}).
		Where("role IN (?)", []string{"creator", "both"}).
		Count(&stats.TotalCreators)

	// Total works
	h.db.Model(&models.MusicMetadata{}).
		Where("is_active = ?", true).
		Count(&stats.TotalWorks)

	// Total earnings
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Scan(&stats.TotalEarnings)

	// Verified creators
	h.db.Model(&models.User{}).
		Where("role IN (?) AND is_verified = ?", []string{"creator", "both"}, true).
		Count(&stats.VerifiedCreators)

	c.JSON(http.StatusOK, gin.H{
		"total_creators":    stats.TotalCreators,
		"total_works":       stats.TotalWorks,
		"total_earnings":    stats.TotalEarnings,
		"verified_creators": stats.VerifiedCreators,
	})
}
