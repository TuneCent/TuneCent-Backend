package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
	"github.com/tunecent/backend/pkg/mockdata"
)

// AnalyticsHandler handles analytics-related endpoints
type AnalyticsHandler struct {
	db *database.DB
}

func NewAnalyticsHandler(db *database.DB) *AnalyticsHandler {
	return &AnalyticsHandler{db: db}
}

// GetPlatformStats returns platform-specific statistics (Spotify, TikTok, Apple Music)
// GET /api/v1/analytics/:tokenId/platform-stats
func (h *AnalyticsHandler) GetPlatformStats(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Get music metadata to get registration date
	var music models.MusicMetadata
	if err := h.db.Where("token_id = ?", tokenID).First(&music).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
		return
	}

	// Generate mock platform stats
	stats := mockdata.GeneratePlatformStats(tokenID, music.RegisteredAt)

	c.JSON(http.StatusOK, gin.H{
		"token_id": tokenID,
		"title":    music.Title,
		"artist":   music.Artist,
		"stats":    stats,
	})
}

// GetViralScore returns viral score and breakdown
// GET /api/v1/analytics/:tokenId/viral-score
func (h *AnalyticsHandler) GetViralScore(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Get music metadata
	var music models.MusicMetadata
	if err := h.db.Where("token_id = ?", tokenID).First(&music).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_id":       tokenID,
		"viral_score":    music.ViralScore,
		"trending_rank":  music.TrendingRank,
		"is_trending":    music.TrendingRank > 0,
		"play_count":     music.PlayCount,
		"view_count":     music.ViewCount,
		"listener_count": music.ListenerCount,
	})
}

// GetGrowthMetrics returns growth percentages over time
// GET /api/v1/analytics/:tokenId/growth?period=week
func (h *AnalyticsHandler) GetGrowthMetrics(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	period := c.DefaultQuery("period", "week") // week, month, all

	// Get analytics data
	var analytics models.Analytics
	if err := h.db.Where("token_id = ?", tokenID).First(&analytics).Error; err != nil {
		// If no analytics exist, return zeros
		c.JSON(http.StatusOK, gin.H{
			"token_id": tokenID,
			"period":   period,
			"growth": gin.H{
				"spotify":      0,
				"tiktok":       0,
				"apple_music":  0,
				"overall":      0,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_id": tokenID,
		"period":   period,
		"growth": gin.H{
			"spotify":      analytics.SpotifyGrowth,
			"tiktok":       analytics.TikTokGrowth,
			"apple_music":  analytics.AppleMusicGrowth,
			"overall":      analytics.WeeklyGrowth,
		},
	})
}

// GetListenerMetrics returns listener counts over time
// GET /api/v1/analytics/:tokenId/listeners
func (h *AnalyticsHandler) GetListenerMetrics(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Get music and analytics
	var music models.MusicMetadata
	if err := h.db.Where("token_id = ?", tokenID).First(&music).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
		return
	}

	// For PoC, return mock historical data (in production, track daily)
	// Generate a trend based on current value
	dailyListeners := make([]uint64, 30)
	current := music.ListenerCount
	for i := 29; i >= 0; i-- {
		// Simulate growth over 30 days
		growth := float64(i) / 30.0
		dailyListeners[29-i] = uint64(float64(current) * growth)
	}

	c.JSON(http.StatusOK, gin.H{
		"token_id":         tokenID,
		"current":          music.ListenerCount,
		"daily_trend":      dailyListeners,
		"period_days":      30,
	})
}

// GetViewMetrics returns view counts over time
// GET /api/v1/analytics/:tokenId/views
func (h *AnalyticsHandler) GetViewMetrics(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Get music
	var music models.MusicMetadata
	if err := h.db.Where("token_id = ?", tokenID).First(&music).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token_id":     tokenID,
		"total_views":  music.ViewCount,
		"total_plays":  music.PlayCount,
		"view_to_play_ratio": func() float64 {
			if music.PlayCount == 0 {
				return 0
			}
			return float64(music.ViewCount) / float64(music.PlayCount)
		}(),
	})
}

// GetTopSongs returns top ranked songs globally or for a creator
// GET /api/v1/analytics/global/top-songs?address=0x...&limit=10
func (h *AnalyticsHandler) GetTopSongs(c *gin.Context) {
	address := c.Query("address") // Optional: filter by creator
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	type TopSong struct {
		TokenID       uint64  `json:"token_id"`
		Title         string  `json:"title"`
		Artist        string  `json:"artist"`
		CreatorAddress string  `json:"creator_address"`
		ViralScore    float64 `json:"viral_score"`
		PlayCount     uint64  `json:"play_count"`
		ViewCount     uint64  `json:"view_count"`
		TrendingRank  int     `json:"trending_rank"`
	}

	var topSongs []TopSong
	query := h.db.Table("music_metadata").
		Select("token_id, title, artist, creator_address, viral_score, play_count, view_count, trending_rank").
		Where("is_active = ?", true).
		Order("viral_score DESC, play_count DESC")

	if address != "" {
		query = query.Where("creator_address = ?", address)
	}

	query.Limit(limit).Scan(&topSongs)

	c.JSON(http.StatusOK, gin.H{
		"top_songs": topSongs,
		"total":     len(topSongs),
	})
}

// GetTrendingIndicators returns trending indicators for a song
// GET /api/v1/analytics/:tokenId/trending
func (h *AnalyticsHandler) GetTrendingIndicators(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Get music
	var music models.MusicMetadata
	if err := h.db.Where("token_id = ?", tokenID).First(&music).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
		return
	}

	// Generate platform stats for trending determination
	platformStats := mockdata.GeneratePlatformStats(tokenID, music.RegisteredAt)

	// Determine which platforms are trending
	trendingPlatforms := []string{}
	if platformStats.Spotify.Growth > 300 {
		trendingPlatforms = append(trendingPlatforms, "Spotify")
	}
	if platformStats.TikTok.Growth > 500 {
		trendingPlatforms = append(trendingPlatforms, "TikTok")
	}
	if platformStats.AppleMusic.Growth > 200 {
		trendingPlatforms = append(trendingPlatforms, "Apple Music")
	}

	c.JSON(http.StatusOK, gin.H{
		"token_id":            tokenID,
		"is_trending":         music.TrendingRank > 0,
		"trending_rank":       music.TrendingRank,
		"viral_score":         music.ViralScore,
		"trending_platforms":  trendingPlatforms,
		"momentum":           len(trendingPlatforms) >= 2, // Trending on 2+ platforms
	})
}

// GetEstimatedReach returns estimated reach calculations
// GET /api/v1/analytics/:tokenId/reach
func (h *AnalyticsHandler) GetEstimatedReach(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	// Get music
	var music models.MusicMetadata
	if err := h.db.Where("token_id = ?", tokenID).First(&music).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
		return
	}

	// Generate platform stats
	platformStats := mockdata.GeneratePlatformStats(tokenID, music.RegisteredAt)
	estimatedReach := mockdata.GenerateEstimatedReach(platformStats)

	c.JSON(http.StatusOK, gin.H{
		"token_id":         tokenID,
		"estimated_reach":  estimatedReach,
		"breakdown": gin.H{
			"spotify_listeners":    platformStats.Spotify.Listeners,
			"tiktok_views":         platformStats.TikTok.Views,
			"apple_music_listeners": platformStats.AppleMusic.Listeners,
		},
		"methodology": "Estimated unique reach accounting for 30% cross-platform overlap",
	})
}
