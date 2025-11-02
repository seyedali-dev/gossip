// Package event. middleware provides composable middleware for event handlers.
package event

import (
	"context"
	"log"
	"time"
)

// Middleware wraps an EventHandler with additional behavior.
type Middleware func(EventHandler) EventHandler

// -------------------------------------------- Public Functions --------------------------------------------

// WithRetry retries failed handlers with exponential backoff.
func WithRetry(maxRetries int, initialDelay time.Duration) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event *Event) error {
			var err error
			delay := initialDelay

			for attempt := 0; attempt <= maxRetries; attempt++ {
				err = next(ctx, event)
				if err == nil {
					return nil
				}

				if attempt < maxRetries {
					log.Printf("[Middleware] Retry attempt %d/%d for event %s after error: %v", attempt+1, maxRetries, event.Type, err)

					select {
					case <-time.After(delay):
						delay *= 2

					case <-ctx.Done():
						return ctx.Err()
					}
				}
			}

			return err
		}
	}
}

// WithTimeout adds a timeout to handler execution.
func WithTimeout(timeout time.Duration) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event *Event) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next(ctx, event)
			}()

			select {
			case err := <-done:
				return err

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// WithRecovery recovers from panics in handlers.
func WithRecovery() Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event *Event) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Middleware] Recovered from panic in handler for event %s: %v", event.Type, r)
					err = nil
				}
			}()

			return next(ctx, event)
		}
	}
}

// WithLogging logs handler execution.
func WithLogging() Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event *Event) error {
			start := time.Now()
			err := next(ctx, event)
			duration := time.Since(start)

			if err != nil {
				log.Printf("[Middleware] Handler for %s failed after %v: %v", event.Type, duration, err)
			} else {
				log.Printf("[Middleware] Handler for %s completed in %v", event.Type, duration)
			}

			return err
		}
	}
}

// Chain chains multiple middlewares together.
func Chain(middlewares ...Middleware) Middleware {
	return func(handler EventHandler) EventHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}
