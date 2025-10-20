-- =====================================================
-- TuneCent PoC Database Migrations
-- Adds fields needed for dashboard, analytics, and pools
-- =====================================================

-- Add demo stats to music_metadata
ALTER TABLE music_metadata
ADD COLUMN IF NOT EXISTS play_count BIGINT UNSIGNED DEFAULT 0 COMMENT 'Total play count across platforms',
ADD COLUMN IF NOT EXISTS view_count BIGINT UNSIGNED DEFAULT 0 COMMENT 'Total video views (TikTok, YouTube, etc)',
ADD COLUMN IF NOT EXISTS listener_count BIGINT UNSIGNED DEFAULT 0 COMMENT 'Unique listeners count',
ADD COLUMN IF NOT EXISTS viral_score DECIMAL(5,2) DEFAULT 0.00 COMMENT 'Viral score 0-100',
ADD COLUMN IF NOT EXISTS trending_rank INT DEFAULT 0 COMMENT '0 = not trending, 1+ = rank',
ADD COLUMN IF NOT EXISTS genre VARCHAR(100) DEFAULT NULL COMMENT 'Music genre',
ADD COLUMN IF NOT EXISTS description TEXT DEFAULT NULL COMMENT 'Music description',
ADD COLUMN IF NOT EXISTS duration INT DEFAULT 0 COMMENT 'Duration in seconds';

-- Add index for trending queries
CREATE INDEX IF NOT EXISTS idx_music_trending ON music_metadata(trending_rank, viral_score DESC);
CREATE INDEX IF NOT EXISTS idx_music_creator ON music_metadata(creator_address, created_at DESC);

-- Add pool stats to campaigns
ALTER TABLE campaigns
ADD COLUMN IF NOT EXISTS risk_score TINYINT UNSIGNED DEFAULT 50 COMMENT 'Risk score 0-100 (lower=safer)',
ADD COLUMN IF NOT EXISTS is_trending BOOLEAN DEFAULT FALSE COMMENT 'Is this pool trending',
ADD COLUMN IF NOT EXISTS estimated_roi DECIMAL(10,2) DEFAULT 150.00 COMMENT 'Estimated ROI percentage',
ADD COLUMN IF NOT EXISTS contributor_count INT UNSIGNED DEFAULT 0 COMMENT 'Number of contributors';

-- Add index for trending pools
CREATE INDEX IF NOT EXISTS idx_campaigns_trending ON campaigns(is_trending, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_campaigns_risk ON campaigns(risk_score, status);

-- Add display fields to users
ALTER TABLE users
ADD COLUMN IF NOT EXISTS display_name VARCHAR(255) DEFAULT NULL COMMENT 'Display name',
ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT NULL COMMENT 'User bio',
ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500) DEFAULT NULL COMMENT 'Avatar image URL',
ADD COLUMN IF NOT EXISTS tier VARCHAR(50) DEFAULT 'Registered Creator' COMMENT 'Creator tier badge',
ADD COLUMN IF NOT EXISTS leaderboard_rank INT UNSIGNED DEFAULT 0 COMMENT 'Leaderboard position',
ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE COMMENT 'Verified creator status',
ADD COLUMN IF NOT EXISTS total_earnings VARCHAR(78) DEFAULT '0' COMMENT 'Total lifetime earnings in Wei',
ADD COLUMN IF NOT EXISTS total_works INT UNSIGNED DEFAULT 0 COMMENT 'Total music registered';

-- Add index for leaderboard
CREATE INDEX IF NOT EXISTS idx_users_leaderboard ON users(leaderboard_rank);
CREATE INDEX IF NOT EXISTS idx_users_earnings ON users(total_earnings DESC);

