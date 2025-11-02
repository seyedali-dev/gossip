# Gossip (ﾉ◕ヮ◕)ﾉ*:･ﾟ✧

**Decoupled event handling for Go — publish once, let everyone listen in.** 🗣️

Gossip is a lightweight, type-safe event bus library for Go that implements the observer/pub-sub pattern. It enables clean separation between core business logic and side effects, making your codebase more maintainable and extensible.

## ✨ Features

- **🔒 Strongly-typed events** - No string typos with `EventType` constants
- **⚡ Async by default** - Non-blocking event dispatch with worker pools
- **🔄 Synchronous option** - When you need immediate processing
- **🎯 Event filtering** - Conditional handler execution
- **📦 Batch processing** - Process multiple events efficiently
- **🔧 Middleware support** - Retry, timeout, logging, recovery
- **🎚️ Priority queues** - Handle critical events first
- **🛡️ Thread-safe** - Concurrent publish/subscribe operations
- **🧹 Graceful shutdown** - Wait for in-flight events

## 📦 Installation

```bash
go get github.com/seyedali-dev/gossip
```

## 🚀 Quick Start

```go
package main

import (
    "context"
    "log"
    gossip "github.com/seyedali-dev/gossip/event"
)

// Define event types
const (
    UserCreated gossip.EventType = "user.created"
)

type UserData struct {
    UserID   string
    Username string
}

func main() {
    // Initialize event bus
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    // Subscribe handler
    bus.Subscribe(UserCreated, func(ctx context.Context, event *gossip.Event) error {
        data := event.Data.(*UserData)
        log.Printf("New user: %s", data.Username)
        return nil
    })
    
    // Publish event
    event := gossip.NewEvent(UserCreated, &UserData{
        UserID:   "123",
        Username: "alice",
    })
    bus.Publish(event)
}
```

## 📚 Documentation

For comprehensive documentation, examples, and advanced usage patterns, see:

- **[Examples Directory](./examples/)** - Real-world use cases
    - [Auth Service](./examples/auth_service/) - Authentication with notifications
    - [E-commerce](./examples/ecommerce/) - Order processing with batch emails
    - [Microservices](./examples/microservices/) - Cross-service communication

## 🎯 Use Cases

- **Authentication systems** - Login notifications, security alerts
- **E-commerce platforms** - Order processing, inventory updates
- **Microservices** - Cross-service event communication
- **Audit logging** - Track all system activities
- **Analytics** - Collect metrics without blocking main flow
- **Notification systems** - Email, SMS, push notifications

## 🏗️ Core Concepts

### Event Types
Define strongly-typed event identifiers:
```go
const (
    UserCreated gossip.EventType = "user.created"
    UserUpdated gossip.EventType = "user.updated"
)
```

### Events
Events carry data and metadata:
```go
event := gossip.NewEvent(UserCreated, userData).
    WithMetadata("request_id", "req-123").
    WithMetadata("source", "api")
```

### Handlers
Functions that process events:
```go
func myHandler(ctx context.Context, event *gossip.Event) error {
    // Process event
    return nil
}
```

### Subscribe & Publish
```go
// Subscribe
bus.Subscribe(UserCreated, myHandler)

// Publish (async)
bus.Publish(event)

// Publish (sync)
errors := bus.PublishSync(ctx, event)
```

## 🔧 Advanced Features

### Middleware
Chain behaviors around handlers:
```go
handler := gossip.Chain(
    gossip.WithRetry(3, 100*time.Millisecond),
    gossip.WithTimeout(5*time.Second),
    gossip.WithLogging(),
    gossip.WithRecovery(),
)(myHandler)

bus.Subscribe(UserCreated, handler)
```

### Filtering
Conditionally execute handlers:
```go
filter := func(event *gossip.Event) bool {
    return event.Metadata["priority"] == "high"
}

bus.Subscribe(OrderCreated, gossip.NewFilteredHandler(filter, handler))
```

### Batch Processing
Process events in groups:
```go
batchConfig := gossip.BatchConfig{
    BatchSize:   100,
    FlushPeriod: 5 * time.Second,
}

processor := gossip.NewBatchProcessor(OrderCreated, batchConfig, batchHandler)
bus.Subscribe(OrderCreated, processor.AsEventHandler())
```

## ⚙️ Configuration

```go
config := &gossip.Config{
    Workers:    20,     // Number of worker goroutines
    BufferSize: 2000,   // Event channel buffer size
}

bus := gossip.NewEventBus(config)
```

## 🧪 Testing

```go
func TestMyHandler(t *testing.T) {
    bus := gossip.NewEventBus(gossip.DefaultConfig())
    defer bus.Shutdown()
    
    received := false
    handler := func(ctx context.Context, event *gossip.Event) error {
        received = true
        return nil
    }
    
    bus.Subscribe(UserCreated, handler)
    bus.Publish(gossip.NewEvent(UserCreated, nil))
    
    time.Sleep(100 * time.Millisecond)
    assert.True(t, received)
}
```

## 🤝 Contributing

Contributions welcome! Please open an issue or submit a PR.

## 📄 License

MIT License - see LICENSE file for details

## 📖 Implementation Details

For the geeks, see [IMPLEMENTATION.md](./IMPLEMENTATION.md) - a comprehensive deep-dive into how Gossip is implemented, including the concurrency model, thread safety patterns, middleware system, and architectural decisions.

## 🙏 Acknowledgments

Inspired by the need for clean event-driven architecture in Go applications.

---

<p align="center" style="font-style: italic; font-family: 'Gochi Hand',fantasy">
    Built with ❤️ for the Go community
</p>
