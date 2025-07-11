package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DepositChannel struct {
	AccountName   string `bson:"accountName" json:"accountName"`
	AccountNumber string `bson:"accountNumber" json:"accountNumber"`
	BankName      string `bson:"bankName" json:"bankName"`
	BankCode      string `bson:"bankCode" json:"bankCode"`
	Type          string `bson:"type" json:"type"`
}

type Account struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID      string             `bson:"customerId"`
	AccountID       string             `bson:"accountId"`
	Name            string             `bson:"name"`
	DepositChannels []DepositChannel   `bson:"depositChannels"`
	Status          string             `bson:"status"`
	CreatedAt       time.Time          `bson:"createdAt"`
}
