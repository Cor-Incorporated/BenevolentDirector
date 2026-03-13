package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CompletenessStatus represents the collection state of a checklist item.
// Valid values align with the OpenAPI enum: collected, partial, missing.
type CompletenessStatus string

const (
	StatusCollected CompletenessStatus = "collected"
	StatusPartial   CompletenessStatus = "partial"
	StatusMissing   CompletenessStatus = "missing"
)

// CompletenessChecklistItem is the persisted per-topic status payload.
type CompletenessChecklistItem struct {
	Status     CompletenessStatus `json:"status"`
	Confidence float64            `json:"confidence"`
}

// CompletenessObservation is the API-facing completeness snapshot.
type CompletenessObservation struct {
	OverallCompleteness float64                              `json:"overall_completeness"`
	Checklist           map[string]CompletenessChecklistItem `json:"checklist"`
	SuggestedNextTopics []string                             `json:"suggested_next_topics"`
	TurnCount           int                                  `json:"-"`
}

// CompletenessStore provides read access to completeness feedback-loop state.
type CompletenessStore interface {
	// GetByCaseID returns the latest completeness snapshot for a tenant-scoped case.
	// Returns (nil, nil) when no snapshot exists.
	GetByCaseID(ctx context.Context, tenantID, caseID uuid.UUID) (*CompletenessObservation, error)
}

// SQLCompletenessStore reads completeness snapshots from PostgreSQL.
type SQLCompletenessStore struct {
	db *sql.DB
}

// NewSQLCompletenessStore creates a SQLCompletenessStore backed by the given database.
// Panics if db is nil, matching the guard in NewCompletenessHandler.
func NewSQLCompletenessStore(db *sql.DB) *SQLCompletenessStore {
	if db == nil {
		panic("db must not be nil")
	}
	return &SQLCompletenessStore{db: db}
}

// GetByCaseID returns the latest completeness snapshot for a case.
// Note: the caseID parameter corresponds to the "session_id" column in the
// completeness_tracking table. The REST layer exposes {caseId} as a path
// parameter, but the underlying DB schema uses session_id.
//
// The query filters on source_domain = 'estimation' because completeness
// observations are currently only produced by the estimation pipeline.
// If other domains start writing completeness_tracking rows, consider
// adding a source_domain parameter or renaming to GetByEstimationDomain.
func (s *SQLCompletenessStore) GetByCaseID(ctx context.Context, tenantID, caseID uuid.UUID) (*CompletenessObservation, error) {
	const query = `
		SELECT checklist, overall_completeness, suggested_next_topics, turn_count
		FROM completeness_tracking
		WHERE tenant_id = $1 AND session_id = $2 AND source_domain = 'estimation'
		ORDER BY updated_at DESC, created_at DESC
		LIMIT 1
	`

	var (
		observation  CompletenessObservation
		checklistRaw []byte
		topics       []string
	)

	err := executorFromContext(ctx, s.db).QueryRowContext(ctx, query, tenantID, caseID).Scan(
		&checklistRaw,
		&observation.OverallCompleteness,
		pq.Array(&topics),
		&observation.TurnCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get completeness observation: %w", err)
	}

	observation.Checklist = map[string]CompletenessChecklistItem{}
	if len(checklistRaw) > 0 {
		if err := json.Unmarshal(checklistRaw, &observation.Checklist); err != nil {
			return nil, fmt.Errorf("decode completeness checklist: %w", err)
		}
	}
	observation.SuggestedNextTopics = topics

	return &observation, nil
}
