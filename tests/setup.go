// Copyright (c) 2025 SeyedAli
// Licensed under the MIT License. See LICENSE file in the project root for details.

// Package tests. setup provides reusable test infrastructure for PostgreSQL integration and benchmarks.
package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainer holds the PostgreSQL container, database handle and cleanup function.
type TestContainer struct {
	Container *postgres.PostgresContainer
	DB        *sql.DB
	Cleanup   func(ctx context.Context) error
}

// DBConfig holds database configuration for tests.
type DBConfig struct {
	Database string
	Username string
	Password string
	Image    string   // e.g. "postgres:15-alpine"
	Port     nat.Port // container internal port (usually "5432")
}

// DefaultDBConfig returns default database configuration for PostgreSQL.
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		Database: "testdb",
		Username: "test",
		Password: "test",
		Image:    "postgres:15-alpine",
		Port:     "5432",
	}
}

// -------------------------------------------- Public Functions --------------------------------------------

// SetupTestContainer creates and initializes a PostgreSQL test container and returns TestContainer with a *sql.DB.
// The returned DB is ready for use (Ping succeeded). Caller should call tc.Cleanup(ctx) when done.
func SetupTestContainer(ctx context.Context) (*TestContainer, error) {
	return SetupTestContainerWithConfig(ctx, DefaultDBConfig())
}

// SetupTestContainerWithConfig creates a PostgreSQL test container using the provided config.
func SetupTestContainerWithConfig(ctx context.Context, cfg *DBConfig) (*TestContainer, error) {
	ctr, err := createPostgresContainer(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("createPostgresContainer: %w", err)
	}

	// Get host and mapped port.
	host, err := ctr.Host(ctx)
	if err != nil {
		_ = ctr.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}
	mappedPort, err := ctr.MappedPort(ctx, cfg.Port)
	if err != nil {
		_ = ctr.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	// Build DSN for lib/pq (database/sql).
	// Note: if you prefer pgx, construct an appropriate DSN.
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, mappedPort.Port(), cfg.Username, cfg.Password, cfg.Database)

	// Open database and verify connection.
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		_ = ctr.Terminate(ctx)
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	// set reasonable connection limits for tests
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Wait/poll for DB to be ready (Ping).
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err := db.PingContext(ctx); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = db.Close()
			_ = ctr.Terminate(ctx)
			return nil, fmt.Errorf("database did not become ready in time: %w", err)
		}
		time.Sleep(250 * time.Millisecond)
	}

	cleanup := func(ctx context.Context) error {
		var firstErr error
		if db != nil {
			if err := db.Close(); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("close db: %w", err)
			}
		}
		if ctr != nil {
			if err := ctr.Terminate(ctx); err != nil && firstErr == nil {
				firstErr = fmt.Errorf("terminate container: %w", err)
			}
		}
		return firstErr
	}

	return &TestContainer{
		Container: ctr,
		DB:        db,
		Cleanup:   cleanup,
	}, nil
}

// SetupTestMain is a helper to call from TestMain to provision the DB for package tests.
// Example usage in your package's main_test.go:
//
//	func TestMain(m *testing.M) {
//	    exitCode := tests.SetupTestMain(m)
//	    os.Exit(exitCode)
//	}
func SetupTestMain(m interface{ Run() int }) int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tc, err := SetupTestContainer(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to setup test container: %v\n", err)
		return 1
	}
	defer func() {
		if err := tc.Cleanup(ctx); err != nil {
			fmt.Printf("❌ Failed to cleanup test container: %v\n", err)
		}
	}()

	// Optionally export connection info as env var for downstream tests.
	// e.g. tests will read TEST_DATABASE_DSN from env to open their own connections.
	if host, err := tc.Container.Host(ctx); err == nil {
		if port, err := tc.Container.MappedPort(ctx, "5432"); err == nil {
			dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				DefaultDBConfig().Username, DefaultDBConfig().Password, host, port.Port(), DefaultDBConfig().Database)
			_ = os.Setenv("TEST_DATABASE_DSN", dsn)
		}
	}

	fmt.Println("✅ Test environment initialized successfully!")
	return m.Run()
}

// -------------------------------------------- Private Helper Functions --------------------------------------------

// createPostgresContainer uses testcontainers' postgres helper to start a PostgreSQL container.
func createPostgresContainer(ctx context.Context, cfg *DBConfig) (*postgres.PostgresContainer, error) {
	ctr, err := postgres.Run(
		ctx,
		cfg.Image,
		postgres.WithDatabase(cfg.Database),
		postgres.WithUsername(cfg.Username),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForSQL(cfg.Port, "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
					host, port.Port(), cfg.Username, cfg.Password, cfg.Database)
			}).WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("postgres.RunContainer: %w", err)
	}

	return ctr, nil
}
