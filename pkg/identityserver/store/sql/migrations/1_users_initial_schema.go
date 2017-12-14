// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS users (
			user_id        STRING(36) PRIMARY KEY,
			name           STRING NOT NULL,
			email          TEXT NOT NULL UNIQUE,
			password       TEXT NOT NULL,
			validated_at   TIMESTAMP,
			admin          BOOL DEFAULT false,
			created_at     TIMESTAMP DEFAULT current_timestamp(),
			updated_at     TIMESTAMP DEFAULT current_timestamp()
		);
		CREATE UNIQUE INDEX IF NOT EXISTS users_email ON users (email);

		CREATE TABLE IF NOT EXISTS validation_tokens (
			validation_token   STRING PRIMARY KEY,
			user_id            STRING(36) REFERENCES users(user_id) NOT NULL,
			created_at         TIMESTAMP DEFAULT current_timestamp(),
			expires_in         INTEGER
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS validation_tokens;
		DROP TABLE IF EXISTS users;
	`

	Registry.Register(1, "1_users_initial_schema", forwards, backwards)
}
