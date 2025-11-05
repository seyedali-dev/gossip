# Gossip - Makefile for testing and development

.PHONY: help test test-unit test-integration test-all test-race test-cover bench clean

# Default target
help:
	@echo "Gossip - Available targets:"
	@echo "  make test              - Run all unit tests"
	@echo "  make test-unit         - Run unit tests only (no Docker needed)"
	@echo "  make test-integration  - Run integration tests (requires Docker)"
	@echo "  make test-all          - Run all tests including integration"
	@echo "  make test-race         - Run tests with race detector"
	@echo "  make test-cover        - Run tests with coverage report"
	@echo "  make bench             - Run benchmark tests"
	@echo "  make clean             - Clean test artifacts"

# Run unit tests only (fast, no external dependencies)
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./event/

# Run integration tests (requires Docker)
test-integration:
	@echo "Running integration tests (Docker required)..."
	@echo "Ensure Docker is running: docker ps"
	go test -v -tags=integration ./event/ -run TestIntegration

# Run all tests (unit + integration)
test-all: test-unit test-integration

# Default test target (unit tests only for speed)
test: test-unit

# Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	go test -v -race ./event/

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./event/
	go tool cover -func=coverage.out
	@echo ""
	@echo "HTML coverage report generated: coverage.html"
	go tool cover -html=coverage.out -o coverage.html

# Run benchmark tests
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./event/

# Run benchmarks multiple times for accuracy
bench-stable:
	@echo "Running benchmarks (5 iterations)..."
	go test -bench=. -benchmem -count=5 ./event/

# Run specific test
test-one:
	@echo "Usage: make test-one TEST=TestName"
	@if [ -z "$(TEST)" ]; then \
		echo "Error: TEST variable not set"; \
		echo "Example: make test-one TEST=TestSubscribe_RegistersHandler"; \
		exit 1; \
	fi
	go test -v ./event/ -run $(TEST)

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	rm -f *.prof
	go clean -testcache

# Run tests in CI mode (all tests, race detection, coverage)
ci: clean
	@echo "Running CI test suite..."
	go test -v -race -coverprofile=coverage.out ./event/
	go test -v -tags=integration ./event/ -run TestIntegration
	go tool cover -func=coverage.out

# Install test dependencies
deps:
	@echo "Installing test dependencies..."
	go get -t ./...
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run all checks (format, test)
check: fmt test-all

# Development workflow: watch for changes and run tests
watch:
	@echo "Watching for changes..."
	@command -v fswatch >/dev/null 2>&1 || { \
		echo "fswatch not installed. Install with: brew install fswatch"; \
		exit 1; \
	}
	fswatch -o event/ | xargs -n1 -I{} make test-unit

# Quick feedback loop for development
dev:
	@echo "Running quick tests..."
	go test -short ./event/
	@echo "✅ Tests passed!"