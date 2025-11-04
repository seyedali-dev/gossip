# Gossip Test Guide

Comprehensive guide for running and understanding the test suite.

## Test Structure

```
event/
├── event_bus_unit_test.go     # Core event bus unit tests
├── middleware_test.go          # Middleware functionality tests
├── filter_test.go              # Event filtering tests
├── batch_test.go               # Batch processing tests
├── priority_test.go            # Priority queue tests
├── e2e_test.go                 # End-to-end integration tests
├── integration_test.go         # Database integration tests (requires Docker)
└── testdata/
    └── seed.go                 # Test database seeding utilities
```

## Running Tests

### Quick Start - All Unit Tests

```bash
# Run all tests
go test ./event/...

# Run with verbose output
go test -v ./event/...

# Run with race detection
go test -race ./event/...

# Run with coverage
go test -cover ./event/...

# Generate coverage report
go test -coverprofile=coverage.out ./event/...
go tool cover -html=coverage.out
```

### Unit Tests Only (No Database)

```bash
# Exclude integration tests
go test -v ./event/ -short

# Run specific test file
go test -v ./event/event_bus_unit_test.go

# Run specific test
go test -v ./event/ -run TestSubscribe_RegistersHandler
```

### Integration Tests (Requires Docker)

Integration tests use testcontainers to spin up a PostgreSQL database.

**Prerequisites:**
- Docker must be running
- Sufficient disk space for PostgreSQL image

```bash
# Run integration tests
go test -v ./event/ -tags=integration

# Run only integration tests with specific pattern
go test -v ./event/ -tags=integration -run TestIntegration_
```

### Benchmark Tests

```bash
# Run all benchmarks
go test -bench=. ./event/...

# Run specific benchmark
go test -bench=BenchmarkPublish_SingleHandler ./event/

# Benchmark with memory profiling
go test -bench=. -benchmem ./event/...

# Run benchmarks multiple times for accuracy
go test -bench=. -count=5 ./event/...
```

## Test Categories

### 1. Event Bus Core Tests (`event_bus_unit_test.go`)

Tests the fundamental pub/sub functionality.

**Key Test Cases:**
- `TestNewEventBus_*` - Bus initialization
- `TestSubscribe_*` - Handler registration
- `TestPublish_*` - Async event publishing
- `TestPublishSync_*` - Sync event publishing
- `TestUnsubscribe_*` - Handler removal
- `TestShutdown_*` - Graceful shutdown
- `TestConcurrent*` - Thread safety

**Run:**
```bash
go test -v ./event/ -run TestNewEventBus
go test -v ./event/ -run TestSubscribe
go test -v ./event/ -run TestPublish
```

### 2. Middleware Tests (`middleware_test.go`)

Tests retry, timeout, recovery, logging, and chaining.

**Key Test Cases:**
- `TestWithRetry_*` - Retry logic with exponential backoff
- `TestWithTimeout_*` - Timeout enforcement
- `TestWithRecovery_*` - Panic recovery
- `TestWithLogging_*` - Execution logging
- `TestChain_*` - Middleware composition

**Run:**
```bash
go test -v ./event/ -run TestWithRetry
go test -v ./event/ -run TestChain
```

### 3. Filter Tests (`filter_test.go`)

Tests conditional event processing.

**Key Test Cases:**
- `TestNewFilteredHandler_*` - Basic filtering
- `TestFilterByMetadata_*` - Metadata filtering
- `TestAnd_*` / `TestOr_*` / `TestNot_*` - Filter combinators
- `TestComplexFilter_*` - Nested filter logic

**Run:**
```bash
go test -v ./event/ -run TestFilter
go test -v ./event/ -run TestComplexFilter
```

### 4. Batch Tests (`batch_test.go`)

Tests batch event processing.

**Key Test Cases:**
- `TestBatchProcessor_FlushesOnBatchSize` - Size-based flushing
- `TestBatchProcessor_FlushesOnPeriod` - Time-based flushing
- `TestBatchProcessor_ShutdownFlushesRemaining` - Graceful shutdown
- `TestBatchProcessor_ConcurrentAdds` - Thread safety

**Run:**
```bash
go test -v ./event/ -run TestBatchProcessor
```

### 5. Priority Tests (`priority_test.go`)

Tests priority queue implementation.

**Key Test Cases:**
- `TestPriorityQueue_PriorityOrdering` - High → Normal → Low
- `TestPriorityQueue_MixedPriorities` - Complex scenarios
- `TestPriorityQueue_ConcurrentEnqueue` - Thread safety

**Run:**
```bash
go test -v ./event/ -run TestPriorityQueue
```

### 6. End-to-End Tests (`e2e_test.go`)

Tests complete workflows and real-world scenarios.

**Key Test Cases:**
- `TestE2E_CompleteUserRegistrationFlow` - Multi-handler workflow
- `TestE2E_OrderProcessingPipeline` - Event chaining
- `TestE2E_ErrorHandlingAndRetry` - Failure recovery
- `TestE2E_HighVolumeStressTest` - Performance under load
- `TestE2E_GracefulShutdownUnderLoad` - Shutdown behavior

**Run:**
```bash
go test -v ./event/ -run TestE2E

# Skip stress tests (they're slow)
go test -v ./event/ -run TestE2E -short
```

### 7. Integration Tests (`integration_test.go`)

Tests database persistence, event replay, and transactional scenarios.

**Requirements:**
- Docker running
- Internet connection (first run downloads PostgreSQL image)

