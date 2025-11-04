// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package main. main demonstrates cross-service communication using gossip event bus
package main

import (
	"context"
	"log"
	"time"

	gossip "github.com/seyedali-dev/gossip/event"
)

// Cross-service event types.
const (
	UserServiceUserCreated  gossip.EventType = "user_service.user.created"
	UserServiceUserUpdated  gossip.EventType = "user_service.user.updated"
	EmailServiceSent        gossip.EventType = "email_service.email.sent"
	NotificationServiceSent gossip.EventType = "notification_service.sent"
)

// -------------------------------------------- Event Data Structs --------------------------------------------

// UserCreatedEvent contains user creation information.
type UserCreatedEvent struct {
	UserID   string
	Email    string
	Username string
}

// -------------------------------------------- User Service --------------------------------------------

// UserService manages users and publishes events.
type UserService struct {
	bus *gossip.EventBus
}

// NewUserService creates a user service.
func NewUserService(bus *gossip.EventBus) *UserService {
	return &UserService{bus: bus}
}

// CreateUser creates a user and notifies other services.
func (s *UserService) CreateUser(email, username string) error {
	userID := "user_" + username

	log.Printf("[UserService] Creating user: %s", username)

	// Business logic here...

	// Publish event for other services
	event := gossip.NewEvent(UserServiceUserCreated, &UserCreatedEvent{
		UserID:   userID,
		Email:    email,
		Username: username,
	})

	s.bus.Publish(event)
	return nil
}

// -------------------------------------------- Email Service --------------------------------------------

// EmailService listens to user events and sends emails.
type EmailService struct {
	bus *gossip.EventBus
}

// NewEmailService creates an email service.
func NewEmailService(bus *gossip.EventBus) *EmailService {
	svc := &EmailService{bus: bus}
	svc.registerHandlers()
	return svc
}

// registerHandlers subscribes to relevant events.
func (s *EmailService) registerHandlers() {
	s.bus.Subscribe(UserServiceUserCreated, s.handleUserCreated)
}

// handleUserCreated sends welcome email to new users.
func (s *EmailService) handleUserCreated(ctx context.Context, event *gossip.Event) error {
	data := event.Data.(*UserCreatedEvent)

	log.Printf("[EmailService] Sending welcome email to %s", data.Email)

	// Send email logic...
	time.Sleep(50 * time.Millisecond)

	// Publish email sent event
	s.bus.Publish(gossip.NewEvent(EmailServiceSent, map[string]string{
		"to":      data.Email,
		"subject": "Welcome!",
	}))

	return nil
}

// -------------------------------------------- Notification Service --------------------------------------------

// NotificationService sends push notifications.
type NotificationService struct {
	bus *gossip.EventBus
}

// NewNotificationService creates a notification service.
func NewNotificationService(bus *gossip.EventBus) *NotificationService {
	svc := &NotificationService{bus: bus}
	svc.registerHandlers()
	return svc
}

// registerHandlers subscribes to relevant events.
func (s *NotificationService) registerHandlers() {
	s.bus.Subscribe(UserServiceUserCreated, s.handleUserCreated)
}

// handleUserCreated sends push notification to new users.
func (s *NotificationService) handleUserCreated(ctx context.Context, event *gossip.Event) error {
	data := event.Data.(*UserCreatedEvent)

	log.Printf("[NotificationService] Sending push notification to user %s", data.UserID)

	// Send notification logic...

	return nil
}

// -------------------------------------------- Analytics Service --------------------------------------------

// AnalyticsService tracks user metrics.
type AnalyticsService struct {
	bus *gossip.EventBus
}

// NewAnalyticsService creates an analytics service.
func NewAnalyticsService(bus *gossip.EventBus) *AnalyticsService {
	svc := &AnalyticsService{bus: bus}
	svc.registerHandlers()
	return svc
}

// registerHandlers subscribes to relevant events.
func (s *AnalyticsService) registerHandlers() {
	// Track all user events
	s.bus.Subscribe(UserServiceUserCreated, s.trackEvent)
	s.bus.Subscribe(UserServiceUserUpdated, s.trackEvent)

	// Track all email events
	s.bus.Subscribe(EmailServiceSent, s.trackEvent)
}

// trackEvent records events for analytics.
func (s *AnalyticsService) trackEvent(ctx context.Context, event *gossip.Event) error {
	log.Printf("[AnalyticsService] Tracking event: %s at %s", event.Type, event.Timestamp.Format(time.RFC3339))
	return nil
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	// Shared event bus across services
	config := &gossip.Config{
		Workers:    10,
		BufferSize: 1000,
	}
	bus := gossip.NewEventBus(config)
	defer bus.Shutdown()

	// Initialize services - they self-register handlers
	userService := NewUserService(bus)
	_ = NewEmailService(bus)
	_ = NewNotificationService(bus)
	_ = NewAnalyticsService(bus)

	log.Println("=== Starting Microservices Demo ===")

	// User service creates users, other services react automatically
	userService.CreateUser("alice@example.com", "alice")
	time.Sleep(200 * time.Millisecond)

	userService.CreateUser("bob@example.com", "bob")
	time.Sleep(200 * time.Millisecond)

	log.Println("=== Demo Complete ===")
	time.Sleep(500 * time.Millisecond)
}
