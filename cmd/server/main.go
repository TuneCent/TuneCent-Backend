package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tunecent/backend/internal/config"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/handlers"
	"github.com/tunecent/backend/internal/models"
	"github.com/tunecent/backend/internal/services"
	"github.com/tunecent/backend/pkg/fingerprint"
	"github.com/tunecent/backend/pkg/ipfs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/tunecent/backend/docs"
)

// @title TuneCent Backend API
// @version 1.0
// @description Complete TuneCent Backend API with 68 endpoints for music NFT, campaigns, royalties, analytics, and more
// @termsOfService http://swagger.io/terms/

// @contact.name TuneCent API Support
// @contact.url https://github.com/tunecent
// @contact.email support@tunecent.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @tag.name Health
// @tag.description Health check endpoints

// @tag.name Music
// @tag.description Music NFT registration and management endpoints

// @tag.name Campaigns
// @tag.description Crowdfunding campaign endpoints

// @tag.name Royalties
// @tag.description Royalty payment and simulation endpoints

// @tag.name Users
// @tag.description User profile and reputation endpoints

// @tag.name Dashboard
// @tag.description Dashboard overview and statistics endpoints

// @tag.name Analytics
// @tag.description Music analytics and metrics endpoints

// @tag.name Wallet
// @tag.description Wallet transaction and balance endpoints

// @tag.name Leaderboard
// @tag.description Leaderboard and ranking endpoints

// @tag.name Portfolio
// @tag.description Portfolio and investment tracking endpoints

// @tag.name Distribution
// @tag.description Music distribution management endpoints

// @tag.name Notifications
// @tag.description Notification management endpoints

// @tag.name Ledger
// @tag.description Revenue split ledger endpoints

// @tag.name Audit
// @tag.description Blockchain audit and verification endpoints

