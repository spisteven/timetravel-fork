package entity

import "time"

// RecordVersion represents a specific version of a record
type RecordVersion struct {
	ID        int               `json:"id"`
	RecordID  int               `json:"record_id"`
	Version   int               `json:"version"`
	Data      map[string]string `json:"data"`
	CreatedAt time.Time         `json:"created_at"`
}

// VersionInfo contains metadata about a version
type VersionInfo struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}
