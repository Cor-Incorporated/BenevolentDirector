package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Cor-Incorporated/Grift/services/control-api/internal/domain"
	"github.com/Cor-Incorporated/Grift/services/control-api/internal/middleware"
	"github.com/google/uuid"
)

// mockVelocityStore implements github.VelocityStore for testing.
type mockVelocityStore struct {
	latestFn func(ctx context.Context, repoID, tenantID uuid.UUID) (*domain.VelocityMetric, error)
}

func (m *mockVelocityStore) Insert(_ context.Context, _ *domain.VelocityMetric) error {
	return nil
}

func (m *mockVelocityStore) LatestByRepositoryAndTenant(ctx context.Context, repoID, tenantID uuid.UUID) (*domain.VelocityMetric, error) {
	return m.latestFn(ctx, repoID, tenantID)
}

func (m *mockVelocityStore) ListByRepository(_ context.Context, _ uuid.UUID, _ int) ([]domain.VelocityMetric, error) {
	return nil, nil
}

func TestVelocityHandler_GetRepositoryVelocity(t *testing.T) {
	repoID := uuid.New()
	tenantID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	score := 75.5

	tests := []struct {
		name       string
		tenantID   string
		repoID     string
		store      *mockVelocityStore
		nilStore   bool
		wantStatus int
		wantError  string
	}{
		{
			name:       "missing tenant header",
			tenantID:   "",
			repoID:     repoID.String(),
			store:      &mockVelocityStore{},
			wantStatus: http.StatusBadRequest,
			wantError:  "missing X-Tenant-ID header",
		},
		{
			name:       "invalid tenant ID format",
			tenantID:   "not-a-uuid",
			repoID:     repoID.String(),
			store:      &mockVelocityStore{},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid X-Tenant-ID format",
		},
		{
			name:       "missing repositoryId",
			tenantID:   tenantID.String(),
			repoID:     "",
			store:      &mockVelocityStore{},
			wantStatus: http.StatusBadRequest,
			wantError:  "missing repositoryId path parameter",
		},
		{
			name:       "invalid repositoryId",
			tenantID:   tenantID.String(),
			repoID:     "bad-id",
			store:      &mockVelocityStore{},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid repositoryId format",
		},
		{
			name:       "nil store returns 503",
			tenantID:   tenantID.String(),
			repoID:     repoID.String(),
			nilStore:   true,
			wantStatus: http.StatusServiceUnavailable,
			wantError:  "velocity store not configured",
		},
		{
			name:     "not found",
			tenantID: tenantID.String(),
			repoID:   repoID.String(),
			store: &mockVelocityStore{
				latestFn: func(_ context.Context, _, _ uuid.UUID) (*domain.VelocityMetric, error) {
					return nil, fmt.Errorf("querying latest velocity metric: %w", sql.ErrNoRows)
				},
			},
			wantStatus: http.StatusNotFound,
			wantError:  "no velocity metrics found for this repository",
		},
		{
			name:     "store internal error",
			tenantID: tenantID.String(),
			repoID:   repoID.String(),
			store: &mockVelocityStore{
				latestFn: func(_ context.Context, _, _ uuid.UUID) (*domain.VelocityMetric, error) {
					return nil, errors.New("db connection failed")
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal server error",
		},
		{
			name:     "success",
			tenantID: tenantID.String(),
			repoID:   repoID.String(),
			store: &mockVelocityStore{
				latestFn: func(_ context.Context, _, _ uuid.UUID) (*domain.VelocityMetric, error) {
					return &domain.VelocityMetric{
						ID:            uuid.New(),
						TenantID:      tenantID,
						RepositoryID:  repoID,
						VelocityScore: &score,
						AnalyzedAt:    time.Now(),
						CreatedAt:     time.Now(),
					}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h *VelocityHandler
			if tt.nilStore {
				h = NewVelocityHandler(nil)
			} else {
				h = NewVelocityHandler(tt.store)
			}

			mux := http.NewServeMux()
			mux.HandleFunc("GET /v1/repositories/{repositoryId}/velocity", h.GetRepositoryVelocity)

			// For "missing repositoryId" test, register a route without the path param.
			if tt.repoID == "" {
				mux = http.NewServeMux()
				mux.HandleFunc("GET /v1/repositories/velocity", h.GetRepositoryVelocity)
			}

			path := "/v1/repositories/" + tt.repoID + "/velocity"
			if tt.repoID == "" {
				path = "/v1/repositories/velocity"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			if tt.tenantID != "" {
				req.Header.Set("X-Tenant-ID", tt.tenantID)
			}
			rec := httptest.NewRecorder()

			// Wrap with Tenant middleware (no-DB variant) to inject tenant context.
			middleware.Tenant(mux).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
				return
			}

			if tt.wantError != "" {
				var body map[string]string
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode error body: %v", err)
				}
				if body["error"] != tt.wantError {
					t.Errorf("error = %q, want %q", body["error"], tt.wantError)
				}
			}

			if tt.wantStatus == http.StatusOK {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode success body: %v", err)
				}
				if _, ok := body["data"]; !ok {
					t.Error("success response missing 'data' key")
				}
			}
		})
	}
}
