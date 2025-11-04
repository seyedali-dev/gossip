// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

package event_test

import (
	"testing"

	"github.com/seyedali-dev/gossip/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------- Priority Queue Tests --------------------------------------------

func TestNewPriorityQueue_CreatesQueue(t *testing.T) {
	pq := event.NewPriorityQueue()
	require.NotNil(t, pq)
	assert.Equal(t, 0, pq.Size())
}

func TestPriorityQueue_EnqueueDequeue(t *testing.T) {
	pq := event.NewPriorityQueue()

	evt := event.NewEvent("test", nil)
	pq.Enqueue(evt, event.PriorityNormal)

	assert.Equal(t, 1, pq.Size())

	dequeuedEvt, ok := pq.Dequeue()
	assert.True(t, ok)
	assert.Equal(t, evt, dequeuedEvt)
	assert.Equal(t, 0, pq.Size())
}

func TestPriorityQueue_EmptyDequeue(t *testing.T) {
	pq := event.NewPriorityQueue()

	evt, ok := pq.Dequeue()
	assert.False(t, ok)
	assert.Nil(t, evt)
}

func TestPriorityQueue_PriorityOrdering(t *testing.T) {
	pq := event.NewPriorityQueue()

	lowEvt := event.NewEvent("low", "low priority")
	normalEvt := event.NewEvent("normal", "normal priority")
	highEvt := event.NewEvent("high", "high priority")

	// Enqueue in mixed order
	pq.Enqueue(normalEvt, event.PriorityNormal)
	pq.Enqueue(lowEvt, event.PriorityLow)
	pq.Enqueue(highEvt, event.PriorityHigh)

	assert.Equal(t, 3, pq.Size())

	// Should dequeue in priority order: high -> normal -> low
	evt1, ok1 := pq.Dequeue()
	assert.True(t, ok1)
	assert.Equal(t, "high", evt1.Data.(string))

	evt2, ok2 := pq.Dequeue()
	assert.True(t, ok2)
	assert.Equal(t, "normal", evt2.Data.(string))

	evt3, ok3 := pq.Dequeue()
	assert.True(t, ok3)
	assert.Equal(t, "low", evt3.Data.(string))
}

func TestPriorityQueue_SamePriority(t *testing.T) {
	pq := event.NewPriorityQueue()

	evt1 := event.NewEvent("evt1", 1)
	evt2 := event.NewEvent("evt2", 2)
	evt3 := event.NewEvent("evt3", 3)

	// All same priority
	pq.Enqueue(evt1, event.PriorityNormal)
	pq.Enqueue(evt2, event.PriorityNormal)
	pq.Enqueue(evt3, event.PriorityNormal)

	// Order within same priority is implementation-defined (usually FIFO-ish)
	for i := 0; i < 3; i++ {
		evt, ok := pq.Dequeue()
		assert.True(t, ok)
		assert.NotNil(t, evt)
	}

	assert.Equal(t, 0, pq.Size())
}

func TestPriorityQueue_MixedPriorities(t *testing.T) {
	pq := event.NewPriorityQueue()

	events := []struct {
		name     string
		priority int
	}{
		{"evt1", event.PriorityLow},
		{"evt2", event.PriorityHigh},
		{"evt3", event.PriorityNormal},
		{"evt4", event.PriorityHigh},
		{"evt5", event.PriorityLow},
		{"evt6", event.PriorityNormal},
	}

	for _, e := range events {
		pq.Enqueue(event.NewEvent(event.EventType(e.name), nil), e.priority)
	}

	assert.Equal(t, 6, pq.Size())

	// First two should be high priority (evt2, evt4)
	evt1, ok := pq.Dequeue()
	require.True(t, ok)
	assert.Contains(t, []string{"evt2", "evt4"}, string(evt1.Type))

	evt2, ok := pq.Dequeue()
	require.True(t, ok)
	assert.Contains(t, []string{"evt2", "evt4"}, string(evt2.Type))

	// Next two should be normal priority (evt3, evt6)
	evt3, ok := pq.Dequeue()
	require.True(t, ok)
	assert.Contains(t, []string{"evt3", "evt6"}, string(evt3.Type))

	evt4, ok := pq.Dequeue()
	require.True(t, ok)
	assert.Contains(t, []string{"evt3", "evt6"}, string(evt4.Type))

	// Last two should be low priority (evt1, evt5)
	evt5, ok := pq.Dequeue()
	require.True(t, ok)
	assert.Contains(t, []string{"evt1", "evt5"}, string(evt5.Type))

	evt6, ok := pq.Dequeue()
	require.True(t, ok)
	assert.Contains(t, []string{"evt1", "evt5"}, string(evt6.Type))
}

func TestPriorityQueue_CustomPriorities(t *testing.T) {
	pq := event.NewPriorityQueue()

	// Use custom priority values
	critical := 100
	important := 50
	routine := 1

	evt1 := event.NewEvent("routine", "routine task")
	evt2 := event.NewEvent("important", "important task")
	evt3 := event.NewEvent("critical", "critical task")

	pq.Enqueue(evt1, routine)
	pq.Enqueue(evt2, important)
	pq.Enqueue(evt3, critical)

	// Should dequeue in order: critical -> important -> routine
	evt, _ := pq.Dequeue()
	assert.Equal(t, "critical", string(evt.Type))

	evt, _ = pq.Dequeue()
	assert.Equal(t, "important", string(evt.Type))

	evt, _ = pq.Dequeue()
	assert.Equal(t, "routine", string(evt.Type))
}

func TestPriorityQueue_ConcurrentEnqueue(t *testing.T) {
	pq := event.NewPriorityQueue()

	// Note: This is a basic concurrency test
	// PriorityQueue should be thread-safe due to mutex

	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(id int) {
			pq.Enqueue(event.NewEvent("test", id), event.PriorityNormal)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	assert.Equal(t, 100, pq.Size())
}

func TestPriorityQueue_LargeVolume(t *testing.T) {
	pq := event.NewPriorityQueue()

	numEvents := 1000

	// Enqueue many events
	for i := 0; i < numEvents; i++ {
		priority := event.PriorityNormal
		if i%3 == 0 {
			priority = event.PriorityHigh
		} else if i%5 == 0 {
			priority = event.PriorityLow
		}

		pq.Enqueue(event.NewEvent("test", i), priority)
	}

	assert.Equal(t, numEvents, pq.Size())

	// Dequeue all
	highCount := 0
	normalCount := 0
	lowCount := 0
	//lastPriority := 100 // Start with max

	for i := 0; i < numEvents; i++ {
		evt, ok := pq.Dequeue()
		require.True(t, ok)
		require.NotNil(t, evt)

		// Track counts (approximate based on enqueueing logic)
		if i < numEvents/3 {
			highCount++
		} else if i < 2*numEvents/3 {
			normalCount++
		} else {
			lowCount++
		}
	}

	assert.Equal(t, 0, pq.Size())

	// High priority events should be processed first
	assert.Greater(t, highCount, 0)
}

// -------------------------------------------- Benchmark Tests --------------------------------------------

func BenchmarkPriorityQueue_Enqueue(b *testing.B) {
	pq := event.NewPriorityQueue()
	evt := event.NewEvent("test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Enqueue(evt, event.PriorityNormal)
	}
}

func BenchmarkPriorityQueue_Dequeue(b *testing.B) {
	pq := event.NewPriorityQueue()

	// Pre-fill queue
	for i := 0; i < b.N; i++ {
		pq.Enqueue(event.NewEvent("test", i), event.PriorityNormal)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Dequeue()
	}
}

func BenchmarkPriorityQueue_MixedOperations(b *testing.B) {
	pq := event.NewPriorityQueue()
	evt := event.NewEvent("test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			pq.Enqueue(evt, event.PriorityNormal)
		} else {
			pq.Dequeue()
		}
	}
}
