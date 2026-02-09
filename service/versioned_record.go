package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rainbowmga/timetravel/database"
	"github.com/rainbowmga/timetravel/entity"
)

var (
	ErrVersionDoesNotExist = errors.New("version does not exist")
	ErrInvalidVersion      = errors.New("invalid version number")
)

// VersionedRecordService extends RecordService with versioning capabilities
type VersionedRecordService interface {
	RecordService

	// GetRecordVersion retrieves a record at a specific version
	GetRecordVersion(ctx context.Context, id int, version int) (entity.Record, error)

	// ListVersions returns all versions for a record
	ListVersions(ctx context.Context, id int) ([]entity.VersionInfo, error)

	// CreateOrUpdateRecord creates or updates a record while preserving history
	CreateOrUpdateRecord(ctx context.Context, record entity.Record) (entity.Record, error)
}

// SQLiteVersionedRecordService implements VersionedRecordService using SQLite
type SQLiteVersionedRecordService struct {
	db *database.DB
}

// NewSQLiteVersionedRecordService creates a new SQLiteVersionedRecordService instance
func NewSQLiteVersionedRecordService(db *database.DB) *SQLiteVersionedRecordService {
	return &SQLiteVersionedRecordService{db: db}
}

// GetRecord retrieves the latest version of a record
func (s *SQLiteVersionedRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
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

// GetRecordVersion retrieves a record at a specific version
func (s *SQLiteVersionedRecordService) GetRecordVersion(ctx context.Context, id int, version int) (entity.Record, error) {
	if id <= 0 {
		return entity.Record{}, ErrRecordIDInvalid
	}
	if version <= 0 {
		return entity.Record{}, ErrInvalidVersion
	}

	var dataJSON string
	err := s.db.QueryRowContext(ctx,
		"SELECT data FROM record_versions WHERE record_id = ? AND version = ?",
		id, version,
	).Scan(&dataJSON)
	if err == sql.ErrNoRows {
		return entity.Record{}, ErrVersionDoesNotExist
	}
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to query record version: %w", err)
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

// ListVersions returns all versions for a record, ordered by version descending
func (s *SQLiteVersionedRecordService) ListVersions(ctx context.Context, id int) ([]entity.VersionInfo, error) {
	if id <= 0 {
		return nil, ErrRecordIDInvalid
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT version, created_at FROM record_versions WHERE record_id = ? ORDER BY version DESC",
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var versions []entity.VersionInfo
	for rows.Next() {
		var v entity.VersionInfo
		if err := rows.Scan(&v.Version, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating versions: %w", err)
	}

	// If no versions found, check if record exists
	if len(versions) == 0 {
		var exists bool
		err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM records WHERE id = ?)", id).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("failed to check record existence: %w", err)
		}
		if !exists {
			return nil, ErrRecordDoesNotExist
		}
	}

	return versions, nil
}

// CreateRecord inserts a new record and creates its first version
func (s *SQLiteVersionedRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
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

	now := time.Now()

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert record
	_, err = tx.ExecContext(ctx,
		"INSERT INTO records (id, data, created_at, updated_at) VALUES (?, ?, ?, ?)",
		id, string(dataJSON), now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	// Insert first version
	_, err = tx.ExecContext(ctx,
		"INSERT INTO record_versions (record_id, version, data, created_at) VALUES (?, ?, ?, ?)",
		id, 1, string(dataJSON), now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert record version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateRecord updates a record and creates a new version
func (s *SQLiteVersionedRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
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
			delete(record.Data, key)
		} else {
			record.Data[key] = *value
		}
	}

	// Serialize updated data to JSON
	dataJSON, err := json.Marshal(record.Data)
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to marshal record data: %w", err)
	}

	now := time.Now()

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get next version number
	var nextVersion int
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version), 0) + 1 FROM record_versions WHERE record_id = ?",
		id,
	).Scan(&nextVersion)
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to get next version: %w", err)
	}

	// Update record in database
	_, err = tx.ExecContext(ctx,
		"UPDATE records SET data = ?, updated_at = ? WHERE id = ?",
		string(dataJSON), now, id,
	)
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to update record: %w", err)
	}

	// Insert new version
	_, err = tx.ExecContext(ctx,
		"INSERT INTO record_versions (record_id, version, data, created_at) VALUES (?, ?, ?, ?)",
		id, nextVersion, string(dataJSON), now,
	)
	if err != nil {
		return entity.Record{}, fmt.Errorf("failed to insert record version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return entity.Record{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return record, nil
}

// CreateOrUpdateRecord creates a new record or updates an existing one, preserving history
func (s *SQLiteVersionedRecordService) CreateOrUpdateRecord(ctx context.Context, record entity.Record) (entity.Record, error) {
	id := record.ID
	if id <= 0 {
		return entity.Record{}, ErrRecordIDInvalid
	}

	// Check if record exists
	_, err := s.GetRecord(ctx, id)
	if err != nil && !errors.Is(err, ErrRecordDoesNotExist) {
		return entity.Record{}, err
	}

	if errors.Is(err, ErrRecordDoesNotExist) {
		// Create new record
		if err := s.CreateRecord(ctx, record); err != nil {
			return entity.Record{}, err
		}
		return record, nil
	}

	// Update existing record - merge updates
	updates := make(map[string]*string)
	for key, value := range record.Data {
		val := value
		updates[key] = &val
	}

	return s.UpdateRecord(ctx, id, updates)
}
