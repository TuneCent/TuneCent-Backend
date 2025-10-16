package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tunecent/backend/internal/config"
)

type Client struct {
	client                    *ethclient.Client
	chainID                   *big.Int
	musicRegistryAddress      common.Address
	royaltyDistributorAddress common.Address
	crowdfundingPoolAddress   common.Address
	reputationScoreAddress    common.Address
}

func NewClient(cfg *config.Config) (*Client, error) {
	client, err := ethclient.Dial(cfg.Blockchain.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}

	// Verify connection
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if chainID.Int64() != cfg.Blockchain.ChainID {
		return nil, fmt.Errorf("chain ID mismatch: expected %d, got %d", cfg.Blockchain.ChainID, chainID.Int64())
	}

	return &Client{
		client:                    client,
		chainID:                   chainID,
		musicRegistryAddress:      common.HexToAddress(cfg.Blockchain.MusicRegistryAddress),
		royaltyDistributorAddress: common.HexToAddress(cfg.Blockchain.RoyaltyDistributorAddress),
		crowdfundingPoolAddress:   common.HexToAddress(cfg.Blockchain.CrowdfundingPoolAddress),
		reputationScoreAddress:    common.HexToAddress(cfg.Blockchain.ReputationScoreAddress),
	}, nil
}

func (c *Client) GetClient() *ethclient.Client {
	return c.client
}

func (c *Client) ChainID() *big.Int {
	return c.chainID
}

func (c *Client) MusicRegistryAddress() common.Address {
	return c.musicRegistryAddress
}

func (c *Client) RoyaltyDistributorAddress() common.Address {
	return c.royaltyDistributorAddress
}

func (c *Client) CrowdfundingPoolAddress() common.Address {
	return c.crowdfundingPoolAddress
}

func (c *Client) ReputationScoreAddress() common.Address {
	return c.reputationScoreAddress
}

func (c *Client) Close() {
	c.client.Close()
}
