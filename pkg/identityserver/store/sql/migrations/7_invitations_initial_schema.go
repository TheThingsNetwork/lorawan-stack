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
		CREATE TABLE IF NOT EXISTS invitations (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			token        STRING NOT NULL,
			email        STRING UNIQUE NOT NULL,
			issued_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at   TIMESTAMP NOT NULL
		);
	`
	const backwards = `
		DROP TABLE IF EXISTS invitations;
	`

	Registry.Register(7, "7_invitations_initial_schema", forwards, backwards)
}
