package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "PENDING"
	OrderStatusPaid    OrderStatus = "PAID"
	OrderStatusFailed  OrderStatus = "FAILED"
)

type OrderItem struct {
	ProductID primitive.ObjectID `bson:"productId" json:"productId"`
	Size      ProductSize        `bson:"size" json:"size"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Price     string             `bson:"price" json:"price"`
}

type Order struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`
	Items         []OrderItem        `bson:"items" json:"items"`
	TotalPrice    string             `bson:"totalPrice" json:"totalPrice"`
	Status        OrderStatus        `bson:"status" json:"status"`
	WalletAddress string             `bson:"walletAddress" json:"walletAddress"`
	TxHash        string             `bson:"txHash,omitempty" json:"txHash,omitempty"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}
