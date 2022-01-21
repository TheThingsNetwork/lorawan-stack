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

package joinserver

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

var (
	errComputeMIC                     = errors.DefineInvalidArgument("compute_mic", "failed to compute MIC")
	errDecodePayload                  = errors.DefineInvalidArgument("decode_payload", "failed to decode payload")
	errDeriveAppSKey                  = errors.Define("derive_app_s_key", "failed to derive application session key")
	errDeriveNwkSKeys                 = errors.Define("derive_nwk_s_keys", "failed to derive network session keys")
	errDeviceNotFound                 = errors.DefineNotFound("device_not_found", "device not found")
	errDevNonceTooSmall               = errors.DefineInvalidArgument("dev_nonce_too_small", "DevNonce is too small")
	errDevNonceLimitInvalid           = errors.DefineInvalidArgument("dev_nonce_limit_invalid", "DevNonce limit can not be less than 1")
	errDuplicateIdentifiers           = errors.DefineAlreadyExists("duplicate_identifiers", "a device identified by the identifiers already exists")
	errEncodePayload                  = errors.DefineInvalidArgument("encode_payload", "failed to encode payload")
	errEncryptPayload                 = errors.Define("encrypt_payload", "failed to encrypt JoinAccept")
	errGenerateSessionKeyID           = errors.Define("generate_session_key_id", "failed to generate session key ID")
	errJoinNonceTooHigh               = errors.Define("join_nonce_too_high", "JoinNonce is too high")
	errLookupNetID                    = errors.Define("lookup_net_id", "lookup NetID")
	errMICMismatch                    = errors.DefineInvalidArgument("mic_mismatch", "MIC mismatch")
	errNetIDMismatch                  = errors.DefineInvalidArgument("net_id_mismatch", "NetID `{net_id}` does not match")
	errNoAppKey                       = errors.DefineFailedPrecondition("no_app_key", "no AppKey specified")
	errNoApplicationServerID          = errors.DefineFailedPrecondition("no_application_server_id", "no AS-ID specified")
	errNoAppSKey                      = errors.DefineCorruption("no_app_s_key", "no AppSKey specified")
	errNoDevEUI                       = errors.DefineInvalidArgument("no_dev_eui", "no DevEUI specified")
	errNoFNwkSIntKey                  = errors.DefineCorruption("no_f_nwk_s_int_key", "no FNwkSIntKey specified")
	errNoJoinEUI                      = errors.DefineInvalidArgument("no_join_eui", "no JoinEUI specified")
	errNoJoinRequest                  = errors.DefineInvalidArgument("no_join_request", "no JoinRequest specified")
	errNoNetID                        = errors.DefineFailedPrecondition("no_net_id", "no NetID specified")
	errNoNwkKey                       = errors.DefineFailedPrecondition("no_nwk_key", "no NwkKey specified")
	errNoNwkSEncKey                   = errors.DefineCorruption("no_nwk_s_enc_key", "no NwkSEncKey specified")
	errNoSNwkSIntKey                  = errors.DefineCorruption("no_s_nwk_s_int_key", "no SNwkSIntKey specified")
	errProvisionerNotFound            = errors.DefineNotFound("provisioner_not_found", "provisioner `{id}` not found")
	errRegistryOperation              = errors.Define("registry_operation", "registry operation failed")
	errReuseDevNonce                  = errors.DefineInvalidArgument("reuse_dev_nonce", "DevNonce has already been used")
	errUnauthenticated                = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
	errUnknownJoinEUI                 = errors.DefineInvalidArgument("unknown_join_eui", "JoinEUI specified is not known")
	errUnsupportedLoRaWANMajorVersion = errors.DefineInvalidArgument("lorawan_major_version", "unsupported LoRaWAN major version: `{major}`")
	errUnsupportedMACVersion          = errors.DefineInvalidArgument("mac_version", "unsupported MAC version: `{version}`")
	errUnwrapKey                      = errors.Define("unwrap_key", "failed to unwrap key")
	errWrapKey                        = errors.Define("wrap_key", "failed to wrap key with KEK label `{label}`")
	errWrongPayloadType               = errors.DefineInvalidArgument("payload_type", "wrong payload type: {type}")
)
