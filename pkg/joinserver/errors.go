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

package joinserver

import errors "go.thethings.network/lorawan-stack/pkg/errorsv3"

var (
	errAddressMismatch           = errors.DefinePermissionDenied("address_mismatch", "{component} address mismatch")
	errDevNonceReused            = errors.DefineInvalidArgument("dev_nonce_reused", "DevNonce has already been used")
	errDevNonceTooHigh           = errors.DefineInvalidArgument("dev_nonce_too_high", "DevNonce is too high")
	errDevNonceTooSmall          = errors.DefineInvalidArgument("dev_nonce_too_small", "DevNonce is too small")
	errEncryptPayloadFailed      = errors.Define("encrypt_payload", "failed to encrypt JoinAccept")
	errForwardJoinRequest        = errors.Define("forward_join_request", "failed to forward JoinRequest")
	errMACVersionMismatch        = errors.DefineInternal("mac_version_mismatch", "Device MAC version mismatch, in registry: {registered}, selected: {selected}")
	errMICCheckFailed            = errors.Define("mic_check", "MIC check failed")
	errMICComputeFailed          = errors.DefineInvalidArgument("mic_compute", "failed to compute MIC")
	errMICMismatch               = errors.DefineInvalidArgument("mic_mismatch", "MIC mismatch")
	errMarshalPayloadFailed      = errors.DefineInvalidArgument("marshal_payload", "failed to marshal payload")
	errMissingAppKey             = errors.DefineCorruption("missing_app_key", "AppKey is missing")
	errMissingAppSKey            = errors.DefineCorruption("missing_app_s_key", "AppSKey is missing")
	errMissingDevAddr            = errors.DefineCorruption("missing_dev_addr", "DevAddr is missing")
	errMissingDevEUI             = errors.DefineCorruption("missing_dev_eui", "DevEUI is missing")
	errMissingFNwkSIntKey        = errors.DefineCorruption("missing_f_nwk_s_int_key", "FNwkSIntKey is missing")
	errMissingJoinEUI            = errors.DefineCorruption("missing_join_eui", "JoinEUI is missing")
	errMissingJoinRequest        = errors.DefineCorruption("missing_join_request", "JoinRequest is missing")
	errMissingNwkKey             = errors.DefineCorruption("missing_nwk_key", "NwkKey is missing")
	errMissingNwkSEncKey         = errors.DefineCorruption("missing_nwk_s_enc_key", "NwkSEncKey is missing")
	errMissingPayload            = errors.DefineCorruption("missing_payload", "message payload is missing")
	errMissingSNwkSIntKey        = errors.DefineCorruption("missing_s_nwk_s_int_key", "SNwkSIntKey is missing")
	errMissingSessionKeyID       = errors.DefineCorruption("missing_session_key_id", "SessionKeyID is missing")
	errNoSession                 = errors.DefineCorruption("missing_session", "Session is missing")
	errPayloadLengthMismatch     = errors.DefineInvalidArgument("payload_length", "expected length of payload to be equal to 23 got {length}")
	errSessionKeyIDMismatch      = errors.DefineInvalidArgument("session_key_id_mismatch", "SessionKeyID mismatch")
	errUnknownAppEUI             = errors.Define("unknown_app_eui", "AppEUI specified is not known")
	errUnmarshalPayloadFailed    = errors.DefineInvalidArgument("unmarshal_payload", "failed to unmarshal payload")
	errUnsupportedLoRaWANVersion = errors.DefineInvalidArgument("lorawan_version", "unsupported LoRaWAN version: {version}", "version")
	errWrongPayloadType          = errors.DefineInvalidArgument("payload_type", "wrong payload type: {type}")
)
