package api

type CardControls struct {
	AllowedChannels   []string        `json:"allowedChannels" bson:"allowedChannels"`
	BlockedChannels   []string        `json:"blockedChannels" bson:"blockedChannels"`
	AllowedMerchants  []string        `json:"allowedMerchants" bson:"allowedMerchants"`
	BlockedMerchants  []string        `json:"blockedMerchants" bson:"blockedMerchants"`
	AllowedCategories []string        `json:"allowedCategories" bson:"allowedCategories"`
	BlockedCategories []string        `json:"blockedCategories" bson:"blockedCategories"`
	SpendingLimits    []SpendingLimit `json:"spendingLimits" bson:"spendingLimits"`
}

type SpendingLimit struct {
	Amount   int    `json:"amount"`
	Interval string `json:"interval"` // e.g., "daily", "weekly", "monthly"
}

type CardMetadata struct {
	Name string `json:"name,omitempty" bson:"name"`
}

type CardDetails struct {
	Last4          string `json:"last4" bson:"last4"`
	Expiry         string `json:"expiry" bson:"expiry"`
	CardHolderName string `json:"cardHolderName" bson:"cardHolderName"`
}
type WebhookEvent struct {
	Event    string      `json:"event"`
	Data     interface{} `json:"data"`
	Metadata struct {
		SentAt string `json:"sentAt"`
		Event  string `json:"event"`
	} `json:"metadata"`
}
type NetworkData struct {
	CardAcceptorNameLocation string `json:"cardAcceptorNameLocation" bson:"cardAcceptorNameLocation"`
	TerminalID               string `json:"terminalId" bson:"terminalId"`
	Network                  string `json:"network" bson:"network"`     // Network used for the transaction (e.g., "Visa", "Mastercard")
	Reference                string `json:"reference" bson:"reference"` // Reference number for the transaction
	RRN                      string `json:"rrn" bson:"rrn"`             // Retrieval Reference Number
	STAN                     string `json:"stan" bson:"stan"`           // System Trace Audit Number
}
type TransactionEvent struct {
	ID            string      `json:"id"`
	Authorization string      `json:"authorization"`
	CardID        string      `json:"card"` // Maps to "card"
	CustomerID    string      `json:"customer"`
	Amount        int64       `json:"amount"`
	Currency      string      `json:"currency"`
	Type          string      `json:"type"` // "check" or "capture"
	Fees          int64       `json:"fees"`
	Channel       string      `json:"channel"`
	NetworkData   NetworkData `json:"networkData"`
	CreatedAt     string      `json:"createdAt"`
}
type AuthorizationRequestEvent struct {
	ID            string      `json:"id" bson:"id"`
	CardID        string      `json:"cardId" bson:"cardId"`
	Authorization string      `json:"authorizationId" bson:"authorizationId"`
	CustomerID    string      `json:"customerId" bson:"customerId"`
	Amount        int64       `json:"amount" bson:"amount"`
	Currency      string      `json:"currency" bson:"currency"`
	Type          string      `json:"type" bson:"type"` // Transaction type (e.g.,	"Authorization", "Deposit")
	Fees          int64       `json:"fees" bson:"fees"`
	Channel       string      `json:"channel" bson:"channel"` // Channel through which the transaction was made (e.g., "POS", "ATM", "Online")
	Status        string      `json:"status" bson:"status"`
	NetworkData   NetworkData `json:"networkData" bson:"networkData"` // Network data related to the transaction
	CreatedAt     string      `json:"createdAt" bson:"createdAt"`
	UpdatedAt     string      `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"` // Optional field for the last update time
}
type AuthorizationUpdateEvent struct {
	ID           string      `json:"id"`
	CardID       string      `json:"card"`
	Amount       int64       `json:"amount"`
	Currency     string      `json:"currency"`
	Type         string      `json:"type"`
	Fees         int64       `json:"fees"`
	Channel      string      `json:"channel"`
	NetworkData  NetworkData `json:"networkData"`
	CreatedAt    string      `json:"createdAt"`
	DecisionType string      `json:"decisionType"`
	Status       string      `json:"status"` // "pending" or "reversed"
}
type AuthorizationClosedEvent struct {
	ID           string      `json:"id"`
	CardID       string      `json:"card"`
	Amount       int64       `json:"amount"`
	Currency     string      `json:"currency"`
	Type         string      `json:"type"`
	Fees         int64       `json:"fees"`
	Channel      string      `json:"channel"`
	NetworkData  NetworkData `json:"networkData"`
	CreatedAt    string      `json:"createdAt"`
	DecisionType string      `json:"decisionType"`
	Status       string      `json:"status"` // "approved" or "declined"
}

type AuthorizationResponse struct {
	Action         string                 `json:"action"` // "approve" or "decline"
	Code           string                 `json:"code,omitempty"`
	CardBalance    int64                  `json:"cardBalance,omitempty"`
	CardHolderName string                 `json:"cardHolderName,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}
type GetAccountBalanceResponse struct {
	Code string `json:"code"`
	Data struct {
		Available       int64  `json:"available"`
		AvailableChange int64  `json:"availableChange"`
		Currency        string `json:"currency"`
		Mode            string `json:"mode"`
		CreatedAt       string `json:"createdAt"`
		Source          string `json:"source"`
		ID              string `json:"id"`
	} `json:"data"`
}
