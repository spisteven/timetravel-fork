package v2

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/api"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/service"
)

// PostRecord creates or updates a record with versioning (v2 API)
func (a *API) PostRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]
	idNumber, err := strconv.ParseInt(id, 10, 32)

	if err != nil || idNumber <= 0 {
		err := api.WriteError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		api.LogError(err)
		return
	}

	var body map[string]*string
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		err := api.WriteError(w, "invalid input; could not parse json", http.StatusBadRequest)
		api.LogError(err)
		return
	}

	// Check if record exists
	_, err = a.versionedService.GetRecord(ctx, int(idNumber))
	recordExists := !errors.Is(err, service.ErrRecordDoesNotExist)

	var record entity.Record
	if recordExists {
		// Update existing record
		record, err = a.versionedService.UpdateRecord(ctx, int(idNumber), body)
	} else {
		// Create new record - exclude null values
		recordMap := map[string]string{}
		for key, value := range body {
			if value != nil {
				recordMap[key] = *value
			}
		}

		record = entity.Record{
			ID:   int(idNumber),
			Data: recordMap,
		}
		err = a.versionedService.CreateRecord(ctx, record)
	}

	if err != nil {
		if err == service.ErrRecordAlreadyExists {
			// This shouldn't happen, but handle it gracefully
			err := api.WriteError(w, api.ErrInternal.Error(), http.StatusInternalServerError)
			api.LogError(err)
			return
		}
		err := api.WriteError(w, api.ErrInternal.Error(), http.StatusInternalServerError)
		api.LogError(err)
		return
	}

	err = api.WriteJSON(w, record, http.StatusOK)
	api.LogError(err)
}
