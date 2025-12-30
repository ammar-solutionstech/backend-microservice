package communication

import (
	"context"
	"fmt"
	"math"
	"time"

	"backend/agent/internal/utils"
	"google.golang.org/grpc"
)

// ReconnectManager manages automatic reconnection with exponential backoff
type ReconnectManager struct {
	logger        *utils.Logger
	maxRetries    int
	initialDelay  time.Duration
	maxDelay      time.Duration
	retryCount    int
	lastError     error
}

// NewReconnectManager creates a new reconnect manager
func NewReconnectManager(logger *utils.Logger) *ReconnectManager {
	return &ReconnectManager{
		logger:       logger,
		maxRetries:   -1, // -1 means infinite retries
		initialDelay: 1 * time.Second,
		maxDelay:     60 * time.Second,
		retryCount:   0,
	}
}

// ConnectWithRetry attempts to establish a connection with automatic retry
func (rm *ReconnectManager) ConnectWithRetry(
	ctx context.Context,
	connectFunc func() (*grpc.ClientConn, error),
) (*grpc.ClientConn, error) {
	for {
		conn, err := connectFunc()
		if err == nil {
			rm.retryCount = 0
			rm.lastError = nil
			return conn, nil
		}

		rm.lastError = err
		rm.retryCount++

		// Check if we should stop retrying
		if rm.maxRetries > 0 && rm.retryCount > rm.maxRetries {
			return nil, fmt.Errorf("max retries exceeded: %v", err)
		}

		// Calculate backoff delay
		delay := rm.calculateBackoff()

		rm.logger.Warn("Connection failed, retrying", map[string]interface{}{
			"error":      err.Error(),
			"retry":      rm.retryCount,
			"delay_secs":  delay.Seconds(),
		})

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue retry loop
		}
	}
}

// calculateBackoff calculates exponential backoff delay
func (rm *ReconnectManager) calculateBackoff() time.Duration {
	// Exponential backoff: initialDelay * 2^retryCount, capped at maxDelay
	delay := float64(rm.initialDelay) * math.Pow(2, float64(rm.retryCount-1))
	if delay > float64(rm.maxDelay) {
		delay = float64(rm.maxDelay)
	}
	return time.Duration(delay)
}

// Reset resets the retry counter
func (rm *ReconnectManager) Reset() {
	rm.retryCount = 0
	rm.lastError = nil
}

// GetLastError returns the last connection error
func (rm *ReconnectManager) GetLastError() error {
	return rm.lastError
}

