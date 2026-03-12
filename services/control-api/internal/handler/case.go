package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Cor-Incorporated/Grift/services/control-api/internal/domain"
	"github.com/Cor-Incorporated/Grift/services/control-api/internal/middleware"
	"github.com/Cor-Incorporated/Grift/services/control-api/internal/service"
	"github.com/Cor-Incorporated/Grift/services/control-api/internal/store"
)

// CaseHandler handles case CRUD operations via the service layer.
type CaseHandler struct {
	svc *service.CaseService
}

// NewCaseHandler constructs a CaseHandler backed by the given service.
func NewCaseHandler(svc *service.CaseService) *CaseHandler {
	return &CaseHandler{svc: svc}
}

// RegisterCaseRoutes registers case routes.
func RegisterCaseRoutes(mux *http.ServeMux, h *CaseHandler) {
	mux.HandleFunc("GET /v1/cases", h.ListCases)
	mux.HandleFunc("POST /v1/cases", h.CreateCase)
	mux.HandleFunc("GET /v1/cases/{caseId}", h.GetCase)
	mux.HandleFunc("PATCH /v1/cases/{caseId}", h.UpdateCase)
	mux.HandleFunc("DELETE /v1/cases/{caseId}", h.DeleteCase)
}

// CreateCase handles POST /v1/cases.
func (h *CaseHandler) CreateCase(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantUUID(w, r)
	if !ok {
		return
	}

	var req createCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.Create(r.Context(), tenantID, service.CreateInput{
		Title:             req.Title,
		CaseType:          domain.CaseType(req.Type),
		ExistingSystemURL: nilIfBlank(req.ExistingSystemURL),
		CompanyName:       nilIfBlank(req.CompanyName),
		ContactName:       nilIfBlank(req.ContactName),
		ContactEmail:      nilIfBlank(req.ContactEmail),
		CreatedByUID:      nilIfBlank(middleware.UserIDFromContext(r.Context())),
	})
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"data": result})
}

// ListCases handles GET /v1/cases.
func (h *CaseHandler) ListCases(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantUUID(w, r)
	if !ok {
		return
	}

	limit, offset := parsePagination(r)
	statusFilter := r.URL.Query().Get("status")
	typeFilter := r.URL.Query().Get("type")

	records, total, err := h.svc.List(r.Context(), tenantID, statusFilter, typeFilter, limit, offset)
	if err != nil {
		writeJSONError(w, "failed to list cases", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": records, "total": total})
}

// GetCase handles GET /v1/cases/{caseId}.
func (h *CaseHandler) GetCase(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantUUID(w, r)
	if !ok {
		return
	}
	caseID, ok := parseCaseUUID(w, r)
	if !ok {
		return
	}

	record, err := h.svc.Get(r.Context(), tenantID, caseID)
	if err != nil {
		writeJSONError(w, "failed to get case", http.StatusInternalServerError)
		return
	}
	if record == nil {
		writeJSONError(w, "case not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": caseWithDetails{
			Case:          *record,
			Conversations: []store.ConversationTurn{},
			SourceDocs:    []any{},
			Estimates:     []any{},
		},
	})
}

// UpdateCase handles PATCH /v1/cases/{caseId}.
func (h *CaseHandler) UpdateCase(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantUUID(w, r)
	if !ok {
		return
	}
	caseID, ok := parseCaseUUID(w, r)
	if !ok {
		return
	}

	var req updateCaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.Update(r.Context(), tenantID, caseID, service.UpdateInput{
		Title:    req.Title,
		Type:     req.Type,
		Status:   req.Status,
		Priority: req.Priority,
	})
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSONError(w, "case not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": result})
}

// DeleteCase handles DELETE /v1/cases/{caseId}.
func (h *CaseHandler) DeleteCase(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := parseTenantUUID(w, r)
	if !ok {
		return
	}
	caseID, ok := parseCaseUUID(w, r)
	if !ok {
		return
	}

	err := h.svc.Delete(r.Context(), tenantID, caseID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeJSONError(w, "case not found", http.StatusNotFound)
			return
		}
		writeJSONError(w, "failed to delete case", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateCaseRequest is the JSON body for PATCH /v1/cases/{caseId}.
type updateCaseRequest struct {
	Title    *string `json:"title"`
	Type     *string `json:"type"`
	Status   *string `json:"status"`
	Priority *string `json:"priority"`
}

// createCaseRequest is the JSON body for POST /v1/cases.
type createCaseRequest struct {
	Title             string `json:"title"`
	Type              string `json:"type"`
	ExistingSystemURL string `json:"existing_system_url"`
	CompanyName       string `json:"company_name"`
	ContactName       string `json:"contact_name"`
	ContactEmail      string `json:"contact_email"`
}

// caseWithDetails wraps a case with related sub-resources for GET /v1/cases/{caseId}.
type caseWithDetails struct {
	domain.Case
	Conversations []store.ConversationTurn `json:"conversations"`
	SourceDocs    []any                      `json:"source_documents"`
	Estimates     []any                      `json:"estimates"`
}