-- Create transactions table (NEW)
CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_address VARCHAR(42) NOT NULL COMMENT 'User wallet address',
    type VARCHAR(20) NOT NULL COMMENT 'Transaction type: royalty, invest, withdraw, etc',
    amount VARCHAR(78) DEFAULT NULL COMMENT 'Amount in Wei',
    tx_hash VARCHAR(66) DEFAULT NULL COMMENT 'Blockchain transaction hash',
    status VARCHAR(20) DEFAULT 'confirmed' COMMENT 'Transaction status',
    description TEXT DEFAULT NULL COMMENT 'Human-readable description',
    related_id BIGINT UNSIGNED DEFAULT NULL COMMENT 'Related entity ID (token_id, campaign_id)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_created_at (created_at DESC),
    INDEX idx_type (type),
    INDEX idx_tx_hash (tx_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Transaction history for wallet page';

-- Create activities table (NEW)
CREATE TABLE IF NOT EXISTS activities (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_address VARCHAR(42) NOT NULL COMMENT 'User wallet address',
    type VARCHAR(50) NOT NULL COMMENT 'Activity type: music_registered, royalty_received, etc',
    title VARCHAR(255) NOT NULL COMMENT 'Activity title',
    description TEXT DEFAULT NULL COMMENT 'Activity description',
    related_id BIGINT UNSIGNED DEFAULT NULL COMMENT 'Related entity ID',
    tx_hash VARCHAR(66) DEFAULT NULL COMMENT 'Related transaction hash',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_address (user_address),
    INDEX idx_created_at (created_at DESC),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Activity feed for dashboard';

-- =====================================================
-- Trigger: Auto-update contributor_count on campaigns
-- =====================================================
DELIMITER //

CREATE TRIGGER IF NOT EXISTS update_campaign_contributor_count
AFTER INSERT ON contributions
FOR EACH ROW
BEGIN
    UPDATE campaigns
    SET contributor_count = (
        SELECT COUNT(DISTINCT contributor_address)
        FROM contributions
        WHERE campaign_id = NEW.campaign_id
    )
    WHERE campaign_id = NEW.campaign_id;
END//

DELIMITER ;

-- =====================================================
-- Stored Procedure: Calculate Risk Score
-- =====================================================
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS calculate_risk_score(IN p_campaign_id BIGINT UNSIGNED)
BEGIN
    DECLARE v_raised DECIMAL(30,0);
    DECLARE v_goal DECIMAL(30,0);
    DECLARE v_contributors INT;
    DECLARE v_creator_reputation INT;
    DECLARE v_funding_progress DECIMAL(10,2);
    DECLARE v_risk_score INT;

    -- Get campaign data
    SELECT
        CAST(raised_amount AS DECIMAL(30,0)),
        CAST(goal_amount AS DECIMAL(30,0)),
        contributor_count,
        creator_address
    INTO v_raised, v_goal, v_contributors, @creator_addr
    FROM campaigns
    WHERE campaign_id = p_campaign_id;

    -- Get creator reputation (simplified: count of works)
    SELECT COUNT(*) INTO v_creator_reputation
    FROM music_metadata
    WHERE creator_address = @creator_addr;

    -- Calculate funding progress percentage
    SET v_funding_progress = IF(v_goal > 0, (v_raised / v_goal) * 100, 0);

    -- Calculate risk score (lower = safer)
    -- Formula: 100 - (funding% * 0.4 + contributors * 2 * 0.3 + reputation * 0.3)
    SET v_risk_score = 100 - LEAST(100,
        (v_funding_progress * 0.4) +
        (LEAST(v_contributors * 2, 100) * 0.3) +
        (LEAST(v_creator_reputation * 10, 100) * 0.3)
    );

    -- Update campaign
    UPDATE campaigns
    SET risk_score = v_risk_score
    WHERE campaign_id = p_campaign_id;
END//

DELIMITER ;

-- =====================================================
-- Sample Data: Add mock stats to existing music
-- =====================================================

-- Update existing music with realistic mock stats (only if play_count is 0)
UPDATE music_metadata
SET
    play_count = FLOOR(1000 + (RAND() * 50000)),
    view_count = FLOOR(5000 + (RAND() * 100000)),
    listener_count = FLOOR(500 + (RAND() * 30000)),
    viral_score = ROUND(20 + (RAND() * 80), 2),
    trending_rank = IF(RAND() > 0.7, FLOOR(1 + (RAND() * 20)), 0)
WHERE play_count = 0 OR play_count IS NULL;

-- Update existing campaigns with risk scores
UPDATE campaigns c
JOIN (
    SELECT campaign_id, COUNT(DISTINCT contributor_address) as cnt
    FROM contributions
    GROUP BY campaign_id
) contrib ON c.campaign_id = contrib.campaign_id
SET
    c.contributor_count = contrib.cnt,
    c.is_trending = IF(RAND() > 0.6 AND c.status = 'active', TRUE, FALSE),
    c.estimated_roi = ROUND(100 + (RAND() * 200), 2);

-- Mark top campaigns as trending
UPDATE campaigns
SET is_trending = TRUE
WHERE status = 'active'
ORDER BY (CAST(raised_amount AS DECIMAL) / CAST(goal_amount AS DECIMAL)) DESC
LIMIT 5;

-- =====================================================
-- Views: Helpful queries for dashboard
-- =====================================================

-- View: Top Artists Leaderboard
CREATE OR REPLACE VIEW vw_leaderboard AS
SELECT
    u.wallet_address,
    u.display_name,
    u.tier,
    u.is_verified,
    COUNT(DISTINCT m.token_id) as total_works,
    COALESCE(SUM(CAST(rp.amount AS DECIMAL(30,0))), 0) as total_earnings,
    COUNT(DISTINCT c.campaign_id) as total_campaigns,
    (
        COUNT(DISTINCT m.token_id) * 100 +
        COALESCE(SUM(CAST(rp.amount AS DECIMAL(30,0))) / 1e18, 0) * 10 +
        COUNT(DISTINCT c.campaign_id) * 50
    ) as score
FROM users u
LEFT JOIN music_metadata m ON u.wallet_address = m.creator_address
LEFT JOIN royalty_payments rp ON m.token_id = rp.token_id AND rp.is_distributed = TRUE
LEFT JOIN campaigns c ON u.wallet_address = c.creator_address
WHERE u.role = 'creator'
GROUP BY u.wallet_address
ORDER BY score DESC;

-- View: Trending Music
CREATE OR REPLACE VIEW vw_trending_music AS
SELECT
    m.*,
    u.display_name as creator_name,
    u.is_verified as creator_verified,
    COALESCE(SUM(CAST(rp.amount AS DECIMAL(30,0))), 0) as total_royalties
FROM music_metadata m
LEFT JOIN users u ON m.creator_address = u.wallet_address
LEFT JOIN royalty_payments rp ON m.token_id = rp.token_id
WHERE m.trending_rank > 0
GROUP BY m.token_id
ORDER BY m.trending_rank ASC, m.viral_score DESC;

-- View: Active Trending Pools
CREATE OR REPLACE VIEW vw_trending_pools AS
SELECT
    c.*,
    m.title as music_title,
    m.artist as music_artist,
    u.display_name as creator_name,
    u.is_verified as creator_verified,
    (CAST(c.raised_amount AS DECIMAL(30,0)) / CAST(c.goal_amount AS DECIMAL(30,0)) * 100) as funding_percentage
FROM campaigns c
JOIN music_metadata m ON c.token_id = m.token_id
JOIN users u ON c.creator_address = u.wallet_address
WHERE c.status = 'active' AND c.is_trending = TRUE
ORDER BY funding_percentage DESC, c.created_at DESC;

COMMIT;
