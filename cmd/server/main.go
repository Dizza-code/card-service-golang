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
	// Update the arguments to match the actual NewClient signature in your api package
	apiClient := api.NewClient(cfg.CardAPIBaseURL, cfg.SecureAPIBaseURL, cfg.CardAPIKey)
	customerService := services.NewCustomerService(db, apiClient, logger)
	cardService := services.NewCardService(db, apiClient, logger)

	// Initialize handlers
	customerHandler := handlers.NewCustomerHandler(customerService, cfg.SettlementAccount, logger)
	cardHandler := handlers.NewCardHandler(cardService, logger)

	// Set up Gin router
	r := gin.Default()
	r.POST("/api/customers", customerHandler.CreateCustomer)
	r.POST("/api/cards", cardHandler.LinkCard)
	// Start server
	logger.Info("Starting server", zap.String("port", cfg.Port))
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
