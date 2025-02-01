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
	var req CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	if !common.IsHexAddress(req.WalletAddress) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid wallet address"})
	}

	userID, ok := c.Get("userID").(primitive.ObjectID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not authenticated"})
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

		// Validate stock
		if stock, ok := product.Stock[item.Size]; !ok || stock < item.Quantity {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": fmt.Sprintf("Insufficient stock for product %s size %s", product.Name, item.Size),
			})
		}

		price, ok := new(big.Int).SetString(product.Price, 10)
		if !ok {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Invalid product price format"})
		}

		itemTotal := new(big.Int).Mul(price, big.NewInt(int64(item.Quantity)))
		totalPrice.Add(totalPrice, itemTotal)

		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Size:      item.Size,
			Quantity:  item.Quantity,
			Price:     product.Price,
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
