// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS invitations (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			token        STRING NOT NULL,
			email        STRING UNIQUE NOT NULL,
			issued_at    TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			expires_at   TIMESTAMP NOT NULL
		);
	`
	const backwards = `
		DROP TABLE IF EXISTS invitations;
	`

	Registry.Register(7, "7_invitations_initial_schema", forwards, backwards)
}
