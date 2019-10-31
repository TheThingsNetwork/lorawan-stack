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
	"sync"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// Data represents QR code data.
type Data interface {
	Validate() error
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// EndDeviceData represents end device QR code data.
type EndDeviceData interface {
	Data
	Encode(*ttnpb.EndDevice) error
}

// AuthenticatedEndDeviceIdentifiers defines end device identifiers with authentication code.
type AuthenticatedEndDeviceIdentifiers interface {
	AuthenticatedEndDeviceIdentifiers() (joinEUI, devEUI types.EUI64, authenticationCode string)
}

var (
	errFormat    = errors.DefineInvalidArgument("format", "invalid format")
	errCharacter = errors.DefineInvalidArgument("character", "invalid character `{r}`")
	errNoJoinEUI = errors.DefineFailedPrecondition("no_join_eui", "no JoinEUI")
	errNoDevEUI  = errors.DefineFailedPrecondition("no_dev_eui", "no DevEUI")
)

// Parse attempts to parse the given QR code data.
func Parse(data []byte) (Data, error) {
	for _, model := range [...]Data{
		&LoRaAllianceTR005Draft3{},
		&LoRaAllianceTR005Draft2{},
	} {
		if err := model.UnmarshalText(data); err == nil {
			return model, nil
		}
	}
	return nil, errFormat
}

// EndDeviceFormat is a end device QR code format.
type EndDeviceFormat interface {
	Format() *ttnpb.QRCodeFormat
	New() EndDeviceData
}

var (
	endDeviceFormats   = map[string]EndDeviceFormat{}
	endDeviceFormatsMu sync.RWMutex
)

// GetEndDeviceFormats returns the registered end device QR code formats.
func GetEndDeviceFormats() map[string]EndDeviceFormat {
	res := make(map[string]EndDeviceFormat)
	endDeviceFormatsMu.RLock()
	for k, v := range endDeviceFormats {
		res[k] = v
	}
	endDeviceFormatsMu.RUnlock()
	return res
}

// GetEndDeviceFormat returns the converter by ID.
func GetEndDeviceFormat(id string) EndDeviceFormat {
	endDeviceFormatsMu.RLock()
	res := endDeviceFormats[id]
	endDeviceFormatsMu.RUnlock()
	return res
}

// RegisterEndDeviceFormat registers the given end device QR code format.
// Existing registrations with the same ID will be overwritten.
func RegisterEndDeviceFormat(id string, f EndDeviceFormat) {
	endDeviceFormatsMu.Lock()
	endDeviceFormats[id] = f
	endDeviceFormatsMu.Unlock()
}
