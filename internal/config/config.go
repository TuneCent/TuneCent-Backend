package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Blockchain BlockchainConfig
	IPFS       IPFSConfig
	JWT        JWTConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type BlockchainConfig struct {
	RPCURL                    string
	ChainID                   int64
	MusicRegistryAddress      string
	RoyaltyDistributorAddress string
	CrowdfundingPoolAddress   string
	ReputationScoreAddress    string
}

type IPFSConfig struct {
	Gateway       string
	PinataAPIKey  string
	PinataSecret  string
}

type JWTConfig struct {
	Secret string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	chainID, err := strconv.ParseInt(getEnv("CHAIN_ID", "84532"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid CHAIN_ID: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "tunecent_db"),
		},
		Blockchain: BlockchainConfig{
			RPCURL:                    getEnv("RPC_URL", "https://sepolia.base.org"),
			ChainID:                   chainID,
			MusicRegistryAddress:      getEnv("MUSIC_REGISTRY_ADDRESS", ""),
			RoyaltyDistributorAddress: getEnv("ROYALTY_DISTRIBUTOR_ADDRESS", ""),
			CrowdfundingPoolAddress:   getEnv("CROWDFUNDING_POOL_ADDRESS", ""),
			ReputationScoreAddress:    getEnv("REPUTATION_SCORE_ADDRESS", ""),
		},
		IPFS: IPFSConfig{
			Gateway:      getEnv("IPFS_GATEWAY", "https://gateway.pinata.cloud/ipfs/"),
			PinataAPIKey: getEnv("PINATA_API_KEY", ""),
			PinataSecret: getEnv("PINATA_SECRET_KEY", ""),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default-secret-change-in-production"),
		},
	}

	return config, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
