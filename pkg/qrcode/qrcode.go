// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// Package qrcode implements working with QR codes.
package qrcode

import (
	"encoding"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// Data represents QR code data.
type Data interface {
	Validate() error
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// AuthenticatedEndDeviceIdentifiers defines end device identifiers with authentication code.
type AuthenticatedEndDeviceIdentifiers interface {
	AuthenticatedEndDeviceIdentifiers() (joinEUI, devEUI types.EUI64, authenticationCode string)
}

var (
	errFormat    = errors.DefineInvalidArgument("format", "invalid format")
	errCharacter = errors.DefineInvalidArgument("character", "invalid character `{r}`")
)

// Parse attempts to parse the given QR code data.
func Parse(data []byte) (Data, error) {
	for _, model := range [...]Data{
		&LoRaAllianceTR005Draft2{},
	} {
		if err := model.UnmarshalText(data); err == nil {
			return model, nil
		}
	}
	return nil, errFormat
}
