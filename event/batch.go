// Package event. batch provides batch processing capabilities for events.
package event

import (
	"context"
	"sync"
	"time"
)

// BatchHandler processes multiple events at once.
type BatchHandler func(ctx context.Context, events []*Event) error

// BatchProcessor collects events and processes them in batches.
type BatchProcessor struct {
	mu          sync.Mutex
	eventType   EventType
	batchSize   int
	flushPeriod time.Duration
	handler     BatchHandler
	buffer      []*Event
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// BatchConfig holds configuration for batch processing.
type BatchConfig struct {
	BatchSize   int           // Max events per batch
	FlushPeriod time.Duration // Max time to wait before flushing
}

// NewBatchProcessor creates a new batch processor.
func NewBatchProcessor(eventType EventType, config BatchConfig, handler BatchHandler) *BatchProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	bp := &BatchProcessor{
		eventType:   eventType,
		batchSize:   config.BatchSize,
		flushPeriod: config.FlushPeriod,
		handler:     handler,
		buffer:      make([]*Event, 0, config.BatchSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	bp.start()
	return bp
}

// -------------------------------------------- Public Functions --------------------------------------------

// Add adds an event to the batch buffer.
func (bp *BatchProcessor) Add(event *Event) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.buffer = append(bp.buffer, event)

	if len(bp.buffer) >= bp.batchSize {
		bp.flush()
	}
}

// Flush processes all buffered events immediately.
func (bp *BatchProcessor) Flush() {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.flush()
}

// Shutdown stops the batch processor and flushes remaining events.
func (bp *BatchProcessor) Shutdown() {
	bp.cancel()
	bp.wg.Wait()
	bp.Flush()
}

// AsEventHandler returns an EventHandler that adds events to the batch.
func (bp *BatchProcessor) AsEventHandler() EventHandler {
	return func(ctx context.Context, event *Event) error {
		bp.Add(event)
		return nil
	}
}

// -------------------------------------------- Private Helper Functions --------------------------------------------

// start begins the periodic flush goroutine.
func (bp *BatchProcessor) start() {
	bp.wg.Add(1)
	go bp.periodicFlush()
}

// periodicFlush flushes the buffer at regular intervals.
func (bp *BatchProcessor) periodicFlush() {
	defer bp.wg.Done()

	ticker := time.NewTicker(bp.flushPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bp.Flush()

		case <-bp.ctx.Done():
			return
		}
	}
}

// flush processes the current buffer (must be called with lock held).
func (bp *BatchProcessor) flush() {
	if len(bp.buffer) == 0 {
		return
	}

	events := make([]*Event, len(bp.buffer))
	copy(events, bp.buffer)
	bp.buffer = bp.buffer[:0]

	go func() {
		if err := bp.handler(bp.ctx, events); err != nil {
			// Log error but don't block
		}
	}()
}
