// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	errBandNotFound                         = errors.DefineNotFound("band_not_found", "band `{id}` not found")
	errDataRateIndexTooHigh                 = errors.DefineInvalidArgument("data_rate_index_too_high", "data rate index must be lower or equal to {max}")
	errDataRateOffsetTooHigh                = errors.DefineInvalidArgument("data_rate_offset_too_high", "data rate offset must be lower or equal to {max}")
	errInvalidChannelCount                  = errors.DefineInvalidArgument("invalid_channel_count", "invalid number of channels defined")
	errUnknownPHYVersion                    = errors.DefineNotFound("unknown_phy_version", "unknown LoRaWAN PHY version `{version}`")
	errUnsupportedChMaskCntl                = errors.DefineInvalidArgument("chmaskcntl_unsupported", "ChMaskCntl `{chmaskcntl}` unsupported")
	errUnsupportedLoRaWANRegionalParameters = errors.DefineInvalidArgument("lorawan_version_unsupported", "LoRaWAN Regional Parameters version not supported; supported versions: {supported}")
)
