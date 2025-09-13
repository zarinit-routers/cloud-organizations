-- Organizations queries
-- name: CreateOrganization :one
INSERT INTO organizations (
  tenant_id, name, legal_code, status, tags, passphrase_hash
) VALUES (
  $1, $2, $3, $4, COALESCE($5, '{}'), $6
) RETURNING id, tenant_id, name, legal_code, status, tags, passphrase_hash, created_at, updated_at, deleted_at;

-- name: GetOrganization :one
SELECT id, tenant_id, name, legal_code, status, tags, passphrase_hash, created_at, updated_at, deleted_at
FROM organizations WHERE id = $1;

-- name: ListOrganizations :many
SELECT id, tenant_id, name, legal_code, status, tags, passphrase_hash, created_at, updated_at, deleted_at
FROM organizations
WHERE ($1::uuid IS NULL OR tenant_id = $1)
  AND ($2::text IS NULL OR (lower(name) LIKE '%' || lower($2) || '%' OR lower(COALESCE(legal_code,'')) LIKE '%' || lower($2) || '%'))
  AND ($3::text IS NULL OR lower(status) = lower($3))
  AND ($4::text[] IS NULL OR tags @> $4)
  AND deleted_at IS NULL
ORDER BY
  CASE WHEN $5 = 'name' THEN name END ASC,
  CASE WHEN $5 = 'createdAt' AND $6 = 'asc' THEN created_at END ASC,
  CASE WHEN $5 = 'createdAt' AND $6 = 'desc' THEN created_at END DESC,
  CASE WHEN $5 = 'updatedAt' AND $6 = 'asc' THEN updated_at END ASC,
  CASE WHEN $5 = 'updatedAt' AND $6 = 'desc' THEN updated_at END DESC
LIMIT $7 OFFSET $8;

-- name: ReplaceOrganization :one
UPDATE organizations SET
  tenant_id = $2,
  name = $3,
  legal_code = $4,
  status = $5,
  tags = COALESCE($6, '{}')
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, tenant_id, name, legal_code, status, tags, passphrase_hash, created_at, updated_at, deleted_at;

-- name: PatchOrganization :one
UPDATE organizations SET
  tenant_id = COALESCE($2, tenant_id),
  name = COALESCE($3, name),
  legal_code = $4,
  status = COALESCE($5, status),
  tags = COALESCE($6, tags)
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, tenant_id, name, legal_code, status, tags, passphrase_hash, created_at, updated_at, deleted_at;

-- name: SoftDeleteOrganization :execrows
UPDATE organizations SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL;

-- name: RestoreOrganization :one
UPDATE organizations SET deleted_at = NULL WHERE id = $1 AND deleted_at IS NOT NULL
RETURNING id, tenant_id, name, legal_code, status, tags, passphrase_hash, created_at, updated_at, deleted_at;

-- Members
-- name: AddMember :exec
INSERT INTO org_members (user_id, org_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveMember :execrows
DELETE FROM org_members WHERE user_id = $1 AND org_id = $2;

-- name: ListMembers :many
SELECT user_id FROM org_members WHERE org_id = $1 ORDER BY user_id LIMIT $2 OFFSET $3;