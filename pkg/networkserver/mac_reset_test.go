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
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleResetInd(t *testing.T) {
	frequencyPlan := ttnpb.NewPopulatedFrequencyPlan(test.Randy, false)

	band := test.Must(band.GetByID(frequencyPlan.GetBandID())).(band.Band)

	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_ResetInd
		Error            error
	}{
		{
			Name:     "nil payload",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Payload:  nil,
			Error:    common.ErrMissingPayload.New(nil),
		},
		{
			Name: "empty queue",
			Device: &ttnpb.EndDevice{
				MaxTxPower:        42,
				MACState:          ttnpb.NewPopulatedMACState(test.Randy, false),
				MACStateDesired:   ttnpb.NewPopulatedMACState(test.Randy, false),
				QueuedMACCommands: []*ttnpb.MACCommand{},
			},
			Expected: &ttnpb.EndDevice{
				MaxTxPower:      42,
				MACState:        NewMACState(&band, 42, frequencyPlan.DwellTime != nil),
				MACStateDesired: NewMACState(&band, 42, frequencyPlan.DwellTime != nil),
				QueuedMACCommands: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					}).MACCommand(),
				},
			},
			Payload: &ttnpb.MACCommand_ResetInd{
				MinorVersion: 1,
			},
			Error: nil,
		},
		{
			Name: "non-empty queue",
			Device: &ttnpb.EndDevice{
				MaxTxPower:      42,
				MACState:        ttnpb.NewPopulatedMACState(test.Randy, false),
				MACStateDesired: ttnpb.NewPopulatedMACState(test.Randy, false),
				QueuedMACCommands: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 42,
					}).MACCommand(),
				},
			},
			Expected: &ttnpb.EndDevice{
				MaxTxPower:      42,
				MACState:        NewMACState(&band, 42, frequencyPlan.DwellTime != nil),
				MACStateDesired: NewMACState(&band, 42, frequencyPlan.DwellTime != nil),
				QueuedMACCommands: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 42,
					}).MACCommand(),
					(&ttnpb.MACCommand_ResetConf{
						MinorVersion: 1,
					}).MACCommand(),
				},
			},
			Payload: &ttnpb.MACCommand_ResetInd{
				MinorVersion: 1,
			},
			Error: nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleResetInd(test.Context(), dev, tc.Payload, frequencyPlan)
			if tc.Error != nil {
				a.So(err, should.BeError)
				return
			}

			a.So(err, should.BeNil)
			a.So(dev, should.Resemble, tc.Expected)
		})
	}
}
