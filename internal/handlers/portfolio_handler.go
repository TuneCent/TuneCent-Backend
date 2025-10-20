package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

// PortfolioHandler handles portfolio-related endpoints
type PortfolioHandler struct {
	db *database.DB
}

func NewPortfolioHandler(db *database.DB) *PortfolioHandler {
	return &PortfolioHandler{db: db}
}

// GetPortfolio returns comprehensive portfolio overview
// GET /api/v1/portfolio/:address
func (h *PortfolioHandler) GetPortfolio(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// Get total music count
	var totalMusic int64
	h.db.Model(&models.MusicMetadata{}).
		Where("creator_address = ? AND is_active = ?", address, true).
		Count(&totalMusic)

	// Get total earnings
	var earnings struct {
		Total string
	}
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ?", address).
		Scan(&earnings)

	// Get total invested in campaigns
	var invested struct {
		Total string
	}
	h.db.Model(&models.Contribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Where("contributor_address = ?", address).
		Scan(&invested)

	// Get active campaigns count
	var activeCampaigns int64
	h.db.Model(&models.Campaign{}).
		Where("creator_address = ? AND status = ?", address, "active").
		Count(&activeCampaigns)

	// Get successful campaigns count
	var successfulCampaigns int64
	h.db.Model(&models.Campaign{}).
		Where("creator_address = ? AND status = ?", address, "successful").
		Count(&successfulCampaigns)

	// Get aggregate stats from music
	var musicStats struct {
		TotalPlays     uint64
		TotalViews     uint64
		TotalListeners uint64
		AvgViralScore  float64
	}
	h.db.Model(&models.MusicMetadata{}).
		Select(`
			COALESCE(SUM(play_count), 0) as total_plays,
			COALESCE(SUM(view_count), 0) as total_views,
			COALESCE(SUM(listener_count), 0) as total_listeners,
			COALESCE(AVG(viral_score), 0) as avg_viral_score
		`).
		Where("creator_address = ? AND is_active = ?", address, true).
		Scan(&musicStats)

	// Calculate portfolio value (mock calculation for PoC)
	// In production, calculate based on NFT floor prices, pending royalties, etc.
	portfolioValueETH := 15.5    // Mock value
	portfolioValueUSD := 38750.0 // Mock value at ~$2500/ETH

	// Get user info
	var user models.User
	h.db.Where("wallet_address = ?", address).First(&user)

	c.JSON(http.StatusOK, gin.H{
		"address":               address,
		"display_name":          user.DisplayName,
		"tier":                  user.Tier,
		"is_verified":           user.IsVerified,
		"total_music":           totalMusic,
		"total_earnings":        earnings.Total,
		"total_invested":        invested.Total,
		"active_campaigns":      activeCampaigns,
		"successful_campaigns":  successfulCampaigns,
		"portfolio_value_eth":   portfolioValueETH,
		"portfolio_value_usd":   portfolioValueUSD,
		"music_stats": gin.H{
			"total_plays":     musicStats.TotalPlays,
			"total_views":     musicStats.TotalViews,
			"total_listeners": musicStats.TotalListeners,
			"avg_viral_score": musicStats.AvgViralScore,
		},
	})
}

// GetGrowthStats returns growth statistics over time
// GET /api/v1/portfolio/:address/growth?period=month
func (h *PortfolioHandler) GetGrowthStats(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	period := c.DefaultQuery("period", "month") // week, month, year

	// Calculate period start date
	var periodStart time.Time
	now := time.Now()
	switch period {
	case "week":
		periodStart = now.AddDate(0, 0, -7)
	case "month":
		periodStart = now.AddDate(0, -1, 0)
	case "year":
		periodStart = now.AddDate(-1, 0, 0)
	default:
		periodStart = now.AddDate(0, -1, 0) // default to month
	}

	// Get earnings in current period
	var currentPeriodEarnings struct {
		Total string
	}
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ? AND royalty_distributions.distributed_at >= ?", address, periodStart).
		Scan(&currentPeriodEarnings)

	// Get earnings in previous period (for comparison)
	periodDuration := now.Sub(periodStart)
	previousPeriodStart := periodStart.Add(-periodDuration)
	var previousPeriodEarnings struct {
		Total string
	}
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ? AND royalty_distributions.distributed_at >= ? AND royalty_distributions.distributed_at < ?",
			address, previousPeriodStart, periodStart).
		Scan(&previousPeriodEarnings)

	// Get new music registered in period
	var newMusicCount int64
	h.db.Model(&models.MusicMetadata{}).
		Where("creator_address = ? AND created_at >= ?", address, periodStart).
		Count(&newMusicCount)

	// Get new campaigns in period
	var newCampaignsCount int64
	h.db.Model(&models.Campaign{}).
		Where("creator_address = ? AND created_at >= ?", address, periodStart).
		Count(&newCampaignsCount)

	// Mock growth percentages for PoC
	// In production, calculate from actual period comparisons
	earningsGrowth := 25.7    // %
	listenersGrowth := 18.3   // %
	playsGrowth := 22.1       // %
	campaignsGrowth := 150.0  // %

	c.JSON(http.StatusOK, gin.H{
		"period":                   period,
		"period_start":             periodStart,
		"period_end":               now,
		"current_period_earnings":  currentPeriodEarnings.Total,
		"previous_period_earnings": previousPeriodEarnings.Total,
		"new_music_count":          newMusicCount,
		"new_campaigns_count":      newCampaignsCount,
		"growth": gin.H{
			"earnings":   earningsGrowth,
			"listeners":  listenersGrowth,
			"plays":      playsGrowth,
			"campaigns":  campaignsGrowth,
		},
	})
}

