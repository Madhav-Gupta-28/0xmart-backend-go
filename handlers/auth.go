package handlers

import (
	"net/http"
	"time"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/models"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// NextAuthLogin handles login requests from NextAuth.js
func NextAuthLogin(c echo.Context) error {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&credentials); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	var user models.User
	err := database.DB.Collection("users").FindOne(
		c.Request().Context(),
		bson.M{"email": credentials.Email},
	).Decode(&user)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

// NextAuthCallback handles OAuth callbacks from NextAuth.js
func NextAuthCallback(c echo.Context) error {
	var userData struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		Image      string `json:"image"`
		Provider   string `json:"provider"`
		ProviderId string `json:"providerId"`
	}

	if err := c.Bind(&userData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Check if user exists
	var user models.User
	err := database.DB.Collection("users").FindOne(
		c.Request().Context(),
		bson.M{"email": userData.Email},
	).Decode(&user)

	if err != nil {
		// Create new user if not exists
		user = models.User{
			ID:            primitive.NewObjectID(),
			Email:         userData.Email,
			Name:          userData.Name,
			Image:         userData.Image,
			Provider:      userData.Provider,
			ProviderId:    userData.ProviderId,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		_, err = database.DB.Collection("users").InsertOne(c.Request().Context(), user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
		}
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":  user,
		"token": token,
	})
}
