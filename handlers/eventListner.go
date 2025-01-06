package handlers

import (
	"net/http"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/utils"

	"github.com/labstack/echo/v4"
)

var listener *utils.BlockchainEventListener

// StartListener starts the blockchain event listener
func StartListener(c echo.Context) error {
	if listener == nil {
		listener = utils.NewBlockchainEventListener()
	}
	err := listener.Start()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "Listener started"})
}

// RestartListener restarts the blockchain event listener
func RestartListener(c echo.Context) error {
	if listener == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Listener not initialized"})
	}
	err := listener.Restart()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "Listener restarted"})
}
