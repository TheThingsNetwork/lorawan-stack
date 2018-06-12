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
		CREATE TABLE IF NOT EXISTS clients (
			id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			client_id            VARCHAR(36) UNIQUE NOT NULL,
			description          VARCHAR NOT NULL DEFAULT '',
			secret               VARCHAR NOT NULL,
			redirect_uri         VARCHAR NOT NULL,
			state                INT NOT NULL DEFAULT 0,
			skip_authorization   BOOL NOT NULL DEFAULT false,
			grants               VARCHAR NOT NULL,
			rights               VARCHAR NOT NULL,
			creator_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS clients;
	`

	Registry.Register(4, "4_clients_initial_schema", forwards, backwards)
}
