package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductSize string

const (
	SizeS  ProductSize = "S"
	SizeM  ProductSize = "M"
	SizeL  ProductSize = "L"
	SizeXL ProductSize = "XL"
)

type Product struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Name        string              `bson:"name" json:"name"`
	Description string              `bson:"description" json:"description"`
	Price       string              `bson:"price" json:"price"` // Price in ETH
	PriceUSD    float64             `bson:"priceUSD" json:"priceUSD"`
	Sizes       []ProductSize       `bson:"sizes" json:"sizes"`
	Colors      []string            `bson:"colors" json:"colors"`
	Images      []string            `bson:"images" json:"images"`
	Stock       map[ProductSize]int `bson:"stock" json:"stock"`
	CreatedAt   time.Time           `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time           `bson:"updatedAt" json:"updatedAt"`
}
