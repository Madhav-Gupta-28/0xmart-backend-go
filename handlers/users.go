package handlers

import (
	"context"
	"net/http"
	"time"

	"net/mail"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/middleware"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/models"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Validate email format
	if !isValidEmail(user.Email) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid email format"})
	}

	// Check if email already exists
	collection := database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	existingUser := collection.FindOne(ctx, bson.M{"email": user.Email})
	if existingUser.Err() == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
	}

	// Validate password strength
	if len(user.Password) < 8 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password must be at least 8 characters"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process password"})
	}
	user.Password = string(hashedPassword)
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	// Don't return password in response
	user.Password = ""
	return c.JSON(http.StatusCreated, user)
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Basic email validation
	_, err := mail.ParseAddress(email)
	return err == nil
}

func LoginUser(c echo.Context) error {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&loginRequest); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	collection := database.DB.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": loginRequest.Email}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password"})
	}

	token, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// GetUserProfile retrieves the user's profile
func GetUserProfile(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)

	var user models.User
	err := database.DB.Collection("users").FindOne(
		c.Request().Context(),
		bson.M{"_id": userID},
	).Decode(&user)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUserProfile updates the user's profile information
func UpdateUserProfile(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)

	var updateData struct {
		Name        string                 `json:"name"`
		PhoneNumber string                 `json:"phoneNumber"`
		Preferences map[string]interface{} `json:"preferences"`
	}

	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	update := bson.M{
		"$set": bson.M{
			"name":        updateData.Name,
			"phoneNumber": updateData.PhoneNumber,
			"preferences": updateData.Preferences,
			"updatedAt":   time.Now(),
		},
	}

	_, err := database.DB.Collection("users").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": userID},
		update,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update profile"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Profile updated successfully"})
}

// AddUserAddress adds a new address to the user's profile
func AddUserAddress(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)

	var address models.Address
	if err := c.Bind(&address); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid address data"})
	}

	address.ID = primitive.NewObjectID()

	update := bson.M{
		"$push": bson.M{"addresses": address},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err := database.DB.Collection("users").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": userID},
		update,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add address"})
	}

	return c.JSON(http.StatusOK, address)
}

func GetUserAddresses(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)

	var user models.User
	err := database.DB.Collection("users").FindOne(
		c.Request().Context(),
		bson.M{"_id": userID},
	).Decode(&user)

	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, user.Addresses)
}

func UpdateUserAddress(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)
	addressID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid address ID"})
	}

	var address models.Address
	if err := c.Bind(&address); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	address.ID = addressID

	update := bson.M{
		"$set": bson.M{
			"addresses.$[elem]": address,
			"updatedAt":         time.Now(),
		},
	}

	arrayFilters := options.ArrayFilters{
		Filters: []interface{}{
			bson.M{"elem._id": addressID},
		},
	}

	result, err := database.DB.Collection("users").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": userID},
		update,
		options.Update().SetArrayFilters(arrayFilters),
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update address"})
	}

	if result.MatchedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Address not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Address updated successfully"})
}

func DeleteUserAddress(c echo.Context) error {
	userID := c.Get("userID").(primitive.ObjectID)
	addressID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid address ID"})
	}

	update := bson.M{
		"$pull": bson.M{
			"addresses": bson.M{"_id": addressID},
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	result, err := database.DB.Collection("users").UpdateOne(
		c.Request().Context(),
		bson.M{"_id": userID},
		update,
	)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete address"})
	}

	if result.ModifiedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Address not found"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Address deleted successfully"})
}
