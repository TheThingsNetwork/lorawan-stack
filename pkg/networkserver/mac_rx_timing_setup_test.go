// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleRxTimingSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Error            error
		EventAssertion   func(*testing.T, ...events.Event) bool
	}{
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Error: errMACRequestNotFound,
			EventAssertion: func(t *testing.T, evs ...events.Event) bool {
				return assertions.New(t).So(evs, should.BeEmpty)
			},
		},
		{
			Name: "42",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RxTimingSetupReq{
							Delay: 42,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1Delay: 42,
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			EventAssertion: func(t *testing.T, evs ...events.Event) bool {
				a := assertions.New(t)
				return a.So(evs, should.HaveLength, 1) &&
					a.So(evs[0].Name(), should.Equal, "ns.mac.rx_timing_setup.answer") &&
					a.So(evs[0].Data(), should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			var err error
			evs := collectEvents(func() {
				err = handleRxTimingSetupAns(test.Context(), dev)
			})
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(tc.EventAssertion(t, evs...), should.BeTrue)
		})
	}
}
