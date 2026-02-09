package v2

import (
	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/service"
)

// API handles v2 API endpoints with versioning support
type API struct {
	versionedService service.VersionedRecordService
}

// NewAPI creates a new v2 API instance
func NewAPI(versionedService service.VersionedRecordService) *API {
	return &API{
		versionedService: versionedService,
	}
}

// CreateRoutes registers all v2 API routes
func (a *API) CreateRoutes(routes *mux.Router) {
	// GET /api/v2/records/{id} - get latest version
	routes.Path("/records/{id}").HandlerFunc(a.GetRecord).Methods("GET")

	// GET /api/v2/records/{id}/versions - list all versions
	routes.Path("/records/{id}/versions").HandlerFunc(a.GetVersions).Methods("GET")

	// GET /api/v2/records/{id}/versions/{version} - get specific version
	routes.Path("/records/{id}/versions/{version}").HandlerFunc(a.GetRecordVersion).Methods("GET")

	// POST /api/v2/records/{id} - create or update with versioning
	routes.Path("/records/{id}").HandlerFunc(a.PostRecord).Methods("POST")
}
