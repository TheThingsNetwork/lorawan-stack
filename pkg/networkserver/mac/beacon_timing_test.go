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
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsBeaconTimingReq(t *testing.T) {
	type TestCase struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Needs       bool
	}
	var tcs []TestCase

	tcs = append(tcs,
		TestCase{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
	)
	ForEachClass(t, func(makeClassName func(parts ...string) string, class ttnpb.Class) {
		tcs = append(tcs,
			TestCase{
				Name: makeClassName("empty parameters"),
				InputDevice: &ttnpb.EndDevice{
					MacState: &ttnpb.MACState{
						DeviceClass: class,
					},
				},
			},
		)
	})

	for _, tc := range tcs {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)
				res := DeviceNeedsBeaconTimingReq(dev)
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

func TestHandleBeaconTimingReq(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Events           events.Builders
		Error            error
	}{
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						// TODO: Support BeaconTimingReq. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2431)
					},
				},
			},
		},
		{
			Name: "non-empty queue",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
						// TODO: Support BeaconTimingReq. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2431)
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
				dev := ttnpb.Clone(tc.Device)

				evs, err := HandleBeaconTimingReq(ctx, dev)
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
