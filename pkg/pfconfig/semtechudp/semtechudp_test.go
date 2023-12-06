// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package semtechudp_test

import (
	"encoding/json"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/semtechudp"
	"go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func getSX1301Conf(t *testing.T) *shared.SX1301Config {
	t.Helper()

	return &shared.SX1301Config{
		LoRaWANPublic: true,
		ClockSource:   1,
		AntennaGain:   0,
		Radios: []shared.RFConfig{
			{
				Enable:     true,
				Type:       "SX1257",
				Frequency:  867500000,
				TxEnable:   true,
				TxFreqMin:  863000000,
				TxFreqMax:  870000000,
				RSSIOffset: -166,
			},
			{
				Enable: true, Type: "SX1257",
				Frequency:  868500000,
				TxEnable:   false,
				TxFreqMin:  0,
				TxFreqMax:  0,
				RSSIOffset: -166,
			},
		},
		Channels: []shared.IFConfig{
			{Enable: true, Radio: 1, IFValue: -400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
			{Enable: true, Radio: 1, IFValue: -200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
			{Enable: true, Radio: 1, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
			{Enable: true, Radio: 0, IFValue: -400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
			{Enable: true, Radio: 0, IFValue: -200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
			{Enable: true, Radio: 0, IFValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
			{Enable: true, Radio: 0, IFValue: 200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
			{Enable: true, Radio: 0, IFValue: 400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
		},
		LoRaStandardChannel: &shared.IFConfig{
			Enable:       true,
			Radio:        1,
			IFValue:      -200000,
			Bandwidth:    250000,
			SpreadFactor: 7,
			Datarate:     0,
		},
		FSKChannel: &shared.IFConfig{
			Enable:       true,
			Radio:        1,
			IFValue:      -200000,
			Bandwidth:    250000,
			SpreadFactor: 7,
			Datarate:     0,
		},
		TxLUTConfigs: []shared.TxLUTConfig{
			{PAGain: 0, MixGain: 8, RFPower: -6, DigGain: 0},
			{PAGain: 0, MixGain: 10, RFPower: -3, DigGain: 0},
			{PAGain: 0, MixGain: 12, RFPower: 0, DigGain: 0},
			{PAGain: 1, MixGain: 8, RFPower: 3, DigGain: 0},
			{PAGain: 1, MixGain: 10, RFPower: 6, DigGain: 0},
			{PAGain: 1, MixGain: 12, RFPower: 10, DigGain: 0},
			{PAGain: 1, MixGain: 13, RFPower: 11, DigGain: 0},
			{PAGain: 2, MixGain: 9, RFPower: 12, DigGain: 0},
			{PAGain: 1, MixGain: 15, RFPower: 13, DigGain: 0},
			{PAGain: 2, MixGain: 10, RFPower: 14, DigGain: 0},
			{PAGain: 2, MixGain: 11, RFPower: 16, DigGain: 0},
			{PAGain: 3, MixGain: 9, RFPower: 20, DigGain: 0},
			{PAGain: 3, MixGain: 10, RFPower: 23, DigGain: 0},
			{PAGain: 3, MixGain: 11, RFPower: 25, DigGain: 0},
			{PAGain: 3, MixGain: 12, RFPower: 26, DigGain: 0},
			{PAGain: 3, MixGain: 14, RFPower: 27, DigGain: 0},
		},
	}
}

func getGtwConfig(t *testing.T) semtechudp.GatewayConf {
	t.Helper()

	return semtechudp.GatewayConf{
		ServerAddress:  "localhost",
		ServerPortUp:   1700,
		ServerPortDown: 1700,
		Servers: []semtechudp.GatewayConf{
			{
				ServerAddress:  "localhost",
				ServerPortUp:   1700,
				ServerPortDown: 1700,
				Enabled:        true,
			},
		},
	}
}

func TestConfigSerialization(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name string
		*semtechudp.Config
	}{
		{
			Name: "Single Frequency Plan",
			Config: &semtechudp.Config{
				SX1301Conf: []*shared.SX1301Config{
					getSX1301Conf(t),
				},
				GatewayConf: getGtwConfig(t),
			},
		},
		{
			Name: "Multiple Frequency Plans",
			Config: &semtechudp.Config{
				SX1301Conf: []*shared.SX1301Config{
					getSX1301Conf(t),
					getSX1301Conf(t),
					getSX1301Conf(t),
				},
				GatewayConf: getGtwConfig(t),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)
			marshalled, err := json.Marshal(tc.Config)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			var cfg2 semtechudp.Config
			err = json.Unmarshal(marshalled, &cfg2)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			a.So(cfg2, should.Resemble, *tc.Config)
		})
	}
}
