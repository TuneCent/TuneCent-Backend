package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

// MusicMetadata represents on-chain music metadata structure
type MusicMetadata struct {
	IPFSCID         string
	FingerprintHash [32]byte
	Creator         common.Address
	RegisteredAt    *big.Int
	Title           string
	Artist          string
	IsActive        bool
}

// CampaignInfo represents on-chain campaign data
type CampaignInfo struct {
	TokenID           *big.Int
	Creator           common.Address
	GoalAmount        *big.Int
	RaisedAmount      *big.Int
	RoyaltyPercentage uint16
	Deadline          *big.Int
	LockupPeriod      *big.Int
	Status            uint8
	FundsWithdrawn    bool
	CreatedAt         *big.Int
}

// GetMusicMetadata retrieves music metadata from MusicRegistry contract
func (s *Service) GetMusicMetadata(ctx context.Context, tokenID *big.Int) (*MusicMetadata, error) {
	// Note: In production, you would use generated contract bindings via abigen
	// For PoC, we return a simplified implementation

	// This is a placeholder that would call the actual contract method:
	// registry, err := contracts.NewMusicRegistry(s.client.MusicRegistryAddress(), s.client.GetClient())
	// metadata, err := registry.GetMusicMetadata(&bind.CallOpts{Context: ctx}, tokenID)

	return nil, fmt.Errorf("contract bindings not generated - run 'make generate-bindings'")
}

// VerifyFingerprint checks if a fingerprint exists on-chain
func (s *Service) VerifyFingerprint(ctx context.Context, fingerprintHash [32]byte) (bool, *big.Int, common.Address, error) {
	// Placeholder for contract call
	// registry, err := contracts.NewMusicRegistry(s.client.MusicRegistryAddress(), s.client.GetClient())
	// return registry.VerifyFingerprint(&bind.CallOpts{Context: ctx}, fingerprintHash)

	return false, nil, common.Address{}, fmt.Errorf("contract bindings not generated")
}

// GetCampaign retrieves campaign information
func (s *Service) GetCampaign(ctx context.Context, campaignID *big.Int) (*CampaignInfo, error) {
	// Placeholder for contract call
	return nil, fmt.Errorf("contract bindings not generated")
}

// GetPendingRoyalties gets pending royalties for a token
func (s *Service) GetPendingRoyalties(ctx context.Context, tokenID *big.Int) (*big.Int, error) {
	// Placeholder for contract call
	return nil, fmt.Errorf("contract bindings not generated")
}

// GetReputationScore gets creator reputation score
func (s *Service) GetReputationScore(ctx context.Context, creator common.Address) (*big.Int, error) {
	// Placeholder for contract call
	return nil, fmt.Errorf("contract bindings not generated")
}

// WaitForTransaction waits for a transaction to be mined
func (s *Service) WaitForTransaction(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return bind.WaitMined(ctx, s.client.GetClient(), &types.Transaction{})
}

// GetBlockNumber returns the latest block number
func (s *Service) GetBlockNumber(ctx context.Context) (uint64, error) {
	return s.client.GetClient().BlockNumber(ctx)
}

// Helper function to convert string to bytes32
func StringToBytes32(s string) [32]byte {
	var b [32]byte
	copy(b[:], s)
	return b
}

// Helper to check if address is valid
func IsValidAddress(address string) bool {
	return common.IsHexAddress(address)
}

// NOTE: To generate contract bindings, run:
// abigen --sol=../TuneCent-SmartContract/src/MusicRegistry.sol --pkg=contracts --out=internal/blockchain/contracts/MusicRegistry.go
// Repeat for all contracts
