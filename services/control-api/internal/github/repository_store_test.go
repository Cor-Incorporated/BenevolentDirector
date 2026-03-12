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

func ptr[T any](v T) *T { return &v }

var repoColumns = []string{
	"id", "tenant_id", "installation_id", "github_id",
	"org_name", "repo_name", "full_name",
	"description", "language", "stars", "topics", "tech_stack",
	"total_commits", "contributor_count",
	"is_private", "is_archived",
	"synced_at", "created_at", "updated_at",
}

func TestSQLRepositoryStore_UpsertRepository(t *testing.T) {
	tenantID := uuid.New()
	installID := uuid.New()

	tests := []struct {
		name    string
		repo    *domain.Repository
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "happy path",
			repo: &domain.Repository{
				ID:               uuid.New(),
				TenantID:         tenantID,
				InstallationID:   &installID,
				GitHubID:         ptr[int64](67890),
				OrgName:          ptr("acme"),
				RepoName:         "api",
				FullName:         "acme/api",
				Description:      ptr("API server"),
				Language:         ptr("Go"),
				Stars:            42,
				Topics:           []string{"go", "api"},
				TechStack:        []string{"grpc"},
				TotalCommits:     100,
				ContributorCount: 5,
				IsPrivate:        false,
				IsArchived:       false,
			},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO repository_snapshots`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "generates ID when nil",
			repo: &domain.Repository{
				TenantID: tenantID,
				RepoName: "web",
				FullName: "acme/web",
			},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO repository_snapshots`).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "SQL error",
			repo: &domain.Repository{
				ID:       uuid.New(),
				TenantID: tenantID,
				FullName: "acme/broken",
			},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`INSERT INTO repository_snapshots`).
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
			store := &SQLRepositoryStore{DB: db}

			err = store.UpsertRepository(context.Background(), tt.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsertRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLRepositoryStore_GetByID(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	repoID := uuid.New()
	installID := uuid.New()

	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantNil bool
		wantErr bool
	}{
		{
			name: "found",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM repository_snapshots`).
					WithArgs(repoID, tenantID).
					WillReturnRows(sqlmock.NewRows(repoColumns).AddRow(
						repoID, tenantID, installID, int64(67890),
						"acme", "api", "acme/api",
						"desc", "Go", 10, `{"go","api"}`, `{"grpc"}`,
						100, 5,
						false, false,
						now, now, now,
					))
			},
		},
		{
			name: "not found returns nil",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM repository_snapshots`).
					WithArgs(repoID, tenantID).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
		},
		{
			name: "SQL error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT .+ FROM repository_snapshots`).
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
			store := &SQLRepositoryStore{DB: db}

			got, err := store.GetByID(context.Background(), repoID, tenantID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && got != nil {
				t.Error("GetByID() expected nil for not-found repo")
			}
			if !tt.wantNil && !tt.wantErr && got == nil {
				t.Error("GetByID() returned nil on success")
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLRepositoryStore_ListByTenant(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	repoID := uuid.New()
	installID := uuid.New()
	orgName := "acme"

	tests := []struct {
		name      string
		opts      ListOptions
		mock      func(sqlmock.Sqlmock)
		wantCount int
		wantTotal int
		wantErr   bool
	}{
		{
			name: "no filter",
			opts: ListOptions{Limit: 20, Offset: 0},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectQuery(`SELECT .+ FROM repository_snapshots`).
					WithArgs(tenantID, 20, 0).
					WillReturnRows(sqlmock.NewRows(repoColumns).AddRow(
						repoID, tenantID, installID, int64(67890),
						"acme", "api", "acme/api",
						"desc", "Go", 10, `{}`, `{}`,
						100, 5,
						false, false,
						now, now, now,
					))
			},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "with org filter",
			opts: ListOptions{OrgName: &orgName, Limit: 10, Offset: 0},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, "acme").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectQuery(`SELECT .+ FROM repository_snapshots`).
					WithArgs(tenantID, "acme", 10, 0).
					WillReturnRows(sqlmock.NewRows(repoColumns).AddRow(
						repoID, tenantID, installID, int64(67890),
						"acme", "api", "acme/api",
						"desc", "Go", 10, `{}`, `{}`,
						100, 5,
						false, false,
						now, now, now,
					))
			},
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name: "limit defaults to 20 when invalid",
			opts: ListOptions{Limit: -1, Offset: -5},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				m.ExpectQuery(`SELECT .+ FROM repository_snapshots`).
					WithArgs(tenantID, 20, 0).
					WillReturnRows(sqlmock.NewRows(repoColumns))
			},
			wantCount: 0,
			wantTotal: 0,
		},
		{
			name: "count query error",
			opts: ListOptions{Limit: 20, Offset: 0},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "list query error",
			opts: ListOptions{Limit: 20, Offset: 0},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectQuery(`SELECT .+ FROM repository_snapshots`).
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
			store := &SQLRepositoryStore{DB: db}

			repos, total, err := store.ListByTenant(context.Background(), tenantID, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListByTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(repos) != tt.wantCount {
					t.Errorf("ListByTenant() returned %d repos, want %d", len(repos), tt.wantCount)
				}
				if total != tt.wantTotal {
					t.Errorf("ListByTenant() total = %d, want %d", total, tt.wantTotal)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLRepositoryStore_FindNewAndArchived(t *testing.T) {
	store := &SQLRepositoryStore{DB: nil}
	_, _, err := store.FindNewAndArchived(context.Background(), uuid.New(), []int64{1, 2, 3})
	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("FindNewAndArchived() error = %v, want ErrNotImplemented", err)
	}
}

func TestPqStringArray(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{name: "empty", input: []string{}, want: "{}"},
		{name: "single", input: []string{"go"}, want: `{"go"}`},
		{name: "multiple", input: []string{"go", "api"}, want: `{"go","api"}`},
		{name: "with special chars", input: []string{`a"b`, `c\d`}, want: `{"a\"b","c\\d"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pqStringArray(tt.input)
			if got != tt.want {
				t.Errorf("pqStringArray() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPqStringArrayScanner(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    []string
		wantErr bool
	}{
		{name: "nil", input: nil, want: nil},
		{name: "empty string", input: "", want: nil},
		{name: "empty array", input: "{}", want: nil},
		{name: "single element string", input: `{"go"}`, want: []string{"go"}},
		{name: "multiple elements string", input: `{"go","api","grpc"}`, want: []string{"go", "api", "grpc"}},
		{name: "byte input", input: []byte(`{"rust","wasm"}`), want: []string{"rust", "wasm"}},
		{name: "unsupported type", input: 42, wantErr: true},
		{name: "unquoted elements", input: `{go,api}`, want: []string{"go", "api"}},
		{name: "escaped chars", input: `{"a\"b","c\\d"}`, want: []string{`a"b`, `c\d`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var scanner pqStringArrayScanner
			err := scanner.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got := scanner.Value()
			if tt.want == nil {
				if len(got) != 0 {
					t.Errorf("Value() = %v, want empty", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Value() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("Value()[%d] = %q, want %q", i, got[i], v)
				}
			}
		})
	}
}
