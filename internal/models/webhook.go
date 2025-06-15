package models

import "time"

// AllaweeEventBody represents the webhook payload from the card issuing service.
type AllaweeEventBody struct {
	Event    string                   `json:"event"`    // Event type (e.g., card.authorization.request)
	Data     AllaweeAuthorizationData `json:"data"`     // Authorization details
	Metadata AllaweeEventBodyMetadata `json:"metadata"` // Event metadata
}

// AllaweeAuthorizationData contains details of an authorization request.
type AllaweeAuthorizationData struct {
	Card          string                          `json:"card"`          // Card ID
	Type          string                          `json:"type"`          // Authorization type (check, capture)
	Id            string                          `json:"id"`            // Authorization ID
	Amount        int64                           `json:"amount"`        // Transaction amount
	Fees          int64                           `json:"fees"`          // Transaction fees
	Channel       string                          `json:"channel"`       // Transaction channel
	Reserved      bool                            `json:"reserved"`      // Whether funds are reserved
	NetworkData   AllaweeEventBodyDataNetworkData `json:"networkData"`   // Network-specific data
	CreatedAt     time.Time                       `json:"createdAt"`     // Timestamp of authorization
	Status        string                          `json:"status"`        // Authorization status
	DeclineReason string                          `json:"declineReason"` // Reason for decline (if any)
	DecisionType  string                          `json:"decisionType"`  // Decision type
	Currency      string                          `json:"currency"`      // Currency code
}

// AllaweeEventBodyDataNetworkData contains network-specific transaction details.
type AllaweeEventBodyDataNetworkData struct {
	Rrn                      string `json:"rrn"`                      // Retrieval Reference Number
	Stan                     string `json:"stan"`                     // System Trace Audit Number
	Network                  string `json:"network"`                  // Card network (e.g., Visa, Mastercard)
	TxnReference             string `json:"txnReference"`             // Transaction reference
	CardAcceptorNameLocation string `json:"cardAcceptorNameLocation"` // Merchant details
}

// AllaweeEventBodyMetadata contains metadata about the webhook event.
type AllaweeEventBodyMetadata struct {
	CreatedAt time.Time `json:"createdAt"` // Timestamp of event
	Event     string    `json:"event"`     // Event type (redundant with top-level event)
}

// paymentData for payment related webhook events.
type PaymentData struct {
	ID               string    `json:"id"`               // Unique identifier for the payment
	CustomerID       string    `json:"customerId"`       // ID of the customer associated with the payment
	VirtualAccountID string    `json:"virtualAccountId"` // ID of the virtual account used for the payment
	Amount           int64     `json:"amount"`           // Amount of the payment
	Currency         string    `json:"currency"`         // Currency of the payment
	Status           string    `json:"status"`           // Status of the payment (e.g., pending, completed, failed)
	CreatedAt        time.Time `json:"createdAt"`        // Timestamp when the payment was created
}
