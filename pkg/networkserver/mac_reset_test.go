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

	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var frequencyPlansStore *frequencyplans.Store

func TestMain(t *testing.M) {
	testFPS := test.Must(test.NewFrequencyPlansStore()).(test.FrequencyPlansStore)

	frequencyPlansStore = (&config.FrequencyPlans{
		StoreDirectory: testFPS.Directory(),
	}).Store()

	ret := t.Run()

	test.Must(nil, testFPS.Destroy())
	os.Exit(ret)
}

func TestHandleResetInd(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_ResetInd
		Error            error
	}{
		{
			Name: "nil payload",
			Device: &ttnpb.EndDevice{
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					SupportsJoin: false,
				},
				MACState: &ttnpb.MACState{},
			},
			Expected: &ttnpb.EndDevice{
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					SupportsJoin: false,
				},
				MACState: &ttnpb.MACState{},
			},
			Payload: nil,
			Error:   errMissingPayload,
		},
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					DefaultMACParameters: &ttnpb.MACParameters{
						MaxEIRP: 42,
					},
					SupportsJoin: false,
				},
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					MACParameters:        *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					DesiredMACParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					QueuedResponses:      []*ttnpb.MACCommand{},
				},
			},
			Expected: func() *ttnpb.EndDevice {
				dev := &ttnpb.EndDevice{
					EndDeviceVersion: ttnpb.EndDeviceVersion{
						DefaultMACParameters: &ttnpb.MACParameters{
							MaxEIRP: 42,
						},
						SupportsJoin: false,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
				}
				if err := ResetMACState(frequencyPlansStore, dev); err != nil {
					panic(errors.NewWithCause(err, "failed to reset MACState"))
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
			Error: nil,
		},
		{
			Name: "non-empty queue",
			Device: &ttnpb.EndDevice{
				EndDeviceVersion: ttnpb.EndDeviceVersion{
					DefaultMACParameters: &ttnpb.MACParameters{
						MaxEIRP: 42,
					},
					SupportsJoin: false,
				},
				FrequencyPlanID: test.EUFrequencyPlanID,
				MACState: &ttnpb.MACState{
					MACParameters:        *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					DesiredMACParameters: *ttnpb.NewPopulatedMACParameters(test.Randy, false),
					QueuedResponses: []*ttnpb.MACCommand{
						{},
						{},
						{},
					},
				},
			},
			Expected: func() *ttnpb.EndDevice {
				dev := &ttnpb.EndDevice{
					EndDeviceVersion: ttnpb.EndDeviceVersion{
						DefaultMACParameters: &ttnpb.MACParameters{
							MaxEIRP: 42,
						},
						SupportsJoin: false,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
				}
				if err := ResetMACState(frequencyPlansStore, dev); err != nil {
					panic(errors.NewWithCause(err, "failed to reset MACState"))
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
			Error: nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleResetInd(test.Context(), dev, tc.Payload, frequencyPlansStore)
			if tc.Error != nil {
				a.So(err, should.DescribeError, errors.Descriptor(tc.Error))
			} else {
				a.So(err, should.BeNil)
			}

			if !a.So(dev, should.Resemble, tc.Expected) {
				pretty.Ldiff(t, dev, tc.Expected)
			}
		})
	}
}
