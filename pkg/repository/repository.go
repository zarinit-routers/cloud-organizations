package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/zarinit-routers/cloud-organizations/pkg/models"
)

// Organizations abstracts DB ops. Concrete impl will be generated/wrapped around sqlc/pgx.
// Keeping interface small enables service tests with mocks.
type Organizations interface {
	Create(ctx context.Context, req models.CreateOrganizationRequest, userID *uuid.UUID) (*models.Organization, error)
	Get(ctx context.Context, id uuid.UUID) (*models.Organization, bool, error)
	List(ctx context.Context, q models.ListOrganizationsQuery) ([]models.Organization, int, error)
	Replace(ctx context.Context, id uuid.UUID, req models.ReplaceOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool, error)
	Patch(ctx context.Context, id uuid.UUID, req models.PatchOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (bool, error)
	Restore(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.Organization, bool, error)

	BulkCreate(ctx context.Context, reqs []models.CreateOrganizationRequest, userID *uuid.UUID) ([]models.Organization, error)
	BulkUpdate(ctx context.Context, ids []uuid.UUID, patch models.PatchOrganizationRequest, userID *uuid.UUID) (int, error)
	BulkDelete(ctx context.Context, ids []uuid.UUID, userID *uuid.UUID) (int, error)

	AddMember(ctx context.Context, orgID, userID uuid.UUID) error
	RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error
	ListMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error)
}
