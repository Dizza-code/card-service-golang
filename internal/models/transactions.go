package models

type NetworkData struct {
	CardAcceptorNameLocation string `bson:"cardAcceptorNameLocation"`
	TerminalID               string `bson:"terminalId"`
	Network                  string `bson:"network"`   // Network used for the transaction (e.g Visa, Mastercard)
	Reference                string `bson:"reference"` // Reference number for the transaction
	RRN                      string `bson:"rrn"`       // Retrieval Reference Number
	STAN                     string `bson:"stan"`      // System Trace Audit Number
}
type Transaction struct {
	ID            string      `bson:"id"`
	Authorization string      `bson:"authorizationId"`
	CardID        string      `bson:"cardId"`
	CustomerID    string      `bson:"customerId"`
	Amount        int64       `bson:"amount"`
	Currency      string      `bson:"currency"`
	Type          string      `bson:"type"` // Transaction type(e.g Authorization, deposits)
	Fees          int64       `bson:"fees"`
	Channel       string      `bson:"channel"`     // Channel through which the transaction was made (e.g POS, ATM, Online)
	NetworkData   NetworkData `bson:"networkData"` // Network data related to the transaction
	Status        string      `bson:"status"`

	CreatedAt string `bson:"createdAt"`
	UpdatedAt string `bson:"updatedAt,omitempty"` // Optional field for the last update time
}

type AuthorizationRequestEvent struct {
	CardID        string      `bson:"cardId"`
	Authorization string      `bson:"authorizationId"`
	CustomerID    string      `bson:"customerId"`
	Amount        int64       `bson:"amount"`
	Currency      string      `bson:"currency"`
	Type          string      `bson:"type"` // Transaction type (e.g., Authorization,
	Fees          int64       `bson:"fees"`
	Channel       string      `bson:"channel"` // Channel through which the transaction was made (e.g., POS, ATM, Online)
	Status        string      `bson:"status"`
	NetworkData   NetworkData `bson:"networkData"` // Network data related to the
	CreatedAt     string      `bson:"createdAt"`
	UpdatedAt     string      `bson:"updatedAt,omitempty"` // Optional field for the last update time
}
