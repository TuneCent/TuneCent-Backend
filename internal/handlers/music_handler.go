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
// @Summary Register new music NFT
// @Description Upload and register a new music NFT with metadata and audio file
// @Tags Music
// @Accept multipart/form-data
// @Produce json
// @Param creator_address formData string true "Creator's wallet address"
// @Param title formData string true "Music title"
// @Param artist formData string true "Artist name"
// @Param genre formData string false "Music genre"
// @Param description formData string false "Music description"
// @Param duration formData integer false "Duration in seconds"
// @Param audio_file formData file true "Audio file"
// @Success 201 {object} map[string]interface{} "Music registered successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /music/register [post]
func (h *MusicHandler) RegisterMusic(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil { // 50 MB limit
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
// @Summary Get music by token ID
// @Description Retrieve music NFT metadata by token ID
// @Tags Music
// @Produce json
// @Param tokenId path integer true "Music Token ID"
// @Success 200 {object} map[string]interface{} "Music metadata"
// @Failure 400 {object} map[string]interface{} "Invalid token ID"
// @Failure 404 {object} map[string]interface{} "Music not found"
// @Router /music/{tokenId} [get]
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
// @Summary List all music NFTs
// @Description Get paginated list of music NFTs with optional filtering
// @Tags Music
// @Produce json
// @Param limit query integer false "Limit (max 100)" default(20)
// @Param offset query integer false "Offset" default(0)
// @Param creator query string false "Filter by creator address"
// @Success 200 {object} map[string]interface{} "List of music"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /music [get]
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
// @Summary Get music analytics
// @Description Retrieve analytics data for a specific music NFT
// @Tags Music
// @Produce json
// @Param tokenId path integer true "Music Token ID"
// @Success 200 {object} map[string]interface{} "Music analytics"
// @Failure 400 {object} map[string]interface{} "Invalid token ID"
// @Failure 404 {object} map[string]interface{} "Analytics not found"
// @Router /music/{tokenId}/analytics [get]
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
