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

package gatewayserver

import errors "go.thethings.network/lorawan-stack/pkg/errorsv3"

var (
	errUDPSocket = errors.DefineFailedPrecondition("udp_socket", "could not open UDP socket")

	errNoNetworkServerFound              = errors.DefineInternal("no_network_server_found", "no Network Server found for this message")
	errNoIdentityServerFound             = errors.DefineInternal("no_identity_server_found", "no Identity Server found")
	errNoReadyConnectionToIdentityServer = errors.DefineUnavailable("no_ready_connection_to_is", "no ready connection to the Identity Server")

	errNoCredentialsPassed = errors.DefineUnauthenticated("no_credentials_passed", "no credentials passed")
	errAPIKeyNeedsRights   = errors.DefinePermissionDenied(
		"api_key_needs_rights",
		"API key needs the following rights for the gateway `{gateway_uid}` to perform this operation: {rights}",
	)
	errGatewayNotConnected = errors.DefineNotFound("gateway_not_connected", "gateway `{gateway_id}` not connected")
	errNoPULLDATAReceived  = errors.Define("no_pull_data_received", "no PULL_DATA received in the last `{delay}`")

	errNoMetadata                   = errors.DefineInternal("no_metadata", "No metadata present")
	errUnsupportedTopicFormat       = errors.DefineInvalidArgument("topic_format", "unsupported topic format `{topic}`")
	errInvalidAPIVersion            = errors.DefineInvalidArgument("api_version", "invalid API version tag `{version}`")
	errPermissionDeniedForThisTopic = errors.DefinePermissionDenied("topic_rights", "permission denied to subscribe to `{topic}`")
	errUnexpectedAuthenticationType = errors.DefineInvalidArgument("authentication_type", "received `{passed}` authentication type, but expected `{expected}`")

	errMarshalToProtobuf       = errors.DefineInvalidArgument("marshal_to_protobuf", "could not marshal to protobuf")
	errUnmarshalFromProtobuf   = errors.DefineInvalidArgument("unmarshal_from_protobuf", "could not unmarshal message from protobuf")
	errTranslationFromProtobuf = errors.DefineInternal("protocol_translation", "could not translate from the protobuf format to UDP")

	errListGatewayRights = errors.Define("list_gateway_rights", "could not list gateway rights for the passed authentication method")

	errCouldNotRetrieveGatewayInformation     = errors.Define("retrieve_gateway_info", "could not retrieve information about the gateway")
	errCouldNotRetrieveFrequencyPlanOfGateway = errors.Define("retrieve_gtw_frequency_plan", "could not retrieve the frequency plan `{fp_id}` of the gateway")

	errCouldNotBeScheduled          = errors.Define("schedule", "could not schedule downlink")
	errCouldNotComputeTOAOfDownlink = errors.Define("compute_toa_of_downlink", "could not compute the time on air of a downlink")

	errNoDevAddr = errors.DefineInvalidArgument("no_dev_addr_specified", "no DevAddr specified")
	errNoDevEUI  = errors.DefineInvalidArgument("no_dev_eui_specified", "no DevEUI specified")
)
