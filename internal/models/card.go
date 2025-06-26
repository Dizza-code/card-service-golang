package models

import (
	"card-service/internal/api"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SpendingLimits struct {
	Amount   int    `bson:"amount"`
	Interval string `bson:"interval"`
}

type Controls struct {
	AllowedChannels   []string         `bson:"allowedChannels"`
	BlockedChannels   []string         `bson:"blockedChannels"`
	AllowedMerchants  []string         `bson:"allowedMerchants"`
	BlockedMerchants  []string         `bson:"blockedMerchants"`
	AllowedCategories []string         `bson:"allowedCategories"`
	BlockedCategories []string         `bson:"blockedCategories"`
	SpendingLimits    []SpendingLimits `bson:"spendingLimits"`
	// Metadata         string   `bson:"Metadata"`
}

//	type CardDetails struct {
//		Last4          string `bson:"last4"`
//		Expiry         string `bson:"expiry"`
//		CardholderName string `bson:"cardHolderName"`
//	}
type CardMetadata struct {
	Name string `bson:"name"`
}
type Card struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	CardID         string             `bson:"cardId"`
	Reference      string             `bson:"reference"`
	CustomerID     string             `bson:"customerId"`
	FundingSource  string             `bson:"fundingSource"`
	Last4          string             `bson:"last4"`
	Expiry         string             `bson:"expiry"`
	CardHolderName string             `bson:"cardHolderName"`
	Type           string             `bson:"type"`
	Status         string             `bson:"status"`
	Program        string             `bson:"program"`
	Controls       api.CardControls   `bson:"controls"`
	Metadata       api.CardMetadata   `bson:"metadata"`
	CreatedAt      time.Time          `bson:"createdAt"`
}
