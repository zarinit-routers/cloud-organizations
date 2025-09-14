-- +migrate Up
ALTER TABLE members ADD PRIMARY KEY (organization_id, user_id);

-- +migrate Down
ALTER TABLE members
DROP PRIMARY KEY;
