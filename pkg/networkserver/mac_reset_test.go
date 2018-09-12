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
	"os"
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var frequencyPlansStore *frequencyplans.Store

func TestMain(t *testing.M) {
	frequencyPlansStore = frequencyplans.NewStore(test.FrequencyPlansFetcher)

	ret := t.Run()

	os.Exit(ret)
}

func TestHandleResetInd(t *testing.T) {
	events := collectEvents("ns.mac.reset_ind")

	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_ResetInd
		Error            error
		ExpectedEvents   int
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				SupportsJoin: false,
				MACState:     &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				SupportsJoin: false,
				MACState:     &ttnpb.MACState{},
			},
			Payload: nil,
			Error:   errMissingPayload,
		},
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				DefaultMACParameters: &ttnpb.MACParameters{
					MaxEIRP: 42,
				},
				SupportsJoin:    false,
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					CurrentParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					DesiredParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					QueuedResponses:   []*ttnpb.MACCommand{},
				},
			},
			Expected: func() *ttnpb.EndDevice {
				dev := &ttnpb.EndDevice{
					DefaultMACParameters: &ttnpb.MACParameters{
						MaxEIRP: 42,
					},
					SupportsJoin:    false,
					FrequencyPlanID: test.EUFrequencyPlanID,
				}
				if err := ResetMACState(frequencyPlansStore, dev); err != nil {
					panic(errors.New("failed to reset MACState").WithCause(err))
				}

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
			Error:          nil,
			ExpectedEvents: 1,
		},
		{
			Name: "non-empty queue",
			Device: &ttnpb.EndDevice{
				DefaultMACParameters: &ttnpb.MACParameters{
					MaxEIRP: 42,
				},
				SupportsJoin:    false,
				FrequencyPlanID: test.EUFrequencyPlanID,
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
					DefaultMACParameters: &ttnpb.MACParameters{
						MaxEIRP: 42,
					},
					SupportsJoin:    false,
					FrequencyPlanID: test.EUFrequencyPlanID,
				}
				if err := ResetMACState(frequencyPlansStore, dev); err != nil {
					panic(errors.New("failed to reset MACState").WithCause(err))
				}

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
			Error:          nil,
			ExpectedEvents: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleResetInd(test.Context(), dev, tc.Payload, frequencyPlansStore)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}

			if tc.ExpectedEvents > 0 {
				events.expect(t, tc.ExpectedEvents)
			}
			a.So(dev, should.Resemble, tc.Expected)
		})
	}
}
