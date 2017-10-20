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
	// Valid FieldMask path values for the `update_mask` in UpdateGatewayRequest message.

	// PathGatewayDescription is the path value for the description field.
	PathGatewayDescription = "description"

	// PathGatewayFrequencyPlanID is the path value for the frequency plan ID field.
	PathGatewayFrequencyPlanID = "frequency_plan_id"

	// PathGatewayPrivacySettingsStatusPublic is the path value for the privacy setting
	// that denotes if the status is public or not.
	PathGatewayPrivacySettingsStatusPublic = "privacy_settings.status_public"

	// PathGatewayPrivacySettingsLocationPublic is the path value for the privacy setting
	// that denotes if the gateway location is public or not.
	PathGatewayPrivacySettingsLocationPublic = "privacy_settings.location_public"

	// PathGatewayPrivacySettingsContactable is the path value for the privacy setting
	// that denotes if the contact account information is public or not.
	PathGatewayPrivacySettingsContactable = "privacy_settings.contactable"

	// PathGatewayAutoUpdate is the path value for the auto update field.
	PathGatewayAutoUpdate = "auto_update"

	// PathGatewayPlatform is the path value for the gateway platform.
	PathGatewayPlatform = "platform"

	// PathGatewayAntennas is the path value for the gateway antennas.
	PathGatewayAntennas = "antennas"

	// PathGatewayAttributes is the path value for the attributes map.
	PathGatewayAttributes = "attributes"

	// PathGatewayClusterAddress is the path value for the cluster address field.
	PathGatewayClusterAddress = "cluster_address"

	// PathGatewayContactAccount is the path value for the User ID that reflects
	// the contact account of the gateway.
	PathGatewayContactAccount = "contact_account.user_id"
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
