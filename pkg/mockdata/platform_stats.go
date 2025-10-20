package mockdata

import (
	"math"
	"math/rand"
	"time"
)

// PlatformStat represents statistics for a single platform
type PlatformStat struct {
	Platform  string  `json:"platform"`
	Plays     uint64  `json:"plays,omitempty"`
	Views     uint64  `json:"views,omitempty"`
	Listeners uint64  `json:"listeners,omitempty"`
	Uses      uint64  `json:"uses,omitempty"`
	Growth    float64 `json:"growth"` // percentage
}

// PlatformStats represents all platform statistics
type PlatformStats struct {
	Spotify    PlatformStat `json:"spotify"`
	TikTok     PlatformStat `json:"tiktok"`
	AppleMusic PlatformStat `json:"apple_music"`
}

// GeneratePlatformStats generates realistic mock platform stats based on token ID and registration date
// Uses tokenID as seed for consistent "random" data
func GeneratePlatformStats(tokenID uint64, registeredAt time.Time) PlatformStats {
	// Use tokenID as seed for deterministic randomness
	seed := int64(tokenID)
	r := rand.New(rand.NewSource(seed))

	// Calculate days since registration
	daysSince := time.Since(registeredAt).Hours() / 24
	if daysSince < 1 {
		daysSince = 1 // Minimum 1 day
	}

	// Growth factor increases over time (logarithmic growth)
	growthFactor := 1.0 + math.Log10(daysSince/7+1) // +1 to avoid log(0)

	// Base multipliers (different platforms have different scales)
	spotifyMultiplier := 1.0 + r.Float64()*2.0  // 1.0 - 3.0
	tiktokMultiplier := 2.0 + r.Float64()*3.0   // 2.0 - 5.0 (TikTok typically higher)
	appleMultiplier := 0.6 + r.Float64()*1.0    // 0.6 - 1.6 (Apple Music typically lower)

	// Generate Spotify stats
	spotifyPlays := uint64(randomRange(r, 5000, 50000) * spotifyMultiplier * growthFactor)
	spotifyListeners := uint64(float64(spotifyPlays) * 0.65) // ~65% play-to-listener ratio
	spotifyGrowth := randomRange(r, 100, 800) // 100-800% growth

	// Generate TikTok stats
	tiktokViews := uint64(randomRange(r, 10000, 200000) * tiktokMultiplier * growthFactor)
	tiktokUses := uint64(randomRange(r, 50, 500) * tiktokMultiplier)
	tiktokGrowth := randomRange(r, 150, 1000) // 150-1000% growth (more viral)

	// Generate Apple Music stats
	applePlays := uint64(randomRange(r, 3000, 40000) * appleMultiplier * growthFactor)
	appleListeners := uint64(float64(applePlays) * 0.70) // ~70% play-to-listener ratio
	appleGrowth := randomRange(r, 50, 500) // 50-500% growth

	return PlatformStats{
		Spotify: PlatformStat{
			Platform:  "Spotify",
			Plays:     spotifyPlays,
			Listeners: spotifyListeners,
			Growth:    spotifyGrowth,
		},
		TikTok: PlatformStat{
			Platform: "TikTok",
			Views:    tiktokViews,
			Uses:     tiktokUses,
			Growth:   tiktokGrowth,
		},
		AppleMusic: PlatformStat{
			Platform:  "Apple Music",
			Plays:     applePlays,
			Listeners: appleListeners,
			Growth:    appleGrowth,
		},
	}
}

// GenerateViralScore calculates a viral score (0-100) based on engagement metrics
func GenerateViralScore(playCount, viewCount, listenerCount uint64, daysSince float64) float64 {
	if daysSince < 1 {
		daysSince = 1
	}

	// Normalize metrics (per day)
	playsPerDay := float64(playCount) / daysSince
	viewsPerDay := float64(viewCount) / daysSince
	listenersPerDay := float64(listenerCount) / daysSince

	// Weighted scoring
	playScore := math.Min(playsPerDay/1000*30, 30)          // Max 30 points for plays
	viewScore := math.Min(viewsPerDay/2000*30, 30)          // Max 30 points for views
	listenerScore := math.Min(listenersPerDay/500*20, 20)   // Max 20 points for listeners
	timeBonus := math.Min(daysSince/30*20, 20)              // Max 20 points for longevity

	viralScore := playScore + viewScore + listenerScore + timeBonus

	// Cap at 100
	if viralScore > 100 {
		viralScore = 100
	}

	return math.Round(viralScore*100) / 100 // Round to 2 decimals
}

