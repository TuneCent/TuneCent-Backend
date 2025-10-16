-- TuneCent Database Schema
-- MySQL 8.0+

CREATE DATABASE IF NOT EXISTS tunecent_db
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE tunecent_db;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    wallet_address VARCHAR(42) NOT NULL UNIQUE,
    username VARCHAR(100) UNIQUE,
    email VARCHAR(255) UNIQUE,
    role ENUM('creator', 'contributor', 'both') DEFAULT 'contributor',
    is_verified BOOLEAN DEFAULT FALSE,
    reputation_score INT UNSIGNED DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_wallet (wallet_address),
    INDEX idx_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Music metadata table
CREATE TABLE IF NOT EXISTS music_metadata (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_id BIGINT UNSIGNED NOT NULL UNIQUE,
    creator_address VARCHAR(42) NOT NULL,
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    genre VARCHAR(100),
    description TEXT,
    ipfs_cid VARCHAR(100) NOT NULL,
    fingerprint_hash VARCHAR(66) NOT NULL UNIQUE,
    audio_file_url VARCHAR(500),
    cover_image_url VARCHAR(500),
    duration INT UNSIGNED COMMENT 'Duration in seconds',
    is_active BOOLEAN DEFAULT TRUE,
    tx_hash VARCHAR(66),
    registered_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_token_id (token_id),
    INDEX idx_creator (creator_address),
    INDEX idx_fingerprint (fingerprint_hash),
    INDEX idx_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Campaigns table
CREATE TABLE IF NOT EXISTS campaigns (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    campaign_id BIGINT UNSIGNED NOT NULL UNIQUE,
    token_id BIGINT UNSIGNED NOT NULL,
    creator_address VARCHAR(42) NOT NULL,
    goal_amount VARCHAR(78) NOT NULL COMMENT 'Wei as string',
    raised_amount VARCHAR(78) DEFAULT '0',
    royalty_percentage SMALLINT UNSIGNED NOT NULL COMMENT 'Basis points',
    deadline TIMESTAMP NOT NULL,
    lockup_period INT UNSIGNED NOT NULL COMMENT 'Days',
    status ENUM('active', 'successful', 'failed', 'cancelled') DEFAULT 'active',
    funds_withdrawn BOOLEAN DEFAULT FALSE,
    tx_hash VARCHAR(66),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_campaign_id (campaign_id),
    INDEX idx_token_id (token_id),
    INDEX idx_creator (creator_address),
    INDEX idx_status (status),
    INDEX idx_deleted (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Contributions table
CREATE TABLE IF NOT EXISTS contributions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    campaign_id BIGINT UNSIGNED NOT NULL,
    contributor_address VARCHAR(42) NOT NULL,
    amount VARCHAR(78) NOT NULL COMMENT 'Wei as string',
    share_percentage DECIMAL(10,8) NOT NULL,
    tx_hash VARCHAR(66),
    contributed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_campaign_id (campaign_id),
    INDEX idx_contributor (contributor_address),
    FOREIGN KEY (campaign_id) REFERENCES campaigns(campaign_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Royalty payments table
CREATE TABLE IF NOT EXISTS royalty_payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_id BIGINT UNSIGNED NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    amount VARCHAR(78) NOT NULL COMMENT 'Wei as string',
    platform VARCHAR(100) NOT NULL,
    usage_type VARCHAR(100),
    tx_hash VARCHAR(66) NOT NULL,
    is_distributed BOOLEAN DEFAULT FALSE,
    distributed_at TIMESTAMP NULL,
    paid_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token_id (token_id),
    INDEX idx_platform (platform),
    INDEX idx_distributed (is_distributed)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Royalty distributions table
CREATE TABLE IF NOT EXISTS royalty_distributions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_id BIGINT UNSIGNED NOT NULL,
    token_id BIGINT UNSIGNED NOT NULL,
    beneficiary VARCHAR(42) NOT NULL,
    amount VARCHAR(78) NOT NULL COMMENT 'Wei as string',
    tx_hash VARCHAR(66) NOT NULL,
    distributed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_payment_id (payment_id),
    INDEX idx_token_id (token_id),
    INDEX idx_beneficiary (beneficiary),
    FOREIGN KEY (payment_id) REFERENCES royalty_payments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Usage detection table (for PoC demo)
CREATE TABLE IF NOT EXISTS usage_detections (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_id BIGINT UNSIGNED NOT NULL,
    platform VARCHAR(100) NOT NULL,
    content_id VARCHAR(255),
    content_url VARCHAR(500),
    detected_at TIMESTAMP NOT NULL,
    payment_sent BOOLEAN DEFAULT FALSE,
    payment_tx_hash VARCHAR(66),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_token_id (token_id),
    INDEX idx_platform (platform),
    INDEX idx_payment_sent (payment_sent)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Analytics table
CREATE TABLE IF NOT EXISTS analytics (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token_id BIGINT UNSIGNED NOT NULL UNIQUE,
    total_views BIGINT UNSIGNED DEFAULT 0,
    total_embeds BIGINT UNSIGNED DEFAULT 0,
    total_usages BIGINT UNSIGNED DEFAULT 0,
    total_royalties VARCHAR(78) DEFAULT '0' COMMENT 'Wei as string',
    last_updated TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_token_id (token_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample data for testing (optional)
-- INSERT INTO users (wallet_address, username, role, is_verified) VALUES
-- ('0x1234567890123456789012345678901234567890', 'demo_creator', 'creator', TRUE),
-- ('0x0987654321098765432109876543210987654321', 'demo_fan', 'contributor', FALSE);
