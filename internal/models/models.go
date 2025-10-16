package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a platform user (creator or contributor)
type User struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	WalletAddress   string         `gorm:"uniqueIndex;not null" json:"wallet_address"`
	Username        string         `gorm:"unique" json:"username,omitempty"`
	Email           string         `gorm:"unique" json:"email,omitempty"`
	Role            string         `gorm:"type:enum('creator','contributor','both');default:'contributor'" json:"role"`
	IsVerified      bool           `gorm:"default:false" json:"is_verified"`
	ReputationScore uint           `json:"reputation_score"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// MusicMetadata stores off-chain music metadata
type MusicMetadata struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	TokenID           uint64         `gorm:"uniqueIndex;not null" json:"token_id"`
	CreatorAddress    string         `gorm:"not null;index" json:"creator_address"`
	Title             string         `gorm:"not null" json:"title"`
	Artist            string         `gorm:"not null" json:"artist"`
	Genre             string         `json:"genre,omitempty"`
	Description       string         `gorm:"type:text" json:"description,omitempty"`
	IPFSCID           string         `gorm:"not null" json:"ipfs_cid"`
	FingerprintHash   string         `gorm:"uniqueIndex;not null" json:"fingerprint_hash"`
	AudioFileURL      string         `json:"audio_file_url,omitempty"`
	CoverImageURL     string         `json:"cover_image_url,omitempty"`
	Duration          int            `json:"duration,omitempty"` // in seconds
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	TxHash            string         `json:"tx_hash,omitempty"`
	RegisteredAt      time.Time      `json:"registered_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// Campaign represents a crowdfunding campaign
type Campaign struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	CampaignID        uint64         `gorm:"uniqueIndex;not null" json:"campaign_id"` // On-chain campaign ID
	TokenID           uint64         `gorm:"not null;index" json:"token_id"`
	CreatorAddress    string         `gorm:"not null;index" json:"creator_address"`
	GoalAmount        string         `gorm:"not null" json:"goal_amount"` // Wei as string
	RaisedAmount      string         `gorm:"default:'0'" json:"raised_amount"`
	RoyaltyPercentage uint16         `json:"royalty_percentage"` // Basis points
	Deadline          time.Time      `json:"deadline"`
	LockupPeriod      int            `json:"lockup_period"` // in days
	Status            string         `gorm:"type:enum('active','successful','failed','cancelled');default:'active'" json:"status"`
	FundsWithdrawn    bool           `gorm:"default:false" json:"funds_withdrawn"`
	TxHash            string         `json:"tx_hash,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// Contribution represents a crowdfunding contribution
type Contribution struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	CampaignID        uint64         `gorm:"not null;index" json:"campaign_id"`
	ContributorAddress string        `gorm:"not null;index" json:"contributor_address"`
	Amount            string         `gorm:"not null" json:"amount"` // Wei as string
	SharePercentage   float64        `json:"share_percentage"`
	TxHash            string         `json:"tx_hash,omitempty"`
	ContributedAt     time.Time      `json:"contributed_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// RoyaltyPayment tracks royalty payments
type RoyaltyPayment struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	TokenID         uint64    `gorm:"not null;index" json:"token_id"`
	From            string    `gorm:"not null" json:"from"`
	Amount          string    `gorm:"not null" json:"amount"` // Wei as string
	Platform        string    `gorm:"not null" json:"platform"`
	UsageType       string    `json:"usage_type,omitempty"`
	TxHash          string    `json:"tx_hash"`
	IsDistributed   bool      `gorm:"default:false" json:"is_distributed"`
	DistributedAt   *time.Time `json:"distributed_at,omitempty"`
	PaidAt          time.Time `json:"paid_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// RoyaltyDistribution tracks individual distributions
type RoyaltyDistribution struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	PaymentID     uint      `gorm:"not null;index" json:"payment_id"`
	TokenID       uint64    `gorm:"not null;index" json:"token_id"`
	Beneficiary   string    `gorm:"not null;index" json:"beneficiary"`
	Amount        string    `gorm:"not null" json:"amount"`
	TxHash        string    `json:"tx_hash"`
	DistributedAt time.Time `json:"distributed_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// UsageDetection stores detected music usage events (mock for PoC)
type UsageDetection struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	TokenID      uint64    `gorm:"not null;index" json:"token_id"`
	Platform     string    `gorm:"not null" json:"platform"`
	ContentID    string    `json:"content_id,omitempty"` // e.g., TikTok video ID
	ContentURL   string    `json:"content_url,omitempty"`
	DetectedAt   time.Time `json:"detected_at"`
	PaymentSent  bool      `gorm:"default:false" json:"payment_sent"`
	PaymentTxHash string   `json:"payment_tx_hash,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Analytics stores aggregated analytics data
type Analytics struct {
	ID                uint      `gorm:"primarykey" json:"id"`
	TokenID           uint64    `gorm:"uniqueIndex;not null" json:"token_id"`
	TotalViews        uint64    `gorm:"default:0" json:"total_views"`
	TotalEmbeds       uint64    `gorm:"default:0" json:"total_embeds"`
	TotalUsages       uint64    `gorm:"default:0" json:"total_usages"`
	TotalRoyalties    string    `gorm:"default:'0'" json:"total_royalties"` // Wei as string
	LastUpdated       time.Time `json:"last_updated"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
