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

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	errCheckMIC                       = errors.Define("check_mic", "MIC check failed")
	errComputeMIC                     = errors.DefineInvalidArgument("compute_mic", "failed to compute MIC")
	errDecodePayload                  = errors.DefineInvalidArgument("decode_payload", "failed to decode payload")
	errDeriveAppSKey                  = errors.Define("derive_app_s_key", "failed to derive application session key")
	errDeriveNwkSKeys                 = errors.Define("derive_nwk_s_keys", "failed to derive network session keys")
	errDevNonceTooHigh                = errors.DefineInvalidArgument("dev_nonce_too_high", "DevNonce is too high")
	errDevNonceTooSmall               = errors.DefineInvalidArgument("dev_nonce_too_small", "DevNonce is too small")
	errDuplicateIdentifiers           = errors.DefineAlreadyExists("duplicate_identifiers", "a device identified by the identifiers already exists")
	errEncodePayload                  = errors.DefineInvalidArgument("encode_payload", "failed to encode payload")
	errEncryptPayload                 = errors.Define("encrypt_payload", "failed to encrypt JoinAccept")
	errEndDeviceRequest               = errors.DefineInvalidArgument("end_device_request", "GetEndDeviceRequest is invalid")
	errForwardJoinRequest             = errors.Define("forward_join_request", "failed to forward JoinRequest")
	errGenerateSessionKeyID           = errors.Define("generate_session_key_id", "failed to generate session key ID")
	errDeviceNotFound                 = errors.DefineNotFound("device_not_found", "device not found")
	errInvalidIdentifiers             = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errJoinNonceTooHigh               = errors.Define("join_nonce_too_high", "JoinNonce is too high")
	errMICMismatch                    = errors.DefineInvalidArgument("mic_mismatch", "MIC mismatch")
	errNetIDMismatch                  = errors.DefineInvalidArgument("net_id_mismatch", "NetID `{net_id}` does not match")
	errNoAppKey                       = errors.DefineCorruption("no_app_key", "no AppKey specified")
	errNoAppSKey                      = errors.DefineCorruption("no_app_s_key", "no AppSKey specified")
	errNoDevAddr                      = errors.DefineCorruption("no_dev_addr", "no DevAddr specified")
	errNoDevEUI                       = errors.DefineInvalidArgument("no_dev_eui", "no DevEUI specified")
	errNoFNwkSIntKey                  = errors.DefineCorruption("no_f_nwk_s_int_key", "no FNwkSIntKey specified")
	errNoJoinEUI                      = errors.DefineInvalidArgument("no_join_eui", "no JoinEUI specified")
	errNoJoinRequest                  = errors.DefineInvalidArgument("no_join_request", "no JoinRequest specified")
	errNoNwkKey                       = errors.DefineCorruption("no_nwk_key", "no NwkKey specified")
	errNoNwkSEncKey                   = errors.DefineCorruption("no_nwk_s_enc_key", "no NwkSEncKey specified")
	errNoNetID                        = errors.DefineFailedPrecondition("no_net_id", "no NetID specified")
	errNoApplicationServerID          = errors.DefineFailedPrecondition("no_application_server_id", "no AS-ID specified")
	errNoPayload                      = errors.DefineInvalidArgument("no_payload", "no message payload specified")
	errNoRootKeys                     = errors.DefineCorruption("no_root_keys", "no root keys specified")
	errNoSNwkSIntKey                  = errors.DefineCorruption("no_s_nwk_s_int_key", "no SNwkSIntKey specified")
	errPayloadLengthMismatch          = errors.DefineInvalidArgument("payload_length", "expected length of payload to be equal to 23 got {length}")
	errProvisionerNotFound            = errors.DefineNotFound("provisioner_not_found", "provisioner `{id}` not found")
	errProvisionerDecode              = errors.Define("provisioner_decode", "failed to decode provisioning data")
	errProvisionEntryCount            = errors.DefineInvalidArgument("provision_entry_count", "expected `{expected}` but have `{actual}` entries to provision")
	errProvisioning                   = errors.DefineAborted("provisioning", "provisioning failed")
	errRegistryOperation              = errors.DefineInternal("registry_operation", "registry operation failed")
	errReuseDevNonce                  = errors.DefineInvalidArgument("reuse_dev_nonce", "DevNonce has already been used")
	errUnauthenticated                = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
	errCallerNotAuthorized            = errors.DefinePermissionDenied("caller_not_authorized", "caller `{name}` is not authorized for the entity")
	errUnknownAppEUI                  = errors.Define("unknown_app_eui", "AppEUI specified is not known")
	errUnsupportedLoRaWANMajorVersion = errors.DefineInvalidArgument("lorawan_major_version", "unsupported LoRaWAN major version: `{major}`")
	errUnsupportedMACVersion          = errors.DefineInvalidArgument("mac_version", "unsupported MAC version: `{version}`")
	errWrapKey                        = errors.Define("wrap_key", "failed to wrap key")
	errWrongPayloadType               = errors.DefineInvalidArgument("payload_type", "wrong payload type: {type}")
)
