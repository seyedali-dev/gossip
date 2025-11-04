// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package main. main demonstrates authentication service with event-driven architecture.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	gossip "github.com/seyedali-dev/gossip/event"
)

// Custom event types for auth service.
const (
	UserRegistered   gossip.EventType = "auth.user.registered"
	UserLoggedIn     gossip.EventType = "auth.user.logged_in"
	UserLoggedOut    gossip.EventType = "auth.user.logged_out"
	PasswordChanged  gossip.EventType = "auth.password.changed"
	AccountLocked    gossip.EventType = "auth.account.locked"
	TwoFactorEnabled gossip.EventType = "auth.2fa.enabled"
)

// -------------------------------------------- Event Data Structs --------------------------------------------

// UserRegisteredData contains user registration information.
type UserRegisteredData struct {
	UserID    string
	Email     string
	Username  string
	IPAddress string
}

// LoginData contains login information.
type LoginData struct {
	UserID    string
	Username  string
	IPAddress string
	UserAgent string
	Success   bool
}

// PasswordChangedData contains password change information.
type PasswordChangedData struct {
	UserID string
	Method string // "reset" or "change"
}

// -------------------------------------------- Handlers --------------------------------------------

// emailNotificationHandler sends email notifications for auth events.
func emailNotificationHandler(ctx context.Context, event *gossip.Event) error {
	switch event.Type {
	case UserRegistered:
		data := event.Data.(*UserRegisteredData)
		log.Printf("[Email] Sending welcome email to %s (%s)", data.Username, data.Email)

	case UserLoggedIn:
		data := event.Data.(*LoginData)
		log.Printf("[Email] Login notification sent to user %s from IP %s", data.Username, data.IPAddress)

	case PasswordChanged:
		data := event.Data.(*PasswordChangedData)
		log.Printf("[Email] Password change alert sent to user %s", data.UserID)

	case AccountLocked:
		log.Printf("[Email] Account locked notification sent")
	}

	return nil
}

// auditLogHandler records all auth events to audit log.
func auditLogHandler(ctx context.Context, event *gossip.Event) error {
	log.Printf("[Audit] Event: %s | Timestamp: %s | Data: %+v",
		event.Type,
		event.Timestamp.Format(time.RFC3339),
		event.Data,
	)
	return nil
}

// metricsHandler tracks authentication metrics.
func metricsHandler(ctx context.Context, event *gossip.Event) error {
	log.Printf("[Metrics] Incrementing counter: %s", event.Type)
	return nil
}

// securityHandler handles security-related logic.
func securityHandler(ctx context.Context, event *gossip.Event) error {
	switch event.Type {
	case UserLoggedIn:
		data := event.Data.(*LoginData)
		if !data.Success {
			log.Printf("[Security] Failed login attempt from IP: %s", data.IPAddress)
			// TODO: Implement rate limiting or IP blocking
		}

	case AccountLocked:
		log.Printf("[Security] Account locked - triggering security review")
	}

	return nil
}

// -------------------------------------------- Service Logic --------------------------------------------

// AuthService handles user authentication.
type AuthService struct {
	bus *gossip.EventBus
}

// NewAuthService creates a new authentication service.
func NewAuthService(bus *gossip.EventBus) *AuthService {
	return &AuthService{bus: bus}
}

// RegisterUser registers a new user and publishes an event.
func (s *AuthService) RegisterUser(email, username, ipAddress string) error {
	userID := fmt.Sprintf("user_%d", time.Now().Unix())

	// Core registration logic here...
	log.Printf("[Service] Registering user: %s", username)

	// Publish event
	eventData := &UserRegisteredData{
		UserID:    userID,
		Email:     email,
		Username:  username,
		IPAddress: ipAddress,
	}

	event := gossip.NewEvent(UserRegistered, eventData).
		WithMetadata("source", "api").
		WithMetadata("version", "v1")

	s.bus.Publish(event)

	return nil
}

// Login authenticates a user and publishes an event.
func (s *AuthService) Login(username, password, ipAddress, userAgent string) error {
	// Authentication logic here...
	success := true // Assume success for demo
	userID := "user_123"

	log.Printf("[Service] User %s logged in from %s", username, ipAddress)

	// Publish event
	eventData := &LoginData{
		UserID:    userID,
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Success:   success,
	}

	event := gossip.NewEvent(UserLoggedIn, eventData).
		WithMetadata("session_id", "sess_abc123")

	s.bus.Publish(event)

	return nil
}

// ChangePassword changes user password and publishes an event.
func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	// Password change logic here...
	log.Printf("[Service] Password changed for user: %s", userID)

	// Publish event
	eventData := &PasswordChangedData{
		UserID: userID,
		Method: "change",
	}

	event := gossip.NewEvent(PasswordChanged, eventData)
	s.bus.Publish(event)

	return nil
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	// Initialize event bus
	config := &gossip.Config{
		Workers:    10,
		BufferSize: 1000,
	}
	bus := gossip.NewEventBus(config)
	defer bus.Shutdown()

	// Register handlers with middleware
	bus.Subscribe(
		UserRegistered,
		gossip.Chain(
			gossip.WithLogging(),
			gossip.WithTimeout(5*time.Second),
			gossip.WithRecovery(),
		)(emailNotificationHandler),
	)

	bus.Subscribe(UserRegistered, auditLogHandler)
	bus.Subscribe(UserRegistered, metricsHandler)

	bus.Subscribe(UserLoggedIn, emailNotificationHandler)
	bus.Subscribe(UserLoggedIn, auditLogHandler)
	bus.Subscribe(UserLoggedIn, metricsHandler)
	bus.Subscribe(UserLoggedIn, securityHandler)

	bus.Subscribe(PasswordChanged, emailNotificationHandler)
	bus.Subscribe(PasswordChanged, auditLogHandler)

	// Create service
	authService := NewAuthService(bus)

	// Simulate user activities
	log.Println("=== Starting Auth Service Demo ===")

	authService.RegisterUser("john@example.com", "john_doe", "192.168.1.100")
	time.Sleep(100 * time.Millisecond)

	authService.Login("john_doe", "password123", "192.168.1.100", "Mozilla/5.0")
	time.Sleep(100 * time.Millisecond)

	authService.ChangePassword("user_123", "oldpass", "newpass")
	time.Sleep(100 * time.Millisecond)

	log.Println("=== Demo Complete ===")
	time.Sleep(500 * time.Millisecond)
}
