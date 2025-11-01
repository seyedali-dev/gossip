// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package event. global provides a singleton event bus instance for application-wide use.
package event

import "sync"

var (
	globalBus *EventBus
	once      sync.Once
)

// -------------------------------------------- Public Functions --------------------------------------------

// GetGlobalBus returns the singleton event bus instance.
func GetGlobalBus() *EventBus {
	once.Do(func() {
		globalBus = NewEventBus(DefaultConfig())
	})
	return globalBus
}

// InitGlobalBus initializes the global event bus with custom configuration.
func InitGlobalBus(cfg *Config) {
	once.Do(func() {
		globalBus = NewEventBus(cfg)
	})
}

// ShutdownGlobalBus shuts down the global event bus.
func ShutdownGlobalBus() {
	GetGlobalBus().Shutdown()
}

// Publish is a convenience function to publish events to the global bus.
func Publish(event *Event) {
	GetGlobalBus().Publish(event)
}

// Subscribe is a convenience function to subscribe to the global bus.
func Subscribe(eventType EventType, handler EventHandler) string {
	return GetGlobalBus().Subscribe(eventType, handler)
}
