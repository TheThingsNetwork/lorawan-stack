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

package networkserver

import (
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNeedsADRParamSetupReq(t *testing.T) {
	type TestCase struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Band        band.Band
		Needs       bool
	}
	var tcs []TestCase

	ForEachBand(t, func(makeBandName func(parts ...string) string, phy band.Band) {
		tcs = append(tcs,
			TestCase{
				Name:        makeBandName("no MAC state"),
				InputDevice: &ttnpb.EndDevice{},
				Band:        phy,
			},
		)
		for _, conf := range []struct {
			Suffix                               string
			CurrentParameters, DesiredParameters ttnpb.MACParameters
			Needs                                bool
		}{
			{
				Suffix: "current(limit:nil,delay:nil),desired(limit:nil,delay:nil)",
			},
			{
				Suffix: "current(limit:32768,delay:1024),desired(limit:32768,delay:1024)",
				CurrentParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
				DesiredParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
			},
			{
				Suffix: "current(limit:32768,delay:1024),desired(limit:nil,delay:nil)",
				CurrentParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
			},
			{
				Suffix: "current(limit:nil,delay:1024),desired(limit:32768,delay:1024)",
				CurrentParameters: ttnpb.MACParameters{
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
				DesiredParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
				Needs: phy.ADRAckLimit != ttnpb.ADR_ACK_LIMIT_32768,
			},
			{
				Suffix: "current(limit:nil,delay:nil),desired(limit:32768,delay:1024)",
				DesiredParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
				Needs: phy.ADRAckLimit != ttnpb.ADR_ACK_LIMIT_32768 || phy.ADRAckDelay != ttnpb.ADR_ACK_DELAY_1024,
			},
			{
				Suffix: "current(limit:32768,delay:nil),desired(limit:nil,delay:1024)",
				CurrentParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
				},
				DesiredParameters: ttnpb.MACParameters{
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
				Needs: phy.ADRAckDelay != ttnpb.ADR_ACK_DELAY_1024,
			},
			{
				Suffix: "current(limit:32768,delay:1024),desired(limit:32768,delay:2048)",
				CurrentParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_1024,
					},
				},
				DesiredParameters: ttnpb.MACParameters{
					ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADR_ACK_LIMIT_32768,
					},
					ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADR_ACK_DELAY_2048,
					},
				},
				Needs: true,
			},
		} {
			ForEachMACVersion(func(makeMACName func(parts ...string) string, macVersion ttnpb.MACVersion) {
				tcs = append(tcs,
					TestCase{
						Name: makeBandName(makeMACName(conf.Suffix)),
						InputDevice: &ttnpb.EndDevice{
							MACState: &ttnpb.MACState{
								LoRaWANVersion:    macVersion,
								CurrentParameters: conf.CurrentParameters,
								DesiredParameters: conf.DesiredParameters,
							},
						},
						Band:  phy,
						Needs: conf.Needs && macVersion.Compare(ttnpb.MAC_V1_1) >= 0,
					},
				)
			})
		}
	})

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.InputDevice)
			res := needsADRParamSetupReq(dev, tc.Band)
			if tc.Needs {
				a.So(res, should.BeTrue)
			} else {
				a.So(res, should.BeFalse)
			}
			a.So(dev, should.Resemble, tc.InputDevice)
		})
	}
}

func TestHandleADRParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Events           []events.DefinitionDataClosure
		Error            error
	}{
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveADRParamSetupAnswer.BindData(nil),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "limit 32768, delay 1024",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_ADRParamSetupReq{
							ADRAckLimitExponent: ttnpb.ADR_ACK_LIMIT_32768,
							ADRAckDelayExponent: ttnpb.ADR_ACK_DELAY_1024,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						ADRAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADR_ACK_LIMIT_32768},
						ADRAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{Value: ttnpb.ADR_ACK_DELAY_1024},
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Events: []events.DefinitionDataClosure{
				evtReceiveADRParamSetupAnswer.BindData(nil),
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			evs, err := handleADRParamSetupAns(test.Context(), dev)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(evs, should.ResembleEventDefinitionDataClosures, tc.Events)
		})
	}
}
