package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Cor-Incorporated/Grift/services/control-api/internal/domain"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestSQLCaseStore_Create(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	caseID := uuid.New()

	tests := []struct {
		name    string
		input   *domain.Case
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "happy path",
			input: &domain.Case{
				ID:       caseID,
				TenantID: tenantID,
				Title:    "Test Case",
				Type:     domain.CaseTypeNewProject,
				Status:   domain.CaseStatusDraft,
			},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO cases`).
					WithArgs(caseID, tenantID, "Test Case", domain.CaseTypeNewProject, domain.CaseStatusDraft,
						nil, nil, nil, nil, nil).
					WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at"}).AddRow(now, now))
			},
			wantErr: false,
		},
		{
			name: "SQL error on insert",
			input: &domain.Case{
				ID:       caseID,
				TenantID: tenantID,
				Title:    "Test Case",
				Type:     domain.CaseTypeNewProject,
				Status:   domain.CaseStatusDraft,
			},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO cases`).
					WithArgs(caseID, tenantID, "Test Case", domain.CaseTypeNewProject, domain.CaseStatusDraft,
						nil, nil, nil, nil, nil).
					WillReturnError(errors.New("unique constraint violation"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mock(mock)
			store := NewSQLCaseStore(db)

			got, err := store.Create(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("Create() returned nil on success")
			}
			if !tt.wantErr {
				if got.CreatedAt.IsZero() {
					t.Error("Create() CreatedAt should be populated")
				}
				if got.UpdatedAt.IsZero() {
					t.Error("Create() UpdatedAt should be populated")
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLCaseStore_Get(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	caseID := uuid.New()

	caseColumns := []string{
		"id", "tenant_id", "title", "type", "status", "priority", "business_line",
		"existing_system_url", "spec_markdown", "contact_name", "contact_email",
		"company_name", "created_by_uid", "created_at", "updated_at",
	}

	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantNil bool
		wantErr bool
	}{
		{
			name: "found",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM cases`).
					WithArgs(tenantID, caseID).
					WillReturnRows(sqlmock.NewRows(caseColumns).AddRow(
						caseID, tenantID, "My Case", domain.CaseTypeNewProject, domain.CaseStatusDraft,
						nil, nil, nil, nil, nil, nil, nil, nil, now, now,
					))
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "not found returns nil",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM cases`).
					WithArgs(tenantID, caseID).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "SQL error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM cases`).
					WithArgs(tenantID, caseID).
					WillReturnError(errors.New("connection refused"))
			},
			wantNil: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mock(mock)
			store := NewSQLCaseStore(db)

			got, err := store.Get(context.Background(), tenantID, caseID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && got != nil {
				t.Error("Get() expected nil for not-found case")
			}
			if !tt.wantNil && !tt.wantErr && got == nil {
				t.Error("Get() returned nil on success")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLCaseStore_List(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	caseID := uuid.New()

	caseColumns := []string{
		"id", "tenant_id", "title", "type", "status", "priority", "business_line",
		"existing_system_url", "spec_markdown", "contact_name", "contact_email",
		"company_name", "created_by_uid", "created_at", "updated_at",
	}

	tests := []struct {
		name         string
		statusFilter string
		typeFilter   string
		limit        int
		offset       int
		mock         func(sqlmock.Sqlmock)
		wantCount    int
		wantTotal    int
		wantErr      bool
	}{
		{
			name:         "no filters",
			statusFilter: "",
			typeFilter:   "",
			limit:        20,
			offset:       0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectQuery(`SELECT .+ FROM cases`).
					WithArgs(tenantID, 20, 0).
					WillReturnRows(sqlmock.NewRows(caseColumns).AddRow(
						caseID, tenantID, "Case 1", domain.CaseTypeNewProject, domain.CaseStatusDraft,
						nil, nil, nil, nil, nil, nil, nil, nil, now, now,
					))
			},
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:         "with status filter",
			statusFilter: "draft",
			typeFilter:   "",
			limit:        10,
			offset:       0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, "draft").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
				m.ExpectQuery(`SELECT .+ FROM cases`).
					WithArgs(tenantID, "draft", 10, 0).
					WillReturnRows(sqlmock.NewRows(caseColumns).
						AddRow(caseID, tenantID, "Case 1", domain.CaseTypeNewProject, domain.CaseStatusDraft,
							nil, nil, nil, nil, nil, nil, nil, nil, now, now).
						AddRow(uuid.New(), tenantID, "Case 2", domain.CaseTypeBugReport, domain.CaseStatusDraft,
							nil, nil, nil, nil, nil, nil, nil, nil, now, now),
					)
			},
			wantCount: 2,
			wantTotal: 2,
			wantErr:   false,
		},
		{
			name:         "with both filters",
			statusFilter: "draft",
			typeFilter:   "new_project",
			limit:        10,
			offset:       0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, "draft", "new_project").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectQuery(`SELECT .+ FROM cases`).
					WithArgs(tenantID, "draft", "new_project", 10, 0).
					WillReturnRows(sqlmock.NewRows(caseColumns).AddRow(
						caseID, tenantID, "Case 1", domain.CaseTypeNewProject, domain.CaseStatusDraft,
						nil, nil, nil, nil, nil, nil, nil, nil, now, now,
					))
			},
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:         "count query error",
			statusFilter: "",
			typeFilter:   "",
			limit:        20,
			offset:       0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:         "list query error",
			statusFilter: "",
			typeFilter:   "",
			limit:        20,
			offset:       0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectQuery(`SELECT .+ FROM cases`).
					WithArgs(tenantID, 20, 0).
					WillReturnError(errors.New("query failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sqlmock: %v", err)
			}
			defer db.Close()

			tt.mock(mock)
			store := NewSQLCaseStore(db)

			cases, total, err := store.List(context.Background(), tenantID, tt.statusFilter, tt.typeFilter, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(cases) != tt.wantCount {
					t.Errorf("List() returned %d cases, want %d", len(cases), tt.wantCount)
				}
				if total != tt.wantTotal {
					t.Errorf("List() total = %d, want %d", total, tt.wantTotal)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
