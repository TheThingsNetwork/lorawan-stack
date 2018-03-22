// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "github.com/TheThingsNetwork/ttn/pkg/errors"

var (
	// ErrBandNotFound describes the errors returned when looking for an unknown band.
	ErrBandNotFound = &errors.ErrDescriptor{
		MessageFormat:  "Band `{band}` not found",
		Type:           errors.NotFound,
		Code:           1,
		SafeAttributes: []string{"band"},
	}
	// ErrUnsupportedLoRaWANVersion is returned if an operation could not be completed because of an invalid LoRaWAN version.
	ErrUnsupportedLoRaWANVersion = &errors.ErrDescriptor{
		MessageFormat: "LoRaWAN version not supported",
		Code:          2,
		Type:          errors.InvalidArgument,
	}
	// ErrUnknownLoRaWANVersion is returned if the LoRaWAN version is not recognized.
	ErrUnknownLoRaWANVersion = &errors.ErrDescriptor{
		MessageFormat: "Unknown LoRaWAN version",
		Code:          3,
		Type:          errors.Unknown,
	}
)

func init() {
	ErrBandNotFound.Register()
	ErrUnsupportedLoRaWANVersion.Register()
	ErrUnknownLoRaWANVersion.Register()
}
