package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

// WalletHandler handles wallet and transaction endpoints
type WalletHandler struct {
	db *database.DB
}

func NewWalletHandler(db *database.DB) *WalletHandler {
	return &WalletHandler{db: db}
}

// GetTransactions returns transaction history for a wallet
// GET /api/v1/wallet/:address/transactions?limit=20&offset=0&type=royalty
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// Query parameters
	limit := c.DefaultQuery("limit", "20")
	offset := c.DefaultQuery("offset", "0")
	txType := c.Query("type") // Optional: filter by type

	var transactions []models.Transaction
	query := h.db.Where("user_address = ?", address).Order("created_at DESC")

	if txType != "" {
		query = query.Where("type = ?", txType)
	}

	query.Limit(atoi(limit)).Offset(atoi(offset)).Find(&transactions)

	// Get total count
	var total int64
	countQuery := h.db.Model(&models.Transaction{}).Where("user_address = ?", address)
	if txType != "" {
		countQuery = countQuery.Where("type = ?", txType)
	}
	countQuery.Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"total":        total,
		"limit":        atoi(limit),
		"offset":       atoi(offset),
	})
}

// GetBalance returns wallet balance (ETH + USD conversion)
// GET /api/v1/wallet/:address/balance
func (h *WalletHandler) GetBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// Calculate total earnings from royalty distributions
	var totalEarnings struct {
		Total string
	}
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ?", address).
		Scan(&totalEarnings)

	// Calculate total invested in campaigns
	var totalInvested struct {
		Total string
	}
	h.db.Model(&models.Contribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Where("contributor_address = ?", address).
		Scan(&totalInvested)

	// Mock ETH price for PoC (in production, fetch from oracle/API)
	ethPriceUSD := 2500.0

	// Calculate balance in ETH (earnings - invested)
	// For PoC, simplified calculation
	balanceWei := totalEarnings.Total
	if balanceWei == "" {
		balanceWei = "0"
	}

	// Convert Wei to ETH (mock calculation for display)
	// In production, use proper big number math
	balanceETH := 0.32 // Mock value for PoC
	balanceUSD := balanceETH * ethPriceUSD

	c.JSON(http.StatusOK, gin.H{
		"address":         address,
		"balance_wei":     balanceWei,
		"balance_eth":     balanceETH,
		"balance_usd":     balanceUSD,
		"total_earnings":  totalEarnings.Total,
		"total_invested":  totalInvested.Total,
		"eth_price_usd":   ethPriceUSD,
	})
}

// SearchTransactions searches transactions by description or tx hash
// GET /api/v1/wallet/:address/search?q=royalty&limit=20
func (h *WalletHandler) SearchTransactions(c *gin.Context) {
	address := c.Param("address")
	query := c.Query("q")

	if address == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address and q parameters are required"})
		return
	}

	limit := c.DefaultQuery("limit", "20")

	var transactions []models.Transaction
	h.db.Where("user_address = ? AND (description LIKE ? OR tx_hash LIKE ? OR type LIKE ?)",
		address, "%"+query+"%", "%"+query+"%", "%"+query+"%").
		Order("created_at DESC").
		Limit(atoi(limit)).
		Find(&transactions)

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
		"query":        query,
		"total":        len(transactions),
	})
}

// GetSavings returns total savings and estimated savings
// GET /api/v1/wallet/:address/savings
func (h *WalletHandler) GetSavings(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	// For PoC, calculate savings based on staking fee discount
	// In production, track actual savings from fee reductions

	// Get total royalties received
	var totalRoyalties struct {
		Total string
	}
	h.db.Model(&models.RoyaltyDistribution{}).
		Select("COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total").
		Joins("JOIN music_metadata ON royalty_distributions.token_id = music_metadata.token_id").
		Where("music_metadata.creator_address = ?", address).
		Scan(&totalRoyalties)

	// Mock: Assume 10% platform fee normally, but user gets 10% discount from staking
	// So they save 1% of total royalties
	// For PoC, use simplified calculation
	totalSavedUSD := 201.56 // Mock value from Figma design
	estimatedSavingsUSD := 351.67 // Mock estimated future savings

	c.JSON(http.StatusOK, gin.H{
		"address":            address,
		"total_saved":        totalSavedUSD,
		"estimated_savings":  estimatedSavingsUSD,
		"savings_source":     "Staking fee discount (10%)",
	})
}

// GetTransactionAudit returns detailed audit information for a transaction
// GET /api/v1/audit/transaction/:txHash
func (h *WalletHandler) GetTransactionAudit(c *gin.Context) {
	txHash := c.Param("txHash")
	if txHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "txHash parameter is required"})
		return
	}

	// Find transaction in database
	var transaction models.Transaction
	if err := h.db.Where("tx_hash = ?", txHash).First(&transaction).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	// Get related royalty distribution if applicable
	var royaltyDist models.RoyaltyDistribution
	var hasRoyalty bool
	if transaction.Type == "royalty" {
		if err := h.db.Where("tx_hash = ?", txHash).First(&royaltyDist).Error; err == nil {
			hasRoyalty = true
		}
	}

	// Build audit response
	auditData := gin.H{
		"tx_hash":       txHash,
		"type":          transaction.Type,
		"status":        transaction.Status,
		"amount":        transaction.Amount,
		"from_address":  transaction.UserAddress,
		"description":   transaction.Description,
		"timestamp":     transaction.CreatedAt,
		"block_number":  nil, // Would be fetched from blockchain in production
		"gas_used":      nil, // Would be fetched from blockchain in production
		"explorer_url":  "https://etherscan.io/tx/" + txHash, // Mock explorer URL
	}

	if hasRoyalty {
		auditData["royalty_details"] = gin.H{
			"token_id":    royaltyDist.TokenID,
			"beneficiary": royaltyDist.Beneficiary,
			"amount":      royaltyDist.Amount,
			"distributed_at": royaltyDist.DistributedAt,
		}
	}

	c.JSON(http.StatusOK, auditData)
}

// VerifyTransaction verifies a transaction on-chain
// GET /api/v1/audit/verify/:txHash
func (h *WalletHandler) VerifyTransaction(c *gin.Context) {
	txHash := c.Param("txHash")
	if txHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "txHash parameter is required"})
		return
	}

	// In production, this would call blockchain RPC to verify the transaction
	// For PoC, return mock verification data
	c.JSON(http.StatusOK, gin.H{
		"tx_hash":       txHash,
		"verified":      true,
		"confirmations": 12,
		"block_number":  18234567,
		"timestamp":     "2025-10-20T10:30:45Z",
		"status":        "confirmed",
		"message":       "Transaction verified on-chain",
	})
}

// GetBlockDetails returns block information
// GET /api/v1/audit/block/:blockNumber
func (h *WalletHandler) GetBlockDetails(c *gin.Context) {
	blockNumberStr := c.Param("blockNumber")
	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid block number"})
		return
	}

	// In production, fetch from blockchain
	// For PoC, return mock data
	c.JSON(http.StatusOK, gin.H{
		"block_number":  blockNumber,
		"timestamp":     "2025-10-20T10:30:45Z",
		"miner":         "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		"gas_used":      21000,
		"gas_limit":     30000000,
		"transactions":  156,
		"explorer_url":  "https://etherscan.io/block/" + blockNumberStr,
	})
}

// Helper function to convert string to int
func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
