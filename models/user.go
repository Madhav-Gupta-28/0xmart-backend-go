package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type       string             `bson:"type" json:"type"` // shipping/billing
	Street     string             `bson:"street" json:"street"`
	City       string             `bson:"city" json:"city"`
	State      string             `bson:"state" json:"state"`
	Country    string             `bson:"country" json:"country"`
	PostalCode string             `bson:"postalCode" json:"postalCode"`
	IsDefault  bool               `bson:"isDefault" json:"isDefault"`
}

type User struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	Name          string                 `bson:"name" json:"name"`
	Email         string                 `bson:"email" json:"email"`
	Password      string                 `bson:"password,omitempty" json:"-"` // "-" means don't include in JSON
	EmailVerified bool                   `bson:"emailVerified" json:"emailVerified"`
	Image         string                 `bson:"image,omitempty" json:"image,omitempty"`
	Provider      string                 `bson:"provider" json:"provider"` // "credentials", "google", etc.
	ProviderId    string                 `bson:"providerId,omitempty" json:"providerId,omitempty"`
	PhoneNumber   string                 `bson:"phoneNumber,omitempty" json:"phoneNumber,omitempty"`
	Addresses     []Address              `bson:"addresses" json:"addresses"`
	Preferences   map[string]interface{} `bson:"preferences" json:"preferences"`
	CreatedAt     time.Time              `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time              `bson:"updatedAt" json:"updatedAt"`
}
