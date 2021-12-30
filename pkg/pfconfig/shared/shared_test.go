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

package shared_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/pfconfig/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func TestSX1301Conf(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name         string
		FP           *frequencyplans.FrequencyPlan
		SX1301Config ttnpb.SX1301Config
	}{
		{
			"EU_863_870",
			&frequencyplans.FrequencyPlan{
				BandID: "EU_863_870",
				UplinkChannels: []frequencyplans.Channel{
					{Frequency: 868100000, Radio: 1, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 868300000, Radio: 1, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 868500000, Radio: 1, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867100000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867300000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867500000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867700000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867900000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
				},
				DownlinkChannels: []frequencyplans.Channel{
					{Frequency: 868100000, Radio: 1, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 868300000, Radio: 1, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 868500000, Radio: 1, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867100000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867300000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867500000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867700000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
					{Frequency: 867900000, Radio: 0, MinDataRate: 0, MaxDataRate: 5},
				},
				LoRaStandardChannel: &frequencyplans.LoRaStandardChannel{
					Frequency: 868300000,
					DataRate:  6,
					Radio:     1,
				},
				FSKChannel: &frequencyplans.FSKChannel{
					Frequency: 868800000,
					DataRate:  7,
					Radio:     1,
				},
				Radios: []frequencyplans.Radio{
					{
						Enable:     true,
						ChipType:   "SX1257",
						Frequency:  867500000,
						RSSIOffset: -166,
						TxConfiguration: &frequencyplans.RadioTxConfiguration{
							MinFrequency: 863000000,
							MaxFrequency: 870000000,
						},
					},
					{
						Enable:     true,
						ChipType:   "SX1257",
						RSSIOffset: -166,
						Frequency:  868500000,
					},
				},
				ClockSource: 1,
			},
			ttnpb.SX1301Config{
				LorawanPublic: true,
				ClockSource:   1,
				AntennaGain:   0,
				Radios: []*ttnpb.RFConfig{
					{
						Enable:     true,
						Type:       "SX1257",
						Frequency:  867500000,
						TxEnable:   true,
						TxFreqMin:  863000000,
						TxFreqMax:  870000000,
						RssiOffset: -166,
					},
					{
						Enable: true, Type: "SX1257",
						Frequency:  868500000,
						TxEnable:   false,
						TxFreqMin:  0,
						TxFreqMax:  0,
						RssiOffset: -166,
					},
				},
				Channels: []*ttnpb.IFConfig{
					{Enable: true, Radio: 1, IfValue: -400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
					{Enable: true, Radio: 1, IfValue: -200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
					{Enable: true, Radio: 1, IfValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
					{Enable: true, Radio: 0, IfValue: -400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
					{Enable: true, Radio: 0, IfValue: -200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
					{Enable: true, Radio: 0, IfValue: 0, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
					{Enable: true, Radio: 0, IfValue: 200000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
					{Enable: true, Radio: 0, IfValue: 400000, Bandwidth: 0, SpreadFactor: 0, Datarate: 0},
				},
				LoraStandardChannel: &ttnpb.IFConfig{Enable: true, Radio: 1, IfValue: -200000, Bandwidth: 250000, SpreadFactor: 7, Datarate: 0},
				FskChannel:          &ttnpb.IFConfig{Enable: true, Radio: 1, IfValue: 300000, Bandwidth: 125000, SpreadFactor: 0, Datarate: 50000},
				TxLutConfigs: []*ttnpb.TxLUTConfig{
					{PaGain: 0, MixGain: 8, RfPower: -6, DigGain: 0},
					{PaGain: 0, MixGain: 10, RfPower: -3, DigGain: 0},
					{PaGain: 0, MixGain: 12, RfPower: 0, DigGain: 0},
					{PaGain: 1, MixGain: 8, RfPower: 3, DigGain: 0},
					{PaGain: 1, MixGain: 10, RfPower: 6, DigGain: 0},
					{PaGain: 1, MixGain: 12, RfPower: 10, DigGain: 0},
					{PaGain: 1, MixGain: 13, RfPower: 11, DigGain: 0},
					{PaGain: 2, MixGain: 9, RfPower: 12, DigGain: 0},
					{PaGain: 1, MixGain: 15, RfPower: 13, DigGain: 0},
					{PaGain: 2, MixGain: 10, RfPower: 14, DigGain: 0},
					{PaGain: 2, MixGain: 11, RfPower: 16, DigGain: 0},
					{PaGain: 3, MixGain: 9, RfPower: 20, DigGain: 0},
					{PaGain: 3, MixGain: 10, RfPower: 23, DigGain: 0},
					{PaGain: 3, MixGain: 11, RfPower: 25, DigGain: 0},
					{PaGain: 3, MixGain: 12, RfPower: 26, DigGain: 0},
					{PaGain: 3, MixGain: 14, RfPower: 27, DigGain: 0},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			cfg, err := BuildSX1301Config(tc.FP)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(*cfg, should.Resemble, tc.SX1301Config)) {
				t.Fatalf("Invalid config: %v", cfg)
			}
			msg, err := json.Marshal(cfg)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Unexpected error: %v", err)
			}
			var unmarshaledCfg ttnpb.SX1301Config
			err = json.Unmarshal(msg, &unmarshaledCfg)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(unmarshaledCfg, should.Resemble, *cfg)) {
				t.Fatalf("Invalid config after unmarshaling: \n '%v' \n\n '%v'", *cfg, unmarshaledCfg)
			}
		})
	}
}

func TestParseGatewayServerAddress(t *testing.T) {
	for _, tc := range []struct {
		Address        string
		Host           string
		Port           uint16
		ErrorAssertion func(t *testing.T, err error) bool
	}{
		{
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.BeError) && a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Address: ":1701",
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.BeError) && a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Address: "test.example.com",
			Host:    "test.example.com",
			Port:    1700,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Address: "test.example.com:1701",
			Host:    "test.example.com",
			Port:    1701,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(fmt.Sprintf("address:'%s'", tc.Address), func(t *testing.T) {
			a := assertions.New(t)
			host, port, err := ParseGatewayServerAddress(tc.Address)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(host, should.Equal, tc.Host)
				a.So(port, should.Equal, tc.Port)
			}
		})
	}
}
