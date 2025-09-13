package organizations

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/zarinit-routers/cloud-organizations/pkg/models"
	"github.com/zarinit-routers/cloud-organizations/pkg/repository"
)

// Publisher is a minimal interface for event publishing.
type Publisher interface {
	OrganizationCreated(ctx context.Context, org models.Organization, traceID string) error
	OrganizationUpdated(ctx context.Context, org models.Organization, traceID string) error
	OrganizationDeleted(ctx context.Context, org models.Organization, traceID string) error
}

// Service keeps organizations in memory (can optionally be backed by a repository).
type Service struct {
	mu    sync.RWMutex
	data  map[uuid.UUID]models.Organization
	pub   Publisher
	nowFn func() time.Time
	repo  repository.Organizations
}

func NewService(pub Publisher) *Service {
	return &Service{
		data:  make(map[uuid.UUID]models.Organization),
		pub:   pub,
		nowFn: time.Now,
	}
}

// WithRepository enables DB-backed mode.
func (s *Service) WithRepository(r repository.Organizations) *Service { s.repo = r; return s }

func (s *Service) Create(req models.CreateOrganizationRequest, userID *uuid.UUID) (*models.Organization, error) {
	if s.repo != nil {
		org, err := s.repo.Create(context.Background(), req, userID)
		if err != nil {
			return nil, err
		}
		_ = s.publishCreated(*org)
		return org, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFn()

	id := uuid.New()
	status := "active"
	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		status = models.NormalizeStatus(*req.Status)
	}
	org := models.Organization{
		ID:        id,
		TenantID:  req.TenantID,
		Name:      strings.TrimSpace(req.Name),
		LegalCode: req.LegalCode,
		Status:    status,
		Tags:      append([]string{}, req.Tags...),
		Addresses: cloneAddresses(req.Addresses, id),
		Contacts:  cloneContacts(req.Contacts, id),
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: userID,
		UpdatedBy: userID,
	}
	s.data[id] = org

	_ = s.publishCreated(org)
	return &org, nil
}

