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

	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleRxParamSetupAns(t *testing.T) {
	events := collectEvents("ns.mac.rx_param.*")

	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RxParamSetupAns
		Error            error
		ExpectedEvents   int
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: nil,
			Error:   errMissingPayload,
		},
		{
			Name: "no request",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{},
			},
			Payload: ttnpb.NewPopulatedMACCommand_RxParamSetupAns(test.Randy, false),
			Error:   errMACRequestNotFound,
		},
		{
			Name: "all ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					MACParameters: ttnpb.MACParameters{
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
					MACParameters: ttnpb.MACParameters{
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
			ExpectedEvents: 1,
		},
		{
			Name: "data rate ack",
			Device: &ttnpb.EndDevice{
				MACState: &ttnpb.MACState{
					MACParameters: ttnpb.MACParameters{
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
					MACParameters: ttnpb.MACParameters{
						Rx1DataRateOffset: 99,
						Rx2Frequency:      99,
					},
					PendingRequests: []*ttnpb.MACCommand{},
				},
			},
			Payload: &ttnpb.MACCommand_RxParamSetupAns{
				Rx1DataRateOffsetAck: true,
				Rx2DataRateIndexAck:  true,
				Rx2FrequencyAck:      false,
			},
			ExpectedEvents: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleRxParamSetupAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil {
				a.So(err, should.EqualErrorOrDefinition, tc.Error)
			} else {
				a.So(err, should.BeNil)
			}

			if !a.So(dev, should.Resemble, tc.Expected) {
				pretty.Ldiff(t, dev, tc.Expected)
			}

			if tc.ExpectedEvents > 0 {
				events.expect(t, tc.ExpectedEvents)
			}
		})
	}
}
