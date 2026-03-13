package handler

import (
	"errors"
	"net/http"

	artifacttrigger "github.com/Cor-Incorporated/Grift/services/control-api/internal/requirementartifact"
	"github.com/Cor-Incorporated/Grift/services/control-api/internal/store"
)

// RequirementArtifactHandler provides HTTP handlers for requirement artifact endpoints.
type RequirementArtifactHandler struct {
	store   store.RequirementArtifactStore
	trigger artifacttrigger.Trigger
}

// NewRequirementArtifactHandler creates a RequirementArtifactHandler with the given store.
func NewRequirementArtifactHandler(
	s store.RequirementArtifactStore,
	trigger artifacttrigger.Trigger,
) *RequirementArtifactHandler {
	return &RequirementArtifactHandler{store: s, trigger: trigger}
}

// GetLatestByCaseID handles GET /v1/cases/{caseId}/requirement-artifact.
// Returns the latest versioned requirement artifact for the given case.
func (h *RequirementArtifactHandler) GetLatestByCaseID(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorBody("requirement artifact store not configured"))
		return
	}

	tenantID, ok := parseTenantUUID(w, r)
	if !ok {
		return
	}
	caseID, ok := parseCaseUUID(w, r)
	if !ok {
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

// TriggerGeneration handles POST /v1/cases/{caseId}/requirement-artifact.
func (h *RequirementArtifactHandler) TriggerGeneration(w http.ResponseWriter, r *http.Request) {
	if h.trigger == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorBody("requirement artifact trigger not configured"))
		return
	}

	tenantID, ok := parseTenantUUID(w, r)
	if !ok {
		return
	}
	caseID, ok := parseCaseUUID(w, r)
	if !ok {
		return
	}

	result, err := h.trigger.Trigger(r.Context(), tenantID, caseID)
	if err == nil {
		writeJSON(w, http.StatusAccepted, result)
		return
	}

	var thresholdErr *artifacttrigger.CompletenessThresholdError
	switch {
	case errors.Is(err, artifacttrigger.ErrCompletenessObservationNotFound):
		writeJSON(w, http.StatusNotFound, errorBody("completeness observation not found"))
	case errors.As(err, &thresholdErr):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":                 thresholdErr.Error(),
			"overall_completeness":  thresholdErr.OverallCompleteness,
			"suggested_next_topics": thresholdErr.SuggestedNextTopics,
		})
	default:
		writeJSON(w, http.StatusInternalServerError, errorBody("internal server error"))
	}
}
