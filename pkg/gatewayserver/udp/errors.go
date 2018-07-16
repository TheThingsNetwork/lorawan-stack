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

import "go.thethings.network/lorawan-stack/pkg/errorsv3"

var (
	errGatewayNotConnected = errors.DefineNotFound("gateway_not_connected", "gateway not connected")

	errTooSmallToHaveGatewayEUI    = errors.DefineInvalidArgument("no_gateway_eui", "packet is not long enough to contain the gateway EUI")
	errDecodingPayloadFromBase64   = errors.DefineInvalidArgument("decoding_payload_from_b64", "could not decode payload from base64")
	errParsingBandwidth            = errors.DefineInvalidArgument("parse_bandwidth", "could not parse bandwidth")
	errParsingSpreadingFactor      = errors.DefineInvalidArgument("parse_spreading_factor", "could not parse spreading factor")
	errUnmarshalPayloadFromLoRaWAN = errors.DefineInvalidArgument("unmarshal_payload_to_lorawan", "failed to unmarshal payload from LoRaWAN")
	errUnmarshalEUI                = errors.DefineInvalidArgument("unmarshal_eui", "failed to unmarshal an EUI")
	errUnmarshalTimestamp          = errors.DefineInvalidArgument("unmarshal_timestamp", "failed to unmarshal timestamp")
	errMarshalPayloadToLoRaWAN     = errors.DefineInvalidArgument("marshal_payload_to_lorawan", "failed to marshalling payload to LoRaWAN format")
	errMarshalPacketToUDP          = errors.DefineInvalidArgument("marshal_packet_to_udp_format", "failed to marshal packet to UDP format")
	errUnknownModulation           = errors.DefineInvalidArgument("unknown_modulation", "unknown modulation `{modulation}`")

	errNoConnectionAssociated = errors.DefineCorruption("no_connection_associated", "no gateway connection associated to this packet")
)
