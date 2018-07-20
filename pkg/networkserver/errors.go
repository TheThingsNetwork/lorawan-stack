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
	errCIDOutOfRange             = errors.DefineInvalidArgument("cid_out_of_range", "CID must be in range from {min} to {max}")
	errChannelIndexTooHigh       = errors.DefineInvalidArgument("channel_index_too_high", "channel index is too high")
	errComputeMIC                = errors.DefineInvalidArgument("compute_mic", "failed to compute MIC")
	errDecryptionFailed          = errors.DefineInvalidArgument("decryption", "decryption failed")
	errDeviceNotFound            = errors.DefineNotFound("device_not_found", "device not found")
	errDeviceRegistryInitialize  = errors.DefineInternal("device_registry_initialization", "Device Registry initialization failed")
	errDeviceStoreFailed         = errors.DefineInternal("device_store", "failed to store device")
	errDuplicateCIDHandler       = errors.DefineAlreadyExists("duplicate_cid_handler", "a handler for MAC command with CID {cid} is already registered")
	errDuplicateSubscription     = errors.DefineAlreadyExists("duplicate_subscription", "another subscription already started")
	errEmptySession              = errors.DefineFailedPrecondition("empty_session", "session in empty")
	errFCntTooHigh               = errors.DefineInvalidArgument("f_cnt_too_high", "FCnt is too high")
	errGatewayServerNotFound     = errors.DefineNotFound("gateway_server_not_found", "Gateway Server not found")
	errInvalidConfiguration      = errors.DefineInvalidArgument("configuration", "invalid configuration")
	errInvalidDataRate           = errors.DefineInvalidArgument("data_rate", "invalid data rate")
	errInvalidRx2DataRateIndex   = errors.DefineInvalidArgument("rx2_data_rate_index", "invalid Rx2 data rate index")
	errJoinFailed                = errors.Define("join", "all Join Servers failed to handle join")
	errLoRaAndFSK                = errors.DefineInvalidArgument("lora_and_fsk", "both LoRa and FSK modulation is specified")
	errMACEncodeFailed           = errors.DefineInternal("mac_encode", "failed to encode MAC commands")
	errMACEncryptFailed          = errors.DefineInternal("mac_encrypt", "failed to encrypt MAC commands")
	errMACRequestNotFound        = errors.DefineInvalidArgument("mac_request_not_found", "MAC response received, but corresponding request not found")
	errMarshalPayloadFailed      = errors.Define("marshal_payload", "failed to marshal payload")
	errMissingApplicationID      = errors.DefineNotFound("missing_application_id", "application ID is unknown")
	errMissingFNwkSIntKey        = errors.DefineNotFound("missing_f_nwk_s_int_key", "FNwkSIntKey is unknown")
	errMissingGatewayID          = errors.DefineNotFound("missing_gateway_id", "gateway ID is unknown")
	errMissingNwkSEncKey         = errors.DefineNotFound("missing_nwk_s_enc_key", "NwkSEncKey is unknown")
	errMissingPayload            = errors.DefineInvalidArgument("missing_payload", "message payload is missing")
	errMissingSNwkSIntKey        = errors.DefineNotFound("missing_s_nwk_s_int_key", "SNwkSIntKey is unknown")
	errRawPayloadTooLong         = errors.Define("raw_payload_too_long", "length of RawPayload must not be less than 4")
	errScheduleFailed            = errors.Define("schedule", "all Gateway Servers failed to schedule the downlink")
	errScheduleTooSoon           = errors.DefineUnavailable("schedule_too_soon", "confirmed downlink is scheduled too soon")
	errUnknownBand               = errors.Define("unknown_band", "band is unknown")
	errUnknownFrequencyPlan      = errors.Define("unknown_frequency_plan", "frequency plan is unknown")
	errUnknownMACState           = errors.DefineFailedPrecondition("unknown_mac_state", "MAC state is unknown")
	errUnmarshalPayloadFailed    = errors.DefineInvalidArgument("unmarshal_payload", "failed to unmarshal payload")
	errUnsupportedLoRaWANVersion = errors.DefineInvalidArgument("unsupported_lorawan_version", "unsupported LoRaWAN version: {version}", "version")
	errUplinkNotFound            = errors.DefineNotFound("uplink_not_found", "uplink not found")
)
