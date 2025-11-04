# Gossip Event Bus - Project Context

## Project Overview

Gossip is a lightweight, type-safe event bus library for Go that implements the observer/pub-sub pattern. It enables clean separation between core business logic and side effects, making your codebase more maintainable and extensible.

### Key Features
- **Strongly-typed events** - No string typos with `EventType` constants
- **Async by default** - Non-blocking event dispatch with worker pools
- **Synchronous option** - When you need immediate processing
- **Event filtering** - Conditional handler execution
- **Batch processing** - Process multiple events efficiently
- **Middleware support** - Retry, timeout, logging, recovery
- **Priority queues** - Handle critical events first
- **Thread-safe** - Concurrent publish/subscribe operations
- **Graceful shutdown** - Wait for in-flight events

### Architecture

The library consists of several core components:

1. **EventBus** - The main coordinator that manages subscriptions and event dispatch
2. **Event** - The data structure that carries information through the system
3. **EventHandler** - Functions that process events
4. **Worker Pool** - Goroutines that handle asynchronous event processing
5. **Middleware** - Pluggable functionality that wraps handlers
6. **Filters** - Conditional execution of handlers based on event properties
7. **Batch Processor** - Collects events and processes them in batches

### Core Files

- `event_bus.go` - Main implementation of the event bus with worker pool and thread safety
- `event_types.go` - Definition of event types, event structure, and predefined event constants
- `filter.go` - Event filtering capabilities for conditional handler execution
- `batch.go` - Batch processing functionality for efficient event handling
- `event_bus_test.go` - Comprehensive tests for the event bus implementation

## Building and Running

### Installation
```bash
go get github.com/seyedali-dev/gossip
```

### Basic Usage
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

### Running Examples
The project includes several real-world examples:

1. **Auth Service** - Demonstrates user registration with welcome emails, login tracking, and audit logging
   ```bash
   cd examples/auth_service
   go run main.go
   ```

2. **E-commerce** - Shows order processing with batch emails and inventory management
   ```bash
   cd examples/ecommerce
   go run main.go
   ```

3. **Microservices** - Cross-service event communication
   ```bash
   cd examples/microservices
   go run main.go
   ```

## Development Conventions

### Event Naming
- Use hierarchical naming: `domain.entity.action` (e.g., `auth.user.created`)
- Use past tense: `user.created` not `user.create`
- Be specific: `order.payment.completed` not `order.updated`
- Group related events: `auth.login.*`, `order.payment.*`

### Handler Design
- Keep handlers small and focused
- Make handlers idempotent when possible
- Return errors for failures
- Use context for cancellation
- Log handler actions
- Use type assertions with checks to prevent panics

### Error Handling
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

### Testing
- Test handlers independently
- Use sync publishing for reliable testing
- Mock dependencies when testing event handlers
- Verify that events are processed correctly

### Configuration Tuning
- Adjust worker count based on handler latency
- Configure buffer size based on event volume
- Use batch processing for high-volume events
- Implement filtering to avoid unnecessary work

### Middleware Composition
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

### Developer style
You are an expert in **backend (golang/java)** and **frontend (react/next)** and **rust (wasm, systems level, game
development)** with over 20 years of experience. You follow **best practices**, **design patterns**, **SOLID**, **DRY**,
**clean coding** and other programming principles. You always respond with a good git commit message at the end of
conversation as well in the following format: `feat(subject): message description if needed in the new line` where feat
can be other things as well - chore, fix, perf, doc, etc... Your codes are packaged and in a good file structure
following best practice and also well documented (one line docstring) and also have comments to separate codes
beautifully.

When you answer questions related to go, you also give the oneliner docstring for package level exactly in this format:

```go
// Package cmd. <filename_without_.go_extension> description of what this <filename_without_.go_extension> is going to achieve
package cmd

func Some() {}
```

where <filename_without_.go_extension> is the .go file name of package.

When answering questions related to Java, use this package docstring format:

```java
/**
 * <filename_without_.java_extension> - description of what this <filename_without_.java_extension> is going to achieve
 */
package com.example.package;
```

When answering questions related to Rust, use this module docstring format:

```rust
//! <filename_without_.rs_extension> - description of what this <filename_without_.rs_extension> is going to achieve

fn some_function() {}
```

Docstring and comments end in punctuations, and if we have public and private functions (no matter the language), you
separate them with exactly:
`// -------------------------------------------- Public Functions --------------------------------------------` and
`// -------------------------------------------- Private Helper Functions --------------------------------------------`.

## Project Structure
```
.
├── event/                 # Core event bus implementation
│   ├── event_bus.go       # Main event bus logic with worker pool
│   ├── event_bus_test.go  # Comprehensive tests
│   ├── event_types.go     # Event definitions and types
│   ├── filter.go          # Event filtering capabilities
│   └── batch.go           # Batch processing functionality
├── examples/              # Real-world example applications
│   ├── auth_service/      # Authentication service example
│   ├── ecommerce/         # E-commerce platform example
│   └── microservices/     # Microservices communication example
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── README.md              # Main project documentation
├── IMPLEMENTATION.md      # Detailed implementation documentation
└── PROJECT_STRUCTURE.md   # Project architecture overview
```

## Key Design Decisions

1. **Async by Default**: Events are processed asynchronously to prevent blocking the publisher, allowing for better performance and responsiveness.

2. **Strongly-Typed Events**: Using a custom `EventType` type instead of raw strings to provide compile-time type safety and prevent typos.

3. **Worker Pool**: Multiple goroutines consume from a shared channel for parallel processing while maintaining load balancing.

4. **Thread Safety**: RWMutex for subscription map allows multiple readers while protecting writes, ensuring concurrent safety.

5. **Channel Buffering**: Buffered channels decouple publishers from processors and absorb traffic spikes.

6. **Middleware Pattern**: Functional approach to add cross-cutting concerns that can be composed across different handlers.

7. **Graceful Shutdown**: Context cancellation and wait groups ensure all in-flight events are processed before shutdown.

## Testing Strategy

The project includes comprehensive tests covering:
- Basic publish/subscribe functionality
- Multiple subscribers for the same event
- Unsubscribe functionality
- Synchronous publishing with error handling
- Concurrent publishing from multiple goroutines
- Graceful shutdown behavior
- Event metadata handling

Tests use atomic counters and wait times to verify asynchronous behavior. The library is designed to be testable with both synchronous and asynchronous operations.

## Performance Considerations

1. **Worker Pool Size**: Should match workload characteristics (CPU-bound vs I/O-bound handlers)
2. **Channel Buffering**: Prevents blocking publishers while allowing for traffic spikes
3. **Memory Management**: Immutable event objects and pre-allocated buffers for efficiency
4. **Concurrency Patterns**: RWMutex for read-heavy workloads, shared channel for load balancing
5. **Batch Processing**: Reduces per-event overhead for high-volume scenarios

### Documentation

The project has extensive documentation in the `README.md` and `IMPLEMENTATION.md` files. Please keep these documents up-to-date with any changes to the codebase.
