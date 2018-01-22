// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS invitations (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			token        STRING UNIQUE NOT NULL,
			email        STRING NOT NULL,
			sent_at      TIMESTAMP DEFAULT current_timestamp(),
			ttl          INT NOT NULL,
			used_at      TIMESTAMP,
			user_id      STRING(36) REFERENCES users(user_id)
		);
	`
	const backwards = `
		DROP TABLE IF EXISTS invitations;
	`

	Registry.Register(7, "7_invitations_initial_schema", forwards, backwards)
}
