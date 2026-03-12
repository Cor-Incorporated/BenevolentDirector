package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Cor-Incorporated/Grift/services/control-api/internal/domain"
	"github.com/google/uuid"
)

// RequirementArtifactStore provides persistence operations for requirement artifacts.
type RequirementArtifactStore interface {
	// GetLatestByCaseID returns the latest versioned requirement artifact for
	// the given tenant and case. Returns (nil, nil) if no artifact exists.
	GetLatestByCaseID(ctx context.Context, tenantID, caseID uuid.UUID) (*domain.RequirementArtifact, error)
}

// SQLRequirementArtifactStore implements RequirementArtifactStore using *sql.DB.
type SQLRequirementArtifactStore struct {
	DB *sql.DB
}

// NewSQLRequirementArtifactStore creates a new SQLRequirementArtifactStore backed by the given database.
func NewSQLRequirementArtifactStore(db *sql.DB) *SQLRequirementArtifactStore {
	return &SQLRequirementArtifactStore{DB: db}
}

// GetLatestByCaseID returns the most recent requirement artifact for the given
// tenant and case, ordered by version descending. Returns (nil, nil) when no
// matching row exists.
func (s *SQLRequirementArtifactStore) GetLatestByCaseID(ctx context.Context, tenantID, caseID uuid.UUID) (*domain.RequirementArtifact, error) {
	const query = `
		SELECT id, tenant_id, case_id, version, markdown,
			source_chunks, status, created_by_uid,
			created_at, updated_at
		FROM requirement_artifacts
		WHERE tenant_id = $1 AND case_id = $2
		ORDER BY version DESC
		LIMIT 1
	`

	var a domain.RequirementArtifact
	var sourceChunks pqUUIDArrayScanner
	var status string

	err := s.DB.QueryRowContext(ctx, query, tenantID, caseID).Scan(
		&a.ID, &a.TenantID, &a.CaseID, &a.Version, &a.Markdown,
		&sourceChunks, &status, &a.CreatedByUID,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting latest requirement artifact for case %s: %w", caseID, err)
	}

	a.SourceChunks = sourceChunks.Value()
	a.Status = domain.ArtifactStatus(status)

	return &a, nil
}

// pqUUIDArrayScanner scans a PostgreSQL uuid[] column into a Go uuid.UUID slice.
type pqUUIDArrayScanner struct {
	data []uuid.UUID
}

// Scan implements the sql.Scanner interface for PostgreSQL uuid[] columns.
func (s *pqUUIDArrayScanner) Scan(src any) error {
	if src == nil {
		s.data = nil
		return nil
	}

	var raw string
	switch v := src.(type) {
	case []byte:
		raw = string(v)
	case string:
		raw = v
	default:
		return fmt.Errorf("unsupported type for uuid[]: %T", src)
	}

	if raw == "{}" || raw == "" {
		s.data = nil
		return nil
	}

	// Trim braces: {uuid1,uuid2,...}
	if len(raw) >= 2 && raw[0] == '{' && raw[len(raw)-1] == '}' {
		raw = raw[1 : len(raw)-1]
	}

	var result []uuid.UUID
	start := 0
	for i := 0; i <= len(raw); i++ {
		if i == len(raw) || raw[i] == ',' {
			elem := raw[start:i]
			id, err := uuid.Parse(elem)
			if err != nil {
				return fmt.Errorf("parsing uuid array element %q: %w", elem, err)
			}
			result = append(result, id)
			start = i + 1
		}
	}

	s.data = result
	return nil
}

// Value returns the scanned UUID slice. Returns an empty slice if nil.
func (s *pqUUIDArrayScanner) Value() []uuid.UUID {
	if s.data == nil {
		return []uuid.UUID{}
	}
	return s.data
}
