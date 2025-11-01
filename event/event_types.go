// Package event. event_types provides a strongly-typed event bus system for decoupling core logic from side effects.
package event

import "time"

// EventType represents a strongly-typed event identifier to prevent typos and ensure consistency.
type EventType string

// Event represents a generic event that flows through the system.
type Event struct {
	Type      EventType      // Strongly-typed event identifier
	Timestamp time.Time      // When the event occurred
	Data      any            // Event-specific payload
	Metadata  map[string]any // Additional context (e.g., request ID, user agent)
}

// NewEvent creates a new event with the given type and data.
func NewEvent(eventType EventType, data any) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]any),
	}
}

// WithMetadata adds metadata to the event and returns the event for chaining.
func (e *Event) WithMetadata(key string, value any) *Event {
	e.Metadata[key] = value
	return e
}

// -------------------------------------------- Event Data Structs --------------------------------------------

// UserCreatedData contains information about a newly created user.
type UserCreatedData struct {
	UserID       string
	Username     string
	Email        string
	Organization string
}

// LoginSuccessData contains information about a successful login.
type LoginSuccessData struct {
	UserID       string
	Username     string
	Organization string
	IPAddress    string
	UserAgent    string
}

// LoginFailedData contains information about a failed login attempt.
type LoginFailedData struct {
	Username  string
	IPAddress string
	Reason    string
}

// TokenRevokedData contains information about a revoked token.
type TokenRevokedData struct {
	TokenID   string
	UserID    string
	RevokedBy string
	Reason    string
}

// UserUpdatedData contains information about user profile updates.
type UserUpdatedData struct {
	UserID        string
	Username      string
	ChangedFields []string
	UpdatedBy     string
}

// Event type constants for different domains in the system.
// This is intended for example you can extend it as you want.
// Simply wrap your string in EventType.
//
//goland:noinspection GoCommentStart
const (
	// Authentication Events
	AuthEventUserCreated     EventType = "auth.user.created"
	AuthEventLoginSuccess    EventType = "auth.login.success"
	AuthEventLoginFailed     EventType = "auth.login.failed"
	AuthEventLogout          EventType = "auth.logout"
	AuthEventPasswordChanged EventType = "auth.password.changed"
	AuthEventTokenRevoked    EventType = "auth.token.revoked"

	// User Management Events
	UserEventUpdated  EventType = "user.updated"
	UserEventDeleted  EventType = "user.deleted"
	UserEventEnabled  EventType = "user.enabled"
	UserEventDisabled EventType = "user.disabled"

	// Organization Events
	OrgEventCreated EventType = "org.created"
	OrgEventUpdated EventType = "org.updated"
	OrgEventDeleted EventType = "org.deleted"

	// Permission Events
	PermissionEventGranted EventType = "permission.granted"
	PermissionEventRevoked EventType = "permission.revoked"
)
