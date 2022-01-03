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

// Package cpf implements the JSON configuration for the Common Packet Forwarder.
package cpf

import (
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// BuildLorad builds Lorad configuration for the given gateway, using the given frequency plan store.
func BuildLorad(gtw *ttnpb.Gateway, fps *frequencyplans.Store) (*ttnpb.LoradConfig, error) {
	fp, err := fps.GetByID(gtw.FrequencyPlanId)
	if err != nil {
		return nil, err
	}
	sx1301Conf := &ttnpb.SX1301Config{}
	if len(fp.Radios) != 0 {
		sx1301Conf, err = shared.BuildSX1301Config(fp)
		if err != nil {
			return nil, err
		}
	}
	var gatewayConf ttnpb.LoradConfig_GatewayConfig
	if antennas := gtw.GetAntennas(); len(antennas) > 0 {
		antenna := antennas[0]
		sx1301Conf.AntennaGain = antenna.Gain
		if location := antenna.Location; location != nil {
			gatewayConf.BeaconLatitude = location.Latitude
			gatewayConf.BeaconLongitude = location.Longitude
		}
	}
	// TODO: Configure Class B (https://github.com/TheThingsNetwork/lorawan-stack/issues/1748).
	return &ttnpb.LoradConfig{
		Sx1301Config: &ttnpb.LoradConfig_LoradSX1301Config{
			GlobalConfig: sx1301Conf,
			// Following fields are set equal to defaults present in CPF 1.1.6 DOTA for Kerlink Wirnet Station.
			AntennaGainDesc:   "Antenna gain, in dBi",
			InsertionLoss:     0.5,
			InsertionLossDesc: "Insertion loss, in dBi",
		},
		GatewayConfig: &gatewayConf,
	}, nil
}

// BuildLorafwd builds Lorafwd configuration for the given gateway.
func BuildLorafwd(gtw *ttnpb.Gateway) (*ttnpb.LoraFwdConfig, error) {
	host, port, err := shared.ParseGatewayServerAddress(gtw.GatewayServerAddress)
	if err != nil {
		return nil, err
	}

	return &ttnpb.LoraFwdConfig{
		Gateway: &ttnpb.GatewayIdentifiers{
			Eui: gtw.GetIds().GetEui(),
		},
		Gwmp: &ttnpb.LoraFwdConfig_GWMPConfig{
			Node:            host,
			ServiceUplink:   uint32(port),
			ServiceDownlink: uint32(port),
		},
	}, nil
}
