package handlers

import (
	"context"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/models"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetProduct(c echo.Context) error {
	productID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	var product models.Product
	err = database.DB.Collection("products").FindOne(c.Request().Context(), bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch product"})
	}

	return c.JSON(http.StatusOK, product)
}

func GetProducts(c echo.Context) error {
	var products []models.Product
	collection := database.DB.Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product models.Product
		cursor.Decode(&product)
		products = append(products, product)
	}

	return c.JSON(http.StatusOK, products)
}

func CreateProduct(c echo.Context) error {
	var product models.Product
	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Validate and format price
	if product.Price == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Price is required"})
	}

	// Convert price to Wei (multiply by 10^18)
	priceFloat, ok := new(big.Float).SetString(product.Price)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid price format"})
	}

	multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	priceInWei := new(big.Float).Mul(priceFloat, multiplier)

	priceInt, _ := priceInWei.Int(nil)
	product.Price = priceInt.String()

	// Generate new ObjectID for the product
	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	collection := database.DB.Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, product)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create product"})
	}

	return c.JSON(http.StatusCreated, product)
}

func GetOrderStatus(c echo.Context) error {
	orderID := c.Param("orderId")
	objID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid order ID"})
	}

	var order models.Order
	err = database.DB.Collection("orders").FindOne(c.Request().Context(), bson.M{"_id": objID}).Decode(&order)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Order not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": string(order.Status)})
}

// SearchProducts handles product search with filters
func SearchProducts(c echo.Context) error {
	query := c.QueryParam("q")
	category := c.QueryParam("category")
	minPrice := c.QueryParam("minPrice")
	maxPrice := c.QueryParam("maxPrice")

	filter := bson.M{}
	if query != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$in": []string{query}}},
		}
	}

	if category != "" {
		categoryID, _ := primitive.ObjectIDFromHex(category)
		filter["categories"] = categoryID
	}

	// Add price range filter if provided
	if minPrice != "" || maxPrice != "" {
		priceFilter := bson.M{}
		if minPrice != "" {
			min, _ := strconv.ParseFloat(minPrice, 64)
			priceFilter["$gte"] = min
		}
		if maxPrice != "" {
			max, _ := strconv.ParseFloat(maxPrice, 64)
			priceFilter["$lte"] = max
		}
		filter["priceUSD"] = priceFilter
	}

	var products []models.Product
	cursor, err := database.DB.Collection("products").Find(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer cursor.Close(c.Request().Context())

	if err = cursor.All(c.Request().Context(), &products); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, products)
}

// RateProduct adds a rating to a product
func RateProduct(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)
	productID, err := primitive.ObjectIDFromHex(c.Param("productId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
	}

	var req struct {
		Rating  float64 `json:"rating"`
		Comment string  `json:"comment"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	rating := models.ProductRating{
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: time.Now(),
	}

	update := bson.M{
		"$push": bson.M{"ratings": rating},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err = database.DB.Collection("products").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": productID},
		update,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Update average rating
	pipeline := []bson.M{
		{"$match": bson.M{"_id": productID}},
		{"$unwind": "$ratings"},
		{"$group": bson.M{
			"_id":       nil,
			"avgRating": bson.M{"$avg": "$ratings.rating"},
		}},
	}

	cursor, err := database.DB.Collection("products").Aggregate(c.Request().Context(), pipeline)
	if err == nil {
		var result struct {
			AvgRating float64 `bson:"avgRating"`
		}
		if cursor.Next(c.Request().Context()) {
			cursor.Decode(&result)
			database.DB.Collection("products").UpdateOne(
				c.Request().Context(),
				bson.M{"_id": productID},
				bson.M{"$set": bson.M{"avgRating": result.AvgRating}},
			)
		}
	}

	return c.JSON(http.StatusOK, rating)
}
