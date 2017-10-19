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
	PathGatewayDescription = "description"

	PathGatewayFrequencyPlanID = "frequency_plan_id"

	PathGatewayPrivacySettingStatusPublic = "privacy_settings.status_public"

	PathGatewayPrivacySettingLocationPublic = "privacy_settings.location_public"

	PathGatewayPrivacySettingsContactable = "privacy_settings.contactable"

	PathGatewayAutoUpdate = "auto_update"

	PathGatewayPlatform = "platform"

	PathGatewayAntennas = "antennas"

	PathGatewayAttributes = "attributes"

	PathGatewayClusterAddress = "cluster_address"

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
