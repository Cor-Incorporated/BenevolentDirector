package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	// Override env for deterministic tests.
	t.Setenv(corsEnvKey, "http://localhost:5173,https://app.example.com")

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := CORS()(inner)

	tests := []struct {
		name           string
		method         string
		path           string
		origin         string
		wantStatus     int
		wantCORSOrigin string // expected Access-Control-Allow-Origin value ("" means absent)
	}{
		{
			name:           "preflight with allowed origin returns 204",
			method:         http.MethodOptions,
			path:           "/v1/cases",
			origin:         "http://localhost:5173",
			wantStatus:     http.StatusNoContent,
			wantCORSOrigin: "http://localhost:5173",
		},
		{
			name:           "GET with allowed origin sets CORS header",
			method:         http.MethodGet,
			path:           "/v1/cases",
			origin:         "https://app.example.com",
			wantStatus:     http.StatusOK,
			wantCORSOrigin: "https://app.example.com",
		},
		{
			name:           "disallowed origin does not get CORS header",
			method:         http.MethodGet,
			path:           "/v1/cases",
			origin:         "https://evil.example.com",
			wantStatus:     http.StatusOK,
			wantCORSOrigin: "",
		},
		{
			name:           "health path skips CORS",
			method:         http.MethodGet,
			path:           "/health",
			origin:         "http://localhost:5173",
			wantStatus:     http.StatusOK,
			wantCORSOrigin: "",
		},
		{
			name:           "no origin header omits CORS headers",
			method:         http.MethodGet,
			path:           "/v1/cases",
			origin:         "",
			wantStatus:     http.StatusOK,
			wantCORSOrigin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			got := rec.Header().Get("Access-Control-Allow-Origin")
			if got != tt.wantCORSOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantCORSOrigin)
			}
		})
	}
}

func TestCORS_PreflightHeaders(t *testing.T) {
	t.Setenv(corsEnvKey, "http://localhost:5173")

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("inner handler should not be called for OPTIONS preflight")
	})
	handler := CORS()(inner)

	req := httptest.NewRequest(http.MethodOptions, "/v1/cases", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if v := rec.Header().Get("Access-Control-Allow-Methods"); v != "GET,POST,PUT,DELETE,OPTIONS" {
		t.Errorf("Allow-Methods = %q, want GET,POST,PUT,DELETE,OPTIONS", v)
	}
	if v := rec.Header().Get("Access-Control-Allow-Headers"); v != "Content-Type, Authorization, X-Tenant-ID, X-Data-Classification" {
		t.Errorf("Allow-Headers = %q, want Content-Type, Authorization, X-Tenant-ID, X-Data-Classification", v)
	}
	if v := rec.Header().Get("Access-Control-Max-Age"); v != "86400" {
		t.Errorf("Max-Age = %q, want 86400", v)
	}
}

func TestParseCORSOrigins_Default(t *testing.T) {
	t.Setenv(corsEnvKey, "")

	origins := parseCORSOrigins()
	if len(origins) != 1 || origins[0] != corsDefaultOrigin {
		t.Errorf("parseCORSOrigins() = %v, want [%s]", origins, corsDefaultOrigin)
	}
}
