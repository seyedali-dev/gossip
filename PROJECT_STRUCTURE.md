# Gossip - Complete Project Structure

```
gossip/
├── README.md                    # Project overview and quick start
├── LICENSE                      # MIT License
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── PROJECT_STRUCTURE.md        # This file
│
├── event_types.go              # Event type definitions and data structs
├── event_bus.go                # Core pub/sub implementation
├── event_bus_test.go           # Comprehensive tests
├── global.go                   # Singleton instance management
├── middleware.go               # Middleware: retry, timeout, logging, recovery
├── filter.go                   # Event filtering capabilities
├── priority.go                 # Priority-based event processing
├── batch.go                    # Batch event processing
│
└── examples/                   # Real-world examples
    ├── README.md               # Comprehensive usage guide
    ├── TESTING.md              # Testing guide
    │
    ├── auth_service/           # Authentication example
    │   └── main.go
    │
    ├── ecommerce/              # E-commerce example
    │   └── main.go
    │
    └── microservices/          # Microservices communication
        └── main.go
```

## Core Files

### event_types.go

- `EventType` - Strongly-typed event identifier
- `Event` - Event structure with type, timestamp, data, metadata
- `NewEvent()` - Event constructor
- `WithMetadata()` - Add metadata to events

### event_bus.go

- `EventBus` - Main pub/sub coordinator
- `EventHandler` - Handler function signature
- `Config` - Bus configuration
- `NewEventBus()` - Create new bus
- `Subscribe()` - Register handlers
- `Publish()` - Async event publishing
- `PublishSync()` - Sync event publishing
- `Shutdown()` - Graceful cleanup

### global.go

- `GetGlobalBus()` - Singleton bus instance
- `InitGlobalBus()` - Initialize with custom config
- Convenience functions for global usage

### middleware.go

- `Middleware` - Handler wrapper type
- `WithRetry()` - Retry failed handlers
- `WithTimeout()` - Add execution timeout
- `WithRecovery()` - Recover from panics
- `WithLogging()` - Log handler execution
- `Chain()` - Combine middlewares

### filter.go

- `Filter` - Event filter function type
- `NewFilteredHandler()` - Conditional handler
- `FilterByMetadata()` - Filter by metadata
- `And()`, `Or()`, `Not()` - Filter combinators

### priority.go

- `PriorityQueue` - Priority-based queue
- `PriorityEvent` - Event with priority
- Priority constants: `PriorityLow`, `PriorityNormal`, `PriorityHigh`
- `Enqueue()`, `Dequeue()` - Queue operations

### batch.go

- `BatchProcessor` - Batch event processor
- `BatchHandler` - Batch handler function type
- `BatchConfig` - Batch configuration
- `NewBatchProcessor()` - Create processor
- `AsEventHandler()` - Convert to event handler