**Key Test Cases:**
- `TestIntegration_EventPersistence` - Saving events to DB
- `TestIntegration_EventReplay` - Replaying events from DB
- `TestIntegration_MetricsTracking` - Event metrics in DB
- `TestIntegration_BatchPersistence` - Batch DB inserts
- `TestIntegration_TransactionalEventPublishing` - Tx + events

**Run:**
```bash
# Make sure Docker is running
docker ps

# Run integration tests
go test -v ./event/ -tags=integration -run TestIntegration
```

## Common Test Patterns

### Testing Async Handlers

Async handlers require sleeping to wait for processing:

```go
bus.Publish(event)
time.Sleep(100 * time.Millisecond) // Wait for processing
assert.True(t, handlerWasCalled)
```

### Testing Sync Handlers

Sync handlers execute immediately:

```go
errors := bus.PublishSync(ctx, event)
assert.Empty(t, errors)
assert.True(t, handlerWasCalled) // No sleep needed
```

### Testing Concurrency

Use atomic operations for concurrent tests:

```go
counter := int32(0)

handler := func(ctx context.Context, e *event.Event) error {
    atomic.AddInt32(&counter, 1)
    return nil
}

// ... publish events concurrently ...

assert.Equal(t, int32(expectedCount), atomic.LoadInt32(&counter))
```

### Testing with Filters

```go
filter := event.FilterByMetadata("priority", "high")
handler := event.NewFilteredHandler(filter, myHandler)

// Will execute
evt1 := event.NewEvent("test", nil).WithMetadata("priority", "high")
handler(ctx, evt1)

// Will NOT execute
evt2 := event.NewEvent("test", nil).WithMetadata("priority", "low")
handler(ctx, evt2)
```

## Test Coverage Goals

- **Overall:** 80%+ coverage
- **Core bus logic:** 90%+ coverage
- **Middleware:** 85%+ coverage
- **Edge cases:** All major error paths covered

**Check coverage:**
```bash
go test -coverprofile=coverage.out ./event/...
go tool cover -func=coverage.out

# Visual HTML report
go tool cover -html=coverage.out
```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./event/...
      
      - name: Run integration tests
        run: go test -v -tags=integration ./event/...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Debugging Tests

### Verbose Output

```bash
# Show all test output
go test -v ./event/

# Show even more detail
go test -v -x ./event/
```

### Running Single Test

```bash
# Run one specific test
go test -v ./event/ -run TestSubscribe_RegistersHandler

# Run tests matching pattern
go test -v ./event/ -run "TestPublish.*"
```

### Test Timeout

```bash
# Increase timeout for slow tests
go test -v -timeout 30m ./event/

# Default timeout is 10 minutes
```

### Memory Profiling

```bash
# Generate memory profile
go test -memprofile=mem.prof ./event/

# Analyze profile
go tool pprof mem.prof
```

### CPU Profiling

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof ./event/

# Analyze profile
go tool pprof cpu.prof
```

## Troubleshooting

### Integration Tests Fail

**Issue:** `failed to setup test container`

**Solutions:**
1. Ensure Docker is running: `docker ps`
2. Check Docker daemon is accessible
3. Verify internet connection (first run downloads image)
4. Ensure sufficient disk space

### Race Detector Fails

**Issue:** `DATA RACE` detected

**Solutions:**
1. Review concurrent access patterns
2. Use `sync.Mutex` for shared state
3. Use `atomic` operations for counters
4. Check test logs for specific race location

### Flaky Tests

**Issue:** Tests pass sometimes, fail others

**Common Causes:**
1. **Timing issues:** Increase sleep durations
2. **Concurrency:** Use proper synchronization
3. **Resource cleanup:** Ensure proper `defer` usage
4. **External dependencies:** Mock or stub external services

**Fix timing issues:**
```go
// Instead of fixed sleep
time.Sleep(50 * time.Millisecond)

// Use retry with timeout
require.Eventually(t, func() bool {
    return condition == true
}, 1*time.Second, 10*time.Millisecond)
```

## Best Practices

1. **Always use `defer`** for cleanup
   ```go
   bus := event.NewEventBus(config)
   defer bus.Shutdown()
   ```

2. **Use `require` for critical assertions**
   ```go
   require.NoError(t, err) // Stops test immediately
   assert.NoError(t, err)  // Continues test
   ```

3. **Use atomic operations for counters**
   ```go
   counter := int32(0)
   atomic.AddInt32(&counter, 1)
   ```

4. **Sleep after async publishes**
   ```go
   bus.Publish(event)
   time.Sleep(100 * time.Millisecond)
   ```

5. **Clean up resources**
   ```go
   defer bus.Shutdown()
   defer processor.Shutdown()
   defer tc.Cleanup(ctx)
   ```

## Performance Benchmarks

Expected performance on modern hardware:

- **Publish (async):** 1-2 million ops/sec
- **Publish (sync):** 100k-500k ops/sec
- **Subscribe:** 500k+ ops/sec
- **Batch processing:** Depends on batch handler, but overhead is minimal

**Run benchmarks:**
```bash
go test -bench=. -benchmem ./event/...
```

## Contributing Tests

When adding new features, always include:

1. **Unit tests** - Test the feature in isolation
2. **Integration tests** - Test with real dependencies (if applicable)
3. **Benchmark tests** - Measure performance impact
4. **E2E tests** - Test complete workflows

**Test naming convention:**
```
TestName_ActionOrTheThingBeingTested
```

Examples:
- `TestEventBus_PublishesEventsAsynchronously`
- `TestMiddleware_RetriesFailedHandlers`
- `TestFilter_CombinesMultipleConditions`