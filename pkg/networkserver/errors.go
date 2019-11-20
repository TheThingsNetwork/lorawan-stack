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

package networkserver

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errABPJoinRequest             = errors.DefineInvalidArgument("abp_join_request", "received a join-request from ABP device")
	errClassAMulticast            = errors.DefineInvalidArgument("class_a_multicast", "multicast device in class A mode")
	errClassBCForClassA           = errors.DefineInvalidArgument("class_b_c_for_class_a", "class B/C downlink queued for device in class A mode")
	errComputeMIC                 = errors.DefineInvalidArgument("compute_mic", "failed to compute MIC")
	errConfirmedDownlinkTooSoon   = errors.DefineUnavailable("confirmed_too_soon", "confirmed downlink is scheduled too soon")
	errConfirmedMulticastDownlink = errors.DefineInvalidArgument("confirmed_multicast_downlink", "confirmed downlink queued for multicast device")
	errCorruptedMACState          = errors.DefineCorruption("corrupted_mac_state", "MAC state is corrupted")
	errDataRateNotFound           = errors.DefineNotFound("data_rate_not_found", "data rate not found")
	errDecodePayload              = errors.DefineInvalidArgument("decode_payload", "failed to decode payload")
	errDeviceNotFound             = errors.DefineNotFound("device_not_found", "device not found")
	errEmptySession               = errors.DefineFailedPrecondition("empty_session", "session in empty")
	errEncodeMAC                  = errors.DefineInternal("encode_mac", "failed to encode MAC commands")
	errEncodePayload              = errors.Define("encode_payload", "failed to encode payload")
	errEncryptMAC                 = errors.DefineInternal("encrypt_mac", "failed to encrypt MAC commands")
	errExpiredDownlink            = errors.DefineFailedPrecondition("downlink_expired", "queued downlink is expired")
	errFCntTooLow                 = errors.DefineInvalidArgument("f_cnt_too_low", "FCnt is too low")
	errInvalidAbsoluteTime        = errors.DefineInvalidArgument("absolute_time", "invalid absolute time set in application downlink")
	errInvalidChannelIndex        = errors.DefineInvalidArgument("channel_index", "invalid channel index")
	errInvalidConfiguration       = errors.DefineInvalidArgument("configuration", "invalid configuration")
	errInvalidDataRate            = errors.DefineInvalidArgument("data_rate", "invalid data rate")
	errInvalidFieldMask           = errors.DefineInvalidArgument("field_mask", "invalid field mask")
	errInvalidFieldValue          = errors.DefineInvalidArgument("field_value", "invalid value of field `{field}`")
	errInvalidFixedPaths          = errors.DefineInvalidArgument("fixed_paths", "invalid fixed paths set in application downlink")
	errInvalidPayload             = errors.DefineInvalidArgument("payload", "invalid payload")
	errJoinServerNotFound         = errors.DefineNotFound("join_server_not_found", "Join Server not found")
	errMACRequestNotFound         = errors.DefineInvalidArgument("mac_request_not_found", "MAC response received, but corresponding request not found")
	errNoDevEUI                   = errors.DefineInvalidArgument("no_dev_eui", "no DevEUI specified")
	errNoJoinEUI                  = errors.DefineInvalidArgument("no_join_eui", "no JoinEUI specified")
	errNoPath                     = errors.DefineNotFound("no_downlink_path", "no downlink path available")
	errNoPayload                  = errors.DefineInvalidArgument("no_payload", "no message payload specified")
	errOutdatedData               = errors.DefineNotFound("outdated_data", "data is outdated")
	errRawPayloadTooShort         = errors.Define("raw_payload_too_short", "length of RawPayload must not be less than 4")
	errSchedule                   = errors.Define("schedule", "all downlink scheduling attempts failed")
	errUnknownChannel             = errors.Define("unknown_chanel", "channel is unknown")
	errUnknownMACState            = errors.DefineFailedPrecondition("unknown_mac_state", "MAC state is unknown")
	errUnknownNwkSEncKey          = errors.DefineNotFound("unknown_nwk_s_enc_key", "NwkSEncKey is unknown")
	errUnknownSession             = errors.DefineNotFound("unknown_session", "unknown session")
	errUnknownSNwkSIntKey         = errors.DefineNotFound("unknown_s_nwk_s_int_key", "SNwkSIntKey is unknown")
	errUnsupportedLoRaWANVersion  = errors.DefineInvalidArgument("unsupported_lorawan_version", "unsupported LoRaWAN version: {version}", "version")
	errUplinkChannelNotFound      = errors.DefineNotFound("uplink_channel_not_found", "uplink channel not found")
)
