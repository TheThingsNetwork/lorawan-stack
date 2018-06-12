// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS authorization_codes (
			id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			authorization_code   VARCHAR(64) UNIQUE NOT NULL,
			client_id            UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
			created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_in           INTEGER NOT NULL,
			scope                VARCHAR NOT NULL,
			redirect_uri         VARCHAR NOT NULL,
			state                VARCHAR NOT NULL,
			user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS access_tokens (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			access_token    VARCHAR UNIQUE NOT NULL,
			client_id       UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
			user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_in      INTEGER NOT NULL,
			scope           VARCHAR NOT NULL,
			redirect_uri    VARCHAR NOT NULL
		);

		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			refresh_token   VARCHAR(64) UNIQUE NOT NULL,
			client_id       UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
			user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			scope           VARCHAR NOT NULL,
			redirect_uri    VARCHAR NOT NULL
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS refresh_tokens;
		DROP TABLE IF EXISTS access_tokens;
		DROP TABLE IF EXISTS authorization_codes;
	`

	Registry.Register(5, "5_oauth_initial_schema", forwards, backwards)
}