func (s *Service) Get(id uuid.UUID) (*models.Organization, bool) {
	if s.repo != nil {
		org, ok, _ := s.repo.Get(context.Background(), id)
		if !ok || org == nil {
			return nil, false
		}
		return org, true
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	org, ok := s.data[id]
	if !ok || org.DeletedAt != nil {
		return nil, false
	}
	return &org, true
}

func (s *Service) List(q models.ListOrganizationsQuery) ([]models.Organization, int) {
	if s.repo != nil {
		items, total, _ := s.repo.List(context.Background(), q)
		return items, total
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var items []models.Organization
	for _, org := range s.data {
		if org.DeletedAt != nil {
			continue
		}
		if q.TenantID != nil {
			if org.TenantID == nil || org.TenantID.String() != q.TenantID.String() {
				continue
			}
		}
		if q.Status != "" && org.Status != q.Status {
			continue
		}
		if q.Q != "" {
			v := strings.ToLower(q.Q)
			lc := ""
			if org.LegalCode != nil {
				lc = strings.ToLower(*org.LegalCode)
			}
			if !strings.Contains(strings.ToLower(org.Name), v) && !strings.Contains(lc, v) {
				continue
			}
		}
		if len(q.Tags) > 0 && !containsAll(org.Tags, q.Tags) {
			continue
		}
		items = append(items, org)
	}
	// Sort
	sort.Slice(items, func(i, j int) bool {
		field := strings.ToLower(q.SortBy)
		dir := strings.ToLower(q.SortDir)
		if field == "" {
			field = "createdat"
		}
		if dir == "" {
			dir = "desc"
		}
		less := false
		switch field {
		case "name":
			less = strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		case "updatedat":
			less = items[i].UpdatedAt.Before(items[j].UpdatedAt)
		default: // createdAt
			less = items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		if dir == "desc" {
			return !less
		}
		return less
	})
	total := len(items)
	// Pagination
	start := q.Offset
	if start > total {
		start = total
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total
}

func (s *Service) Replace(id uuid.UUID, r models.ReplaceOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool) {
	if s.repo != nil {
		org, ok, _ := s.repo.Replace(context.Background(), id, r, userID)
		if !ok || org == nil {
			return nil, false
		}
		_ = s.publishUpdated(*org)
		return org, true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	org, ok := s.data[id]
	if !ok || org.DeletedAt != nil {
		return nil, false
	}
	now := s.nowFn()
	org.TenantID = r.TenantID
	org.Name = strings.TrimSpace(r.Name)
	org.LegalCode = r.LegalCode
	org.Status = models.NormalizeStatus(r.Status)
	org.Tags = append([]string{}, r.Tags...)
	org.Addresses = cloneAddresses(r.Addresses, id)
	org.Contacts = cloneContacts(r.Contacts, id)
	org.UpdatedAt = now
	org.UpdatedBy = userID
	s.data[id] = org
	_ = s.publishUpdated(org)
	return &org, true

}

func (s *Service) Patch(id uuid.UUID, r models.PatchOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool) {
	if s.repo != nil {
		org, ok, _ := s.repo.Patch(context.Background(), id, r, userID)
		if !ok || org == nil {
			return nil, false
		}
		_ = s.publishUpdated(*org)
		return org, true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	org, ok := s.data[id]
	if !ok || org.DeletedAt != nil {
		return nil, false
	}
	now := s.nowFn()
	if r.TenantID != nil {
		org.TenantID = *r.TenantID
	}
	if r.Name != nil {
		org.Name = strings.TrimSpace(*r.Name)
	}
	if r.LegalCode != nil {
		org.LegalCode = *r.LegalCode
	}
	if r.Status != nil {
		org.Status = models.NormalizeStatus(*r.Status)
	}
	if r.Tags != nil {
		org.Tags = append([]string{}, (*r.Tags)...)
	}
	if r.Addresses != nil {
		org.Addresses = cloneAddresses(*r.Addresses, id)
	}
	if r.Contacts != nil {
		org.Contacts = cloneContacts(*r.Contacts, id)
	}
	org.UpdatedAt = now
	org.UpdatedBy = userID
	s.data[id] = org
	_ = s.publishUpdated(org)
	return &org, true
}

func (s *Service) SoftDelete(id uuid.UUID, userID *uuid.UUID) bool {
	if s.repo != nil {
		ok, _ := s.repo.SoftDelete(context.Background(), id, userID)
		return ok
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	org, ok := s.data[id]
	if !ok || org.DeletedAt != nil {
		return false
	}
	now := s.nowFn()
	org.DeletedAt = &now
	org.UpdatedAt = now
	org.UpdatedBy = userID
	s.data[id] = org
	_ = s.publishDeleted(org)
	return true
}

func (s *Service) Restore(id uuid.UUID, userID *uuid.UUID) (*models.Organization, bool) {
	if s.repo != nil {
		org, ok, _ := s.repo.Restore(context.Background(), id, userID)
		if !ok || org == nil {
			return nil, false
		}
		_ = s.publishUpdated(*org)
		return org, true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	org, ok := s.data[id]
	if !ok || org.DeletedAt == nil {
		return nil, false
	}
	org.DeletedAt = nil
	org.UpdatedAt = s.nowFn()
	org.UpdatedBy = userID
	s.data[id] = org
	_ = s.publishUpdated(org)
	return &org, true
}

func (s *Service) BulkCreate(reqs []models.CreateOrganizationRequest, userID *uuid.UUID) ([]models.Organization, error) {
	if s.repo != nil {
		return s.repo.BulkCreate(context.Background(), reqs, userID)
	}
	res := make([]models.Organization, 0, len(reqs))
	for _, r := range reqs {
		org, err := s.Create(r, userID)
		if err != nil {
			return nil, err
		}
		res = append(res, *org)
	}
	return res, nil
}

func (s *Service) BulkUpdate(ids []uuid.UUID, patch models.PatchOrganizationRequest, userID *uuid.UUID) (int, error) {
	if s.repo != nil {
		return s.repo.BulkUpdate(context.Background(), ids, patch, userID)
	}
	count := 0
	for _, id := range ids {
		if _, ok := s.Patch(id, patch, userID); ok {
			count++
		}
	}
	return count, nil
}

func (s *Service) BulkDelete(ids []uuid.UUID, userID *uuid.UUID) int {
	if s.repo != nil {
		count, _ := s.repo.BulkDelete(context.Background(), ids, userID)
		return count
	}
	count := 0
	for _, id := range ids {
		if s.SoftDelete(id, userID) {
			count++
		}
	}
	return count
}

// helpers

func containsAll(have, need []string) bool {
	h := map[string]struct{}{}
	for _, v := range have {
		h[strings.ToLower(v)] = struct{}{}
	}
	for _, v := range need {
		if _, ok := h[strings.ToLower(v)]; !ok {
			return false
		}
	}
	return true
}

func cloneAddresses(in []models.OrgAddress, orgID uuid.UUID) []models.OrgAddress {
	out := make([]models.OrgAddress, len(in))
	for i, a := range in {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		a.OrganizationID = orgID
		out[i] = a
	}
	return out
}

func cloneContacts(in []models.OrgContact, orgID uuid.UUID) []models.OrgContact {
	out := make([]models.OrgContact, len(in))
	for i, a := range in {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		a.OrganizationID = orgID
		out[i] = a
	}
	return out
}

// event helpers

func (s *Service) publishCreated(org models.Organization) error {
	if s.pub == nil {
		return nil
	}
	return s.pub.OrganizationCreated(context.Background(), org, "")
}
func (s *Service) publishUpdated(org models.Organization) error {
	if s.pub == nil {
		return nil
	}
	return s.pub.OrganizationUpdated(context.Background(), org, "")
}
func (s *Service) publishDeleted(org models.Organization) error {
	if s.pub == nil {
		return nil
	}
	return s.pub.OrganizationDeleted(context.Background(), org, "")
}
