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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsRxTimingSetupReq(t *testing.T) {
	for _, tc := range []struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Needs       bool
	}{
		{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
		{
			Name: "current(delay:1),desired(delay:1)",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
					},
				},
			},
		},
		{
			Name: "current(delay:1),desired(delay:5)",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_5,
					},
				},
			},
			Needs: true,
		},
		{
			Name: "current(delay:1),desired(delay:5),recent",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
					},
					DesiredParameters: &ttnpb.MACParameters{
						Rx1Delay: ttnpb.RxDelay_RX_DELAY_5,
					},
					RecentMacCommandIdentifiers: []ttnpb.MACCommandIdentifier{
						ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP,
					},
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)
				res := DeviceNeedsRxTimingSetupReq(dev)
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

func TestHandleRxTimingSetupAns(t *testing.T) {
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
				EvtReceiveRxTimingSetupAnswer,
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP),
		},
		{
			Name: "42",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RxTimingSetupReq{
							Delay: 42,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Rx1Delay: 42,
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests:   []*ttnpb.MACCommand{},
				},
			},
			Events: events.Builders{
				EvtReceiveRxTimingSetupAnswer,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.Device)

				evs, err := HandleRxTimingSetupAns(ctx, dev)
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
