package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

// CampaignHandler handles crowdfunding campaign endpoints
type CampaignHandler struct {
	db *database.DB
}

func NewCampaignHandler(db *database.DB) *CampaignHandler {
	return &CampaignHandler{db: db}
}

func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
	var req struct {
		TokenID           uint64 `json:"token_id" binding:"required"`
		CreatorAddress    string `json:"creator_address" binding:"required"`
		GoalAmount        string `json:"goal_amount" binding:"required"`
		RoyaltyPercentage uint16 `json:"royalty_percentage" binding:"required"`
		DurationDays      int    `json:"duration_days" binding:"required"`
		LockupDays        int    `json:"lockup_days" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock campaign creation - in production, call smart contract
	campaign := &models.Campaign{
		CampaignID:        uint64(1), // Mock
		TokenID:           req.TokenID,
		CreatorAddress:    req.CreatorAddress,
		GoalAmount:        req.GoalAmount,
		RaisedAmount:      "0",
		RoyaltyPercentage: req.RoyaltyPercentage,
		LockupPeriod:      req.LockupDays,
		Status:            "active",
		TxHash:            "0xmock",
	}

	if err := h.db.Create(campaign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create campaign"})
		return
	}

	c.JSON(http.StatusCreated, campaign)
}

func (h *CampaignHandler) GetCampaign(c *gin.Context) {
	campaignID, _ := strconv.ParseUint(c.Param("campaignId"), 10, 64)

	var campaign models.Campaign
	if err := h.db.Where("campaign_id = ?", campaignID).First(&campaign).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Campaign not found"})
		return
	}

	c.JSON(http.StatusOK, campaign)
}

func (h *CampaignHandler) ListCampaigns(c *gin.Context) {
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	query := h.db.Model(&models.Campaign{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var campaigns []models.Campaign
	var total int64

	query.Count(&total)
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&campaigns)

	c.JSON(http.StatusOK, gin.H{
		"data":   campaigns,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *CampaignHandler) Contribute(c *gin.Context) {
	campaignID, _ := strconv.ParseUint(c.Param("campaignId"), 10, 64)

	var req struct {
		ContributorAddress string `json:"contributor_address" binding:"required"`
		Amount             string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contribution := &models.Contribution{
		CampaignID:         campaignID,
		ContributorAddress: req.ContributorAddress,
		Amount:             req.Amount,
		SharePercentage:    0, // Calculate based on total
		TxHash:             "0xmock",
	}

	if err := h.db.Create(contribution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record contribution"})
		return
	}

	c.JSON(http.StatusCreated, contribution)
}

// RoyaltyHandler handles royalty endpoints
type RoyaltyHandler struct {
	db *database.DB
}

func NewRoyaltyHandler(db *database.DB) *RoyaltyHandler {
	return &RoyaltyHandler{db: db}
}

func (h *RoyaltyHandler) GetRoyalties(c *gin.Context) {
	tokenID, _ := strconv.ParseUint(c.Param("tokenId"), 10, 64)

	var payments []models.RoyaltyPayment
	h.db.Where("token_id = ?", tokenID).Order("paid_at DESC").Find(&payments)

	c.JSON(http.StatusOK, gin.H{
		"token_id": tokenID,
		"payments": payments,
	})
}

func (h *RoyaltyHandler) SimulateRoyaltyPayment(c *gin.Context) {
	var req struct {
		TokenID  uint64 `json:"token_id" binding:"required"`
		Platform string `json:"platform" binding:"required"`
		Amount   string `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment := &models.RoyaltyPayment{
		TokenID:       req.TokenID,
		From:          "0xPlatformSimulator",
		Amount:        req.Amount,
		Platform:      req.Platform,
		UsageType:     "simulated",
		TxHash:        "0xmock",
		IsDistributed: false,
	}

	if err := h.db.Create(payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Royalty payment simulated successfully",
		"payment": payment,
	})
}

// UserHandler handles user and reputation endpoints
type UserHandler struct {
	db *database.DB
}

func NewUserHandler(db *database.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	address := c.Param("address")

	var user models.User
	if err := h.db.Where("wallet_address = ?", address).First(&user).Error; err != nil {
		// Create new user if not exists
		user = models.User{
			WalletAddress: address,
			Role:          "contributor",
		}
		h.db.Create(&user)
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetReputation(c *gin.Context) {
	address := c.Param("address")

	// Get user's stats
	var totalWorks int64
	h.db.Model(&models.MusicMetadata{}).Where("creator_address = ?", address).Count(&totalWorks)

	var campaigns []models.Campaign
	h.db.Where("creator_address = ? AND status = ?", address, "successful").Find(&campaigns)

	c.JSON(http.StatusOK, gin.H{
		"address":              address,
		"total_works":          totalWorks,
		"successful_campaigns": len(campaigns),
		"reputation_score":     totalWorks*10 + int64(len(campaigns))*50,
	})
}
