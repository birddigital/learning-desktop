// Command migrate runs database migrations for Learning Desktop
//
// Usage:
//   migrate up              - Run all pending migrations
//   migrate down            - Rollback last migration
//   migrate status          - Show migration status
//   migrate version         - Show current version
//   migrate create name     - Create new migration files
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	migrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
		version INT PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`

	selectVersion = `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`
	insertVersion = `INSERT INTO schema_migrations (version) VALUES ($1)`
	deleteVersion = `DELETE FROM schema_migrations WHERE version = $1`
)

var migrationsDir = "migrations"

type Migration struct {
	Version int
	Up      string
	Down    string
}

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	// Get database URL from env or default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost/learning_desktop?sslmode=disable"
	}

	db, err := sqlx.Connect("pgx", dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize migrations table
	if _, err := db.Exec(migrationsTable); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating migrations table: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runUp(db)
	case "down":
		runDown(db)
	case "status":
		showStatus(db)
	case "version":
		showVersion(db)
	case "create":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: migrate create <name>")
			os.Exit(1)
		}
		createMigration(os.Args[2])
	case "redo":
		runDown(db)
		runUp(db)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("Learning Desktop Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate up              Run all pending migrations")
	fmt.Println("  migrate down            Rollback last migration")
	fmt.Println("  migrate status          Show migration status")
	fmt.Println("  migrate version         Show current version")
	fmt.Println("  migrate create NAME     Create new migration files")
	fmt.Println("  migrate redo            Rollback and re-run last migration")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  DATABASE_URL           PostgreSQL connection string")
	fmt.Println("                         (default: postgres://localhost/learning_desktop?sslmode=disable)")
}

func loadMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("reading migrations dir: %w", err)
	}

	var migrations []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}

		// Parse version from filename (e.g., "001_schema.up.sql" -> version 1)
		name := strings.TrimSuffix(e.Name(), ".up.sql")
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		var version int
		if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
			continue
		}

		upPath := filepath.Join(migrationsDir, e.Name())
		downPath := filepath.Join(migrationsDir, fmt.Sprintf("%s.down.sql", parts[0]))

		migrations = append(migrations, Migration{
			Version: version,
			Up:      upPath,
			Down:    downPath,
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func runUp(db *sqlx.DB) {
	migrations, err := loadMigrations()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading migrations: %v\n", err)
		os.Exit(1)
	}

	// Get current version
	var current int
	if err := db.Get(&current, selectVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current version: %v\n", err)
		os.Exit(1)
	}

	// Find pending migrations
	var pending []Migration
	for _, m := range migrations {
		if m.Version > current {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		fmt.Println("No pending migrations.")
		return
	}

	fmt.Printf("Running %d migration(s):\n", len(pending))
	for _, m := range pending {
		if err := runMigration(db, m, true); err != nil {
			fmt.Fprintf(os.Stderr, "Error running migration %d: %v\n", m.Version, err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ %d\n", m.Version)
	}
	fmt.Println("Migrations complete!")
}

func runDown(db *sqlx.DB) {
	// Get current version
	var current int
	if err := db.Get(&current, selectVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current version: %v\n", err)
		os.Exit(1)
	}

	if current == 0 {
		fmt.Println("No migrations to rollback.")
		return
	}

	// Load migrations and find the one at current version
	migrations, err := loadMigrations()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading migrations: %v\n", err)
		os.Exit(1)
	}

	var target *Migration
	for _, m := range migrations {
		if m.Version == current {
			target = &m
			break
		}
	}

	if target == nil {
		fmt.Fprintf(os.Stderr, "Migration %d not found\n", current)
		os.Exit(1)
	}

	fmt.Printf("Rolling back migration %d...\n", target.Version)
	if err := runMigration(db, *target, false); err != nil {
		fmt.Fprintf(os.Stderr, "Error rolling back migration: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Rollback complete!")
}

func runMigration(db *sqlx.DB, m Migration, up bool) error {
	file := m.Up
	if !up {
		file = m.Down
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading migration file: %w", err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}

	// Update version
	if up {
		if _, err := tx.Exec(insertVersion, m.Version); err != nil {
			return fmt.Errorf("recording version: %w", err)
		}
	} else {
		if _, err := tx.Exec(deleteVersion, m.Version); err != nil {
			return fmt.Errorf("removing version: %w", err)
		}
	}

	return tx.Commit()
}

func showStatus(db *sqlx.DB) {
	migrations, err := loadMigrations()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading migrations: %v\n", err)
		os.Exit(1)
	}

	// Get current version
	var current int
	if err := db.Get(&current, selectVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current version: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration Status:")
	fmt.Println()

	for _, m := range migrations {
		status := "pending"
		if m.Version <= current {
			status = "applied"
		}
		fmt.Printf("  %s %3d %s\n", map[bool]string{true: "✓", false: "→"}[status == "applied"], m.Version, status)
	}
}

func showVersion(db *sqlx.DB) {
	var current int
	if err := db.Get(&current, selectVersion); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current version: %v\n", err)
		os.Exit(1)
	}

	if current == 0 {
		fmt.Println("No migrations applied.")
	} else {
		fmt.Printf("Current version: %d\n", current)
	}
}

func createMigration(name string) {
	// Find next version number
	migrations, err := loadMigrations()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading migrations: %v\n", err)
		os.Exit(1)
	}

	var nextVersion int
	if len(migrations) > 0 {
		last := migrations[len(migrations)-1]
		nextVersion = last.Version + 1
	} else {
		nextVersion = 1
	}

	prefix := fmt.Sprintf("%03d", nextVersion)

	upPath := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.up.sql", prefix, name))
	downPath := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.down.sql", prefix, name))

	// Check if files already exist
	if _, err := os.Stat(upPath); err == nil {
		fmt.Fprintf(os.Stderr, "Migration file already exists: %s\n", upPath)
		os.Exit(1)
	}

	// Create up file
	upContent := fmt.Sprintf("-- Migration %d: %s\n-- %s\n\n",
		nextVersion, name, time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating up file: %v\n", err)
		os.Exit(1)
	}

	// Create down file
	downContent := fmt.Sprintf("-- Rollback migration %d: %s\n\n", nextVersion, name)
	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating down file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created migration files:\n  %s\n  %s\n", upPath, downPath)

	// Open in editor if EDITOR is set
	if editor := os.Getenv("EDITOR"); editor != "" {
		fmt.Printf("\nOpening %s in %s...\n", upPath, editor)
		// Open editor (simplified - just print the command)
		fmt.Printf("Tip: Run '%s %s' to edit\n", editor, upPath)
	}
}

func init() {
	// Allow running from any directory
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try relative to binary
		if exe, err := os.Executable(); err == nil {
			migrationsDir = filepath.Join(filepath.Dir(exe), migrationsDir)
			if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
				// Try parent directory
				migrationsDir = filepath.Join(filepath.Dir(exe), "..", migrationsDir)
			}
		}
	}
}

// confirmAction prompts for user confirmation
func confirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", prompt)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}
