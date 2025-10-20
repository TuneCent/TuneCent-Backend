package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

// DashboardHandler handles dashboard-related endpoints
type DashboardHandler struct {
	db *database.DB
}

func NewDashboardHandler(db *database.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetOverview returns dashboard overview stats for a creator
// GET /api/v1/dashboard/overview?address=0x...
func (h *DashboardHandler) GetOverview(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// Get total music count
	var musicCount int64
	h.db.Model(&models.MusicMetadata{}).
		Where("creator_address = ? AND is_active = ?", address, true).
		Count(&musicCount)

	// Get total royalties earned
	var totalEarnings string
	var royaltySum struct {
		Total string
	}
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ?", address).
		Scan(&royaltySum)
	totalEarnings = royaltySum.Total

	// Get total listeners (sum from music metadata)
	var listenerStats struct {
		TotalListeners uint64
		TotalViews     uint64
		TotalPlays     uint64
	}
	h.db.Model(&models.MusicMetadata{}).
		Select("COALESCE(SUM(listener_count), 0) as total_listeners, COALESCE(SUM(view_count), 0) as total_views, COALESCE(SUM(play_count), 0) as total_plays").
		Where("creator_address = ?", address).
		Scan(&listenerStats)

	// Get active campaign count
	var activeCampaigns int64
	h.db.Model(&models.Campaign{}).
		Where("creator_address = ? AND status = ?", address, "active").
		Count(&activeCampaigns)

	// Get successful campaign count
	var successfulCampaigns int64
	h.db.Model(&models.Campaign{}).
		Where("creator_address = ? AND status = ?", address, "successful").
		Count(&successfulCampaigns)

	// Get user tier and verified status
	var user models.User
	h.db.Where("wallet_address = ?", address).First(&user)

	c.JSON(http.StatusOK, gin.H{
		"address":              address,
		"total_music":          musicCount,
		"total_earnings":       totalEarnings,
		"total_listeners":      listenerStats.TotalListeners,
		"total_views":          listenerStats.TotalViews,
		"total_plays":          listenerStats.TotalPlays,
		"active_campaigns":     activeCampaigns,
		"successful_campaigns": successfulCampaigns,
		"tier":                 user.Tier,
		"is_verified":          user.IsVerified,
		"leaderboard_rank":     user.LeaderboardRank,
	})
}

// GetQuickStats returns quick stats for dashboard cards
// GET /api/v1/dashboard/quick-stats?address=0x...
func (h *DashboardHandler) GetQuickStats(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// Get today's earnings (last royalty payment)
	var todayEarnings string
	var lastRoyalty struct {
		Amount string
	}
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("amount").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ?", address).
		Order("royalty_distributions.distributed_at DESC").
		Limit(1).
		Scan(&lastRoyalty)
	todayEarnings = lastRoyalty.Amount
	if todayEarnings == "" {
		todayEarnings = "0"
	}

	// Get weekly growth (mock calculation based on recent activity)
	weeklyGrowth := 15.5 // Mock value for PoC

	// Get new listeners this week (mock)
	newListeners := uint64(1250) // Mock value for PoC

	// Get trending songs count (where trending_rank > 0)
	var trendingSongs int64
	h.db.Model(&models.MusicMetadata{}).
		Where("creator_address = ? AND trending_rank > ?", address, 0).
		Count(&trendingSongs)

	c.JSON(http.StatusOK, gin.H{
		"today_earnings":   todayEarnings,
		"weekly_growth":    weeklyGrowth,
		"new_listeners":    newListeners,
		"trending_songs":   trendingSongs,
	})
}

// GetTrendingPools returns trending crowdfunding pools
// GET /api/v1/dashboard/trending-pools?limit=5
func (h *DashboardHandler) GetTrendingPools(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "5")
	limit, _ := strconv.Atoi(limitStr)

	type PoolWithMusic struct {
		models.Campaign
		MusicTitle       string  `json:"music_title"`
		MusicArtist      string  `json:"music_artist"`
		CreatorName      string  `json:"creator_name"`
		CreatorVerified  bool    `json:"creator_verified"`
		FundingPercentage float64 `json:"funding_percentage"`
	}

	var pools []PoolWithMusic
	h.db.Table("campaigns").
		Select(`campaigns.*,
			music_metadata.title as music_title,
			music_metadata.artist as music_artist,
			users.display_name as creator_name,
			users.is_verified as creator_verified,
			(CAST(campaigns.raised_amount AS DECIMAL(30,0)) / CAST(campaigns.goal_amount AS DECIMAL(30,0)) * 100) as funding_percentage`).
		Joins("JOIN music_metadata ON campaigns.token_id = music_metadata.token_id").
		Joins("JOIN users ON campaigns.creator_address = users.wallet_address").
		Where("campaigns.status = ? AND campaigns.is_trending = ?", "active", true).
		Order("funding_percentage DESC, campaigns.created_at DESC").
		Limit(limit).
		Scan(&pools)

	c.JSON(http.StatusOK, gin.H{
		"pools": pools,
		"total": len(pools),
	})
}

