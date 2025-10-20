package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/services"
)

type LedgerHandler struct {
	ledgerService *services.LedgerService
}

func NewLedgerHandler(ledgerService *services.LedgerService) *LedgerHandler {
	return &LedgerHandler{
		ledgerService: ledgerService,
	}
}

// GetSplitHistory handles GET /api/v1/ledger/:tokenId/splits
func (h *LedgerHandler) GetSplitHistory(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 100 {
		limit = 100
	}

	history, err := h.ledgerService.GetSplitHistory(c.Request.Context(), tokenID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetContributorBreakdown handles GET /api/v1/ledger/:tokenId/contributors
func (h *LedgerHandler) GetContributorBreakdown(c *gin.Context) {
	tokenIDStr := c.Param("tokenId")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	breakdown, err := h.ledgerService.GetContributorBreakdown(c.Request.Context(), tokenID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, breakdown)
}

// GetSplitByTxHash handles GET /api/v1/ledger/audit/:txHash
func (h *LedgerHandler) GetSplitByTxHash(c *gin.Context) {
	txHash := c.Param("txHash")

	splitRecord, err := h.ledgerService.GetSplitRecordByTxHash(c.Request.Context(), txHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, splitRecord)
}

// GetUserLedger handles GET /api/v1/ledger/user/:address
func (h *LedgerHandler) GetUserLedger(c *gin.Context) {
	userAddress := c.Param("address")

	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit > 100 {
		limit = 100
	}

	distributions, total, err := h.ledgerService.GetUserLedger(c.Request.Context(), userAddress, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_address": userAddress,
		"data":         distributions,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}
