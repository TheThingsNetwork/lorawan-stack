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

package commands

import (
	"go.thethings.network/lorawan-stack/v3/cmd/internal/commands"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	shared_applicationserver "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/applicationserver"
	shared_console "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/console"
	shared_deviceclaimingserver "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/deviceclaimingserver"
	shared_devicerepository "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/devicerepository"
	shared_devicetemplateconverter "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/devicetemplateconverter"
	shared_gatewayconfigurationserver "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/gatewayconfigurationserver"
	shared_gatewayserver "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/gatewayserver"
	shared_identityserver "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/identityserver"
	shared_joinserver "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/joinserver"
	shared_networkserver "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/networkserver"
	shared_packetbrokeragent "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/packetbrokeragent"
	shared_qrcodegenerator "go.thethings.network/lorawan-stack/v3/cmd/internal/shared/qrcodegenerator"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver"
	conf "go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/console"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/v3/pkg/devicetemplateconverter"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator"
)

// Config for the ttn-lw-stack binary.
type Config struct {
	conf.ServiceBase `name:",squash"`
	IS               identityserver.Config             `name:"is"`
	GS               gatewayserver.Config              `name:"gs"`
	NS               networkserver.Config              `name:"ns"`
	AS               applicationserver.Config          `name:"as"`
	JS               joinserver.Config                 `name:"js"`
	Console          console.Config                    `name:"console"`
	GCS              gatewayconfigurationserver.Config `name:"gcs"`
	DTC              devicetemplateconverter.Config    `name:"dtc"`
	QRG              qrcodegenerator.Config            `name:"qrg"`
	PBA              packetbrokeragent.Config          `name:"pba"`
	DR               devicerepository.Config           `name:"dr"`
	DCS              deviceclaimingserver.Config       `name:"dcs"`
	OutputFormat     string                            `name:"output-format" yaml:"output-format" description:"Output format"`
}

// DefaultConfig contains the default config for the ttn-lw-stack binary.
var DefaultConfig = Config{
	ServiceBase:  shared.DefaultServiceBase,
	IS:           shared_identityserver.DefaultIdentityServerConfig,
	GS:           shared_gatewayserver.DefaultGatewayServerConfig,
	NS:           shared_networkserver.DefaultNetworkServerConfig,
	AS:           shared_applicationserver.DefaultApplicationServerConfig,
	JS:           shared_joinserver.DefaultJoinServerConfig,
	Console:      shared_console.DefaultConsoleConfig,
	GCS:          shared_gatewayconfigurationserver.DefaultGatewayConfigurationServerConfig,
	DTC:          shared_devicetemplateconverter.DefaultDeviceTemplateConverterConfig,
	QRG:          shared_qrcodegenerator.DefaultQRCodeGeneratorConfig,
	PBA:          shared_packetbrokeragent.DefaultPacketBrokerAgentConfig,
	DR:           shared_devicerepository.DefaultDeviceRepositoryConfig,
	DCS:          shared_deviceclaimingserver.DefaultDeviceClaimingServerConfig,
	OutputFormat: "json",
}

func init() {
	Root.AddCommand(commands.Config(mgr))
}
