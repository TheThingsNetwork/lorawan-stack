// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS authorization_codes (
			authorization_code   STRING(64) PRIMARY KEY,
			client_id            STRING(36) NOT NULL REFERENCES clients(client_id),
			created_at           TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			expires_in           INTEGER NOT NULL,
			scope                STRING NOT NULL,
			redirect_uri         STRING NOT NULL,
			state                STRING NOT NULL,
			user_id              STRING(36) NOT NULL REFERENCES users(user_id)
		);

		CREATE TABLE IF NOT EXISTS access_tokens (
			access_token    STRING PRIMARY KEY,
			client_id       STRING(36) NOT NULL REFERENCES clients(client_id),
			user_id         STRING(36) NOT NULL REFERENCES users(user_id),
			created_at      TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			expires_in      INTEGER NOT NULL,
			scope           STRING NOT NULL,
			redirect_uri    STRING NOT NULL
		);

		CREATE TABLE IF NOT EXISTS refresh_tokens (
			refresh_token   STRING(64) PRIMARY KEY,
			client_id       STRING(36) NOT NULL REFERENCES clients(client_id),
			user_id         STRING(36) NOT NULL REFERENCES users(user_id),
			created_at      TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			scope           STRING NOT NULL,
			redirect_uri    STRING NOT NULL
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS refresh_tokens;
		DROP TABLE IF EXISTS access_tokens;
		DROP TABLE IF EXISTS authorization_codes;
	`

	Registry.Register(5, "5_oauth_initial_schema", forwards, backwards)
}
