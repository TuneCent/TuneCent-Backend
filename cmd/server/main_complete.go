package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tunecent/backend/internal/blockchain"
	"github.com/tunecent/backend/internal/config"
	"github.com/tunecent/backend/internal/database"
	"github.com/tunecent/backend/internal/handlers"
	"github.com/tunecent/backend/internal/services"
	"github.com/tunecent/backend/pkg/fingerprint"
	"github.com/tunecent/backend/pkg/ipfs"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	log.Printf("Starting TuneCent Backend API v1.0.0-poc in %s mode", cfg.Server.Env)

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize blockchain client (optional for PoC without contract addresses)
	var blockchainClient *blockchain.Client
	var blockchainService *blockchain.Service
	if cfg.Blockchain.MusicRegistryAddress != "" {
		blockchainClient, err = blockchain.NewClient(cfg)
		if err != nil {
			log.Printf("Warning: Failed to connect to blockchain: %v", err)
			log.Println("Continuing in database-only mode")
		} else {
			blockchainService = blockchain.NewService(blockchainClient)
			defer blockchainClient.Close()
			log.Println("Blockchain client connected successfully")
		}
	} else {
		log.Println("No blockchain addresses configured, running in database-only mode")
	}

	// Initialize services
	ipfsService := ipfs.NewService(cfg)
	fingerprintService := fingerprint.NewService()

	// Initialize business logic services
	musicService := services.NewMusicService(db, ipfsService, fingerprintService, blockchainService)

	// Initialize handlers
	musicHandler := handlers.NewMusicHandler(musicService)
	campaignHandler := handlers.NewCampaignHandler(db)
	royaltyHandler := handlers.NewRoyaltyHandler(db)
	userHandler := handlers.NewUserHandler(db)

	// Setup Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Middleware
	r.Use(CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		dbHealth := "ok"
		if err := db.Ping(); err != nil {
			dbHealth = "error"
		}

		blockchainHealth := "not_configured"
		if blockchainClient != nil {
			blockchainHealth = "ok"
		}

		c.JSON(200, gin.H{
			"status":     "ok",
			"service":    "TuneCent Backend API",
			"version":    "1.0.0-poc",
			"database":   dbHealth,
			"blockchain": blockchainHealth,
		})
	})

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
	}

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("🚀 Server listening on %s", addr)
	log.Printf("📊 Health check: http://localhost%s/health", addr)
	log.Printf("📝 API docs: http://localhost%s/api/v1", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
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
