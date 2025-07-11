package handlers

import (
	"card-service/internal/api"
	"card-service/internal/services"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	webhookService *services.WebhookService
	logger         *zap.Logger
	signingKey     string
}

func NewWebhookHandler(webhookService *services.WebhookService, logger *zap.Logger, signingKey string) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		logger:         logger,
		signingKey:     signingKey,
	}
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	//verify signature
	body, err := c.GetRawData()
	if err != nil {
		h.logger.Error("failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request"})
		return
	}
	signature := c.GetHeader("Allawee-Signature")
	hash := hmac.New(sha512.New, []byte(h.signingKey))
	hash.Write(body)
	expected := hex.EncodeToString(hash.Sum(nil))
	if signature != expected {
		h.logger.Error("invalid signature",
			zap.String("received", signature),
			zap.String("expected", expected),
		)
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid-signature"})
		return
	}

	var event api.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		h.logger.Error("Failed to bind webhook event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch event.Event {
	case "card.transaction.created":
		var transaction api.TransactionEvent
		if err := json.Unmarshal(body, &event); err != nil {
			h.logger.Error("failed to bind transaction event", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		event.Data = transaction

	case "card.authorization.request":
		var authRequest api.AuthorizationRequestEvent
		if err := json.Unmarshal(body, &err); err != nil {
			h.logger.Error("failed to bind authorization event", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		event.Data = authRequest

	case "card.authorization.closed":
		var authClosed api.AuthorizationClosedEvent
		if err := json.Unmarshal(body, &err); err != nil {
			h.logger.Error("failed to bind authorization closed", zap.Error(err))
			return
		}
		event.Data = authClosed

	default:
		h.logger.Error("unknown event type", zap.String("event", event.Event))
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown event type: %s", event.Event)})
		return
	}
	h.logger.Info("Received webhook event", zap.String("event", event.Event))

	response, err := h.webhookService.HandleWebhook(c.Request.Context(), event)
	if err != nil {
		h.logger.Error("failed to handle event webhook", zap.String("event", event.Event), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if response.Action != "" {
		c.JSON(http.StatusOK, response)
	} else {
		c.JSON(http.StatusOK, gin.H{"code": "success"})
		return
	}

	// Log the response sent to Allawee
	responseBytes, _ := json.Marshal(response)
	h.logger.Info("Webhook response sent",
		zap.String("event", event.Event),
		zap.String("response", string(responseBytes)),
	)

	c.JSON(200, response)
}
