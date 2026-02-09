package v2

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/api"
	"github.com/rainbowmga/timetravel/service"
)

// GetVersions lists all versions of a record
func (a *API) GetVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNumber, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNumber <= 0 {
		err := api.WriteError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		api.LogError(err)
		return
	}

	versions, err := a.versionedService.ListVersions(ctx, int(idNumber))
	if err != nil {
		if err == service.ErrRecordDoesNotExist {
			err := api.WriteError(w, fmt.Sprintf("record of id %v does not exist", idNumber), http.StatusNotFound)
			api.LogError(err)
			return
		}
		err := api.WriteError(w, api.ErrInternal.Error(), http.StatusInternalServerError)
		api.LogError(err)
		return
	}

	err = api.WriteJSON(w, map[string]interface{}{
		"id":       idNumber,
		"versions": versions,
	}, http.StatusOK)
	api.LogError(err)
}
