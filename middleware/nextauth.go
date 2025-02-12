package middleware

import (
	"net/http"
	"strings"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/utils"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NextAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the token from the Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "No authorization header",
				})
			}

			// Extract the token from the "Bearer" scheme
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization header format",
				})
			}

			// Verify the JWT token
			claims, err := utils.ValidateJWT(parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			// Convert string ID to ObjectID
			userID, err := primitive.ObjectIDFromHex(claims.UserID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid user ID",
				})
			}

			// Add the user ID to the context
			c.Set("userID", userID)
			return next(c)
		}
	}
}
