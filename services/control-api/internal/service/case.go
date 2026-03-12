package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Cor-Incorporated/Grift/services/control-api/internal/domain"
	"github.com/Cor-Incorporated/Grift/services/control-api/internal/store"
	"github.com/google/uuid"
)

// CaseService provides business logic for case operations.
type CaseService struct {
	store store.CaseStore
}

// NewCaseService creates a CaseService with the given store dependency.
func NewCaseService(s store.CaseStore) *CaseService {
	return &CaseService{store: s}
}

// CreateInput holds the parameters for creating a new case.
type CreateInput struct {
	Title             string
	CaseType          domain.CaseType
	ExistingSystemURL *string
	CompanyName       *string
	ContactName       *string
	ContactEmail      *string
	CreatedByUID      *string
}

// Create validates input and creates a new case in draft status.
func (s *CaseService) Create(ctx context.Context, tenantID uuid.UUID, in CreateInput) (*domain.Case, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(title) > 200 {
		return nil, fmt.Errorf("title must be 200 characters or fewer")
	}
	if !in.CaseType.IsValid() {
		return nil, fmt.Errorf("invalid case type: %s", in.CaseType)
	}

	c := &domain.Case{
		ID:                uuid.New(),
		TenantID:          tenantID,
		Title:             title,
		Type:              in.CaseType,
		Status:            domain.CaseStatusDraft,
		ExistingSystemURL: in.ExistingSystemURL,
		CompanyName:       in.CompanyName,
		ContactName:       in.ContactName,
		ContactEmail:      in.ContactEmail,
		CreatedByUID:      in.CreatedByUID,
	}

	result, err := s.store.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("creating case: %w", err)
	}

	return result, nil
}

// List returns cases for a tenant with optional filters and pagination.
func (s *CaseService) List(ctx context.Context, tenantID uuid.UUID, statusFilter, typeFilter string, limit, offset int) ([]domain.Case, int, error) {
	cases, total, err := s.store.List(ctx, tenantID, statusFilter, typeFilter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing cases: %w", err)
	}

	return cases, total, nil
}

// Get returns a single case by ID, scoped to a tenant.
// Returns nil if the case does not exist.
func (s *CaseService) Get(ctx context.Context, tenantID, caseID uuid.UUID) (*domain.Case, error) {
	c, err := s.store.Get(ctx, tenantID, caseID)
	if err != nil {
		return nil, fmt.Errorf("getting case: %w", err)
	}

	return c, nil
}

// UpdateInput holds the optional fields for patching a case.
type UpdateInput struct {
	Title    *string
	Type     *string
	Status   *string
	Priority *string
}

// ErrNotFound indicates the requested resource does not exist.
var ErrNotFound = errors.New("not found")

// Update validates the non-nil fields and patches the case.
func (s *CaseService) Update(ctx context.Context, tenantID, caseID uuid.UUID, in UpdateInput) (*domain.Case, error) {
	var fields store.UpdateCaseFields

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, fmt.Errorf("title must not be empty")
		}
		if len(title) > 200 {
			return nil, fmt.Errorf("title must be 200 characters or fewer")
		}
		fields.Title = &title
	}
	if in.Type != nil {
		ct := domain.CaseType(*in.Type)
		if !ct.IsValid() {
			return nil, fmt.Errorf("invalid case type: %s", *in.Type)
		}
		fields.Type = &ct
	}
	if in.Status != nil {
		cs := domain.CaseStatus(*in.Status)
		if !cs.IsValid() {
			return nil, fmt.Errorf("invalid case status: %s", *in.Status)
		}
		fields.Status = &cs
	}
	if in.Priority != nil {
		cp := domain.CasePriority(*in.Priority)
		if !cp.IsValid() {
			return nil, fmt.Errorf("invalid case priority: %s", *in.Priority)
		}
		fields.Priority = &cp
	}

	result, err := s.store.Update(ctx, tenantID, caseID, fields)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("updating case: %w", err)
	}

	return result, nil
}

// Delete removes a case by ID scoped to a tenant.
// Returns ErrNotFound if the case does not exist.
func (s *CaseService) Delete(ctx context.Context, tenantID, caseID uuid.UUID) error {
	err := s.store.Delete(ctx, tenantID, caseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("deleting case: %w", err)
	}

	return nil
}
