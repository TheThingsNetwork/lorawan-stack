// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS gateways (
			id                  STRING(36) PRIMARY KEY,
			description         TEXT,
			frequency_plan      STRING(36) NOT NULL,
			key                 STRING(36),
			activated           BOOL DEFAULT false,
			status_public       BOOL DEFAULT false,
			location_public     BOOL DEFAULT false,
			owner_public        BOOL DEFAULT false,
			auto_update         BOOL DEFAULT true,
			brand               STRING,
			model               STRING,
			antenna_type        STRING,
			antenna_model       STRING,
			antenna_placement   STRING,
			antenna_altitude    STRING,
			antenna_location    STRING,
			routers             TEXT NOT NULL,
			created             TIMESTAMP DEFAULT current_timestamp(),
			archived            TIMESTAMP DEFAULT null
		);
		CREATE TABLE IF NOT EXISTS gateways_attributes (
			gateway_id   STRING(36) REFERENCES gateways(id),
			attribute    STRING NOT NULL,
			value        STRING NOT NULL,
			PRIMARY KEY(gateway_id, attribute)
		);
		CREATE TABLE IF NOT EXISTS gateways_collaborators (
			gateway_id   STRING(36) REFERENCES gateways(id),
			username     STRING(36) REFERENCES users(username),
			"right"      STRING NOT NULL,
			PRIMARY KEY(gateway_id, username, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS gateways_attributes;
		DROP TABLE IF EXISTS gateways_collaborators;
		DROP TABLE IF EXISTS gateways;
	`

	Registry.Register(3, "3_gateways_initial_schema", forwards, backwards)
}
