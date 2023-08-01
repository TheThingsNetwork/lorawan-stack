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
	"time"

	"go.thethings.network/lorawan-stack/v3/cmd/internal/commands"
	conf "go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/discover"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
)

var (
	defaultInsecure                  = false
	defaultClusterHost               = "localhost"
	defaultGRPCAddress, _            = discover.DefaultPort(defaultClusterHost, discover.DefaultPorts[!defaultInsecure])
	defaultOAuthServerBaseAddress, _ = discover.DefaultURL(defaultClusterHost, discover.DefaultHTTPPorts[!defaultInsecure], !defaultInsecure)
	defaultOAuthServerAddress        = defaultOAuthServerBaseAddress + "/oauth"
	defaultRetryConfig               = RetryConfig{
		DefaultTimeout: 100 * time.Millisecond,
		EnableMetadata: true,
	}
	defaultTelemetryConfig = telemetry.CLI{
		Enable: true,
		Target: "https://telemetry.thethingsstack.io/collect",
	}
)

// Config for the ttn-lw-cli binary.
type Config struct {
	conf.Base                          `name:",squash"`
	CredentialsID                      string        `name:"credentials-id" yaml:"credentials-id" description:"Credentials ID (if using multiple configurations)"`                                 //nolint:lll
	InputFormat                        string        `name:"input-format" yaml:"input-format" description:"Input format"`                                                                          //nolint:lll
	OutputFormat                       string        `name:"output-format" yaml:"output-format" description:"Output format"`                                                                       //nolint:lll
	AllowUnknownHosts                  bool          `name:"allow-unknown-hosts" yaml:"allow-unknown-hosts" description:"Allow sending credentials to unknown hosts"`                              //nolint:lll
	OAuthServerAddress                 string        `name:"oauth-server-address" yaml:"oauth-server-address" description:"OAuth Server address"`                                                  //nolint:lll
	IdentityServerGRPCAddress          string        `name:"identity-server-grpc-address" yaml:"identity-server-grpc-address" description:"Identity Server address"`                               //nolint:lll
	GatewayServerEnabled               bool          `name:"gateway-server-enabled" yaml:"gateway-server-enabled" description:"Gateway Server enabled"`                                            //nolint:lll
	GatewayServerGRPCAddress           string        `name:"gateway-server-grpc-address" yaml:"gateway-server-grpc-address" description:"Gateway Server address"`                                  //nolint:lll
	NetworkServerEnabled               bool          `name:"network-server-enabled" yaml:"network-server-enabled" description:"Network Server enabled"`                                            //nolint:lll
	NetworkServerGRPCAddress           string        `name:"network-server-grpc-address" yaml:"network-server-grpc-address" description:"Network Server address"`                                  //nolint:lll
	ApplicationServerEnabled           bool          `name:"application-server-enabled" yaml:"application-server-enabled" description:"Application Server enabled"`                                //nolint:lll
	ApplicationServerGRPCAddress       string        `name:"application-server-grpc-address" yaml:"application-server-grpc-address" description:"Application Server address"`                      //nolint:lll
	JoinServerEnabled                  bool          `name:"join-server-enabled" yaml:"join-server-enabled" description:"Join Server enabled"`                                                     //nolint:lll
	JoinServerGRPCAddress              string        `name:"join-server-grpc-address" yaml:"join-server-grpc-address" description:"Join Server address"`                                           //nolint:lll
	DeviceTemplateConverterGRPCAddress string        `name:"device-template-converter-grpc-address" yaml:"device-template-converter-grpc-address" description:"Device Template Converter address"` //nolint:lll
	DeviceClaimingServerGRPCAddress    string        `name:"device-claiming-server-grpc-address" yaml:"device-claiming-server-grpc-address" description:"Device Claiming Server address"`          //nolint:lll
	QRCodeGeneratorGRPCAddress         string        `name:"qr-code-generator-grpc-address" yaml:"qr-code-generator-grpc-address" description:"QR Code Generator address"`                         //nolint:lll
	PacketBrokerAgentGRPCAddress       string        `name:"packet-broker-agent-grpc-address" yaml:"packet-broker-agent-grpc-address" description:"Packet Broker Agent address"`                   //nolint:lll
	Insecure                           bool          `name:"insecure" yaml:"insecure" description:"Connect without TLS"`                                                                           //nolint:lll
	CA                                 string        `name:"ca" yaml:"ca" description:"CA certificate file"`
	DumpRequests                       bool          `name:"dump-requests" yaml:"dump-requests" description:"When log level is set to debug, also dump request payload as JSON"` //nolint:lll
	SkipVersionCheck                   bool          `name:"skip-version-check" yaml:"skip-version-check" description:"Do not perform version checks"`                           //nolint:lll
	Retry                              RetryConfig   `name:"retry" yaml:"retry"`
	Telemetry                          telemetry.CLI `name:"telemetry" yaml:"telemetry" description:"Telemetry configuration"` //nolint:lll
}

// RetryConfig defines the values for the retry behavior in the CLI.
type RetryConfig struct {
	Max            uint          `name:"max" yaml:"max" description:"Maximum amount of times that a request can be reattempted"`                                                     //nolint:lll
	DefaultTimeout time.Duration `name:"default-timeout" yaml:"default-timeout" description:"Default timeout between retry attempts"`                                                //nolint:lll
	EnableMetadata bool          `name:"enable-metadata" yaml:"enable-metadata" description:"Use request response metadata to dynamically calculate timeout between retry attempts"` //nolint:lll
	Jitter         float64       `name:"jitter" yaml:"jitter" description:"Fraction that creates a deviation of the timeout used between retry attempts"`                            //nolint:lll
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
	hosts = append(hosts, c.DeviceClaimingServerGRPCAddress)
	hosts = append(hosts, c.QRCodeGeneratorGRPCAddress)
	hosts = append(hosts, c.PacketBrokerAgentGRPCAddress)
	return getHosts(hosts...)
}

// MakeDefaultConfig builds the default config for the ttn-lw-cli binary for a given host.
func MakeDefaultConfig(clusterGRPCAddress string, oauthServerAddress string, insecure bool) Config {
	return Config{
		Base: conf.Base{
			Log: conf.Log{
				Format: "console",
				Level:  log.InfoLevel,
			},
		},
		InputFormat:                        "json",
		OutputFormat:                       "json",
		OAuthServerAddress:                 oauthServerAddress,
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
		DeviceClaimingServerGRPCAddress:    clusterGRPCAddress,
		QRCodeGeneratorGRPCAddress:         clusterGRPCAddress,
		PacketBrokerAgentGRPCAddress:       clusterGRPCAddress,
		Insecure:                           insecure,
		Retry:                              defaultRetryConfig,
		Telemetry:                          defaultTelemetryConfig,
	}
}

// DefaultConfig contains the default config for the ttn-lw-cli binary.
var DefaultConfig = MakeDefaultConfig(defaultGRPCAddress, defaultOAuthServerAddress, defaultInsecure)

var configCommand = commands.Config(mgr)

func init() {
	configCommand.PersistentPreRunE = preRun()
	Root.AddCommand(configCommand)
}
