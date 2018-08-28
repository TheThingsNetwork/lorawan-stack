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

import (
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var (
	errNoEUI           = errors.DefineInvalidArgument("no_eui", "packet is not long enough to contain the EUI")
	errPayload         = errors.DefineInvalidArgument("payload", "failed to parse binary payload")
	errBandwidth       = errors.DefineInvalidArgument("bandwidth", "failed to parse bandwidth")
	errSpreadingFactor = errors.DefineInvalidArgument("spreading_factor", "failed to parse spreading factor")
	errEUI             = errors.DefineInvalidArgument("eui", "failed to parse EUI")
	errTimestamp       = errors.DefineInvalidArgument("timestamp", "failed to parse timestamp")
	errModulation      = errors.DefineInvalidArgument("modulation", "invalid modulation `{modulation}`")
)
