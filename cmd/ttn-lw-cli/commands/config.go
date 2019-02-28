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
	conf.Base                `name:",squash"`
	InputFormat              string `name:"input-format" description:"Input format"`
	OutputFormat             string `name:"output-format" description:"Output format"`
	OAuthServerAddress       string `name:"oauth-server-address" description:"OAuth Server Address"`
	IdentityServerAddress    string `name:"identity-server-address" description:"Identity Server Address"`
	GatewayServerAddress     string `name:"gateway-server-address" description:"Gateway Server Address"`
	NetworkServerAddress     string `name:"network-server-address" description:"Network Server Address"`
	ApplicationServerAddress string `name:"application-server-address" description:"Application Server Address"`
	JoinServerAddress        string `name:"join-server-address" description:"Join Server Address"`
	Insecure                 bool   `name:"insecure" description:"Connect without TLS"`
	CA                       string `name:"ca" description:"CA certificate file"`
}

// DefaultConfig contains the default config for the ttn-lw-cli binary.
var DefaultConfig = Config{
	Base: conf.Base{
		Log: conf.Log{
			Level: log.InfoLevel,
		},
	},
	InputFormat:              "json",
	OutputFormat:             "json",
	OAuthServerAddress:       clusterHTTPAddress,
	IdentityServerAddress:    clusterGRPCAddress,
	GatewayServerAddress:     clusterGRPCAddress,
	NetworkServerAddress:     clusterGRPCAddress,
	ApplicationServerAddress: clusterGRPCAddress,
	JoinServerAddress:        clusterGRPCAddress,
}

var configCommand = commands.Config(mgr)

func init() {
	versionCommand.PersistentPreRunE = preRun()
	Root.AddCommand(configCommand)
}
