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
	errCorruptedMACState         = errors.DefineCorruption("corrupted_mac_state", "MAC state is corrupted")
	errDecodePayload             = errors.DefineInvalidArgument("decode_payload", "failed to decode payload")
	errDecrypt                   = errors.DefineInvalidArgument("decrypt", "failed to decrypt")
	errDeviceNotFound            = errors.DefineNotFound("device_not_found", "device not found")
	errDuplicateCIDHandler       = errors.DefineAlreadyExists("duplicate_cid_handler", "a handler for MAC command with CID {cid} is already registered")
	errDuplicateIdentifiers      = errors.DefineAlreadyExists("duplicate_identifiers", "a device identified by the identifiers already exists")
	errDuplicateSubscription     = errors.DefineAlreadyExists("duplicate_subscription", "another subscription already started")
	errEmptySession              = errors.DefineFailedPrecondition("empty_session", "session in empty")
	errEncodeMAC                 = errors.DefineInternal("encode_mac", "failed to encode MAC commands")
	errEncodePayload             = errors.Define("encode_payload", "failed to encode payload")
	errEncryptMAC                = errors.DefineInternal("encrypt_mac", "failed to encrypt MAC commands")
	errFCntTooHigh               = errors.DefineInvalidArgument("f_cnt_too_high", "FCnt is too high")
	errGatewayServerNotFound     = errors.DefineNotFound("gateway_server_not_found", "Gateway Server not found")
	errInvalidConfiguration      = errors.DefineInvalidArgument("configuration", "invalid configuration")
	errInvalidDataRate           = errors.DefineInvalidArgument("data_rate", "invalid data rate")
	errInvalidFrequencyPlan      = errors.DefineInvalidArgument("frequency_plan", "invalid frequency plan")
	errInvalidPayload            = errors.DefineInvalidArgument("payload", "invalid payload")
	errInvalidRx2DataRateIndex   = errors.DefineInvalidArgument("rx2_data_rate_index", "invalid Rx2 data rate index")
	errJoin                      = errors.Define("join", "all Join Servers failed to handle join")
	errLoRaAndFSK                = errors.DefineInvalidArgument("lora_and_fsk", "both LoRa and FSK modulation is specified")
	errMACRequestNotFound        = errors.DefineInvalidArgument("mac_request_not_found", "MAC response received, but corresponding request not found")
	errNoPayload                 = errors.DefineInvalidArgument("no_payload", "no message payload specified")
	errOutdatedData              = errors.DefineNotFound("outdated_data", "data is outdated")
	errRawPayloadTooLong         = errors.Define("raw_payload_too_long", "length of RawPayload must not be less than 4")
	errSchedule                  = errors.Define("schedule", "all Gateway Servers failed to schedule the downlink")
	errScheduleTooSoon           = errors.DefineUnavailable("schedule_too_soon", "confirmed downlink is scheduled too soon")
	errUnknownApplicationID      = errors.DefineNotFound("unknown_application_id", "application ID is unknown")
	errUnknownBand               = errors.Define("unknown_band", "band is unknown")
	errUnknownChannel            = errors.Define("unknown_chanel", "channel is unknown")
	errUnknownFNwkSIntKey        = errors.DefineNotFound("unknown_f_nwk_s_int_key", "FNwkSIntKey is unknown")
	errUnknownFrequencyPlan      = errors.Define("unknown_frequency_plan", "frequency plan is unknown")
	errUnknownMACState           = errors.DefineFailedPrecondition("unknown_mac_state", "MAC state is unknown")
	errUnknownNwkSEncKey         = errors.DefineNotFound("unknown_nwk_s_enc_key", "NwkSEncKey is unknown")
	errUnknownSNwkSIntKey        = errors.DefineNotFound("unknown_s_nwk_s_int_key", "SNwkSIntKey is unknown")
	errUnsupportedLoRaWANVersion = errors.DefineInvalidArgument("unsupported_lorawan_version", "unsupported LoRaWAN version: {version}", "version")
	errUplinkNotFound            = errors.DefineNotFound("uplink_not_found", "uplink not found")
)
