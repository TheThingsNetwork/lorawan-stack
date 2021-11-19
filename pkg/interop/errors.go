// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package interop

import (
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var (
	errNoPublicTLSAddress  = errors.DefineFailedPrecondition("no_public_tls_address", "no public TLS address configured for interop")
	errUnknownMACVersion   = errors.DefineInvalidArgument("unknown_mac_version", "unknown MAC version")
	errInvalidLength       = errors.DefineInvalidArgument("invalid_length", "invalid length")
	errInvalidRequestType  = errors.DefineInvalidArgument("invalid_request_type", "invalid request type `{type}`")
	errNotRegistered       = errors.DefineNotFound("not_registered", "not registered")
	errUnexpectedResult    = errors.Define("unexpected_result", "unexpected result code {code}", "code")
	errCallerNotAuthorized = errors.DefinePermissionDenied("caller_not_authorized", "caller is not authorized for `{target}`")
	errUnauthenticated     = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
	errInvalidVendorID     = errors.DefineInvalidArgument("invalid_vendor_id", "invalid vendor ID")

	ErrNoAction           = errors.DefineAborted("no_action", "no action")
	ErrMIC                = errors.DefineCorruption("mic", "MIC failed")
	ErrFrameReplayed      = errors.DefineAborted("frame_replayed", "frame replayed")
	ErrJoinReq            = errors.DefineAborted("join_req", "join-request failed")
	ErrNoRoamingAgreement = errors.DefineFailedPrecondition("no_roaming_agreement", "no roaming agreement")
	ErrDeviceRoaming      = errors.DefineFailedPrecondition("device_roaming", "device roaming disallowed")
	ErrRoamingActivation  = errors.DefineFailedPrecondition("roaming_activation", "roaming activation disallowed")
	ErrActivation         = errors.DefineFailedPrecondition("activation", "activation disallowed")
	ErrUnknownDevEUI      = errors.DefineNotFound("unknown_dev_eui", "unknown DevEUI")
	ErrUnknownDevAddr     = errors.DefineNotFound("unknown_dev_addr", "unknown DevAddr")
	ErrUnknownSender      = errors.DefineNotFound("unknown_sender", "unknown sender")
	ErrUnknownReceiver    = errors.DefineNotFound("unknown_receiver", "unknown receiver")
	ErrDeferred           = errors.DefineAborted("deferred", "deferred")
	ErrTransmitFailed     = errors.DefineAborted("transmit_failed", "transmit failed")
	ErrFPort              = errors.DefineInvalidArgument("f_port", "invalid FPort")
	ErrProtocolVersion    = errors.DefineInvalidArgument("protocol_version", "invalid protocol version")
	ErrStaleDeviceProfile = errors.DefineFailedPrecondition("stale_device_profile", "stale device profile")
	ErrMalformedMessage   = errors.DefineInvalidArgument("malformed_message", "malformed message")
	ErrFrameSize          = errors.DefineInvalidArgument("frame_size", "frame size error")

	resultErrors = map[ResultCode]*errors.Definition{
		ResultNoAction:               ErrNoAction,
		ResultMICFailed:              ErrMIC,
		ResultFrameReplayed:          ErrFrameReplayed,
		ResultJoinReqFailed:          ErrJoinReq,
		ResultNoRoamingAgreement:     ErrNoRoamingAgreement,
		ResultDevRoamingDisallowed:   ErrDeviceRoaming,
		ResultRoamingActDisallowed:   ErrRoamingActivation,
		ResultActivationDisallowed:   ErrActivation,
		ResultUnknownDevEUI:          ErrUnknownDevEUI,
		ResultUnknownDevAddr:         ErrUnknownDevAddr,
		ResultUnknownSender:          ErrUnknownSender,
		ResultUnkownReceiver:         ErrUnknownReceiver,
		ResultUnknownReceiver:        ErrUnknownReceiver,
		ResultDeferred:               ErrDeferred,
		ResultXmitFailed:             ErrTransmitFailed,
		ResultInvalidFPort:           ErrFPort,
		ResultInvalidProtocolVersion: ErrProtocolVersion,
		ResultStaleDeviceProfile:     ErrStaleDeviceProfile,
		ResultMalformedMessage:       ErrMalformedMessage,
		ResultMalformedRequest:       ErrMalformedMessage,
		ResultFrameSizeError:         ErrFrameSize,
	}
	errorResultCodes = map[*errors.Definition]map[ProtocolVersion]ResultCode{
		ErrNoAction: {
			ProtocolV1_1: ResultNoAction,
		},
		ErrMIC: {
			ProtocolV1_0: ResultMICFailed,
			ProtocolV1_1: ResultMICFailed,
		},
		ErrFrameReplayed: {
			ProtocolV1_1: ResultFrameReplayed,
		},
		ErrJoinReq: {
			ProtocolV1_0: ResultJoinReqFailed,
			ProtocolV1_1: ResultJoinReqFailed,
		},
		ErrNoRoamingAgreement: {
			ProtocolV1_0: ResultNoRoamingAgreement,
			ProtocolV1_1: ResultNoRoamingAgreement,
		},
		ErrDeviceRoaming: {
			ProtocolV1_0: ResultDevRoamingDisallowed,
			ProtocolV1_1: ResultDevRoamingDisallowed,
		},
		ErrRoamingActivation: {
			ProtocolV1_0: ResultRoamingActDisallowed,
			ProtocolV1_1: ResultRoamingActDisallowed,
		},
		ErrActivation: {
			ProtocolV1_0: ResultActivationDisallowed,
			ProtocolV1_1: ResultActivationDisallowed,
		},
		ErrUnknownDevEUI: {
			ProtocolV1_0: ResultUnknownDevEUI,
			ProtocolV1_1: ResultUnknownDevEUI,
		},
		ErrUnknownDevAddr: {
			ProtocolV1_0: ResultUnknownDevAddr,
			ProtocolV1_1: ResultUnknownDevAddr,
		},
		ErrUnknownSender: {
			ProtocolV1_0: ResultUnknownSender,
			ProtocolV1_1: ResultUnknownSender,
		},
		ErrUnknownReceiver: {
			ProtocolV1_0: ResultUnkownReceiver,
			ProtocolV1_1: ResultUnkownReceiver,
		},
		ErrDeferred: {
			ProtocolV1_0: ResultDeferred,
			ProtocolV1_1: ResultDeferred,
		},
		ErrTransmitFailed: {
			ProtocolV1_0: ResultXmitFailed,
			ProtocolV1_1: ResultXmitFailed,
		},
		ErrFPort: {
			ProtocolV1_0: ResultInvalidFPort,
			ProtocolV1_1: ResultInvalidFPort,
		},
		ErrProtocolVersion: {
			ProtocolV1_0: ResultInvalidProtocolVersion,
			ProtocolV1_1: ResultInvalidProtocolVersion,
		},
		ErrStaleDeviceProfile: {
			ProtocolV1_0: ResultStaleDeviceProfile,
			ProtocolV1_1: ResultStaleDeviceProfile,
		},
		ErrMalformedMessage: {
			ProtocolV1_0: ResultMalformedRequest,
			ProtocolV1_1: ResultMalformedMessage,
		},
		ErrFrameSize: {
			ProtocolV1_0: ResultFrameSizeError,
			ProtocolV1_1: ResultFrameSizeError,
		},
	}
)
