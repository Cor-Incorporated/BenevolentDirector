package handler

import (
	"net/http"

	"github.com/Cor-Incorporated/Grift/services/control-api/internal/middleware"
	"github.com/Cor-Incorporated/Grift/services/control-api/internal/store"
	"github.com/google/uuid"
)

// RequirementArtifactHandler provides HTTP handlers for requirement artifact endpoints.
type RequirementArtifactHandler struct {
	store store.RequirementArtifactStore
}

// NewRequirementArtifactHandler creates a RequirementArtifactHandler with the given store.
func NewRequirementArtifactHandler(s store.RequirementArtifactStore) *RequirementArtifactHandler {
	return &RequirementArtifactHandler{store: s}
}

// GetLatestByCaseID handles GET /v1/cases/{caseId}/requirement-artifact.
// Returns the latest versioned requirement artifact for the given case.
func (h *RequirementArtifactHandler) GetLatestByCaseID(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorBody("requirement artifact store not configured"))
		return
	}

	tenantIDStr := middleware.TenantIDFromContext(r.Context())
	if tenantIDStr == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("missing tenant context"))
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid tenant ID"))
		return
	}

	caseIDStr := r.PathValue("caseId")
	if caseIDStr == "" {
		writeJSON(w, http.StatusBadRequest, errorBody("missing caseId path parameter"))
		return
	}

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody("invalid caseId format"))
		return
	}

	artifact, err := h.store.GetLatestByCaseID(r.Context(), tenantID, caseID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
		return
	}
	if artifact == nil {
		writeJSON(w, http.StatusNotFound, errorBody("no requirement artifact found for this case"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": artifact})
}
