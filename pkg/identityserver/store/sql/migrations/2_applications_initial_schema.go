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
		CREATE TABLE IF NOT EXISTS applications (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			application_id   STRING(36) UNIQUE NOT NULL,
			description      STRING NOT NULL DEFAULT '',
			created_at       TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			updated_at       TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys (
			application_id   UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			key_name         STRING(36) NOT NULL,
			key              STRING UNIQUE NOT NULL,
			PRIMARY KEY(application_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS applications_api_keys_rights (
			application_id   UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			key_name         STRING(36) NOT NULL,
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, key_name, "right")
		);
		CREATE TABLE IF NOT EXISTS applications_collaborators (
			application_id   UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
			account_id       UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, account_id, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS applications_collaborators;
		DROP TABLE IF EXISTS applications_api_keys_rights;
		DROP TABLE IF EXISTS applications_api_keys;
		DROP TABLE IF EXISTS applications;
	`

	Registry.Register(2, "2_applications_initial_schema", forwards, backwards)
}
