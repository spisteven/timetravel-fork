package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/api"
	v2api "github.com/rainbowmga/timetravel/api/v2"
	"github.com/rainbowmga/timetravel/database"
	"github.com/rainbowmga/timetravel/service"
)

// logError logs all non-nil errors
func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func main() {
	// Initialize database
	db, err := database.NewDB(database.DefaultDBPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	router := mux.NewRouter()

	// Use SQLiteRecordService for v1 API (backward compatibility)
	recordService := service.NewSQLiteRecordService(db)
	v1API := api.NewAPI(recordService)

	// Use SQLiteVersionedRecordService for v2 API (with versioning)
	versionedService := service.NewSQLiteVersionedRecordService(db)
	v2API := v2api.NewAPI(versionedService)

	// Register v1 routes
	v1Route := router.PathPrefix("/api/v1").Subrouter()
	v1Route.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		logError(err)
	})
	v1API.CreateRoutes(v1Route)

	// Register v2 routes
	v2Route := router.PathPrefix("/api/v2").Subrouter()
	v2API.CreateRoutes(v2Route)

	address := "127.0.0.1:8000"
	srv := &http.Server{
		Handler:      router,
		Addr:         address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("listening on %s", address)
	log.Fatal(srv.ListenAndServe())
}
