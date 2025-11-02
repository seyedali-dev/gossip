# Testing Guide for Gossip

Complete guide for testing event-driven code using Gossip.

## Unit Testing Handlers

### Basic Handler Test

```go
func TestEmailHandler(t *testing.T) {
    event := gossip.NewEvent(UserCreated, &UserData{
        UserID: "123",
        Email:  "test@example.com",
    })
    
    err := emailHandler(context.Background(), event)
    
    assert.NoError(t, err)
}
```

### Testing with Mock Dependencies

```go
type MockEmailSender struct {
    SentEmails []string
}

func (m *MockEmailSender) Send(to string) error {
    m.SentEmails = append(m.SentEmails, to)
    return nil
}

func TestEmailHandlerWithMock(t *testing.T) {
    mockSender := &MockEmailSender{}
    handler := createEmailHandler(mockSender)
    
    event := gossip.NewEvent(UserCreated, &UserData{
        Email: "test@example.com",
    })
    
    err := handler(context.Background(), event)
    
    assert.NoError(t, err)
    assert.Equal(t, 1, len(mockSender.SentEmails))
    assert.Equal(t, "test@example.com", mockSender.SentEmails[0])
}
```

## Integration Testing

### Testing Event Flow

```go
func TestUserCreationFlow(t *testing.T) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    // Track handler executions
    var emailSent, auditLogged, metricsRecorded bool
    
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        emailSent = true
        return nil
    })
    
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        auditLogged = true
        return nil
    })
    
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        metricsRecorded = true
        return nil
    })
    
    // Publish event
    event := gossip.NewEvent(UserCreated, &UserData{UserID: "123"})
    bus.Publish(event)
    
    // Wait for async processing
    time.Sleep(100 * time.Millisecond)
    
    assert.True(t, emailSent)
    assert.True(t, auditLogged)
    assert.True(t, metricsRecorded)
}
```

### Testing Event Ordering

```go
func TestEventOrdering(t *testing.T) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    var events []string
    var mu sync.Mutex
    
    handler := func(ctx context.Context, e *gossip.Event) error {
        mu.Lock()
        defer mu.Unlock()
        events = append(events, string(e.Type))
        return nil
    }
    
    bus.Subscribe(EventA, handler)
    bus.Subscribe(EventB, handler)
    bus.Subscribe(EventC, handler)
    
    bus.Publish(gossip.NewEvent(EventA, nil))
    bus.Publish(gossip.NewEvent(EventB, nil))
    bus.Publish(gossip.NewEvent(EventC, nil))
    
    time.Sleep(100 * time.Millisecond)
    
    assert.Equal(t, 3, len(events))
}
```

## Testing Synchronous Publishing

```go
func TestSyncPublish(t *testing.T) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    executed := false
    
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        executed = true
        return nil
    })
    
    errors := bus.PublishSync(context.Background(), gossip.NewEvent(UserCreated, nil))
    
    assert.Empty(t, errors)
    assert.True(t, executed, "Handler should execute immediately in sync mode")
}
```

## Testing Error Handling

```go
func TestHandlerError(t *testing.T) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    expectedError := errors.New("processing failed")
    
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        return expectedError
    })
    
    event := gossip.NewEvent(UserCreated, nil)
    errors := bus.PublishSync(context.Background(), event)
    
    assert.Equal(t, 1, len(errors))
    assert.Contains(t, errors[0].Error(), "processing failed")
}
```

## Testing Middleware

```go
func TestRetryMiddleware(t *testing.T) {
    attempts := 0
    
    handler := func(ctx context.Context, e *gossip.Event) error {
        attempts++
        if attempts < 3 {
            return errors.New("temporary error")
        }
        return nil
    }
    
    wrappedHandler := gossip.WithRetry(3, 10*time.Millisecond)(handler)
    
    err := wrappedHandler(context.Background(), gossip.NewEvent(UserCreated, nil))
    
    assert.NoError(t, err)
    assert.Equal(t, 3, attempts)
}
```

