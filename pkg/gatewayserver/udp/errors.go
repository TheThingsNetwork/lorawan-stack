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

package udp

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	// ErrTooSmallToHaveGatewayEUI is returned if a message isn't long enough
	// to contain a Gateway EUI.
	ErrTooSmallToHaveGatewayEUI = &errors.ErrDescriptor{
		MessageFormat: "Packet is not long enough to contain the gateway EUI",
		Code:          1,
		Type:          errors.InvalidArgument,
	}
	// ErrUnmarshalFailed is returned if a value could not be unmarshalled.
	ErrUnmarshalFailed = &errors.ErrDescriptor{
		MessageFormat: "Could not unmarshal value",
		Code:          2,
		Type:          errors.InvalidArgument,
	}
	// ErrMarshalFailed is returned if a value could not be marshalled.
	ErrMarshalFailed = &errors.ErrDescriptor{
		MessageFormat: "Could not marshal value",
		Code:          3,
		Type:          errors.InvalidArgument,
	}
	// ErrDecodingPayloadFromBase64 is returned if the value of a base64 payload could not be decoded.
	ErrDecodingPayloadFromBase64 = &errors.ErrDescriptor{
		MessageFormat: "Could not decode payload from base64",
		Code:          4,
		Type:          errors.InvalidArgument,
	}
	// ErrParsingBandwidth is returned if the bandwidth could not be parsed from a message.
	ErrParsingBandwidth = &errors.ErrDescriptor{
		MessageFormat: "Could not parse bandwidth",
		Code:          5,
		Type:          errors.InvalidArgument,
	}
	// ErrParsingSpreadingFactor is returned if the spreading factor could not be parsed from a message.
	ErrParsingSpreadingFactor = &errors.ErrDescriptor{
		MessageFormat: "Could not parse spreading factor",
		Code:          6,
		Type:          errors.InvalidArgument,
	}
	// ErrUnknownModulation is returned if the modulation of a packet is unknown.
	ErrUnknownModulation = &errors.ErrDescriptor{
		MessageFormat: "Unknown modulation",
		Code:          7,
		Type:          errors.InvalidArgument,
	}
)

func init() {
	ErrTooSmallToHaveGatewayEUI.Register()
	ErrUnmarshalFailed.Register()
	ErrMarshalFailed.Register()
	ErrDecodingPayloadFromBase64.Register()
	ErrParsingBandwidth.Register()
	ErrParsingSpreadingFactor.Register()
	ErrUnknownModulation.Register()
}
