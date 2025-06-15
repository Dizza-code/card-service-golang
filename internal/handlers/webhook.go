package handlers

// import (
// 	"bytes"
// 	"card-service/internal/services"
// 	"crypto/hmac"
// 	"crypto/sha512"
// 	"encoding/hex"
// 	"encoding/json"
// 	"io/ioutil"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// type WebhookHandler struct {
// 	signingKey  string
// 	authService *services.AuthorizationService
// }

// func NewWebhookHandler(signingKey string, authService *services.AuthorizationService) *WebhookHandler {
// 	return &WebhookHandler{signingKey: signingKey, authService: authService}
// }

// type AllaweeEventBody struct {
// 	Event    string                   `json:"event"`
// 	Data     AllaweeAuthorizationData `json:"data"`
// 	Metadata AllaweeEventBodyMetadata `json:"metadata"`
// }

// type AllaweeAuthorizationData struct {
// 	Card          string                          `json:"card"`
// 	Type          string                          `json:"type"`
// 	Id            string                          `json:"id"`
// 	Amount        int64                           `json:"amount"`
// 	Fees          int64                           `json:"fees"`
// 	Channel       string                          `json:"channel"`
// 	Reserved      bool                            `json:"reserved"`
// 	NetworkData   AllaweeEventBodyDataNetworkData `json:"networkData"`
// 	CreatedAt     time.Time                       `json:"createdAt"`
// 	Status        string                          `json:"status"`
// 	DeclineReason string                          `json:"declineReason"`
// 	DecisionType  string                          `json:"decisionType"`
// 	Currency      string                          `json:"currency"`
// }

// type AllaweeEventBodyDataNetworkData struct {
// 	Rrn                      string `json:"rrn"`
// 	Stan                     string `json:"stan"`
// 	Network                  string `json:"network"`
// 	TxnReference             string `json:"txnReference"`
// 	CardAcceptorNameLocation string `json:"cardAcceptorNameLocation"`
// }

// type AllaweeEventBodyMetadata struct {
// 	CreatedAt time.Time `json:"createdAt"`
// 	Event     string    `json:"event"`
// }

// func (h *WebhookHandler) Handle(c *gin.Context) {
// 	body := h.verifyRequest(c)
// 	if body == nil {
// 		return
// 	}

// 	switch body.Event {
// 	case "card.authorization.request":
// 		response := h.authService.ProcessRequest(body)
// 		c.JSON(200, response)
// 	case "card.authorization.closed":
// 		response := h.authService.ProcessClosed(body)
// 		c.JSON(200, response)
// 	case "card.authorization.update":
// 		response := h.authService.ProcessUpdate(body)
// 		c.JSON(200, response)
// 	case "card.transaction.created":
// 		c.JSON(200, gin.H{"code": "success"})
// 	default:
// 		c.JSON(400, gin.H{"error": "Invalid Request"})
// 	}
// }

// func (h *WebhookHandler) verifyRequest(c *gin.Context) *AllaweeEventBody {
// 	signature := c.GetHeader("Allawee-Signature")
// 	reqBody, _ := ioutil.ReadAll(c.Request.Body)
// 	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Restore body for binding

// 	hash := hmacHashHex(h.signingKey, string(reqBody))
// 	if hash != signature {
// 		c.JSON(400, gin.H{"error": "Invalid Signature"})
// 		return nil
// 	}

// 	var body AllaweeEventBody
// 	err := json.Unmarshal(reqBody, &body)
// 	if err != nil {
// 		c.JSON(400, gin.H{"error": "Invalid Request"})
// 		return nil
// 	}

// 	return &body
// }

// func hmacHashHex(key, secret string) string {
// 	h := hmac.New(sha512.New, []byte(key))
// 	h.Write([]byte(secret))
// 	return hex.EncodeToString(h.Sum(nil))
// }