## Testing Filters

```go
func TestEventFilter(t *testing.T) {
    executed := false
    
    filter := func(e *gossip.Event) bool {
        return e.Metadata["priority"] == "high"
    }
    
    handler := gossip.NewFilteredHandler(filter, func(ctx context.Context, e *gossip.Event) error {
        executed = true
        return nil
    })
    
    // Low priority - should not execute
    event := gossip.NewEvent(UserCreated, nil).WithMetadata("priority", "low")
    handler(context.Background(), event)
    assert.False(t, executed)
    
    // High priority - should execute
    event = gossip.NewEvent(UserCreated, nil).WithMetadata("priority", "high")
    handler(context.Background(), event)
    assert.True(t, executed)
}
```

## Testing Batch Processing

```go
func TestBatchProcessor(t *testing.T) {
    var processedEvents []*gossip.Event
    var mu sync.Mutex
    
    batchHandler := func(ctx context.Context, events []*gossip.Event) error {
        mu.Lock()
        defer mu.Unlock()
        processedEvents = append(processedEvents, events...)
        return nil
    }
    
    config := gossip.BatchConfig{
        BatchSize:   3,
        FlushPeriod: 1 * time.Second,
    }
    
    processor := gossip.NewBatchProcessor(OrderCreated, config, batchHandler)
    defer processor.Shutdown()
    
    // Add events
    for i := 0; i < 5; i++ {
        processor.Add(gossip.NewEvent(OrderCreated, i))
    }
    
    // Wait for batch processing
    time.Sleep(100 * time.Millisecond)
    
    mu.Lock()
    count := len(processedEvents)
    mu.Unlock()
    
    assert.GreaterOrEqual(t, count, 3, "Should process at least one batch")
}
```

## Benchmarking

```go
func BenchmarkEventPublish(b *testing.B) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        return nil
    })
    
    event := gossip.NewEvent(UserCreated, &UserData{UserID: "123"})
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bus.Publish(event)
    }
}

func BenchmarkEventPublishSync(b *testing.B) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        return nil
    })
    
    event := gossip.NewEvent(UserCreated, &UserData{UserID: "123"})
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        bus.PublishSync(ctx, event)
    }
}
```

## Test Helpers

```go
// WaitForEvent waits for a handler to execute
func WaitForEvent(timeout time.Duration, condition func() bool) bool {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if condition() {
            return true
        }
        time.Sleep(10 * time.Millisecond)
    }
    return false
}

// Usage
func TestWithHelper(t *testing.T) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    executed := false
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        executed = true
        return nil
    })
    
    bus.Publish(gossip.NewEvent(UserCreated, nil))
    
    assert.True(t, WaitForEvent(1*time.Second, func() bool {
        return executed
    }))
}
```

## Common Testing Patterns

### Table-Driven Tests

```go
func TestEventHandlers(t *testing.T) {
    tests := []struct {
        name      string
        eventType gossip.EventType
        data      interface{}
        wantError bool
    }{
        {"valid user", UserCreated, &UserData{UserID: "123"}, false},
        {"nil data", UserCreated, nil, true},
        {"invalid type", UserCreated, "string", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            event := gossip.NewEvent(tt.eventType, tt.data)
            err := myHandler(context.Background(), event)
            
            if tt.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Testing Cleanup

```go
func TestGracefulShutdown(t *testing.T) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    
    processed := int32(0)
    bus.Subscribe(UserCreated, func(ctx context.Context, e *gossip.Event) error {
        time.Sleep(50 * time.Millisecond)
        atomic.AddInt32(&processed, 1)
        return nil
    })
    
    // Publish many events
    for i := 0; i < 10; i++ {
        bus.Publish(gossip.NewEvent(UserCreated, nil))
    }
    
    // Shutdown waits for processing
    bus.Shutdown()
    
    assert.Greater(t, atomic.LoadInt32(&processed), int32(0))
}
```