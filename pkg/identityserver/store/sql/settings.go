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

package sql

import (
	"go.thethings.network/lorawan-stack/pkg/identityserver/db"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// SettingStore implements store.SettingStore.
type SettingStore struct {
	storer
}

// NewSettingStore returns a settings store.
func NewSettingStore(store storer) *SettingStore {
	return &SettingStore{
		storer: store,
	}
}

// Get returns the settings.
func (s *SettingStore) Get() (*ttnpb.IdentityServerSettings, error) {
	return s.get(s.queryer())
}

func (s *SettingStore) get(q db.QueryContext) (*ttnpb.IdentityServerSettings, error) {
	var res struct {
		*ttnpb.IdentityServerSettings
		AllowedEmailsConverted  db.StringSlice
		BlacklistedIDsConverted db.StringSlice
	}
	err := q.SelectOne(
		&res,
		`SELECT
				blacklisted_ids AS blacklisted_ids_converted,
				allowed_emails AS allowed_emails_converted,
				validation_token_ttl,
				invitation_token_ttl,
				skip_validation,
				invitation_only,
				admin_approval,
				updated_at
			FROM settings
			WHERE id = 1`)
	if db.IsNoRows(err) {
		return nil, store.ErrSettingsNotFound.New(nil)
	}
	if err != nil {
		return nil, err
	}
	settings := new(ttnpb.IdentityServerSettings)
	settings = res.IdentityServerSettings
	res.AllowedEmailsConverted.SetInto(&settings.AllowedEmails)
	res.BlacklistedIDsConverted.SetInto(&settings.BlacklistedIDs)
	return settings, nil
}

// Set sets the settings.
func (s *SettingStore) Set(settings ttnpb.IdentityServerSettings) error {
	return s.set(s.queryer(), settings)
}

func (s *SettingStore) set(q db.QueryContext, settings ttnpb.IdentityServerSettings) error {
	var input struct {
		*ttnpb.IdentityServerSettings
		BlacklistedIDsConverted db.StringSlice
		AllowedEmailsConverted  db.StringSlice
	}
	input.IdentityServerSettings = &settings
	input.BlacklistedIDsConverted = db.StringSlice(settings.BlacklistedIDs)
	input.AllowedEmailsConverted = db.StringSlice(settings.AllowedEmails)
	_, err := q.NamedExec(
		`UPSERT
			INTO settings (
				id,
				updated_at,
				blacklisted_ids,
				skip_validation,
				invitation_only,
				admin_approval,
				validation_token_ttl,
				invitation_token_ttl,
				allowed_emails)
			VALUES (
				1,
				:updated_at,
				:blacklisted_ids_converted,
				:skip_validation,
				:invitation_only,
				:admin_approval,
				:validation_token_ttl,
				:invitation_token_ttl,
				:allowed_emails_converted)
		`,
		input)
	return err
}
