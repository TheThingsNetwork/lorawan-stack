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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleResetInd(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_ResetInd
		AssertEvents     func(*testing.T, ...events.Event) bool
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				SupportsJoin:      false,
				MACState:          &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				SupportsJoin:      false,
				MACState:          &ttnpb.MACState{},
			},
			AssertEvents: func(t *testing.T, evs ...events.Event) bool {
				return assertions.New(t).So(evs, should.BeEmpty)
			},
			Error: errNoPayload,
		},
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				SupportsJoin:      false,
				FrequencyPlanID:   test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					CurrentParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					DesiredParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					QueuedResponses:   []*ttnpb.MACCommand{},
				},
			},
			Expected: func() *ttnpb.EndDevice {
				dev := &ttnpb.EndDevice{
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					SupportsJoin:      false,
					FrequencyPlanID:   test.EUFrequencyPlanID,
				}
				macState, err := newMACState(dev, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})
				if err != nil {
					t.Fatalf("Failed to reset MACState: %v", errors.Stack(err))
				}
				dev.MACState = macState
				dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1
				dev.MACState.QueuedResponses = []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					}).MACCommand(),
				}
				return dev
			}(),
			Payload: &ttnpb.MACCommand_ResetInd{
				MinorVersion: 1,
			},
			AssertEvents: func(t *testing.T, evs ...events.Event) bool {
				a := assertions.New(t)
				return a.So(evs, should.HaveLength, 2) &&
					a.So(evs[0].Name(), should.Equal, "ns.mac.reset.indication") &&
					a.So(evs[0].Data(), should.Resemble, &ttnpb.MACCommand_ResetInd{
						MinorVersion: 1,
					}) &&
					a.So(evs[1].Name(), should.Equal, "ns.mac.reset.confirmation") &&
					a.So(evs[1].Data(), should.Resemble, &ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					})
			},
		},
		{
			Name: "non-empty queue",
			Device: &ttnpb.EndDevice{
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				SupportsJoin:      false,
				FrequencyPlanID:   test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					CurrentParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					DesiredParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: func() *ttnpb.EndDevice {
				dev := &ttnpb.EndDevice{
					LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
					SupportsJoin:      false,
					FrequencyPlanID:   test.EUFrequencyPlanID,
				}
				macState, err := newMACState(dev, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})
				if err != nil {
					t.Fatalf("Failed to reset MACState: %v", errors.Stack(err))
				}
				dev.MACState = macState
				dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1
				dev.MACState.QueuedResponses = []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					}).MACCommand(),
				}
				return dev
			}(),
			Payload: &ttnpb.MACCommand_ResetInd{
				MinorVersion: 1,
			},
			AssertEvents: func(t *testing.T, evs ...events.Event) bool {
				a := assertions.New(t)
				return a.So(evs, should.HaveLength, 2) &&
					a.So(evs[0].Name(), should.Equal, "ns.mac.reset.indication") &&
					a.So(evs[0].Data(), should.Resemble, &ttnpb.MACCommand_ResetInd{
						MinorVersion: 1,
					}) &&
					a.So(evs[1].Name(), should.Equal, "ns.mac.reset.confirmation") &&
					a.So(evs[1].Data(), should.Resemble, &ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					})
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			var err error
			evs := collectEvents(func() {
				err = handleResetInd(test.Context(), dev, tc.Payload, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})
			})
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(dev, should.Resemble, tc.Expected)
			a.So(tc.AssertEvents(t, evs...), should.BeTrue)
		})
	}
}
