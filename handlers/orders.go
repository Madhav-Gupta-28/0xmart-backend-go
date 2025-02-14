package handlers

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"log"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateOrderRequest struct {
	WalletAddress string `json:"walletAddress"`
}

func CreateOrder(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)
	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Make wallet address validation optional
	if req.WalletAddress != "" && !common.IsHexAddress(req.WalletAddress) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid wallet address format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user's cart
	var cart models.Cart
	err := database.DB.Collection("carts").FindOne(ctx, bson.M{"userId": userID}).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Cart is empty"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cart"})
	}

	if len(cart.Items) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Cart is empty"})
	}

	// Calculate total price and validate items
	totalPrice := big.NewInt(0)
	var orderItems []models.OrderItem
	productsCollection := database.DB.Collection("products")

	for _, item := range cart.Items {
		var product models.Product
		err := productsCollection.FindOne(ctx, bson.M{"_id": item.ProductID}).Decode(&product)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to fetch product %s", item.ProductID.Hex()),
			})
		}

		// Convert string size to ProductSize
		productSize := models.ProductSize(item.Size)

		// Validate stock
		if stock, ok := product.Stock[productSize]; !ok || stock < item.Quantity {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Insufficient stock for product %s size %s", product.Name, item.Size),
			})
		}

		// Parse the price (already in Wei)
		price := new(big.Int)
		price, ok := price.SetString(product.Price, 10)
		if !ok {
			// If parsing fails, try to convert it to Wei
			priceFloat, ok := new(big.Float).SetString(product.Price)
			if !ok {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": fmt.Sprintf("Invalid price format for product %s", product.Name),
				})
			}
			multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
			priceInWei := new(big.Float).Mul(priceFloat, multiplier)
			price, _ = priceInWei.Int(nil)
		}

		itemTotal := new(big.Int).Mul(price, big.NewInt(int64(item.Quantity)))
		totalPrice.Add(totalPrice, itemTotal)

		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Size:      productSize,
			Quantity:  item.Quantity,
			Price:     price.String(),
		})
	}

	// Create order
	order := models.Order{
		ID:            primitive.NewObjectID(),
		UserID:        userID,
		Items:         orderItems,
		TotalPrice:    totalPrice.String(),
		Status:        models.OrderStatusPending,
		WalletAddress: req.WalletAddress,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert order
	_, err = database.DB.Collection("orders").InsertOne(ctx, order)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create order"})
	}

	// Clear cart after successful order creation
	_, err = database.DB.Collection("carts").UpdateOne(
		ctx,
		bson.M{"userId": userID},
		bson.M{"$set": bson.M{"items": []models.CartItem{}, "updatedAt": time.Now()}},
	)
	if err != nil {
		log.Printf("Failed to clear cart after order creation: %v", err)
	}

	return c.JSON(http.StatusCreated, order)
}

func ProcessPayment(c echo.Context) error {
	_, ok := c.Get("userID").(primitive.ObjectID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
	}

	orderID, err := primitive.ObjectIDFromHex(c.Param("orderId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid order ID format"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var order models.Order
	err = database.DB.Collection("orders").FindOne(ctx, bson.M{"_id": orderID}).Decode(&order)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Order not found"})
	}

	// Update order status to pending
	update := bson.M{
		"$set": bson.M{
			"status":    models.OrderStatusPending,
			"updatedAt": time.Now(),
		},
	}

	_, err = database.DB.Collection("orders").UpdateOne(ctx, bson.M{"_id": orderID}, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update order"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Payment processing initiated"})
}

func updateOrderStatus(ctx context.Context, orderID primitive.ObjectID, status models.OrderStatus, txHash string) error {
	update := bson.M{
		"$set": bson.M{
			"status":    status,
			"updatedAt": time.Now(),
		},
	}

	if txHash != "" {
		update["$set"].(bson.M)["txHash"] = txHash
	}

	_, err := database.DB.Collection("orders").UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		update,
	)
	return err
}

func GetOrders(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var orders []models.Order
	cursor, err := database.DB.Collection("orders").Find(ctx, bson.M{"userId": userID})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch orders"})
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &orders); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode orders"})
	}

	return c.JSON(http.StatusOK, orders)
}

// UpdateOrderFulfillment updates the order fulfillment status
func UpdateOrderFulfillment(c echo.Context) error {
	orderID := c.Param("orderId")
	var req struct {
		Status         models.FulfillmentStatus `json:"status"`
		TrackingNumber string                   `json:"trackingNumber,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	objID, _ := primitive.ObjectIDFromHex(orderID)
	update := bson.M{
		"$set": bson.M{
			"fulfillmentStatus": req.Status,
			"trackingNumber":    req.TrackingNumber,
			"updatedAt":         time.Now(),
		},
	}

	_, err := database.DB.Collection("orders").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": objID},
		update,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}

// GetOrder retrieves a specific order
func GetOrder(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)
	orderID, err := primitive.ObjectIDFromHex(c.Param("orderId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid order ID"})
	}

	var order models.Order
	err = database.DB.Collection("orders").FindOne(
		c.Request().Context(),
		bson.M{
			"_id":    orderID,
			"userId": userID, // Ensure user can only access their own orders
		},
	).Decode(&order)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Order not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch order"})
	}

	return c.JSON(http.StatusOK, order)
}

// Update GetOrderStatus to be more detailed
