/*
Copyright 2024 Keiailab.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mongodb

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// RetryConfig contains configuration for retry behavior
type RetryConfig struct {
	// InitialDelay is the initial delay before the first retry
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Factor is the multiplier for each retry
	Factor float64
	// Jitter adds randomness to the delay
	Jitter float64
	// MaxRetries is the maximum number of retries (0 for unlimited)
	MaxRetries int
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Factor:       2.0,
		Jitter:       0.1,
		MaxRetries:   10,
	}
}

// QuickRetryConfig returns a faster retry configuration for quick operations
func QuickRetryConfig() RetryConfig {
	return RetryConfig{
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Factor:       1.5,
		Jitter:       0.1,
		MaxRetries:   5,
	}
}

// LongRetryConfig returns a longer retry configuration for operations that may take time
func LongRetryConfig() RetryConfig {
	return RetryConfig{
		InitialDelay: 2 * time.Second,
		MaxDelay:     60 * time.Second,
		Factor:       2.0,
		Jitter:       0.2,
		MaxRetries:   20,
	}
}

// RetryWithBackoff retries a function with exponential backoff
func RetryWithBackoff(ctx context.Context, config RetryConfig, fn func() error) error {
	backoff := wait.Backoff{
		Duration: config.InitialDelay,
		Factor:   config.Factor,
		Jitter:   config.Jitter,
		Steps:    config.MaxRetries,
		Cap:      config.MaxDelay,
	}

	var lastErr error
	err := wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		if err := fn(); err != nil {
			lastErr = err
			return false, nil // Continue retrying
		}
		return true, nil // Success
	})

	if err != nil {
		if lastErr != nil {
			return lastErr
		}
		return err
	}

	return nil
}

// RetryUntilSuccess retries a function until it succeeds or context is cancelled
func RetryUntilSuccess(ctx context.Context, interval time.Duration, fn func() error) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Try immediately first
	if err := fn(); err == nil {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := fn(); err == nil {
				return nil
			}
		}
	}
}

// WaitForCondition waits for a condition to be true
func WaitForCondition(ctx context.Context, interval time.Duration, condition func() (bool, error)) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Check immediately first
	if done, err := condition(); err != nil {
		return err
	} else if done {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			done, err := condition()
			if err != nil {
				return err
			}
			if done {
				return nil
			}
		}
	}
}

// WaitForConditionWithBackoff waits for a condition with exponential backoff
func WaitForConditionWithBackoff(ctx context.Context, config RetryConfig, condition func() (bool, error)) error {
	backoff := wait.Backoff{
		Duration: config.InitialDelay,
		Factor:   config.Factor,
		Jitter:   config.Jitter,
		Steps:    config.MaxRetries,
		Cap:      config.MaxDelay,
	}

	return wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		return condition()
	})
}

// WithTimeout creates a context with timeout for retry operations
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}

// WithDeadline creates a context with deadline for retry operations
func WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(ctx, deadline)
}
