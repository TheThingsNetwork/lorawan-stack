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

import "go.thethings.network/lorawan-stack/v3/pkg/errors"

// Errors returned by component initialization.
var (
	ErrInitializeBaseComponent              = errors.Define("initialize_base_component", "initialize base component")
	ErrInvalidLogFormat                     = errors.DefineInvalidArgument("log_format", "invalid log format `{format}`")
	ErrInitializeLogger                     = errors.Define("initialize_logger", "initialize logger")
	ErrInitializeIdentityServer             = errors.Define("initialize_identity_server", "initialize Identity Server")
	ErrInitializeGatewayServer              = errors.Define("initialize_gateway_server", "initialize Gateway Server")
	ErrInitializeNetworkServer              = errors.Define("initialize_network_server", "initialize Network Server")
	ErrInitializeApplicationServer          = errors.Define("initialize_application_server", "initialize Application Server")
	ErrInitializeJoinServer                 = errors.Define("initialize_join_server", "initialize Join Server")
	ErrInitializeConsole                    = errors.Define("initialize_console", "initialize Console")
	ErrInitializeGatewayConfigurationServer = errors.Define("initialize_gateway_configuration_server", "initialize Gateway Configuration Server")
	ErrInitializeDeviceTemplateConverter    = errors.Define("initialize_device_template_converter", "initialize Device Template Converter")
	ErrInitializeQRCodeGenerator            = errors.Define("initialize_qr_code_generator", "initialize QR Code Generator")
	ErrInitializePacketBrokerAgent          = errors.Define("initialize_packet_broker_agent", "initialize Packet Broker Agent")
	ErrInitializeDeviceRepository           = errors.Define("initialize_device_repository", "initialize Device Repository")
	ErrInitializeDeviceClaimingServer       = errors.Define("initialize_device_claiming_server", "initialize Device Claiming Server")
)
