// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package event. event_bus provides the core pub/sub event bus implementation.
package event

import (
	"context"
	"fmt"
	"sync"
)

// EventHandler is a function that processes events.
type EventHandler func(ctx context.Context, event *Event) error

// Subscription represents a registered event handler.
type Subscription struct {
	ID      string
	Type    EventType
	Handler EventHandler
}

// EventBus manages event publishing and subscription.
type EventBus struct {
	mu            sync.RWMutex
	subscriptions map[EventType][]*Subscription
	workers       int
	eventChan     chan *Event
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// Config holds configuration for the event bus.
type Config struct {
	Workers    int // Number of worker goroutines
	BufferSize int // Size of the event channel buffer
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() *Config {
	return &Config{
		Workers:    10,
		BufferSize: 1000,
	}
}

// NewEventBus creates a new event bus with the given configuration.
func NewEventBus(cfg *Config) *EventBus {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	bus := &EventBus{
		subscriptions: make(map[EventType][]*Subscription),
		workers:       cfg.Workers,
		eventChan:     make(chan *Event, cfg.BufferSize),
		ctx:           ctx,
		cancel:        cancel,
	}

	bus.start()
	return bus
}

// -------------------------------------------- Public Functions --------------------------------------------

// Subscribe registers a handler for a specific event type and returns a subscription ID.
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subscriptionID := fmt.Sprintf("%s-%d", eventType, len(eb.subscriptions[eventType]))

	sub := &Subscription{
		ID:      subscriptionID,
		Type:    eventType,
		Handler: handler,
	}

	eb.subscriptions[eventType] = append(eb.subscriptions[eventType], sub)

	return subscriptionID
}

// Unsubscribe removes a subscription by ID.
func (eb *EventBus) Unsubscribe(subscriptionID string) bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for eventType, subs := range eb.subscriptions {
		for i, sub := range subs {
			if sub.ID == subscriptionID {
				// Unsubscribed handler
				eb.subscriptions[eventType] = append(subs[:i], subs[i+1:]...)
				return true
			}
		}
	}

	return false
}

// Publish sends an event to all registered handlers asynchronously.
func (eb *EventBus) Publish(event *Event) {
	select {
	case eb.eventChan <- event:

	case <-eb.ctx.Done():

	default:
		// Event channel full, dropping event
	}
}

// PublishSync sends an event to all registered handlers synchronously.
func (eb *EventBus) PublishSync(ctx context.Context, event *Event) []error {
	eb.mu.RLock()
	handlers := eb.subscriptions[event.Type]
	eb.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	errors := make([]error, 0)

	for _, sub := range handlers {
		if err := sub.Handler(ctx, event); err != nil {
			errors = append(errors, fmt.Errorf("handler %s: %w", sub.ID, err))
		}
	}

	return errors
}

// Shutdown gracefully stops the event bus and waits for all workers to finish.
func (eb *EventBus) Shutdown() {
	eb.cancel()
	close(eb.eventChan)
	eb.wg.Wait()
}

// -------------------------------------------- Private Helper Functions --------------------------------------------

// start initializes worker goroutines to process events.
func (eb *EventBus) start() {
	for i := 0; i < eb.workers; i++ {
		eb.wg.Add(1)
		go eb.worker()
	}
}

// worker processes events from the channel.
func (eb *EventBus) worker() {
	defer eb.wg.Done()

	for {
		select {
		case event, ok := <-eb.eventChan:
			if !ok {
				return // Worker stopped
			}
			eb.dispatch(event)

		case <-eb.ctx.Done():
			return // Worker shutting down
		}
	}
}

// dispatch sends an event to all registered handlers.
func (eb *EventBus) dispatch(event *Event) {
	eb.mu.RLock()
	handlers := eb.subscriptions[event.Type]
	eb.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	for _, sub := range handlers {
		_ = sub.Handler(eb.ctx, event)
	}
}
