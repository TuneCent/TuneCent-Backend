package services

import (
	"context"
	"fmt"
	"time"

	"github.com/tunecent/backend/internal/blockchain"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/models"
	"github.com/tunecent/backend/pkg/fingerprint"
	"github.com/tunecent/backend/pkg/ipfs"
)

type MusicService struct {
	db          *database.DB
	ipfs        *ipfs.Service
	fingerprint *fingerprint.Service
	blockchain  *blockchain.Service
}

func NewMusicService(db *database.DB, ipfsService *ipfs.Service, fpService *fingerprint.Service, bcService *blockchain.Service) *MusicService {
	return &MusicService{
		db:          db,
		ipfs:        ipfsService,
		fingerprint: fpService,
		blockchain:  bcService,
	}
}

type RegisterMusicRequest struct {
	CreatorAddress string `json:"creator_address" binding:"required"`
	Title          string `json:"title" binding:"required"`
	Artist         string `json:"artist" binding:"required"`
	Genre          string `json:"genre"`
	Description    string `json:"description"`
	AudioData      []byte `json:"-"` // Binary audio data
	Duration       int    `json:"duration"`
}

type RegisterMusicResponse struct {
	TokenID         uint64    `json:"token_id"`
	IPFSCID         string    `json:"ipfs_cid"`
	FingerprintHash string    `json:"fingerprint_hash"`
	TxHash          string    `json:"tx_hash"`
	Message         string    `json:"message"`
	RegisteredAt    time.Time `json:"registered_at"`
}

func (s *MusicService) RegisterMusic(ctx context.Context, req *RegisterMusicRequest) (*RegisterMusicResponse, error) {
	// Step 1: Generate fingerprint
	fingerprintHash, err := s.fingerprint.Generate(req.AudioData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate fingerprint: %w", err)
	}

	// Step 2: Check if fingerprint already exists
	var existingMusic models.MusicMetadata
	if err := s.db.Where("fingerprint_hash = ?", fingerprintHash).First(&existingMusic).Error; err == nil {
		return nil, fmt.Errorf("music already registered with token ID: %d", existingMusic.TokenID)
	}

	// Step 3: Upload metadata to IPFS
	metadata := ipfs.MusicMetadata{
		Title:           req.Title,
		Artist:          req.Artist,
		Genre:           req.Genre,
		Description:     req.Description,
		Duration:        req.Duration,
		FingerprintHash: fingerprintHash,
		Creator:         req.CreatorAddress,
		Timestamp:       time.Now().Unix(),
	}

	ipfsCID, err := s.ipfs.UploadJSON(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS: %w", err)
	}

	// Step 4: In production, call smart contract to register music
	// For PoC, we simulate with a mock token ID and tx hash
	tokenID := uint64(time.Now().Unix()) // Mock token ID
	txHash := fmt.Sprintf("0x%064x", time.Now().UnixNano()) // Mock tx hash

	// Note: Real implementation would be:
	// tx, err := s.blockchain.RegisterMusic(ctx, ipfsCID, fingerprintHash, req.Title, req.Artist)
	// tokenID := tx.TokenID
	// txHash := tx.Hash

	// Step 5: Save to database
	musicMetadata := &models.MusicMetadata{
		TokenID:         tokenID,
		CreatorAddress:  req.CreatorAddress,
		Title:           req.Title,
		Artist:          req.Artist,
		Genre:           req.Genre,
		Description:     req.Description,
		IPFSCID:         ipfsCID,
		FingerprintHash: fingerprintHash,
		Duration:        req.Duration,
		IsActive:        true,
		TxHash:          txHash,
		RegisteredAt:    time.Now(),
	}

	if err := s.db.Create(musicMetadata).Error; err != nil {
		return nil, fmt.Errorf("failed to save to database: %w", err)
	}

	// Step 6: Initialize analytics
	analytics := &models.Analytics{
		TokenID:        tokenID,
		TotalViews:     0,
		TotalEmbeds:    0,
		TotalUsages:    0,
		TotalRoyalties: "0",
		LastUpdated:    time.Now(),
	}
	s.db.Create(analytics)

	return &RegisterMusicResponse{
		TokenID:         tokenID,
		IPFSCID:         ipfsCID,
		FingerprintHash: fingerprintHash,
		TxHash:          txHash,
		Message:         "Music registered successfully",
		RegisteredAt:    musicMetadata.RegisteredAt,
	}, nil
}

func (s *MusicService) GetMusic(ctx context.Context, tokenID uint64) (*models.MusicMetadata, error) {
	var music models.MusicMetadata
	if err := s.db.Where("token_id = ?", tokenID).First(&music).Error; err != nil {
		return nil, fmt.Errorf("music not found: %w", err)
	}
	return &music, nil
}

func (s *MusicService) ListMusic(ctx context.Context, limit, offset int, creatorAddress string) ([]*models.MusicMetadata, int64, error) {
	var musics []*models.MusicMetadata
	var total int64

	query := s.db.Model(&models.MusicMetadata{})

	if creatorAddress != "" {
		query = query.Where("creator_address = ?", creatorAddress)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Order("registered_at DESC").Limit(limit).Offset(offset).Find(&musics).Error; err != nil {
		return nil, 0, err
	}

	return musics, total, nil
}

func (s *MusicService) GetAnalytics(ctx context.Context, tokenID uint64) (*models.Analytics, error) {
	var analytics models.Analytics
	if err := s.db.Where("token_id = ?", tokenID).First(&analytics).Error; err != nil {
		return nil, fmt.Errorf("analytics not found: %w", err)
	}
	return &analytics, nil
}

func (s *MusicService) VerifyFingerprint(ctx context.Context, fingerprintHash string) (*models.MusicMetadata, error) {
	var music models.MusicMetadata
	if err := s.db.Where("fingerprint_hash = ? AND is_active = ?", fingerprintHash, true).First(&music).Error; err != nil {
		return nil, fmt.Errorf("fingerprint not found or inactive: %w", err)
	}
	return &music, nil
}
