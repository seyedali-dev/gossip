// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

package event_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seyedali-dev/gossip/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------- End-to-End Tests --------------------------------------------

func TestE2E_CompleteUserRegistrationFlow(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	// Simulate complete user registration workflow
	var mu sync.Mutex
	actions := make([]string, 0)

	// Email handler
	emailHandler := func(ctx context.Context, e *event.Event) error {
		mu.Lock()
		defer mu.Unlock()
		actions = append(actions, "email_sent")
		return nil
	}

	// Audit handler
	auditHandler := func(ctx context.Context, e *event.Event) error {
		mu.Lock()
		defer mu.Unlock()
		actions = append(actions, "audit_logged")
		return nil
	}

	// Metrics handler
	metricsHandler := func(ctx context.Context, e *event.Event) error {
		mu.Lock()
		defer mu.Unlock()
		actions = append(actions, "metrics_recorded")
		return nil
	}

	// Profile creation handler
	profileHandler := func(ctx context.Context, e *event.Event) error {
		mu.Lock()
		defer mu.Unlock()
		actions = append(actions, "profile_created")
		return nil
	}

	// Register all handlers
	bus.Subscribe("user.registered", emailHandler)
	bus.Subscribe("user.registered", auditHandler)
	bus.Subscribe("user.registered", metricsHandler)
	bus.Subscribe("user.registered", profileHandler)

	// Simulate user registration
	userData := map[string]interface{}{
		"user_id":  "user_123",
		"email":    "john@example.com",
		"username": "john_doe",
	}

	evt := event.NewEvent("user.registered", userData).
		WithMetadata("source", "web").
		WithMetadata("ip_address", "192.168.1.100")

	bus.Publish(evt)

	// Wait for all handlers
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Verify all actions occurred
	assert.Len(t, actions, 4)
	assert.Contains(t, actions, "email_sent")
	assert.Contains(t, actions, "audit_logged")
	assert.Contains(t, actions, "metrics_recorded")
	assert.Contains(t, actions, "profile_created")
}

func TestE2E_OrderProcessingPipeline(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	orderState := struct {
		sync.Mutex
		inventoryReserved bool
		paymentProcessed  bool
		shippingScheduled bool
		emailSent         bool
	}{}

	// Inventory handler
	bus.Subscribe("order.created", func(ctx context.Context, e *event.Event) error {
		orderState.Lock()
		defer orderState.Unlock()
		orderState.inventoryReserved = true

		// Publish next event in pipeline
		bus.Publish(event.NewEvent("order.inventory_reserved", e.Data))
		return nil
	})

	// Payment handler
	bus.Subscribe("order.inventory_reserved", func(ctx context.Context, e *event.Event) error {
		time.Sleep(50 * time.Millisecond) // Simulate payment processing

		orderState.Lock()
		defer orderState.Unlock()
		orderState.paymentProcessed = true

		// Publish next event
		bus.Publish(event.NewEvent("order.paid", e.Data))
		return nil
	})

	// Shipping handler
	bus.Subscribe("order.paid", func(ctx context.Context, e *event.Event) error {
		orderState.Lock()
		defer orderState.Unlock()
		orderState.shippingScheduled = true

		// Publish next event
		bus.Publish(event.NewEvent("order.shipped", e.Data))
		return nil
	})

	// Email notification handler
	bus.Subscribe("order.shipped", func(ctx context.Context, e *event.Event) error {
		orderState.Lock()
		defer orderState.Unlock()
		orderState.emailSent = true
		return nil
	})

	// Start the pipeline
	orderData := map[string]interface{}{
		"order_id": "order_123",
		"amount":   199.99,
	}

	bus.Publish(event.NewEvent("order.created", orderData))

	// Wait for pipeline to complete
	time.Sleep(500 * time.Millisecond)

	orderState.Lock()
	defer orderState.Unlock()

	assert.True(t, orderState.inventoryReserved)
	assert.True(t, orderState.paymentProcessed)
	assert.True(t, orderState.shippingScheduled)
	assert.True(t, orderState.emailSent)
}

func TestE2E_ErrorHandlingAndRetry(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	attempts := int32(0)
	successCount := int32(0)

	// Handler that fails first 2 times, then succeeds
	failingHandler := func(ctx context.Context, e *event.Event) error {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			return errors.New("temporary failure")
		}
		atomic.AddInt32(&successCount, 1)
		return nil
	}

	// Wrap with retry middleware
	retryHandler := event.WithRetry(3, 50*time.Millisecond)(failingHandler)

	bus.Subscribe("critical.task", retryHandler)

	bus.Publish(event.NewEvent("critical.task", nil))

	time.Sleep(300 * time.Millisecond)

	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	assert.Equal(t, int32(1), atomic.LoadInt32(&successCount))
}

