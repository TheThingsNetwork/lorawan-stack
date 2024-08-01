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

import (
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	gs "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver"
)

// DefaultGatewayConfigurationServerConfig is the default configuration for the Gateway Configuration Server.
var DefaultGatewayConfigurationServerConfig = gatewayconfigurationserver.Config{
	RequireAuth: true,
}

func init() {
	DefaultGatewayConfigurationServerConfig.TheThingsKickstarterGateway.Default.UpdateChannel = "stable"
	DefaultGatewayConfigurationServerConfig.TheThingsKickstarterGateway.Default.MQTTServer = "mqtts://" + gs.DefaultGatewayServerConfig.MQTTV2.PublicTLSAddress
	DefaultGatewayConfigurationServerConfig.TheThingsKickstarterGateway.Default.FirmwareURL = "https://ttkg-fw.thethingsindustries.com/v1"
	DefaultGatewayConfigurationServerConfig.BasicStation.Default.LNSURI = "wss://" + shared.DefaultPublicHost + gs.DefaultGatewayServerConfig.BasicStation.ListenTLS
}
