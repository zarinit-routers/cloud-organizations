-- +migrate Up
CREATE TABLE
    IF NOT EXISTS organizations (
        id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
        name VARCHAR(255) NOT NULL,
        passphrase VARCHAR(255),
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP
    );

CREATE TABLE
    IF NOT EXISTS members (
        organization_id uuid REFERENCES organizations (id) ON DELETE CASCADE,
        user_id uuid NOT NULL,
        created_at TIMESTAMP NOT NULL
    );

-- +migrate Down
DROP TABLE organizations;
