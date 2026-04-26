// Package cron - removed module stub
package cron

import "time"

// Stub package - cron module has been removed

// RetryConfig stub
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig returns default retry config (stub)
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
	}
}
