package main

import (
	"fmt"
	"log"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/config"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Initialize Echo
	e := echo.New()
	fmt.Println("Echo initialized") // Debug log

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Connect to MongoDB
	if err := database.ConnectDB(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Setup routes
	routes.SetupRoutes(e)

	// Start the server
	port := config.GetEnv("PORT", "3000")
	fmt.Printf("Server starting on port %s...\n", port) // Debug log
	e.Logger.Fatal(e.Start(":" + port))
}
