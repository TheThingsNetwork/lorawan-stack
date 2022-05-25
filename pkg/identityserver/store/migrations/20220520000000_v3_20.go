// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// If the accounts table does not exist, we're working on an empty database,
		// so we initialize it.
		if exists, err := tableExists(ctx, db, "accounts"); err != nil {
			return err
		} else if !exists {
			_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;")
			if err != nil {
				log.FromContext(ctx).
					WithError(err).
					Warn("Failed to enable pgcrypto extension, but trying to continue without it")
			}

			err = migrate.NewSQLMigrationFunc(sqlMigrations, "20220520000000_v3_20.init.sql")(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to run SQL migration: %w", err)
			}

			return nil
		}

		// If the notifications table does not exist, we're working with a pre-v3.19 database,
		// so we ask the operator to upgrade to v3.19 first.
		if exists, err := tableExists(ctx, db, "notifications"); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("database needs to be upgraded to v3.19 before upgrading to v3.20")
		}

		// v3.1
		//
		// - rename index uix_authorization_codes_code to authorization_code_code_index
		// - rename index uix_invitations_email to invitation_email_index
		// - rename index uix_invitations_token to invitation_token_index
		//
		// Instead of renaming the old index, the GORM migrator creates new indexes
		// and leaves the old ones, so we may still need to drop those.

		if _, err := db.NewDropIndex().
			IfExists().
			Index("uix_authorization_codes_code").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewDropIndex().
			IfExists().
			Index("uix_invitations_email").
			Exec(ctx); err != nil {
			return err
		}
		if _, err := db.NewDropIndex().
			IfExists().
			Index("uix_invitations_token").
			Exec(ctx); err != nil {
			return err
		}

		// v3.2
		//
		// (no schema changes)

		// v3.3
		//
		// - Add column picture_id to table end_devices (after service_profile_id)
		//   + Add index end_device_picture_index to table end_devices
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.4
		//
		// (no schema changes)

		// v3.5
		//
		// - Add column schedule_anytime_delay to table gateways (after enforce_duty_cycle)
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.6
		//
		// (no schema changes)

		// v3.7
		//
		// - Add column user_session_id to table access_tokens (after user_id)
		//   + Add index idx_access_tokens_user_session_id to table access_tokens
		// - Add column user_session_id to table authorization_codes (after user_id)
		//   + Add index idx_authorization_codes_user_session_id to table authorization_codes
		// - Add column update_location_from_status to table gateways
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.8
		//
		// - Add column logout_redirect_uris to table clients (after redirect_uris)
		// - Create table migrations
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.9
		//
		// - Add column session_secret to table user_sessions (after user_id)
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.10
		//
		// - Add column lbs_lns_secret to table gateways (after update_location_from_status)
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.11
		//
		// - Add column band_id to table end_devices (after firmware_version)
		// - Add columns claim_authentication_code_secret, claim_authentication_code_valid_from,
		//   claim_authentication_code_valid_to, target_cups_uri, target_cups_key to table gateways
		//   (after lbs_lns_secret)
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.12
		//
		// - Add column state_description to table clients (after state)
		// - Add column used to table contact_info_validations (after value)
		// - Add column require_authenticated_connection to table gateways (after target_cups_key)
		// - Create table login_tokens
		// - Add column state_description to table users (after state)
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.13
		//
		// - Add column expires_at to table api_keys (after entity_type)
		// - Add column dev_eui_counter to table applications (after description)
		// - Create table eui_blocks
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.14
		//
		// - Add column activated_at to table end_devices (after picture_id)
		// - Add column placement to table gateway_antennas (after accuracy)
		// - Add columns supports_lrfhss, disable_packet_broker_forwarding to table gateways
		//   (after require_authenticated_connection)
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.15
		//
		// (no schema changes)

		// v3.16
		//
		// (no schema changes)

		// v3.17
		//
		// - Add columns administrative_contact_id, technical_contact_id to table applications (after description)
		// - Add columns administrative_contact_id, technical_contact_id to table clients (after description)
		// - Add columns administrative_contact_id, technical_contact_id to table gateways (after description)
		// - Add columns administrative_contact_id, technical_contact_id to table organizations (after description)
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.18
		//
		// - Drop index eui_block_index

		if _, err := db.NewDropIndex().
			IfExists().
			Index("eui_block_index").
			Exec(ctx); err != nil {
			return err
		}

		// v3.19
		//
		// - Add columns network_server_address, application_server_address, join_server_address to table applications
		//   (after technical_contact_id)
		// - Add column last_seen_at to table end_devices (after activated_at)
		// - Create table notification_receivers
		// - Create table notifications
		//
		// New columns are added at the end of the table, but in order to re-order
		// them we'd have to recreate the table, which is currently not worth the effort.

		// v3.20
		//
		// Additional v3.20 migrations can be added in this folder.

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.NewDropTable().
			Table(
				"notification_receivers",
				"notifications",
				"api_keys",
				"access_tokens",
				"authorization_codes",
				"client_authorizations",
				"clients",
				"end_device_locations",
				"end_devices",
				"eui_blocks",
				"applications",
				"gateway_antennas",
				"gateways",
				"organizations",
				"memberships",
				"invitations",
				"login_tokens",
				"user_sessions",
				"users",
				"pictures",
				"contact_info_validations",
				"contact_infos",
				"attributes",
				"accounts",
				"migrations",
			).
			IfExists().
			Cascade().
			Exec(ctx)
		return err
	})
}
