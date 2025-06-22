package handlers

import (
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
	AllowedChannels   []string               `bson:"allowedChannels"`
	BlockedChannels   []string               `bson:"blockedChannels"`
	AllowedMerchants  []string               `bson:"allowedMerchants"`
	BlockedMerchants  []string               `bson:"blockedMerchants"`
	AllowedCategories []string               `bson:"allowedCategories"`
	BlockedCategories []string               `bson:"blockedCategories"`
	SpendingLimits    []SpendingLimitRequest `bson:"spendingLimits"`
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
	Exp            string `json:"exp" bson:"exp"`
	CardHolderName string `json:"cardHolderName" bson:"cardHolderName"`
}
type LinkCardRequestResponse struct {
	CardID         string                `json:"cardId"`
	CustomerID     string                `json:"customerId"`
	FundingSource  string                `json:"fundingSource"`
	CardProgram    string                `json:"cardProgram"`
	Currency       string                `json:"currency"`
	Status         string                `json:"status"`
	Details        CardDetails           `json:"details"`
	Controls       services.CardControls `json:"controls"`
	CardHolderName string                `json:"cardHolderName"`
	Type           string                `json:"type"`
}

func convertSpendingLimits(limits []SpendingLimitRequest) []services.SpendingLimit {
	result := make([]services.SpendingLimit, len(limits))
	for i, l := range limits {
		result[i] = services.SpendingLimit{
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

	var controls services.CardControls
	if req.Controls != nil {
		controls = services.CardControls{
			AllowedChannels:   req.Controls.AllowedChannels,
			BlockedChannels:   req.Controls.BlockedChannels,
			AllowedMerchants:  req.Controls.AllowedMerchants,
			BlockedMerchants:  req.Controls.BlockedMerchants,
			AllowedCategories: req.Controls.AllowedCategories,
			BlockedCategories: req.Controls.BlockedCategories,
			SpendingLimits:    convertSpendingLimits(req.Controls.SpendingLimits),
		}
	}

	var metadata *services.CardMetadata
	if req.Metadata != nil {
		metadata = &services.CardMetadata{
			Name: req.Metadata.Name,
		}
	}

	var cardID string
	var err error
	if metadata != nil {
		cardID, err = h.cardService.LinkCard(
			req.Pan,
			req.Customer,
			req.FundingSource,
			req.Reference,
			controls,
			*metadata,
		)
	} else {
		cardID, err = h.cardService.LinkCard(
			req.Pan,
			req.Customer,
			req.FundingSource,
			req.Reference,
			controls,
			services.CardMetadata{},
		)
	}
	if err != nil {
		h.logger.Error("Failed to link card", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := LinkCardRequestResponse{
		CardID:        cardID,
		CustomerID:    req.Customer,
		FundingSource: req.FundingSource,
		Controls:      controls,
		Details:       CardDetails{},
		Status:        "",
	}
	c.JSON(http.StatusCreated, response)
}
