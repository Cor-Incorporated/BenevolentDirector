package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Cor-Incorporated/BenevolentDirector/services/control-api/internal/github"
	"github.com/Cor-Incorporated/BenevolentDirector/services/control-api/internal/middleware"
	"github.com/google/uuid"
)

// VelocityHandler handles velocity metric HTTP requests.
type VelocityHandler struct {
	store github.VelocityStore
}

// NewVelocityHandler creates a new VelocityHandler with the given store.
func NewVelocityHandler(store github.VelocityStore) *VelocityHandler {
	return &VelocityHandler{store: store}
}

// GetRepositoryVelocity returns the latest velocity metric for a repository.
// GET /v1/repositories/{repositoryId}/velocity
func (h *VelocityHandler) GetRepositoryVelocity(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("missing tenant context"))
		return
	}

	repoIDStr := r.PathValue("repositoryId")
	if repoIDStr == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("missing repositoryId path parameter"))
		return
	}

	repoID, err := uuid.Parse(repoIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid repositoryId format"))
		return
	}

	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorBody("velocity store not configured"))
		return
	}

	metric, err := h.store.LatestByRepository(r.Context(), repoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, errorBody("no velocity metrics found for this repository"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}

	// Verify the metric belongs to the requesting tenant
	parsedTenantID, err := uuid.Parse(tenantID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid tenant ID"))
		return
	}
	if metric.TenantID != parsedTenantID {
		writeJSON(w, http.StatusForbidden, errorBody("access denied"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": metric})
}

// writeJSON and errorBody are defined in repository.go (shared handler helpers)
