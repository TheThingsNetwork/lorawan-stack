// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS applications (
			application_id   STRING(36) PRIMARY KEY,
			description      TEXT,
			created_at       TIMESTAMP DEFAULT current_timestamp(),
			updated_at       TIMESTAMP DEFAULT current_timestamp(),
			archived_at      TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys (
			application_id   STRING(36) REFERENCES applications(application_id),
			name             STRING(36) NOT NULL,
			key              STRING NOT NULL,
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, name, "right"),
			UNIQUE(application_id, key, "right")
		);
		CREATE TABLE IF NOT EXISTS applications_collaborators (
			application_id   STRING(36) REFERENCES applications(application_id),
			user_id          STRING(36) REFERENCES users(user_id),
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, user_id, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS applications_collaborators;
		DROP TABLE IF EXISTS applications_api_keys;
		DROP TABLE IF EXISTS applications;
	`

	Registry.Register(2, "2_applications_initial_schema", forwards, backwards)
}
