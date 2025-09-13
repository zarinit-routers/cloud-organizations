package contracts

import (
	"github.com/google/uuid"
	"github.com/zarinit-routers/cloud-organizations/pkg/models"
)

// OrganizationsService describes the behavior required by HTTP handlers.
type OrganizationsService interface {
	Create(req models.CreateOrganizationRequest, userID *uuid.UUID) (*models.Organization, error)
	Get(id uuid.UUID) (*models.Organization, bool)
	List(q models.ListOrganizationsQuery) (items []models.Organization, total int)
	Replace(id uuid.UUID, req models.ReplaceOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool)
	Patch(id uuid.UUID, req models.PatchOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool)
	SoftDelete(id uuid.UUID, userID *uuid.UUID) bool
	Restore(id uuid.UUID, userID *uuid.UUID) (*models.Organization, bool)

	BulkCreate(reqs []models.CreateOrganizationRequest, userID *uuid.UUID) ([]models.Organization, error)
	BulkUpdate(ids []uuid.UUID, patch models.PatchOrganizationRequest, userID *uuid.UUID) (int, error)
	BulkDelete(ids []uuid.UUID, userID *uuid.UUID) int
}
