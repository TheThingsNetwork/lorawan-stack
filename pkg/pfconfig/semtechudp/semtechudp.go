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
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Config represents the full configuration for Semtech's UDP Packet Forwarder.
type Config struct {
	SX1301Conf  shared.SX1301Config `json:"SX1301_conf"`
	GatewayConf GatewayConf         `json:"gateway_conf"`
}

// GatewayConf contains the configuration for the gateway's server connection.
type GatewayConf struct {
	ServerAddress  string        `json:"server_address"`
	ServerPortUp   uint32        `json:"serv_port_up"`
	ServerPortDown uint32        `json:"serv_port_down"`
	Enabled        bool          `json:"serv_enabled,omitempty"` // only used inside servers
	Servers        []GatewayConf `json:"servers,omitempty"`
}

// Build builds a packet forwarder configuration for the given gateway, using the given frequency plan store.
func Build(gateway *ttnpb.Gateway, store *frequencyplans.Store) (*Config, error) {
	var c Config

	host, port, err := shared.ParseGatewayServerAddress(gateway.GatewayServerAddress)
	if err != nil {
		return nil, err
	}
	c.GatewayConf.ServerAddress, c.GatewayConf.ServerPortUp, c.GatewayConf.ServerPortDown = host, uint32(port), uint32(port)
	server := c.GatewayConf
	server.Enabled = true
	c.GatewayConf.Servers = append(c.GatewayConf.Servers, server)

	frequencyPlan, err := store.GetByID(gateway.FrequencyPlanID)
	if err != nil {
		return nil, err
	}
	sx1301Config, err := shared.BuildSX1301Config(frequencyPlan)
	if err != nil {
		return nil, err
	}

	c.SX1301Conf = *sx1301Config

	return &c, nil
}
