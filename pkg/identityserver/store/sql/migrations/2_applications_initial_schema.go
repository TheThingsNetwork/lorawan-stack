// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS applications (
			id            STRING(36) PRIMARY KEY,
			description   TEXT,
			created       TIMESTAMP DEFAULT current_timestamp(),
			archived      TIMESTAMP DEFAULT null
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys (
			app_id    STRING(36) REFERENCES applications(id),
			name      STRING(36) NOT NULL,
			key       STRING NOT NULL,
			"right"   TEXT NOT NULL,
			PRIMARY KEY(app_id, name, "right"),
			UNIQUE(app_id, key, "right")
		);
		CREATE TABLE IF NOT EXISTS applications_euis (
			app_id   STRING(36) REFERENCES applications(id),
			eui      BYTES,
			PRIMARY KEY(app_id, eui)
		);
		CREATE TABLE IF NOT EXISTS applications_collaborators (
			app_id     STRING(36) REFERENCES applications(id),
			username   STRING(36) REFERENCES users(username),
			"right"    STRING NOT NULL,
			PRIMARY KEY(app_id, username, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS applications_collaborators;
		DROP TABLE IF EXISTS applications_euis;
		DROP TABLE IF EXISTS applications_api_keys;
		DROP TABLE IF EXISTS applications;
	`

	Registry.Register(2, "2_applications_initial_schema", forwards, backwards)
}
