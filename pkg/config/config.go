package config

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	DatabaseURL       string
	WebhookSigningKey string
	CardAPIKey        string
	CardAPIBaseURL    string
	Port              string
	SettlementAccount string
}

// func Load() (*Config, error) {
// 	godotenv.Load()
// 	return &Config{
// 		DatabaseURL:       os.Getenv("DATABASE_URL"),
// 		WebhookSigningKey: os.Getenv("WEBHOOK_SIGNING_KEY"),
// 		CardAPIKey:        os.Getenv("CARD_API_KEY"),
// 		CardAPIBaseURL:    os.Getenv("CARD_API_BASE_URL"),
// 		Port:              os.Getenv("PORT"),
// 		SettlementAccount: os.Getenv("SETTLEMENT_ACCOUNT"),
// 	}, nil

// }

func Load(logger *zap.Logger) (*Config, error) {
	wd, _ := os.Getwd()
	logger.Info("Attempting to load .env file", zap.String("workingDir", wd))
	if err := godotenv.Load(); err != nil {
		logger.Warn("Failed to load .env file; relying on environment variables", zap.Error(err))
	}
	settlementAccount := os.Getenv("SETTLEMENT_ACCOUNT")
	logger.Info("Raw SETTLEMENT_ACCOUNT", zap.String("value", settlementAccount))

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		CardAPIKey:  "sk.test.b8mWEuSx.pGgN2smXZekawW1$F6F66x0CO-RG6zx$mcxYS4c5x1Ikdx7kZhK@ZKKrQNm8Wk00",
		// CardAPIKey:        os.Getenv("CARD_SERVER_API_KEY"),
		CardAPIBaseURL:    os.Getenv("CARD_API_BASE_URL"),
		Port:              os.Getenv("PORT"),
		SettlementAccount: settlementAccount,
	}
	if cfg.CardAPIKey == "" {
		logger.Error("CARD_SERVER_API_KEY is empty")
		return nil, fmt.Errorf("CARD_SERVER_API_KEY is required")
	}

	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(cfg.CardAPIKey)))
	logger.Info("Loaded configuration",
		zap.String("databaseURL", cfg.DatabaseURL),
		zap.String("cardAPIKeyPrefix", cfg.CardAPIKey[:4]),
		zap.Int("cardAPIKeyLength", len(cfg.CardAPIKey)), // Debug length
		zap.String("cardAPIKeyHash", keyHash),
		zap.String("cardAPIBaseURL", cfg.CardAPIBaseURL),
		zap.String("port", cfg.Port),
		zap.String("settlementAccount", cfg.SettlementAccount),
	)
	return cfg, nil
}
