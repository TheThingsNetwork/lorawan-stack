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

// Server provides methods for end device QR codes.
type Server struct {
	endDeviceFormats sync.Map
}

// New returns a new Server.
func New(ctx context.Context) *Server {
	return &Server{}
}

// GetEndDeviceFormats returns the registered end device QR code formats.
func (s *Server) GetEndDeviceFormats() map[string]Format {
	res := make(map[string]Format)
	s.endDeviceFormats.Range(func(key, value interface{}) bool {
		res[key.(string)] = value.(Format)
		return true
	})
	return res
}

// GetEndDeviceFormat returns the converter by ID.
func (s *Server) GetEndDeviceFormat(id string) Format {
	res, ok := s.endDeviceFormats.Load(id)
	if !ok {
		return nil
	}
	return res.(Format)
}

// RegisterEndDeviceFormat registers the given end device QR code format.
// Existing registrations with the same ID will be overwritten.
func (s *Server) RegisterEndDeviceFormat(id string, f Format) {
	s.endDeviceFormats.Store(id, f)
}

var (
	errUnknownFormat = errors.DefineInvalidArgument("unknown_format", "format `{format_id}` unknown")
)

// Parse attempts to parse the given QR code data.
// It returns the parser and the format ID that successfully parsed the QR code.
func (s *Server) Parse(formatID string, data []byte) (ret Data, err error) {
	s.endDeviceFormats.Range(func(key, value interface{}) bool {
		id := key.(string)
		// If format ID is provided, use only that.
		if formatID != "" && formatID != id {
			return true
		}
		f := value.(Format).New()
		if err = f.UnmarshalText(data); err == nil {
			ret = f
			return false
		} else if formatID == id {
			// Return the unmarshaling error since this was the requested format.
			return false
		}
		return true
	})
	if ret == nil && err == nil {
		return nil, errUnknownFormat.WithAttributes("format_id", formatID)
	}
	return
}
