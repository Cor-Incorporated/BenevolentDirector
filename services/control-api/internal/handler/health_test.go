package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Health() status = %d, want %d", rec.Code, http.StatusOK)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Health() Content-Type = %q, want %q", ct, "application/json")
	}

	var body healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("Health() failed to decode response: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf("Health() status = %q, want %q", body.Status, "ok")
	}
	if body.Service != "control-api" {
		t.Errorf("Health() service = %q, want %q", body.Service, "control-api")
	}
}
