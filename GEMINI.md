# GEMINI.md - Gossip Event Bus

This document provides a comprehensive overview of the Gossip event bus library, its architecture, and development conventions. It is intended to be used as a guide for developers working with the codebase.

## Project Overview

Gossip is a lightweight, type-safe event bus library for Go that implements the observer/pub-sub pattern. It enables clean separation between core business logic and side effects, making your codebase more maintainable and extensible.

The library is designed with the following key features:

*   **Strongly-typed events**: Prevents typos and ensures consistency with `EventType` constants.
*   **Asynchronous by default**: Non-blocking event dispatch with a worker pool.
*   **Synchronous option**: For when immediate processing is required.
*   **Event filtering**: Conditional handler execution based on event properties.
*   **Batch processing**: Process multiple events efficiently.
*   **Middleware support**: Chainable middleware for retry, timeout, logging, and recovery.
*   **Priority queues**: Handle critical events first.
*   **Thread-safe**: Concurrent publish/subscribe operations.
*   **Graceful shutdown**: Waits for in-flight events to be processed.

The core of the library is the `EventBus`, which manages event subscriptions and dispatching. Events are published to the bus, and the bus is responsible for notifying all subscribed handlers.

## Building and Running

### Building the Library

To build the library, you can use the standard Go build command:

```sh
go build ./...
```

### Running Examples

The `examples` directory contains several examples of how to use the library. To run an example, navigate to its directory and use `go run`:

```sh
cd examples/microservices
go run main.go
```

### Running Tests

To run the tests, you can use the standard Go test command:

```sh
go test ./...
```

To run the tests with coverage, use the following command:

```sh
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Development Conventions

### Code Style

The codebase follows standard Go formatting and linting conventions. Please run `go fmt` and `go vet` before submitting any changes.

### Testing

All new features and bug fixes should be accompanied by tests. The tests are located in the `event` directory and follow the `_test.go` naming convention.

### Commits

Commit messages should be clear and concise, and should explain the "why" behind the change, not just the "what".

### Documentation

The project has extensive documentation in the `README.md` and `IMPLEMENTATION.md` files. Please keep these documents up-to-date with any changes to the codebase.

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
