package handlers

import (
	"net/http"
	"sync"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/utils"

	"github.com/labstack/echo/v4"
)

var (
	listener *utils.BlockchainEventListener
	mu       sync.Mutex // Add mutex for thread safety
)

// StartListener starts the blockchain event listener
func StartListener(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if listener != nil {
		return c.JSON(http.StatusOK, map[string]string{"status": "Listener already running"})
	}

	listener = utils.NewBlockchainEventListener()
	err := listener.Start()
	if err != nil {
		listener = nil // Reset on failure
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "Listener started successfully"})
}

// RestartListener restarts the blockchain event listener
func RestartListener(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if listener == nil {
		// Automatically start if not initialized
		listener = utils.NewBlockchainEventListener()
		err := listener.Start()
		if err != nil {
			listener = nil
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "Listener started successfully"})
	}

	err := listener.Restart()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "Listener restarted successfully"})
}
