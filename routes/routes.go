package routes

import (
	"fmt"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/handlers"
	customMiddleware "github.com/Madhav-Gupta-28/0xmart-backend-go/middleware"
	"github.com/labstack/echo/v4"
)

// routes/routes.go
func SetupRoutes(e *echo.Echo) {
	fmt.Println("Setting up routes...") // Debug log

	// Public routes
	e.POST("/register", handlers.SignUp)
	e.POST("/login", handlers.LoginUser)

	// NextAuth routes (public)
	e.POST("/api/auth/signup", handlers.SignUp)
	fmt.Println("Registered /api/auth/signup route") // Debug log
	e.POST("/api/auth/signin", handlers.NextAuthSignIn)
	e.GET("/api/auth/csrf", handlers.NextAuthCSRF)
	e.GET("/api/auth/session", handlers.NextAuthSession, customMiddleware.NextAuthMiddleware())

	// Protected API routes
	api := e.Group("/api")
	api.Use(customMiddleware.NextAuthMiddleware())

	fmt.Println("Routes setup complete") // Debug log

	// User routes
	api.GET("/users/me", handlers.GetUserProfile)
	api.PUT("/users/me", handlers.UpdateUserProfile)
	api.GET("/users/me/addresses", handlers.GetUserAddresses)
	api.POST("/users/me/addresses", handlers.AddUserAddress)
	api.PUT("/users/me/addresses/:id", handlers.UpdateUserAddress)
	api.DELETE("/users/me/addresses/:id", handlers.DeleteUserAddress)

	// Product routes
	api.GET("/products", handlers.GetProducts)
	api.GET("/products/:id", handlers.GetProduct)
	api.POST("/products", handlers.CreateProduct)
	// Polling route for order status
	api.GET("/orders/:orderId/status", handlers.GetOrderStatus)

	// Cart routes
	api.GET("/cart", handlers.GetCart)
	api.POST("/cart", handlers.AddToCart)
	api.DELETE("/cart/:productId", handlers.RemoveFromCart)
	api.PUT("/cart/quantity", handlers.UpdateCartItemQuantity)

	// Order routes
	api.POST("/orders", handlers.CreateOrder)
	api.POST("/orders/:orderId/payment", handlers.ProcessPayment)

	// Add this line in SetupRoutes
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
}
