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

package cluster

// Config represents clustering configuration.
type Config struct {
	Join                       []string `name:"join" description:"Addresses of cluster peers to join"`
	Name                       string   `name:"name" description:"Name of the current cluster peer (default: $HOSTNAME)"`
	Address                    string   `name:"address" description:"Address to use for cluster communication"`
	IdentityServer             string   `name:"identity-server" description:"Address for the Identity Server"`
	GatewayServer              string   `name:"gateway-server" description:"Address for the Gateway Server"`
	NetworkServer              string   `name:"network-server" description:"Address for the Network Server"`
	ApplicationServer          string   `name:"application-server" description:"Address for the Application Server"`
	JoinServer                 string   `name:"join-server" description:"Address for the Join Server"`
	CryptoServer               string   `name:"crypto-server" description:"Address for the Crypto Server"`
	PacketBrokerAgent          string   `name:"packet-broker-agent" description:"Address of the Packet Broker Agent"`
	DeviceRepository           string   `name:"device-repository" description:"Address for the Device Repository"`
	GatewayConfigurationServer string   `name:"gateway-configuration-server" description:"Address for the Gateway Configuration Server"` //nolint:lll
	DeviceClaimingServer       string   `name:"device-claiming-server" description:"Address for the Device Claiming Server"`             //nolint:lll
	TLS                        bool     `name:"tls" description:"Do cluster gRPC over TLS"`
	TLSServerName              string   `name:"tls-server-name" description:"Server name to use in TLS handshake to cluster peers"`                                                          //nolint:lll
	Keys                       []string `name:"keys" description:"Keys used to communicate between components of the cluster. The first one will be used by the cluster to identify itself"` //nolint:lll
}
