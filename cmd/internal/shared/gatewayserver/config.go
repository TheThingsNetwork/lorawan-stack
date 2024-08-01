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
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ttigw"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/packetbroker"
)

// DefaultGatewayServerConfig is the default configuration for the GatewayServer.
var DefaultGatewayServerConfig = gatewayserver.Config{
	RequireRegisteredGateways:         false,
	FetchGatewayInterval:              10 * time.Minute,
	FetchGatewayJitter:                0.2,
	UpdateGatewayLocationDebounceTime: time.Hour,
	UpdateConnectionStatsInterval:     time.Minute,
	UpdateConnectionStatsDebounceTime: 30 * time.Second,
	ConnectionStatsTTL:                12 * time.Hour,
	ConnectionStatsDisconnectTTL:      48 * time.Hour,
	UpdateVersionInfoDelay:            5 * time.Second,
	Forward: map[string][]string{
		"": {"00000000/0"},
	},
	PacketBroker: gatewayserver.PacketBrokerConfig{
		UpdateGatewayInterval: packetbroker.DefaultUpdateGatewayInterval,
		UpdateGatewayJitter:   packetbroker.DefaultUpdateGatewayJitter,
		OnlineTTLMargin:       packetbroker.DefaultOnlineTTLMargin,
	},
	UDP: gatewayserver.UDPConfig{
		Config: udp.DefaultConfig,
		Listeners: map[string]string{
			":1700": "",
		},
	},
	MQTTV2: config.MQTT{
		Listen:           ":1881",
		ListenTLS:        ":8881",
		PublicAddress:    fmt.Sprintf("%s:1881", shared.DefaultPublicHost),
		PublicTLSAddress: fmt.Sprintf("%s:8881", shared.DefaultPublicHost),
	},
	MQTT: config.MQTT{
		Listen:           ":1882",
		ListenTLS:        ":8882",
		PublicAddress:    fmt.Sprintf("%s:1882", shared.DefaultPublicHost),
		PublicTLSAddress: fmt.Sprintf("%s:8882", shared.DefaultPublicHost),
	},
	BasicStation: gatewayserver.BasicStationConfig{
		Config:                 semtechws.DefaultConfig,
		MaxValidRoundTripDelay: 10 * time.Second,
		Listen:                 ":1887",
		ListenTLS:              ":8887",
	},
	TheThingsIndustriesGateway: gatewayserver.TheThingsIndustriesGatewayConfig{
		Config:    ttigw.DefaultConfig,
		Listen:    ":1889",
		ListenTLS: ":8889",
	},
}
