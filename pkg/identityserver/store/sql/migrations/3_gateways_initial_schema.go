// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS gateways (
			gateway_id          STRING(36) PRIMARY KEY,
			description         STRING NOT NULL DEFAULT '',
			frequency_plan_id   STRING(36) NOT NULL,
			activated_at        TIMESTAMP DEFAULT NULL,
			privacy_settings    STRING NOT NULL DEFAULT '',
			auto_update         BOOL DEFAULT TRUE,
			platform            STRING NOT NULL DEFAULT '',
			cluster_address     STRING NOT NULL,
			created_at          TIMESTAMP NOT NULL DEFAULT current_timestamp(),
			updated_at          TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS gateways_attributes (
			gateway_id   STRING(36) NOT NULL REFERENCES gateways(gateway_id),
			attribute    STRING NOT NULL,
			value        STRING NOT NULL,
			PRIMARY KEY(gateway_id, attribute)
		);
		CREATE TABLE IF NOT EXISTS gateways_antennas (
			antenna_id   STRING DEFAULT to_hex(unique_rowid()) PRIMARY KEY,
			gateway_id   STRING(36) NOT NULL REFERENCES gateways(gateway_id),
			gain         FLOAT NOT NULL DEFAULT 0.0,
			type         STRING NOT NULL DEFAULT '',
			model        STRING NOT NULL DEFAULT '',
			placement    STRING NOT NULL DEFAULT '',
			longitude    FLOAT NOT NULL DEFAULT 0,
			latitude     FLOAT NOT NULL DEFAULT 0,
			altitude     INT NOT NULL DEFAULT 0,
			created_at   TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS gateways_radios (
			radio_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			gateway_id         STRING(36) NOT NULL REFERENCES gateways(gateway_id),
			frequency          INT NOT NULL CHECK (frequency > 0),
			tx_configuration   STRING DEFAULT NULL,
			created_at         TIMESTAMP NOT NULL DEFAULT current_timestamp()
		);
		CREATE TABLE IF NOT EXISTS gateways_collaborators (
			gateway_id   STRING(36) REFERENCES gateways(gateway_id),
			account_id   STRING(36) REFERENCES accounts(account_id),
			"right"      STRING NOT NULL,
			PRIMARY KEY(gateway_id, account_id, "right")
		);
		CREATE TABLE IF NOT EXISTS gateways_api_keys (
			key          STRING NOT NULL PRIMARY KEY,
			gateway_id   STRING(36) NOT NULL REFERENCES gateways(gateway_id),
			key_name     STRING(36) NOT NULL,
			UNIQUE(gateway_id, key_name)
		);
		CREATE TABLE IF NOT EXISTS gateways_api_keys_rights (
			gateway_id   STRING(36) NOT NULL REFERENCES gateways(gateway_id),
			key_name     STRING(36) NOT NULL,
			"right"      STRING NOT NULL,
			PRIMARY KEY(gateway_id, key_name, "right")
		);
	`

	const backwards = `
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
