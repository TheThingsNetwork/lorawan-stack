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

	// ErrLoRaWANParametersInvalid is returned if the LoRaWAN parameters specified are not valid or are not compatible with the band.
	ErrLoRaWANParametersInvalid = &errors.ErrDescriptor{
		MessageFormat: "Invalid LoRaWAN parameters",
		Code:          4,
		Type:          errors.InvalidArgument,
	}
)

func init() {
	ErrBandNotFound.Register()
	ErrUnsupportedLoRaWANRegionalParameters.Register()
	ErrUnknownLoRaWANRegionalParameters.Register()
}
