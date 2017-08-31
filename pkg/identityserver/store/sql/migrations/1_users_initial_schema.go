// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS users (
			username    STRING(36) PRIMARY KEY,
			email       TEXT NOT NULL UNIQUE,
			password    TEXT NOT NULL,
			joined      TIMESTAMP DEFAULT current_timestamp(),
			archived    TIMESTAMP,
			validated   BOOL DEFAULT false,
			admin       BOOL DEFAULT false,
			god         BOOL DEFAULT false
		);
		CREATE UNIQUE INDEX IF NOT EXISTS user_email ON users (email);
	`

	const backwards = `
		DROP TABLE IF EXISTS users;
	`

	Registry.Register(1, "1_users_initial_schema", forwards, backwards)
}
