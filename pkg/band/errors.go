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

import errors "go.thethings.network/lorawan-stack/pkg/errorsv3"

var (
	errBandNotFound                         = errors.DefineNotFound("band_not_found", "band `{band_id}` not found")
	errUnsupportedLoRaWANRegionalParameters = errors.DefineInvalidArgument(
		"lorawan_version_unsupported",
		"LoRaWAN Regional Parameters version not supported; supported versions: {supported}",
	)
	errUnknownLoRaWANRegionalParameters = errors.DefineNotFound("unknown_lorawan_version", "unknown LoRaWAN version")
	errDataRateOffsetTooHigh            = errors.DefineInvalidArgument("data_rate_offset_too_high", "data rate offset must be lower or equal to {max}")
	errDataRateIndexTooHigh             = errors.DefineInvalidArgument("data_rate_index_too_high", "data rate index must be lower or equal to {max}")

	errUnsupportedChMaskCntl = errors.DefineInvalidArgument("chmaskcntl_unsupported", "ChMaskCntl `{chmaskcntl}` unsupported")
)
