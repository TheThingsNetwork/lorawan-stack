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

package encoding

import "go.thethings.network/lorawan-stack/pkg/errorsv3"

var (
	errNoGatewayEUI         = errors.DefineInvalidArgument("no_gateway_eui", "packet is not long enough to contain the gateway EUI")
	errDecodePayload        = errors.DefineInvalidArgument("decode_payload", "could not decode binary payload")
	errParseBandwidth       = errors.DefineInvalidArgument("parse_bandwidth", "could not parse bandwidth")
	errParseSpreadingFactor = errors.DefineInvalidArgument("parse_spreading_factor", "could not parse spreading factor")
	errUnmarshalEUI         = errors.DefineInvalidArgument("unmarshal_eui", "failed to unmarshal EUI")
	errUnmarshalTimestamp   = errors.DefineInvalidArgument("unmarshal_timestamp", "failed to unmarshal timestamp")
	errUnknownModulation    = errors.DefineInvalidArgument("unknown_modulation", "unknown modulation `{modulation}`")
)
