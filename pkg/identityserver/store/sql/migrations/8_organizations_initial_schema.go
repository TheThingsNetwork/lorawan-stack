// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS organizations (
			id                UUID PRIMARY KEY REFERENCES accounts(id),
			organization_id   VARCHAR(36) UNIQUE NOT NULL REFERENCES accounts(account_id),
			name              VARCHAR NOT NULL,
			description       VARCHAR NOT NULL DEFAULT '',
			url               VARCHAR NOT NULL DEFAULT '',
			location          VARCHAR NOT NULL DEFAULT '',
			email             VARCHAR NOT NULL,
			created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS organizations_organization_id ON organizations (organization_id);
		CREATE TABLE IF NOT EXISTS organizations_members (
			organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			"right"           VARCHAR NOT NULL,
			PRIMARY KEY(organization_id, user_id, "right")
		);
		CREATE TABLE IF NOT EXISTS organizations_api_keys (
			organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			key_name          VARCHAR(36) NOT NULL,
			key               VARCHAR NOT NULL UNIQUE,
			PRIMARY KEY(organization_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS organizations_api_keys_rights (
			organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			key_name          VARCHAR(36) NOT NULL,
			"right"           VARCHAR NOT NULL,
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
