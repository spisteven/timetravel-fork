package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rainbowmga/timetravel/database"
	"github.com/rainbowmga/timetravel/entity"
)

// SQLiteRecordService implements RecordService using SQLite database
type SQLiteRecordService struct {
	db *database.DB
}

// NewSQLiteRecordService creates a new SQLiteRecordService instance
func NewSQLiteRecordService(db *database.DB) *SQLiteRecordService {
	return &SQLiteRecordService{db: db}
}

// GetRecord retrieves a record by ID
func (s *SQLiteRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	if id <= 0 {
		return entity.Record{}, ErrRecordIDInvalid
	}

	var dataJSON string
	err := s.db.QueryRowContext(ctx, "SELECT data FROM records WHERE id = ?", id).Scan(&dataJSON)
	if err == sql.ErrNoRows {
		return entity.Record{}, ErrRecordDoesNotExist
	}
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to query record: %w", err)
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return entity.Record{}, fmt.Errorf("failed to unmarshal record data: %w", err)
	}

	return entity.Record{
		ID:   id,
		Data: data,
	}, nil
}

// CreateRecord inserts a new record
func (s *SQLiteRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	id := record.ID
	if id <= 0 {
		return ErrRecordIDInvalid
	}

	// Check if record already exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM records WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check record existence: %w", err)
	}
	if exists {
		return ErrRecordAlreadyExists
	}

	// Serialize data to JSON
	dataJSON, err := json.Marshal(record.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal record data: %w", err)
	}

	// Insert record
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO records (id, data, created_at, updated_at) VALUES (?, ?, ?, ?)",
		id, string(dataJSON), time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	return nil
}

// UpdateRecord updates an existing record's data
func (s *SQLiteRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	if id <= 0 {
		return entity.Record{}, ErrRecordIDInvalid
	}

	// Get current record
	record, err := s.GetRecord(ctx, id)
	if err != nil {
		return entity.Record{}, err
	}

	// Apply updates
	for key, value := range updates {
		if value == nil {
			// Delete key
			delete(record.Data, key)
		} else {
			// Update or add key
			record.Data[key] = *value
		}
	}

	// Serialize updated data to JSON
	dataJSON, err := json.Marshal(record.Data)
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to marshal record data: %w", err)
	}

	// Update record in database
	_, err = s.db.ExecContext(ctx,
		"UPDATE records SET data = ?, updated_at = ? WHERE id = ?",
		string(dataJSON), time.Now(), id,
	)
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to update record: %w", err)
	}

	return record, nil
}
