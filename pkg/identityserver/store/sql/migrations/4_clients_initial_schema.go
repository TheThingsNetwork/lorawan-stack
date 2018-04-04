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
			id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			client_id          STRING(36) UNIQUE NOT NULL,
			description        STRING NOT NULL DEFAULT '',
			secret             STRING NOT NULL,
			redirect_uri       STRING NOT NULL,
			state	             INT NOT NULL DEFAULT 0,
			official_labeled   BOOL NOT NULL DEFAULT false,
			grants             STRING NOT NULL,
			rights             STRING NOT NULL,
			creator_id         UUID NOT NULL REFERENCES users(id),
			created_at         TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			updated_at         TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS clients;
	`

	Registry.Register(4, "4_clients_initial_schema", forwards, backwards)
}
