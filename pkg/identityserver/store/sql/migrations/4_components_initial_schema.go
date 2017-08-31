// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS components (
			id        STRING(36) PRIMARY KEY,
			type      STRING(20) NOT NULL,
			created   TIMESTAMP DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS components_collaborators (
			component_id   STRING(36) REFERENCES components(id),
			username       STRING(36) REFERENCES users(username),
			"right"        STRING NOT NULL,
			PRIMARY KEY(component_id, username, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS components_collaborators;
		DROP TABLE IF EXISTS components;
	`

	Registry.Register(4, "4_components_initial_schema", forwards, backwards)
}
