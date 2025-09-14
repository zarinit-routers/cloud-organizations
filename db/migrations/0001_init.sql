CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- organizations table with soft delete and audit timestamps
CREATE TABLE IF NOT EXISTS organizations (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  tenant_id uuid NULL,
  name varchar(256) NOT NULL,
  legal_code varchar(256) NULL,
  status varchar(32) NOT NULL DEFAULT 'active',
  tags text[] NOT NULL DEFAULT '{}',
  passphrase_hash varchar(256) NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_org_tenant_legalcode
  ON organizations(tenant_id, legal_code) WHERE legal_code IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_org_status ON organizations(status);
CREATE INDEX IF NOT EXISTS idx_org_updated_at ON organizations(updated_at);
CREATE INDEX IF NOT EXISTS idx_org_tags_gin ON organizations USING GIN(tags);

-- org_members mapping
CREATE TABLE IF NOT EXISTS org_members (
  user_id uuid NOT NULL,
  org_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, org_id)
);

CREATE INDEX IF NOT EXISTS idx_org_members_org_id ON org_members(org_id);
CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON org_members(user_id);

-- updated_at trigger
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END; $$ LANGUAGE plpgsql;
CREATE TRIGGER trg_org_updated_at BEFORE UPDATE ON organizations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
