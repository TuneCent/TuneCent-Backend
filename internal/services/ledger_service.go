package services

import (
	"context"
	"fmt"
	"time"

	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
)

type LedgerService struct {
	db *database.DB
}

func NewLedgerService(db *database.DB) *LedgerService {
	return &LedgerService{db: db}
}

type SplitHistoryResponse struct {
	TokenID       uint64                        `json:"token_id"`
	TotalSplits   int64                         `json:"total_splits"`
	TotalAmount   string                        `json:"total_amount"`
	SplitRecords  []SplitRecordDetail           `json:"split_records"`
}

type SplitRecordDetail struct {
	ID             uint                          `json:"id"`
	PaymentID      uint                          `json:"payment_id"`
	TotalAmount    string                        `json:"total_amount"`
	SplitCount     int                           `json:"split_count"`
	TxHash         string                        `json:"tx_hash"`
	BlockNumber    uint64                        `json:"block_number"`
	BlockTimestamp time.Time                     `json:"block_timestamp"`
	Distributions  []models.RoyaltyDistribution  `json:"distributions"`
	CreatedAt      time.Time                     `json:"created_at"`
}

type ContributorBreakdown struct {
	TokenID        uint64                 `json:"token_id"`
	TotalPayments  int64                  `json:"total_payments"`
	Contributors   []ContributorSummary   `json:"contributors"`
}

type ContributorSummary struct {
	Beneficiary    string    `json:"beneficiary"`
	TotalAmount    string    `json:"total_amount"`
	PaymentCount   int64     `json:"payment_count"`
	LastPayment    time.Time `json:"last_payment"`
}

func (s *LedgerService) GetSplitHistory(ctx context.Context, tokenID uint64, limit, offset int) (*SplitHistoryResponse, error) {
	var splitRecords []models.SplitRecord
	var total int64

	// Get split records
	query := s.db.Model(&models.SplitRecord{}).Where("token_id = ?", tokenID)
	query.Count(&total)
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&splitRecords)

	// Calculate total amount
	var totalAmountSum struct {
		Total string
	}
	s.db.Model(&models.SplitRecord{}).
		Select("COALESCE(SUM(CAST(total_amount AS DECIMAL(30,0))), 0) as total").
		Where("token_id = ?", tokenID).
		Scan(&totalAmountSum)

	// Build detailed records with distributions
	details := make([]SplitRecordDetail, len(splitRecords))
	for i, record := range splitRecords {
		var distributions []models.RoyaltyDistribution
		s.db.Where("payment_id = ?", record.PaymentID).Find(&distributions)

		details[i] = SplitRecordDetail{
			ID:             record.ID,
			PaymentID:      record.PaymentID,
			TotalAmount:    record.TotalAmount,
			SplitCount:     record.SplitCount,
			TxHash:         record.TxHash,
			BlockNumber:    record.BlockNumber,
			BlockTimestamp: record.BlockTimestamp,
			Distributions:  distributions,
			CreatedAt:      record.CreatedAt,
		}
	}

	return &SplitHistoryResponse{
		TokenID:      tokenID,
		TotalSplits:  total,
		TotalAmount:  totalAmountSum.Total,
		SplitRecords: details,
	}, nil
}

func (s *LedgerService) GetContributorBreakdown(ctx context.Context, tokenID uint64) (*ContributorBreakdown, error) {
	type ContributorData struct {
		Beneficiary  string
		TotalAmount  string
		PaymentCount int64
		LastPayment  time.Time
	}

	var contributors []ContributorData

	s.db.Table("royalty_distributions").
		Select(`beneficiary,
			COALESCE(SUM(CAST(amount AS DECIMAL(30,0))), 0) as total_amount,
			COUNT(*) as payment_count,
			MAX(distributed_at) as last_payment`).
		Where("token_id = ?", tokenID).
		Group("beneficiary").
		Order("total_amount DESC").
		Scan(&contributors)

	// Convert to response format
	summaries := make([]ContributorSummary, len(contributors))
	for i, c := range contributors {
		summaries[i] = ContributorSummary{
			Beneficiary:  c.Beneficiary,
			TotalAmount:  c.TotalAmount,
			PaymentCount: c.PaymentCount,
			LastPayment:  c.LastPayment,
		}
	}

	return &ContributorBreakdown{
		TokenID:       tokenID,
		TotalPayments: int64(len(contributors)),
		Contributors:  summaries,
	}, nil
}

func (s *LedgerService) CreateSplitRecord(ctx context.Context, tokenID uint64, paymentID uint, totalAmount string, splitCount int, txHash string, blockNumber uint64) (*models.SplitRecord, error) {
	splitRecord := &models.SplitRecord{
		TokenID:        tokenID,
		PaymentID:      paymentID,
		TotalAmount:    totalAmount,
		SplitCount:     splitCount,
		TxHash:         txHash,
		BlockNumber:    blockNumber,
		BlockTimestamp: time.Now(),
	}

	if err := s.db.Create(splitRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create split record: %w", err)
	}

	return splitRecord, nil
}

func (s *LedgerService) GetSplitRecordByTxHash(ctx context.Context, txHash string) (*SplitRecordDetail, error) {
	var splitRecord models.SplitRecord
	if err := s.db.Where("tx_hash = ?", txHash).First(&splitRecord).Error; err != nil {
		return nil, fmt.Errorf("split record not found: %w", err)
	}

	var distributions []models.RoyaltyDistribution
	s.db.Where("payment_id = ?", splitRecord.PaymentID).Find(&distributions)

	return &SplitRecordDetail{
		ID:             splitRecord.ID,
		PaymentID:      splitRecord.PaymentID,
		TotalAmount:    splitRecord.TotalAmount,
		SplitCount:     splitRecord.SplitCount,
		TxHash:         splitRecord.TxHash,
		BlockNumber:    splitRecord.BlockNumber,
		BlockTimestamp: splitRecord.BlockTimestamp,
		Distributions:  distributions,
		CreatedAt:      splitRecord.CreatedAt,
	}, nil
}

func (s *LedgerService) GetUserLedger(ctx context.Context, userAddress string, limit, offset int) ([]models.RoyaltyDistribution, int64, error) {
	var distributions []models.RoyaltyDistribution
	var total int64

	query := s.db.Model(&models.RoyaltyDistribution{}).Where("beneficiary = ?", userAddress)
	query.Count(&total)
	query.Order("distributed_at DESC").Limit(limit).Offset(offset).Find(&distributions)

	return distributions, total, nil
}
