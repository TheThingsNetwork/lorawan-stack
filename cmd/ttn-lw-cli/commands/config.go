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
	"go.thethings.network/lorawan-stack/cmd/internal/commands"
	conf "go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
)

var (
	clusterHost        = "localhost"
	clusterGRPCAddress = clusterHost + ":8884"
	clusterHTTPAddress = "https://" + clusterHost + ":8885"
)

// Config for the ttn-lw-cli binary.
type Config struct {
	conf.Base                          `name:",squash"`
	CredentialsID                      string `name:"credentials-id" description:"Credentials ID (if using multiple configurations)"`
	InputFormat                        string `name:"input-format" description:"Input format"`
	OutputFormat                       string `name:"output-format" description:"Output format"`
	AllowUnknownHosts                  bool   `name:"allow-unknown-hosts" description:"Allow sending credentials to unknown hosts"`
	OAuthServerAddress                 string `name:"oauth-server-address" description:"OAuth Server address"`
	IdentityServerGRPCAddress          string `name:"identity-server-grpc-address" description:"Identity Server address"`
	GatewayServerEnabled               bool   `name:"gateway-server-enabled" description:"Gateway Server enabled"`
	GatewayServerGRPCAddress           string `name:"gateway-server-grpc-address" description:"Gateway Server address"`
	NetworkServerEnabled               bool   `name:"network-server-enabled" description:"Network Server enabled"`
	NetworkServerGRPCAddress           string `name:"network-server-grpc-address" description:"Network Server address"`
	ApplicationServerEnabled           bool   `name:"application-server-enabled" description:"Application Server enabled"`
	ApplicationServerGRPCAddress       string `name:"application-server-grpc-address" description:"Application Server address"`
	JoinServerEnabled                  bool   `name:"join-server-enabled" description:"Join Server enabled"`
	JoinServerGRPCAddress              string `name:"join-server-grpc-address" description:"Join Server address"`
	DeviceTemplateConverterGRPCAddress string `name:"device-template-converter-grpc-address" description:"Device Template Converter address"`
	DeviceClaimServerGRPCAddress       string `name:"device-claim-server-grpc-address" description:"Device Claim Server address"`
	Insecure                           bool   `name:"insecure" description:"Connect without TLS"`
	CA                                 string `name:"ca" description:"CA certificate file"`
}

func (c Config) getHosts() []string {
	hosts := make([]string, 0, 7)
	hosts = append(hosts, c.OAuthServerAddress)
	hosts = append(hosts, c.IdentityServerGRPCAddress)
	if c.GatewayServerEnabled {
		hosts = append(hosts, c.GatewayServerGRPCAddress)
	}
	if c.NetworkServerEnabled {
		hosts = append(hosts, c.NetworkServerGRPCAddress)
	}
	if c.ApplicationServerEnabled {
		hosts = append(hosts, c.ApplicationServerGRPCAddress)
	}
	if c.JoinServerEnabled {
		hosts = append(hosts, c.JoinServerGRPCAddress)
	}
	hosts = append(hosts, c.DeviceTemplateConverterGRPCAddress)
	hosts = append(hosts, c.DeviceClaimServerGRPCAddress)
	return getHosts(hosts...)
}

// DefaultConfig contains the default config for the ttn-lw-cli binary.
var DefaultConfig = Config{
	Base: conf.Base{
		Log: conf.Log{
			Level: log.InfoLevel,
		},
	},
	InputFormat:                        "json",
	OutputFormat:                       "json",
	OAuthServerAddress:                 clusterHTTPAddress + "/oauth",
	IdentityServerGRPCAddress:          clusterGRPCAddress,
	GatewayServerEnabled:               true,
	GatewayServerGRPCAddress:           clusterGRPCAddress,
	NetworkServerEnabled:               true,
	NetworkServerGRPCAddress:           clusterGRPCAddress,
	ApplicationServerEnabled:           true,
	ApplicationServerGRPCAddress:       clusterGRPCAddress,
	JoinServerEnabled:                  true,
	JoinServerGRPCAddress:              clusterGRPCAddress,
	DeviceTemplateConverterGRPCAddress: clusterGRPCAddress,
	DeviceClaimServerGRPCAddress:       clusterGRPCAddress,
}

var configCommand = commands.Config(mgr)

func init() {
	configCommand.PersistentPreRunE = preRun()
	Root.AddCommand(configCommand)
}
