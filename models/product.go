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

type ProductCategory struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	ParentID primitive.ObjectID `bson:"parentId,omitempty" json:"parentId,omitempty"`
}

type ProductRating struct {
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Rating    float64            `bson:"rating" json:"rating"`
	Comment   string             `bson:"comment" json:"comment"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type Product struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description" json:"description"`
	Price       string               `bson:"price" json:"price"` // Price in ETH
	PriceUSD    float64              `bson:"priceUSD" json:"priceUSD"`
	Sizes       []ProductSize        `bson:"sizes" json:"sizes"`
	Colors      []string             `bson:"colors" json:"colors"`
	Images      []string             `bson:"images" json:"images"`
	Stock       map[ProductSize]int  `bson:"stock" json:"stock"`
	CreatedAt   time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time            `bson:"updatedAt" json:"updatedAt"`
	Categories  []primitive.ObjectID `bson:"categories" json:"categories"`
	Tags        []string             `bson:"tags" json:"tags"`
	Ratings     []ProductRating      `bson:"ratings" json:"ratings"`
	AvgRating   float64              `bson:"avgRating" json:"avgRating"`
}
