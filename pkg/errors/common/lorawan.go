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

package common

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	// ErrUnsupportedLoRaWANVersion is returned by operations which failed because of an unsupported LoRaWAN version.
	ErrUnsupportedLoRaWANVersion = &errors.ErrDescriptor{
		MessageFormat: "Unsupport LoRaWAN version: {version}",
		Type:          errors.InvalidArgument,
		Code:          6,
	}
	// ErrUnsupportedLoRaWANMACVersion is returned by operations which failed
	// because of an unsupported LoRaWAN MAC version.
	ErrUnsupportedLoRaWANMACVersion = &errors.ErrDescriptor{
		MessageFormat: "Unsupported LoRaWAN MAC version: {version}",
		Type:          errors.InvalidArgument,
		Code:          7,
	}
	// ErrComputeMIC represents error occurring when computation of the MIC fails.
	ErrComputeMIC = &errors.ErrDescriptor{
		MessageFormat: "Failed to compute MIC",
		Type:          errors.InvalidArgument,
		Code:          8,
	}
	// ErrMissingPayload represents the error occurring when the message payload is missing.
	ErrMissingPayload = &errors.ErrDescriptor{
		MessageFormat: "Message payload is missing",
		Type:          errors.InvalidArgument,
		Code:          9,
	}
	// ErrInvalidModulation is returned if the passed modulation is invalid.
	ErrInvalidModulation = &errors.ErrDescriptor{
		MessageFormat: "Invalid modulation",
		Type:          errors.InvalidArgument,
		Code:          10,
	}
	// ErrMissingDevAddr represents an error occurring when the DevAddr is missing.
	ErrMissingDevAddr = &errors.ErrDescriptor{
		MessageFormat: "DevAddr is missing",
		Type:          errors.InvalidArgument,
		Code:          11,
	}
	// ErrMissingDevEUI represents an error occurring when the DevEUI is missing.
	ErrMissingDevEUI = &errors.ErrDescriptor{
		MessageFormat: "DevEUI is missing",
		Type:          errors.InvalidArgument,
		Code:          12,
	}
	// ErrMissingJoinEUI represents an error occurring when the JoinEUI is missing.
	ErrMissingJoinEUI = &errors.ErrDescriptor{
		MessageFormat: "JoinEUI is missing",
		Type:          errors.InvalidArgument,
		Code:          13,
	}
	// ErrFCntTooLow represents an error occurring when FCnt is too low.
	ErrFCntTooLow = &errors.ErrDescriptor{
		MessageFormat: "FCnt is too low",
		Type:          errors.InvalidArgument,
		Code:          14,
	}
	// ErrFCntTooHigh represents an error occurring when FCnt is too high.
	ErrFCntTooHigh = &errors.ErrDescriptor{
		MessageFormat: "FCnt is too high",
		Type:          errors.InvalidArgument,
		Code:          15,
	}
)

func init() {
	ErrUnsupportedLoRaWANVersion.Register()
	ErrUnsupportedLoRaWANMACVersion.Register()
	ErrComputeMIC.Register()
	ErrMissingPayload.Register()
	ErrInvalidModulation.Register()
	ErrMissingDevEUI.Register()
	ErrMissingJoinEUI.Register()
	ErrMissingDevAddr.Register()
	ErrFCntTooLow.Register()
	ErrFCntTooHigh.Register()
}