func TestE2E_FilteredHandlersWithComplexLogic(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	highPriorityCount := int32(0)
	apiOnlyCount := int32(0)
	vipCustomerCount := int32(0)

	// High priority filter
	highPriorityFilter := event.FilterByMetadata("priority", "high")
	bus.Subscribe("order.created", event.NewFilteredHandler(highPriorityFilter, func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&highPriorityCount, 1)
		return nil
	}))

	// API-only filter
	apiFilter := event.FilterByMetadata("source", "api")
	bus.Subscribe("order.created", event.NewFilteredHandler(apiFilter, func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&apiOnlyCount, 1)
		return nil
	}))

	// VIP customer filter (complex)
	vipFilter := event.And(
		event.FilterByMetadata("customer_type", "vip"),
		event.Or(
			event.FilterByMetadata("priority", "high"),
			event.FilterByMetadata("amount_threshold", "exceeded"),
		),
	)
	bus.Subscribe("order.created", event.NewFilteredHandler(vipFilter, func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&vipCustomerCount, 1)
		return nil
	}))

	// Publish various orders
	orders := []struct {
		metadata map[string]interface{}
	}{
		{map[string]interface{}{"priority": "high", "source": "api"}},                                     // high+api
		{map[string]interface{}{"priority": "low", "source": "web"}},                                      // none
		{map[string]interface{}{"priority": "high", "source": "web"}},                                     // high only
		{map[string]interface{}{"customer_type": "vip", "priority": "high", "source": "api"}},             // all
		{map[string]interface{}{"customer_type": "vip", "amount_threshold": "exceeded", "source": "api"}}, // vip+api
		{map[string]interface{}{"customer_type": "regular", "priority": "high", "source": "api"}},         // high+api
	}

	for _, order := range orders {
		evt := event.NewEvent("order.created", nil)
		for k, v := range order.metadata {
			evt.WithMetadata(k, v)
		}
		bus.Publish(evt)
	}

	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(4), atomic.LoadInt32(&highPriorityCount), "4 high priority orders")
	assert.Equal(t, int32(4), atomic.LoadInt32(&apiOnlyCount), "4 API orders")
	assert.Equal(t, int32(2), atomic.LoadInt32(&vipCustomerCount), "2 VIP customer orders")
}

