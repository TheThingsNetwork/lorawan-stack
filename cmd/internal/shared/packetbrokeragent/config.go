// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// DefaultPacketBrokerAgentConfig is the default configuration for the Packet Broker Agent.
var DefaultPacketBrokerAgentConfig = packetbrokeragent.Config{
	IAMAddress:          "iam.packetbroker.net:443",
	ControlPlaneAddress: "cp.packetbroker.net:443",
	MapperAddress:       "mapper.packetbroker.net:443",
	AuthenticationMode:  "oauth2",
	OAuth2: packetbrokeragent.OAuth2Config{
		TokenURL: "https://iam.packetbroker.net/token",
	},
	Registration: packetbrokeragent.RegistrationConfig{
		Listed: true,
	},
	HomeNetwork: packetbrokeragent.HomeNetworkConfig{
		WorkerPool: packetbrokeragent.WorkerPoolConfig{
			Limit: 4096,
		},
		IncludeHops:     false,
		DevAddrPrefixes: []types.DevAddrPrefix{{}}, // Subscribe to all DevAddr prefixes.
	},
	Forwarder: packetbrokeragent.ForwarderConfig{
		WorkerPool: packetbrokeragent.WorkerPoolConfig{
			Limit: 1024,
		},
		IncludeGatewayEUI: true,
		IncludeGatewayID:  true,
		HashGatewayID:     false,
		GatewayOnlineTTL:  packetbrokeragent.DefaultGatewayOnlineTTL,
	},
}
