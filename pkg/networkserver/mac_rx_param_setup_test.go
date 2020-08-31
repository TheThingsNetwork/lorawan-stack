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
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsRxParamSetupReq(t *testing.T) {
	for _, tc := range []struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Defaults    ttnpb.MACSettings
		Needs       bool
	}{
		{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
		{
			Name: "current(data-rate-offset:1,data-rate-index:2,frequency:123),desired(data-rate-offset:1,data-rate-index:2,frequency:123)",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 1,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_2,
						Rx2Frequency:      123,
					},
					DesiredParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 1,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_2,
						Rx2Frequency:      123,
					},
				},
			},
		},
		{
			Name: "current(data-rate-offset:1,data-rate-index:2,frequency:123),desired(data-rate-offset:1,data-rate-index:3,frequency:123)",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 1,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_2,
						Rx2Frequency:      123,
					},
					DesiredParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 1,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_3,
						Rx2Frequency:      123,
					},
				},
			},
			Needs: true,
		},
		{
			Name: "current(data-rate-offset:1,data-rate-index:2,frequency:123),desired(data-rate-offset:1,data-rate-index:2,frequency:124)",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 1,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_2,
						Rx2Frequency:      123,
					},
					DesiredParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 1,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_2,
						Rx2Frequency:      124,
					},
				},
			},
			Needs: true,
		},
		{
			Name: "current(data-rate-offset:1,data-rate-index:2,frequency:123),desired(data-rate-offset:2,data-rate-index:2,frequency:123)",
			InputDevice: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 1,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_2,
						Rx2Frequency:      123,
					},
					DesiredParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 2,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_2,
						Rx2Frequency:      123,
					},
				},
			},
			Needs: true,
		},
	} {
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.InputDevice)
				res := deviceNeedsRxParamSetupReq(dev)
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

func TestHandleRxParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RxParamSetupAns
		Events           events.Builders
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Error: errNoPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: &ttnpb.MACCommand_RxParamSetupAns{
				Rx1DataRateOffsetAck: true,
				Rx2DataRateIndexAck:  true,
				Rx2FrequencyAck:      true,
			},
			Events: events.Builders{
				evtReceiveRxParamSetupAccept.With(events.WithData(&ttnpb.MACCommand_RxParamSetupAns{
					Rx1DataRateOffsetAck: true,
					Rx2DataRateIndexAck:  true,
					Rx2FrequencyAck:      true,
				})),
			},
			Error: errMACRequestNotFound,
		},
		{
			Name: "all ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 99,
						Rx2Frequency:      99,
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RxParamSetupReq{
							Rx1DataRateOffset: 42,
							Rx2DataRateIndex:  43,
							Rx2Frequency:      44,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 42,
						Rx2DataRateIndex:  43,
						Rx2Frequency:      44,
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_RxParamSetupAns{
				Rx1DataRateOffsetAck: true,
				Rx2DataRateIndexAck:  true,
				Rx2FrequencyAck:      true,
			},
			Events: events.Builders{
				evtReceiveRxParamSetupAccept.With(events.WithData(&ttnpb.MACCommand_RxParamSetupAns{
					Rx1DataRateOffsetAck: true,
					Rx2DataRateIndexAck:  true,
					Rx2FrequencyAck:      true,
				})),
			},
		},
		{
			Name: "data rate ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 99,
						Rx2Frequency:      99,
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RxParamSetupReq{
							Rx1DataRateOffset: 42,
							Rx2DataRateIndex:  43,
							Rx2Frequency:      44,
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					CurrentParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 99,
						Rx2Frequency:      99,
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_RxParamSetupAns{
				Rx1DataRateOffsetAck: true,
				Rx2DataRateIndexAck:  true,
			},
			Events: events.Builders{
				evtReceiveRxParamSetupReject.With(events.WithData(&ttnpb.MACCommand_RxParamSetupAns{
					Rx1DataRateOffsetAck: true,
					Rx2DataRateIndexAck:  true,
				})),
			},
		},
	} {
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := CopyEndDevice(tc.Device)

				evs, err := handleRxParamSetupAns(ctx, dev, tc.Payload)
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
