package github

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

func ptrFloat64(v float64) *float64 { return &v }
func ptrInt(v int) *int             { return &v }

var velocityColumns = []string{
	"id", "tenant_id", "repository_id",
	"commits_per_week", "active_days_per_week", "pr_merge_frequency",
	"issue_close_speed", "churn_rate", "contributor_count",
	"velocity_score", "estimated_hours",
	"analyzed_at", "created_at",
}

func TestNewSQLVelocityStore(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	store := NewSQLVelocityStore(db)
	if store == nil {
		t.Fatal("NewSQLVelocityStore returned nil")
	}
	if store.DB != db {
		t.Error("NewSQLVelocityStore did not set DB field")
	}
}

func TestSQLVelocityStore_Insert(t *testing.T) {
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name    string
		metric  *domain.VelocityMetric
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "happy path",
			metric: &domain.VelocityMetric{
				ID:                uuid.New(),
				TenantID:          uuid.New(),
				RepositoryID:      uuid.New(),
				CommitsPerWeek:    ptrFloat64(12.5),
				ActiveDaysPerWeek: ptrFloat64(4.0),
				PRMergeFrequency:  ptrFloat64(3.2),
				IssueCloseSpeed:   ptrFloat64(2.1),
				ChurnRate:         ptrFloat64(0.15),
				ContributorCount:  ptrInt(8),
				VelocityScore:     ptrFloat64(75.0),
				EstimatedHours:    ptrFloat64(120.0),
				AnalyzedAt:        now,
				CreatedAt:         now,
			},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO velocity_metrics`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:    "nil metric",
			metric:  nil,
			mock:    func(m sqlmock.Sqlmock) {},
			wantErr: true,
		},
		{
			name: "SQL error",
			metric: &domain.VelocityMetric{
				ID:           uuid.New(),
				TenantID:     uuid.New(),
				RepositoryID: uuid.New(),
				AnalyzedAt:   now,
				CreatedAt:    now,
			},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO velocity_metrics`).
					WillReturnError(errors.New("constraint violation"))
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
			store := &SQLVelocityStore{DB: db}

			err = store.Insert(context.Background(), tt.metric)
			if (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLVelocityStore_LatestByRepositoryAndTenant(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	repoID := uuid.New()
	tenantID := uuid.New()
	metricID := uuid.New()

	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantNil bool
		wantErr bool
	}{
		{
			name: "found",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, tenantID).
					WillReturnRows(sqlmock.NewRows(velocityColumns).AddRow(
						metricID, tenantID, repoID,
						12.5, 4.0, 3.2,
						2.1, 0.15, 8,
						75.0, 120.0,
						now, now,
					))
			},
		},
		{
			name: "not found",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, tenantID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "SQL error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, tenantID).
					WillReturnError(errors.New("connection refused"))
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
			store := &SQLVelocityStore{DB: db}

			got, err := store.LatestByRepositoryAndTenant(context.Background(), repoID, tenantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LatestByRepositoryAndTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("LatestByRepositoryAndTenant() returned nil on success")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLVelocityStore_ListByRepository(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	repoID := uuid.New()
	tenantID := uuid.New()
	metricID := uuid.New()

	tests := []struct {
		name      string
		limit     int
		mock      func(sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name:  "happy path",
			limit: 5,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, 5).
					WillReturnRows(sqlmock.NewRows(velocityColumns).
						AddRow(metricID, tenantID, repoID, 12.5, 4.0, 3.2, 2.1, 0.15, 8, 75.0, 120.0, now, now).
						AddRow(uuid.New(), tenantID, repoID, 10.0, 3.0, 2.5, 1.8, 0.12, 7, 68.0, 100.0, now, now),
					)
			},
			wantCount: 2,
		},
		{
			name:  "negative limit defaults to 10",
			limit: -1,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, 10).
					WillReturnRows(sqlmock.NewRows(velocityColumns))
			},
			wantCount: 0,
		},
		{
			name:  "zero limit defaults to 10",
			limit: 0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, 10).
					WillReturnRows(sqlmock.NewRows(velocityColumns))
			},
			wantCount: 0,
		},
		{
			name:  "query error",
			limit: 5,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, 5).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:  "scan error",
			limit: 5,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, 5).
					WillReturnRows(sqlmock.NewRows(velocityColumns).
						AddRow("not-a-uuid", tenantID, repoID, 12.5, 4.0, 3.2, 2.1, 0.15, 8, 75.0, 120.0, now, now),
					)
			},
			wantErr: true,
		},
		{
			name:  "row iteration error",
			limit: 5,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM velocity_metrics`).
					WithArgs(repoID, 5).
					WillReturnRows(sqlmock.NewRows(velocityColumns).
						AddRow(metricID, tenantID, repoID, 12.5, 4.0, 3.2, 2.1, 0.15, 8, 75.0, 120.0, now, now).
						RowError(0, errors.New("row iteration failure")),
					)
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
			store := &SQLVelocityStore{DB: db}

			got, err := store.ListByRepository(context.Background(), repoID, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListByRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.wantCount {
				t.Errorf("ListByRepository() returned %d metrics, want %d", len(got), tt.wantCount)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
