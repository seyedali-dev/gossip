// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

//go:build integration
// +build integration

package event_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seyedali-dev/gossip/event"
	"github.com/seyedali-dev/gossip/event/testdata"
	"github.com/seyedali-dev/gossip/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------- Database Integration Tests --------------------------------------------

func TestIntegration_EventPersistence(t *testing.T) {
	ctx := context.Background()

	tc, err := tests.SetupTestContainer(ctx)
	require.NoError(t, err)
	defer tc.Cleanup(ctx)

	// Seed database
	err = testdata.SeedEventLogTable(ctx, tc.DB)
	require.NoError(t, err)

	// Create event bus
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	// Create handler that persists events to database
	persistHandler := createPersistenceHandler(tc.DB)
	bus.Subscribe("user.created", persistHandler)

	// Publish events
	for i := 0; i < 5; i++ {
		data := map[string]interface{}{
			"user_id":  fmt.Sprintf("user_%d", i),
			"username": fmt.Sprintf("user%d", i),
			"email":    fmt.Sprintf("user%d@example.com", i),
		}

		evt := event.NewEvent("user.created", data).
			WithMetadata("source", "test")

		bus.Publish(evt)
	}

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify events were persisted
	var count int
	err = tc.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM event_log WHERE event_type = $1", "user.created").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 5, count, "All events should be persisted")
}

func TestIntegration_EventReplay(t *testing.T) {
	ctx := context.Background()

	tc, err := tests.SetupTestContainer(ctx)
	require.NoError(t, err)
	defer tc.Cleanup(ctx)

	// Seed database
	err = testdata.SeedEventLogTable(ctx, tc.DB)
	require.NoError(t, err)

	err = testdata.InsertSampleEvents(ctx, tc.DB)
	require.NoError(t, err)

	// Create event bus
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	replayCount := int32(0)
	bus.Subscribe("user.created", func(ctx context.Context, e *event.Event) error {
		atomic.AddInt32(&replayCount, 1)
		return nil
	})

	// Fetch and replay events from database
	rows, err := tc.DB.QueryContext(ctx, "SELECT event_type, payload, metadata FROM event_log WHERE event_type = $1", "user.created")
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var eventType string
		var payload, metadata []byte

		err := rows.Scan(&eventType, &payload, &metadata)
		require.NoError(t, err)

		var data map[string]interface{}
		err = json.Unmarshal(payload, &data)
		require.NoError(t, err)

		var meta map[string]interface{}
		err = json.Unmarshal(metadata, &meta)
		require.NoError(t, err)

		evt := event.NewEvent(event.EventType(eventType), data)
		for k, v := range meta {
			evt.WithMetadata(k, v)
		}

		bus.Publish(evt)
	}

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(2), atomic.LoadInt32(&replayCount), "Should replay 2 user.created events")
}

func TestIntegration_MetricsTracking(t *testing.T) {
	ctx := context.Background()

	tc, err := tests.SetupTestContainer(ctx)
	require.NoError(t, err)
	defer tc.Cleanup(ctx)

	// Seed database
	err = testdata.SeedMetricsTable(ctx, tc.DB)
	require.NoError(t, err)

	// Create event bus
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	// Metrics handler
	metricsHandler := createMetricsHandler(tc.DB)
	bus.Subscribe("user.created", metricsHandler)
	bus.Subscribe("order.created", metricsHandler)

	// Publish events
	for i := 0; i < 10; i++ {
		bus.Publish(event.NewEvent("user.created", nil))
	}

	for i := 0; i < 5; i++ {
		bus.Publish(event.NewEvent("order.created", nil))
	}

	time.Sleep(300 * time.Millisecond)

	// Verify metrics
	var userCount, orderCount int

	err = tc.DB.QueryRowContext(ctx, "SELECT count FROM event_metrics WHERE event_type = $1", "user.created").Scan(&userCount)
	require.NoError(t, err)
	assert.Equal(t, 10, userCount)

	err = tc.DB.QueryRowContext(ctx, "SELECT count FROM event_metrics WHERE event_type = $1", "order.created").Scan(&orderCount)
	require.NoError(t, err)
	assert.Equal(t, 5, orderCount)
}

