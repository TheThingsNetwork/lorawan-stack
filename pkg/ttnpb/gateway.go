// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"database/sql/driver"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// GetGateway returns the base Gateway itself.
func (g *Gateway) GetGateway() *Gateway {
	return g
}

// SetAttributes sets the free-form attributes.
func (g *Gateway) SetAttributes(attributes map[string]string) {
	g.Attributes = attributes
}

// SetAntennas sets the antennas.
func (g *Gateway) SetAntennas(antennas []GatewayAntenna) {
	g.Antennas = antennas
}

const (
	// These are the valid FieldMask path values for the `update_mask` in
	// the UpdateGatewayRequest message.

	// FieldPathGatewayDescription is the path value for the `description` field.
	FieldPathGatewayDescription = "description"

	// FieldPathGatewayFrequencyPlanID is the path value for the `frequency_plan_id` field.
	FieldPathGatewayFrequencyPlanID = "frequency_plan_id"

	// FieldPathGatewayPrivacySettingsStatusPublic is the path value for the
	// `status_public` field in the privacy settings.
	FieldPathGatewayPrivacySettingsStatusPublic = "privacy_settings.status_public"

	// FieldPathGatewayPrivacySettingsLocationPublic is the path value for the
	// `location_public` field in the privacy settings.
	FieldPathGatewayPrivacySettingsLocationPublic = "privacy_settings.location_public"

	// FieldPathGatewayPrivacySettingsContactable is the path value for the
	// `contactable` field in the privacy settings.
	FieldPathGatewayPrivacySettingsContactable = "privacy_settings.contactable"

	// FieldPathGatewayAutoUpdate is the path value for the `auto_update` field.
	FieldPathGatewayAutoUpdate = "auto_update"

	// FieldPathGatewayPlatform is the path value for the `platform` field.
	FieldPathGatewayPlatform = "platform"

	// FieldPathGatewayAntennas is the path value for the `antennas` field.
	// This field path affect to all the antennas inside the slice.
	FieldPathGatewayAntennas = "antennas"

	// FieldPathGatewayAttributes is the path value for the `attributes` field.
	FieldPathGatewayAttributes = "attributes"

	// FieldPathGatewayClusterAddress is the path value for the `cluster_address` field.
	FieldPathGatewayClusterAddress = "cluster_address"

	// FieldPathGatewayContactAccount is the path value for the `user_id` field
	// that denotes the contact account of the gateway.
	FieldPathGatewayContactAccount = "contact_account.user_id"
)

// GatewayPrivacySetting is an enum that defines the different gateway privacy settings.
type GatewayPrivacySetting int32

const (
	PrivacySettingLocationPublic GatewayPrivacySetting = iota
	PrivacySettingStatusPublic
	PrivacySettingContactable
)

// Value implements driver.Valuer interface.
func (p GatewayPrivacySettings) Value() (driver.Value, error) {
	settings := make([]string, 0)

	if p.LocationPublic {
		settings = append(settings, string(PrivacySettingLocationPublic))
	}

	if p.StatusPublic {
		settings = append(settings, string(PrivacySettingStatusPublic))
	}

	if p.Contactable {
		settings = append(settings, string(PrivacySettingContactable))
	}

	return strings.Join(settings, ","), nil
}

// Scan implements sql.Scanner interface.
func (p *GatewayPrivacySettings) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return errors.Errorf("Invalid type assertion. Got %T instead of string", src)
	}

	for _, part := range strings.Split(str, ",") {
		switch part {
		case string(PrivacySettingLocationPublic):
			p.LocationPublic = true
		case string(PrivacySettingStatusPublic):
			p.StatusPublic = true
		case string(PrivacySettingContactable):
			p.Contactable = true
		}
	}

	return nil
}
