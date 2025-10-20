package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/services"
)

type ReinvestmentHandler struct {
	reinvestmentService *services.ReinvestmentService
}

func NewReinvestmentHandler(reinvestmentService *services.ReinvestmentService) *ReinvestmentHandler {
	return &ReinvestmentHandler{
		reinvestmentService: reinvestmentService,
	}
}

// GetSuggestions handles GET /api/v1/reinvest/suggestions
func (h *ReinvestmentHandler) GetSuggestions(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	suggestions, err := h.reinvestmentService.GetSuggestions(c.Request.Context(), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, suggestions)
}

// QuickReinvest handles POST /api/v1/reinvest/quick
func (h *ReinvestmentHandler) QuickReinvest(c *gin.Context) {
	var req services.QuickReinvestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history, err := h.reinvestmentService.QuickReinvest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Reinvestment successful",
		"history": history,
	})
}

// GetHistory handles GET /api/v1/reinvest/history
func (h *ReinvestmentHandler) GetHistory(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 100 {
		limit = 100
	}

	history, total, err := h.reinvestmentService.GetReinvestmentHistory(c.Request.Context(), userAddress, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   history,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetStats handles GET /api/v1/reinvest/stats
func (h *ReinvestmentHandler) GetStats(c *gin.Context) {
	userAddress := c.Query("user_address")
	if userAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_address is required"})
		return
	}

	stats, err := h.reinvestmentService.GetReinvestmentStats(c.Request.Context(), userAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
