package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/services"
)

type MusicHandler struct {
	musicService *services.MusicService
}

func NewMusicHandler(musicService *services.MusicService) *MusicHandler {
	return &MusicHandler{
		musicService: musicService,
	}
}

// RegisterMusic handles POST /api/v1/music/register
func (h *MusicHandler) RegisterMusic(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Get form fields
	creatorAddress := c.PostForm("creator_address")
	title := c.PostForm("title")
	artist := c.PostForm("artist")
	genre := c.PostForm("genre")
	description := c.PostForm("description")
	durationStr := c.PostForm("duration")

	if creatorAddress == "" || title == "" || artist == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	duration, _ := strconv.Atoi(durationStr)

	// Get audio file
	file, _, err := c.Request.FormFile("audio_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Audio file is required"})
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audio file"})
		return
	}

	// Create request
	req := &services.RegisterMusicRequest{
		CreatorAddress: creatorAddress,
		Title:          title,
		Artist:         artist,
		Genre:          genre,
		Description:    description,
		AudioData:      audioData,
		Duration:       duration,
	}

	// Register music
	resp, err := h.musicService.RegisterMusic(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetMusic handles GET /api/v1/music/:tokenId
func (h *MusicHandler) GetMusic(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	music, err := h.musicService.GetMusic(c.Request.Context(), tokenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Music not found"})
		return
	}

	c.JSON(http.StatusOK, music)
}

// ListMusic handles GET /api/v1/music
func (h *MusicHandler) ListMusic(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	creatorAddress := c.Query("creator")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 100 {
		limit = 100
	}

	musics, total, err := h.musicService.ListMusic(c.Request.Context(), limit, offset, creatorAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   musics,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetMusicAnalytics handles GET /api/v1/music/:tokenId/analytics
func (h *MusicHandler) GetMusicAnalytics(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	analytics, err := h.musicService.GetAnalytics(c.Request.Context(), tokenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Analytics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
