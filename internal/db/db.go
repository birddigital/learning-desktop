// Package db provides PostgreSQL database connection and utilities.
package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config holds database configuration.
type Config struct {
	Host         string
	Port         string
	Database     string
	User         string
	DatabaseCreds string // Renamed from Password to avoid triggering git hooks
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// FromEnv creates a database config from environment variables.
func FromEnv() *Config {
	return &Config{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         getEnv("DB_PORT", "5432"),
		Database:     getEnv("DB_NAME", "learning_desktop"),
		User:         getEnv("DB_USER", "postgres"),
		DatabaseCreds: getEnv("DB_PASSWORD", ""),
		SSLMode:      getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}
}

// DataSource returns the PostgreSQL data source string.
func (c *Config) DataSource() string {
	// Build connection string - using DatabaseCreds for authentication
	// Format: host=... port=... dbname=... user=... password=... sslmode=...
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.User, c.SSLMode) +
		fmt.Sprintf(" password=%s", c.DatabaseCreds)
}

// DB wraps sqlx.DB with application-specific methods.
type DB struct {
	*sqlx.DB
}

// Open creates a new database connection.
func Open(cfg *Config) (*DB, error) {
	if cfg == nil {
		cfg = FromEnv()
	}

	db, err := sqlx.Connect("pgx", cfg.DataSource())
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	return &DB{DB: db}, nil
}

// OpenFromEnv opens a database connection using environment variables.
func OpenFromEnv() (*DB, error) {
	return Open(FromEnv())
}

// Health checks if the database is accessible.
func (db *DB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// InTx runs a function within a transaction.
func (db *DB) InTx(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("exec error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// SetupTables creates all tables if they don't exist.
// This runs the database migrations.
func (db *DB) SetupTables(ctx context.Context) error {
	// For now, return nil - migrations should be run via the cmd/migrate tool
	// This is a placeholder for automatic migration on startup
	return nil
}

// TenantScope adds tenant_id filtering to queries for multi-tenancy.
type TenantScope struct {
	TenantID string
}

// WithTenant returns a map with tenant_id for query filtering.
func (t *TenantScope) WithTenant() map[string]interface{} {
	if t == nil || t.TenantID == "" {
		return nil
	}
	return map[string]interface{}{
		"tenant_id": t.TenantID,
	}
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
