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
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator/qrcode"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

const (
	serialNumberAttribute = "serial-number"
	vendorIDAttribute     = "vendor-id"
	profileIDAttribute    = "profile-id"
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
	// Return the ID of this format as a string.
	ID() string
}

// Data represents end device QR code data.
type Data interface {
	qrcode.Data
	Encode(*ttnpb.EndDevice) error
}

// Server provides methods for end device onboarding.
type Server struct {
	endDeviceFormats   map[string]Format
	endDeviceFormatsMu sync.RWMutex
}

// New returns a new Server.
func New(ctx context.Context) *Server {
	return &Server{
		endDeviceFormats: make(map[string]Format),
	}
}

// GetEndDeviceFormats returns the registered end device QR code formats.
func (s *Server) GetEndDeviceFormats() map[string]Format {
	res := make(map[string]Format)
	s.endDeviceFormatsMu.RLock()
	for k, v := range s.endDeviceFormats {
		res[k] = v
	}
	s.endDeviceFormatsMu.RUnlock()
	return res
}

// GetEndDeviceFormat returns the converter by ID.
func (s *Server) GetEndDeviceFormat(id string) Format {
	s.endDeviceFormatsMu.RLock()
	res := s.endDeviceFormats[id]
	s.endDeviceFormatsMu.RUnlock()
	return res
}

// RegisterEndDeviceFormat registers the given end device QR code format.
// Existing registrations with the same ID will be overwritten.
func (s *Server) RegisterEndDeviceFormat(id string, f Format) {
	s.endDeviceFormatsMu.Lock()
	s.endDeviceFormats[id] = f
	s.endDeviceFormatsMu.Unlock()
}

var (
	errUnknownFormat = errors.DefineInvalidArgument("unknown_format", "format `{format_id}` unknown")
)

// Parse attempts to parse the given QR code data.
// It returns the parser and the format ID that successfully parsed the QR code.
func (s *Server) Parse(formatID string, data []byte) (Data, error) {
	for id, format := range s.endDeviceFormats {
		// If format ID is provided, use only that.
		if formatID != "" && formatID != id {
			continue
		}
		edFormat := format.New()
		if err := edFormat.UnmarshalText(data); err == nil {
			return edFormat, nil
		} else if formatID == id {
			return nil, err
		}
	}
	return nil, errUnknownFormat.WithAttributes("format_id", formatID)
}
