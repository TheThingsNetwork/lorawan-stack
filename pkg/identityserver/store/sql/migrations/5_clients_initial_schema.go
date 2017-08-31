// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS clients (
			id            STRING(36) PRIMARY KEY,
			description   TEXT,
			secret        STRING NOT NULL,
			uri           STRING NOT NULL,
			state	      INT NOT NULL DEFAULT 0,
			official      BOOL DEFAULT false,
			grants        TEXT NOT NULL,
			scope         TEXT NOT NULL,
			created       TIMESTAMP DEFAULT current_timestamp(),
			archived      TIMESTAMP DEFAULT null
		);
		CREATE TABLE IF NOT EXISTS clients_collaborators (
			client_id   STRING(36) REFERENCES clients(id),
			username    STRING(36) REFERENCES users(username),
			"right"     STRING NOT NULL,
			PRIMARY KEY(client_id, username, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS clients_collaborators;
		DROP TABLE IF EXISTS clients;
	`

	Registry.Register(5, "5_clients_initial_schema", forwards, backwards)
}
