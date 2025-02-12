package handlers

import (
	"context"
	"fmt"
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

type NextAuthSignInRequest struct {
	Provider   string `json:"provider"`
	Email      string `json:"email,omitempty"`
	Password   string `json:"password,omitempty"`
	Name       string `json:"name,omitempty"`
	Image      string `json:"image,omitempty"`
	ProviderId string `json:"providerId,omitempty"`
}

// SignUpRequest represents the expected request body for signup
type SignUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignUp handles user registration
// SignUp handles user registration
func SignUp(c echo.Context) error {
	fmt.Println("Signing up")
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	// Validate input
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "All fields are required"})
	}

	// Check if user already exists
	var existingUser models.User
	err := database.DB.Collection("users").FindOne(
		c.Request().Context(),
		bson.M{"email": req.Email},
	).Decode(&existingUser)

	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process password"})
	}

	// Create new user
	newUser := models.User{
		ID:            primitive.NewObjectID(),
		Name:          req.Name,
		Email:         req.Email,
		Password:      string(hashedPassword),
		Provider:      "credentials",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Insert user into database
	_, err = database.DB.Collection("users").InsertOne(c.Request().Context(), newUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(newUser.ID.Hex())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	// Don't return password in response
	newUser.Password = ""

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user":  newUser,
		"token": token,
	})
}

// NextAuthSignIn handles sign-in requests from NextAuth
func NextAuthSignIn(c echo.Context) error {
	var req NextAuthSignInRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user exists
	var user models.User
	err := database.DB.Collection("users").FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)

	if err != nil {
		// Create new user if not exists
		user = models.User{
			ID:            primitive.NewObjectID(),
			Email:         req.Email,
			Name:          req.Name,
			Image:         req.Image,
			Provider:      req.Provider,
			ProviderId:    req.ProviderId,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		_, err = database.DB.Collection("users").InsertOne(ctx, user)
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
		"user": map[string]interface{}{
			"id":            user.ID.Hex(),
			"email":         user.Email,
			"name":          user.Name,
			"image":         user.Image,
			"emailVerified": user.EmailVerified,
		},
		"token": token,
	})
}

// NextAuthSession verifies and returns the current session
func NextAuthSession(c echo.Context) error {
	userID, ok := c.Get("userID").(primitive.ObjectID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid session"})
	}

	var user models.User
	err := database.DB.Collection("users").FindOne(c.Request().Context(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":            user.ID.Hex(),
			"email":         user.Email,
			"name":          user.Name,
			"image":         user.Image,
			"emailVerified": user.EmailVerified,
			"addresses":     user.Addresses,
			"phoneNumber":   user.PhoneNumber,
			"preferences":   user.Preferences,
		},
	})
}

// NextAuthCSRF generates and returns a CSRF token
func NextAuthCSRF(c echo.Context) error {
	token := utils.GenerateCSRFToken()
	return c.JSON(http.StatusOK, map[string]string{
		"csrfToken": token,
	})
}
