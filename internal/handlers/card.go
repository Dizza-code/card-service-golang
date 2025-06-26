package handlers

import (
	"card-service/internal/api"
	"card-service/internal/services"

	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CardHandler struct {
	cardService *services.CardService
	logger      *zap.Logger
}

//NewCardHandler creates a new card handler

func NewCardHandler(cardService *services.CardService, logger *zap.Logger) *CardHandler {
	return &CardHandler{
		cardService: cardService,
		logger:      logger,
	}
}

type SpendingLimitRequest struct {
	Amount   int    `json:"amount"`
	Interval string `json:"interval"`
}

type CardControlsRequest struct {
	AllowedChannels   []string               `json:"allowedChannels"`
	BlockedChannels   []string               `json:"blockedChannels"`
	AllowedMerchants  []string               `json:"allowedMerchants"`
	BlockedMerchants  []string               `json:"blockedMerchants"`
	AllowedCategories []string               `json:"allowedCategories"`
	BlockedCategories []string               `json:"blockedCategories"`
	SpendingLimits    []SpendingLimitRequest `json:"spendingLimits"`
	// Metadata         string                 `bson:"Metadata"`
}

type CardMetadataRequest struct {
	Name string `json:"name"`
}

type LinkCardRequest struct {
	Pan           string               `json:"pan" binding:"required"`
	Customer      string               `json:"customerId"`
	FundingSource string               `json:"fundingSource"`
	Reference     string               `json:"reference"`
	Controls      *CardControlsRequest `json:"controls"`
	Metadata      *CardMetadataRequest `json:"metadata"`
}
type CardDetails struct {
	Last4          string `json:"last4" bson:"last4"`
	Expiry         string `json:"exp" bson:"exp"`
	CardHolderName string `json:"cardHolderName" bson:"cardHolderName"`
}
type LinkCardRequestResponse struct {
	CardID        string           `json:"cardId"`
	CustomerID    string           `json:"customerId"`
	FundingSource string           `json:"fundingSource"`
	Program       string           `json:"program"`
	Currency      string           `json:"currency"`
	Type          string           `json:"type"`
	Status        string           `json:"status"`
	Details       CardDetails      `json:"details"`
	Controls      api.CardControls `json:"controls"`
	Metadata      api.CardMetadata `json:"metadata"`
}

func convertSpendingLimits(limits []SpendingLimitRequest) []api.SpendingLimit {
	result := make([]api.SpendingLimit, len(limits))
	for i, l := range limits {
		result[i] = api.SpendingLimit{
			Amount:   l.Amount,
			Interval: l.Interval,
		}
	}
	return result
}

func (h *CardHandler) LinkCard(c *gin.Context) {
	var req LinkCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var controls *api.CardControls
	if req.Controls != nil {
		controls = &api.CardControls{
			AllowedChannels:   req.Controls.AllowedChannels,
			BlockedChannels:   req.Controls.BlockedChannels,
			AllowedMerchants:  req.Controls.AllowedMerchants,
			BlockedMerchants:  req.Controls.BlockedMerchants,
			AllowedCategories: req.Controls.AllowedCategories,
			BlockedCategories: req.Controls.BlockedCategories,
			SpendingLimits:    convertSpendingLimits(req.Controls.SpendingLimits),
		}
	}

	var metadata *api.CardMetadata
	if req.Metadata != nil {
		metadata = &api.CardMetadata{
			Name: req.Metadata.Name,
		}
	}

	cardID, last4, expiry, cardHolderName, cardType, status, program, currency, err := h.cardService.LinkCard(
		c.Request.Context(),
		req.Pan,
		req.Customer,
		req.FundingSource,
		req.Reference,
		controls,
		metadata,
	)
	if err != nil {
		h.logger.Error("Failed to link card", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := LinkCardRequestResponse{
		CardID:        cardID,
		CustomerID:    req.Customer,
		FundingSource: req.FundingSource,
		Program:       program,
		Currency:      currency,
		Status:        status,
		Type:          cardType,
		Details: CardDetails{
			Last4:          last4,
			Expiry:         expiry,
			CardHolderName: cardHolderName,
		},
		Controls: api.CardControls{},
	}
	if req.Controls != nil {
		response.Controls = *controls
	}
	c.JSON(http.StatusCreated, response)
}
