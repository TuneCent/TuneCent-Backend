package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/services"
)

type DistributionHandler struct {
	distributionService *services.DistributionService
}

func NewDistributionHandler(distributionService *services.DistributionService) *DistributionHandler {
	return &DistributionHandler{
		distributionService: distributionService,
	}
}

// SubmitDistribution handles POST /api/v1/distribution/submit
func (h *DistributionHandler) SubmitDistribution(c *gin.Context) {
	var req services.SubmitDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	submission, err := h.distributionService.SubmitDistribution(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Distribution submitted successfully",
		"submission": submission,
	})
}

// GetDistributionStatus handles GET /api/v1/distribution/:tokenId/status
func (h *DistributionHandler) GetDistributionStatus(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	status, err := h.distributionService.GetDistributionStatus(c.Request.Context(), tokenID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetPlatformStatus handles GET /api/v1/distribution/:tokenId/platform/:platform
func (h *DistributionHandler) GetPlatformStatus(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	platform := c.Param("platform")

	platformStatus, err := h.distributionService.GetPlatformStatus(c.Request.Context(), tokenID, platform)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, platformStatus)
}

// UpdatePlatformStatus handles PUT /api/v1/distribution/:tokenId/platform/:platform
func (h *DistributionHandler) UpdatePlatformStatus(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	platform := c.Param("platform")

	var req struct {
		Status      string `json:"status" binding:"required"`
		ExternalID  string `json:"external_id"`
		ExternalURL string `json:"external_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.distributionService.UpdatePlatformStatus(c.Request.Context(), tokenID, platform, req.Status, req.ExternalID, req.ExternalURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Platform status updated successfully",
	})
}

// ListDistributions handles GET /api/v1/distribution/list
func (h *DistributionHandler) ListDistributions(c *gin.Context) {
	userAddress := c.Query("user_address")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 100 {
		limit = 100
	}

	submissions, total, err := h.distributionService.ListDistributions(c.Request.Context(), userAddress, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   submissions,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
