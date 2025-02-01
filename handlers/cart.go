package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/models"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddToCart(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)

	var req struct {
		ProductID string `json:"productId"`
		Size      string `json:"size"`
		Quantity  int    `json:"quantity"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	collection := database.DB.Collection("carts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{
			"items": bson.M{
				"productId": productID,
				"size":      req.Size,
				"quantity":  req.Quantity,
			},
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	result := collection.FindOneAndUpdate(
		ctx,
		bson.M{"userId": userID},
		update,
		options.FindOneAndUpdate().SetUpsert(true),
	)

	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, result.Err().Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Item added to cart"})
}

// GetCart retrieves the user's cart
func GetCart(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)
	collection := database.DB.Collection("carts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var cart models.Cart
	err := collection.FindOne(ctx, bson.M{"userId": userID}).Decode(&cart)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Cart not found"})
	}

	return c.JSON(http.StatusOK, cart)
}

// RemoveFromCart removes an item from the cart
func RemoveFromCart(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)
	productID, err := primitive.ObjectIDFromHex(c.Param("productId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	collection := database.DB.Collection("carts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{"productId": productID},
		},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"userId": userID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if result.ModifiedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Item not found in cart"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Item removed from cart"})
}

// UpdateCartItemQuantity updates the quantity of an item in the cart
func UpdateCartItemQuantity(c echo.Context) error {
	type UpdateQuantityRequest struct {
		ProductID string `json:"productId"`
		Size      string `json:"size"`
		Quantity  int    `json:"quantity"`
	}

	var req UpdateQuantityRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	userID := c.Get("userID").(primitive.ObjectID)
	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	collection := database.DB.Collection("carts")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"items.$[elem].quantity": req.Quantity,
			"updatedAt":              time.Now(),
		},
	}

	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.productId": productID,
				"elem.size":      req.Size,
			},
		},
	}

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"userId": userID},
		update,
		options.Update().SetArrayFilters(arrayFilters),
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if result.ModifiedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Item not found in cart"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Quantity updated successfully"})
}
