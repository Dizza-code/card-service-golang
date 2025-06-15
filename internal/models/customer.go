package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Customer struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID string             `bson:"customerId"`
	Name       string             `bson:"name"`
	Email      string             `bson:"email"`
	// Phone            string    `bson:"phone"`
	// Balance      int64     `bson:"balance"`
	AccountID string    `bson:"accountId"`
	CreatedAt time.Time `bson:"createdAt"`
}