func TestE2E_BatchProcessingWithMultipleEventTypes(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	var mu sync.Mutex
	emailBatches := make([]int, 0)
	smsBatches := make([]int, 0)

	// Email batch processor
	emailConfig := event.BatchConfig{
		BatchSize:   5,
		FlushPeriod: 200 * time.Millisecond,
	}
	emailProcessor := event.NewBatchProcessor("notification.email", emailConfig, func(ctx context.Context, events []*event.Event) error {
		mu.Lock()
		defer mu.Unlock()
		emailBatches = append(emailBatches, len(events))
		return nil
	})
	defer emailProcessor.Shutdown()

	// SMS batch processor
	smsConfig := event.BatchConfig{
		BatchSize:   3,
		FlushPeriod: 200 * time.Millisecond,
	}
	smsProcessor := event.NewBatchProcessor("notification.sms", smsConfig, func(ctx context.Context, events []*event.Event) error {
		mu.Lock()
		defer mu.Unlock()
		smsBatches = append(smsBatches, len(events))
		return nil
	})
	defer smsProcessor.Shutdown()

	bus.Subscribe("notification.email", emailProcessor.AsEventHandler())
	bus.Subscribe("notification.sms", smsProcessor.AsEventHandler())

	// Send 12 email notifications and 7 SMS notifications
	for i := 0; i < 12; i++ {
		bus.Publish(event.NewEvent("notification.email", map[string]interface{}{
			"to":      fmt.Sprintf("user%d@example.com", i),
			"subject": "Update",
		}))
	}

	for i := 0; i < 7; i++ {
		bus.Publish(event.NewEvent("notification.sms", map[string]interface{}{
			"to":      fmt.Sprintf("+1234567%04d", i),
			"message": "Alert",
		}))
	}

	time.Sleep(400 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Email: 2 batches of 5, then 1 batch of 2 (or pending)
	require.GreaterOrEqual(t, len(emailBatches), 2)
	assert.Equal(t, 5, emailBatches[0])
	assert.Equal(t, 5, emailBatches[1])

	// SMS: 2 batches of 3, then 1 batch of 1 (or pending)
	require.GreaterOrEqual(t, len(smsBatches), 2)
	assert.Equal(t, 3, smsBatches[0])
	assert.Equal(t, 3, smsBatches[1])
}

func TestE2E_MiddlewareChainWithComplexScenario(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	attempts := int32(0)
	timeouts := int32(0)
	successes := int32(0)

	// Complex handler that simulates various scenarios
	complexHandler := func(ctx context.Context, e *event.Event) error {
		count := atomic.AddInt32(&attempts, 1)

		// First attempt: timeout
		if count == 1 {
			time.Sleep(200 * time.Millisecond)
			atomic.AddInt32(&timeouts, 1)
			return context.DeadlineExceeded
		}

		// Second attempt: error
		if count == 2 {
			return errors.New("processing error")
		}

		// Third attempt: success
		atomic.AddInt32(&successes, 1)
		return nil
	}

	// Chain middleware: recovery -> retry -> timeout -> logging
	wrappedHandler := event.Chain(
		event.WithRecovery(),
		event.WithRetry(3, 50*time.Millisecond),
		event.WithTimeout(100*time.Millisecond),
		event.WithLogging(),
	)(complexHandler)

	bus.Subscribe("complex.task", wrappedHandler)

	bus.Publish(event.NewEvent("complex.task", nil))

	time.Sleep(1 * time.Second)

	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
	assert.Equal(t, int32(1), atomic.LoadInt32(&successes))
}

func TestE2E_HighVolumeStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	bus := event.NewEventBus(&event.Config{
		Workers:    20,
		BufferSize: 5000,
	})
	defer bus.Shutdown()

	processed := int32(0)

	// Fast handler
	bus.Subscribe("high.volume", func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&processed, 1)
		return nil
	})

	// Publish 10,000 events
	numEvents := 10000

	start := time.Now()
	for i := 0; i < numEvents; i++ {
		bus.Publish(event.NewEvent("high.volume", i))
	}
	elapsed := time.Since(start)

	// Publishing should be fast
	assert.Less(t, elapsed, 1*time.Second, "Publishing should be very fast")

	// Wait for processing
	time.Sleep(2 * time.Second)

	count := atomic.LoadInt32(&processed)
	assert.Equal(t, int32(numEvents), count, "All events should be processed")
}

func TestE2E_GracefulShutdownUnderLoad(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())

	processed := int32(0)

	bus.Subscribe("test.event", func(ctx context.Context, e *event.Event) error {
		time.Sleep(10 * time.Millisecond) // Simulate work
		atomic.AddInt32(&processed, 1)
		return nil
	})

	// Publish many events
	for i := 0; i < 100; i++ {
		bus.Publish(event.NewEvent("test.event", i))
	}

	// Give it a moment to start processing
	time.Sleep(50 * time.Millisecond)

	// Shutdown should wait for all in-flight events
	shutdownStart := time.Now()
	bus.Shutdown()
	shutdownDuration := time.Since(shutdownStart)

	// Verify some events were processed
	count := atomic.LoadInt32(&processed)
	assert.Greater(t, count, int32(0), "Some events should be processed")

	// Shutdown should take some time (waiting for workers)
	assert.Greater(t, shutdownDuration, 0*time.Millisecond)

	t.Logf("Processed %d/%d events before shutdown", count, 100)
}

func TestE2E_EventChaining(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	var mu sync.Mutex
	eventChain := make([]string, 0)

	// Event A triggers Event B
	bus.Subscribe("event.a", func(ctx context.Context, e *event.Event) error {
		mu.Lock()
		eventChain = append(eventChain, "A")
		mu.Unlock()

		bus.Publish(event.NewEvent("event.b", nil))
		return nil
	})

	// Event B triggers Event C
	bus.Subscribe("event.b", func(ctx context.Context, e *event.Event) error {
		mu.Lock()
		eventChain = append(eventChain, "B")
		mu.Unlock()

		bus.Publish(event.NewEvent("event.c", nil))
		return nil
	})

	// Event C is final
	bus.Subscribe("event.c", func(ctx context.Context, e *event.Event) error {
		mu.Lock()
		eventChain = append(eventChain, "C")
		mu.Unlock()
		return nil
	})

	// Start chain
	bus.Publish(event.NewEvent("event.a", nil))

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, eventChain, 3)
	assert.Equal(t, []string{"A", "B", "C"}, eventChain)
}
