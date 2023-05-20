// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package enddevices

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	errCharacter = errors.DefineInvalidArgument("character", "invalid character `{r}`")
	errNoJoinEUI = errors.DefineFailedPrecondition("no_join_eui", "no JoinEUI")
	errNoDevEUI  = errors.DefineFailedPrecondition("no_dev_eui", "no DevEUI")
	errFormat    = errors.DefineInvalidArgument("format", "invalid format")
)

// Format is a end device QR code format.
type Format interface {
	Format() *ttnpb.QRCodeFormat
	New() Data
}

// Data represents end device QR code data.
type Data interface {
	qrcode.Data
	Encode(*ttnpb.EndDevice) error
	// FormatID returns the ID of the format used to parse the QR Code data.
	FormatID() string
	// EndDeviceTemplate returns the End Device Template corresponding to the QR code data.
	EndDeviceTemplate() *ttnpb.EndDeviceTemplate
}

type endDeviceFormat struct {
	id     string
	format Format
}

// Server provides methods for end device QR codes.
type Server struct {
	endDeviceFormats []endDeviceFormat
}

// New returns a new Server.
func New(ctx context.Context) *Server {
	s := &Server{
		// Newer formats should be added to this slice first to preferentially match with those first.
		endDeviceFormats: []endDeviceFormat{
			{
				id:     formatIDLoRaAllianceTR005,
				format: new(LoRaAllianceTR005Format),
			},
			{
				id:     formatIDLoRaAllianceTR005Draft2,
				format: new(LoRaAllianceTR005Draft2Format),
			},
			{
				id:     formatIDLoRaAllianceTR005Draft3,
				format: new(LoRaAllianceTR005Draft3Format),
			},
			{
				id:     formatIDDevEUI,
				format: new(devEUIFormat),
			},
		},
	}
	return s
}

// GetEndDeviceFormats returns the registered end device QR code formats.
func (s *Server) GetEndDeviceFormats() map[string]Format {
	ret := make(map[string]Format)
	for _, edFormat := range s.endDeviceFormats {
		ret[edFormat.id] = edFormat.format
	}
	return ret
}

// GetEndDeviceFormat returns the converter by ID.
func (s *Server) GetEndDeviceFormat(id string) Format {
	for _, edFormat := range s.endDeviceFormats {
		if edFormat.id == id {
			return edFormat.format
		}
	}
	return nil
}

// RegisterEndDeviceFormat registers the given end device QR code format.
// Existing registrations with the same ID will not be overwritten.
// While matching, the slice will be traversed in a FIFO manner.
func (s *Server) RegisterEndDeviceFormat(id string, f Format) {
	s.endDeviceFormats = append(s.endDeviceFormats, endDeviceFormat{
		id:     id,
		format: f,
	})
}

var errUnknownFormat = errors.DefineInvalidArgument("unknown_format", "format unknown")

// Parse attempts to parse the given QR code data.
// It returns the parser and the format ID that successfully parsed the QR code.
func (s *Server) Parse(formatID string, data []byte) (ret Data, err error) {
	for _, edFormat := range s.endDeviceFormats {
		// If format ID is provided, use only that.
		if formatID != "" && formatID != edFormat.id {
			continue
		}
		f := edFormat.format.New()
		if err := f.UnmarshalText(data); err == nil {
			return f, nil
		} else if formatID == edFormat.id {
			return nil, err
		}
	}
	return nil, errUnknownFormat.New()
}