// GetPerformanceMetrics returns detailed performance metrics
// GET /api/v1/portfolio/:address/performance
func (h *PortfolioHandler) GetPerformanceMetrics(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	type MusicPerformance struct {
		TokenID       uint64  `json:"token_id"`
		Title         string  `json:"title"`
		Artist        string  `json:"artist"`
		PlayCount     uint64  `json:"play_count"`
		ViewCount     uint64  `json:"view_count"`
		ListenerCount uint64  `json:"listener_count"`
		ViralScore    float64 `json:"viral_score"`
		TotalEarnings string  `json:"total_earnings"`
	}

	var performance []MusicPerformance
	h.db.Table("music_metadata m").
		Select(`
			m.token_id,
			m.title,
			m.artist,
			m.play_count,
			m.view_count,
			m.listener_count,
			m.viral_score,
			COALESCE(SUM(CAST(rd.amount AS DECIMAL(30,0))), 0) as total_earnings
		`).
		Joins("LEFT JOIN royalty_distributions rd ON m.token_id = rd.token_id").
		Where("m.creator_address = ? AND m.is_active = ?", address, true).
		Group("m.token_id").
		Order("m.viral_score DESC, m.play_count DESC").
		Scan(&performance)

	// Calculate performance stats
	var bestPerformer MusicPerformance
	if len(performance) > 0 {
		bestPerformer = performance[0]
	}

	c.JSON(http.StatusOK, gin.H{
		"music_performance": performance,
		"total_tracks":      len(performance),
		"best_performer":    bestPerformer,
	})
}

// GetPoolsInvested returns campaigns the user has invested in
// GET /api/v1/portfolio/:address/pools
func (h *PortfolioHandler) GetPoolsInvested(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	type PoolInvestment struct {
		CampaignID        uint64  `json:"campaign_id"`
		MusicTitle        string  `json:"music_title"`
		MusicArtist       string  `json:"music_artist"`
		AmountInvested    string  `json:"amount_invested"`
		SharePercentage   float64 `json:"share_percentage"`
		Status            string  `json:"status"`
		RoyaltyPercentage uint16  `json:"royalty_percentage"`
		ContributedAt     time.Time `json:"contributed_at"`
	}

	var investments []PoolInvestment
	h.db.Table("contributions c").
		Select(`
			c.campaign_id,
			m.title as music_title,
			m.artist as music_artist,
			c.amount as amount_invested,
			c.share_percentage,
			camp.status,
			camp.royalty_percentage,
			c.contributed_at
		`).
		Joins("JOIN campaigns camp ON c.campaign_id = camp.campaign_id").
		Joins("JOIN music_metadata m ON camp.token_id = m.token_id").
		Where("c.contributor_address = ?", address).
		Order("c.contributed_at DESC").
		Scan(&investments)

	// Calculate total invested
	var totalInvested struct {
		Total string
	}
	h.db.Model(&models.Contribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Where("contributor_address = ?", address).
		Scan(&totalInvested)

	c.JSON(http.StatusOK, gin.H{
		"investments":    investments,
		"total_pools":    len(investments),
		"total_invested": totalInvested.Total,
	})
}
