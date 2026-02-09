package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

var (
	ErrInternal = errors.New("internal error")
)

// LogError logs an error if it's not nil (exported for use by v2 API)
func LogError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

// logs an error if it's not nil (kept for backward compatibility with v1)
func logError(err error) {
	LogError(err)
}

// WriteJSON writes the data as json (exported for use by v2 API)
func WriteJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	return err
}

// writeJSON writes the data as json (kept for backward compatibility with v1)
func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	return WriteJSON(w, data, statusCode)
}

// WriteError writes the message as an error (exported for use by v2 API)
func WriteError(w http.ResponseWriter, message string, statusCode int) error {
	log.Printf("response errored: %s", message)
	return WriteJSON(
		w,
		map[string]string{"error": message},
		statusCode,
	)
}

// writeError writes the message as an error (kept for backward compatibility with v1)
func writeError(w http.ResponseWriter, message string, statusCode int) error {
	return WriteError(w, message, statusCode)
}
