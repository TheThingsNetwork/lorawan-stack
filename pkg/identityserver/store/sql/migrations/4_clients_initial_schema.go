// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS clients (
			client_id          STRING(36) PRIMARY KEY,
			description        TEXT,
			secret             STRING NOT NULL,
			redirect_uri       STRING NOT NULL,
			state	           INT NOT NULL DEFAULT 0,
				official_labeled   BOOL DEFAULT false,
			grants             TEXT,
			rights             TEXT,
			created_at         TIMESTAMP DEFAULT current_timestamp(),
			updated_at         TIMESTAMP,
			archived_at        TIMESTAMP
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS clients;
	`

	Registry.Register(4, "4_clients_initial_schema", forwards, backwards)
}
