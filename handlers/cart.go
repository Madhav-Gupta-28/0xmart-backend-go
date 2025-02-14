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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil || productID.IsZero() {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// Verify product exists
	var product models.Product
	err = database.DB.Collection("products").FindOne(
		c.Request().Context(),
		bson.M{"_id": productID},
	).Decode(&product)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
	}

	// Check if item already exists in cart
	update := bson.M{
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	// Try to update existing item
	result := database.DB.Collection("carts").FindOneAndUpdate(
		c.Request().Context(),
		bson.M{
			"userId": userID,
			"items": bson.M{
				"$elemMatch": bson.M{
					"productId": productID,
					"size":      req.Size,
				},
			},
		},
		bson.M{
			"$inc": bson.M{
				"items.$.quantity": req.Quantity,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if result.Err() == mongo.ErrNoDocuments {
		// Item doesn't exist, add new item
		update = bson.M{
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

		result = database.DB.Collection("carts").FindOneAndUpdate(
			c.Request().Context(),
			bson.M{"userId": userID},
			update,
			options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
		)
	}

	if result.Err() != nil && result.Err() != mongo.ErrNoDocuments {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update cart"})
	}

	// Return updated cart
	var cart models.Cart
	err = result.Decode(&cart)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"message": "Cart updated successfully"})
	}

	return c.JSON(http.StatusOK, cart)
}

// GetCart retrieves the user's cart and cleans invalid products
func GetCart(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)

	var cart models.Cart
	err := database.DB.Collection("carts").FindOne(
		c.Request().Context(),
		bson.M{"userId": userID},
	).Decode(&cart)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create new cart if none exists
			cart = models.Cart{
				ID:        primitive.NewObjectID(),
				UserID:    userID,
				Items:     []models.CartItem{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_, err = database.DB.Collection("carts").InsertOne(c.Request().Context(), cart)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create cart"})
			}
		} else {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cart"})
		}
	}

	// Clean invalid products from cart
	validItems := []models.CartItem{}
	for _, item := range cart.Items {
		if !item.ProductID.IsZero() {
			// Verify product exists
			var product models.Product
			err := database.DB.Collection("products").FindOne(
				c.Request().Context(),
				bson.M{"_id": item.ProductID},
			).Decode(&product)
			if err == nil {
				validItems = append(validItems, item)
			}
		}
	}

	// Update cart if invalid items were removed
	if len(validItems) != len(cart.Items) {
		_, err = database.DB.Collection("carts").UpdateOne(
			c.Request().Context(),
			bson.M{"_id": cart.ID},
			bson.M{
				"$set": bson.M{
					"items":     validItems,
					"updatedAt": time.Now(),
				},
			},
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to clean cart"})
		}
		cart.Items = validItems
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
	userID := c.Get("userID").(primitive.ObjectID)

	var req struct {
		ProductID string `json:"productId"`
		Quantity  int    `json:"quantity"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate quantity
	if req.Quantity < 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Quantity must be at least 1"})
	}

	productID, err := primitive.ObjectIDFromHex(req.ProductID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	// First, check if the product exists and has enough stock
	var product models.Product
	err = database.DB.Collection("products").FindOne(
		c.Request().Context(),
		bson.M{"_id": productID},
	).Decode(&product)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
	}

	// Update the cart item quantity
	update := bson.M{
		"$set": bson.M{
			"items.$[elem].quantity": req.Quantity,
			"updatedAt":              time.Now(),
		},
	}

	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem.productId": productID},
		},
	}

	result, err := database.DB.Collection("carts").UpdateOne(
		c.Request().Context(),
		bson.M{"userId": userID},
		update,
		options.Update().SetArrayFilters(arrayFilters),
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update quantity: " + err.Error()})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Cart or product not found"})
	}

	if result.ModifiedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No changes made to cart"})
	}

	// Get updated cart
	var cart models.Cart
	err = database.DB.Collection("carts").FindOne(
		c.Request().Context(),
		bson.M{"userId": userID},
	).Decode(&cart)

	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{"message": "Quantity updated successfully"})
	}

	return c.JSON(http.StatusOK, cart)
}
