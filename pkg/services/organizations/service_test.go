package organizations_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/zarinit-routers/cloud-organizations/pkg/models"
	"github.com/zarinit-routers/cloud-organizations/pkg/services/organizations"
)

func TestServiceCreate(t *testing.T) {
	svc := organizations.NewService(nil)
	userID := uuid.New()
	req := models.CreateOrganizationRequest{
		Name: "Test Org",
		Tags: []string{"test"},
	}
	org, err := svc.Create(req, &userID)
	if err != nil {
		t.Fatal(err)
	}
	if org.Name != "Test Org" {
		t.Errorf("name mismatch: %s", org.Name)
	}
	if len(org.Tags) != 1 || org.Tags[0] != "test" {
		t.Errorf("tags mismatch: %v", org.Tags)
	}

	fetched, ok := svc.Get(org.ID)
	if !ok {
		t.Fatal("not found")
	}
	if fetched.ID != org.ID {
		t.Error("id mismatch")
	}
}

func TestServiceList(t *testing.T) {
	svc := organizations.NewService(nil)
	userID := uuid.New()
	for i := 0; i < 5; i++ {
		req := models.CreateOrganizationRequest{Name: "Org"}
		_, _ = svc.Create(req, &userID)
	}
	items, total := svc.List(models.ListOrganizationsQuery{Limit: 10})
	if total != 5 {
		t.Errorf("expected 5, got %d", total)
	}
	if len(items) != 5 {
		t.Errorf("expected 5 items, got %d", len(items))
	}
}
