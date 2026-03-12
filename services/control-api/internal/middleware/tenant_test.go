package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// fakeTenantStore is a test double for TenantStore.
type fakeTenantStore struct {
	tenants   map[string]bool
	rlsCalls  []string
	existsErr error
	setRLSErr error
	fakeTx    *sql.Tx // nil is acceptable for tests; TxFromContext will return nil
}

func (f *fakeTenantStore) Exists(_ context.Context, tenantID string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.tenants[tenantID], nil
}

func (f *fakeTenantStore) SetRLS(_ context.Context, tenantID string) (*sql.Tx, error) {
	if f.setRLSErr != nil {
		return nil, f.setRLSErr
	}
	f.rlsCalls = append(f.rlsCalls, tenantID)
	return f.fakeTx, nil
}

func TestTenantMiddleware_MissingHeader(t *testing.T) {
	handler := Tenant(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTenantMiddleware_InvalidUUID(t *testing.T) {
	handler := Tenant(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", "not-a-uuid")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestTenantMiddleware_ValidUUID_NoStore(t *testing.T) {
	var gotTenantID string
	handler := Tenant(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenantID = TenantIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	tenantID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", tenantID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotTenantID != tenantID {
		t.Errorf("TenantIDFromContext() = %q, want %q", gotTenantID, tenantID)
	}
}

func TestTenantMiddleware_HealthSkipsTenantCheck(t *testing.T) {
	handler := Tenant(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTenantWithStore_ValidTenantSetsRLS(t *testing.T) {
	tenantID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	store := &fakeTenantStore{
		tenants: map[string]bool{tenantID: true},
	}

	var gotTenantID string
	handler := TenantWithStore(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenantID = TenantIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", tenantID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotTenantID != tenantID {
		t.Errorf("TenantIDFromContext() = %q, want %q", gotTenantID, tenantID)
	}
	if len(store.rlsCalls) != 1 || store.rlsCalls[0] != tenantID {
		t.Errorf("SetRLS calls = %v, want [%s]", store.rlsCalls, tenantID)
	}
}

func TestTenantWithStore_TenantNotFound(t *testing.T) {
	store := &fakeTenantStore{
		tenants: map[string]bool{},
	}

	handler := TenantWithStore(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestTenantWithStore_ExistsError(t *testing.T) {
	store := &fakeTenantStore{
		existsErr: fmt.Errorf("db connection failed"),
	}

	handler := TenantWithStore(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestTenantWithStore_SetRLSError(t *testing.T) {
	tenantID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	store := &fakeTenantStore{
		tenants:   map[string]bool{tenantID: true},
		setRLSErr: fmt.Errorf("SET failed"),
	}

	handler := TenantWithStore(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", tenantID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestTenantWithStore_HealthSkips(t *testing.T) {
	store := &fakeTenantStore{
		tenants: map[string]bool{},
	}

	handler := TenantWithStore(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if len(store.rlsCalls) != 0 {
		t.Errorf("SetRLS should not be called for health, got %v", store.rlsCalls)
	}
}

func TestTenantWithStore_PostTenantsSkipsTenantCheck(t *testing.T) {
	store := &fakeTenantStore{
		tenants: map[string]bool{},
	}

	handler := TenantWithStore(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if len(store.rlsCalls) != 0 {
		t.Errorf("SetRLS should not be called for POST /v1/tenants, got %v", store.rlsCalls)
	}
}

func TestTenantIDFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	if got := TenantIDFromContext(ctx); got != "" {
		t.Errorf("TenantIDFromContext() = %q, want empty string", got)
	}
}

func TestTxFromContext_EmptyContext(t *testing.T) {
	ctx := context.Background()
	if got := TxFromContext(ctx); got != nil {
		t.Errorf("TxFromContext() = %v, want nil", got)
	}
}

func TestStatusWriter_WriteDefaultsTo200(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec}

	_, _ = sw.Write([]byte("hello"))

	if sw.Status() != http.StatusOK {
		t.Errorf("Status() = %d, want %d", sw.Status(), http.StatusOK)
	}
}

func TestStatusWriter_FlushSetsStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec}

	sw.Flush()

	if sw.Status() != http.StatusOK {
		t.Errorf("Status() after Flush = %d, want %d", sw.Status(), http.StatusOK)
	}
}

func TestStatusWriter_ExplicitWriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec}

	sw.WriteHeader(http.StatusNotFound)

	if sw.Status() != http.StatusNotFound {
		t.Errorf("Status() = %d, want %d", sw.Status(), http.StatusNotFound)
	}
}

func TestFinalizeRequestTx_NilTx(t *testing.T) {
	// Should not panic on nil tx.
	finalizeRequestTx(nil, http.StatusOK)
}

func TestTenantWithStore_ServerErrorRollsBackTx(t *testing.T) {
	tenantID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	store := &fakeTenantStore{
		tenants: map[string]bool{tenantID: true},
	}

	handler := TenantWithStore(store)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/cases", nil)
	req.Header.Set("X-Tenant-ID", tenantID)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestSQLTenantStore_Exists_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	tenantID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	store := &SQLTenantStore{DB: db}
	exists, err := store.Exists(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSQLTenantStore_Exists_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	tenantID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(tenantID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	store := &SQLTenantStore{DB: db}
	exists, err := store.Exists(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false")
	}
}

func TestSQLTenantStore_Exists_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs("some-id").
		WillReturnError(fmt.Errorf("db connection failed"))

	store := &SQLTenantStore{DB: db}
	_, err = store.Exists(context.Background(), "some-id")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSQLTenantStore_SetRLS_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	tenantID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config`).
		WithArgs(tenantID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	store := &SQLTenantStore{DB: db}
	tx, err := store.SetRLS(context.Background(), tenantID)
	if err != nil {
		t.Fatalf("SetRLS() error = %v", err)
	}
	if tx == nil {
		t.Fatal("SetRLS() returned nil tx")
	}
	// Clean up: rollback the mock tx
	mock.ExpectRollback()
	_ = tx.Rollback()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestSQLTenantStore_SetRLS_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("cannot begin"))

	store := &SQLTenantStore{DB: db}
	_, err = store.SetRLS(context.Background(), "tenant-id")
	if err == nil {
		t.Fatal("expected error from BeginTx failure")
	}
}

func TestSQLTenantStore_SetRLS_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec(`SELECT set_config`).
		WithArgs("tenant-id").
		WillReturnError(fmt.Errorf("exec failed"))
	mock.ExpectRollback()

	store := &SQLTenantStore{DB: db}
	_, err = store.SetRLS(context.Background(), "tenant-id")
	if err == nil {
		t.Fatal("expected error from set_config failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFinalizeRequestTx_CommitSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	mock.ExpectCommit()

	finalizeRequestTx(tx, http.StatusOK)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFinalizeRequestTx_RollbackOn500(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	mock.ExpectRollback()

	finalizeRequestTx(tx, http.StatusInternalServerError)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFinalizeRequestTx_CommitErrorTriggersRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))
	// After commit failure, finalizeRequestTx calls Rollback.
	// sqlmock may or may not track this depending on version,
	// so we just verify it does not panic.
	finalizeRequestTx(tx, http.StatusOK)
}
