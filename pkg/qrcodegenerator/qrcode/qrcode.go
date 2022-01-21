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
	"context"
	"encoding"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// QRCode represents oboarding data from QR Code.
type QRCode struct {
	endDeviceFormats   map[string]EndDeviceFormat
	endDeviceFormatsMu sync.RWMutex
}

// New returns a QRCode.
func New(ctx context.Context) *QRCode {
	return &QRCode{
		endDeviceFormats: make(map[string]EndDeviceFormat),
	}
}

// Data represents QR code data.
type Data interface {
	Validate() error
	GetOnboardingEntityData() *ttnpb.OnboardingEntityData
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// EndDeviceData represents end device QR code data.
type EndDeviceData interface {
	Data
	Encode(*ttnpb.EndDevice) error
}

var (
	errUnknownFormat     = errors.DefineInvalidArgument("unknown_format", "format `{format_id}` unknown")
	errUnsupportedEntity = errors.DefineInvalidArgument("unsupported_entity", "entity `{entity}` unsupported")
)

// Parse attempts to parse the given QR code data.
// It returns the parser and the format ID that successfully parsed the QR code.
func (c *QRCode) Parse(formatID string, entity ttnpb.OnboardingEntityType, data []byte) (Data, string, error) {
	switch entity {
	case ttnpb.OnboardingEntityType_END_DEVICE:
		for id, format := range c.endDeviceFormats {
			// If format ID is provided, use  only that.
			if formatID != "" && formatID != id {
				continue
			}
			edFormat := format.New()
			if err := edFormat.UnmarshalText(data); err == nil {
				return edFormat, id, nil
			}
		}
	default:
		return nil, "", errUnsupportedEntity.WithAttributes("entity", entity)
	}
	return nil, "", errUnknownFormat.WithAttributes("format_id", formatID)
}

// EndDeviceFormat is a end device QR code format.
type EndDeviceFormat interface {
	Format() *ttnpb.QRCodeFormat
	New() EndDeviceData
}

// GetEndDeviceFormats returns the registered end device QR code formats.
func (c *QRCode) GetEndDeviceFormats() map[string]EndDeviceFormat {
	res := make(map[string]EndDeviceFormat)
	c.endDeviceFormatsMu.RLock()
	for k, v := range c.endDeviceFormats {
		res[k] = v
	}
	c.endDeviceFormatsMu.RUnlock()
	return res
}

// GetEndDeviceFormat returns the converter by ID.
func (c *QRCode) GetEndDeviceFormat(id string) EndDeviceFormat {
	c.endDeviceFormatsMu.RLock()
	res := c.endDeviceFormats[id]
	c.endDeviceFormatsMu.RUnlock()
	return res
}

// RegisterEndDeviceFormat registers the given end device QR code format.
// Existing registrations with the same ID will be overwritten.
func (c *QRCode) RegisterEndDeviceFormat(id string, f EndDeviceFormat) {
	c.endDeviceFormatsMu.Lock()
	c.endDeviceFormats[id] = f
	c.endDeviceFormatsMu.Unlock()
}
