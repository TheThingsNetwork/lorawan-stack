// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS organizations (
			id                UUID PRIMARY KEY REFERENCES accounts(id),
			organization_id   STRING(36) UNIQUE NOT NULL REFERENCES accounts(account_id),
			name              STRING NOT NULL,
			description       STRING NOT NULL DEFAULT '',
			url               STRING NOT NULL DEFAULT '',
			location          STRING NOT NULL DEFAULT '',
			email             STRING NOT NULL,
			created_at        TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			updated_at        TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS organizations_organization_id ON organizations (organization_id);
		CREATE TABLE IF NOT EXISTS organizations_members (
			organization_id   UUID NOT NULL REFERENCES organizations(id),
			user_id           UUID NOT NULL REFERENCES users(id),
			"right"           STRING NOT NULL,
			PRIMARY KEY(organization_id, user_id, "right")
		);
		CREATE TABLE IF NOT EXISTS organizations_api_keys (
			organization_id   UUID NOT NULL REFERENCES organizations(id),
			key_name          STRING(36) NOT NULL,
			key               STRING NOT NULL UNIQUE,
			PRIMARY KEY(organization_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS organizations_api_keys_rights (
			organization_id    UUID NOT NULL REFERENCES organizations(id),
			key_name           STRING(36) NOT NULL,
			"right"             STRING NOT NULL,
			PRIMARY KEY(organization_id, key_name, "right")
		);
	`
	const backwards = `
		DROP TABLE IF EXISTS organizations_api_keys_rights;
		DROP TABLE IF EXISTS organizations_api_keys;
		DROP TABLE IF EXISTS organizations_members;
		DROP TABLE IF EXISTS organizations;
	`

	Registry.Register(8, "8_organizations_initial_schema", forwards, backwards)
}
