package main

import (
	"0xmart-backend-go/config"
	"0xmart-backend-go/handlers"
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Create a new Echo instance
	e := echo.New()

	// Define routes
	e.GET("/start", handlers.StartListener)
	e.GET("/restart", handlers.RestartListener)

	// Start the server
	port := config.GetEnv("PORT", "3000")
	log.Printf("Server running on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
