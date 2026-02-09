package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DefaultDBPath = "timetravel.db"
)

// DB wraps the sql.DB connection
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection and initializes the schema
func NewDB(dbPath string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	database := &DB{DB: db}

	if err := database.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return database, nil
}

// initSchema creates the necessary tables for records and versions
func (db *DB) initSchema() error {
	// Records table stores the current state of each record
	createRecordsTable := `
	CREATE TABLE IF NOT EXISTS records (
		id INTEGER PRIMARY KEY CHECK(id > 0),
		data TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Record versions table stores historical versions of records
	createVersionsTable := `
	CREATE TABLE IF NOT EXISTS record_versions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id INTEGER NOT NULL,
		version INTEGER NOT NULL,
		data TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (record_id) REFERENCES records(id) ON DELETE CASCADE,
		UNIQUE(record_id, version)
	);
	`

	// Index for faster lookups
	createIndexes := `
	CREATE INDEX IF NOT EXISTS idx_record_versions_record_id ON record_versions(record_id);
	CREATE INDEX IF NOT EXISTS idx_record_versions_record_id_version ON record_versions(record_id, version);
	`

	if _, err := db.Exec(createRecordsTable); err != nil {
		return fmt.Errorf("failed to create records table: %w", err)
	}

	if _, err := db.Exec(createVersionsTable); err != nil {
		return fmt.Errorf("failed to create record_versions table: %w", err)
	}

	if _, err := db.Exec(createIndexes); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
