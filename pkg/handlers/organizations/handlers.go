package organizations

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zarinit-routers/cloud-organizations/pkg/handlers"
	"github.com/zarinit-routers/cloud-organizations/pkg/models"
	contracts "github.com/zarinit-routers/cloud-organizations/pkg/services/contracts"
)

// Handlers wires HTTP endpoints to service

type Handlers struct {
	Svc contracts.OrganizationsService
}

func NewHandlers(svc contracts.OrganizationsService) *Handlers { return &Handlers{Svc: svc} }

// Create POST /api/v1/organizations/
func (h *Handlers) Create(c *gin.Context) {
	var req models.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid payload", gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "name is required", nil)
		return
	}
	org, err := h.Svc.Create(req, userIDFromContext(c))
	if err != nil {
		handlers.JSONError(c, http.StatusInternalServerError, "internal", "failed to create", nil)
		return
	}
	c.JSON(http.StatusCreated, org)
}

// Get /:id
func (h *Handlers) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid id", nil)
		return
	}
	org, ok := h.Svc.Get(id)
	if !ok {
		handlers.JSONError(c, http.StatusNotFound, "not_found", "organization not found", nil)
		return
	}
	c.JSON(http.StatusOK, org)
}

// List /
func (h *Handlers) List(c *gin.Context) {
	var q models.ListOrganizationsQuery
	if v := c.Query("tenantId"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			q.TenantID = &id
		} else {
			handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid tenantId", nil)
			return
		}
	}
	q.Q = c.Query("q")
	q.Status = c.Query("status")
	if tags := c.Query("tags"); tags != "" {
		q.Tags = strings.Split(tags, ",")
	}
	q.SortBy = c.DefaultQuery("sortBy", "createdAt")
	q.SortDir = c.DefaultQuery("sortDir", "desc")
	q.Limit = mustAtoiDefault(c.Query("limit"), 50)
	if q.Limit > 200 {
		q.Limit = 200
	}
	q.Offset = mustAtoiDefault(c.Query("offset"), 0)

	items, total := h.Svc.List(q)
	c.JSON(http.StatusOK, models.ListOrganizationsResponse{Items: items, Total: total, Limit: q.Limit, Offset: q.Offset})
}

// Replace PUT /:id
func (h *Handlers) Replace(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid id", nil)
		return
	}
	var req models.ReplaceOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid payload", gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "name is required", nil)
		return
	}
	org, ok := h.Svc.Replace(id, req, userIDFromContext(c))
	if !ok {
		handlers.JSONError(c, http.StatusNotFound, "not_found", "organization not found", nil)
		return
	}
	c.JSON(http.StatusOK, org)
}

// Patch PATCH /:id
func (h *Handlers) Patch(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid id", nil)
		return
	}
	var req models.PatchOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid payload", gin.H{"error": err.Error()})
		return
	}
	org, ok := h.Svc.Patch(id, req, userIDFromContext(c))
	if !ok {
		handlers.JSONError(c, http.StatusNotFound, "not_found", "organization not found", nil)
		return
	}
	c.JSON(http.StatusOK, org)
}

// Delete DELETE /:id
func (h *Handlers) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid id", nil)
		return
	}
	ok := h.Svc.SoftDelete(id, userIDFromContext(c))
	if !ok {
		handlers.JSONError(c, http.StatusNotFound, "not_found", "organization not found", nil)
		return
	}
	c.Status(http.StatusNoContent)
}

// Restore POST /:id/restore
func (h *Handlers) Restore(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "invalid id", nil)
		return
	}
	org, ok := h.Svc.Restore(id, userIDFromContext(c))
	if !ok {
		handlers.JSONError(c, http.StatusNotFound, "not_found", "organization not found or not deleted", nil)
		return
	}
	c.JSON(http.StatusOK, org)
}

// Bulk endpoints DTOs

type bulkCreateRequest struct {
	Items []models.CreateOrganizationRequest `json:"items"`
}

type bulkUpdateRequest struct {
	IDs   []uuid.UUID                     `json:"ids"`
	Patch models.PatchOrganizationRequest `json:"patch"`
}

type bulkDeleteRequest struct {
	IDs []uuid.UUID `json:"ids"`
}

func (h *Handlers) BulkCreate(c *gin.Context) {
	var req bulkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "items are required", nil)
		return
	}
	items, err := h.Svc.BulkCreate(req.Items, userIDFromContext(c))
	if err != nil {
		handlers.JSONError(c, http.StatusInternalServerError, "internal", "failed to create", nil)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"items": items, "total": len(items)})
}

func (h *Handlers) BulkUpdate(c *gin.Context) {
	var req bulkUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "ids are required", nil)
		return
	}
	count, _ := h.Svc.BulkUpdate(req.IDs, req.Patch, userIDFromContext(c))
	c.JSON(http.StatusOK, gin.H{"updated": count})
}

func (h *Handlers) BulkDelete(c *gin.Context) {
	var req bulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		handlers.JSONError(c, http.StatusBadRequest, "bad_request", "ids are required", nil)
		return
	}
	count := h.Svc.BulkDelete(req.IDs, userIDFromContext(c))
	c.JSON(http.StatusOK, gin.H{"deleted": count})
}

// helpers

func mustAtoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// Pull user id from JWT middleware claims if available.
// We keep it simple: header X-User-ID or claim `sub` UUID.
func userIDFromContext(c *gin.Context) *uuid.UUID {
	if v := c.GetHeader("X-User-ID"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			return &id
		}
	}
	if v, ok := c.Get("user_id"); ok {
		if s, ok := v.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				return &id
			}
		}
	}
	return nil
}
