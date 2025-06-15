package models

import "time"

type Card struct {
	ID         string    `bson:"id"`
	CustomerID string    `bson:"customerId"`
	Active     bool      `bson:"active"`
	CreatedAt  time.Time `bson:"createdAt"`
}