// GenerateTrendingRank assigns a trending rank based on viral score
// Returns 0 if not trending, 1+ if trending
func GenerateTrendingRank(viralScore float64, totalSongs int) int {
	// Only top 20% get trending rank
	trendingThreshold := 60.0

	if viralScore >= trendingThreshold {
		// Rank based on score (higher score = lower rank number = better)
		rank := int((100 - viralScore) / 5) + 1
		if rank < 1 {
			rank = 1
		}
		if rank > totalSongs/5 { // Max 20% of songs can be trending
			return 0
		}
		return rank
	}

	return 0 // Not trending
}

// GenerateEstimatedReach calculates estimated reach based on platform stats
func GenerateEstimatedReach(stats PlatformStats) uint64 {
	// Estimated reach = unique listeners/viewers across platforms
	// Assume 30% overlap between platforms
	total := stats.Spotify.Listeners + stats.TikTok.Views/10 + stats.AppleMusic.Listeners
	reach := uint64(float64(total) * 0.7) // Account for overlap

	return reach
}

// randomRange generates a random float between min and max
func randomRange(r *rand.Rand, min, max float64) float64 {
	return min + r.Float64()*(max-min)
}

// GenerateRiskScore calculates campaign risk score (0-100, lower = safer)
func GenerateRiskScore(fundingPercentage float64, contributorCount uint, creatorReputation uint) uint8 {
	// Start with high risk (100)
	risk := 100.0

	// Reduce risk based on funding progress (max 40 points reduction)
	risk -= math.Min(fundingPercentage, 100) * 0.4

	// Reduce risk based on number of contributors (max 30 points reduction)
	contributorScore := math.Min(float64(contributorCount)*2, 100)
	risk -= contributorScore * 0.3

	// Reduce risk based on creator reputation (max 30 points reduction)
	reputationScore := math.Min(float64(creatorReputation)*10, 100)
	risk -= reputationScore * 0.3

	// Ensure risk is between 0 and 100
	if risk < 0 {
		risk = 0
	}
	if risk > 100 {
		risk = 100
	}

	return uint8(risk)
}

// GenerateROI calculates estimated ROI based on campaign metrics
func GenerateROI(fundingPercentage float64, riskScore uint8, daysSinceCreation float64) float64 {
	// Base ROI
	baseROI := 150.0

	// Higher funding = higher confidence = higher ROI
	fundingBonus := (fundingPercentage / 100) * 50 // Up to +50%

	// Lower risk = higher ROI
	riskPenalty := float64(riskScore) * 0.5 // Risk reduces ROI

	// Time factor (diminishing returns)
	timeFactor := math.Min(daysSinceCreation/30, 5) * 10 // Up to +50% for older campaigns

	estimatedROI := baseROI + fundingBonus - riskPenalty + timeFactor

	// ROI typically ranges from 80% to 300%
	if estimatedROI < 80 {
		estimatedROI = 80
	}
	if estimatedROI > 300 {
		estimatedROI = 300
	}

	return math.Round(estimatedROI*100) / 100
}

// GenerateTier returns creator tier based on total works and earnings
func GenerateTier(totalWorks uint, totalEarningsEth float64) string {
	// Legendary: 50+ works OR 100+ ETH earned
	if totalWorks >= 50 || totalEarningsEth >= 100 {
		return "Legendary Creator"
	}

	// Rising Star: 20+ works OR 50+ ETH earned
	if totalWorks >= 20 || totalEarningsEth >= 50 {
		return "Rising Star"
	}

	// Established: 10+ works OR 20+ ETH earned
	if totalWorks >= 10 || totalEarningsEth >= 20 {
		return "Established Creator"
	}

	// Verified: 5+ works OR 5+ ETH earned
	if totalWorks >= 5 || totalEarningsEth >= 5 {
		return "Verified Creator"
	}

	// Default
	return "Registered Creator"
}

// WeeklyGrowth calculates week-over-week growth percentage (mock)
func WeeklyGrowth(currentValue, previousValue float64) float64 {
	if previousValue == 0 {
		return 0
	}

	growth := ((currentValue - previousValue) / previousValue) * 100
	return math.Round(growth*100) / 100
}
