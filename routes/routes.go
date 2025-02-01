package routes

import (
	"github.com/Madhav-Gupta-28/0xmart-backend-go/handlers"
	"github.com/labstack/echo/v4"
)

// routes/routes.go
func SetupRoutes(e *echo.Echo) {
	// Public routes
	e.POST("/register", handlers.RegisterUser)
	e.POST("/login", handlers.LoginUser)

	// Protected routes
	api := e.Group("/api")
	// api.Use(middleware.JWT([]byte(os.Getenv("JWT_SECRET"))))

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
}
