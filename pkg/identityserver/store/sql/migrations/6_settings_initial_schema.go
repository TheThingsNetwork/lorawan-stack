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
	// By using the `CHECK` constraint in the column `id` we ensure that there is only a single row.
	const forwards = `
		CREATE TABLE IF NOT EXISTS settings (
			id                     INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
			updated_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			blacklisted_ids        STRING NOT NULL DEFAULT '[]',
			skip_validation        BOOL NOT NULL DEFAULT false,
			invitation_only        BOOL NOT NULL DEFAULT false,
			admin_approval         BOOL NOT NULL DEFAULT false,
			validation_token_ttl   INT NOT NULL,
			allowed_emails         STRING NOT NULL DEFAULT '[]',
			invitation_token_ttl   INT NOT NULL
		);
	`
	const backwards = `
		DROP TABLE IF EXISTS settings;
	`

	Registry.Register(6, "6_settings_initial_schema", forwards, backwards)
}
