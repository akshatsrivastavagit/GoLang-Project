// Package main Omnichannel Inventory API
//
// @title Omnichannel Inventory API
// @version 1.0
// @description Real-time multi-warehouse inventory sync for omnichannel retail.
// @BasePath /
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"omnichannel_inventory/internal/db"
	"omnichannel_inventory/internal/events"
	"omnichannel_inventory/internal/handlers"
	"omnichannel_inventory/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Debug environment variables
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Printf("WARNING: WEBHOOK_URL is not set in environment variables")
	} else {
		log.Printf("Webhook URL configured: %s", webhookURL)
	}

	// Initialize database connections
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	if err := db.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer db.CloseRedis()

	// Debug environment variables
	slackURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackURL == "" {
		log.Printf("WARNING: SLACK_WEBHOOK_URL is not set in environment variables")
	} else {
		log.Printf("Slack webhook URL configured: %s", slackURL)
	}

	// Initialize services
	inventoryService := services.NewInventoryService(db.GetDB(), db.GetRedis())
	handlers.SetInventoryService(inventoryService)

	// Initialize event processor
	processor := events.NewEventProcessor()
	ctx := context.Background()
	processor.Start(ctx)
	defer processor.Stop()

	// Start event consumer
	go events.StartInventoryEventConsumer(ctx, processor)

	// Create Gin router
	router := gin.Default()

	// Add logging middleware
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
				param.ClientIP,
				param.TimeStamp.Format(time.RFC1123),
				param.Method,
				param.Path,
				param.Request.Proto,
				param.StatusCode,
				param.Latency,
				param.Request.UserAgent(),
				param.ErrorMessage,
			)
		},
		Output: gin.DefaultWriter,
	}))

	// Serve static files
	router.Static("/static", "./static")
	router.LoadHTMLGlob("static/*.html")

	// API routes
	api := router.Group("/api")
	{
		api.POST("/stock", handlers.AddOrUpdateStock)
		api.GET("/stock/:sku", handlers.GetConsolidatedStock)
		api.POST("/orders/simulate", handlers.SimulateOrder)
		api.GET("/history/:sku", handlers.GetInventoryHistory)
	}

	// Debug endpoint
	router.GET("/debug/env", func(c *gin.Context) {
		envVars := map[string]string{
			"SLACK_WEBHOOK_URL": os.Getenv("SLACK_WEBHOOK_URL"),
		}
		c.JSON(200, envVars)
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Web interface routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 