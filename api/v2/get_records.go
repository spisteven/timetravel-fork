package v2

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/api"
	"github.com/rainbowmga/timetravel/service"
)

// GetRecord retrieves the latest version of a record (v2 API)
func (a *API) GetRecord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	idNumber, err := strconv.ParseInt(id, 10, 32)
	if err != nil || idNumber <= 0 {
		err := api.WriteError(w, "invalid id; id must be a positive number", http.StatusBadRequest)
		api.LogError(err)
		return
	}

	record, err := a.versionedService.GetRecord(ctx, int(idNumber))
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

	err = api.WriteJSON(w, record, http.StatusOK)
	api.LogError(err)
}
