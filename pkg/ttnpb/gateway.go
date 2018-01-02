// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"database/sql/driver"
	"regexp"
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

var (
	// FieldPathGatewayDescription is the field path for the gateway description field.
	FieldPathGatewayDescription = regexp.MustCompile(`^description$`)

	// FieldPathGatewayFrequencyPlanID is the field path for the gateway frequency plan ID field.
	FieldPathGatewayFrequencyPlanID = regexp.MustCompile(`^frequency_plan_id$`)

	// FieldPathGatewaySettingsStatusPublic is the field path for the gateway privacy setting status public field.
	FieldPathGatewayPrivacySettingsStatusPublic = regexp.MustCompile(`^privacy_settings.status_public$`)

	// FieldPathGatewaySettingsLocationPublic is the field path for the gateway privacy setting location public field.
	FieldPathGatewayPrivacySettingsLocationPublic = regexp.MustCompile(`^privacy_settings.location_public$`)

	// FieldPathGatewaySettingsContactable is the field path for the gateway privacy setting contactable field.
	FieldPathGatewayPrivacySettingsContactable = regexp.MustCompile(`^privacy_settings.contactable$`)

	// FieldPathGatewayAutoUpdate is the field path for the gateway auto update field.
	FieldPathGatewayAutoUpdate = regexp.MustCompile(`^auto_update$`)

	// FieldPathGatewayPlatform is the field path for the gateway platform field.
	FieldPathGatewayPlatform = regexp.MustCompile(`^platform$`)

	// FieldPathGatewayAntennaGain is the field path for the gain field of an indexed antenna.
	FieldPathGatewayAntennaGain = regexp.MustCompile(`^antennas\.(\d).gain$`)

	// FieldPathGatewayAntennaLocationLatitude is the field path for the latitude field of an indexed antenna.
	FieldPathGatewayAntennaLocationLatitude = regexp.MustCompile(`^antennas\.(\d).latitude$`)

	// FieldPathGatewayAntennaLocationLongitude is the field path for the longitude field of an indexed antenna.
	FieldPathGatewayAntennaLocationLongitude = regexp.MustCompile(`^antennas\.(\d).longitude$`)

	// FieldPathGatewayAntennaLocationAltitude is the field path for the altitude field of an indexed antenna.
	FieldPathGatewayAntennaLocationAltitude = regexp.MustCompile(`^antennas\.(\d).altitude$`)

	// FieldPathGatewayAntennaLocationAccuracy is the field path for the accuracy field of an indexed antenna.
	FieldPathGatewayAntennaLocationAccuracy = regexp.MustCompile(`^antennas\.(\d).accuracy$`)

	// FieldPathGatewayAntennaLocationSource is the field path for the source field of an indexed antenna.
	FieldPathGatewayAntennaLocationSource = regexp.MustCompile(`^antennas\.(\d).source$`)

	// FieldPathGatewayAntennaType is the field path for the type field of an indexed antenna.
	FieldPathGatewayAntennaType = regexp.MustCompile(`^antennas\.(\d).type$`)

	// FieldPathGatewayAntennaModel is the field path for the model field of an indexed antenna.
	FieldPathGatewayAntennaModel = regexp.MustCompile(`^antennas\.(\d).model$`)

	// FieldPathGatewayAntennaPlacement is the field path for the placement field of an indexed antenna.
	FieldPathGatewayAntennaPlacement = regexp.MustCompile(`^antennas\.(\d).placement$`)

	// FieldPathGatewayAttributes is the field path for an attribute in the attributes map.
	FieldPathGatewayAttributes = regexp.MustCompile(`^attributes\.(.*)$`)

	// FieldPathGatewayClusterAddress is the field path for the gateway cluster address field.
	FieldPathGatewayClusterAddress = regexp.MustCompile(`^cluster_address$`)

	// FieldPathGatewayContactAccountUserID is the field path for the gateway contact account user ID field.
	FieldPathGatewayContactAccountUserID = regexp.MustCompile(`^contact_account.user_id$`)
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