// @tag.name Reinvestment
// @tag.description Reinvestment suggestions and tracking endpoints

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	gormDB, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Wrap GORM DB in our database wrapper
	db := &database.DB{DB: gormDB}

	// Run migrations - DISABLED for PoC (using schema.sql instead)
	// if err := runMigrations(gormDB); err != nil {
	// 	log.Fatal("Failed to run migrations:", err)
	// }

	// Initialize services
	ipfsService := ipfs.NewService(cfg)
	fingerprintService := fingerprint.NewService()
	musicService := services.NewMusicService(db, ipfsService, fingerprintService, nil)
	distributionService := services.NewDistributionService(db)
	notificationService := services.NewNotificationService(db)
	ledgerService := services.NewLedgerService(db)
	reinvestmentService := services.NewReinvestmentService(db)

	// Initialize handlers
	musicHandler := handlers.NewMusicHandler(musicService)
	campaignHandler := handlers.NewCampaignHandler(db)
	royaltyHandler := handlers.NewRoyaltyHandler(db)
	userHandler := handlers.NewUserHandler(db)

	// PoC handlers
	dashboardHandler := handlers.NewDashboardHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(db)
	walletHandler := handlers.NewWalletHandler(db)
	leaderboardHandler := handlers.NewLeaderboardHandler(db)
	portfolioHandler := handlers.NewPortfolioHandler(db)

	// New service handlers
	distributionHandler := handlers.NewDistributionHandler(distributionService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	ledgerHandler := handlers.NewLedgerHandler(ledgerService)
	reinvestmentHandler := handlers.NewReinvestmentHandler(reinvestmentService)

	// Initialize Gin router
	r := gin.Default()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	r.GET("/health", HealthCheck)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Music routes
		music := v1.Group("/music")
		{
			music.POST("/register", musicHandler.RegisterMusic)
			music.GET("/:tokenId", musicHandler.GetMusic)
			music.GET("/", musicHandler.ListMusic)
			music.GET("/:tokenId/analytics", musicHandler.GetMusicAnalytics)
		}

		// Campaign routes
		campaigns := v1.Group("/campaigns")
		{
			campaigns.POST("/", campaignHandler.CreateCampaign)
			campaigns.GET("/:campaignId", campaignHandler.GetCampaign)
			campaigns.GET("/", campaignHandler.ListCampaigns)
			campaigns.POST("/:campaignId/contribute", campaignHandler.Contribute)
		}

		// Royalty routes
		royalties := v1.Group("/royalties")
		{
			royalties.GET("/token/:tokenId", royaltyHandler.GetRoyalties)
			royalties.POST("/simulate", royaltyHandler.SimulateRoyaltyPayment)
		}

		// User/Reputation routes
		users := v1.Group("/users")
		{
			users.GET("/:address", userHandler.GetUserProfile)
			users.GET("/:address/reputation", userHandler.GetReputation)
		}

		// Dashboard routes (PoC)
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/overview", dashboardHandler.GetOverview)
			dashboard.GET("/quick-stats", dashboardHandler.GetQuickStats)
			dashboard.GET("/trending-pools", dashboardHandler.GetTrendingPools)
			dashboard.GET("/activities", dashboardHandler.GetRecentActivities)
			dashboard.GET("/music-trends", dashboardHandler.GetMusicTrends)
			dashboard.GET("/viral-performance", dashboardHandler.GetViralPerformance)
			dashboard.GET("/weekly-progress", dashboardHandler.GetWeeklyProgress)
			dashboard.GET("/royalty-pulse", dashboardHandler.GetRoyaltyPulse)
		}

		// Analytics routes (PoC)
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/:tokenId/platform-stats", analyticsHandler.GetPlatformStats)
			analytics.GET("/:tokenId/viral-score", analyticsHandler.GetViralScore)
			analytics.GET("/:tokenId/growth", analyticsHandler.GetGrowthMetrics)
			analytics.GET("/:tokenId/listeners", analyticsHandler.GetListenerMetrics)
			analytics.GET("/:tokenId/views", analyticsHandler.GetViewMetrics)
			analytics.GET("/:tokenId/trending", analyticsHandler.GetTrendingIndicators)
			analytics.GET("/:tokenId/reach", analyticsHandler.GetEstimatedReach)
			analytics.GET("/global/top-songs", analyticsHandler.GetTopSongs)
		}

		// Wallet routes (PoC)
		wallet := v1.Group("/wallet")
		{
			wallet.GET("/:address/transactions", walletHandler.GetTransactions)
			wallet.GET("/:address/balance", walletHandler.GetBalance)
			wallet.GET("/:address/search", walletHandler.SearchTransactions)
			wallet.GET("/:address/savings", walletHandler.GetSavings)
		}

		// Leaderboard routes (PoC)
		leaderboard := v1.Group("/leaderboard")
		{
			leaderboard.GET("/top-artists", leaderboardHandler.GetTopArtists)
			leaderboard.GET("/:address/rank", leaderboardHandler.GetUserRank)
			leaderboard.GET("/stats", leaderboardHandler.GetLeaderboardStats)
		}

		// Portfolio routes (PoC)
		portfolio := v1.Group("/portfolio")
		{
			portfolio.GET("/:address", portfolioHandler.GetPortfolio)
			portfolio.GET("/:address/growth", portfolioHandler.GetGrowthStats)
			portfolio.GET("/:address/performance", portfolioHandler.GetPerformanceMetrics)
			portfolio.GET("/:address/pools", portfolioHandler.GetPoolsInvested)
		}

		// Distribution routes
		distribution := v1.Group("/distribution")
		{
			distribution.POST("/submit", distributionHandler.SubmitDistribution)
			distribution.GET("/:tokenId/status", distributionHandler.GetDistributionStatus)
			distribution.GET("/:tokenId/platform/:platform", distributionHandler.GetPlatformStatus)
			distribution.PUT("/:tokenId/platform/:platform", distributionHandler.UpdatePlatformStatus)
			distribution.GET("/list", distributionHandler.ListDistributions)
		}

		// Notification routes
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.GET("/unread/count", notificationHandler.GetUnreadCount)
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
			notifications.PUT("/read-all", notificationHandler.MarkAllAsRead)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			notifications.GET("/preferences", notificationHandler.GetPreferences)
			notifications.PUT("/preferences", notificationHandler.UpdatePreferences)
		}

		// Ledger routes
		ledger := v1.Group("/ledger")
		{
			ledger.GET("/:tokenId/splits", ledgerHandler.GetSplitHistory)
			ledger.GET("/:tokenId/contributors", ledgerHandler.GetContributorBreakdown)
			ledger.GET("/audit/:txHash", ledgerHandler.GetSplitByTxHash)
			ledger.GET("/user/:address", ledgerHandler.GetUserLedger)
		}

		// Audit routes
		audit := v1.Group("/audit")
		{
			audit.GET("/transaction/:txHash", walletHandler.GetTransactionAudit)
			audit.GET("/verify/:txHash", walletHandler.VerifyTransaction)
			audit.GET("/block/:blockNumber", walletHandler.GetBlockDetails)
		}

		// Reinvestment routes
		reinvest := v1.Group("/reinvest")
		{
			reinvest.GET("/suggestions", reinvestmentHandler.GetSuggestions)
			reinvest.POST("/quick", reinvestmentHandler.QuickReinvest)
			reinvest.GET("/history", reinvestmentHandler.GetHistory)
			reinvest.GET("/stats", reinvestmentHandler.GetStats)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 TuneCent Backend API starting on port %s", port)
	log.Printf("📊 Total endpoints: 68")
	log.Printf("✅ Music endpoints: 4")
	log.Printf("✅ Campaign endpoints: 4")
	log.Printf("✅ Royalty endpoints: 2")
	log.Printf("✅ User endpoints: 2")
	log.Printf("✅ Dashboard endpoints: 8")
	log.Printf("✅ Analytics endpoints: 8")
	log.Printf("✅ Wallet endpoints: 4")
	log.Printf("✅ Leaderboard endpoints: 3")
	log.Printf("✅ Portfolio endpoints: 4")
	log.Printf("✅ Distribution endpoints: 5")
	log.Printf("✅ Notification endpoints: 7")
	log.Printf("✅ Ledger endpoints: 4")
	log.Printf("✅ Audit endpoints: 3")
	log.Printf("✅ Reinvestment endpoints: 4")
	log.Printf("🎯 PoC Mode: Using mock data for platform stats")
	log.Printf("🆕 New Features: Distribution Hub, Notifications, Split Ledger, Audit Tools, Reinvestment")

	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDB() (*gorm.DB, error) {
	// Get database credentials from environment
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "tunecent_db"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("✅ Database connected successfully")
	return db, nil
}

func runMigrations(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.MusicMetadata{},
		&models.Campaign{},
		&models.Contribution{},
		&models.RoyaltyPayment{},
		&models.RoyaltyDistribution{},
		&models.UsageDetection{},
		&models.Analytics{},
		&models.Transaction{},
		&models.Activity{},
		&models.DistributionSubmission{},
		&models.PlatformDistribution{},
		&models.Notification{},
		&models.NotificationPreference{},
		&models.SplitRecord{},
		&models.ReinvestmentSuggestion{},
		&models.ReinvestmentHistory{},
	)

	if err != nil {
		return err
	}

	log.Println("✅ Migrations completed successfully")
	return nil
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the API service
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status"
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "TuneCent Backend API",
		"version": "1.0.0-poc",
	})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
