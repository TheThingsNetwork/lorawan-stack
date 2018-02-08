// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS applications (
			application_id   STRING(36) PRIMARY KEY,
			description      STRING NOT NULL DEFAULT '',
			created_at       TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			updated_at       TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys (
			key              STRING NOT NULL PRIMARY KEY,
			application_id   STRING(36) NOT NULL REFERENCES applications(application_id),
			key_name         STRING(36) NOT NULL,
			UNIQUE(application_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys_rights (
			application_id   STRING(36) NOT NULL REFERENCES applications(application_id),
			key_name         STRING(36) NOT NULL,
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, key_name, "right")
		);
		CREATE TABLE IF NOT EXISTS applications_collaborators (
			application_id   STRING(36) REFERENCES applications(application_id),
			account_id       STRING(36) REFERENCES accounts(account_id),
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
