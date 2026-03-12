package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestSQLConversationStore_ListTurns(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	caseID := uuid.New()
	turnID := uuid.New()

	turnColumns := []string{"id", "case_id", "role", "content", "metadata", "created_at", "turn_number"}

	tests := []struct {
		name      string
		limit     int
		offset    int
		mock      func(sqlmock.Sqlmock)
		wantCount int
		wantTotal int
		wantErr   bool
	}{
		{
			name:   "happy path with one turn",
			limit:  20,
			offset: 0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, caseID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				meta, _ := json.Marshal(map[string]any{"model": "claude"})
				m.ExpectQuery(`SELECT id, case_id, role, content, metadata, created_at, turn_number`).
					WithArgs(tenantID, caseID, 20, 0).
					WillReturnRows(sqlmock.NewRows(turnColumns).AddRow(
						turnID, caseID, "user", "hello", meta, now, 1,
					))
			},
			wantCount: 1,
			wantTotal: 1,
			wantErr:   false,
		},
		{
			name:   "empty result",
			limit:  20,
			offset: 0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, caseID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				m.ExpectQuery(`SELECT id, case_id, role, content, metadata, created_at, turn_number`).
					WithArgs(tenantID, caseID, 20, 0).
					WillReturnRows(sqlmock.NewRows(turnColumns))
			},
			wantCount: 0,
			wantTotal: 0,
			wantErr:   false,
		},
		{
			name:   "count query error",
			limit:  20,
			offset: 0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, caseID).
					WillReturnError(errors.New("connection lost"))
			},
			wantErr: true,
		},
		{
			name:   "list query error",
			limit:  20,
			offset: 0,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, caseID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectQuery(`SELECT id, case_id, role, content, metadata, created_at, turn_number`).
					WithArgs(tenantID, caseID, 20, 0).
					WillReturnError(errors.New("query timeout"))
			},
			wantErr: true,
		},
		{
			name:   "pagination offset",
			limit:  10,
			offset: 5,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT COUNT`).
					WithArgs(tenantID, caseID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))
				m.ExpectQuery(`SELECT id, case_id, role, content, metadata, created_at, turn_number`).
					WithArgs(tenantID, caseID, 10, 5).
					WillReturnRows(sqlmock.NewRows(turnColumns))
			},
			wantCount: 0,
			wantTotal: 15,
			wantErr:   false,
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
			store := NewSQLConversationStore(db)

			turns, total, err := store.ListTurns(context.Background(), tenantID, caseID, tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListTurns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(turns) != tt.wantCount {
					t.Errorf("ListTurns() returned %d turns, want %d", len(turns), tt.wantCount)
				}
				if total != tt.wantTotal {
					t.Errorf("ListTurns() total = %d, want %d", total, tt.wantTotal)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLConversationStore_ListTurns_MetadataJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	caseID := uuid.New()
	turnID := uuid.New()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	metadata := map[string]any{"model": "claude-4", "tokens": float64(150)}
	metaBytes, _ := json.Marshal(metadata)

	mock.ExpectQuery(`SELECT COUNT`).
		WithArgs(tenantID, caseID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT id, case_id, role, content, metadata, created_at, turn_number`).
		WithArgs(tenantID, caseID, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "case_id", "role", "content", "metadata", "created_at", "turn_number"}).
			AddRow(turnID, caseID, "assistant", "response", metaBytes, now, 1))

	store := NewSQLConversationStore(db)
	turns, _, err := store.ListTurns(context.Background(), tenantID, caseID, 20, 0)
	if err != nil {
		t.Fatalf("ListTurns() unexpected error: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("ListTurns() returned %d turns, want 1", len(turns))
	}

	if turns[0].Metadata["model"] != "claude-4" {
		t.Errorf("metadata model = %v, want claude-4", turns[0].Metadata["model"])
	}
	if turns[0].Metadata["tokens"] != float64(150) {
		t.Errorf("metadata tokens = %v, want 150", turns[0].Metadata["tokens"])
	}
}

func TestSQLConversationStore_InsertTurn(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	tenantID := uuid.New()
	caseID := uuid.New()

	tests := []struct {
		name    string
		role    string
		content string
		meta    map[string]any
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name:    "happy path",
			role:    "user",
			content: "hello world",
			meta:    map[string]any{"source": "web"},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO conversation_turns`).
					WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(now))
				m.ExpectQuery(`SELECT turn_number FROM ordered`).
					WillReturnRows(sqlmock.NewRows([]string{"turn_number"}).AddRow(1))
			},
			wantErr: false,
		},
		{
			name:    "nil metadata",
			role:    "assistant",
			content: "response",
			meta:    nil,
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO conversation_turns`).
					WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(now))
				m.ExpectQuery(`SELECT turn_number FROM ordered`).
					WillReturnRows(sqlmock.NewRows([]string{"turn_number"}).AddRow(2))
			},
			wantErr: false,
		},
		{
			name:    "insert SQL error",
			role:    "user",
			content: "test",
			meta:    map[string]any{},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO conversation_turns`).
					WillReturnError(errors.New("insert failed"))
			},
			wantErr: true,
		},
		{
			name:    "turn number query error",
			role:    "user",
			content: "test",
			meta:    map[string]any{},
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`INSERT INTO conversation_turns`).
					WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(now))
				m.ExpectQuery(`SELECT turn_number FROM ordered`).
					WillReturnError(errors.New("count failed"))
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
			store := NewSQLConversationStore(db)

			got, err := store.InsertTurn(context.Background(), tenantID, caseID, tt.role, tt.content, tt.meta)
			if (err != nil) != tt.wantErr {
				t.Errorf("InsertTurn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Fatal("InsertTurn() returned nil on success")
				}
				if got.Role != tt.role {
					t.Errorf("InsertTurn() role = %q, want %q", got.Role, tt.role)
				}
				if got.Content != tt.content {
					t.Errorf("InsertTurn() content = %q, want %q", got.Content, tt.content)
				}
				if got.CaseID != caseID {
					t.Errorf("InsertTurn() caseID = %v, want %v", got.CaseID, caseID)
				}
				if got.ID == uuid.Nil {
					t.Error("InsertTurn() should generate a non-nil UUID")
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSQLConversationStore_EnsureCaseExists(t *testing.T) {
	tenantID := uuid.New()
	caseID := uuid.New()

	tests := []struct {
		name    string
		mock    func(sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "case exists",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id FROM cases`).
					WithArgs(tenantID, caseID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(caseID))
			},
			wantErr: false,
		},
		{
			name: "case not found",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id FROM cases`).
					WithArgs(tenantID, caseID).
					WillReturnError(errors.New("sql: no rows in result set"))
			},
			wantErr: true,
		},
		{
			name: "SQL error",
			mock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id FROM cases`).
					WithArgs(tenantID, caseID).
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
			store := NewSQLConversationStore(db)

			err = store.EnsureCaseExists(context.Background(), tenantID, caseID)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureCaseExists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
