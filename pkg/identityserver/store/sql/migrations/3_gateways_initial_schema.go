// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS gateways (
			gateway_id          STRING(36) PRIMARY KEY,
			description         STRING,
			frequency_plan_id   STRING(36) NOT NULL,
			activated_at        TIMESTAMP DEFAULT null,
			privacy_settings    STRING,
			auto_update         BOOL DEFAULT true,
			platform            STRING,
			cluster_address     STRING NOT NULL,
			created_at          TIMESTAMP DEFAULT current_timestamp(),
			updated_at          TIMESTAMP DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS gateways_attributes (
			gateway_id   STRING(36) REFERENCES gateways(gateway_id),
			attribute    STRING NOT NULL,
			value        STRING NOT NULL,
			PRIMARY KEY(gateway_id, attribute)
		);
		CREATE TABLE IF NOT EXISTS gateways_antennas (
			antenna_id   STRING DEFAULT to_hex(unique_rowid()) PRIMARY KEY,
			gateway_id   STRING(36) REFERENCES gateways(gateway_id) NOT NULL,
			gain         FLOAT,
			type         STRING,
			model        STRING,
			placement    STRING,
			longitude    FLOAT,
			latitude     FLOAT,
			altitude     INT,
			created_at   TIMESTAMP DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS gateways_collaborators (
			gateway_id   STRING(36) REFERENCES gateways(gateway_id),
			user_id      STRING(36) REFERENCES users(user_id),
			"right"      STRING NOT NULL,
			PRIMARY KEY(gateway_id, user_id, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS gateways_attributes;
		DROP TABLE IF EXISTS gateways_antennas;
		DROP TABLE IF EXISTS gateways_collaborators;
		DROP TABLE IF EXISTS gateways;
	`

	Registry.Register(3, "3_gateways_initial_schema", forwards, backwards)
}
