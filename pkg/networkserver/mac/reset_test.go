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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHandleResetInd(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_ResetInd
		Events           events.Builders
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				SupportsJoin:      false,
				MacState:          &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				SupportsJoin:      false,
				MacState:          &ttnpb.MACState{},
			},
			Error: ErrNoPayload,
		},
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				SupportsJoin:      false,
				FrequencyPlanId:   test.EUFrequencyPlanID,
				MacState: &ttnpb.MACState{
					CurrentParameters: genMACParameters(),
					DesiredParameters: genMACParameters(),
					QueuedResponses:   []*ttnpb.MACCommand{},
				},
			},
			Expected: func() *ttnpb.EndDevice {
				dev := &ttnpb.EndDevice{
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					SupportsJoin:      false,
					FrequencyPlanId:   test.EUFrequencyPlanID,
				}
				macState, err := NewState(dev, frequencyplans.NewStore(test.FrequencyPlansFetcher), &ttnpb.MACSettings{})
				if err != nil {
					t.Fatalf("Failed to reset MACState: %v", errors.Stack(err))
				}
				dev.MacState = macState
				dev.MacState.LorawanVersion = ttnpb.MACVersion_MAC_V1_1
				dev.MacState.QueuedResponses = []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					}).MACCommand(),
				}
				return dev
			}(),
			Payload: &ttnpb.MACCommand_ResetInd{
				MinorVersion: 1,
			},
			Events: events.Builders{
				EvtReceiveResetIndication.With(events.WithData(&ttnpb.MACCommand_ResetInd{
					MinorVersion: 1,
				})),
				EvtEnqueueResetConfirmation.With(events.WithData(&ttnpb.MACCommand_ResetConf{
					MinorVersion: 1,
				})),
			},
		},
		{
			Name: "non-empty queue",
			Device: &ttnpb.EndDevice{
				LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
				LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
				SupportsJoin:      false,
				FrequencyPlanId:   test.EUFrequencyPlanID,
				MacState: &ttnpb.MACState{
					CurrentParameters: genMACParameters(),
					DesiredParameters: genMACParameters(),
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: func() *ttnpb.EndDevice {
				dev := &ttnpb.EndDevice{
					LorawanVersion:    ttnpb.MACVersion_MAC_V1_1,
					LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_1_REV_B,
					SupportsJoin:      false,
					FrequencyPlanId:   test.EUFrequencyPlanID,
				}
				macState, err := NewState(dev, frequencyplans.NewStore(test.FrequencyPlansFetcher), &ttnpb.MACSettings{})
				if err != nil {
					t.Fatalf("Failed to reset MACState: %v", errors.Stack(err))
				}
				dev.MacState = macState
				dev.MacState.LorawanVersion = ttnpb.MACVersion_MAC_V1_1
				dev.MacState.QueuedResponses = []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					}).MACCommand(),
				}
				return dev
			}(),
			Payload: &ttnpb.MACCommand_ResetInd{
				MinorVersion: 1,
			},
			Events: events.Builders{
				EvtReceiveResetIndication.With(events.WithData(&ttnpb.MACCommand_ResetInd{
					MinorVersion: 1,
				})),
				EvtEnqueueResetConfirmation.With(events.WithData(&ttnpb.MACCommand_ResetConf{
					MinorVersion: 1,
				})),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)

				evs, err := HandleResetInd(ctx, dev, tc.Payload, frequencyplans.NewStore(test.FrequencyPlansFetcher), &ttnpb.MACSettings{})
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

func genMACParameters() *ttnpb.MACParameters {
	randomVal := random.Int63n(100)
	channels := make([]*ttnpb.MACParameters_Channel, randomVal%16)
	for i := range channels {
		channels[i] = &ttnpb.MACParameters_Channel{
			MinDataRateIndex:  ttnpb.DataRateIndex(randomVal % 5),
			MaxDataRateIndex:  ttnpb.DataRateIndex(randomVal % 16),
			UplinkFrequency:   uint64((randomVal + int64(i)) * 1000000000),
			DownlinkFrequency: uint64((randomVal + int64(i)) * 1000000000),
		}
	}
	return &ttnpb.MACParameters{
		MaxEirp:           float32(randomVal),
		AdrDataRateIndex:  ttnpb.DataRateIndex(randomVal % 16),
		AdrTxPowerIndex:   uint32(randomVal % 16),
		AdrNbTrans:        uint32(1 + randomVal%15),
		MaxDutyCycle:      ttnpb.AggregatedDutyCycle(randomVal % 16),
		Channels:          channels,
		Rx1Delay:          5,
		Rx1DataRateOffset: 2,
		Rx2DataRateIndex:  2,
		Rx2Frequency:      868100000,
		BeaconFrequency:   869525000,
		PingSlotFrequency: 869525000,
	}
}
