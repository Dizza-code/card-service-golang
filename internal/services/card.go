package services

import (
	"card-service/internal/api"
	"card-service/internal/models"
	"card-service/internal/store"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

//card service handles card operations

type CardService struct {
	store  *store.Store
	client *api.Client
	logger *zap.Logger
}

// New card service intialize a new card service instances with provided client, store
func NewCardService(store *store.Store, client *api.Client, logger *zap.Logger) *CardService {
	return &CardService{store: store, client: client, logger: logger}
}

// LinkCard links a card to a customer and stores them in the mongoDB
type SpendingLimit struct {
	Amount   int
	Interval string
}
type CardControls struct {
	AllowedChannels   []string
	BlockedChannels   []string
	AllowedMerchants  []string
	BlockedMerchants  []string
	AllowedCategories []string
	BlockedCategories []string
	SpendingLimits    []SpendingLimit
}

type CardMetadata struct {
	Name string
}

// Helper function to convert internal CardControls to *api.CardControls
func convertToApiCardControls(c CardControls) *api.CardControls {
	var spendingLimits []api.SpendingLimits
	for _, sl := range c.SpendingLimits {
		spendingLimits = append(spendingLimits, api.SpendingLimits{
			Amount:   sl.Amount,
			Interval: sl.Interval,
		})
	}
	return &api.CardControls{
		AllowedChannels:   c.AllowedChannels,
		BlockedChannels:   c.BlockedChannels,
		AllowedMerchants:  c.AllowedMerchants,
		BlockedMerchants:  c.BlockedMerchants,
		AllowedCategories: c.AllowedCategories,
		BlockedCategories: c.BlockedCategories,
		SpendingLimits:    spendingLimits,
	}
}

func convertToModelCardDetails(details api.CardDetails) models.CardDetails {
	return models.CardDetails{
		// Map each field from api.CardDetails to models.CardDetails here.
		// Example:
		// Field1: details.Field1,
		// Field2: details.Field2,
		// Add all necessary fields according to your struct definitions.
	}
}

// Helper function to convert api.CardControls to models.Controls
func convertToModelControls(apiControls api.CardControls) models.Controls {
	var spendingLimits []models.SpendingLimits
	for _, sl := range apiControls.SpendingLimits {
		spendingLimits = append(spendingLimits, models.SpendingLimits{
			Amount:   sl.Amount,
			Interval: sl.Interval,
		})
	}
	return models.Controls{
		AllowedChannels:   fmt.Sprintf("%v", apiControls.AllowedChannels),
		BlockedChannels:   fmt.Sprintf("%v", apiControls.BlockedChannels),
		AllowedMerchants:  fmt.Sprintf("%v", apiControls.AllowedMerchants),
		BlockedMerchants:  fmt.Sprintf("%v", apiControls.BlockedMerchants),
		AllowedCategories: fmt.Sprintf("%v", apiControls.AllowedCategories),
		BlockedCategories: fmt.Sprintf("%v", apiControls.BlockedCategories),
		SpendingLimits: func() models.SpendingLimits {
			if len(spendingLimits) > 0 {
				return spendingLimits[0]
			}
			return models.SpendingLimits{}
		}(),
	}
}

func (s *CardService) LinkCard(pan, customer, fundingSource, reference string, controls CardControls, metadata CardMetadata) (string, error) {
	s.logger.Info("Starting linking card",
		zap.String("pan", pan),
		zap.String("customer", customer),
		zap.String("fundingSource", fundingSource),
		zap.String("controls", fmt.Sprintf("%+v", controls)),
	)
	session, err := s.store.Client.StartSession()
	if err != nil {
		s.logger.Error("Failed to start MongoDB session", zap.Error(err))
		return "", err
	}
	defer session.EndSession(context.Background())
	ctx := context.Background()
	//Determine funding source
	// selectedFundingSource := fundingSource
	// if selectedFundingSource == "" {
	// 	//fetch primary sub-account

	// 		)
	// 		return "", "", "", "", "", "", "", fmt.Errorf("customer has not active sub-account")
	// 	}
	// 	if err != nil {
	// 		s.logger.Error("Failed to fetch sub-account",
	// 			zap.String("CustomerID", customer),
	// 			zap.Error(err),
	// 		)
	// 		return "", "", "", "", "", "", "", fmt.Errorf("failed to fetch sub accounts: %w", err)
	// 	}
	// }
	// Build the request to link the card
	req := &api.LinkCardRequest{
		Pan:           pan,
		Customer:      customer,
		FundingSource: fundingSource,
		Controls:      convertToApiCardControls(controls),
		Metadata: &api.CardMetadata{
			Name: metadata.Name,
		},
	}

	// Call the API to link the card
	resp, err := s.client.LinkCard(*req)
	if err != nil {
		s.logger.Error("Failed to link card via API", zap.Error(err))
		return "", err
	}

	card := models.Card{
		CardID:        resp.Data.ID,
		CustomerID:    resp.Data.Customer,
		FundingSource: resp.Data.FundingSource,
		Details:       convertToModelCardDetails(resp.Data.Details),
		Controls:      convertToModelControls(resp.Data.Controls),
		Type:          resp.Data.Type,
		Status:        resp.Data.Status,
		Program:       resp.Data.Program,
		Reference:     resp.Data.Reference,
		CreatedAt:     time.Now(),
	}

	_, err = s.store.Cards.InsertOne(ctx, card)
	if err != nil {
		s.logger.Error("Failed to store card in MongoDB", zap.Error(err))
		return "", err
	}
	s.logger.Info("Stored card in MongoDB", zap.String("cardID", resp.Data.ID))
	return resp.Data.ID, nil
}
