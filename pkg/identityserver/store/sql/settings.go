// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
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
		return nil, ErrSettingsNotFound.New(nil)
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
func (s *SettingStore) Set(settings *ttnpb.IdentityServerSettings) error {
	return s.set(s.queryer(), settings)
}

func (s *SettingStore) set(q db.QueryContext, settings *ttnpb.IdentityServerSettings) error {
	var input struct {
		*ttnpb.IdentityServerSettings
		BlacklistedIDsConverted db.StringSlice
		AllowedEmailsConverted  db.StringSlice
	}
	input.IdentityServerSettings = settings
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
				current_timestamp(),
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
