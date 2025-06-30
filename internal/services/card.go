package services

import (
	"card-service/internal/api"
	"card-service/internal/models"
	"card-service/internal/store"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

type CardMetadata struct {
	Name string
}

func (s *CardService) LinkCard(
	ctx context.Context,
	pan,
	customer,
	fundingSource,
	reference string,
	controls *api.CardControls, metadata *api.CardMetadata) (string,
	string, string, string, string, string, string, string, error) {
	s.logger.Info("Starting linking card",
		zap.String("pan", pan),
		zap.String("customer", customer),
		zap.String("fundingSource", fundingSource),
		zap.String("controls", fmt.Sprintf("%+v", controls)),
	)
	session, err := s.store.Client.StartSession()
	if err != nil {
		s.logger.Error("Failed to start MongoDB session", zap.Error(err))
		return "", "", "", "", "", "", "", "", err
	}
	defer session.EndSession(context.Background())
	// ctx := context.Background()
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
	req := api.LinkCardRequest{
		Pan:           pan,
		Customer:      customer,
		FundingSource: fundingSource,
		Controls:      controls,
		Metadata:      metadata,
	}

	// Call the API to link the card
	resp, err := s.client.LinkCard(req)
	if err != nil {
		s.logger.Error("Failed to link card via API", zap.Error(err))
		return "", "", "", "", "", "", "", "", err
	}

	card := models.Card{
		CardID:         resp.Data.ID,
		CustomerID:     resp.Data.Customer,
		FundingSource:  resp.Data.FundingSource,
		Last4:          resp.Data.Details.Last4,
		Expiry:         resp.Data.Details.Expiry,
		CardHolderName: resp.Data.Details.CardHolderName,
		Controls:       resp.Data.Controls,
		Type:           resp.Data.Type,
		Status:         resp.Data.Status,
		Program:        resp.Data.Program,
		Reference:      resp.Data.Reference,
		Metadata:       resp.Data.Metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = s.store.Cards.InsertOne(ctx, card)
	if err != nil {
		s.logger.Error("Failed to store card in MongoDB", zap.Error(err))
		return "", "", "", "", "", "", "", "", err
	}
	s.logger.Info("Stored card in MongoDB", zap.String("cardID", resp.Data.ID))
	return resp.Data.ID,
		resp.Data.Details.Last4,
		resp.Data.Details.Expiry,
		resp.Data.Details.CardHolderName,
		resp.Data.Type,
		resp.Data.Status,
		resp.Data.Program,
		resp.Data.Currency,
		nil
}

func (s *CardService) ActivateCard(ctx context.Context, cvv, pin string, cardID string) (string, error) {
	s.logger.Info("Activating card",
		zap.String("cvv", cvv),
		zap.String("pin", pin),
	)

	// Check if card exists and is inactive
	var card models.Card
	err := s.store.Cards.FindOne(ctx, bson.M{"cardId": cardID}).Decode(&card)
	if err == mongo.ErrNoDocuments {
		s.logger.Error("Card not found", zap.String("cardID", cardID))
		return "", fmt.Errorf("card not found")
	}
	if err != nil {
		s.logger.Error("Failed to fetch card", zap.String("cardID", cardID), zap.Error(err))
		return "", fmt.Errorf("failed to fetch card: %w", err)
	}

	if card.Status == "active" {
		s.logger.Warn("Card already activated", zap.String("cardID", cardID))
		return "", fmt.Errorf("card already activated")
	}

	req := api.ActivateCardRequest{
		Cvv: cvv,
		Pin: pin,
	}

	resp, err := s.client.ActivateCard(cardID, req)
	if err != nil {
		s.logger.Error("Failed to activate card via API", zap.Error(err))
		return "", err
	}

	//update card status in MongoDB
	update := bson.M{
		"$set": bson.M{
			"status":    "active",
			"updatedAt": time.Now(),
		},
	}
	_, err = s.store.Cards.UpdateOne(ctx, bson.M{"cardId": cardID}, update)
	if err != nil {
		s.logger.Error("Failed to update card status in MongoDB",
			zap.String("cardID", cardID), zap.Error(err))
		return "", fmt.Errorf("failed to update card status: %w", err)
	}

	s.logger.Info("Card activated successfully",
		zap.String("code", resp.Code),
		zap.String("message", resp.Message),
	)

	return resp.Code, nil
}
