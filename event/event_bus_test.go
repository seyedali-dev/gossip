// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package event_test. event_bus_test provides comprehensive tests for the event bus implementation
package event_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seyedali-dev/gossip/event"
)

// -------------------------------------------- Test Functions --------------------------------------------

func TestEventBus_PublishSubscribe(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	received := int32(0)
	handler := func(ctx context.Context, event *event.Event) error {
		atomic.AddInt32(&received, 1)
		return nil
	}

	bus.Subscribe(event.AuthEventLoginSuccess, handler)

	event := event.NewEvent(event.AuthEventLoginSuccess, &event.LoginSuccessData{
		UserID:   "test-user",
		Username: "testuser",
	})

	bus.Publish(event)

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&received) != 1 {
		t.Errorf("Expected 1 event, got %d", received)
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	counter := int32(0)

	for i := 0; i < 5; i++ {
		bus.Subscribe(event.AuthEventUserCreated, func(ctx context.Context, event *event.Event) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})
	}

	event := event.NewEvent(event.AuthEventUserCreated, &event.UserCreatedData{
		UserID:   "test-user",
		Username: "testuser",
	})

	bus.Publish(event)

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&counter) != 5 {
		t.Errorf("Expected 5 invocations, got %d", counter)
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	counter := int32(0)
	handler := func(ctx context.Context, event *event.Event) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}

	subID := bus.Subscribe(event.AuthEventLoginSuccess, handler)

	event := event.NewEvent(event.AuthEventLoginSuccess, &event.LoginSuccessData{})
	bus.Publish(event)
	time.Sleep(50 * time.Millisecond)

	bus.Unsubscribe(subID)

	bus.Publish(event)
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&counter) != 1 {
		t.Errorf("Expected 1 invocation, got %d", counter)
	}
}

func TestEventBus_SynchronousPublish(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	executed := false
	handler := func(ctx context.Context, event *event.Event) error {
		executed = true
		return nil
	}

	bus.Subscribe(event.AuthEventPasswordChanged, handler)

	event := event.NewEvent(event.AuthEventPasswordChanged, nil)
	errors := bus.PublishSync(context.Background(), event)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	if !executed {
		t.Error("Handler was not executed synchronously")
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	counter := int32(0)
	handler := func(ctx context.Context, event *event.Event) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}

	bus.Subscribe(event.AuthEventLoginSuccess, handler)

	var wg sync.WaitGroup
	numEvents := 100

	for i := 0; i < numEvents; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := event.NewEvent(event.AuthEventLoginSuccess, &event.LoginSuccessData{})
			bus.Publish(event)
		}()
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt32(&counter) != int32(numEvents) {
		t.Errorf("Expected %d events, got %d", numEvents, counter)
	}
}

func TestEventBus_GracefulShutdown(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())

	counter := int32(0)
	handler := func(ctx context.Context, event *event.Event) error {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
		return nil
	}

	bus.Subscribe(event.AuthEventUserCreated, handler)

	for i := 0; i < 10; i++ {
		event := event.NewEvent(event.AuthEventUserCreated, &event.UserCreatedData{})
		bus.Publish(event)
	}

	bus.Shutdown()

	if atomic.LoadInt32(&counter) == 0 {
		t.Error("No events were processed before shutdown")
	}
}

func TestEventBus_EventMetadata(t *testing.T) {
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	receivedMetadata := make(map[string]interface{})
	handler := func(ctx context.Context, event *event.Event) error {
		for k, v := range event.Metadata {
			receivedMetadata[k] = v
		}
		return nil
	}

	bus.Subscribe(event.AuthEventLoginSuccess, handler)

	event := event.NewEvent(event.AuthEventLoginSuccess, &event.LoginSuccessData{}).
		WithMetadata("request_id", "12345").
		WithMetadata("user_agent", "test")

	bus.Publish(event)
	time.Sleep(100 * time.Millisecond)

	if receivedMetadata["request_id"] != "12345" {
		t.Error("Metadata was not properly passed to handler")
	}
}
