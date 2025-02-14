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

	// Public routes (no authentication required)
	e.POST("/register", handlers.SignUp)
	e.POST("/login", handlers.LoginUser)
	e.POST("/api/auth/signup", handlers.SignUp)
	e.POST("/api/auth/signin", handlers.NextAuthSignIn)
	e.GET("/api/auth/csrf", handlers.NextAuthCSRF)

	// Public Product routes
	e.GET("/api/products", handlers.GetProducts)           // Make this public
	e.GET("/api/products/:id", handlers.GetProduct)        // Make this public
	e.GET("/api/products/search", handlers.SearchProducts) // Make this public

	// Protected API routes (require authentication)
	api := e.Group("/api")
	api.Use(customMiddleware.NextAuthMiddleware())

	// Protected Product routes
	api.POST("/products", handlers.CreateProduct) // Only creating products needs auth
	api.POST("/products/:productId/ratings", handlers.RateProduct)

	// Protected User routes
	api.GET("/users/me", handlers.GetUserProfile)
	api.PUT("/users/me", handlers.UpdateUserProfile)
	api.GET("/users/me/addresses", handlers.GetUserAddresses)
	api.POST("/users/me/addresses", handlers.AddUserAddress)
	api.PUT("/users/me/addresses/:id", handlers.UpdateUserAddress)
	api.DELETE("/users/me/addresses/:id", handlers.DeleteUserAddress)

	// Cart routes
	api.GET("/cart", handlers.GetCart)
	api.POST("/cart", handlers.AddToCart)
	api.DELETE("/cart/:productId", handlers.RemoveFromCart)
	api.PUT("/cart/quantity", handlers.UpdateCartItemQuantity)

	// Order routes
	api.GET("/orders", handlers.GetOrders)                        // Get all orders
	api.GET("/orders/:orderId", handlers.GetOrder)                // Get single order
	api.GET("/orders/:orderId/status", handlers.GetOrderStatus)   // Get order status
	api.POST("/orders", handlers.CreateOrder)                     // Create order
	api.POST("/orders/:orderId/payment", handlers.ProcessPayment) // Process payment

	// Add this line in SetupRoutes
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
}
