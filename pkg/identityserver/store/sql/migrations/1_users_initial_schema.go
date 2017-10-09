// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS users (
			user_id       STRING(36) PRIMARY KEY,
			name          STRING NOT NULL,
			email         TEXT NOT NULL UNIQUE,
			password      TEXT NOT NULL,
			validated     BOOL DEFAULT false,
			admin         BOOL DEFAULT false,
			created_at    TIMESTAMP DEFAULT current_timestamp(),
			updated_at    TIMESTAMP,
			archived_at   TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS users_email ON users (email);
	`

	const backwards = `
		DROP TABLE IF EXISTS users;
	`

	Registry.Register(1, "1_users_initial_schema", forwards, backwards)
}
