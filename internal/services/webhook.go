package services

import (
	"card-service/internal/api"
	"card-service/internal/models"
	"card-service/internal/store"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type WebhookService struct {
	store     *store.Store
	apiClient *api.Client
	logger    *zap.Logger
}

func NewWebhookService(store *store.Store, apiClient *api.Client, logger *zap.Logger) *WebhookService {
	return &WebhookService{
		store:     store,
		apiClient: apiClient,
		logger:    logger,
	}
}

func (s *WebhookService) HandleWebhook(ctx context.Context, event api.WebhookEvent) (api.AuthorizationResponse, error) {
	var response api.AuthorizationResponse

	switch event.Event {
	case "card.transaction.created":
		transaction, ok := event.Data.(api.TransactionEvent)
		if !ok {
			s.logger.Error("Invalid transaction event data", zap.Any("data", event.Data))
			return response, fmt.Errorf("invalid transaction event data")
		}
		return s.handleTransactionEvent(ctx, transaction)

	case "card.authorization.request":
		authRequest, ok := event.Data.(api.AuthorizationRequestEvent)
		if !ok {
			s.logger.Error("Invalid authorization request event data", zap.Any("data", event.Data))
			return response, fmt.Errorf("invalid authorization request event data")
		}
		return s.HandleAuthorizationRequest(ctx, authRequest)

	case "card.authorization.closed":
		authClosed, ok := event.Data.(api.AuthorizationClosedEvent)
		if !ok {
			s.logger.Error("Invalid authorization closed event data", zap.Any("data", event.Data))
			return response, fmt.Errorf("invalid authorization closed event data")
		}
		return s.HandleAuthorizationClosed(ctx, authClosed)

	default:
		s.logger.Error("Unknown event type", zap.String("event", event.Event))
		return response, fmt.Errorf("unknown event type: %s", event.Event)
	}

}

// handleTransactionEvent processes a transaction event and returns an authorization request event.
func (s *WebhookService) handleTransactionEvent(ctx context.Context, event api.TransactionEvent) (api.AuthorizationResponse, error) {
	if event.Type == "check" {
		s.logger.Info("processing balance check", zap.String("CardID", event.CardID))
		return api.AuthorizationResponse{Code: "success"}, nil
	}

	if event.Type == "capture" {
		//validate positive amount
		if event.Amount < 0 || event.Fees < 0 {
			s.logger.Error("invalid amount or fees",
				zap.Int64("amount", event.Amount),
				zap.Int64("fees", event.Fees),
			)
		}
		return api.AuthorizationResponse{Code: "error"}, fmt.Errorf("invalid ammount or fees")
	}

	//create transaction and store it in the database
	transaction := models.Transaction{
		ID:            event.ID,
		Authorization: event.Authorization,
		CardID:        event.CardID,
		CustomerID:    event.CustomerID,
		Amount:        event.Amount,
		Currency:      event.Currency,
		Type:          event.Type,
		Fees:          event.Fees,
		Channel:       event.Channel,
		Status:        "approved", // Assuming the transaction is approved
		NetworkData: models.NetworkData{
			CardAcceptorNameLocation: event.NetworkData.CardAcceptorNameLocation,
			Network:                  event.NetworkData.Network,
			Reference:                event.NetworkData.Reference,
			RRN:                      event.NetworkData.RRN,
			STAN:                     event.NetworkData.STAN,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	//store in the database
	_, err := s.store.InsertTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to store transaction in database",
			zap.String("transactionID", event.ID),
			zap.Error(err),
		)
		return api.AuthorizationResponse{Code: "error"}, fmt.Errorf("failed to store transaction: %w", err)
	}
	s.logger.Info("Stored transaction in database",
		zap.String("transactionID", event.ID),
		zap.String("CardID", event.CardID),
	)
	return api.AuthorizationResponse{Code: "success"}, nil
}

// handle AuthorizationRequestEvent processes an authorization request event and returns an authorization response.
func (s *WebhookService) HandleAuthorizationRequest(ctx context.Context, event api.AuthorizationRequestEvent) (api.AuthorizationResponse, error) {
	if event.Type != "pending" {
		s.logger.Error("Invalid authorization request",
			zap.String("stauts", event.Status),
			zap.String("cardID", event.CardID),
		)
		return api.AuthorizationResponse{Action: "declined", Code: "invalid transaction"}, fmt.Errorf("invalid status:%s", event.Status)
	}

	//fetch the card from the database
	var card models.Card
	err := s.store.Cards.FindOne(ctx, bson.M{"cardID": event.CardID}).Decode(&card)
	if err != nil {
		s.logger.Error("failed to fetch card",
			zap.String("CardID", event.CardID),
			zap.Error(err))
		return api.AuthorizationResponse{Action: "declined", Code: "account-not-found"}, fmt.Errorf("failed to fetch card: %w", err)
	}

	//validate card status
	if card.Status != "active" {
		s.logger.Warn("Card is not active",
			zap.String("CardID", event.CardID),
			zap.String("status", card.Status),
		)
		return api.AuthorizationResponse{Action: "declined", Code: "account-inactive"}, fmt.Errorf("card is not active")
	}

	//fetch the customer from the db
	var customer models.Customer
	err = s.store.Customers.FindOne(ctx, bson.M{"accountId": card.FundingSource}).Decode(&customer)
	if err != nil {
		s.logger.Error("Failed to fetch customer", zap.String("accountID", card.FundingSource), zap.Error(err))
		return api.AuthorizationResponse{Action: "decline", Code: "account-not-found"}, fmt.Errorf("failed to fetch customer: %w", err)
	}
	// fetch balance
	balance, err := s.apiClient.GetAccountBalance(card.FundingSource)
	if err != nil {
		s.logger.Error("Failed to fetch balance", zap.String("accountID", card.FundingSource), zap.Error(err))
		return api.AuthorizationResponse{Action: "decline", Code: "error"}, fmt.Errorf("failed to fetch balance: %w", err)
	}

	//validate balance(amount + fees)
	totalAmount := event.Amount + event.Fees
	if event.Type == "capture" && totalAmount > balance.Data.Available {
		s.logger.Warn("Insufficient balance",
			zap.String("cardID", event.CardID),
			zap.Int64("totalAmount", totalAmount),
			zap.Int64("availableBalance", balance.Data.Available),
		)
		return api.AuthorizationResponse{Action: "decline", Code: "insufficient-funds"}, fmt.Errorf("insufficient balance")
	}

	//validate controls
	if !s.isChannelAllowed(card.Controls, event.Channel) {
		s.logger.Warn("Channel not allowed",
			zap.String("cardID", event.CardID),
			zap.String("channel", event.Channel),
		)
		return api.AuthorizationResponse{Action: "decline", Code: "spending-control"}, fmt.Errorf("channel not allowed")
	}

	//check for duplicate transaction
	existing, err := s.store.GetTransaction(ctx, event.ID)
	if err == nil && existing != nil {
		s.logger.Warn("Duplicate transaction",
			zap.String("transactionID", event.ID),
			zap.String("cardID", event.CardID),
		)
		return api.AuthorizationResponse{Action: "approve", Code: "duplicate-transaction"}, fmt.Errorf("duplicate transaction")
	}
	s.logger.Info("Authorization approved",
		zap.String("cardID", event.CardID),
		zap.String("type", event.Type),
		zap.Int64("totalAmount", totalAmount),
	)

	//handle closed authorization

	return api.AuthorizationResponse{
		Action:         "approve",
		CardBalance:    balance.Data.Available,
		CardHolderName: customer.Name,
	}, nil
}

// isChannelAllowed checks if a channel is allowed for the card controls.
func (s *WebhookService) isChannelAllowed(controls api.CardControls, channel string) bool {
	for _, allowed := range controls.AllowedChannels {
		if allowed == channel {
			return true
		}
	}
	for _, blocked := range controls.BlockedChannels {
		if blocked == channel {
			return false
		}
	}
	return len(controls.AllowedChannels) == 0
}

// HandleAuthorizationClosed processes an authorization closed event and returns an authorization response.
func (s *WebhookService) HandleAuthorizationClosed(ctx context.Context, event api.AuthorizationClosedEvent) (api.AuthorizationResponse, error) {
	// fetch original transaction
	original, err := s.store.GetTransaction(ctx, event.ID)
	if err != nil || original == nil {
		s.logger.Error("Failed to fetch original transaction",
			zap.String("authorizationID", event.ID),
			zap.Error(err),
		)
		return api.AuthorizationResponse{Action: "decline", Code: "invalid-transaction"}, fmt.Errorf("failed to fetch original transaction: %w", err)
	}

	if event.Status == "approved" {
		update := bson.M{
			"$set": bson.M{
				"status":    event.Status,
				"amount":    event.Amount,
				"fees":      event.Fees,
				"updatedAt": time.Now().UTC().Format(time.RFC3339),
			},
		}
		_, err = s.store.Transactions.UpdateOne(ctx, bson.M{"id": event.ID}, update)
		if err != nil {
			s.logger.Error("Failed to update transaction for approval",
				zap.String("authorizationID", event.ID),
				zap.Error(err),
			)
			return api.AuthorizationResponse{Action: "decline", Code: "error"}, fmt.Errorf("failed to update transaction: %w", err)
		}
		s.logger.Info("Authorization closed approved",
			zap.String("cardID", event.CardID),
			zap.Int64("totalAmount", event.Amount+event.Fees),
		)
		return api.AuthorizationResponse{Action: "approve"}, nil
	}
	return api.AuthorizationResponse{Action: "decline"}, nil
}
