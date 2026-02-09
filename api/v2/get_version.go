package v2

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/api"
	"github.com/rainbowmga/timetravel/service"
)

// GetRecordVersion retrieves a record at a specific version
func (a *API) GetRecordVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]
	versionStr := vars["version"]

	idNumber, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNumber <= 0 {
		err := api.WriteError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		api.LogError(err)
		return
	}

	versionNumber, err := strconv.ParseInt(versionStr, 10, 32)
	if err != nil || versionNumber <= 0 {
		err := api.WriteError(w, "invalid version; version must be a positive number", http.StatusBadRequest)
		api.LogError(err)
		return
	}

	record, err := a.versionedService.GetRecordVersion(ctx, int(idNumber), int(versionNumber))
	if err != nil {
		if err == service.ErrRecordDoesNotExist || err == service.ErrVersionDoesNotExist {
			err := api.WriteError(w, fmt.Sprintf("record version %v@%v does not exist", idNumber, versionNumber), http.StatusNotFound)
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
