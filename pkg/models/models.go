package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Organization domain model
// Soft delete by DeletedAt.
// CreatedAt/UpdatedAt maintained in service.
type Organization struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  *uuid.UUID `json:"tenantId,omitempty"`
	Name      string     `json:"name"`
	LegalCode *string    `json:"legalCode,omitempty"`
	Status    string     `json:"status"` // active/inactive/blocked
	Tags      []string   `json:"tags,omitempty"`

	Addresses []OrgAddress `json:"addresses,omitempty"`
	Contacts  []OrgContact `json:"contacts,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`

	CreatedBy *uuid.UUID `json:"createdBy,omitempty"`
	UpdatedBy *uuid.UUID `json:"updatedBy,omitempty"`
}

type OrgAddress struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Type           string    `json:"type"` // legal, actual, shipping
	Country        string    `json:"country,omitempty"`
	Region         string    `json:"region,omitempty"`
	City           string    `json:"city,omitempty"`
	Street         string    `json:"street,omitempty"`
	ZIP            string    `json:"zip,omitempty"`
}

type OrgContact struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Type           string    `json:"type"` // email, phone
	Value          string    `json:"value"`
	IsPrimary      bool      `json:"isPrimary"`
}

// Requests / DTOs

type CreateOrganizationRequest struct {
	TenantID  *uuid.UUID   `json:"tenantId"`
	Name      string       `json:"name"` // required
	LegalCode *string      `json:"legalCode"`
	Status    *string      `json:"status"`
	Tags      []string     `json:"tags"`
	Addresses []OrgAddress `json:"addresses"`
	Contacts  []OrgContact `json:"contacts"`
}

type ReplaceOrganizationRequest struct {
	TenantID  *uuid.UUID   `json:"tenantId"`
	Name      string       `json:"name"`
	LegalCode *string      `json:"legalCode"`
	Status    string       `json:"status"`
	Tags      []string     `json:"tags"`
	Addresses []OrgAddress `json:"addresses"`
	Contacts  []OrgContact `json:"contacts"`
}

// Patch with pointers for partial update
// nil pointer means not provided
// empty slice means clear

type PatchOrganizationRequest struct {
	TenantID  **uuid.UUID   `json:"tenantId"`
	Name      *string       `json:"name"`
	LegalCode **string      `json:"legalCode"`
	Status    *string       `json:"status"`
	Tags      *[]string     `json:"tags"`
	Addresses *[]OrgAddress `json:"addresses"`
	Contacts  *[]OrgContact `json:"contacts"`
}

type ListOrganizationsQuery struct {
	TenantID *uuid.UUID
	Q        string // name or legal code contains (ILIKE)
	Status   string
	Tags     []string
	Limit    int
	Offset   int
	SortBy   string // name|createdAt|updatedAt
	SortDir  string // asc|desc
}

type ListOrganizationsResponse struct {
	Items  []Organization `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// Helpers

func NormalizeStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return "active"
	}
	switch s {
	case "active", "inactive", "blocked":
		return s
	default:
		return "active"
	}
}
