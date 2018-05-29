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

import "go.thethings.network/lorawan-stack/pkg/errors"

var (
	// ErrNoNetworkServerFound is returned if no Network Server was found for a passed DevAddr.
	ErrNoNetworkServerFound = &errors.ErrDescriptor{
		MessageFormat: "No Network Server found for this message",
		Code:          1,
		Type:          errors.NotFound,
	}
	// ErrNoIdentityServerFound is returned if no Identity Server was found.
	ErrNoIdentityServerFound = &errors.ErrDescriptor{
		MessageFormat: "No Identity Server found",
		Code:          2,
		Type:          errors.NotFound,
	}
	// ErrUnauthorized is returned if there are no credentials passed.
	ErrUnauthorized = &errors.ErrDescriptor{
		MessageFormat: "No credentials passed",
		Code:          3,
		Type:          errors.Unauthorized,
	}
	// ErrGatewayNotConnected is returned when a send operation failed because a gateway is not connected.
	ErrGatewayNotConnected = &errors.ErrDescriptor{
		MessageFormat:  "Gateway `{gateway_id}` not connected",
		Code:           4,
		Type:           errors.NotFound,
		SafeAttributes: []string{"gateway_id"},
	}
	// ErrNoReadyConnectionToIdentityServer is returned when the connection to the identity server is not ready yet.
	ErrNoReadyConnectionToIdentityServer = &errors.ErrDescriptor{
		MessageFormat: "No ready connection to the identity server",
		Code:          5,
		Type:          errors.Internal,
	}
	// ErrTranslationFromProtobuf is returned when the translation of a
	// message between the proto format and the UDP format fails.
	ErrTranslationFromProtobuf = &errors.ErrDescriptor{
		MessageFormat: "Could not translate from the protobuf format to UDP",
		Code:          6,
		Type:          errors.Internal,
	}
	// ErrUnsupportedTopicFormat is returned if a topic is unsupported.
	ErrUnsupportedTopicFormat = &errors.ErrDescriptor{
		MessageFormat:  "Unsupported topic format `{topic}`",
		Code:           7,
		Type:           errors.InvalidArgument,
		SafeAttributes: []string{"topic"},
	}
	// ErrInvalidAPIVersion is returned when an invalid API version is passed.
	ErrInvalidAPIVersion = &errors.ErrDescriptor{
		MessageFormat:  "Invalid API version tag `{version}`",
		Code:           8,
		Type:           errors.InvalidArgument,
		SafeAttributes: []string{"version"},
	}
)

func init() {
	ErrNoNetworkServerFound.Register()
	ErrNoIdentityServerFound.Register()
	ErrUnauthorized.Register()
	ErrGatewayNotConnected.Register()
	ErrNoReadyConnectionToIdentityServer.Register()
	ErrTranslationFromProtobuf.Register()
	ErrUnsupportedTopicFormat.Register()
	ErrInvalidAPIVersion.Register()
}
