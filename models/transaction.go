package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	OrderID         uint64             `bson:"orderId"`
	CustomerAddress string             `bson:"customerAddress"`
	Amount          string             `bson:"amount"`
	Timestamp       time.Time          `bson:"timestamp"`
	Status          string             `bson:"status"`
}
