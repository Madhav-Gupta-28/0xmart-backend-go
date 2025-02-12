package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/database"
	"github.com/Madhav-Gupta-28/0xmart-backend-go/utils"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	listener *utils.BlockchainEventListener
	mu       sync.Mutex

	// Prometheus metrics
	eventProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "event_processing_duration_seconds",
		Help: "Time spent processing blockchain events",
	})
	failedTransactions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "failed_transactions_total",
		Help: "Total number of failed transaction processing attempts",
	})
	successfulTransactions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "successful_transactions_total",
		Help: "Total number of successfully processed transactions",
	})
)

// RetryQueue holds failed events for retry
type RetryQueue struct {
	events []utils.EventData
	mu     sync.Mutex
}

// Add health check endpoint
func HealthCheck(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if listener == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status": "down",
			"error":  "Event listener not initialized",
		})
	}

	health := listener.GetHealth()
	if !health.IsHealthy {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":            "degraded",
			"lastEventTime":     health.LastEventTime,
			"reconnectAttempts": health.ReconnectAttempts,
			"error":             health.LastError,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":        "healthy",
		"lastEventTime": health.LastEventTime,
		"uptime":        health.Uptime,
	})
}

// GetMetrics returns the current metrics
func GetMetrics(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if listener == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Event listener not initialized",
		})
	}

	metrics := listener.GetMetrics()
	return c.JSON(http.StatusOK, metrics)
}

// StartListener with enhanced error handling and metrics
func StartListener(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if listener != nil {
		return c.JSON(http.StatusOK, map[string]string{"status": "Listener already running"})
	}

	listener = utils.NewBlockchainEventListener()
	err := listener.Start()
	if err != nil {
		failedTransactions.Inc()
		listener = nil // Reset on failure
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Start the retry mechanism
	go retryFailedEvents()

	successfulTransactions.Inc()
	return c.JSON(http.StatusOK, map[string]string{"status": "Listener started successfully"})
}

// RetryFailedEvents handles the retry mechanism for failed events
func retryFailedEvents() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if listener == nil {
			continue
		}

		failedEvents := listener.GetFailedEvents()
		for _, event := range failedEvents {
			startTime := time.Now()

			err := processEvent(event)
			if err != nil {
				failedTransactions.Inc()
				// Log the error and keep the event in the retry queue
				continue
			}

			duration := time.Since(startTime).Seconds()
			eventProcessingDuration.Observe(duration)
			successfulTransactions.Inc()
		}
	}
}

// processEvent handles the processing of a single event
func processEvent(event utils.EventData) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add exponential backoff for retries
	backoff := 1 * time.Second
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		err := saveEventToMongoDB(ctx, event)
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		return err
	}

	return nil
}

// saveEventToMongoDB persists the event to MongoDB with retry logic
func saveEventToMongoDB(ctx context.Context, event utils.EventData) error {
	collection := database.DB.Collection("transactions")

	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// Skip duplicate events
			return nil
		}
		return err
	}

	return nil
}

// RestartListener with enhanced error handling
func RestartListener(c echo.Context) error {
	mu.Lock()
	defer mu.Unlock()

	if listener == nil {
		listener = utils.NewBlockchainEventListener()
		err := listener.Start()
		if err != nil {
			failedTransactions.Inc()
			listener = nil
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		successfulTransactions.Inc()
		return c.JSON(http.StatusOK, map[string]string{"status": "Listener started successfully"})
	}

	err := listener.Restart()
	if err != nil {
		failedTransactions.Inc()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	successfulTransactions.Inc()
	return c.JSON(http.StatusOK, map[string]string{"status": "Listener restarted successfully"})
}