func TestIntegration_BatchPersistence(t *testing.T) {
	ctx := context.Background()

	tc, err := tests.SetupTestContainer(ctx)
	require.NoError(t, err)
	defer tc.Cleanup(ctx)

	// Seed database
	err = testdata.SeedEventLogTable(ctx, tc.DB)
	require.NoError(t, err)

	// Create event bus
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	// Batch handler for efficient persistence
	config := event.BatchConfig{
		BatchSize:   5,
		FlushPeriod: 100 * time.Millisecond,
	}

	batchPersistHandler := func(ctx context.Context, events []*event.Event) error {
		tx, err := tc.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		stmt, err := tx.PrepareContext(ctx, "INSERT INTO event_log (event_type, payload, metadata) VALUES ($1, $2, $3)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, evt := range events {
			payload, _ := json.Marshal(evt.Data)
			metadata, _ := json.Marshal(evt.Metadata)

			_, err = stmt.ExecContext(ctx, string(evt.Type), payload, metadata)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	}

	processor := event.NewBatchProcessor("order.created", config, batchPersistHandler)
	defer processor.Shutdown()

	bus.Subscribe("order.created", processor.AsEventHandler())

	// Publish 13 events (2 full batches + partial)
	for i := 0; i < 13; i++ {
		data := map[string]interface{}{
			"order_id": fmt.Sprintf("order_%d", i),
			"amount":   99.99,
		}
		bus.Publish(event.NewEvent("order.created", data))
	}

	time.Sleep(300 * time.Millisecond)

	// Verify all events persisted
	var count int
	err = tc.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM event_log WHERE event_type = $1", "order.created").Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 10, "At least 2 full batches should be persisted")
}

func TestIntegration_UserEventTracking(t *testing.T) {
	ctx := context.Background()

	tc, err := tests.SetupTestContainer(ctx)
	require.NoError(t, err)
	defer tc.Cleanup(ctx)

	// Seed database
	err = testdata.SeedUserEventsTable(ctx, tc.DB)
	require.NoError(t, err)

	// Create event bus
	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	// Handler to track user events
	trackingHandler := func(ctx context.Context, e *event.Event) error {
		data, ok := e.Data.(map[string]interface{})
		if !ok {
			return nil
		}

		userID, ok := data["user_id"].(string)
		if !ok {
			return nil
		}

		_, err := tc.DB.ExecContext(ctx,
			"INSERT INTO user_events (user_id, event_type) VALUES ($1, $2)",
			userID, string(e.Type))

		return err
	}

	bus.Subscribe("user.login", trackingHandler)
	bus.Subscribe("user.logout", trackingHandler)

	// Simulate user activity
	users := []string{"user1", "user2", "user1", "user3", "user1"}

	for _, userID := range users {
		bus.Publish(event.NewEvent("user.login", map[string]interface{}{
			"user_id": userID,
		}))
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)

	// Verify tracking
	var user1Count int
	err = tc.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_events WHERE user_id = $1", "user1").Scan(&user1Count)
	require.NoError(t, err)
	assert.Equal(t, 3, user1Count, "user1 should have 3 login events")
}

func TestIntegration_TransactionalEventPublishing(t *testing.T) {
	ctx := context.Background()

	tc, err := tests.SetupTestContainer(ctx)
	require.NoError(t, err)
	defer tc.Cleanup(ctx)

	err = testdata.SeedEventLogTable(ctx, tc.DB)
	require.NoError(t, err)

	bus := event.NewEventBus(event.DefaultConfig())
	defer bus.Shutdown()

	persistHandler := createPersistenceHandler(tc.DB)
	bus.Subscribe("order.created", persistHandler)

	// Simulate transactional business logic
	processOrder := func() error {
		tx, err := tc.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Business logic here...
		// For demo, just commit

		if err := tx.Commit(); err != nil {
			return err
		}

		// Only publish event after successful commit
		bus.Publish(event.NewEvent("order.created", map[string]interface{}{
			"order_id": "order_123",
			"status":   "confirmed",
		}))

		return nil
	}

	err = processOrder()
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	var count int
	err = tc.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM event_log WHERE event_type = $1", "order.created").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// -------------------------------------------- Helper Functions --------------------------------------------

func createPersistenceHandler(db *sql.DB) event.EventHandler {
	return func(ctx context.Context, e *event.Event) error {
		payload, err := json.Marshal(e.Data)
		if err != nil {
			return err
		}

		metadata, err := json.Marshal(e.Metadata)
		if err != nil {
			return err
		}

		_, err = db.ExecContext(ctx,
			"INSERT INTO event_log (event_type, payload, metadata) VALUES ($1, $2, $3)",
			string(e.Type), payload, metadata)

		return err
	}
}

func createMetricsHandler(db *sql.DB) event.EventHandler {
	return func(ctx context.Context, e *event.Event) error {
		_, err := db.ExecContext(ctx, `
			INSERT INTO event_metrics (event_type, count, last_occurred)
			VALUES ($1, 1, CURRENT_TIMESTAMP)
			ON CONFLICT (event_type) 
			DO UPDATE SET 
				count = event_metrics.count + 1,
				last_occurred = CURRENT_TIMESTAMP
		`, string(e.Type))

		return err
	}
}
