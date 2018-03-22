// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS applications (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			application_id   STRING(36) UNIQUE NOT NULL,
			description      STRING NOT NULL DEFAULT '',
			created_at       TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			updated_at       TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys (
			application_id   UUID NOT NULL REFERENCES applications(id),
			key_name         STRING(36) NOT NULL,
			key              STRING UNIQUE NOT NULL,
			PRIMARY KEY(application_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys_rights (
			application_id   UUID NOT NULL REFERENCES applications(id),
			key_name         STRING(36) NOT NULL,
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, key_name, "right")
		);
		CREATE TABLE IF NOT EXISTS applications_collaborators (
			application_id   UUID NOT NULL REFERENCES applications(id),
			account_id       UUID NOT NULL REFERENCES accounts(id),
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, account_id, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS applications_collaborators;
		DROP TABLE IF EXISTS applications_api_keys_rights;
		DROP TABLE IF EXISTS applications_api_keys;
		DROP TABLE IF EXISTS applications;
	`

	Registry.Register(2, "2_applications_initial_schema", forwards, backwards)
}
