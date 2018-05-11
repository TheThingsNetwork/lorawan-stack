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

package ttnpb

import (
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
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

// SetRadios sets the radios.
func (g *Gateway) SetRadios(radios []GatewayRadio) {
	g.Radios = radios
}

var (
	// FieldPathGatewayDescription is the field path for the gateway description field.
	FieldPathGatewayDescription = regexp.MustCompile(`^description$`)

	// FieldPathGatewayFrequencyPlanID is the field path for the gateway frequency plan ID field.
	FieldPathGatewayFrequencyPlanID = regexp.MustCompile(`^frequency_plan_id$`)

	// FieldPathGatewayClusterAddress is the field path for the gateway cluster address field.
	FieldPathGatewayClusterAddress = regexp.MustCompile(`^cluster_address$`)

	// FieldPathGatewayAntennas is the field path for the antennas.
	FieldPathGatewayAntennas = regexp.MustCompile(`^antennas$`)

	// FieldPathGatewayRadios is the field path for the radios.
	FieldPathGatewayRadios = regexp.MustCompile(`^radios$`)

	// FieldPathGatewayPrivacySettingsStatusPublic is the field path for the gateway privacy setting status public field.
	FieldPathGatewayPrivacySettingsStatusPublic = regexp.MustCompile(`^privacy_settings.status_public$`)

	// FieldPathGatewayPrivacySettingsLocationPublic is the field path for the gateway privacy setting location public field.
	FieldPathGatewayPrivacySettingsLocationPublic = regexp.MustCompile(`^privacy_settings.location_public$`)

	// FieldPathGatewayPrivacySettingsContactable is the field path for the gateway privacy setting contactable field.
	FieldPathGatewayPrivacySettingsContactable = regexp.MustCompile(`^privacy_settings.contactable$`)

	// FieldPathGatewayAutoUpdate is the field path for the gateway auto update field.
	FieldPathGatewayAutoUpdate = regexp.MustCompile(`^auto_update$`)

	// FieldPathGatewayPlatform is the field path for the gateway platform field.
	FieldPathGatewayPlatform = regexp.MustCompile(`^platform$`)

	// FieldPathGatewayAttributes is the field path for an attribute in the attributes map.
	FieldPathGatewayAttributes = regexp.MustCompile(`^attributes\.(.+)$`)

	// FieldPathGatewayContactAccount is the field path for the gateway contact account identifiers field.
	FieldPathGatewayContactAccountIDs = regexp.MustCompile(`^contact_account_ids$`)

	// FieldPathGatewayDisableTxDelay is the field path for the gateway disable Tx delay field.
	FieldPathGatewayDisableTxDelay = regexp.MustCompile(`^disable_tx_delay$`)
)

// gatewayPrivacySetting is an enum that defines the different gateway privacy settings.
type gatewayPrivacySetting int32

const (
	privacySettingLocationPublic gatewayPrivacySetting = iota
	privacySettingStatusPublic
	privacySettingContactable
)

// Value implements driver.Valuer interface.
func (p GatewayPrivacySettings) Value() (driver.Value, error) {
	settings := make([]string, 0)

	if p.LocationPublic {
		settings = append(settings, string(privacySettingLocationPublic))
	}

	if p.StatusPublic {
		settings = append(settings, string(privacySettingStatusPublic))
	}

	if p.Contactable {
		settings = append(settings, string(privacySettingContactable))
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
		case string(privacySettingLocationPublic):
			p.LocationPublic = true
		case string(privacySettingStatusPublic):
			p.StatusPublic = true
		case string(privacySettingContactable):
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
