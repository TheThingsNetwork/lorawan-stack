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
	// errDeviceNotFound represents error ocurring when device is not found.
	errDeviceNotFound = errors.DefineNotFound("device", "device not found")

	// errMissingFNwkSIntKey represents error ocurring when FNwkSIntKey is missing.
	errMissingFNwkSIntKey = errors.DefineNotFound("missing_f_nwk_s_int_key", "FNwkSIntKey is unknown")

	// errMissingSNwkSIntKey represents error ocurring when SNwkSIntKey is missing.
	errMissingSNwkSIntKey = errors.DefineNotFound("missing_s_nwk_s_int_key", "SNwkSIntKey is unknown")

	// errMissingNwkSEncKey represents error ocurring when NwkSEncKey is missing.
	errMissingNwkSEncKey = errors.DefineNotFound("missing_nwk_s_enc_key", "NwkSEncKey is unknown")

	// errMissingApplicationID represents error ocurring when ApplicationID is missing.
	errMissingApplicationID = errors.DefineNotFound("missing_application_id", "application ID is unknown")

	// errMissingGatewayID represents error ocurring when GatewayID is missing.
	errMissingGatewayID = errors.DefineNotFound("missing_gateway_id", "gateway ID is unknown")

	// errDuplicateSubscription represents error ocurring when a duplicate subscription is opened.
	errDuplicateSubscription = errors.DefineAlreadyExists("duplicate_subscription", "another subscription already started")

	// errInvalidConfiguration represents error ocurring when the configuration is invalid.
	errInvalidConfiguration = errors.DefineInvalidArgument("invalid_configuration", "invalid configuration")

	// errUplinkNotFound represents error ocurring when there were no uplinks found.
	errUplinkNotFound = errors.DefineNotFound("uplink_not_found", "uplink not found")

	// errGatewayServerNotFound represents error ocurring when there were no uplinks found.
	errGatewayServerNotFound = errors.DefineNotFound("gateway_server_not_found", "gateway server not found")

	// errChannelIndexTooHigh represents error ocurring when the channel index is too high.
	errChannelIndexTooHigh = errors.DefineInvalidArgument("channel_index_too_high", "channel index is too high")

	// errDecryptionFailed represents error ocurring when the decryption fails.
	errDecryptionFailed = errors.DefineInvalidArgument("decryption_failed", "decryption failed")

	// errMACRequestNotFound represents error ocurring when the a response to a MAC response
	// is received, but a corresponding request is not found.
	errMACRequestNotFound = errors.DefineInvalidArgument("mac_request_not_found", "MAC response received, but corresponding request not found")

	// errInvalidDataRate represents error ocurring when the data rate is invalid.
	errInvalidDataRate = errors.DefineInvalidArgument("invalid_data_rate", "invalid data rate")

	// errScheduleTooSoon represents error ocurring when a confirmed downlink is scheduled too soon.
	errScheduleTooSoon = errors.DefineUnavailable("schedule_too_soon", "confirmed downlink is scheduled too soon")

	// errCorruptRegistry represents error occurring when the registry of a component is corrupted.
	errCorruptRegistry = errors.DefineCorruption("corrupt_registry", "registry is corrupt")

	// errUnsupportedLoRaWANVersion is returned by operations which failed because of an unsupported LoRaWAN version.
	errUnsupportedLoRaWANVersion = errors.DefineInvalidArgument("unsupported_lorawan_version", "unsupported LoRaWAN version: {version}", "version")

	// errComputeMIC represents error occurring when computation of the MIC fails.
	errComputeMIC = errors.DefineInvalidArgument("compute_mic", "failed to compute MIC")

	// errMissingPayload represents the error occurring when the message payload is missing.
	errMissingPayload = errors.DefineInvalidArgument("missing_payload", "message payload is missing")

	// errInvalidArgument is returned if the arguments passed to a function are invalid.
	errInvalidArgument = errors.DefineInvalidArgument("invalid_argument", "Invalid arguments")

	// errCheckFailed is returned if the arguments didn't pass a specifically-defined
	// argument check.
	errCheckFailed = errors.DefineInvalidArgument("CheckFailed", "Arguments check failed")

	// errUnmarshalPayloadFailed is returned when a payload couldn't be unmarshalled.
	errUnmarshalPayloadFailed = errors.DefineInvalidArgument("unmarshal_payload_failed", "Failed to unmarshal payload")

	// errMarshalPayloadFailed is returned when a payload couldn't be marshalled.
	errMarshalPayloadFailed = errors.DefineInvalidArgument("marshal_payload_failed", "Failed to marshal payload")

	// errPermissionDenied is returned when a request is not allowed to access a protected resource.
	errPermissionDenied = errors.DefinePermissionDenied("permission_denied", "Permission denied to perform this operation")

	// errInvalidModulation is returned if the passed modulation is invalid.
	errInvalidModulation = errors.DefineInvalidArgument("invalid_modulation", "Invalid modulation")

	// errMissingDevAddr represents an error occurring when the DevAddr is missing.
	errMissingDevAddr = errors.DefineInvalidArgument("missing_dev_addr", "DevAddr is missing")

	// errMissingDevEUI represents an error occurring when the DevEUI is missing.
	errMissingDevEUI = errors.DefineInvalidArgument("missing_dev_eui", "DevEUI is missing")

	// errMissingJoinEUI represents an error occurring when the JoinEUI is missing.
	errMissingJoinEUI = errors.DefineInvalidArgument("missing_join_eui", "JoinEUI is missing")

	// errFCntTooLow represents an error occurring when FCnt is too low.
	errFCntTooLow = errors.DefineInvalidArgument("f_cnt_too_low", "FCnt is too low")

	// errFCntTooHigh represents an error occurring when FCnt is too high.
	errFCntTooHigh = errors.DefineInvalidArgument("f_cnt_too_high", "FCnt is too high")
)
