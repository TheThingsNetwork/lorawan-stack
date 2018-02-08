// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS organizations (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id   STRING(36) UNIQUE NOT NULL REFERENCES accounts(account_id),
			name              STRING NOT NULL,
			description       STRING NOT NULL,
			url               STRING NOT NULL,
			location          STRING NOT NULL,
			email             STRING NOT NULL,
			created_at        TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			updated_at        TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS organizations_organization_id ON organizations (organization_id);
		CREATE TABLE IF NOT EXISTS organizations_members (
			organization_id   STRING(36) NOT NULL REFERENCES organizations(organization_id),
			user_id           STRING(36) NOT NULL REFERENCES users(user_id),
			"right"           STRING NOT NULL,
			PRIMARY KEY(organization_id, user_id, "right")
		);
		CREATE TABLE IF NOT EXISTS organizations_api_keys (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id   STRING(36) NOT NULL REFERENCES organizations(organization_id),
			key               STRING NOT NULL UNIQUE,
			key_name          STRING(36) NOT NULL,
			UNIQUE(organization_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS organizations_api_keys_rights (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id   STRING(36) NOT NULL,
			key_name          STRING(36) NOT NULL,
			"right"           STRING NOT NULL,
			UNIQUE(organization_id, key_name, "right"),
			CONSTRAINT fk_key FOREIGN KEY (organization_id, key_name) REFERENCES organizations_api_keys (organization_id, key_name)
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
