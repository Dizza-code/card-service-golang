package main

import (
	"card-service/internal/api"
	"card-service/internal/handlers"
	"card-service/internal/services"
	"card-service/internal/store"
	"card-service/pkg/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// main initializes and starts the application.
func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration from .env
	cfg, err := config.Load(logger)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to MongoDB
	db, err := store.NewStore(cfg.DatabaseURL, "card_service")
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer db.Close()

	// Initialize services
	apiClient := api.NewClient(cfg.CardAPIBaseURL, cfg.CardAPIKey)
	customerService := services.NewCustomerService(db, apiClient, logger)

	// Initialize handlers
	customerHandler := handlers.NewCustomerHandler(customerService, cfg.SettlementAccount, logger)

	// Set up Gin router
	r := gin.Default()
	r.POST("/api/customers", customerHandler.CreateCustomer)

	// Start server
	logger.Info("Starting server", zap.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
