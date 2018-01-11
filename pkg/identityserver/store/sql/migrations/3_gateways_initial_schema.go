// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
		CREATE TABLE IF NOT EXISTS gateways_radios (
			radio_id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
			gateway_id         STRING(36) REFERENCES gateways(gateway_id) NOT NULL,
			frequency          INT,
			tx_configuration   STRING,
			created_at         TIMESTAMP DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS gateways_collaborators (
			gateway_id   STRING(36) REFERENCES gateways(gateway_id),
			user_id      STRING(36) REFERENCES users(user_id),
			"right"      STRING NOT NULL,
			PRIMARY KEY(gateway_id, user_id, "right")
		);
		CREATE TABLE IF NOT EXISTS gateways_api_keys (
			gateway_id   STRING(36) NOT NULL REFERENCES gateways(gateway_id),
			key          STRING PRIMARY KEY,
			key_name     STRING(36) NOT NULL,
			UNIQUE(gateway_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS gateways_api_keys_rights (
			gateway_id   STRING(36) NOT NULL REFERENCES gateways(gateway_id),
			key_name     STRING(36) NOT NULL,
			"right"      STRING NOT NULL,
			PRIMARY KEY(gateway_id, key_name, "right")
		);
		CREATE TABLE IF NOT EXISTS gateways_locked_api_keys (
			gateway_id   STRING(36) PRIMARY KEY REFERENCES gateways(gateway_id),
			key          STRING NOT NULL REFERENCES gateways_api_keys(key)
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS gateways_locked_api_keys;
		DROP TABLE IF EXISTS gateways_api_keys_rights;
		DROP TABLE IF EXISTS gateways_api_keys;
		DROP TABLE IF EXISTS gateways_attributes;
		DROP TABLE IF EXISTS gateways_antennas;
		DROP TABLE IF EXISTS gateways_radios;
		DROP TABLE IF EXISTS gateways_collaborators;
		DROP TABLE IF EXISTS gateways;
	`

	Registry.Register(3, "3_gateways_initial_schema", forwards, backwards)
}
