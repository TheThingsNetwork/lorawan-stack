// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"database/sql/driver"
	"encoding/json"
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

// SetAntennas sets the radios.
func (g *Gateway) SetRadios(radios []GatewayRadio) {
	g.Radios = radios
}

// SetAntennas sets the API key.
func (g *Gateway) SetAPIKey(key *APIKey) {
	g.APIKey = *key
}

var (
	// FieldPathGatewayDescription is the field path for the gateway description field.
	FieldPathGatewayDescription = regexp.MustCompile(`^description$`)

	// FieldPathGatewayAPIKey is the field path for the gateway API Key that is used
	// in the device configuration.
	FieldPathGatewayAPIKey = regexp.MustCompile(`^api_key$`)

	// FieldPathGatewayFrequencyPlanID is the field path for the gateway frequency plan ID field.
	FieldPathGatewayFrequencyPlanID = regexp.MustCompile(`^frequency_plan_id$`)

	// FieldPathGatewayClusterAddress is the field path for the gateway cluster address field.
	FieldPathGatewayClusterAddress = regexp.MustCompile(`^cluster_address$`)

	// FieldPathGatewayAntennas is the field path for the antennas.
	FieldPathGatewayAntennas = regexp.MustCompile(`^antennas$`)

	// FieldPathGatewayRadios is the field path for the radios.
	FieldPathGatewayRadios = regexp.MustCompile(`^radios$`)

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

	// FieldPathGatewayAttributes is the field path for an attribute in the attributes map.
	FieldPathGatewayAttributes = regexp.MustCompile(`^attributes\.(.+)$`)

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

// Value implements driver.Valuer interface.
func (p GatewayRadio_TxConfiguration) Value() (driver.Value, error) {
	b, err := json.Marshal([]uint32{p.MinFrequency, p.MaxFrequency, p.NotchFrequency})
	if err != nil {
		return nil, err
	}

	return string(b[:]), nil
}

// Scan implements sql.Scanner interface.
func (p *GatewayRadio_TxConfiguration) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return errors.Errorf("Invalid type assertion. Got %T instead of string", src)
	}

	var values []uint32
	if err := json.Unmarshal([]byte(str), &values); err != nil {
		return err
	}

	p.MinFrequency = values[0]
	p.MaxFrequency = values[1]
	p.NotchFrequency = values[2]

	return nil
}
