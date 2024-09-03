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

// Package gateways provides a QR code parser for gateways.
package gateways

import (
	"context"
	"encoding"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var (
	errUnknownFormat = errors.DefineInvalidArgument("unknown_format", "format unknown")
	errInvalidLength = errors.DefineInvalidArgument("invalid_length", "invalid length")
	errInvalidFormat = errors.DefineInvalidArgument("invalid_format", "invalid format")
)

// Format is a gateway QR code format.
type Format interface {
	Format() *ttnpb.QRCodeFormat
	New() Data
}

// Data represents gateway QR code data.
type Data interface {
	// FormatID returns the ID of the format used to parse the QR Code data.
	FormatID() string
	GatewayEUI() types.EUI64
	OwnerToken() string
	encoding.TextUnmarshaler
}

type gatewayFormat struct {
	id     string
	format Format
}

// Server provides methods for gateways QR codes.
type Server struct {
	gatewayFormats []gatewayFormat
}

// New returns a new Server.
func New(_ context.Context) *Server {
	s := &Server{
		// Newer formats should be added to this slice first to
		// preferentially match with those first.
		gatewayFormats: []gatewayFormat{
			{
				id:     formatIDTTIGPRO1,
				format: new(TTIGPRO1Format),
			},
		},
	}
	return s
}

// RegisterGatewayFormat registers the given gateway QR code format.
func (s *Server) RegisterGatewayFormat(id string, f Format) {
	s.gatewayFormats = append(s.gatewayFormats, gatewayFormat{
		id:     id,
		format: f,
	})
}

// GetGatewayFormats returns the registered gateway QR code formats.
func (s *Server) GetGatewayFormats() map[string]Format {
	ret := make(map[string]Format)
	for _, gtwFormat := range s.gatewayFormats {
		ret[gtwFormat.id] = gtwFormat.format
	}
	return ret
}

// GetGatewayFormat returns the format by ID.
func (s *Server) GetGatewayFormat(id string) Format {
	for _, gtwFormat := range s.gatewayFormats {
		if gtwFormat.id == id {
			return gtwFormat.format
		}
	}
	return nil
}

// Formats returns the registered gateway QR code formats.
func (s *Server) Formats() []*ttnpb.QRCodeFormat {
	formats := make([]*ttnpb.QRCodeFormat, 0, len(s.gatewayFormats))
	for _, gtwFormat := range s.gatewayFormats {
		formats = append(formats, gtwFormat.format.Format())
	}
	return formats
}

// Parse the given QR code data. If formatID is provided, only that format is used.
// Otherwise, the first format registered will be used.
func (s *Server) Parse(formatID string, data []byte) (ret Data, err error) {
	for _, gtwFormat := range s.gatewayFormats {
		// If format ID is provided, use only that. Otherwise,
		// default to the first format listed in gatewayFormats.
		if formatID != "" && formatID != gtwFormat.id {
			continue
		}

		f := gtwFormat.format.New()
		if err := f.UnmarshalText(data); err != nil {
			return nil, err
		}

		return f, nil
	}

	return nil, errUnknownFormat
}
