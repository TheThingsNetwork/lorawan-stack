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
	// ErrUnsupportedLoRaWANRegionalParameters is returned if an operation could not be completed because of an unsupported LoRaWAN regional parameters version.
	ErrUnsupportedLoRaWANRegionalParameters = &errors.ErrDescriptor{
		MessageFormat: "LoRaWAN version not supported",
		Code:          2,
		Type:          errors.InvalidArgument,
	}
	// ErrUnknownLoRaWANRegionalParameters is returned if the LoRaWAN regional parameters version is not recognized.
	ErrUnknownLoRaWANRegionalParameters = &errors.ErrDescriptor{
		MessageFormat: "Unknown LoRaWAN version",
		Code:          3,
		Type:          errors.Unknown,
	}
)

func init() {
	ErrBandNotFound.Register()
	ErrUnsupportedLoRaWANRegionalParameters.Register()
	ErrUnknownLoRaWANRegionalParameters.Register()
}
