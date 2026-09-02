// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package ttgc

import (
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// LBSCUPSConfig is the configuration for LoRa Basics Station gateways that connect
// to The Things Gateway Controller via CUPS.
type LBSCUPSConfig struct {
	LNSPort uint16 `name:"lns-port" description:"LoRa Basics Station LNS port of the Gateway Server"`
}

// Config is the configuration for The Things Gateway Controller.
type Config struct {
	Enabled            bool                 `name:"enabled" description:"Enable The Things Gateway Controller"`
	GatewayEUIs        []types.EUI64Range   `name:"gateway-euis" description:"Gateway EUI prefixes or ranges of gateways that connect to The Things Gateway Controller"` //nolint:lll
	ManagedGatewayEUIs []types.EUI64Range   `name:"managed-gateway-euis" description:"Gateway EUI prefixes or ranges of managed gateways"`                               //nolint:lll
	Address            string               `name:"address" description:"The address of The Things Gateway Controller"`
	Domain             string               `name:"domain" description:"The domain of this cluster"`
	TLS                tlsconfig.ClientAuth `name:"tls" description:"TLS configuration"`
	LBSCUPS            LBSCUPSConfig        `name:"lbscups" description:"Configuration for LoRa Basics Station gateways connecting via CUPS"` //nolint:lll
}
