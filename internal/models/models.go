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
	// PoC additions for dashboard/leaderboard
	DisplayName     string         `json:"display_name,omitempty"`
	Bio             string         `gorm:"type:text" json:"bio,omitempty"`
	AvatarURL       string         `json:"avatar_url,omitempty"`
	Tier            string         `gorm:"default:'Registered Creator'" json:"tier"`
	LeaderboardRank uint           `gorm:"default:0" json:"leaderboard_rank"`
	TotalEarnings   string         `gorm:"default:'0'" json:"total_earnings"` // Wei as string
	TotalWorks      uint           `gorm:"default:0" json:"total_works"`
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
	IPFSCID           string         `gorm:"column:ipfs_cid;not null" json:"ipfs_cid"`
	FingerprintHash   string         `gorm:"uniqueIndex;not null" json:"fingerprint_hash"`
	AudioFileURL      string         `json:"audio_file_url,omitempty"`
	CoverImageURL     string         `json:"cover_image_url,omitempty"`
	Duration          int            `json:"duration,omitempty"` // in seconds
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	TxHash            string         `json:"tx_hash,omitempty"`
	RegisteredAt      time.Time      `json:"registered_at"`
	// PoC additions for analytics and trending
	PlayCount         uint64         `gorm:"default:0" json:"play_count"`
	ViewCount         uint64         `gorm:"default:0" json:"view_count"`
	ListenerCount     uint64         `gorm:"default:0" json:"listener_count"`
	ViralScore        float64        `gorm:"type:decimal(5,2);default:0" json:"viral_score"`
	TrendingRank      int            `gorm:"default:0" json:"trending_rank"` // 0 = not trending
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
	// PoC additions for pool stats and trending
	RiskScore         uint8          `gorm:"default:50" json:"risk_score"` // 0-100, lower = safer
	IsTrending        bool           `gorm:"default:false" json:"is_trending"`
	EstimatedROI      float64        `gorm:"type:decimal(10,2);default:150" json:"estimated_roi"`
	ContributorCount  uint           `gorm:"default:0" json:"contributor_count"`
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
	// PoC additions for platform-specific stats
	SpotifyPlays      uint64    `gorm:"default:0" json:"spotify_plays"`
	SpotifyGrowth     float64   `gorm:"type:decimal(10,2);default:0" json:"spotify_growth"`
	TikTokViews       uint64    `gorm:"default:0" json:"tiktok_views"`
	TikTokGrowth      float64   `gorm:"type:decimal(10,2);default:0" json:"tiktok_growth"`
	AppleMusicPlays   uint64    `gorm:"default:0" json:"apple_music_plays"`
	AppleMusicGrowth  float64   `gorm:"type:decimal(10,2);default:0" json:"apple_music_growth"`
	EstimatedReach    uint64    `gorm:"default:0" json:"estimated_reach"`
	WeeklyGrowth      float64   `gorm:"type:decimal(10,2);default:0" json:"weekly_growth"`
	LastUpdated       time.Time `json:"last_updated"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Transaction represents a wallet transaction history entry
type Transaction struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserAddress string    `gorm:"not null;index" json:"user_address"`
	Type        string    `gorm:"not null" json:"type"` // royalty, invest, withdraw, etc.
	Amount      string    `json:"amount,omitempty"` // Wei as string
	TxHash      string    `gorm:"index" json:"tx_hash,omitempty"`
	Status      string    `gorm:"default:'confirmed'" json:"status"` // pending, confirmed, failed
	Description string    `gorm:"type:text" json:"description,omitempty"`
	RelatedID   uint64    `json:"related_id,omitempty"` // token_id, campaign_id, etc.
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Activity represents a user activity feed entry
type Activity struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserAddress string    `gorm:"not null;index" json:"user_address"`
	Type        string    `gorm:"not null" json:"type"` // music_registered, royalty_received, pool_invested, etc.
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	RelatedID   uint64    `json:"related_id,omitempty"` // token_id, campaign_id, etc.
	TxHash      string    `json:"tx_hash,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// DistributionSubmission tracks music distribution to external platforms
