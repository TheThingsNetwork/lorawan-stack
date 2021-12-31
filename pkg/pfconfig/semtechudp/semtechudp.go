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

// Package semtechudp implements the JSON configuration for the Semtech UDP Packet Forwarder.
package semtechudp

import (
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Build builds a packet forwarder configuration for the given gateway, using the given frequency plan store.
func Build(gateway *ttnpb.Gateway, store *frequencyplans.Store) (*ttnpb.SemtechUDPConfig, error) {
	var c ttnpb.SemtechUDPConfig

	host, port, err := shared.ParseGatewayServerAddress(gateway.GatewayServerAddress)
	if err != nil {
		return nil, err
	}

	if gateway.GetIds().GetEui() != nil {
		c.GatewayConfig.GatewayId = gateway.GetEntityIdentifiers().GetGatewayIds()
	}
	c.GatewayConfig.ServerAddress, c.GatewayConfig.ServerPortUp, c.GatewayConfig.ServerPortDown = host, uint32(port), uint32(port)
	server := c.GatewayConfig
	server.Enabled = true
	c.GatewayConfig.Servers = append(c.GatewayConfig.Servers, server)

	frequencyPlan, err := store.GetByID(gateway.FrequencyPlanId)
	if err != nil {
		return nil, err
	}
	if len(frequencyPlan.Radios) != 0 {
		sx1301Config, err := shared.BuildSX1301Config(frequencyPlan)
		if err != nil {
			return nil, err
		}
		c.Sx1301Config = sx1301Config
	}

	return &c, nil
}
