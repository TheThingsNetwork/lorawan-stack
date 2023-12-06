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
	"encoding/json"

	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// Config represents the full configuration for Semtech's UDP Packet Forwarder.
type Config struct {
	SX1301Conf  []*shared.SX1301Config `json:"SX1301_conf"`
	GatewayConf GatewayConf            `json:"gateway_conf"`
}

// SingleSX1301Config is a helper type for marshaling a config with a single SX1301Config.
type singleSX1301Config struct {
	SX1301Conf  *shared.SX1301Config `json:"SX1301_conf"`
	GatewayConf GatewayConf          `json:"gateway_conf"`
}

// MarshalJSON implements json.Marshaler.
// Serializes the SX1301Conf field as an object if it contains a single element and as an array otherwise.
func (c Config) MarshalJSON() ([]byte, error) {
	if len(c.SX1301Conf) == 1 {
		return json.Marshal(singleSX1301Config{
			SX1301Conf:  c.SX1301Conf[0],
			GatewayConf: c.GatewayConf,
		})
	}
	type alias Config
	return json.Marshal(alias(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Config) UnmarshalJSON(data []byte) error {
	var single singleSX1301Config
	if err := json.Unmarshal(data, &single); err == nil {
		c.SX1301Conf = []*shared.SX1301Config{single.SX1301Conf}
		c.GatewayConf = single.GatewayConf
		return nil
	}
	type alias Config
	return json.Unmarshal(data, (*alias)(c))
}

// GatewayConf contains the configuration for the gateway's server connection.
type GatewayConf struct {
	GatewayID      string        `json:"gateway_ID,omitempty"`
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

	if gateway.GetIds().GetEui() != nil {
		c.GatewayConf.GatewayID = types.MustEUI64(gateway.GetIds().GetEui()).String()
	}
	c.GatewayConf.ServerAddress = host
	c.GatewayConf.ServerPortUp = uint32(port)
	c.GatewayConf.ServerPortDown = uint32(port)
	server := c.GatewayConf
	server.Enabled = true
	c.GatewayConf.Servers = append(c.GatewayConf.Servers, server)

	c.SX1301Conf = make([]*shared.SX1301Config, 0, len(gateway.FrequencyPlanIds))
	for _, frequencyPlanID := range gateway.FrequencyPlanIds {
		frequencyPlan, err := store.GetByID(frequencyPlanID)
		if err != nil {
			return nil, err
		}
		if len(frequencyPlan.Radios) != 0 {
			sx1301Config, err := shared.BuildSX1301Config(frequencyPlan)
			if err != nil {
				return nil, err
			}
			c.SX1301Conf = append(c.SX1301Conf, sx1301Config)
		}
	}
	return &c, nil
}
