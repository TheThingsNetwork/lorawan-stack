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

package shared

import "go.thethings.network/lorawan-stack/pkg/errors"

// Errors returned by component initialization.
var (
	ErrInitializeBaseComponent              = errors.Define("initialize_base_component", "could not initialize base component")
	ErrInitializeIdentityServer             = errors.Define("initialize_identity_server", "could not initialize Identity Server")
	ErrInitializeGatewayServer              = errors.Define("initialize_gateway_server", "could not initialize Gateway Server")
	ErrInitializeNetworkServer              = errors.Define("initialize_network_server", "could not initialize Network Server")
	ErrInitializeApplicationServer          = errors.Define("initialize_application_server", "could not initialize Application Server")
	ErrInitializeJoinServer                 = errors.Define("initialize_join_server", "could not initialize Join Server")
	ErrInitializeConsole                    = errors.Define("initialize_console", "could not initialize Console")
	ErrInitializeGatewayConfigurationServer = errors.Define("initialize_gateway_configuration_server", "could not initialize Gateway Configuration Server")
	ErrInitializeDeviceTemplateConverter    = errors.Define("initialize_device_template_converter", "could not initialize Device Template Converter")
)
