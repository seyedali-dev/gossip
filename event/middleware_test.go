// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

package event_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seyedali-dev/gossip/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------- Retry Middleware Tests --------------------------------------------

func TestWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	attempts := int32(0)

	handler := func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&attempts, 1)
		return nil
	}

	wrappedHandler := event.WithRetry(3, 10*time.Millisecond)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts), "Should succeed on first attempt")
}

func TestWithRetry_RetriesUntilSuccess(t *testing.T) {
	attempts := int32(0)

	handler := func(ctx context.Context, e *event.Event) error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	wrappedHandler := event.WithRetry(3, 10*time.Millisecond)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts), "Should retry until success")
}

func TestWithRetry_ExhaustsRetries(t *testing.T) {
	attempts := int32(0)
	expectedErr := errors.New("persistent error")

	handler := func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&attempts, 1)
		return expectedErr
	}

	wrappedHandler := event.WithRetry(3, 10*time.Millisecond)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, int32(4), atomic.LoadInt32(&attempts), "Initial attempt + 3 retries")
}

func TestWithRetry_ExponentialBackoff(t *testing.T) {
	attempts := make([]time.Time, 0)

	handler := func(ctx context.Context, e *event.Event) error {
		attempts = append(attempts, time.Now())
		return errors.New("error")
	}

	wrappedHandler := event.WithRetry(2, 50*time.Millisecond)(handler)

	_ = wrappedHandler(context.Background(), event.NewEvent("test", nil))

	require.Len(t, attempts, 3, "Should have 3 attempts total")

	// Check backoff intervals
	firstDelay := attempts[1].Sub(attempts[0])
	secondDelay := attempts[2].Sub(attempts[1])

	assert.GreaterOrEqual(t, firstDelay, 50*time.Millisecond, "First retry delay")
	assert.GreaterOrEqual(t, secondDelay, 100*time.Millisecond, "Second retry delay (doubled)")
}

func TestWithRetry_RespectsContext(t *testing.T) {
	attempts := int32(0)

	handler := func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("error")
	}

	wrappedHandler := event.WithRetry(10, 100*time.Millisecond)(handler)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err := wrappedHandler(ctx, event.NewEvent("test", nil))

	assert.Error(t, err)
	assert.Less(t, atomic.LoadInt32(&attempts), int32(11), "Should stop retrying when context is cancelled")
}

// -------------------------------------------- Timeout Middleware Tests --------------------------------------------