// GetRecentActivities returns recent activities feed
// GET /api/v1/dashboard/activities?address=0x...&limit=10
func (h *DashboardHandler) GetRecentActivities(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	var activities []models.Activity
	h.db.Where("user_address = ?", address).
		Order("created_at DESC").
		Limit(limit).
		Find(&activities)

	c.JSON(http.StatusOK, gin.H{
		"activities": activities,
		"total":      len(activities),
	})
}

// GetMusicTrends returns music trends chart data
// GET /api/v1/dashboard/music-trends?address=0x...&days=30
func (h *DashboardHandler) GetMusicTrends(c *gin.Context) {
	address := c.Query("address")
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)

	// Get all music for this creator with stats
	type MusicTrend struct {
		TokenID       uint64  `json:"token_id"`
		Title         string  `json:"title"`
		PlayCount     uint64  `json:"play_count"`
		ViewCount     uint64  `json:"view_count"`
		ListenerCount uint64  `json:"listener_count"`
		ViralScore    float64 `json:"viral_score"`
		TrendingRank  int     `json:"trending_rank"`
	}

	var trends []MusicTrend
	query := h.db.Table("music_metadata").
		Select("token_id, title, play_count, view_count, listener_count, viral_score, trending_rank").
		Where("creator_address = ? AND is_active = ?", address, true).
		Order("play_count DESC")

	if address != "" {
		query.Find(&trends)
	}

	c.JSON(http.StatusOK, gin.H{
		"trends": trends,
		"period": days,
	})
}

// GetViralPerformance returns viral performance showcase
// GET /api/v1/dashboard/viral-performance?address=0x...
func (h *DashboardHandler) GetViralPerformance(c *gin.Context) {
	address := c.Query("address")

	type ViralMusic struct {
		TokenID       uint64  `json:"token_id"`
		Title         string  `json:"title"`
		Artist        string  `json:"artist"`
		ViralScore    float64 `json:"viral_score"`
		ViewCount     uint64  `json:"view_count"`
		PlayCount     uint64  `json:"play_count"`
		ListenerCount uint64  `json:"listener_count"`
		TrendingRank  int     `json:"trending_rank"`
	}

	var viralMusic []ViralMusic
	query := h.db.Table("music_metadata").
		Select("token_id, title, artist, viral_score, view_count, play_count, listener_count, trending_rank").
		Where("is_active = ? AND viral_score > ?", true, 50.0). // Only high viral scores
		Order("viral_score DESC, trending_rank ASC")

	if address != "" {
		query = query.Where("creator_address = ?", address)
	}

	query.Limit(10).Find(&viralMusic)

	c.JSON(http.StatusOK, gin.H{
		"viral_music": viralMusic,
		"threshold":   50.0,
	})
}

// GetWeeklyProgress returns weekly progress indicators
// GET /api/v1/dashboard/weekly-progress?address=0x...
func (h *DashboardHandler) GetWeeklyProgress(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// For PoC, return mock weekly progress data
	// In production, this would calculate actual week-over-week changes
	c.JSON(http.StatusOK, gin.H{
		"listeners_growth":  12.5, // percentage
		"plays_growth":      18.3,
		"earnings_growth":   25.7,
		"engagement_growth": 15.2,
		"week_start":        "2025-10-13",
		"week_end":          "2025-10-20",
	})
}

// GetRoyaltyPulse returns live royalty pulse data
// GET /api/v1/dashboard/royalty-pulse?address=0x...
func (h *DashboardHandler) GetRoyaltyPulse(c *gin.Context) {
	address := c.Query("address")

	// Get recent royalty payments (last 24 hours or last 10)
	type RoyaltyPulse struct {
		TokenID   uint64    `json:"token_id"`
		Title     string    `json:"title"`
		Amount    string    `json:"amount"`
		Platform  string    `json:"platform"`
		PaidAt    string    `json:"paid_at"`
	}

	var pulseData []RoyaltyPulse
	query := h.db.Table("royalty_payments").
		Select("royalty_payments.token_id, music_metadata.title, royalty_payments.amount, royalty_payments.platform, royalty_payments.paid_at").
		Joins("JOIN music_metadata ON royalty_payments.token_id = music_metadata.token_id").
		Where("royalty_payments.is_distributed = ?", true).
		Order("royalty_payments.paid_at DESC").
		Limit(10)

	if address != "" {
		query = query.Where("music_metadata.creator_address = ?", address)
	}

	query.Scan(&pulseData)

	// Calculate total in pulse period
	var totalPulse string
	h.db.Table("royalty_payments").
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_payments.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ? AND royalty_payments.paid_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)", address).
		Scan(&totalPulse)

	c.JSON(http.StatusOK, gin.H{
		"pulse_data":      pulseData,
		"total_24h":       totalPulse,
		"payment_count":   len(pulseData),
	})
}
