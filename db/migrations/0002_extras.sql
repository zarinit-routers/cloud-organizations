-- Additional tables for addresses and contacts
CREATE TABLE IF NOT EXISTS org_addresses (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  type varchar(32) NOT NULL,
  country varchar(128) NULL,
  region varchar(128) NULL,
  city varchar(128) NULL,
  street varchar(256) NULL,
  zip varchar(32) NULL
);

CREATE INDEX IF NOT EXISTS idx_org_addresses_org_id ON org_addresses(organization_id);

CREATE TABLE IF NOT EXISTS org_contacts (
  id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  type varchar(32) NOT NULL,
  value varchar(256) NOT NULL,
  is_primary boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_org_contacts_org_id ON org_contacts(organization_id);
