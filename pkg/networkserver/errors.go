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

package networkserver

import (
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var (
	errUnknownFrequencyPlan      = errors.Define("unknown_frequency_plan", "frequency plan is unknown")
	errUnknownBand               = errors.Define("unknown_band", "band is unknown")
	errDeviceNotFound            = errors.DefineNotFound("device_not_found", "device not found")
	errMissingFNwkSIntKey        = errors.DefineNotFound("missing_f_nwk_s_int_key", "FNwkSIntKey is unknown")
	errMissingSNwkSIntKey        = errors.DefineNotFound("missing_s_nwk_s_int_key", "SNwkSIntKey is unknown")
	errMissingNwkSEncKey         = errors.DefineNotFound("missing_nwk_s_enc_key", "NwkSEncKey is unknown")
	errMissingApplicationID      = errors.DefineNotFound("missing_application_id", "application ID is unknown")
	errMissingGatewayID          = errors.DefineNotFound("missing_gateway_id", "gateway ID is unknown")
	errUplinkNotFound            = errors.DefineNotFound("uplink_not_found", "uplink not found")
	errGatewayServerNotFound     = errors.DefineNotFound("gateway_server_not_found", "Gateway Server not found")
	errUnknownMACState           = errors.DefineFailedPrecondition("unknown_mac_state", "MAC state is unknown")
	errDuplicateSubscription     = errors.DefineAlreadyExists("duplicate_subscription", "another subscription already started")
	errInvalidConfiguration      = errors.DefineInvalidArgument("invalid_configuration", "invalid configuration")
	errChannelIndexTooHigh       = errors.DefineInvalidArgument("channel_index_too_high", "channel index is too high")
	errDecryptionFailed          = errors.DefineInvalidArgument("decryption_failed", "decryption failed")
	errMACRequestNotFound        = errors.DefineInvalidArgument("mac_request_not_found", "MAC response received, but corresponding request not found")
	errInvalidDataRate           = errors.DefineInvalidArgument("invalid_data_rate", "invalid data rate")
	errUnsupportedLoRaWANVersion = errors.DefineInvalidArgument("unsupported_lorawan_version", "unsupported LoRaWAN version: {version}", "version")
	errComputeMIC                = errors.DefineInvalidArgument("compute_mic_failed", "failed to compute MIC")
	errMissingPayload            = errors.DefineInvalidArgument("missing_payload", "message payload is missing")
	errMarshalPayloadFailed      = errors.Define("marshal_payload_failed", "failed to marshal payload")
	errUnmarshalPayloadFailed    = errors.DefineInvalidArgument("unmarshal_payload_failed", "failed to unmarshal payload")
	errFCntTooHigh               = errors.DefineInvalidArgument("f_cnt_too_high", "FCnt is too high")
	errScheduleTooSoon           = errors.DefineUnavailable("schedule_too_soon", "confirmed downlink is scheduled too soon")
)
