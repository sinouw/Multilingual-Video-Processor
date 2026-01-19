package utils

import (
	"fmt"
	"log/slog"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
}

// Retry executes a function with retry logic
func Retry(fn func() error, config RetryConfig) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			if attempt > 1 {
				slog.Info("Retry succeeded", "attempt", attempt)
			}
			return nil
		}

		lastErr = err
		if attempt < config.MaxAttempts {
			slog.Warn("Retry attempt failed, retrying",
				"attempt", attempt,
				"maxAttempts", config.MaxAttempts,
				"delay", delay,
				"error", err)

			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return fmt.Errorf("retry exhausted after %d attempts: %w", config.MaxAttempts, lastErr)
}
