// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package testdata provides test data seeding utilities for integration tests.
package testdata

import (
	"context"
	"database/sql"
	"fmt"
)

// EventLog represents an event stored in the database for audit/replay.
type EventLog struct {
	ID        int64
	EventType string
	Payload   []byte
	Metadata  []byte
	CreatedAt string
}

// -------------------------------------------- Public Functions --------------------------------------------

// SeedEventLogTable creates the event_log table for testing event persistence.
func SeedEventLogTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS event_log (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(255) NOT NULL,
			payload JSONB,
			metadata JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create event_log table: %w", err)
	}

	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_event_type ON event_log (event_type)"); err != nil {
		return fmt.Errorf("failed to create index on event_log: %w", err)
	}

	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_created_at ON event_log (created_at)"); err != nil {
		return fmt.Errorf("failed to create index on event_log: %w", err)
	}

	return nil
}

// SeedUserEventsTable creates a table for tracking user-related events.
func SeedUserEventsTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_events (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(255) NOT NULL,
			occurred_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create user_events table: %w", err)
	}

	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_user_id ON user_events (user_id)"); err != nil {
		return fmt.Errorf("failed to create index on user_events: %w", err)
	}

	if _, err := db.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_event_type ON user_events (event_type)"); err != nil {
		return fmt.Errorf("failed to create index on user_events: %w", err)
	}

	return nil
}

// SeedMetricsTable creates a table for storing event metrics.
func SeedMetricsTable(ctx context.Context, db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS event_metrics (
			id SERIAL PRIMARY KEY,
			event_type VARCHAR(255) NOT NULL,
			count INTEGER DEFAULT 0,
			last_occurred TIMESTAMP,
			UNIQUE(event_type)
		)
	`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create event_metrics table: %w", err)
	}

	return nil
}

// CleanupTables removes all test tables.
func CleanupTables(ctx context.Context, db *sql.DB) error {
	tables := []string{"event_log", "user_events", "event_metrics"}

	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// InsertSampleEvents inserts sample events for testing queries.
func InsertSampleEvents(ctx context.Context, db *sql.DB) error {
	events := []struct {
		eventType string
		payload   string
		metadata  string
	}{
		{"user.created", `{"user_id": "user_1", "email": "alice@example.com"}`, `{"source": "api"}`},
		{"user.created", `{"user_id": "user_2", "email": "bob@example.com"}`, `{"source": "web"}`},
		{"user.login", `{"user_id": "user_1", "ip": "192.168.1.1"}`, `{"device": "mobile"}`},
		{"order.created", `{"order_id": "order_1", "amount": 99.99}`, `{"priority": "high"}`},
		{"order.paid", `{"order_id": "order_1", "payment_method": "card"}`, `{"priority": "high"}`},
	}

	for _, e := range events {
		query := `INSERT INTO event_log (event_type, payload, metadata) VALUES ($1, $2::jsonb, $3::jsonb)`
		if _, err := db.ExecContext(ctx, query, e.eventType, e.payload, e.metadata); err != nil {
			return fmt.Errorf("failed to insert event %s: %w", e.eventType, err)
		}
	}

	return nil
}
