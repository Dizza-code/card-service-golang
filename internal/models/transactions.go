package models

type Transaction struct {
	ID         string `bson:"id"`
	CardID     string `bson:"cardId"`
	CustomerID string `bson:"customerId"`
	Type       string `bson:"type"` // Transaction type(e.g Authorization, deposits)
	Amount     int64  `bson:"amount"`
	Fees       int64  `bson:"fees"`
	Status     string `bson:"status"`
	CreatedAt  string `bson:"createdAt"`
}