type DistributionSubmission struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	TokenID      uint64         `gorm:"not null;index" json:"token_id"`
	UserAddress  string         `gorm:"not null;index" json:"user_address"`
	Platforms    string         `gorm:"type:text" json:"platforms"` // JSON array of platforms
	Status       string         `gorm:"default:'pending'" json:"status"` // pending, processing, distributed, failed
	SubmittedAt  time.Time      `json:"submitted_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// PlatformDistribution tracks distribution status per platform
type PlatformDistribution struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	TokenID       uint64         `gorm:"not null;index" json:"token_id"`
	Platform      string         `gorm:"not null;index" json:"platform"` // spotify, tiktok, apple_music, youtube_music
	Status        string         `gorm:"default:'pending'" json:"status"` // pending, live, failed, removed
	ExternalID    string         `json:"external_id,omitempty"` // Platform's track ID
	ExternalURL   string         `json:"external_url,omitempty"`
	DistributedAt *time.Time     `json:"distributed_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// Notification represents user notifications
type Notification struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	UserAddress string    `gorm:"not null;index" json:"user_address"`
	Type        string    `gorm:"not null" json:"type"` // payment, contribution, milestone, alert
	Title       string    `gorm:"not null" json:"title"`
	Message     string    `gorm:"type:text" json:"message"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	RelatedID   uint64    `json:"related_id,omitempty"` // token_id, campaign_id, etc.
	TxHash      string    `json:"tx_hash,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NotificationPreference stores user notification preferences
type NotificationPreference struct {
	ID                   uint   `gorm:"primarykey" json:"id"`
	UserAddress          string `gorm:"uniqueIndex;not null" json:"user_address"`
	EmailNotifications   bool   `gorm:"default:true" json:"email_notifications"`
	RoyaltyAlerts        bool   `gorm:"default:true" json:"royalty_alerts"`
	ContributionAlerts   bool   `gorm:"default:true" json:"contribution_alerts"`
	MilestoneAlerts      bool   `gorm:"default:true" json:"milestone_alerts"`
	MarketingEmails      bool   `gorm:"default:false" json:"marketing_emails"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// SplitRecord tracks royalty split records for audit
type SplitRecord struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	TokenID        uint64    `gorm:"not null;index" json:"token_id"`
	PaymentID      uint      `gorm:"not null;index" json:"payment_id"`
	TotalAmount    string    `gorm:"not null" json:"total_amount"` // Wei as string
	SplitCount     int       `gorm:"not null" json:"split_count"`
	TxHash         string    `gorm:"index" json:"tx_hash"`
	BlockNumber    uint64    `json:"block_number,omitempty"`
	BlockTimestamp time.Time `json:"block_timestamp"`
	CreatedAt      time.Time `json:"created_at"`
}

// ReinvestmentSuggestion stores reinvestment opportunities
type ReinvestmentSuggestion struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserAddress     string    `gorm:"not null;index" json:"user_address"`
	AvailableFunds  string    `gorm:"not null" json:"available_funds"` // Wei as string
	SuggestedPools  string    `gorm:"type:text" json:"suggested_pools"` // JSON array of campaign IDs
	ExpectedROI     float64   `gorm:"type:decimal(10,2)" json:"expected_roi"`
	Reasoning       string    `gorm:"type:text" json:"reasoning,omitempty"`
	IsActioned      bool      `gorm:"default:false" json:"is_actioned"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ReinvestmentHistory tracks user reinvestment actions
type ReinvestmentHistory struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	UserAddress     string    `gorm:"not null;index" json:"user_address"`
	FromSource      string    `gorm:"not null" json:"from_source"` // royalty, withdrawal, etc.
	ToCampaignID    uint64    `gorm:"not null;index" json:"to_campaign_id"`
	Amount          string    `gorm:"not null" json:"amount"` // Wei as string
	TxHash          string    `json:"tx_hash,omitempty"`
	SuggestionID    *uint     `json:"suggestion_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}
