// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS authorization_codes (
			authorization_code STRING(64) PRIMARY KEY,
			client_id          STRING(36) REFERENCES clients(client_id) NOT NULL,
			created_at         TIMESTAMP DEFAULT current_timestamp(),
			expires_in         INTEGER,
			scope              STRING,
			redirect_uri       STRING,
			state              STRING,
			username           STRING
		);

		CREATE TABLE IF NOT EXISTS refresh_tokens (
			refresh_token STRING(64) PRIMARY KEY,
			client_id    STRING(36) REFERENCES clients(client_id) NOT NULL,
			created_at   TIMESTAMP DEFAULT current_timestamp(),
			scope		 STRING,
			redirect_uri STRING
		);

	`

	const backwards = `
		DROP TABLE IF EXISTS refresh_tokens;
		DROP TABLE IF EXISTS authorization_codes;
	`

	Registry.Register(5, "5_tokens_initial_schema", forwards, backwards)
}
