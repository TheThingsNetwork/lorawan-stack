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
	"bytes"

	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

type LorafwdGatewayConfig struct {
	ID *types.EUI64
}

type LorafwdGWMPConfig struct {
	Node            string
	ServiceUplink   uint16
	ServiceDownlink uint16
}

// LorafwdConfig represents the Lorafwd configuration of Semtech's UDP Packet Forwarder.
type LorafwdConfig struct {
	Gateway  LorafwdGatewayConfig
	Filter   struct{}
	Database struct{}
	GWMP     LorafwdGWMPConfig
	API      struct{}
}

func (conf LorafwdConfig) MarshalText() ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := lorafwdTmpl.Execute(buf, conf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// LoradGatewayConf contains the Lorad configuration for the gateway's server connection.
type LoradGatewayConf struct {
	BeaconEnable    bool        `json:"beacon_enable"`
	BeaconPeriod    uint        `json:"beacon_period,omitempty"`
	BeaconFreqHz    uint        `json:"beacon_freq_hz,omitempty"`
	BeaconFreqNb    uint        `json:"beacon_freq_nb,omitempty"`
	BeaconStep      uint        `json:"beacon_step,omitempty"`
	BeaconDatarate  uint        `json:"beacon_datarate,omitempty"`
	BeaconBwHz      uint        `json:"beacon_bw_hz,omitempty"`
	BeaconPower     uint        `json:"beacon_power,omitempty"`
	BeaconInfodesc  interface{} `json:"beacon_infodesc,omitempty"`
	BeaconLatitude  float64     `json:"beacon_latitude,omitempty"`
	BeaconLongitude float64     `json:"beacon_longitude,omitempty"`
}

type LoradSX1301Conf struct {
	shared.SX1301Config
	InsertionLoss     float32 `json:"insertion_loss"`
	InsertionLossDesc string  `json:"insertion_loss_desc,omitempty"`
	AntennaGainDesc   string  `json:"antenna_gain_desc,omitempty"`
}

// LoradConfig represents the Lorad configuration of Semtech's UDP Packet Forwarder.
type LoradConfig struct {
	SX1301Conf  LoradSX1301Conf  `json:"SX1301_conf"`
	GatewayConf LoradGatewayConf `json:"gateway_conf"`
}

// BuildLorad builds Lorad configuration for the given gateway, using the given frequency plan store.
func BuildLorad(gtw *ttnpb.Gateway, fps *frequencyplans.Store) (*LoradConfig, error) {
	fp, err := fps.GetByID(gtw.FrequencyPlanID)
	if err != nil {
		return nil, err
	}
	sx1301Conf, err := shared.BuildSX1301Config(fp)
	if err != nil {
		return nil, err
	}
	var gatewayConf LoradGatewayConf
	if len(gtw.Antennas) > 0 {
		a := gtw.Antennas[0]
		sx1301Conf.AntennaGain = a.Gain
		gatewayConf.BeaconLatitude = a.Location.Latitude
		gatewayConf.BeaconLongitude = a.Location.Longitude
	}
	// TODO: Configure Class B (https://github.com/TheThingsNetwork/lorawan-stack/issues/1748).
	return &LoradConfig{
		SX1301Conf: LoradSX1301Conf{
			SX1301Config: *sx1301Conf,
			// Following fields are set equal to defaults present in CPF 1.1.6 DOTA for Kerlink Wirnet Station.
			AntennaGainDesc:   "Antenna gain, in dBi",
			InsertionLoss:     0.5,
			InsertionLossDesc: "Insertion loss, in dBi",
		},
		GatewayConf: gatewayConf,
	}, nil
}

// BuildLorafwd builds Lorafwd configuration for the given gateway.
func BuildLorafwd(gtw *ttnpb.Gateway) (*LorafwdConfig, error) {
	host, port, err := shared.ParseGatewayServerAddress(gtw.GatewayServerAddress)
	if err != nil {
		return nil, err
	}
	return &LorafwdConfig{
		Gateway: LorafwdGatewayConfig{
			ID: gtw.EUI,
		},
		GWMP: LorafwdGWMPConfig{
			Node:            host,
			ServiceUplink:   port,
			ServiceDownlink: port,
		},
	}, nil
}
