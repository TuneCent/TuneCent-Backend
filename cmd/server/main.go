package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tunecent/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "TuneCent Backend API",
			"version": "1.0.0-poc",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Music routes
		music := v1.Group("/music")
		{
			music.POST("/register", RegisterMusic(db))
			music.GET("/:tokenId", GetMusic(db))
			music.GET("/", ListMusic(db))
			music.GET("/:tokenId/analytics", GetMusicAnalytics(db))
		}

		// Campaign routes
		campaigns := v1.Group("/campaigns")
		{
			campaigns.POST("/", CreateCampaign(db))
			campaigns.GET("/:campaignId", GetCampaign(db))
			campaigns.GET("/", ListCampaigns(db))
			campaigns.POST("/:campaignId/contribute", Contribute(db))
		}

		// Royalty routes
		royalties := v1.Group("/royalties")
		{
			royalties.GET("/token/:tokenId", GetRoyalties(db))
			royalties.POST("/simulate", SimulateRoyaltyPayment(db))
		}

		// User/Reputation routes
		users := v1.Group("/users")
		{
			users.GET("/:address", GetUserProfile(db))
			users.GET("/:address/reputation", GetReputation(db))
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting TuneCent Backend API on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func initDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.MusicMetadata{},
		&models.Campaign{},
		&models.Contribution{},
		&models.RoyaltyPayment{},
		&models.RoyaltyDistribution{},
		&models.UsageDetection{},
		&models.Analytics{},
	)
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

// Placeholder handlers (to be implemented)
func RegisterMusic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - register music endpoint"})
	}
}

func GetMusic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - get music endpoint"})
	}
}

func ListMusic(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - list music endpoint"})
	}
}

func GetMusicAnalytics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - get analytics endpoint"})
	}
}

func CreateCampaign(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - create campaign endpoint"})
	}
}

func GetCampaign(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - get campaign endpoint"})
	}
}

func ListCampaigns(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - list campaigns endpoint"})
	}
}

func Contribute(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - contribute endpoint"})
	}
}

func GetRoyalties(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - get royalties endpoint"})
	}
}

func SimulateRoyaltyPayment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - simulate royalty payment endpoint"})
	}
}

func GetUserProfile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - get user profile endpoint"})
	}
}

func GetReputation(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "Not implemented - get reputation endpoint"})
	}
}
