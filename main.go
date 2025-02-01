package main

import (
	"log"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/config"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/handlers"
	"github.com/labstack/echo/v4"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to MongoDB
	err := database.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Create a new Echo instance
	e := echo.New()

	// Define routes
	e.GET("/products", handlers.GetProducts)
	e.POST("/products", handlers.CreateProduct)
	e.GET("/orders", handlers.GetOrders)
	e.POST("/orders", handlers.CreateOrder)
	e.POST("/register", handlers.RegisterUser)
	e.POST("/login", handlers.LoginUser)

	// Start the server
	port := config.GetEnv("PORT", "3000")
	log.Printf("Server running on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
