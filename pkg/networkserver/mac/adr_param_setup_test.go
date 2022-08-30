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

package mac_test

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsADRParamSetupReq(t *testing.T) {
	type TestCase struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Band        *band.Band
		Needs       bool
	}
	var tcs []TestCase

	ForEachBand(t, func(makeBandName func(parts ...string) string, phy *band.Band, _ ttnpb.PHYVersion) {
		tcs = append(tcs,
			TestCase{
				Name:        makeBandName("no MAC state"),
				InputDevice: &ttnpb.EndDevice{},
				Band:        phy,
			},
		)
		for _, conf := range []struct {
			Suffix                               string
			CurrentParameters, DesiredParameters *ttnpb.MACParameters
			Needs                                bool
		}{
			{
				Suffix:            "current(limit:nil,delay:nil),desired(limit:nil,delay:nil)",
				CurrentParameters: &ttnpb.MACParameters{},
				DesiredParameters: &ttnpb.MACParameters{},
			},
			{
				Suffix: "current(limit:32768,delay:1024),desired(limit:32768,delay:1024)",
				CurrentParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
			},
			{
				Suffix: "current(limit:32768,delay:1024),desired(limit:nil,delay:nil)",
				CurrentParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
				DesiredParameters: &ttnpb.MACParameters{},
			},
			{
				Suffix: "current(limit:nil,delay:1024),desired(limit:32768,delay:1024)",
				CurrentParameters: &ttnpb.MACParameters{
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
				Needs: phy.ADRAckLimit != ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
			},
			{
				Suffix:            "current(limit:nil,delay:nil),desired(limit:32768,delay:1024)",
				CurrentParameters: &ttnpb.MACParameters{},
				DesiredParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
				Needs: phy.ADRAckLimit != ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768 || phy.ADRAckDelay != ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
			},
			{
				Suffix: "current(limit:32768,delay:nil),desired(limit:nil,delay:1024)",
				CurrentParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
				Needs: phy.ADRAckDelay != ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
			},
			{
				Suffix: "current(limit:32768,delay:1024),desired(limit:32768,delay:2048)",
				CurrentParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
					},
				},
				DesiredParameters: &ttnpb.MACParameters{
					AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{
						Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
					},
					AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{
						Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2048,
					},
				},
				Needs: true,
			},
		} {
			ForEachMACVersion(t, func(makeMACName func(parts ...string) string, macVersion ttnpb.MACVersion) {
				tcs = append(tcs,
					TestCase{
						Name: makeBandName(makeMACName(conf.Suffix)),
						InputDevice: &ttnpb.EndDevice{
							MacState: &ttnpb.MACState{
								LorawanVersion:    macVersion,
								CurrentParameters: conf.CurrentParameters,
								DesiredParameters: conf.DesiredParameters,
							},
						},
						Band:  phy,
						Needs: conf.Needs && macspec.UseADRParamSetupReq(macVersion),
					},
				)
			})
		}
	})

	for _, tc := range tcs {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)
				res := DeviceNeedsADRParamSetupReq(dev, tc.Band)
				if tc.Needs {
					a.So(res, should.BeTrue)
				} else {
					a.So(res, should.BeFalse)
				}
				a.So(dev, should.Resemble, tc.InputDevice)
			},
		})
	}
}

func TestHandleADRParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Events           events.Builders
		Error            error
	}{
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Events: events.Builders{
				EvtReceiveADRParamSetupAnswer,
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_ADR_PARAM_SETUP),
		},
		{
			Name: "limit 32768, delay 1024",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_ADRParamSetupReq{
							AdrAckLimitExponent: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768,
							AdrAckDelayExponent: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						AdrAckLimitExponent: &ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768},
						AdrAckDelayExponent: &ttnpb.ADRAckDelayExponentValue{Value: ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024},
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests:   []*ttnpb.MACCommand{},
				},
			},
			Events: events.Builders{
				EvtReceiveADRParamSetupAnswer,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.Device)

				evs, err := HandleADRParamSetupAns(ctx, dev)
				if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
					tc.Error == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(dev, should.Resemble, tc.Expected)
				a.So(evs, should.ResembleEventBuilders, tc.Events)
			},
		})
	}
}
