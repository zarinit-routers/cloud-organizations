package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zarinit-routers/cloud-organizations/pkg/models"
)

// Repo is a Postgres implementation of repository.Organizations.
type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) Create(ctx context.Context, req models.CreateOrganizationRequest, userID *uuid.UUID) (*models.Organization, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var id uuid.UUID
	var createdAt, updatedAt time.Time
	var status string
	if req.Status != nil {
		status = models.NormalizeStatus(*req.Status)
	} else {
		status = models.NormalizeStatus("")
	}
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}
	err = tx.QueryRow(ctx,
		`INSERT INTO organizations (tenant_id, name, legal_code, status, tags, passphrase_hash)
		 VALUES ($1,$2,$3,$4,$5,NULL)
		 RETURNING id, created_at, updated_at`,
		req.TenantID, strings.TrimSpace(req.Name), req.LegalCode, status, tags,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert org: %w", err)
	}

	// addresses
	for _, a := range req.Addresses {
		_, err = tx.Exec(ctx,
			`INSERT INTO org_addresses (id, organization_id, type, country, region, city, street, zip)
			 VALUES (COALESCE($1, uuid_generate_v4()), $2, $3, $4, $5, $6, $7, $8)`,
			uuidOrNil(a.ID), id, a.Type, a.Country, a.Region, a.City, a.Street, a.ZIP,
		)
		if err != nil {
			return nil, fmt.Errorf("insert address: %w", err)
		}
	}
	for _, c := range req.Contacts {
		_, err = tx.Exec(ctx,
			`INSERT INTO org_contacts (id, organization_id, type, value, is_primary)
			 VALUES (COALESCE($1, uuid_generate_v4()), $2, $3, $4, $5)`,
			uuidOrNil(c.ID), id, c.Type, c.Value, c.IsPrimary,
		)
		if err != nil {
			return nil, fmt.Errorf("insert contact: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	// fetch full entity
	org, _, err := r.getInternal(ctx, id)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (r *Repo) Get(ctx context.Context, id uuid.UUID) (*models.Organization, bool, error) {
	org, ok, err := r.getInternal(ctx, id)
	return org, ok, err
}

func (r *Repo) getInternal(ctx context.Context, id uuid.UUID) (*models.Organization, bool, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, name, legal_code, status, tags, created_at, updated_at, deleted_at FROM organizations WHERE id=$1`, id)
	var (
		org       models.Organization
		deletedAt *time.Time
	)
	org.Addresses = []models.OrgAddress{}
	org.Contacts = []models.OrgContact{}
	if err := row.Scan(&org.ID, &org.TenantID, &org.Name, &org.LegalCode, &org.Status, &org.Tags, &org.CreatedAt, &org.UpdatedAt, &deletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}
	org.DeletedAt = deletedAt
	// addresses
	rows, err := r.pool.Query(ctx, `SELECT id, organization_id, type, country, region, city, street, zip FROM org_addresses WHERE organization_id=$1`, org.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var a models.OrgAddress
			if err := rows.Scan(&a.ID, &a.OrganizationID, &a.Type, &a.Country, &a.Region, &a.City, &a.Street, &a.ZIP); err != nil {
				return nil, false, err
			}
			org.Addresses = append(org.Addresses, a)
		}
	}
	rows2, err := r.pool.Query(ctx, `SELECT id, organization_id, type, value, is_primary FROM org_contacts WHERE organization_id=$1`, org.ID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var c models.OrgContact
			if err := rows2.Scan(&c.ID, &c.OrganizationID, &c.Type, &c.Value, &c.IsPrimary); err != nil {
				return nil, false, err
			}
			org.Contacts = append(org.Contacts, c)
		}
	}
	if org.DeletedAt != nil {
		return nil, false, nil
	}
	return &org, true, nil
}

func (r *Repo) List(ctx context.Context, q models.ListOrganizationsQuery) ([]models.Organization, int, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	idx := 1
	if q.TenantID != nil {
		where = append(where, fmt.Sprintf("tenant_id=$%d", idx))
		args = append(args, *q.TenantID)
		idx++
	}
	if strings.TrimSpace(q.Q) != "" {
		where = append(where, fmt.Sprintf("(lower(name) LIKE '%%' || lower($%d) || '%%' OR lower(coalesce(legal_code,'')) LIKE '%%' || lower($%d) || '%%')", idx, idx))
		args = append(args, q.Q)
		idx++
	}
	if strings.TrimSpace(q.Status) != "" {
		where = append(where, fmt.Sprintf("lower(status)=lower($%d)", idx))
		args = append(args, q.Status)
		idx++
	}
	if len(q.Tags) > 0 {
		where = append(where, fmt.Sprintf("tags @> $%d", idx))
		args = append(args, q.Tags)
		idx++
	}
	cond := strings.Join(where, " AND ")
	order := "created_at DESC"
	switch strings.ToLower(q.SortBy) {
	case "name":
		order = "name ASC"
	case "updatedat":
		if strings.ToLower(q.SortDir) == "asc" {
			order = "updated_at ASC"
		} else {
			order = "updated_at DESC"
		}
	default:
		if strings.ToLower(q.SortDir) == "asc" {
			order = "created_at ASC"
		} else {
			order = "created_at DESC"
		}
	}
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := q.Offset
	countSQL := fmt.Sprintf("SELECT count(1) FROM organizations WHERE %s", cond)
	row := r.pool.QueryRow(ctx, countSQL, args...)
	var total int
	if err := row.Scan(&total); err != nil {
		return nil, 0, err
	}
	listSQL := fmt.Sprintf("SELECT id, tenant_id, name, legal_code, status, tags, created_at, updated_at, deleted_at FROM organizations WHERE %s ORDER BY %s LIMIT %d OFFSET %d", cond, order, limit, offset)
	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []models.Organization{}
	for rows.Next() {
		var it models.Organization
		var deletedAt *time.Time
		if err := rows.Scan(&it.ID, &it.TenantID, &it.Name, &it.LegalCode, &it.Status, &it.Tags, &it.CreatedAt, &it.UpdatedAt, &deletedAt); err != nil {
			return nil, 0, err
		}
		if deletedAt != nil {
			continue
		}
		items = append(items, it)
	}
	return items, total, nil
}

func (r *Repo) Replace(ctx context.Context, id uuid.UUID, req models.ReplaceOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	cmd, err := tx.Exec(ctx, `UPDATE organizations SET tenant_id=$2, name=$3, legal_code=$4, status=$5, tags=$6 WHERE id=$1 AND deleted_at IS NULL`, id, req.TenantID, strings.TrimSpace(req.Name), req.LegalCode, models.NormalizeStatus(req.Status), req.Tags)
	if err != nil {
		return nil, false, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, false, nil
	}
	// replace addresses/contacts
	if _, err := tx.Exec(ctx, `DELETE FROM org_addresses WHERE organization_id=$1`, id); err != nil {
		return nil, false, err
	}
	for _, a := range req.Addresses {
		_, err = tx.Exec(ctx, `INSERT INTO org_addresses (id, organization_id, type, country, region, city, street, zip) VALUES (COALESCE($1, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8)`, uuidOrNil(a.ID), id, a.Type, a.Country, a.Region, a.City, a.Street, a.ZIP)
		if err != nil {
			return nil, false, err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM org_contacts WHERE organization_id=$1`, id); err != nil {
		return nil, false, err
	}
	for _, c := range req.Contacts {
		_, err = tx.Exec(ctx, `INSERT INTO org_contacts (id, organization_id, type, value, is_primary) VALUES (COALESCE($1, gen_random_uuid()), $2, $3, $4, $5)`, uuidOrNil(c.ID), id, c.Type, c.Value, c.IsPrimary)
		if err != nil {
			return nil, false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	org, ok, err := r.getInternal(ctx, id)
	return org, ok, err
}

func (r *Repo) Patch(ctx context.Context, id uuid.UUID, req models.PatchOrganizationRequest, userID *uuid.UUID) (*models.Organization, bool, error) {
	// dynamic update
	sets := []string{}
	args := []any{}
	idx := 1
	if req.TenantID != nil {
		sets = append(sets, fmt.Sprintf("tenant_id=$%d", idx))
		args = append(args, *req.TenantID)
		idx++
	}
	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name=$%d", idx))
		args = append(args, strings.TrimSpace(*req.Name))
		idx++
	}
	if req.LegalCode != nil {
		sets = append(sets, fmt.Sprintf("legal_code=$%d", idx))
		args = append(args, *req.LegalCode)
		idx++
	}
	if req.Status != nil {
		sets = append(sets, fmt.Sprintf("status=$%d", idx))
		args = append(args, models.NormalizeStatus(*req.Status))
		idx++
	}
	if req.Tags != nil {
		sets = append(sets, fmt.Sprintf("tags=$%d", idx))
		args = append(args, *req.Tags)
		idx++
	}
	if len(sets) > 0 {
		args = append(args, id)
		sql := fmt.Sprintf("UPDATE organizations SET %s WHERE id=$%d AND deleted_at IS NULL", strings.Join(sets, ", "), idx)
		cmd, err := r.pool.Exec(ctx, sql, args...)
		if err != nil {
			return nil, false, err
		}
		if cmd.RowsAffected() == 0 {
			return nil, false, nil
		}
	}
	// addresses
	if req.Addresses != nil {
		if _, err := r.pool.Exec(ctx, `DELETE FROM org_addresses WHERE organization_id=$1`, id); err != nil {
			return nil, false, err
		}
		for _, a := range *req.Addresses {
			_, err := r.pool.Exec(ctx, `INSERT INTO org_addresses (id, organization_id, type, country, region, city, street, zip) VALUES (COALESCE($1, gen_random_uuid()), $2, $3, $4, $5, $6, $7, $8)`, uuidOrNil(a.ID), id, a.Type, a.Country, a.Region, a.City, a.Street, a.ZIP)
			if err != nil {
				return nil, false, err
			}
		}
	}
	if req.Contacts != nil {
		if _, err := r.pool.Exec(ctx, `DELETE FROM org_contacts WHERE organization_id=$1`, id); err != nil {
			return nil, false, err
		}
		for _, c := range *req.Contacts {
			_, err := r.pool.Exec(ctx, `INSERT INTO org_contacts (id, organization_id, type, value, is_primary) VALUES (COALESCE($1, gen_random_uuid()), $2, $3, $4, $5)`, uuidOrNil(c.ID), id, c.Type, c.Value, c.IsPrimary)
			if err != nil {
				return nil, false, err
			}
		}
	}
	org, ok, err := r.getInternal(ctx, id)
	return org, ok, err
}

func (r *Repo) SoftDelete(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (bool, error) {
	cmd, err := r.pool.Exec(ctx, `UPDATE organizations SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() > 0, nil
}

func (r *Repo) Restore(ctx context.Context, id uuid.UUID, userID *uuid.UUID) (*models.Organization, bool, error) {
	cmd, err := r.pool.Exec(ctx, `UPDATE organizations SET deleted_at=NULL WHERE id=$1 AND deleted_at IS NOT NULL`, id)
	if err != nil {
		return nil, false, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, false, nil
	}
	org, ok, err := r.getInternal(ctx, id)
	return org, ok, err
}

func (r *Repo) BulkCreate(ctx context.Context, reqs []models.CreateOrganizationRequest, userID *uuid.UUID) ([]models.Organization, error) {
	out := make([]models.Organization, 0, len(reqs))
	for _, req := range reqs {
		org, err := r.Create(ctx, req, userID)
		if err != nil {
			return nil, err
		}
		out = append(out, *org)
	}
	return out, nil
}

func (r *Repo) BulkUpdate(ctx context.Context, ids []uuid.UUID, patch models.PatchOrganizationRequest, userID *uuid.UUID) (int, error) {
	updated := 0
	for _, id := range ids {
		if _, ok, err := r.Patch(ctx, id, patch, userID); err != nil {
			return updated, err
		} else if ok {
			updated++
		}
	}
	return updated, nil
}

func (r *Repo) BulkDelete(ctx context.Context, ids []uuid.UUID, userID *uuid.UUID) (int, error) {
	deleted := 0
	for _, id := range ids {
		ok, err := r.SoftDelete(ctx, id, userID)
		if err != nil {
			return deleted, err
		}
		if ok {
			deleted++
		}
	}
	return deleted, nil
}

func (r *Repo) AddMember(ctx context.Context, orgID, user uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO org_members (user_id, org_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, user, orgID)
	return err
}

func (r *Repo) RemoveMember(ctx context.Context, orgID, user uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM org_members WHERE user_id=$1 AND org_id=$2`, user, orgID)
	return err
}

func (r *Repo) ListMembers(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]uuid.UUID, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(1) FROM org_members WHERE org_id=$1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT user_id FROM org_members WHERE org_id=$1 ORDER BY user_id LIMIT $2 OFFSET $3`, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	res := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, 0, err
		}
		res = append(res, id)
	}
	return res, total, nil
}

func uuidOrNil(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
