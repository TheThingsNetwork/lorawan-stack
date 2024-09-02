// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package gateways

import (
	"regexp"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

const (
	formatIDttigpro1 = "ttigpro1"

	euiLength        = 16
	ownerTokenLength = 12
)

// TTIGPRO1 is a format for gateway identification QR codes.
type ttigpro1 struct {
	gatewayEUI types.EUI64
	ownerToken string
}

// ttigpro1Regex is the regular expression to match the TTIGPRO1 format.
// The format is as follows: https://ttig.pro/c/{16 lowercase base16 chars}/{12 base62 chars}
var ttigpro1Regex = regexp.MustCompile(`^https://ttig\.pro/c/([a-f0-9]{16})/([a-z0-9]{12})$`)

// UnmarshalText implements the TextUnmarshaler interface.
func (m *ttigpro1) UnmarshalText(text []byte) error {
	// Match the URL against the pattern
	matches := ttigpro1Regex.FindStringSubmatch(string(text))
	if matches == nil || len(matches) != 3 {
		return errInvalidFormat
	}

	if err := m.gatewayEUI.UnmarshalText([]byte(matches[1])); err != nil {
		return err
	}

	m.ownerToken = matches[2]

	if len(m.ownerToken) != ownerTokenLength {
		return errInvalidLength
	}

	return nil
}

// FormatID implements the Data interface.
func (m *ttigpro1) FormatID() string {
	return formatIDttigpro1
}

func (m *ttigpro1) GatewayEUI() string {
	return m.gatewayEUI.String()
}

func (m *ttigpro1) OwnerToken() string {
	return string(m.ownerToken)
}

// TTIGPRO1Format implements the TTIGPRO1 Format.
type TTIGPRO1Format struct{}

// Format implements the Format interface.
func (TTIGPRO1Format) Format() *ttnpb.QRCodeFormat {
	return &ttnpb.QRCodeFormat{
		Name:        "TTIGPRO1",
		Description: "TTI QR code format for gateway devices.",
		FieldMask: ttnpb.FieldMask(
			"ids.eui",
			"claim_authentication_code.secret.value",
		),
	}
}

// ID is the identifier of the format as a string.
func (TTIGPRO1Format) ID() string {
	return formatIDttigpro1
}

// New implements the Format interface.
func (TTIGPRO1Format) New() Data {
	return new(ttigpro1)
}
