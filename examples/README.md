# Gossip Examples - Comprehensive Guide

This guide provides extensive documentation for using Gossip in real-world applications.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Concepts](#basic-concepts)
3. [Setting Up Your Project](#setting-up-your-project)
4. [Real-World Examples](#real-world-examples)
5. [Advanced Patterns](#advanced-patterns)
6. [Best Practices](#best-practices)
7. [Common Pitfalls](#common-pitfalls)

---

## Getting Started

### Installation

```bash
go get github.com/seyedali-dev/gossip
```

### Import in Your Code

```go
import gossip "github.com/seyedali-dev/gossip/event"
```

---

## Basic Concepts

### 1. Event Types

Event types are strongly-typed string constants that identify events. **Never use raw strings.**

```go
// ✅ Good - Strongly typed
const (
    UserCreated gossip.EventType = "user.created"
    OrderPaid   gossip.EventType = "order.paid"
)

// ❌ Bad - Raw strings prone to typos
bus.Publish(gossip.NewEvent("user.created", data))
```

**Convention:** Use hierarchical naming with dots: `domain.entity.action`

Examples:
- `auth.user.created`
- `auth.login.success`
- `order.payment.completed`
- `inventory.stock.depleted`

### 2. Event Structure

Events contain:
- **Type**: Strongly-typed identifier
- **Timestamp**: Automatic creation time
- **Data**: Event-specific payload (any type)
- **Metadata**: Additional context (map)

```go
event := gossip.NewEvent(UserCreated, &UserData{
    UserID: "123",
    Email:  "user@example.com",
})

// Add metadata for context
event.WithMetadata("request_id", "req-abc-123")
event.WithMetadata("source", "api")
event.WithMetadata("priority", "high")
```

### 3. Handlers

Handlers are functions that process events:

```go
func myHandler(ctx context.Context, event *gossip.Event) error {
    // Type assert the data
    data := event.Data.(*UserData)
    
    // Process the event
    log.Printf("Processing user: %s", data.UserID)
    
    // Return error if processing fails
    return nil
}
```

**Key points:**
- Always type-assert `event.Data` to expected type
- Return `error` if processing fails
- Use `ctx` for cancellation/timeout
- Handlers should be idempotent when possible

---

## Setting Up Your Project

### Step 1: Define Event Types

Create a dedicated file for event types:

```go
// events/types.go
package events

import "github.com/seyedali-dev/gossip"

const (
    // User events
    UserCreated      gossip.EventType = "user.created"
    UserUpdated      gossip.EventType = "user.updated"
    UserDeleted      gossip.EventType = "user.deleted"
    
    // Auth events
    LoginSuccess     gossip.EventType = "auth.login.success"
    LoginFailed      gossip.EventType = "auth.login.failed"
    PasswordChanged  gossip.EventType = "auth.password.changed"
    
    // Order events
    OrderCreated     gossip.EventType = "order.created"
    OrderPaid        gossip.EventType = "order.paid"
    OrderShipped     gossip.EventType = "order.shipped"
)
```

### Step 2: Define Event Data Structures

Create structs for event payloads:

```go
// events/data.go
package events

type UserCreatedData struct {
    UserID   string
    Email    string
    Username string
}

type LoginSuccessData struct {
    UserID    string
    IPAddress string
    UserAgent string
}

type OrderCreatedData struct {
    OrderID    string
    CustomerID string
    Amount     float64
    Items      []string
}
```

### Step 3: Initialize Event Bus

In your main application:

```go
// main.go
package main

import (
    "log"
    gossip "github.com/seyedali-dev/gossip/event"
)

var eventBus *gossip.EventBus

func initEventBus() {
    config := &gossip.Config{
        Workers:    10,    // Adjust based on load
        BufferSize: 1000,  // Adjust based on event volume
    }
    
    eventBus = gossip.NewEventBus(config)
    log.Println("Event bus initialized")
}

func main() {
    initEventBus()
    defer eventBus.Shutdown()
    
    // Register handlers
    registerHandlers()
    
    // Start your application
    // ...
}
```

### Step 4: Register Handlers

Create handler registration function:

```go
// handlers/registration.go
package handlers

import (
    "github.com/seyedali-dev/gossip"
    "yourapp/events"
)

func RegisterHandlers(bus *gossip.EventBus) {
    // User handlers
    bus.Subscribe(events.UserCreated, EmailNotificationHandler)
    bus.Subscribe(events.UserCreated, AuditLogHandler)
    bus.Subscribe(events.UserCreated, MetricsHandler)
    
    // Auth handlers
    bus.Subscribe(events.LoginSuccess, SecurityHandler)
    bus.Subscribe(events.LoginSuccess, AnalyticsHandler)
    
    // Order handlers
    bus.Subscribe(events.OrderCreated, InventoryHandler)
    bus.Subscribe(events.OrderPaid, PaymentProcessorHandler)
}
```

### Step 5: Publish Events in Business Logic

```go
// services/user_service.go
package services

import (
    "github.com/seyedali-dev/gossip"
    "yourapp/events"
)

type UserService struct {
    bus *gossip.EventBus
}

func (s *UserService) CreateUser(email, username string) error {
    // Core business logic
    userID := generateUserID()
    
    // ... save to database ...
    
    // Publish event
    eventData := &events.UserCreatedData{
        UserID:   userID,
        Email:    email,
        Username: username,
    }
    
    event := gossip.NewEvent(events.UserCreated, eventData).
        WithMetadata("source", "api").
        WithMetadata("request_id", getRequestID())
    
    s.bus.Publish(event)
    
    return nil
}
```

---

## Real-World Examples

### 1. Authentication Service

**File: `examples/auth_service/main.go`**

Demonstrates:
- User registration with welcome emails
- Login tracking and notifications
- Password change alerts
- Audit logging for all auth events
- Security monitoring

**Run:**
```bash
cd examples/auth_service
go run main.go
```

**Key Patterns:**
- Multiple handlers for single event (fan-out)
- Middleware composition (retry, timeout, recovery)
- Metadata usage for request tracking

### 2. E-commerce Platform

**File: `examples/ecommerce/main.go`**

Demonstrates:
- Order creation and processing
- Batch email notifications
- Inventory management
- Conditional analytics (high-value orders)
- Payment processing pipeline

**Run:**
```bash
cd examples/ecommerce
go run main.go
```

**Key Patterns:**
- Batch processing for efficiency
- Event filtering based on conditions
- Priority-based processing
- Async inventory updates

### 3. Microservices Communication

**File: `examples/microservices/main.go`**

Demonstrates:
- Cross-service event communication
- Service-to-service decoupling
- Multiple services reacting to same event
- Event chaining (service publishes events other services consume)

**Run:**
```bash
cd examples/microservices
go run main.go
```

**Key Patterns:**
- Self-registering handlers in service constructors
- Shared event bus across services
- Event-driven service orchestration

---

## Advanced Patterns

### 1. Middleware Composition

Chain multiple middleware for robust handling:

```go
handler := gossip.Chain(
    gossip.WithRecovery(),                      // Catch panics
    gossip.WithRetry(3, 100*time.Millisecond),  // Retry on failure
    gossip.WithTimeout(5*time.Second),          // Prevent hanging
    gossip.WithLogging(),                       // Log execution
)(myHandler)

bus.Subscribe(UserCreated, handler)
```

**Order matters:** Place recovery first to catch all panics.

### 2. Conditional Processing with Filters

Only execute handlers when conditions are met:

```go
// Only process high-priority events
highPriorityFilter := func(event *gossip.Event) bool {
    priority, exists := event.Metadata["priority"]
    return exists && priority == "high"
}

bus.Subscribe(OrderCreated, gossip.NewFilteredHandler(
    highPriorityFilter,
    urgentOrderHandler,
))

// Only process events from specific source
apiFilter := gossip.FilterByMetadata("source", "api")
bus.Subscribe(UserCreated, gossip.NewFilteredHandler(apiFilter, apiHandler))

// Combine filters with AND/OR logic
complexFilter := gossip.And(
    gossip.FilterByMetadataExists("user_id"),
    gossip.Or(
        gossip.FilterByMetadata("source", "api"),
        gossip.FilterByMetadata("source", "web"),
    ),
)
```

### 3. Batch Processing

Efficient processing of high-volume events:

```go
batchConfig := gossip.BatchConfig{
    BatchSize:   100,              // Process 100 events at once
    FlushPeriod: 5 * time.Second,  // Or every 5 seconds
}

batchHandler := func(ctx context.Context, events []*gossip.Event) error {
    // Process all events together
    log.Printf("Processing batch of %d events", len(events))
    
    // Bulk database insert, bulk email send, etc.
    return bulkInsertToDatabase(events)
}

processor := gossip.NewBatchProcessor(OrderCreated, batchConfig, batchHandler)
defer processor.Shutdown()

bus.Subscribe(OrderCreated, processor.AsEventHandler())
```

**When to use:**
- Email notifications (batch send)
- Database inserts (bulk insert)
- API calls (batch requests)
- Metric aggregation

### 4. Synchronous Processing

When you need immediate results:

```go
event := gossip.NewEvent(CriticalTransaction, data)

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

errors := bus.PublishSync(ctx, event)
if len(errors) > 0 {
    // Some handlers failed
    log.Printf("Handler failures: %v", errors)
    return handleFailure(errors)
}

log.Println("All handlers succeeded")
```

**Use cases:**
- Critical transactions requiring validation
- Rollback scenarios
- Immediate feedback needed
- Testing

### 5. Priority-Based Processing

Handle critical events first:

```go
priorityQueue := gossip.NewPriorityQueue()

// Enqueue with priority
priorityQueue.Enqueue(criticalEvent, gossip.PriorityHigh)
priorityQueue.Enqueue(normalEvent, gossip.PriorityNormal)
priorityQueue.Enqueue(backgroundTask, gossip.PriorityLow)

// Dequeue processes highest priority first
event, ok := priorityQueue.Dequeue()
if ok {
    bus.Publish(event)
}
```

---

## Best Practices

### 1. Event Naming Conventions

✅ **Do:**
- Use hierarchical naming: `domain.entity.action`
- Be specific: `order.payment.completed` not `order.updated`
- Use past tense: `user.created` not `user.create`
- Group related events: `auth.login.*`, `order.payment.*`

❌ **Don't:**
- Use generic names: `updated`, `changed`
- Mix tenses: `user.creating`, `order.created`
- Use abbreviations: `usr.crtd`

### 2. Handler Design

✅ **Do:**
- Keep handlers small and focused
- Make handlers idempotent when possible
- Return errors for failures
- Use context for cancellation
- Log handler actions

❌ **Don't:**
- Block for long periods
- Panic in handlers (use middleware recovery)
- Modify shared state without locking
- Ignore errors

### 3. Error Handling

```go
func myHandler(ctx context.Context, event *gossip.Event) error {
    // Type assertion with check
    data, ok := event.Data.(*UserData)
    if !ok {
        return fmt.Errorf("invalid event data type")
    }
    
    // Business logic with error handling
    if err := processUser(data); err != nil {
        return fmt.Errorf("failed to process user: %w", err)
    }
    
    return nil
}
```

### 4. Testing Handlers

```go
func TestUserCreatedHandler(t *testing.T) {
    // Create test event
    event := gossip.NewEvent(UserCreated, &UserData{
        UserID: "test-123",
        Email:  "test@example.com",
    })
    
    // Execute handler
    err := myHandler(context.Background(), event)
    
    // Assert results
    assert.NoError(t, err)
    assert.True(t, emailWasSent("test@example.com"))
}
```

### 5. Configuration Tuning

```go
// Low volume (< 100 events/sec)
config := &gossip.Config{
    Workers:    5,
    BufferSize: 500,
}

// Medium volume (100-1000 events/sec)
config := &gossip.Config{
    Workers:    10,
    BufferSize: 1000,
}

// High volume (> 1000 events/sec)
config := &gossip.Config{
    Workers:    20,
    BufferSize: 5000,
}
```

---

## Common Pitfalls

### 1. Type Assertion Panics

❌ **Wrong:**
```go
data := event.Data.(*UserData) // Panics if wrong type
```

✅ **Correct:**
```go
data, ok := event.Data.(*UserData)
if !ok {
    return fmt.Errorf("invalid data type")
}
```

### 2. Blocking Handlers

❌ **Wrong:**
```go
func slowHandler(ctx context.Context, event *gossip.Event) error {
    time.Sleep(30 * time.Second) // Blocks worker
    return nil
}
```

✅ **Correct:**
```go
func fastHandler(ctx context.Context, event *gossip.Event) error {
    // Offload heavy work
    go processInBackground(event.Data)
    return nil
}
```

### 3. Forgetting Shutdown

❌ **Wrong:**
```go
func main() {
    bus := gossip.NewEventBus(config)
    // ... application code ...
    // No cleanup!
}
```

✅ **Correct:**
```go
func main() {
    bus := gossip.NewEventBus(config)
    defer bus.Shutdown() // Graceful cleanup
    // ... application code ...
}
```

### 4. Circular Event Dependencies

❌ **Wrong:**
```go
// Handler A publishes Event B
func handlerA(ctx context.Context, event *gossip.Event) error {
    bus.Publish(gossip.NewEvent(EventB, nil))
    return nil
}

// Handler B publishes Event A (circular!)
func handlerB(ctx context.Context, event *gossip.Event) error {
    bus.Publish(gossip.NewEvent(EventA, nil))
    return nil
}
```

✅ **Correct:**
- Avoid circular dependencies
- Use event metadata to track origin and prevent loops
- Design event flow as a DAG (Directed Acyclic Graph)

### 5. Overusing Synchronous Publishing

❌ **Wrong:**
```go
// Using sync for everything defeats the purpose
for _, item := range items {
    bus.PublishSync(ctx, event) // Slow!
}
```

✅ **Correct:**
```go
// Use async for most cases
for _, item := range items {
    bus.Publish(event) // Fast!
}

// Use sync only when necessary
criticalEvent := gossip.NewEvent(CriticalOp, data)
errors := bus.PublishSync(ctx, criticalEvent)
```

---

## Performance Tips

1. **Tune worker count** based on handler latency
2. **Use batch processing** for high-volume events
3. **Implement filtering** to avoid unnecessary work
4. **Monitor buffer fullness** - increase if events are dropped
5. **Profile handler performance** - optimize slow handlers
6. **Use middleware wisely** - each layer adds overhead

---

## Troubleshooting

### Events Not Being Processed

1. Check if handlers are registered before publishing
2. Verify event type matches subscription
3. Ensure bus hasn't been shutdown
4. Check logs for handler errors

### Slow Event Processing

1. Profile handler execution time
2. Check for blocking operations
3. Increase worker count
4. Use batch processing

### Memory Issues

1. Reduce buffer size
2. Implement event filtering
3. Fix handler memory leaks
4. Monitor goroutine count

---

## Additional Resources

- [Main README](../README.md) - Library overview
- [API Documentation](https://pkg.go.dev/github.com/seyedali-dev/gossip)
- [GitHub Issues](https://github.com/seyedali-dev/gossip/issues)

---

**Questions?** Open an issue on GitHub or reach out to the community!