func TestWithTimeout_CompletesBeforeTimeout(t *testing.T) {
	executed := false

	handler := func(ctx context.Context, e *event.Event) error {
		time.Sleep(10 * time.Millisecond)
		executed = true
		return nil
	}

	wrappedHandler := event.WithTimeout(100 * time.Millisecond)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestWithTimeout_ExceedsTimeout(t *testing.T) {
	handler := func(ctx context.Context, e *event.Event) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	wrappedHandler := event.WithTimeout(50 * time.Millisecond)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestWithTimeout_HandlerError(t *testing.T) {
	expectedErr := errors.New("handler error")

	handler := func(ctx context.Context, e *event.Event) error {
		return expectedErr
	}

	wrappedHandler := event.WithTimeout(100 * time.Millisecond)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// -------------------------------------------- Recovery Middleware Tests --------------------------------------------

func TestWithRecovery_HandlerSucceeds(t *testing.T) {
	executed := false

	handler := func(ctx context.Context, e *event.Event) error {
		executed = true
		return nil
	}

	wrappedHandler := event.WithRecovery()(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestWithRecovery_HandlerPanics(t *testing.T) {
	handler := func(ctx context.Context, e *event.Event) error {
		panic("something went wrong")
	}

	wrappedHandler := event.WithRecovery()(handler)

	// Should not panic
	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err, "Recovery should return nil after catching panic")
}

func TestWithRecovery_HandlerPanicsWithError(t *testing.T) {
	panicErr := errors.New("panic error")

	handler := func(ctx context.Context, e *event.Event) error {
		panic(panicErr)
	}

	wrappedHandler := event.WithRecovery()(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err, "Recovery should catch panic and return nil")
}

// -------------------------------------------- Logging Middleware Tests --------------------------------------------

func TestWithLogging_Success(t *testing.T) {
	executed := false

	handler := func(ctx context.Context, e *event.Event) error {
		time.Sleep(10 * time.Millisecond)
		executed = true
		return nil
	}

	wrappedHandler := event.WithLogging()(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestWithLogging_Failure(t *testing.T) {
	expectedErr := errors.New("handler failed")

	handler := func(ctx context.Context, e *event.Event) error {
		return expectedErr
	}

	wrappedHandler := event.WithLogging()(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// -------------------------------------------- Chain Middleware Tests --------------------------------------------

func TestChain_SingleMiddleware(t *testing.T) {
	attempts := int32(0)

	handler := func(ctx context.Context, e *event.Event) error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			return errors.New("error")
		}
		return nil
	}

	wrappedHandler := event.Chain(
		event.WithRetry(2, 10*time.Millisecond),
	)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestChain_MultipleMiddlewares(t *testing.T) {
	attempts := int32(0)

	handler := func(ctx context.Context, e *event.Event) error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			return errors.New("error")
		}
		time.Sleep(20 * time.Millisecond)
		return nil
	}

	wrappedHandler := event.Chain(
		event.WithRecovery(),
		event.WithRetry(2, 10*time.Millisecond),
		event.WithTimeout(500*time.Millisecond),
		event.WithLogging(),
	)(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&attempts), int32(2))
}

func TestChain_OrderMatters(t *testing.T) {
	// Recovery should be outermost to catch panics from other middleware
	handler := func(ctx context.Context, e *event.Event) error {
		panic("test panic")
	}

	// Recovery first - should catch panic
	wrappedHandler1 := event.Chain(
		event.WithRecovery(),
		event.WithLogging(),
	)(handler)

	err1 := wrappedHandler1(context.Background(), event.NewEvent("test", nil))
	assert.NoError(t, err1, "Recovery should catch panic")

	// If logging was first, panic would escape (but WithRecovery still catches it here)
	// This test demonstrates the importance of middleware order
}

func TestChain_EmptyChain(t *testing.T) {
	executed := false

	handler := func(ctx context.Context, e *event.Event) error {
		executed = true
		return nil
	}

	wrappedHandler := event.Chain()(handler)

	err := wrappedHandler(context.Background(), event.NewEvent("test", nil))

	assert.NoError(t, err)
	assert.True(t, executed)
}

// -------------------------------------------- Integration Tests --------------------------------------------

func TestMiddleware_WithEventBus(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	attempts := int32(0)

	handler := func(ctx context.Context, e *event.Event) error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	wrappedHandler := event.Chain(
		event.WithRetry(3, 10*time.Millisecond),
		event.WithTimeout(500*time.Millisecond),
	)(handler)

	bus.Subscribe("test.event", wrappedHandler)

	bus.Publish(event.NewEvent("test.event", nil))
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestMiddleware_MultipleHandlersWithDifferentMiddleware(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	handler1Called := false
	handler2Called := false

	// Handler 1 with retry
	handler1 := event.WithRetry(2, 10*time.Millisecond)(func(ctx context.Context, e *event.Event) error {
		handler1Called = true
		return nil
	})

	// Handler 2 with timeout
	handler2 := event.WithTimeout(100 * time.Millisecond)(func(ctx context.Context, e *event.Event) error {
		handler2Called = true
		return nil
	})

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)

	bus.Publish(event.NewEvent("test.event", nil))
	time.Sleep(100 * time.Millisecond)

	assert.True(t, handler1Called)
	assert.True(t, handler2Called)
}

// -------------------------------------------- Benchmark Tests --------------------------------------------

func BenchmarkWithRetry(b *testing.B) {
	handler := func(ctx context.Context, e *event.Event) error {
		return nil
	}

	wrappedHandler := event.WithRetry(3, 10*time.Millisecond)(handler)
	evt := event.NewEvent("test", nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrappedHandler(ctx, evt)
	}
}

func BenchmarkWithTimeout(b *testing.B) {
	handler := func(ctx context.Context, e *event.Event) error {
		return nil
	}

	wrappedHandler := event.WithTimeout(100 * time.Millisecond)(handler)
	evt := event.NewEvent("test", nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrappedHandler(ctx, evt)
	}
}

func BenchmarkChain(b *testing.B) {
	handler := func(ctx context.Context, e *event.Event) error {
		return nil
	}

	wrappedHandler := event.Chain(
		event.WithRecovery(),
		event.WithRetry(2, 10*time.Millisecond),
		event.WithTimeout(100*time.Millisecond),
		event.WithLogging(),
	)(handler)

	evt := event.NewEvent("test", nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrappedHandler(ctx, evt)
	}
}
